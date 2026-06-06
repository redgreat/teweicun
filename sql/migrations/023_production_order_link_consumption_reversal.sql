-- MIGRATION_ID: 023_production_order_link_consumption_reversal
-- MIGRATION_APPLIED: pending

BEGIN;

-- ============================================
-- 1. consumption_order 新增生产单/生产退货单关联字段
-- ============================================
ALTER TABLE consumption_order
	ADD COLUMN IF NOT EXISTS production_order_id BIGINT NULL REFERENCES production_order(id),
	ADD COLUMN IF NOT EXISTS production_return_order_id BIGINT NULL REFERENCES production_return_order(id);

COMMENT ON COLUMN consumption_order.production_order_id IS '关联的生产单ID（可空；多个领料单可关联同一生产单）';
COMMENT ON COLUMN consumption_order.production_return_order_id IS '关联的生产退货单ID（可空；多个领料单可关联同一生产退货单）';

-- ============================================
-- 2. reversal_order 新增生产单/生产退货单关联字段
-- ============================================
ALTER TABLE reversal_order
	ADD COLUMN IF NOT EXISTS production_order_id BIGINT NULL REFERENCES production_order(id),
	ADD COLUMN IF NOT EXISTS production_return_order_id BIGINT NULL REFERENCES production_return_order(id);

COMMENT ON COLUMN reversal_order.production_order_id IS '关联的生产单ID（可空；多个退料单可关联同一生产单）';
COMMENT ON COLUMN reversal_order.production_return_order_id IS '关联的生产退货单ID（可空；多个退料单可关联同一生产退货单）';

-- ============================================
-- 3. production_order 解除 1:1 绑定，新增可编辑成本字段
-- ============================================
-- 移除 consumption_order_id 的 UNIQUE 约束
ALTER TABLE production_order
	DROP CONSTRAINT IF EXISTS production_order_consumption_order_id_key;

-- 消费 order_id 改为可空（允许直接创建生产单而不绑定领料单）
ALTER TABLE production_order
	ALTER COLUMN consumption_order_id DROP NOT NULL;

-- 新增可编辑成本价格字段
ALTER TABLE production_order
	ADD COLUMN IF NOT EXISTS cost_price NUMERIC(18, 3) NOT NULL DEFAULT 0;

COMMENT ON COLUMN production_order.cost_price IS '生产成本价格（可编辑；系统自动计算 = 关联领料/退料单金额合计）';

-- ============================================
-- 4. production_return_order 新增可编辑成本字段
-- ============================================
ALTER TABLE production_return_order
	ADD COLUMN IF NOT EXISTS cost_price NUMERIC(18, 3) NOT NULL DEFAULT 0;

COMMENT ON COLUMN production_return_order.cost_price IS '生产退货成本价格（可编辑；系统自动计算 = 关联领料/退料单金额合计的负数）';

-- ============================================
-- 5. 重建 v_production_order_list 视图，包含新字段
-- ============================================
DROP VIEW IF EXISTS v_production_order_list;
CREATE VIEW v_production_order_list AS
SELECT
	po.id,
	po.production_no,
	po.status,
	po.consumption_order_id,
	co.order_no AS consumption_order_no,
	po.stock_out_id,
	so.stock_out_no,
	po.stock_in_id,
	si.stock_in_no,
	po.produced_material_id,
	m.material_code AS produced_material_code,
	m.material_name AS produced_material_name,
	po.produced_warehouse_id,
	w.warehouse_code AS produced_warehouse_code,
	w.warehouse_name AS produced_warehouse_name,
	po.produced_quantity,
	po.produced_unit_cost,
	po.cost_price,
	po.remark,
	po.created_at,
	po.updated_at
FROM production_order po
LEFT JOIN consumption_order co ON co.id = po.consumption_order_id AND co.deleted_at IS NULL
LEFT JOIN stock_out so ON so.id = po.stock_out_id AND so.deleted_at IS NULL
LEFT JOIN stock_in si ON si.id = po.stock_in_id AND si.deleted_at IS NULL
LEFT JOIN material m ON m.id = po.produced_material_id
LEFT JOIN warehouse w ON w.id = po.produced_warehouse_id AND w.deleted_at IS NULL;

-- ============================================
-- 6. 重建 v_production_return_order_list 视图，包含新字段
-- ============================================
DROP VIEW IF EXISTS v_production_return_order_list;
CREATE VIEW v_production_return_order_list AS
SELECT
	pro.id,
	pro.return_no,
	pro.status,
	pro.production_order_id,
	po.production_no,
	po.consumption_order_id,
	co.order_no AS consumption_order_no,
	pro.stock_out_id,
	so.stock_out_no,
	po.produced_material_id,
	m.material_code AS produced_material_code,
	m.material_name AS produced_material_name,
	po.produced_warehouse_id,
	w.warehouse_code AS produced_warehouse_code,
	w.warehouse_name AS produced_warehouse_name,
	pro.returned_quantity,
	pro.cost_price,
	pro.remark,
	pro.created_at,
	pro.updated_at
