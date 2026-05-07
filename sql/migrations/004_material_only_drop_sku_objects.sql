-- 功能：物料单模型改造第四步（删除 SKU 相关对象）
-- 创建时间：2026-05-07
-- 创建人：GPT-5.3-Codex

BEGIN;

DROP VIEW IF EXISTS v_sku_list CASCADE;

DROP FUNCTION IF EXISTS fn_generate_sku_code(bigint) CASCADE;
DROP FUNCTION IF EXISTS fn_generate_sku_code(bigint, jsonb) CASCADE;

DROP TABLE IF EXISTS material_sku CASCADE;

ALTER TABLE material
    DROP COLUMN IF EXISTS sku_managed;

DROP TABLE IF EXISTS _sku_material_mapping;

COMMIT;
