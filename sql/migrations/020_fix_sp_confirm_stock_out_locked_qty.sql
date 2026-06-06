-- ============================================================
-- 功能：修复 sp_confirm_stock_out 出库确认时未减少 locked_quantity 的问题
--       当销售订单确认时已锁定库存，出库确认只减 quantity 不行
--       需要同时减少 locked_quantity
-- 创建时间：2026-06-06
-- 创建人：wangcw
-- MIGRATION_ID: 020_fix_sp_confirm_stock_out_locked_qty
-- MIGRATION_APPLIED: pending
-- ============================================================

BEGIN;

DROP PROCEDURE IF EXISTS sp_confirm_stock_out CASCADE;

CREATE OR REPLACE PROCEDURE public.sp_confirm_stock_out(IN p_stock_out_id bigint, IN p_operator_id bigint)
 LANGUAGE plpgsql
AS $procedure$
DECLARE
	v_so           RECORD;
	v_item         RECORD;
	v_inv          RECORD;
	v_balance      NUMERIC(18,3);
	v_operator_name VARCHAR;
	v_ref_wh_id    BIGINT;
	v_min_wh       BIGINT;
	v_max_wh       BIGINT;
    v_need         INTEGER;
    v_selected_cnt INTEGER;
BEGIN
	SELECT id, stock_out_no, stock_out_date, out_type, ref_doc_type, ref_doc_id, status
	INTO v_so
	FROM stock_out
	WHERE id = p_stock_out_id AND deleted_at IS NULL
	FOR UPDATE;

	IF v_so IS NULL THEN
		RAISE EXCEPTION '出库单不存在或已删除 [ID=%]', p_stock_out_id;
	END IF;

	IF COALESCE(v_so.status, 'draft') = 'confirmed' THEN
		RAISE EXCEPTION '出库单已完成，无需重复确认';
	END IF;

	SELECT u.real_name INTO v_operator_name FROM sys_user u WHERE u.id = p_operator_id;

	SELECT MIN(i.warehouse_id), MAX(i.warehouse_id)
	INTO v_min_wh, v_max_wh
	FROM stock_out_item soi
	INNER JOIN inventory i ON i.id = soi.inventory_id
	WHERE soi.stock_out_id = p_stock_out_id
	  AND COALESCE(soi.inventory_id, 0) <> 0;

	IF v_max_wh IS NOT NULL AND v_min_wh IS DISTINCT FROM v_max_wh THEN
		RAISE EXCEPTION '出库明细涉及多个仓库，请拆单后再确认';
	END IF;
	v_ref_wh_id := v_min_wh;

	FOR v_item IN
		SELECT soi.*, m.material_code, m.material_name, m.is_code
		FROM stock_out_item soi
		INNER JOIN material m ON m.id = soi.material_id
		WHERE soi.stock_out_id = p_stock_out_id
	LOOP
		IF COALESCE(v_item.inventory_id, 0) <> 0 THEN
			SELECT i.*
			INTO v_inv
			FROM inventory i
			WHERE i.id = v_item.inventory_id
			FOR UPDATE;

			IF v_inv IS NULL THEN
				RAISE EXCEPTION '库存不存在 [inventory_id=%]', v_item.inventory_id;
			END IF;

			-- 销售出库：sp_confirm_sales_order 已经预先锁定库存，
			-- 此处检查 v_inv.quantity（包含被本单锁定的部分）
			-- 非销售出库：检查可用量（quantity - locked_quantity）
			IF v_so.ref_doc_type = 'sales_order' THEN
				IF v_inv.quantity < v_item.quantity THEN
					RAISE EXCEPTION '物料[%-%]库存不足，需要:%, 现有:%',
						v_item.material_code, v_item.material_name,
						v_item.quantity, v_inv.quantity;
				END IF;
			ELSE
				IF (v_inv.quantity - v_inv.locked_quantity) < v_item.quantity THEN
					RAISE EXCEPTION '物料[%-%]库存不足，需要:%, 可用:%',
						v_item.material_code, v_item.material_name,
						v_item.quantity, (v_inv.quantity - v_inv.locked_quantity);
				END IF;
			END IF;

			UPDATE inventory
			SET quantity = quantity - v_item.quantity,
				locked_quantity = GREATEST(COALESCE(locked_quantity, 0) - v_item.quantity, 0),
				in_transit_quantity = GREATEST(COALESCE(in_transit_quantity, 0) - v_item.quantity, 0),
				updated_at = NOW()
			WHERE id = v_item.inventory_id;

			SELECT quantity INTO v_balance FROM inventory WHERE id = v_item.inventory_id;

			INSERT INTO inventory_transaction (
				material_id, warehouse_id, trans_type, quantity,
				balance, ref_doc_type, ref_doc_no, ref_doc_id,
				operator_id, remark, created_at
			)
			VALUES (
				v_item.material_id, v_inv.warehouse_id, 'out', -v_item.quantity, v_balance,
				'stock_out', v_so.stock_out_no, p_stock_out_id,
				p_operator_id, '出库确认', NOW()
			);

			IF v_item.is_code THEN
				v_need := FLOOR(v_item.quantity)::INT;

				SELECT COUNT(*) INTO v_selected_cnt
				FROM stock_out_item_serial_selection
				WHERE stock_out_item_id = v_item.id;

				IF v_selected_cnt <> v_need THEN
					RAISE EXCEPTION '物料[%]编码未备齐，需:% 已备:%', v_item.material_name, v_need, v_selected_cnt;
				END IF;

				WITH picked AS (
					UPDATE sku_serial_code sc
					SET status = 'issued',
						updated_at = NOW()
					FROM stock_out_item_serial_selection s
					WHERE sc.id = s.serial_code_id
					  AND s.stock_out_item_id = v_item.id
					  AND sc.status = 'in_stock'
					RETURNING sc.id, sc.serial_code
				)
                INSERT INTO sku_serial_trace (
                    serial_code_id, serial_code, action,
                    ref_doc_type, ref_doc_no, ref_doc_id,
                    from_warehouse_id, operator_id, operator_name, remark
                )
                SELECT id, serial_code, 'stock_out', 'stock_out', v_so.stock_out_no, p_stock_out_id,
                       v_inv.warehouse_id, p_operator_id, v_operator_name, '出库确认-编码已领用(手动)'
                FROM picked;

				GET DIAGNOSTICS v_selected_cnt = ROW_COUNT;
				IF v_selected_cnt <> v_need THEN
					RAISE EXCEPTION '部分备货编码状态已变化，请重新备货';
				END IF;
			END IF;
		END IF;
	END LOOP;

	UPDATE stock_out
	SET status = 'confirmed',
		confirmed_at = NOW(),
		updated_by = p_operator_id,
		updated_at = NOW()
	WHERE id = p_stock_out_id;

	IF v_so.ref_doc_type = 'consumption_order' THEN
		UPDATE consumption_order SET status = 'completed', updated_at = NOW() WHERE id = v_so.ref_doc_id;
	END IF;

    DELETE FROM stock_out_item_serial_selection
    WHERE stock_out_item_id IN (SELECT id FROM stock_out_item WHERE stock_out_id = p_stock_out_id);
END;
$procedure$;

COMMIT;