FROM production_return_order pro
LEFT JOIN production_order po ON po.id = pro.production_order_id
LEFT JOIN consumption_order co ON co.id = po.consumption_order_id AND co.deleted_at IS NULL
LEFT JOIN stock_out so ON so.id = pro.stock_out_id AND so.deleted_at IS NULL
LEFT JOIN material m ON m.id = po.produced_material_id
LEFT JOIN warehouse w ON w.id = po.produced_warehouse_id AND w.deleted_at IS NULL;

-- ============================================
-- 7. 更新 sp_generate_production_from_consumption 以支持新字段
--    （如果领料单已关联生产单则不重复创建，而是更新关联关系）
-- ============================================
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
	v_target_production_order_id BIGINT;
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
		co.produced_warehouse_id,
		co.production_order_id
	INTO v_co
	FROM consumption_order co
	WHERE co.id = p_consumption_order_id
	  AND co.deleted_at IS NULL
	FOR UPDATE;

	IF v_co IS NULL THEN
		RETURN;
	END IF;

	-- 如果领料单没有配置成品生产信息，不生成生产单
	IF v_co.produced_material_id IS NULL OR v_co.produced_quantity IS NULL OR v_co.produced_warehouse_id IS NULL THEN
		RETURN;
	END IF;

	-- 查找领料单关联的生产单
	v_target_production_order_id := v_co.production_order_id;

	-- 如果领料单没有直接关联生产单，按旧逻辑查找由该领料单生成的唯一生产单
	IF v_target_production_order_id IS NULL THEN
		SELECT id INTO v_target_production_order_id
		FROM production_order
		WHERE consumption_order_id = p_consumption_order_id
		LIMIT 1;
	END IF;

	-- 如果找到了已存在的生产单，不需要重复创建（但可以更新成本）
	IF v_target_production_order_id IS NOT NULL THEN
		-- 更新生产单成本价格 = 所有关联领料单金额合计
		UPDATE production_order
		SET cost_price = (
			SELECT COALESCE(SUM(coi.quantity * COALESCE(inv.unit_cost, 0)), 0)
			FROM consumption_order_item coi
			LEFT JOIN inventory inv ON inv.id = coi.inventory_id
			WHERE coi.order_id IN (
				SELECT id FROM consumption_order
				WHERE production_order_id = v_target_production_order_id
				  AND deleted_at IS NULL
			)
		),
		updated_at = NOW(),
		updated_by = p_operator_id
		WHERE id = v_target_production_order_id;

		-- 同时更新成品入库单的成本
		UPDATE stock_in_item sii
		SET unit_cost = (
			SELECT COALESCE(po.cost_price, po.produced_unit_cost) / NULLIF(po.produced_quantity, 0)
			FROM production_order po WHERE po.id = v_target_production_order_id
		)
		FROM stock_in si
		JOIN production_order po ON po.stock_in_id = si.id
		WHERE sii.stock_in_id = si.id
		  AND po.id = v_target_production_order_id;

		RETURN;
	END IF;

	-- 以下为旧逻辑：真正创建新的生产单
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

	IF v_mat.is_code THEN
		RAISE EXCEPTION '成品物料为有编码物料，暂不支持自动生产入库（需要生成/选择编码）[%-%]',
			v_mat.material_code, v_mat.material_name;
	END IF;

	-- 计算领料单总成本
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

