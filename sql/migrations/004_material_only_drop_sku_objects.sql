-- 功能：物料单模型改造第四步（删除 SKU 相关对象）
-- 创建时间：2026-05-07
-- 创建人：GPT-5.3-Codex
-- MIGRATION_ID: 004_material_only_drop_sku_objects
-- MIGRATION_APPLIED: applied_20260510T110435

BEGIN;

DROP VIEW IF EXISTS v_sku_list CASCADE;

DROP FUNCTION IF EXISTS fn_generate_sku_code(bigint) CASCADE;
DROP FUNCTION IF EXISTS fn_generate_sku_code(bigint, jsonb) CASCADE;

DROP TABLE IF EXISTS material_sku CASCADE;

ALTER TABLE material
    DROP COLUMN IF EXISTS sku_managed;

DROP TABLE IF EXISTS _sku_material_mapping;

CREATE OR REPLACE VIEW v_material_list AS
SELECT
    m.id,
    m.category_id,
    c.category_name,
    m.material_code,
    m.material_name,
    m.unit,
    fn_dict_label('unit'::character varying, m.unit) AS unit_name,
    m.safety_stock,
    m.max_stock,
    m.is_code,
    false::boolean AS sku_managed,
    m.custom_attributes,
    m.default_warehouse_id,
    w.warehouse_name AS default_warehouse_name,
    m.status,
    fn_dict_label('common_status'::character varying, m.status) AS status_name,
    m.remark,
    m.created_at,
    m.updated_at,
    0::bigint AS sku_count
FROM material m
LEFT JOIN material_category c ON m.category_id = c.id
LEFT JOIN warehouse w ON m.default_warehouse_id = w.id
WHERE m.deleted_at IS NULL;

CREATE OR REPLACE VIEW v_inventory_detail AS
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
    0::bigint AS sku_id,
    ''::character varying AS sku_code,
    ''::character varying AS sku_name,
    i.quantity,
    i.locked_quantity,
    (i.quantity - i.locked_quantity) AS available_quantity,
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
