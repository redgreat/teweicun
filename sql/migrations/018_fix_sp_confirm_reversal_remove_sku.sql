-- ============================================================
-- 功能：修复 sp_confirm_reversal_order 中已删除的 material_sku 表及 sku_id 列引用
-- 创建时间：2026-06-06
-- 创建人：wangcw
-- MIGRATION_ID: 018_fix_sp_confirm_reversal_remove_sku
-- MIGRATION_APPLIED: pending
-- ============================================================

BEGIN;

DROP PROCEDURE IF EXISTS sp_confirm_reversal_order CASCADE;

CREATE OR REPLACE PROCEDURE public.sp_confirm_reversal_order(IN p_order_id bigint, IN p_operator_id bigint)
 LANGUAGE plpgsql
AS $procedure$
DECLARE
	v_order       RECORD;
	v_item        RECORD;
	v_stock_in_id BIGINT;
	v_stock_in_no VARCHAR(20);
BEGIN
	SELECT *
	INTO v_order
	FROM reversal_order
	WHERE id = p_order_id
	  AND deleted_at IS NULL
	FOR UPDATE;

	IF v_order IS NULL THEN
		RAISE EXCEPTION '退料订单不存在 [id=%]', p_order_id;
	END IF;

	IF v_order.status NOT IN ('pending', 'confirmed', 'completed') THEN
		RAISE EXCEPTION '单据状态[%]不允许确认', v_order.status;
	END IF;

	IF COALESCE(v_order.stock_in_id, 0) <> 0 THEN
		UPDATE reversal_order
		SET status = 'confirmed',
			updated_by = p_operator_id,
			updated_at = NOW()
		WHERE id = p_order_id;
		RETURN;
	END IF;

	v_stock_in_no := fn_generate_serial_no('SI');

	INSERT INTO stock_in (
		stock_in_no, warehouse_id, warehouse_code, stock_in_date, stock_in_status, stock_in_type,
		remark, created_by, created_at, updated_by, updated_at
	) VALUES (
		v_stock_in_no, v_order.warehouse_id, v_order.warehouse_code, COALESCE(v_order.order_date, CURRENT_DATE),
		'preparing', 'reversal', '退料订单自动生成', p_operator_id, NOW(), p_operator_id, NOW()
	)
	RETURNING id INTO v_stock_in_id;

	FOR v_item IN
		SELECT roi.material_id,
		       roi.quantity,
		       roi.unit,
		       roi.inventory_id,
		       COALESCE(inv.unit_cost, 0) AS unit_cost,
		       '[]'::jsonb AS custom_attributes,
		       COALESCE(NULLIF(TRIM(inv.unit), ''), NULLIF(TRIM(roi.unit), ''), NULLIF(TRIM(m.unit), ''), '件') AS final_unit
		FROM reversal_order_item roi
		LEFT JOIN inventory inv ON inv.id = roi.inventory_id
		LEFT JOIN material m ON m.id = roi.material_id
		WHERE roi.order_id = p_order_id
		ORDER BY roi.id
	LOOP
		INSERT INTO stock_in_item (
			stock_in_id, material_id, custom_attributes,
			arrived_quantity, accepted_quantity, unit, unit_cost, created_at
		) VALUES (
			v_stock_in_id, v_item.material_id, v_item.custom_attributes,
			v_item.quantity, v_item.quantity, v_item.final_unit, v_item.unit_cost, NOW()
		);
	END LOOP;

	UPDATE reversal_order
	SET status = 'confirmed',
		stock_in_id = v_stock_in_id,
		updated_by = p_operator_id,
		updated_at = NOW()
	WHERE id = p_order_id;
END;
$procedure$;

COMMIT;