-- ============================================
-- 8. 更新消费订单视图，包含新字段
-- ============================================
DROP VIEW IF EXISTS v_consumption_order_list;
CREATE VIEW v_consumption_order_list AS
SELECT
	co.id,
	co.order_no,
	co.project_no,
	co.product_name,
	COALESCE(
		(SELECT coi2.warehouse_id FROM consumption_order_item coi2
		 WHERE coi2.order_id = co.id AND coi2.warehouse_id IS NOT NULL
		 ORDER BY coi2.id DESC LIMIT 1),
		(SELECT i.warehouse_id FROM consumption_order_item coi2
		 INNER JOIN inventory i ON i.id = coi2.inventory_id
		 WHERE coi2.order_id = co.id ORDER BY coi2.id LIMIT 1),
		0
	) AS warehouse_id,
	COALESCE(
		(SELECT coi2.warehouse_code FROM consumption_order_item coi2
		 WHERE coi2.order_id = co.id AND coi2.warehouse_code IS NOT NULL
		   AND btrim(coi2.warehouse_code::text) <> ''
		 ORDER BY coi2.id DESC LIMIT 1),
		(SELECT w2.warehouse_code FROM consumption_order_item coi2
		 INNER JOIN inventory i ON i.id = coi2.inventory_id
		 INNER JOIN warehouse w2 ON w2.id = i.warehouse_id AND w2.deleted_at IS NULL
		 WHERE coi2.order_id = co.id ORDER BY coi2.id LIMIT 1),
		''
	) AS warehouse_code,
	COALESCE(
		(SELECT w3.warehouse_name FROM consumption_order_item coi2
		 INNER JOIN warehouse w3 ON w3.id = coi2.warehouse_id AND w3.deleted_at IS NULL
		 WHERE coi2.order_id = co.id AND coi2.warehouse_id IS NOT NULL
		 ORDER BY coi2.id DESC LIMIT 1),
		(SELECT w4.warehouse_name FROM consumption_order_item coi2
		 INNER JOIN inventory i ON i.id = coi2.inventory_id
		 INNER JOIN warehouse w4 ON w4.id = i.warehouse_id AND w4.deleted_at IS NULL
		 WHERE coi2.order_id = co.id ORDER BY coi2.id LIMIT 1),
		''
	) AS warehouse_name,
	co.order_date,
	co.designer_id,
	COALESCE(co.designer_name, '') AS designer_name,
	co.status,
	COALESCE(co.stock_out_id, 0) AS stock_out_id,
	COALESCE(so.stock_out_no, '') AS stock_out_no,
	COALESCE(co.remark, '') AS remark,
	co.created_at,
	co.updated_at,
	COALESCE((SELECT COUNT(1) FROM consumption_order_item coi WHERE coi.order_id = co.id), 0)::integer AS item_count,
	COALESCE((SELECT COALESCE(SUM(coi.quantity), 0) FROM consumption_order_item coi WHERE coi.order_id = co.id), 0) AS total_quantity,
	COALESCE(co.produced_material_id, 0) AS produced_material_id,
	COALESCE(co.produced_warehouse_id, 0) AS produced_warehouse_id,
	COALESCE(co.produced_quantity, 0) AS produced_quantity,
	COALESCE(co.production_order_id, 0) AS production_order_id,
	COALESCE(po.production_no, '') AS production_no,
	COALESCE(co.production_return_order_id, 0) AS production_return_order_id,
	COALESCE(pro.return_no, '') AS production_return_no
FROM consumption_order co
LEFT JOIN stock_out so ON so.id = co.stock_out_id AND so.deleted_at IS NULL
LEFT JOIN production_order po ON po.id = co.production_order_id
LEFT JOIN production_return_order pro ON pro.id = co.production_return_order_id
WHERE co.deleted_at IS NULL;

-- ============================================
-- 9. 更新退料订单视图，包含新字段
-- ============================================
DROP VIEW IF EXISTS v_reversal_order_list;
CREATE VIEW v_reversal_order_list AS
SELECT
	ro.id,
	ro.order_no,
	ro.project_no,
	ro.product_name,
	ro.warehouse_id,
	ro.warehouse_code,
	COALESCE(w.warehouse_name, '') AS warehouse_name,
	ro.order_date,
	ro.designer_id,
	COALESCE(ro.designer_name, '') AS designer_name,
	ro.status,
	COALESCE(ro.stock_in_id, 0) AS stock_in_id,
	COALESCE(si.stock_in_no, '') AS stock_in_no,
	COALESCE(ro.remark, '') AS remark,
	ro.created_at,
	ro.updated_at,
	COALESCE((SELECT COUNT(1) FROM reversal_order_item roi WHERE roi.order_id = ro.id), 0)::integer AS item_count,
	COALESCE((SELECT COALESCE(SUM(roi.quantity), 0) FROM reversal_order_item roi WHERE roi.order_id = ro.id), 0) AS total_quantity,
	COALESCE(ro.production_order_id, 0) AS production_order_id,
	COALESCE(po.production_no, '') AS production_no,
	COALESCE(ro.production_return_order_id, 0) AS production_return_order_id,
	COALESCE(pro.return_no, '') AS production_return_no
FROM reversal_order ro
LEFT JOIN warehouse w ON w.id = ro.warehouse_id AND w.deleted_at IS NULL
LEFT JOIN stock_in si ON si.id = ro.stock_in_id AND si.deleted_at IS NULL
LEFT JOIN production_order po ON po.id = ro.production_order_id
LEFT JOIN production_return_order pro ON pro.id = ro.production_return_order_id
WHERE ro.deleted_at IS NULL;

COMMIT;
