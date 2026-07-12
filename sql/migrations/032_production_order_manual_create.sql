-- MIGRATION_ID: 032_production_order_manual_create
-- MIGRATION_APPLIED: pending
-- 功能：允许手动创建生产单（不依赖领料单自动生成）
--       使 consumption_order_id/stock_out_id 可为空，支持独立成本录入

BEGIN;

-- 移除旧的唯一约束和外键约束
ALTER TABLE production_order DROP CONSTRAINT IF EXISTS production_order_consumption_order_id_fkey;
ALTER TABLE production_order DROP CONSTRAINT IF EXISTS production_order_consumption_order_id_key;
ALTER TABLE production_order DROP CONSTRAINT IF EXISTS production_order_stock_out_id_fkey;

-- 改为可空
ALTER TABLE production_order ALTER COLUMN consumption_order_id DROP NOT NULL;
ALTER TABLE production_order ALTER COLUMN stock_out_id DROP NOT NULL;

-- 重新添加外键（允许 NULL）
ALTER TABLE production_order ADD CONSTRAINT production_order_consumption_order_id_fkey
    FOREIGN KEY (consumption_order_id) REFERENCES consumption_order(id);

ALTER TABLE production_order ADD CONSTRAINT production_order_stock_out_id_fkey
    FOREIGN KEY (stock_out_id) REFERENCES stock_out(id);

-- 为 consumption_order_id 创建部分唯一索引（仅非 NULL 值唯一）
DROP INDEX IF EXISTS idx_production_order_consumption_unique;
CREATE UNIQUE INDEX idx_production_order_consumption_unique
    ON production_order(consumption_order_id) WHERE consumption_order_id IS NOT NULL;

COMMENT ON COLUMN production_order.consumption_order_id IS '来源领料单ID（可空；手动创建的生产单可不关联领料单）';
COMMENT ON COLUMN production_order.stock_out_id IS '关联领料出库单ID（可空；手动创建时可无出库单）';

-- 更新视图支持可空字段
DROP VIEW IF EXISTS v_production_order_list;
CREATE VIEW v_production_order_list AS
SELECT
    po.id, po.production_no, po.status,
    COALESCE(po.consumption_order_id, 0) AS consumption_order_id,
    COALESCE(co.order_no, '') AS consumption_order_no,
    COALESCE(po.stock_out_id, 0) AS stock_out_id,
    COALESCE(so.stock_out_no, '') AS stock_out_no,
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
    COALESCE(po.cost_price, 0)::float8 AS cost_price,
    COALESCE(po.remark, '') AS remark,
    po.created_at, po.updated_at
FROM production_order po
LEFT JOIN consumption_order co ON co.id = po.consumption_order_id
LEFT JOIN stock_out so ON so.id = po.stock_out_id
LEFT JOIN stock_in si ON si.id = po.stock_in_id
LEFT JOIN material m ON m.id = po.produced_material_id
LEFT JOIN warehouse w ON w.id = po.produced_warehouse_id;

COMMIT;
