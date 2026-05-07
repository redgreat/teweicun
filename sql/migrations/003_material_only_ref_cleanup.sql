-- 功能：物料单模型改造第三步（清理业务表 SKU 外键字段）
-- 创建时间：2026-05-07
-- 创建人：GPT-5.3-Codex

BEGIN;

UPDATE purchase_order_item
SET custom_attributes = '[]'::jsonb
WHERE custom_attributes IS NULL;

UPDATE stock_in_item
SET custom_attributes = '[]'::jsonb
WHERE custom_attributes IS NULL;

UPDATE inventory
SET custom_attributes = '[]'::jsonb
WHERE custom_attributes IS NULL;

ALTER TABLE purchase_order_item
    DROP COLUMN IF EXISTS sku_id;

ALTER TABLE stock_in_item
    DROP COLUMN IF EXISTS sku_id;

ALTER TABLE return_order_item
    DROP COLUMN IF EXISTS sku_id;

ALTER TABLE inventory
    DROP COLUMN IF EXISTS sku_id,
    DROP COLUMN IF EXISTS sku;

COMMIT;
