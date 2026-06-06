-- ============================================================
-- 功能：为 stock_out_item 表添加 remark 列（sp_confirm_sales_order 需要）
-- 创建时间：2026-06-06
-- 创建人：wangcw
-- MIGRATION_ID: 019_add_remark_to_stock_out_item
-- MIGRATION_APPLIED: pending
-- ============================================================

BEGIN;

ALTER TABLE stock_out_item ADD COLUMN IF NOT EXISTS remark text;
COMMENT ON COLUMN stock_out_item.remark IS '备注';

COMMIT;
