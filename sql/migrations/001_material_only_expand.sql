-- 功能：物料单模型改造第一步（扩展 material 字段，准备去 SKU）
-- 创建时间：2026-05-07
-- 创建人：GPT-5.3-Codex
-- MIGRATION_ID: 001_material_only_expand
-- MIGRATION_APPLIED: applied_20260510T110419

BEGIN;

ALTER TABLE material
    ADD COLUMN IF NOT EXISTS reference_price numeric(18,2) NOT NULL DEFAULT 0;

COMMENT ON COLUMN material.reference_price IS '参考价格（元），由原 SKU 价格归并到物料';

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'material'
          AND column_name = 'sku_managed'
    ) THEN
        UPDATE material
        SET sku_managed = false
        WHERE sku_managed IS DISTINCT FROM false;
    END IF;
END $$;

COMMIT;
