-- ============================================================
-- 功能：修复 sp_confirm_stock_in / sp_confirm_stock_out 中使用已删除的 sku_serial_xxx 表
--   统一改为 material_serial_code / material_serial_trace
-- 创建时间：2026-06-06
-- 创建人：wangcw
-- MIGRATION_ID: 025_fix_procedures_use_material_serial_tables
-- MIGRATION_APPLIED: pending
-- ============================================================

BEGIN;

-- ========== 修复 sp_confirm_stock_in ==========
DROP PROCEDURE IF EXISTS sp_confirm_stock_in CASCADE;

CREATE OR REPLACE PROCEDURE public.sp_confirm_stock_in(IN p_stock_in_id bigint, IN p_operator_id bigint)
 LANGUAGE plpgsql
AS $procedure$
DECLARE
    v_stock_in RECORD;
    v_item RECORD;
    v_serial_code VARCHAR;
    v_inv_id BIGINT;
    v_operator_name VARCHAR;
    v_i INT;
BEGIN
    SELECT si.* INTO v_stock_in
    FROM stock_in si
    WHERE si.id = p_stock_in_id
      AND si.deleted_at IS NULL
    FOR UPDATE;

    IF v_stock_in IS NULL THEN
        RAISE EXCEPTION '入库单不存在 [id=%]', p_stock_in_id;
    END IF;

    IF v_stock_in.stock_in_status = 'passed' THEN
        RAISE EXCEPTION '入库单已全部入库，无需重复确认';
    END IF;

    SELECT u.real_name INTO v_operator_name
    FROM sys_user u
    WHERE u.id = p_operator_id;

    FOR v_item IN
        SELECT sii.*, m.is_code, m.material_code, m.material_name
        FROM stock_in_item sii
        JOIN material m ON m.id = sii.material_id
        WHERE sii.stock_in_id = p_stock_in_id
    LOOP
        IF COALESCE(v_item.accepted_quantity, 0) <= 0 THEN
            CONTINUE;
        END IF;

        INSERT INTO inventory (
            material_id, warehouse_id, quantity, locked_quantity, unit, stock_in_date, unit_cost, material_inspection_no
        ) VALUES (
            v_item.material_id, v_stock_in.warehouse_id, v_item.accepted_quantity, 0, v_item.unit,
            v_stock_in.stock_in_date, COALESCE(v_item.unit_cost, 0), NULL
        )
        ON CONFLICT (material_id, warehouse_id, COALESCE(material_inspection_no, ''))
        DO UPDATE SET
            quantity = inventory.quantity + v_item.accepted_quantity,
            unit_cost = CASE WHEN v_item.unit_cost > 0 THEN v_item.unit_cost ELSE inventory.unit_cost END,
            updated_at = NOW()
        RETURNING id INTO v_inv_id;

        INSERT INTO inventory_transaction (
            material_id, warehouse_id, trans_type, quantity, balance, ref_doc_type, ref_doc_no, ref_doc_id, operator_id, created_at
        ) VALUES (
            v_item.material_id, v_stock_in.warehouse_id, 'in', v_item.accepted_quantity,
            (SELECT quantity FROM inventory WHERE id = v_inv_id), 'stock_in', v_stock_in.stock_in_no, p_stock_in_id, p_operator_id, NOW()
        );

        IF v_item.is_code THEN
            FOR v_i IN 1..floor(v_item.accepted_quantity) LOOP
                v_serial_code := fn_generate_serial_no('MAT');

                INSERT INTO material_serial_code (
                    serial_code,
                    material_id, material_code, material_name,
                    stock_in_id, stock_in_item_id,
                    inventory_id, warehouse_id,
                    status, created_at, updated_at
                ) VALUES (
                    v_serial_code,
                    v_item.material_id, v_item.material_code, v_item.material_name,
                    p_stock_in_id, v_item.id,
                    v_inv_id, v_stock_in.warehouse_id,
                    'in_stock', NOW(), NOW()
                );

                INSERT INTO material_serial_trace (
                    serial_code_id, serial_code, action, ref_doc_type, ref_doc_no, ref_doc_id,
                    to_warehouse_id, operator_id, operator_name, remark, created_at
                ) VALUES (
                    (SELECT id FROM material_serial_code WHERE serial_code = v_serial_code),
                    v_serial_code,
                    'stock_in', 'stock_in', v_stock_in.stock_in_no, p_stock_in_id,
                    v_stock_in.warehouse_id, p_operator_id, COALESCE(v_operator_name, ''), '入库确认生成编码', NOW()
                );
            END LOOP;
        END IF;
    END LOOP;

    UPDATE stock_in
    SET stock_in_status = 'passed',
        updated_by = p_operator_id,
        updated_at = NOW()
    WHERE id = p_stock_in_id;
END;
$procedure$;

-- ========== 修复 sp_confirm_stock_out ==========
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

			-- 销售出库：sp_confirm_sales_order 已经预先锁定库存
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
					UPDATE material_serial_code sc
					SET status = 'issued',
						updated_at = NOW()
					FROM stock_out_item_serial_selection s
					WHERE sc.id = s.serial_code_id
					  AND s.stock_out_item_id = v_item.id
					  AND sc.status = 'in_stock'
					RETURNING sc.id, sc.serial_code
				)
                INSERT INTO material_serial_trace (
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
