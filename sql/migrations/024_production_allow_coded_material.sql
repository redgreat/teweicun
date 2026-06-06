-- ============================================================
-- 功能：生产入库支持编码物料的自动生成编码
--   sp_confirm_stock_in 已为编码物料自动生成 serial codes，
--   移除 sp_generate_production_from_consumption 中对编码物料的拦截
-- 创建时间：2026-06-06
-- 创建人：wangcw
-- MIGRATION_ID: 024_production_allow_coded_material
-- MIGRATION_APPLIED: pending
-- ============================================================

BEGIN;

CREATE OR REPLACE PROCEDURE sp_generate_production_from_consumption(
	IN p_consumption_order_id BIGINT,
	IN p_stock_out_id BIGINT,
	IN p_operator_id BIGINT
)
LANGUAGE plpgsql
AS $procedure$
DECLARE
	v_co RECORD;
	v_wh RECORD;
	v_mat RECORD;
	v_exists BIGINT;
	v_total_cost NUMERIC(18, 3);
	v_unit_cost NUMERIC(18, 3);
	v_stock_in_id BIGINT;
	v_stock_in_no VARCHAR(20);
	v_production_no VARCHAR(20);
	v_production_id BIGINT;
BEGIN
	IF COALESCE(p_operator_id, 0) = 0 THEN
		RETURN;
	END IF;

	SELECT
		co.id,
		co.order_no,
		co.order_date,
		co.status,
		co.produced_material_id,
		co.produced_quantity,
		co.produced_warehouse_id
	INTO v_co
	FROM consumption_order co
	WHERE co.id = p_consumption_order_id
	  AND co.deleted_at IS NULL
	FOR UPDATE;

	IF v_co IS NULL THEN
		RETURN;
	END IF;

	IF v_co.produced_material_id IS NULL OR v_co.produced_quantity IS NULL OR v_co.produced_warehouse_id IS NULL THEN
		RETURN;
	END IF;

	SELECT id INTO v_exists
	FROM production_order
	WHERE consumption_order_id = p_consumption_order_id
	LIMIT 1;

	IF v_exists IS NOT NULL THEN
		RETURN;
	END IF;

	SELECT id, warehouse_code, warehouse_name INTO v_wh
	FROM warehouse
	WHERE id = v_co.produced_warehouse_id AND deleted_at IS NULL;

	IF v_wh IS NULL THEN
		RAISE EXCEPTION '成品入库仓库不存在 [warehouse_id=%]', v_co.produced_warehouse_id;
	END IF;

	SELECT id, material_code, material_name, COALESCE(NULLIF(TRIM(unit), ''), '件') AS unit, COALESCE(is_code, FALSE) AS is_code
	INTO v_mat
	FROM material
	WHERE id = v_co.produced_material_id AND deleted_at IS NULL;

	IF v_mat IS NULL THEN
		RAISE EXCEPTION '成品物料不存在 [material_id=%]', v_co.produced_material_id;
	END IF;

	-- 编码物料也允许自动生产入库，sp_confirm_stock_in 会自动生成编码
	-- 移除了原有的 is_code 拦截

	SELECT COALESCE(SUM(coi.quantity * COALESCE(inv.unit_cost, 0)), 0)
	INTO v_total_cost
	FROM consumption_order_item coi
	LEFT JOIN inventory inv ON inv.id = coi.inventory_id
	WHERE coi.order_id = p_consumption_order_id;

	IF v_co.produced_quantity <= 0 THEN
		RAISE EXCEPTION '成品数量必须 > 0';
	END IF;

	v_unit_cost := v_total_cost / v_co.produced_quantity;
	v_stock_in_no := fn_generate_serial_no('SI');

	INSERT INTO stock_in (
		stock_in_no,
		warehouse_id, warehouse_code,
		stock_in_date,
		stock_in_status,
		stock_in_type,
		remark,
		created_by, created_at, updated_by, updated_at
	) VALUES (
		v_stock_in_no,
		v_wh.id, v_wh.warehouse_code,
		COALESCE(v_co.order_date, CURRENT_DATE),
		'preparing',
		'production',
		'生产单自动生成（领料出库确认后）',
		p_operator_id, NOW(), p_operator_id, NOW()
	)
	RETURNING id INTO v_stock_in_id;

	INSERT INTO stock_in_item (
		stock_in_id,
		material_id,
		arrived_quantity,
		accepted_quantity,
		unit,
		unit_cost,
		custom_attributes,
		created_at
	) VALUES (
		v_stock_in_id,
		v_mat.id,
		v_co.produced_quantity,
		v_co.produced_quantity,
		v_mat.unit,
		v_unit_cost,
		'[]'::jsonb,
		NOW()
	);

	CALL sp_confirm_stock_in(v_stock_in_id, p_operator_id);

	v_production_no := fn_generate_serial_no('PR');

	INSERT INTO production_order (
		production_no,
		consumption_order_id,
		stock_out_id,
		stock_in_id,
		produced_material_id,
		produced_warehouse_id,
		produced_quantity,
		produced_unit_cost,
		cost_price,
		status,
		remark,
		created_by,
		updated_by
	) VALUES (
		v_production_no,
		p_consumption_order_id,
		p_stock_out_id,
		v_stock_in_id,
		v_mat.id,
		v_wh.id,
		v_co.produced_quantity,
		v_unit_cost,
		v_total_cost,
		'completed',
		'由领料出库确认自动生成',
		p_operator_id,
		p_operator_id
	)
	RETURNING id INTO v_production_id;

	-- 将新生产单关联回领料单
	UPDATE consumption_order
	SET production_order_id = v_production_id,
	    updated_at = NOW(),
	    updated_by = p_operator_id
	WHERE id = p_consumption_order_id;
END;
$procedure$;

COMMIT;
