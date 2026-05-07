-- 功能：物料单模型改造第二步（历史 SKU 拆分为独立物料并回填业务表）
-- 创建时间：2026-05-07
-- 创建人：GPT-5.3-Codex

BEGIN;

DROP TABLE IF EXISTS _sku_material_mapping;
CREATE TABLE _sku_material_mapping (
    old_sku_id bigint PRIMARY KEY,
    old_material_id bigint NOT NULL,
    new_material_id bigint NOT NULL
);

WITH sku_rows AS (
    SELECT
        s.id AS old_sku_id,
        s.material_id AS old_material_id,
        m.category_id,
        m.material_grade,
        m.standard_no,
        m.unit,
        m.aux_unit,
        m.conversion_factor,
        m.is_code,
        m.safety_stock,
        m.max_stock,
        m.default_warehouse_id,
        m.status,
        m.created_by AS material_created_by,
        m.updated_by AS material_updated_by,
        m.deleted_at AS material_deleted_at,
        s.created_by AS sku_created_by,
        s.updated_by AS sku_updated_by,
        s.created_at AS sku_created_at,
        s.updated_at AS sku_updated_at,
        s.deleted_at AS sku_deleted_at,
        COALESCE(s.reference_price, 0) AS reference_price,
        COALESCE(s.custom_attributes, '[]'::jsonb) AS custom_attributes,
        m.material_name AS base_material_name,
        COALESCE(NULLIF(BTRIM(s.sku_name), ''), '') AS sku_name,
        COALESCE(s.remark, m.remark) AS merged_remark
    FROM material_sku s
    JOIN material m ON m.id = s.material_id
),
prepared AS (
    SELECT
        sr.*,
        (
            SELECT STRING_AGG(
                BTRIM(
                    COALESCE(attr->>'attr_name', '') || ':' ||
                    COALESCE(attr->>'attr_value', '') ||
                    COALESCE(attr->>'attr_unit', '')
                ),
                ' / '
            )
            FROM jsonb_array_elements(sr.custom_attributes) attr
            WHERE COALESCE(BTRIM(attr->>'attr_value'), '') <> ''
        ) AS attr_text
    FROM sku_rows sr
),
inserted AS (
    INSERT INTO material (
        category_id,
        material_code,
        material_name,
        material_grade,
        standard_no,
        unit,
        aux_unit,
        conversion_factor,
        is_code,
        safety_stock,
        max_stock,
        default_warehouse_id,
        remark,
        status,
        created_by,
        created_at,
        updated_by,
        updated_at,
        deleted_at,
        custom_attributes,
        sku_managed,
        reference_price
    )
    SELECT
        p.category_id,
        fn_generate_material_code(),
        CASE
            WHEN p.sku_name <> '' THEN p.sku_name
            WHEN COALESCE(p.attr_text, '') <> '' THEN p.base_material_name || '（' || p.attr_text || '）'
            ELSE p.base_material_name
        END AS material_name,
        p.material_grade,
        p.standard_no,
        p.unit,
        p.aux_unit,
        p.conversion_factor,
        p.is_code,
        p.safety_stock,
        p.max_stock,
        p.default_warehouse_id,
        p.merged_remark,
        p.status,
        COALESCE(p.sku_created_by, p.material_created_by),
        COALESCE(p.sku_created_at, NOW()),
        COALESCE(p.sku_updated_by, p.material_updated_by),
        COALESCE(p.sku_updated_at, NOW()),
        p.sku_deleted_at,
        p.custom_attributes,
        false,
        p.reference_price
    FROM prepared p
    RETURNING id, material_name
),
numbered_inserted AS (
    SELECT id, ROW_NUMBER() OVER (ORDER BY id) AS rn
    FROM inserted
),
numbered_source AS (
    SELECT old_sku_id, old_material_id, ROW_NUMBER() OVER (ORDER BY old_sku_id) AS rn
    FROM sku_rows
)
INSERT INTO _sku_material_mapping (old_sku_id, old_material_id, new_material_id)
SELECT ns.old_sku_id, ns.old_material_id, ni.id
FROM numbered_source ns
JOIN numbered_inserted ni ON ni.rn = ns.rn;

UPDATE purchase_order_item poi
SET material_id = mp.new_material_id,
    custom_attributes = COALESCE(ms.custom_attributes, poi.custom_attributes, '[]'::jsonb)
FROM _sku_material_mapping mp
LEFT JOIN material_sku ms ON ms.id = mp.old_sku_id
WHERE poi.sku_id = mp.old_sku_id;

UPDATE stock_in_item sii
SET material_id = mp.new_material_id,
    custom_attributes = COALESCE(ms.custom_attributes, sii.custom_attributes, '[]'::jsonb)
FROM _sku_material_mapping mp
LEFT JOIN material_sku ms ON ms.id = mp.old_sku_id
WHERE sii.sku_id = mp.old_sku_id;

UPDATE return_order_item roi
SET material_id = mp.new_material_id
FROM _sku_material_mapping mp
WHERE roi.sku_id = mp.old_sku_id;

UPDATE inventory i
SET material_id = mp.new_material_id,
    custom_attributes = COALESCE(ms.custom_attributes, i.custom_attributes, '[]'::jsonb),
    sku = NULL
FROM _sku_material_mapping mp
LEFT JOIN material_sku ms ON ms.id = mp.old_sku_id
WHERE i.sku_id = mp.old_sku_id;

COMMIT;
