-- 功能：移除物料列表视图中 sku_managed/sku_count 等历史 SKU 字段
-- 创建时间：2026-05-17
-- 创建人：GPT-5.4

BEGIN;

DROP VIEW IF EXISTS v_material_list;

CREATE VIEW v_material_list AS
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
    m.custom_attributes,
    m.default_warehouse_id,
    w.warehouse_name AS default_warehouse_name,
    m.status,
    fn_dict_label('common_status'::character varying, m.status) AS status_name,
    m.remark,
    m.created_at,
    m.updated_at
FROM material m
LEFT JOIN material_category c ON m.category_id = c.id
LEFT JOIN warehouse w ON m.default_warehouse_id = w.id
WHERE m.deleted_at IS NULL;

COMMIT;
