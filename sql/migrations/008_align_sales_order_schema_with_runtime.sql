-- 功能：对齐销售订单运行时表结构与当前创建流程
-- 创建时间：2026-05-17
-- 创建人：GPT-5.4

BEGIN;

ALTER TABLE sales_order
    ADD COLUMN IF NOT EXISTS customer_name varchar(200),
    ADD COLUMN IF NOT EXISTS contract_no varchar(50),
    ADD COLUMN IF NOT EXISTS payment_method varchar(50),
    ADD COLUMN IF NOT EXISTS receiver_name varchar(100),
    ADD COLUMN IF NOT EXISTS receiver_phone varchar(20),
    ADD COLUMN IF NOT EXISTS receiver_address varchar(500);

ALTER TABLE sales_order
    ALTER COLUMN sales_person_id DROP NOT NULL;

COMMENT ON COLUMN sales_order.sales_person_id IS '销售员ID，可为空';
COMMENT ON COLUMN sales_order.customer_name IS '客户名称快照';
COMMENT ON COLUMN sales_order.contract_no IS '合同编号';
COMMENT ON COLUMN sales_order.payment_method IS '付款方式';
COMMENT ON COLUMN sales_order.receiver_name IS '收货联系人';
COMMENT ON COLUMN sales_order.receiver_phone IS '收货联系电话';
COMMENT ON COLUMN sales_order.receiver_address IS '收货地址';

ALTER TABLE sales_order_item
    ADD COLUMN IF NOT EXISTS material_code varchar(30),
    ADD COLUMN IF NOT EXISTS material_name varchar(200),
    ADD COLUMN IF NOT EXISTS unit varchar(20);

COMMENT ON COLUMN sales_order_item.material_code IS '物料编码快照';
COMMENT ON COLUMN sales_order_item.material_name IS '物料名称快照';
COMMENT ON COLUMN sales_order_item.unit IS '单位快照';

COMMIT;
