-- =====================================================
-- 迁移：为调拨单明细表添加库存批次字段
-- 说明：仓库调拨需要指定调出库存批次，
--       sp_confirm_transfer_out 存储过程已引用此字段，
--       但表结构尚未添加，导致调拨出库确认失败。
-- 日期：2026-07-12
-- 作者：Hermes Agent
-- =====================================================

-- 1. 添加 inventory_id 列
ALTER TABLE stock_transfer_item 
ADD COLUMN IF NOT EXISTS inventory_id bigint;

COMMENT ON COLUMN stock_transfer_item.inventory_id IS '调出库存批次ID，关联 inventory.id';

-- 2. 添加外键索引（可选，提升查询性能）
CREATE INDEX IF NOT EXISTS idx_sti_inventory 
ON stock_transfer_item(inventory_id) 
WHERE inventory_id IS NOT NULL;

-- 3. 为已有数据补默认值（无历史数据则跳过）
-- UPDATE stock_transfer_item SET inventory_id = 0 WHERE inventory_id IS NULL;
