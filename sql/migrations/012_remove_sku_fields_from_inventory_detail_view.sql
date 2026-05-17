-- 功能：视图层最终收口：移除 v_inventory_detail 中历史 sku_* 字段
-- 创建时间：2026-05-17
-- 创建人：GPT-5.4

BEGIN;

DROP VIEW IF EXISTS v_inventory_detail;

CREATE VIEW v_inventory_detail AS
SELECT
    i.id,
    i.material_id,
    m.material_code,
    m.material_name,
    m.is_code,
    m.unit,
    i.warehouse_id,
    w.warehouse_code,
    w.warehouse_name,
    i.quantity,
    i.locked_quantity,
    i.quantity - i.locked_quantity AS available_quantity,
    i.unit_cost,
    i.stock_in_date,
    i.material_inspection_no,
    i.updated_at AS created_at,
    i.updated_at,
    NULL::bigint AS certificate_id,
    ''::character varying AS certificate_no
FROM inventory i
JOIN material m ON i.material_id = m.id
JOIN warehouse w ON i.warehouse_id = w.id
WHERE m.deleted_at IS NULL;

COMMIT;

