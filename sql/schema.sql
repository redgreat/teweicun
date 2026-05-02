-- =====================================================
-- 特维存（TeWeiCun）数据库结构导出
-- 导出时间: 2026-04-29 06:21:56
-- PostgreSQL 数据库
-- =====================================================

-- =====================================================
-- 表结构定义
-- =====================================================

-- 表: consumption_order
-- 创建时间: 2026-04-29
CREATE TABLE consumption_order (
    id bigint NOT NULL DEFAULT nextval('consumption_order_id_seq'::regclass),
    order_no character varying(20) NOT NULL,
    project_no character varying(50) NOT NULL,
    product_name character varying(200) NOT NULL,
    order_date date NOT NULL DEFAULT CURRENT_DATE,
    designer_id bigint NOT NULL,
    designer_name character varying(50) NOT NULL,
    status character varying(20) NOT NULL DEFAULT 'draft'::character varying,
    stock_out_id bigint,
    remark text,
    created_by bigint,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_by bigint,
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone
);

COMMENT ON TABLE consumption_order IS '领料订单；仓库由明细库存/确认时写入的明细行仓库体现）';

COMMENT ON COLUMN consumption_order.order_no IS '领料单号，规则：CO + 年月日 + 序号';
COMMENT ON COLUMN consumption_order.project_no IS '项目/图纸编号';
COMMENT ON COLUMN consumption_order.product_name IS '产品名称（压力容器名称）';
COMMENT ON COLUMN consumption_order.order_date IS '领料日期';
COMMENT ON COLUMN consumption_order.designer_id IS '设计员ID';
COMMENT ON COLUMN consumption_order.designer_name IS '设计员姓名';
COMMENT ON COLUMN consumption_order.status IS '状态：草稿/待确认/已确认/已完成/已取消';
COMMENT ON COLUMN consumption_order.stock_out_id IS '关联的出库单ID';
COMMENT ON COLUMN consumption_order.remark IS '备注';

CREATE UNIQUE INDEX consumption_order_pkey ON public.consumption_order USING btree (id);
CREATE UNIQUE INDEX consumption_order_order_no_key ON public.consumption_order USING btree (order_no);
CREATE INDEX idx_co_no ON public.consumption_order USING btree (order_no) WHERE (deleted_at IS NULL);
CREATE INDEX idx_co_project ON public.consumption_order USING btree (project_no) WHERE (deleted_at IS NULL);
CREATE INDEX idx_co_status ON public.consumption_order USING btree (status) WHERE (deleted_at IS NULL);
CREATE INDEX idx_co_date ON public.consumption_order USING btree (order_date) WHERE (deleted_at IS NULL);
CREATE INDEX idx_co_designer ON public.consumption_order USING btree (designer_id) WHERE (deleted_at IS NULL);

-- 表: consumption_order_item
-- 创建时间: 2026-04-29
CREATE TABLE consumption_order_item (
    id bigint NOT NULL DEFAULT nextval('consumption_order_item_id_seq'::regclass),
    order_id bigint NOT NULL,
    material_id bigint NOT NULL,
    inventory_id bigint,
    quantity numeric(18,3) NOT NULL,
    unit character varying(20) NOT NULL,
    remark text,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    warehouse_id bigint,
    warehouse_code character varying(10)
);

COMMENT ON TABLE consumption_order_item IS '领料订单明细';

COMMENT ON COLUMN consumption_order_item.order_id IS '领料单ID';
COMMENT ON COLUMN consumption_order_item.material_id IS '物料ID';
COMMENT ON COLUMN consumption_order_item.inventory_id IS '库存批次ID';
COMMENT ON COLUMN consumption_order_item.quantity IS '领料数量';
COMMENT ON COLUMN consumption_order_item.unit IS '单位';
COMMENT ON COLUMN consumption_order_item.remark IS '备注';
COMMENT ON COLUMN consumption_order_item.warehouse_id IS '该明细行出库仓库（按所选库存批次解析）';
COMMENT ON COLUMN consumption_order_item.warehouse_code IS '该明细行出库仓库编码（按所选库存批次写入）';

CREATE UNIQUE INDEX consumption_order_item_pkey ON public.consumption_order_item USING btree (id);
CREATE INDEX idx_coi_order ON public.consumption_order_item USING btree (order_id);
CREATE INDEX idx_coi_material ON public.consumption_order_item USING btree (material_id);
CREATE INDEX idx_coi_inventory ON public.consumption_order_item USING btree (inventory_id);

-- 表: customer
-- 创建时间: 2026-04-29
CREATE TABLE customer (
    id bigint NOT NULL DEFAULT nextval('customer_id_seq'::regclass),
    customer_code character varying(20) NOT NULL,
    customer_name character varying(200) NOT NULL,
    credit_code character varying(18),
    customer_type character varying(50),
    contact_person character varying(50) NOT NULL,
    contact_phone character varying(20) NOT NULL,
    address character varying(500),
    sales_person_id bigint,
    remark text,
    status character varying(20) NOT NULL DEFAULT 'enabled'::character varying,
    created_by bigint,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    updated_by bigint,
    updated_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp(6) with time zone
);

COMMENT ON TABLE customer IS '客户主数据表';

COMMENT ON COLUMN customer.id IS '客户ID，主键';
COMMENT ON COLUMN customer.customer_code IS '客户编码，唯一';
COMMENT ON COLUMN customer.customer_name IS '客户名称';
COMMENT ON COLUMN customer.credit_code IS '统一社会信用代码';
COMMENT ON COLUMN customer.customer_type IS '客户类型: direct_user/trader/engineering/other';
COMMENT ON COLUMN customer.contact_person IS '联系人';
COMMENT ON COLUMN customer.contact_phone IS '联系电话';
COMMENT ON COLUMN customer.address IS '地址';
COMMENT ON COLUMN customer.sales_person_id IS '负责销售员ID';
COMMENT ON COLUMN customer.remark IS '备注';
COMMENT ON COLUMN customer.status IS '状态: enabled-启用, disabled-禁用';
COMMENT ON COLUMN customer.created_by IS '创建人ID';
COMMENT ON COLUMN customer.created_at IS '创建时间';
COMMENT ON COLUMN customer.updated_by IS '更新人ID';
COMMENT ON COLUMN customer.updated_at IS '更新时间';
COMMENT ON COLUMN customer.deleted_at IS '删除时间（软删除标记）';

CREATE INDEX idx_customer_name ON public.customer USING btree (customer_name) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX uk_customer_code ON public.customer USING btree (customer_code) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX customer_pkey ON public.customer USING btree (id);

-- 表: inventory
-- 创建时间: 2026-04-29
CREATE TABLE inventory (
    id bigint NOT NULL DEFAULT nextval('inventory_id_seq'::regclass),
    material_id bigint NOT NULL,
    warehouse_id bigint NOT NULL,
    quantity numeric(18,3) NOT NULL DEFAULT 0,
    locked_quantity numeric(18,3) NOT NULL DEFAULT 0,
    unit character varying(20) NOT NULL,
    unit_cost numeric(18,2) DEFAULT 0,
    stock_in_date date,
    expire_date date,
    updated_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    material_inspection_no character varying(50),
    sku character varying(100),
    sku_id bigint,
    custom_attributes jsonb,
    in_transit_quantity numeric(18,3) NOT NULL DEFAULT 0
);

COMMENT ON TABLE inventory IS '库存台账表（核心表）';

COMMENT ON COLUMN inventory.id IS '库存ID，主键';
COMMENT ON COLUMN inventory.material_id IS '物料ID';
COMMENT ON COLUMN inventory.warehouse_id IS '仓库ID';
COMMENT ON COLUMN inventory.quantity IS '库存数量（不得为负）';
COMMENT ON COLUMN inventory.locked_quantity IS '锁定数量（销售订单锁定）';
COMMENT ON COLUMN inventory.unit IS '计量单位';
COMMENT ON COLUMN inventory.unit_cost IS '单位成本';
COMMENT ON COLUMN inventory.stock_in_date IS '入库日期';
COMMENT ON COLUMN inventory.expire_date IS '有效期';
COMMENT ON COLUMN inventory.updated_at IS '更新时间';
COMMENT ON COLUMN inventory.material_inspection_no IS '材料检号';
COMMENT ON COLUMN inventory.sku IS 'SKU编码(物料+自定义属性唯一标识)';
COMMENT ON COLUMN inventory.sku_id IS 'SKU ID';
COMMENT ON COLUMN inventory.custom_attributes IS '自定义属性值JSON';

CREATE INDEX idx_inventory_expire ON public.inventory USING btree (expire_date) WHERE (expire_date IS NOT NULL);
CREATE INDEX idx_inventory_material_wh ON public.inventory USING btree (material_id, warehouse_id);
CREATE INDEX idx_inventory_stock_date ON public.inventory USING btree (stock_in_date);
CREATE UNIQUE INDEX inventory_pkey ON public.inventory USING btree (id);
CREATE UNIQUE INDEX uk_inventory_key ON public.inventory USING btree (material_id, warehouse_id, COALESCE(material_inspection_no, ''::character varying));

-- 表: inventory_check
-- 创建时间: 2026-04-29
CREATE TABLE inventory_check (
    id bigint NOT NULL DEFAULT nextval('inventory_check_id_seq'::regclass),
    check_no character varying(20) NOT NULL,
    warehouse_id bigint NOT NULL,
    check_date date NOT NULL,
    check_status character varying(20) NOT NULL DEFAULT 'draft'::character varying,
    checker_id bigint NOT NULL,
    approved_by bigint,
    approved_at timestamp(6) with time zone,
    remark text,
    created_by bigint,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    updated_by bigint,
    updated_at timestamp(6) with time zone NOT NULL DEFAULT now()
);

COMMENT ON TABLE inventory_check IS '库存盘点单';

COMMENT ON COLUMN inventory_check.id IS '盘点单ID，主键';
COMMENT ON COLUMN inventory_check.check_no IS '盘点单号，唯一';
COMMENT ON COLUMN inventory_check.warehouse_id IS '盘点仓库ID';
COMMENT ON COLUMN inventory_check.check_date IS '盘点日期';
COMMENT ON COLUMN inventory_check.check_status IS '盘点状态: draft/counting/completed/approved';
COMMENT ON COLUMN inventory_check.checker_id IS '盘点人ID';
COMMENT ON COLUMN inventory_check.approved_by IS '审核人ID';
COMMENT ON COLUMN inventory_check.approved_at IS '审核时间';
COMMENT ON COLUMN inventory_check.remark IS '备注';
COMMENT ON COLUMN inventory_check.created_by IS '创建人ID';
COMMENT ON COLUMN inventory_check.created_at IS '创建时间';
COMMENT ON COLUMN inventory_check.updated_by IS '更新人ID';
COMMENT ON COLUMN inventory_check.updated_at IS '更新时间';

CREATE UNIQUE INDEX uk_check_no ON public.inventory_check USING btree (check_no);
CREATE UNIQUE INDEX inventory_check_pkey ON public.inventory_check USING btree (id);

-- 表: inventory_check_item
-- 创建时间: 2026-04-29
CREATE TABLE inventory_check_item (
    id bigint NOT NULL DEFAULT nextval('inventory_check_item_id_seq'::regclass),
    check_id bigint NOT NULL,
    material_id bigint NOT NULL,
    book_quantity numeric(18,3) NOT NULL,
    actual_quantity numeric(18,3),
    diff_quantity numeric(18,3),
    diff_reason text,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    updated_at timestamp(6) with time zone NOT NULL DEFAULT now()
);

COMMENT ON TABLE inventory_check_item IS '盘点明细表';

COMMENT ON COLUMN inventory_check_item.id IS '明细ID，主键';
COMMENT ON COLUMN inventory_check_item.check_id IS '关联盘点单ID';
COMMENT ON COLUMN inventory_check_item.material_id IS '物料ID';
COMMENT ON COLUMN inventory_check_item.book_quantity IS '账面数量';
COMMENT ON COLUMN inventory_check_item.actual_quantity IS '实盘数量';
COMMENT ON COLUMN inventory_check_item.diff_quantity IS '差异数量（实盘-账面）';
COMMENT ON COLUMN inventory_check_item.diff_reason IS '差异原因';
COMMENT ON COLUMN inventory_check_item.created_at IS '创建时间';
COMMENT ON COLUMN inventory_check_item.updated_at IS '更新时间';

CREATE INDEX idx_ici_check ON public.inventory_check_item USING btree (check_id);
CREATE UNIQUE INDEX inventory_check_item_pkey ON public.inventory_check_item USING btree (id);

-- 表: inventory_transaction
-- 创建时间: 2026-04-29
CREATE TABLE inventory_transaction (
    id bigint NOT NULL DEFAULT nextval('inventory_transaction_id_seq'::regclass),
    material_id bigint NOT NULL,
    warehouse_id bigint NOT NULL,
    trans_type character varying(20) NOT NULL,
    quantity numeric(18,3) NOT NULL,
    balance numeric(18,3),
    ref_doc_type character varying(50),
    ref_doc_no character varying(20),
    ref_doc_id bigint,
    operator_id bigint,
    remark text,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now()
);

COMMENT ON TABLE inventory_transaction IS '库存流水表（记录每次库存变动）';

COMMENT ON COLUMN inventory_transaction.id IS '流水ID，主键';
COMMENT ON COLUMN inventory_transaction.material_id IS '物料ID';
COMMENT ON COLUMN inventory_transaction.warehouse_id IS '仓库ID';
COMMENT ON COLUMN inventory_transaction.trans_type IS '流水类型: in-入库, out-出库, adjust-调整, lock-锁定, unlock-解锁';
COMMENT ON COLUMN inventory_transaction.quantity IS '变动数量（入库为正，出库为负）';
COMMENT ON COLUMN inventory_transaction.balance IS '变动后结余';
COMMENT ON COLUMN inventory_transaction.ref_doc_type IS '关联单据类型';
COMMENT ON COLUMN inventory_transaction.ref_doc_no IS '关联单据号';
COMMENT ON COLUMN inventory_transaction.ref_doc_id IS '关联单据ID';
COMMENT ON COLUMN inventory_transaction.operator_id IS '操作人ID';
COMMENT ON COLUMN inventory_transaction.remark IS '备注';
COMMENT ON COLUMN inventory_transaction.created_at IS '创建时间';

CREATE INDEX idx_inv_trans_date ON public.inventory_transaction USING btree (created_at);
CREATE INDEX idx_inv_trans_material_date ON public.inventory_transaction USING btree (material_id, created_at);
CREATE INDEX idx_inv_trans_ref ON public.inventory_transaction USING btree (ref_doc_type, ref_doc_no);
CREATE UNIQUE INDEX inventory_transaction_pkey ON public.inventory_transaction USING btree (id);

-- 表: material
-- 创建时间: 2026-04-29
CREATE TABLE material (
    id bigint NOT NULL DEFAULT nextval('material_id_seq'::regclass),
    material_code character varying(30) NOT NULL,
    material_name character varying(100) NOT NULL,
    category_id bigint NOT NULL,
    material_grade character varying(50),
    standard_no character varying(50),
    unit character varying(20) NOT NULL,
    aux_unit character varying(20),
    conversion_factor numeric(18,6),
    is_code boolean NOT NULL DEFAULT false,
    safety_stock numeric(18,3) DEFAULT 0,
    max_stock numeric(18,3) DEFAULT 0,
    default_warehouse_id bigint,
    remark text,
    status character varying(20) NOT NULL DEFAULT 'enabled'::character varying,
    created_by bigint,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    updated_by bigint,
    updated_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp(6) with time zone,
    custom_attributes jsonb DEFAULT '[]'::jsonb,
    sku_managed boolean NOT NULL DEFAULT false
);

COMMENT ON TABLE material IS '物料主数据表';

COMMENT ON COLUMN material.id IS '物料ID，主键';
COMMENT ON COLUMN material.material_code IS '物料编码，全局唯一';
COMMENT ON COLUMN material.material_name IS '物料名称';
COMMENT ON COLUMN material.category_id IS '物料分类ID';
COMMENT ON COLUMN material.material_grade IS '材质/材料牌号';
COMMENT ON COLUMN material.standard_no IS '材料标准号';
COMMENT ON COLUMN material.unit IS '主计量单位';
COMMENT ON COLUMN material.aux_unit IS '辅助计量单位';
COMMENT ON COLUMN material.conversion_factor IS '主辅单位换算系数';
COMMENT ON COLUMN material.is_code IS '是否独立编码(入库时是否生成材料检号)';
COMMENT ON COLUMN material.safety_stock IS '安全库存';
COMMENT ON COLUMN material.max_stock IS '最高库存';
COMMENT ON COLUMN material.default_warehouse_id IS '默认仓库ID';
COMMENT ON COLUMN material.remark IS '备注';
COMMENT ON COLUMN material.status IS '状态: enabled-启用, disabled-禁用';
COMMENT ON COLUMN material.created_by IS '创建人ID';
COMMENT ON COLUMN material.created_at IS '创建时间';
COMMENT ON COLUMN material.updated_by IS '更新人ID';
COMMENT ON COLUMN material.updated_at IS '更新时间';
COMMENT ON COLUMN material.deleted_at IS '删除时间（软删除标记）';
COMMENT ON COLUMN material.custom_attributes IS '自定义属性JSON数组';
COMMENT ON COLUMN material.sku_managed IS '是否启用SKU管理（启用后采购/入库可选SKU）';

CREATE INDEX idx_material_category ON public.material USING btree (category_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_material_grade ON public.material USING btree (material_grade) WHERE (deleted_at IS NULL);
CREATE INDEX idx_material_name ON public.material USING btree (material_name) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX uk_material_code ON public.material USING btree (material_code) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX material_pkey ON public.material USING btree (id);

-- 表: material_attribute_def
-- 创建时间: 2026-04-29
CREATE TABLE material_attribute_def (
    id bigint NOT NULL DEFAULT nextval('material_attribute_def_id_seq'::regclass),
    attr_code character varying(50) NOT NULL,
    attr_name character varying(100) NOT NULL,
    attr_type character varying(20) NOT NULL DEFAULT 'text'::character varying,
    attr_unit character varying(20),
    is_required boolean NOT NULL DEFAULT false,
    sort_order integer DEFAULT 0,
    remark text,
    status character varying(20) NOT NULL DEFAULT 'enabled'::character varying,
    created_by bigint,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_by bigint,
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone,
    select_options text
);

COMMENT ON TABLE material_attribute_def IS '物料属性定义表';

COMMENT ON COLUMN material_attribute_def.id IS '属性ID，主键';
COMMENT ON COLUMN material_attribute_def.attr_code IS '属性编码，唯一';
COMMENT ON COLUMN material_attribute_def.attr_name IS '属性名称';
COMMENT ON COLUMN material_attribute_def.attr_type IS '属性类型: text-文本, number-数字, select-下拉选择, date-日期';
COMMENT ON COLUMN material_attribute_def.attr_unit IS '属性单位';
COMMENT ON COLUMN material_attribute_def.is_required IS '是否必填';
COMMENT ON COLUMN material_attribute_def.sort_order IS '排序号';
COMMENT ON COLUMN material_attribute_def.remark IS '备注';
COMMENT ON COLUMN material_attribute_def.status IS '状态: enabled-启用, disabled-禁用';
COMMENT ON COLUMN material_attribute_def.created_by IS '创建人ID';
COMMENT ON COLUMN material_attribute_def.created_at IS '创建时间';
COMMENT ON COLUMN material_attribute_def.updated_by IS '更新人ID';
COMMENT ON COLUMN material_attribute_def.updated_at IS '更新时间';
COMMENT ON COLUMN material_attribute_def.deleted_at IS '删除时间（软删除标记）';
COMMENT ON COLUMN material_attribute_def.select_options IS '下拉选项值（每行一个选项，仅当attr_type=select时使用）';

CREATE UNIQUE INDEX material_attribute_def_pkey ON public.material_attribute_def USING btree (id);
CREATE UNIQUE INDEX uk_mat_attr_def_code ON public.material_attribute_def USING btree (attr_code) WHERE (deleted_at IS NULL);
CREATE INDEX idx_mat_attr_def_status ON public.material_attribute_def USING btree (status) WHERE (deleted_at IS NULL);

-- 表: material_attribute_value
-- 创建时间: 2026-04-29
CREATE TABLE material_attribute_value (
    id bigint NOT NULL DEFAULT nextval('material_attribute_value_id_seq'::regclass),
    material_id bigint NOT NULL,
    attr_id bigint NOT NULL,
    attr_value text NOT NULL,
    created_by bigint,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_by bigint,
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone
);

COMMENT ON TABLE material_attribute_value IS '物料属性值表';

COMMENT ON COLUMN material_attribute_value.id IS 'ID，主键';
COMMENT ON COLUMN material_attribute_value.material_id IS '物料ID';
COMMENT ON COLUMN material_attribute_value.attr_id IS '属性ID';
COMMENT ON COLUMN material_attribute_value.attr_value IS '属性值';
COMMENT ON COLUMN material_attribute_value.created_by IS '创建人ID';
COMMENT ON COLUMN material_attribute_value.created_at IS '创建时间';
COMMENT ON COLUMN material_attribute_value.updated_by IS '更新人ID';
COMMENT ON COLUMN material_attribute_value.updated_at IS '更新时间';
COMMENT ON COLUMN material_attribute_value.deleted_at IS '删除时间（软删除标记）';

CREATE UNIQUE INDEX material_attribute_value_pkey ON public.material_attribute_value USING btree (id);
CREATE UNIQUE INDEX uk_mat_attr_value ON public.material_attribute_value USING btree (material_id, attr_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_mat_attr_value_material ON public.material_attribute_value USING btree (material_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_mat_attr_value_attr ON public.material_attribute_value USING btree (attr_id) WHERE (deleted_at IS NULL);

-- 表: material_category
-- 创建时间: 2026-04-29
CREATE TABLE material_category (
    id bigint NOT NULL DEFAULT nextval('material_category_id_seq'::regclass),
    parent_id bigint DEFAULT 0,
    category_code character varying(10) NOT NULL,
    category_name character varying(100) NOT NULL,
    sort_order integer DEFAULT 0,
    status character varying(20) NOT NULL DEFAULT 'enabled'::character varying,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    updated_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp(6) with time zone
);

COMMENT ON TABLE material_category IS '物料分类表';

COMMENT ON COLUMN material_category.id IS '分类ID，主键';
COMMENT ON COLUMN material_category.parent_id IS '父分类ID，0为顶级分类';
COMMENT ON COLUMN material_category.category_code IS '分类编码，唯一';
COMMENT ON COLUMN material_category.category_name IS '分类名称';
COMMENT ON COLUMN material_category.sort_order IS '排序号';
COMMENT ON COLUMN material_category.status IS '状态: enabled-启用, disabled-禁用';
COMMENT ON COLUMN material_category.created_at IS '创建时间';
COMMENT ON COLUMN material_category.updated_at IS '更新时间';
COMMENT ON COLUMN material_category.deleted_at IS '删除时间（软删除标记）';

CREATE UNIQUE INDEX uk_mat_cat_code ON public.material_category USING btree (category_code) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX material_category_pkey ON public.material_category USING btree (id);

-- 表: material_certificate
-- 创建时间: 2026-04-29
CREATE TABLE material_certificate (
    id bigint NOT NULL DEFAULT nextval('material_certificate_id_seq'::regclass),
    cert_no character varying(30) NOT NULL,
    supplier_id bigint,
    material_id bigint NOT NULL,
    material_grade character varying(50),
    standard_no character varying(50),
    chemical_composition jsonb,
    mechanical_properties jsonb,
    inspection_org character varying(200),
    inspection_date date,
    file_path character varying(500),
    confirmed_by bigint,
    confirmed_at timestamp(6) with time zone,
    remark text,
    created_by bigint,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    updated_by bigint,
    updated_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp(6) with time zone,
    supplier_code character varying(20)
);

COMMENT ON TABLE material_certificate IS '材质证明书表';

COMMENT ON COLUMN material_certificate.id IS '证书ID，主键';
COMMENT ON COLUMN material_certificate.cert_no IS '证书编号，唯一';
COMMENT ON COLUMN material_certificate.supplier_id IS '供应商ID';
COMMENT ON COLUMN material_certificate.material_id IS '物料ID';
COMMENT ON COLUMN material_certificate.material_grade IS '材质牌号';
COMMENT ON COLUMN material_certificate.standard_no IS '材料标准号';
COMMENT ON COLUMN material_certificate.chemical_composition IS '化学成分（JSON格式）';
COMMENT ON COLUMN material_certificate.mechanical_properties IS '力学性能（JSON格式）';
COMMENT ON COLUMN material_certificate.inspection_org IS '检验机构';
COMMENT ON COLUMN material_certificate.inspection_date IS '检验日期';
COMMENT ON COLUMN material_certificate.file_path IS 'PDF扫描件路径';
COMMENT ON COLUMN material_certificate.confirmed_by IS '确认人ID';
COMMENT ON COLUMN material_certificate.confirmed_at IS '确认时间';
COMMENT ON COLUMN material_certificate.remark IS '备注';
COMMENT ON COLUMN material_certificate.created_by IS '创建人ID';
COMMENT ON COLUMN material_certificate.created_at IS '创建时间';
COMMENT ON COLUMN material_certificate.updated_by IS '更新人ID';
COMMENT ON COLUMN material_certificate.updated_at IS '更新时间';
COMMENT ON COLUMN material_certificate.deleted_at IS '删除时间（软删除标记）';

CREATE INDEX idx_mc_chemical ON public.material_certificate USING gin (chemical_composition);
CREATE INDEX idx_mc_material ON public.material_certificate USING btree (material_id);
CREATE UNIQUE INDEX uk_mc_cert_no ON public.material_certificate USING btree (cert_no) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX material_certificate_pkey ON public.material_certificate USING btree (id);

-- 表: material_sku
-- 创建时间: 2026-04-29
CREATE TABLE material_sku (
    id bigint NOT NULL DEFAULT nextval('material_sku_id_seq'::regclass),
    material_id bigint NOT NULL,
    sku_code character varying(50) NOT NULL,
    sku_name character varying(200),
    custom_attributes jsonb NOT NULL DEFAULT '[]'::jsonb,
    status character varying(20) NOT NULL DEFAULT 'enabled'::character varying,
    remark text,
    created_by bigint,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_by bigint,
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone,
    reference_price numeric(18,2) NOT NULL DEFAULT 0
);

COMMENT ON TABLE material_sku IS '物料SKU表';

COMMENT ON COLUMN material_sku.id IS '主键ID';
COMMENT ON COLUMN material_sku.material_id IS '物料ID';
COMMENT ON COLUMN material_sku.sku_code IS 'SKU编码';
COMMENT ON COLUMN material_sku.sku_name IS 'SKU名称';
COMMENT ON COLUMN material_sku.custom_attributes IS '自定义属性值JSON';
COMMENT ON COLUMN material_sku.status IS '状态:enabled-启用,disabled-禁用';
COMMENT ON COLUMN material_sku.remark IS '备注';
COMMENT ON COLUMN material_sku.reference_price IS '参考价格（元）';

CREATE UNIQUE INDEX material_sku_pkey ON public.material_sku USING btree (id);
CREATE INDEX idx_sku_material ON public.material_sku USING btree (material_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_sku_code ON public.material_sku USING btree (sku_code) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX uk_sku_code ON public.material_sku USING btree (sku_code) WHERE (deleted_at IS NULL);

-- 表: purchase_order
-- 创建时间: 2026-04-29
CREATE TABLE purchase_order (
    id bigint NOT NULL DEFAULT nextval('purchase_order_id_seq'::regclass),
    order_no character varying(20) NOT NULL,
    request_id bigint,
    supplier_id bigint NOT NULL,
    buyer_id bigint NOT NULL,
    order_date date NOT NULL,
    expected_date date,
    payment_method character varying(50),
    order_status character varying(20) NOT NULL DEFAULT 'draft'::character varying,
    total_amount numeric(18,2) DEFAULT 0,
    remark text,
    created_by bigint,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    updated_by bigint,
    updated_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp(6) with time zone,
    order_type character varying(20) DEFAULT 'purchase'::character varying,
    supplier_code character varying(20) NOT NULL
);

COMMENT ON TABLE purchase_order IS '采购订单（收货仓库在入库单等下游按批次/仓库选择）';

COMMENT ON COLUMN purchase_order.id IS '采购订单ID，主键';
COMMENT ON COLUMN purchase_order.order_no IS '订单编号，唯一';
COMMENT ON COLUMN purchase_order.request_id IS '关联采购申请ID（可选）';
COMMENT ON COLUMN purchase_order.supplier_id IS '供应商ID';
COMMENT ON COLUMN purchase_order.buyer_id IS '采购员ID';
COMMENT ON COLUMN purchase_order.order_date IS '下单日期';
COMMENT ON COLUMN purchase_order.expected_date IS '预计到货日期';
COMMENT ON COLUMN purchase_order.payment_method IS '付款方式: prepay/on_delivery/monthly/other';
COMMENT ON COLUMN purchase_order.order_status IS '订单状态: draft/ordered/partial_received/full_received/closed';
COMMENT ON COLUMN purchase_order.total_amount IS '订单总金额（由触发器自动汇总）';
COMMENT ON COLUMN purchase_order.remark IS '备注';
COMMENT ON COLUMN purchase_order.created_by IS '创建人ID';
COMMENT ON COLUMN purchase_order.created_at IS '创建时间';
COMMENT ON COLUMN purchase_order.updated_by IS '更新人ID';
COMMENT ON COLUMN purchase_order.updated_at IS '更新时间';
COMMENT ON COLUMN purchase_order.deleted_at IS '删除时间（软删除标记）';
COMMENT ON COLUMN purchase_order.order_type IS '订单类型: purchase-采购订货, return-采购退货';

CREATE INDEX idx_po_date ON public.purchase_order USING btree (order_date) WHERE (deleted_at IS NULL);
CREATE INDEX idx_po_status ON public.purchase_order USING btree (order_status) WHERE (deleted_at IS NULL);
CREATE INDEX idx_po_status_date ON public.purchase_order USING btree (order_status, order_date) WHERE (deleted_at IS NULL);
CREATE INDEX idx_po_supplier ON public.purchase_order USING btree (supplier_id) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX uk_po_no ON public.purchase_order USING btree (order_no) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX purchase_order_pkey ON public.purchase_order USING btree (id);
CREATE INDEX idx_po_supplier_code ON public.purchase_order USING btree (supplier_code) WHERE (deleted_at IS NULL);

-- 表: purchase_order_item
-- 创建时间: 2026-04-29
CREATE TABLE purchase_order_item (
    id bigint NOT NULL DEFAULT nextval('purchase_order_item_id_seq'::regclass),
    order_id bigint NOT NULL,
    material_id bigint NOT NULL,
    quantity numeric(18,3) NOT NULL,
    unit_price numeric(18,2),
    amount numeric(18,2),
    received_quantity numeric(18,3) DEFAULT 0,
    delivery_date date,
    remark text,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    updated_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    unit character varying(20),
    custom_attributes jsonb,
    sku_id bigint
);

COMMENT ON TABLE purchase_order_item IS '采购订单明细表';

COMMENT ON COLUMN purchase_order_item.id IS '明细ID，主键';
COMMENT ON COLUMN purchase_order_item.order_id IS '关联采购订单ID';
COMMENT ON COLUMN purchase_order_item.material_id IS '物料ID';
COMMENT ON COLUMN purchase_order_item.quantity IS '采购数量';
COMMENT ON COLUMN purchase_order_item.unit_price IS '单价';
COMMENT ON COLUMN purchase_order_item.amount IS '金额（数量×单价）';
COMMENT ON COLUMN purchase_order_item.received_quantity IS '已到货数量';
COMMENT ON COLUMN purchase_order_item.delivery_date IS '交货日期';
COMMENT ON COLUMN purchase_order_item.remark IS '备注';
COMMENT ON COLUMN purchase_order_item.created_at IS '创建时间';
COMMENT ON COLUMN purchase_order_item.updated_at IS '更新时间';
COMMENT ON COLUMN purchase_order_item.unit IS '计量单位';
COMMENT ON COLUMN purchase_order_item.custom_attributes IS '自定义属性JSON';
COMMENT ON COLUMN purchase_order_item.sku_id IS 'SKU ID';

CREATE INDEX idx_poi_material ON public.purchase_order_item USING btree (material_id);
CREATE INDEX idx_poi_order ON public.purchase_order_item USING btree (order_id);
CREATE UNIQUE INDEX purchase_order_item_pkey ON public.purchase_order_item USING btree (id);

-- 表: purchase_request
-- 创建时间: 2026-04-29
CREATE TABLE purchase_request (
    id bigint NOT NULL DEFAULT nextval('purchase_request_id_seq'::regclass),
    request_no character varying(20) NOT NULL,
    request_date date NOT NULL,
    requester_id bigint NOT NULL,
    department character varying(100),
    required_date date,
    request_reason character varying(50) NOT NULL,
    project_no character varying(50),
    approval_status character varying(20) NOT NULL DEFAULT 'draft'::character varying,
    approved_by bigint,
    approved_at timestamp(6) with time zone,
    approval_remark text,
    remark text,
    created_by bigint,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    updated_by bigint,
    updated_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp(6) with time zone
);

COMMENT ON TABLE purchase_request IS '采购申请单';

COMMENT ON COLUMN purchase_request.id IS '采购申请ID，主键';
COMMENT ON COLUMN purchase_request.request_no IS '申请单号，唯一';
COMMENT ON COLUMN purchase_request.request_date IS '申请日期';
COMMENT ON COLUMN purchase_request.requester_id IS '申请人ID';
COMMENT ON COLUMN purchase_request.department IS '申请部门';
COMMENT ON COLUMN purchase_request.required_date IS '需求日期';
COMMENT ON COLUMN purchase_request.request_reason IS '申请原因: production/safety_stock/project/other';
COMMENT ON COLUMN purchase_request.project_no IS '关联项目号';
COMMENT ON COLUMN purchase_request.approval_status IS '审批状态: draft/pending/approved/rejected/closed';
COMMENT ON COLUMN purchase_request.approved_by IS '审批人ID';
COMMENT ON COLUMN purchase_request.approved_at IS '审批时间';
COMMENT ON COLUMN purchase_request.approval_remark IS '审批意见';
COMMENT ON COLUMN purchase_request.remark IS '备注';
COMMENT ON COLUMN purchase_request.created_by IS '创建人ID';
COMMENT ON COLUMN purchase_request.created_at IS '创建时间';
COMMENT ON COLUMN purchase_request.updated_by IS '更新人ID';
COMMENT ON COLUMN purchase_request.updated_at IS '更新时间';
COMMENT ON COLUMN purchase_request.deleted_at IS '删除时间（软删除标记）';

CREATE INDEX idx_pr_requester ON public.purchase_request USING btree (requester_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_pr_status ON public.purchase_request USING btree (approval_status) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX uk_pr_no ON public.purchase_request USING btree (request_no) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX purchase_request_pkey ON public.purchase_request USING btree (id);

-- 表: purchase_request_item
-- 创建时间: 2026-04-29
CREATE TABLE purchase_request_item (
    id bigint NOT NULL DEFAULT nextval('purchase_request_item_id_seq'::regclass),
    request_id bigint NOT NULL,
    material_id bigint NOT NULL,
    quantity numeric(18,3) NOT NULL,
    unit character varying(20) NOT NULL,
    current_stock numeric(18,3),
    remark text,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now()
);

COMMENT ON TABLE purchase_request_item IS '采购申请明细表';

COMMENT ON COLUMN purchase_request_item.id IS '明细ID，主键';
COMMENT ON COLUMN purchase_request_item.request_id IS '关联采购申请ID';
COMMENT ON COLUMN purchase_request_item.material_id IS '物料ID';
COMMENT ON COLUMN purchase_request_item.quantity IS '申请数量';
COMMENT ON COLUMN purchase_request_item.unit IS '计量单位';
COMMENT ON COLUMN purchase_request_item.current_stock IS '当前库存快照';
COMMENT ON COLUMN purchase_request_item.remark IS '备注';
COMMENT ON COLUMN purchase_request_item.created_at IS '创建时间';

CREATE INDEX idx_pri_request ON public.purchase_request_item USING btree (request_id);
CREATE UNIQUE INDEX purchase_request_item_pkey ON public.purchase_request_item USING btree (id);

-- 表: return_material_order
-- 创建时间: 2026-04-29
CREATE TABLE return_material_order (
    id bigint NOT NULL DEFAULT nextval('return_material_order_id_seq'::regclass),
    order_no character varying(20) NOT NULL,
    project_no character varying(50) NOT NULL,
    product_name character varying(200) NOT NULL,
    warehouse_id bigint NOT NULL,
    warehouse_code character varying(10) NOT NULL,
    order_date date NOT NULL DEFAULT CURRENT_DATE,
    designer_id bigint NOT NULL,
    designer_name character varying(50) NOT NULL,
    status character varying(20) NOT NULL DEFAULT 'draft'::character varying,
    stock_in_id bigint,
    remark text,
    created_by bigint,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_by bigint,
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone
);

COMMENT ON TABLE return_material_order IS '退料订单';

COMMENT ON COLUMN return_material_order.order_no IS '退料单号，规则：RMO + 年月日 + 序号';
COMMENT ON COLUMN return_material_order.project_no IS '项目/图纸编号';
COMMENT ON COLUMN return_material_order.product_name IS '产品名称';
COMMENT ON COLUMN return_material_order.warehouse_id IS '退料仓库ID';
COMMENT ON COLUMN return_material_order.warehouse_code IS '仓库编码';
COMMENT ON COLUMN return_material_order.order_date IS '退料日期';
COMMENT ON COLUMN return_material_order.designer_id IS '设计员ID';
COMMENT ON COLUMN return_material_order.designer_name IS '设计员姓名';
COMMENT ON COLUMN return_material_order.status IS '状态：草稿/待确认/已确认/已完成/已取消';
COMMENT ON COLUMN return_material_order.stock_in_id IS '关联的入库单ID';
COMMENT ON COLUMN return_material_order.remark IS '备注';

CREATE UNIQUE INDEX return_material_order_pkey ON public.return_material_order USING btree (id);
CREATE UNIQUE INDEX return_material_order_order_no_key ON public.return_material_order USING btree (order_no);
CREATE INDEX idx_rmo_no ON public.return_material_order USING btree (order_no) WHERE (deleted_at IS NULL);
CREATE INDEX idx_rmo_project ON public.return_material_order USING btree (project_no) WHERE (deleted_at IS NULL);
CREATE INDEX idx_rmo_status ON public.return_material_order USING btree (status) WHERE (deleted_at IS NULL);
CREATE INDEX idx_rmo_date ON public.return_material_order USING btree (order_date) WHERE (deleted_at IS NULL);

-- 表: return_material_order_item
-- 创建时间: 2026-04-29
CREATE TABLE return_material_order_item (
    id bigint NOT NULL DEFAULT nextval('return_material_order_item_id_seq'::regclass),
    order_id bigint NOT NULL,
    material_id bigint NOT NULL,
    quantity numeric(18,3) NOT NULL,
    unit character varying(20) NOT NULL,
    remark text,
    created_at timestamp with time zone NOT NULL DEFAULT now()
);

COMMENT ON TABLE return_material_order_item IS '退料订单明细';

COMMENT ON COLUMN return_material_order_item.order_id IS '退料单ID';
COMMENT ON COLUMN return_material_order_item.material_id IS '物料ID';
COMMENT ON COLUMN return_material_order_item.quantity IS '退料数量';
COMMENT ON COLUMN return_material_order_item.unit IS '单位';
COMMENT ON COLUMN return_material_order_item.remark IS '备注';

CREATE UNIQUE INDEX return_material_order_item_pkey ON public.return_material_order_item USING btree (id);
CREATE INDEX idx_rmoi_order ON public.return_material_order_item USING btree (order_id);
CREATE INDEX idx_rmoi_material ON public.return_material_order_item USING btree (material_id);

-- 表: return_order
-- 创建时间: 2026-04-29
CREATE TABLE return_order (
    id bigint NOT NULL DEFAULT nextval('return_order_id_seq'::regclass),
    return_no character varying(20) NOT NULL,
    return_type character varying(20) NOT NULL,
    ref_doc_type character varying(50),
    ref_doc_id bigint,
    warehouse_id bigint NOT NULL,
    return_date date NOT NULL,
    return_status character varying(20) NOT NULL DEFAULT 'draft'::character varying,
    remark text,
    created_by bigint,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    updated_by bigint,
    updated_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp(6) with time zone,
    warehouse_code character varying(10) NOT NULL,
    supplier_code character varying(20),
    stock_out_id bigint
);

COMMENT ON TABLE return_order IS '退货单（采购退货/销售退货共用）';

COMMENT ON COLUMN return_order.id IS '退货单ID，主键';
COMMENT ON COLUMN return_order.return_no IS '退货单号，唯一';
COMMENT ON COLUMN return_order.return_type IS '退货类型: purchase_return/sales_return';
COMMENT ON COLUMN return_order.ref_doc_type IS '关联单据类型';
COMMENT ON COLUMN return_order.ref_doc_id IS '关联单据ID';
COMMENT ON COLUMN return_order.warehouse_id IS '仓库ID';
COMMENT ON COLUMN return_order.return_date IS '退货日期';
COMMENT ON COLUMN return_order.return_status IS '退货状态: draft/confirmed/closed';
COMMENT ON COLUMN return_order.remark IS '备注';
COMMENT ON COLUMN return_order.created_by IS '创建人ID';
COMMENT ON COLUMN return_order.created_at IS '创建时间';
COMMENT ON COLUMN return_order.updated_by IS '更新人ID';
COMMENT ON COLUMN return_order.updated_at IS '更新时间';
COMMENT ON COLUMN return_order.deleted_at IS '删除时间（软删除标记）';

CREATE UNIQUE INDEX uk_return_no ON public.return_order USING btree (return_no) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX return_order_pkey ON public.return_order USING btree (id);
CREATE INDEX idx_ro_warehouse_code ON public.return_order USING btree (warehouse_code) WHERE (deleted_at IS NULL);

-- 表: return_order_item
-- 创建时间: 2026-04-29
CREATE TABLE return_order_item (
    id bigint NOT NULL DEFAULT nextval('return_order_item_id_seq'::regclass),
    return_id bigint NOT NULL,
    material_id bigint NOT NULL,
    quantity numeric(18,3) NOT NULL,
    unit character varying(20) NOT NULL,
    remark text,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    sku_id bigint,
    inventory_id bigint,
    warehouse_id bigint,
    warehouse_code character varying(10)
);

COMMENT ON TABLE return_order_item IS '退货明细表';

COMMENT ON COLUMN return_order_item.id IS '明细ID，主键';
COMMENT ON COLUMN return_order_item.return_id IS '关联退货单ID';
COMMENT ON COLUMN return_order_item.material_id IS '物料ID';
COMMENT ON COLUMN return_order_item.quantity IS '退货数量';
COMMENT ON COLUMN return_order_item.unit IS '计量单位';
COMMENT ON COLUMN return_order_item.remark IS '备注';
COMMENT ON COLUMN return_order_item.created_at IS '创建时间';
COMMENT ON COLUMN return_order_item.warehouse_id IS '库存所在仓库ID（采购退货支持多仓）';
COMMENT ON COLUMN return_order_item.warehouse_code IS '库存所在仓库编码（采购退货支持多仓）';

CREATE INDEX idx_roi_return ON public.return_order_item USING btree (return_id);
CREATE UNIQUE INDEX return_order_item_pkey ON public.return_order_item USING btree (id);

-- 表: reversal_order
-- 创建时间: 2026-04-29
CREATE TABLE reversal_order (
    id bigint NOT NULL DEFAULT nextval('reversal_order_id_seq'::regclass),
    order_no character varying(20) NOT NULL,
    project_no character varying(50) NOT NULL,
    product_name character varying(200) NOT NULL,
    warehouse_id bigint NOT NULL,
    warehouse_code character varying(10) NOT NULL,
    order_date date NOT NULL DEFAULT CURRENT_DATE,
    designer_id bigint NOT NULL,
    designer_name character varying(50) NOT NULL,
    status character varying(20) NOT NULL DEFAULT 'draft'::character varying,
    stock_in_id bigint,
    remark text,
    created_by bigint,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_by bigint,
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone
);

COMMENT ON TABLE reversal_order IS '退料订单';

COMMENT ON COLUMN reversal_order.order_no IS '退料单号，规则：RO + 年月日 + 序号';
COMMENT ON COLUMN reversal_order.project_no IS '项目/图纸编号';
COMMENT ON COLUMN reversal_order.product_name IS '产品名称';
COMMENT ON COLUMN reversal_order.warehouse_id IS '退料仓库ID';
COMMENT ON COLUMN reversal_order.warehouse_code IS '仓库编码';
COMMENT ON COLUMN reversal_order.order_date IS '退料日期';
COMMENT ON COLUMN reversal_order.designer_id IS '设计员ID';
COMMENT ON COLUMN reversal_order.designer_name IS '设计员姓名';
COMMENT ON COLUMN reversal_order.status IS '状态：草稿/待确认/已确认/已完成/已取消';
COMMENT ON COLUMN reversal_order.stock_in_id IS '关联的入库单ID';
COMMENT ON COLUMN reversal_order.remark IS '备注';

CREATE UNIQUE INDEX reversal_order_pkey ON public.reversal_order USING btree (id);
CREATE UNIQUE INDEX reversal_order_order_no_key ON public.reversal_order USING btree (order_no);
CREATE INDEX idx_ro_no ON public.reversal_order USING btree (order_no) WHERE (deleted_at IS NULL);
CREATE INDEX idx_ro_project ON public.reversal_order USING btree (project_no) WHERE (deleted_at IS NULL);
CREATE INDEX idx_ro_status ON public.reversal_order USING btree (status) WHERE (deleted_at IS NULL);
CREATE INDEX idx_ro_date ON public.reversal_order USING btree (order_date) WHERE (deleted_at IS NULL);

-- 表: reversal_order_item
-- 创建时间: 2026-04-29
CREATE TABLE reversal_order_item (
    id bigint NOT NULL DEFAULT nextval('reversal_order_item_id_seq'::regclass),
    order_id bigint NOT NULL,
    material_id bigint NOT NULL,
    quantity numeric(18,3) NOT NULL,
    unit character varying(20) NOT NULL,
    remark text,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    inventory_id bigint
);

COMMENT ON TABLE reversal_order_item IS '退料订单明细';

COMMENT ON COLUMN reversal_order_item.order_id IS '退料单ID';
COMMENT ON COLUMN reversal_order_item.material_id IS '物料ID';
COMMENT ON COLUMN reversal_order_item.quantity IS '退料数量';
COMMENT ON COLUMN reversal_order_item.unit IS '单位';
COMMENT ON COLUMN reversal_order_item.remark IS '备注';
COMMENT ON COLUMN reversal_order_item.inventory_id IS '关联库存ID';

CREATE UNIQUE INDEX reversal_order_item_pkey ON public.reversal_order_item USING btree (id);
CREATE INDEX idx_roi_order ON public.reversal_order_item USING btree (order_id);
CREATE INDEX idx_roi_material ON public.reversal_order_item USING btree (material_id);

-- 表: sales_order
-- 创建时间: 2026-04-29
CREATE TABLE sales_order (
    id bigint NOT NULL DEFAULT nextval('sales_order_id_seq'::regclass),
    order_no character varying(20) NOT NULL,
    customer_id bigint NOT NULL,
    sales_person_id bigint NOT NULL,
    order_date date NOT NULL,
    delivery_date date,
    order_status character varying(20) NOT NULL DEFAULT 'draft'::character varying,
    total_amount numeric(18,2) DEFAULT 0,
    remark text,
    created_by bigint,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    updated_by bigint,
    updated_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp(6) with time zone,
    customer_code character varying(20) NOT NULL
);

COMMENT ON TABLE sales_order IS '销售订单';

COMMENT ON COLUMN sales_order.id IS '销售订单ID，主键';
COMMENT ON COLUMN sales_order.order_no IS '订单编号，唯一';
COMMENT ON COLUMN sales_order.customer_id IS '客户ID';
COMMENT ON COLUMN sales_order.sales_person_id IS '销售员ID';
COMMENT ON COLUMN sales_order.order_date IS '下单日期';
COMMENT ON COLUMN sales_order.delivery_date IS '交货日期';
COMMENT ON COLUMN sales_order.order_status IS '订单状态: draft/confirmed/preparing/shipped/received/closed/cancelled';
COMMENT ON COLUMN sales_order.total_amount IS '订单总金额（由触发器自动汇总）';
COMMENT ON COLUMN sales_order.remark IS '备注';
COMMENT ON COLUMN sales_order.created_by IS '创建人ID';
COMMENT ON COLUMN sales_order.created_at IS '创建时间';
COMMENT ON COLUMN sales_order.updated_by IS '更新人ID';
COMMENT ON COLUMN sales_order.updated_at IS '更新时间';
COMMENT ON COLUMN sales_order.deleted_at IS '删除时间（软删除标记）';

CREATE INDEX idx_sales_customer ON public.sales_order USING btree (customer_id) WHERE (deleted_at IS NULL);
CREATE INDEX idx_sales_status ON public.sales_order USING btree (order_status) WHERE (deleted_at IS NULL);
CREATE INDEX idx_sales_status_date ON public.sales_order USING btree (order_status, order_date) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX uk_sales_no ON public.sales_order USING btree (order_no) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX sales_order_pkey ON public.sales_order USING btree (id);
CREATE INDEX idx_sales_customer_code ON public.sales_order USING btree (customer_code) WHERE (deleted_at IS NULL);

-- 表: sales_order_item
-- 创建时间: 2026-04-29
CREATE TABLE sales_order_item (
    id bigint NOT NULL DEFAULT nextval('sales_order_item_id_seq'::regclass),
    order_id bigint NOT NULL,
    material_id bigint NOT NULL,
    quantity numeric(18,3) NOT NULL,
    unit_price numeric(18,2),
    amount numeric(18,2),
    shipped_quantity numeric(18,3) DEFAULT 0,
    remark text,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    updated_at timestamp(6) with time zone NOT NULL DEFAULT now()
);

COMMENT ON TABLE sales_order_item IS '销售订单明细表';

COMMENT ON COLUMN sales_order_item.id IS '明细ID，主键';
COMMENT ON COLUMN sales_order_item.order_id IS '关联销售订单ID';
COMMENT ON COLUMN sales_order_item.material_id IS '物料ID';
COMMENT ON COLUMN sales_order_item.quantity IS '销售数量';
COMMENT ON COLUMN sales_order_item.unit_price IS '单价';
COMMENT ON COLUMN sales_order_item.amount IS '金额';
COMMENT ON COLUMN sales_order_item.shipped_quantity IS '已发货数量';
COMMENT ON COLUMN sales_order_item.remark IS '备注';
COMMENT ON COLUMN sales_order_item.created_at IS '创建时间';
COMMENT ON COLUMN sales_order_item.updated_at IS '更新时间';

CREATE INDEX idx_sales_item_material ON public.sales_order_item USING btree (material_id);
CREATE INDEX idx_sales_item_order ON public.sales_order_item USING btree (order_id);
CREATE UNIQUE INDEX sales_order_item_pkey ON public.sales_order_item USING btree (id);

-- 表: sku_serial_code
-- 创建时间: 2026-04-29
CREATE TABLE sku_serial_code (
    id bigint NOT NULL DEFAULT nextval('sku_serial_code_id_seq'::regclass),
    serial_code character varying(50) NOT NULL,
    material_id bigint NOT NULL,
    material_code character varying(30) NOT NULL,
    material_name character varying(100) NOT NULL,
    stock_in_id bigint,
    stock_in_item_id bigint,
    inventory_id bigint,
    warehouse_id bigint,
    status character varying(20) NOT NULL DEFAULT 'in_stock'::character varying,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

COMMENT ON TABLE sku_serial_code IS 'SKU具体编码表（每个编码对应一个具体设备/部件）';

COMMENT ON COLUMN sku_serial_code.serial_code IS '具体编码（如材料检号+序号）';
COMMENT ON COLUMN sku_serial_code.material_id IS '物料ID';
COMMENT ON COLUMN sku_serial_code.material_code IS '物料编码';
COMMENT ON COLUMN sku_serial_code.material_name IS '物料名称';
COMMENT ON COLUMN sku_serial_code.stock_in_id IS '入库单ID';
COMMENT ON COLUMN sku_serial_code.stock_in_item_id IS '入库明细ID';
COMMENT ON COLUMN sku_serial_code.inventory_id IS '库存ID';
COMMENT ON COLUMN sku_serial_code.warehouse_id IS '仓库ID';
COMMENT ON COLUMN sku_serial_code.status IS '状态: in_stock-在库, issued-已领用, returned-已退回, scrapped-已报废';

CREATE UNIQUE INDEX sku_serial_code_pkey ON public.sku_serial_code USING btree (id);
CREATE UNIQUE INDEX idx_sku_serial_code_code ON public.sku_serial_code USING btree (serial_code);
CREATE INDEX idx_sku_serial_code_material ON public.sku_serial_code USING btree (material_id);
CREATE INDEX idx_sku_serial_code_inventory ON public.sku_serial_code USING btree (inventory_id);
CREATE INDEX idx_sku_serial_code_status ON public.sku_serial_code USING btree (status);

-- 表: sku_serial_trace
-- 创建时间: 2026-04-29
CREATE TABLE sku_serial_trace (
    id bigint NOT NULL DEFAULT nextval('sku_serial_trace_id_seq'::regclass),
    serial_code_id bigint NOT NULL,
    serial_code character varying(50) NOT NULL,
    action character varying(30) NOT NULL,
    ref_doc_type character varying(50),
    ref_doc_no character varying(20),
    ref_doc_id bigint,
    from_warehouse_id bigint,
    to_warehouse_id bigint,
    operator_id bigint,
    operator_name character varying(50),
    remark text,
    created_at timestamp with time zone NOT NULL DEFAULT now()
);

COMMENT ON TABLE sku_serial_trace IS '编码流转记录表（全生命周期追溯）';

COMMENT ON COLUMN sku_serial_trace.serial_code_id IS '关联编码ID';
COMMENT ON COLUMN sku_serial_trace.serial_code IS '具体编码';
COMMENT ON COLUMN sku_serial_trace.action IS '操作类型';
COMMENT ON COLUMN sku_serial_trace.ref_doc_type IS '关联单据类型';
COMMENT ON COLUMN sku_serial_trace.ref_doc_no IS '关联单据编号';
COMMENT ON COLUMN sku_serial_trace.ref_doc_id IS '关联单据ID';
COMMENT ON COLUMN sku_serial_trace.from_warehouse_id IS '来源仓库';
COMMENT ON COLUMN sku_serial_trace.to_warehouse_id IS '目标仓库';
COMMENT ON COLUMN sku_serial_trace.operator_id IS '操作人ID';
COMMENT ON COLUMN sku_serial_trace.operator_name IS '操作人姓名';

CREATE UNIQUE INDEX sku_serial_trace_pkey ON public.sku_serial_trace USING btree (id);
CREATE INDEX idx_serial_trace_code_id ON public.sku_serial_trace USING btree (serial_code_id);
CREATE INDEX idx_serial_trace_code ON public.sku_serial_trace USING btree (serial_code);
CREATE INDEX idx_serial_trace_action ON public.sku_serial_trace USING btree (action);
CREATE INDEX idx_serial_trace_created ON public.sku_serial_trace USING btree (created_at);

-- 表: stock_in
-- 创建时间: 2026-04-29
CREATE TABLE stock_in (
    id bigint NOT NULL DEFAULT nextval('stock_in_id_seq'::regclass),
    stock_in_no character varying(20) NOT NULL,
    purchase_order_id bigint,
    supplier_id bigint,
    warehouse_id bigint,
    stock_in_date date NOT NULL,
    remark text,
    created_by bigint,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    updated_by bigint,
    updated_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp(6) with time zone,
    stock_in_status character varying(20) NOT NULL DEFAULT 'preparing'::character varying,
    warehouse_code character varying(10),
    supplier_code character varying(20),
    stock_in_type character varying(20) NOT NULL DEFAULT 'purchase'::character varying
);

COMMENT ON TABLE stock_in IS '入库单';

COMMENT ON COLUMN stock_in.id IS '入库单ID，主键';
COMMENT ON COLUMN stock_in.stock_in_no IS '入库单号，唯一';
COMMENT ON COLUMN stock_in.purchase_order_id IS '关联采购订单ID';
COMMENT ON COLUMN stock_in.supplier_id IS '供应商ID';
COMMENT ON COLUMN stock_in.warehouse_id IS '入库仓库ID';
COMMENT ON COLUMN stock_in.stock_in_date IS '入库日期';
COMMENT ON COLUMN stock_in.remark IS '备注';
COMMENT ON COLUMN stock_in.created_by IS '创建人ID';
COMMENT ON COLUMN stock_in.created_at IS '创建时间';
COMMENT ON COLUMN stock_in.updated_by IS '更新人ID';
COMMENT ON COLUMN stock_in.updated_at IS '更新时间';
COMMENT ON COLUMN stock_in.deleted_at IS '删除时间（软删除标记）';
COMMENT ON COLUMN stock_in.stock_in_status IS '入库流程状态: preparing(待备货)/pending(待入库)/passed(已入库)/failed(已拒绝)';
COMMENT ON COLUMN stock_in.stock_in_type IS '入库类型: purchase(采购入库)/return(销售退货入库)/reversal(退料入库)';

CREATE INDEX idx_si_date ON public.stock_in USING btree (stock_in_date) WHERE (deleted_at IS NULL);
CREATE INDEX idx_si_po ON public.stock_in USING btree (purchase_order_id) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX uk_si_no ON public.stock_in USING btree (stock_in_no) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX stock_in_pkey ON public.stock_in USING btree (id);
CREATE INDEX idx_si_warehouse_code ON public.stock_in USING btree (warehouse_code) WHERE (deleted_at IS NULL);
CREATE INDEX idx_si_supplier_code ON public.stock_in USING btree (supplier_code) WHERE (deleted_at IS NULL);

-- 表: stock_in_confirm_log
-- 创建时间: 2026-04-29
CREATE TABLE stock_in_confirm_log (
    id bigint NOT NULL DEFAULT nextval('stock_in_confirm_log_id_seq'::regclass),
    stock_in_id bigint NOT NULL,
    stock_in_item_id bigint NOT NULL,
    material_id bigint NOT NULL,
    purchase_quantity numeric(18,3) NOT NULL DEFAULT 0,
    before_received_quantity numeric(18,3) NOT NULL DEFAULT 0,
    current_received_quantity numeric(18,3) NOT NULL DEFAULT 0,
    after_received_quantity numeric(18,3) NOT NULL DEFAULT 0,
    operator_id bigint NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    remark text
);

COMMENT ON TABLE stock_in_confirm_log IS '入库单确认日志（记录每次确认入库的数量变化）';

COMMENT ON COLUMN stock_in_confirm_log.stock_in_id IS '入库单ID';
COMMENT ON COLUMN stock_in_confirm_log.stock_in_item_id IS '入库单明细ID';
COMMENT ON COLUMN stock_in_confirm_log.material_id IS '物料ID';
COMMENT ON COLUMN stock_in_confirm_log.purchase_quantity IS '采购数量';
COMMENT ON COLUMN stock_in_confirm_log.before_received_quantity IS '确认前已入库数量';
COMMENT ON COLUMN stock_in_confirm_log.current_received_quantity IS '本次入库数量';
COMMENT ON COLUMN stock_in_confirm_log.after_received_quantity IS '确认后累计入库数量';
COMMENT ON COLUMN stock_in_confirm_log.operator_id IS '操作人ID';
COMMENT ON COLUMN stock_in_confirm_log.created_at IS '确认时间';
COMMENT ON COLUMN stock_in_confirm_log.remark IS '备注';

CREATE UNIQUE INDEX stock_in_confirm_log_pkey ON public.stock_in_confirm_log USING btree (id);
CREATE INDEX idx_stock_in_confirm_log_stock_in ON public.stock_in_confirm_log USING btree (stock_in_id, created_at DESC);
CREATE INDEX idx_stock_in_confirm_log_item ON public.stock_in_confirm_log USING btree (stock_in_item_id, created_at DESC);

-- 表: stock_in_item
-- 创建时间: 2026-04-29
CREATE TABLE stock_in_item (
    id bigint NOT NULL DEFAULT nextval('stock_in_item_id_seq'::regclass),
    stock_in_id bigint NOT NULL,
    material_id bigint NOT NULL,
    arrived_quantity numeric(18,3) NOT NULL,
    accepted_quantity numeric(18,3) NOT NULL,
    unit character varying(20) NOT NULL,
    cert_id bigint,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    unit_cost numeric(18,2),
    custom_attributes jsonb,
    sku_id bigint
);

COMMENT ON TABLE stock_in_item IS '入库明细表';

COMMENT ON COLUMN stock_in_item.id IS '明细ID，主键';
COMMENT ON COLUMN stock_in_item.stock_in_id IS '关联入库单ID';
COMMENT ON COLUMN stock_in_item.material_id IS '物料ID';
COMMENT ON COLUMN stock_in_item.arrived_quantity IS '到货数量';
COMMENT ON COLUMN stock_in_item.accepted_quantity IS '合格数量（实际入库数量）';
COMMENT ON COLUMN stock_in_item.unit IS '计量单位';
COMMENT ON COLUMN stock_in_item.cert_id IS '关联材质证明书ID';
COMMENT ON COLUMN stock_in_item.created_at IS '创建时间';
COMMENT ON COLUMN stock_in_item.unit_cost IS '入库单价';
COMMENT ON COLUMN stock_in_item.custom_attributes IS '自定义属性JSON';
COMMENT ON COLUMN stock_in_item.sku_id IS 'SKU ID';

CREATE INDEX idx_sii_material ON public.stock_in_item USING btree (material_id);
CREATE INDEX idx_sii_stock_in ON public.stock_in_item USING btree (stock_in_id);
CREATE UNIQUE INDEX stock_in_item_pkey ON public.stock_in_item USING btree (id);

-- 表: stock_in_item_serial_selection
-- 创建时间: 2026-04-29
CREATE TABLE stock_in_item_serial_selection (
    id bigint NOT NULL DEFAULT nextval('stock_in_item_serial_selection_id_seq'::regclass),
    stock_in_id bigint NOT NULL,
    stock_in_item_id bigint NOT NULL,
    serial_code_id bigint NOT NULL,
    created_by bigint,
    created_at timestamp with time zone NOT NULL DEFAULT now()
);


CREATE UNIQUE INDEX stock_in_item_serial_selection_pkey ON public.stock_in_item_serial_selection USING btree (id);
CREATE UNIQUE INDEX uk_siss_item_serial ON public.stock_in_item_serial_selection USING btree (stock_in_item_id, serial_code_id);
CREATE UNIQUE INDEX uk_siss_serial_only_one ON public.stock_in_item_serial_selection USING btree (serial_code_id);
CREATE INDEX idx_siss_stock_in ON public.stock_in_item_serial_selection USING btree (stock_in_id);
CREATE INDEX idx_siss_item ON public.stock_in_item_serial_selection USING btree (stock_in_item_id);

-- 表: stock_out
-- 创建时间: 2026-04-29
CREATE TABLE stock_out (
    id bigint NOT NULL DEFAULT nextval('stock_out_id_seq'::regclass),
    stock_out_no character varying(20) NOT NULL,
    out_type character varying(20) NOT NULL,
    ref_doc_type character varying(50),
    ref_doc_id bigint,
    stock_out_date date NOT NULL,
    receiver character varying(100),
    remark text,
    created_by bigint,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    updated_by bigint,
    updated_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp(6) with time zone,
    status character varying(20) NOT NULL DEFAULT 'draft'::character varying,
    confirmed_at timestamp with time zone
);

COMMENT ON TABLE stock_out IS '出库单（仓库由明细 inventory_id 对应库存解析）';

COMMENT ON COLUMN stock_out.id IS '出库单ID，主键';
COMMENT ON COLUMN stock_out.stock_out_no IS '出库单号，唯一';
COMMENT ON COLUMN stock_out.out_type IS '出库类型: sales/production/transfer/other';
COMMENT ON COLUMN stock_out.ref_doc_type IS '关联单据类型';
COMMENT ON COLUMN stock_out.ref_doc_id IS '关联单据ID';
COMMENT ON COLUMN stock_out.stock_out_date IS '出库日期';
COMMENT ON COLUMN stock_out.receiver IS '领料人/收货方';
COMMENT ON COLUMN stock_out.remark IS '备注';
COMMENT ON COLUMN stock_out.created_by IS '创建人ID';
COMMENT ON COLUMN stock_out.created_at IS '创建时间';
COMMENT ON COLUMN stock_out.updated_by IS '更新人ID';
COMMENT ON COLUMN stock_out.updated_at IS '更新时间';
COMMENT ON COLUMN stock_out.deleted_at IS '删除时间（软删除标记）';

CREATE INDEX idx_so_date ON public.stock_out USING btree (stock_out_date) WHERE (deleted_at IS NULL);
CREATE INDEX idx_so_type ON public.stock_out USING btree (out_type) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX uk_so_no ON public.stock_out USING btree (stock_out_no) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX stock_out_pkey ON public.stock_out USING btree (id);

-- 表: stock_out_item
-- 创建时间: 2026-04-29
CREATE TABLE stock_out_item (
    id bigint NOT NULL DEFAULT nextval('stock_out_item_id_seq'::regclass),
    stock_out_id bigint NOT NULL,
    material_id bigint NOT NULL,
    quantity numeric(18,3) NOT NULL,
    unit character varying(20) NOT NULL,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    inventory_id bigint
);

COMMENT ON TABLE stock_out_item IS '出库明细表';

COMMENT ON COLUMN stock_out_item.id IS '明细ID，主键';
COMMENT ON COLUMN stock_out_item.stock_out_id IS '关联出库单ID';
COMMENT ON COLUMN stock_out_item.material_id IS '物料ID';
COMMENT ON COLUMN stock_out_item.quantity IS '出库数量';
COMMENT ON COLUMN stock_out_item.unit IS '计量单位';
COMMENT ON COLUMN stock_out_item.created_at IS '创建时间';

CREATE INDEX idx_soi_material ON public.stock_out_item USING btree (material_id);
CREATE INDEX idx_soi_stock_out ON public.stock_out_item USING btree (stock_out_id);
CREATE UNIQUE INDEX stock_out_item_pkey ON public.stock_out_item USING btree (id);

-- 表: stock_out_item_serial_selection
-- 创建时间: 2026-04-29
CREATE TABLE stock_out_item_serial_selection (
    id bigint NOT NULL DEFAULT nextval('stock_out_item_serial_selection_id_seq'::regclass),
    stock_out_item_id bigint NOT NULL,
    serial_code_id bigint NOT NULL,
    created_by bigint,
    created_at timestamp with time zone NOT NULL DEFAULT now()
);

COMMENT ON TABLE stock_out_item_serial_selection IS '出库明细编码选择暂存表（确认前）';

COMMENT ON COLUMN stock_out_item_serial_selection.stock_out_item_id IS '出库明细ID';
COMMENT ON COLUMN stock_out_item_serial_selection.serial_code_id IS '编码ID';
COMMENT ON COLUMN stock_out_item_serial_selection.created_by IS '选择人';

CREATE UNIQUE INDEX stock_out_item_serial_selection_pkey ON public.stock_out_item_serial_selection USING btree (id);
CREATE UNIQUE INDEX uk_stock_out_item_serial_selection ON public.stock_out_item_serial_selection USING btree (stock_out_item_id, serial_code_id);
CREATE INDEX idx_soss_item ON public.stock_out_item_serial_selection USING btree (stock_out_item_id);
CREATE INDEX idx_soss_serial ON public.stock_out_item_serial_selection USING btree (serial_code_id);

-- 表: stock_transfer
-- 创建时间: 2026-04-29
CREATE TABLE stock_transfer (
    id bigint NOT NULL DEFAULT nextval('stock_transfer_id_seq'::regclass),
    transfer_no character varying(20) NOT NULL,
    from_warehouse_id bigint NOT NULL,
    to_warehouse_id bigint NOT NULL,
    transfer_date date NOT NULL,
    transfer_status character varying(20) NOT NULL DEFAULT 'draft'::character varying,
    remark text,
    created_by bigint,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    updated_by bigint,
    updated_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp(6) with time zone
);

COMMENT ON TABLE stock_transfer IS '库存调拨单';

COMMENT ON COLUMN stock_transfer.id IS '调拨单ID，主键';
COMMENT ON COLUMN stock_transfer.transfer_no IS '调拨单号，唯一';
COMMENT ON COLUMN stock_transfer.from_warehouse_id IS '调出仓库ID';
COMMENT ON COLUMN stock_transfer.to_warehouse_id IS '调入仓库ID';
COMMENT ON COLUMN stock_transfer.transfer_date IS '调拨日期';
COMMENT ON COLUMN stock_transfer.transfer_status IS '调拨状态: draft/in_transit/completed/cancelled';
COMMENT ON COLUMN stock_transfer.remark IS '备注';
COMMENT ON COLUMN stock_transfer.created_by IS '创建人ID';
COMMENT ON COLUMN stock_transfer.created_at IS '创建时间';
COMMENT ON COLUMN stock_transfer.updated_by IS '更新人ID';
COMMENT ON COLUMN stock_transfer.updated_at IS '更新时间';
COMMENT ON COLUMN stock_transfer.deleted_at IS '删除时间（软删除标记）';

CREATE UNIQUE INDEX uk_transfer_no ON public.stock_transfer USING btree (transfer_no) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX stock_transfer_pkey ON public.stock_transfer USING btree (id);

-- 表: stock_transfer_item
-- 创建时间: 2026-04-29
CREATE TABLE stock_transfer_item (
    id bigint NOT NULL DEFAULT nextval('stock_transfer_item_id_seq'::regclass),
    transfer_id bigint NOT NULL,
    material_id bigint NOT NULL,
    quantity numeric(18,3) NOT NULL,
    unit character varying(20) NOT NULL,
    remark text,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now()
);

COMMENT ON TABLE stock_transfer_item IS '调拨明细表';

COMMENT ON COLUMN stock_transfer_item.id IS '明细ID，主键';
COMMENT ON COLUMN stock_transfer_item.transfer_id IS '关联调拨单ID';
COMMENT ON COLUMN stock_transfer_item.material_id IS '物料ID';
COMMENT ON COLUMN stock_transfer_item.quantity IS '调拨数量';
COMMENT ON COLUMN stock_transfer_item.unit IS '计量单位';
COMMENT ON COLUMN stock_transfer_item.remark IS '备注';
COMMENT ON COLUMN stock_transfer_item.created_at IS '创建时间';

CREATE INDEX idx_sti_transfer ON public.stock_transfer_item USING btree (transfer_id);
CREATE UNIQUE INDEX stock_transfer_item_pkey ON public.stock_transfer_item USING btree (id);

-- 表: supplier
-- 创建时间: 2026-04-29
CREATE TABLE supplier (
    id bigint NOT NULL DEFAULT nextval('supplier_id_seq'::regclass),
    supplier_code character varying(20) NOT NULL,
    supplier_name character varying(200) NOT NULL,
    credit_code character varying(18),
    supplier_type character varying(50) NOT NULL,
    contact_person character varying(50) NOT NULL,
    contact_phone character varying(20) NOT NULL,
    address character varying(500),
    is_qualified boolean NOT NULL DEFAULT false,
    qualification_expire date,
    supplier_rating character varying(5),
    bank_name character varying(100),
    bank_account character varying(30),
    remark text,
    status character varying(20) NOT NULL DEFAULT 'enabled'::character varying,
    created_by bigint,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    updated_by bigint,
    updated_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp(6) with time zone
);

COMMENT ON TABLE supplier IS '供应商主数据表';

COMMENT ON COLUMN supplier.id IS '供应商ID，主键';
COMMENT ON COLUMN supplier.supplier_code IS '供应商编码，唯一';
COMMENT ON COLUMN supplier.supplier_name IS '供应商名称';
COMMENT ON COLUMN supplier.credit_code IS '统一社会信用代码';
COMMENT ON COLUMN supplier.supplier_type IS '供应商类型: raw_material/welding/standard_part/purchased/other';
COMMENT ON COLUMN supplier.contact_person IS '联系人';
COMMENT ON COLUMN supplier.contact_phone IS '联系电话';
COMMENT ON COLUMN supplier.address IS '地址';
COMMENT ON COLUMN supplier.is_qualified IS '是否合格供方';
COMMENT ON COLUMN supplier.qualification_expire IS '合格供方资质到期日';
COMMENT ON COLUMN supplier.supplier_rating IS '供应商评级: A/B/C/D';
COMMENT ON COLUMN supplier.bank_name IS '开户银行';
COMMENT ON COLUMN supplier.bank_account IS '银行账号';
COMMENT ON COLUMN supplier.remark IS '备注';
COMMENT ON COLUMN supplier.status IS '状态: enabled-启用, disabled-禁用, blacklisted-黑名单';
COMMENT ON COLUMN supplier.created_by IS '创建人ID';
COMMENT ON COLUMN supplier.created_at IS '创建时间';
COMMENT ON COLUMN supplier.updated_by IS '更新人ID';
COMMENT ON COLUMN supplier.updated_at IS '更新时间';
COMMENT ON COLUMN supplier.deleted_at IS '删除时间（软删除标记）';

CREATE INDEX idx_supplier_name ON public.supplier USING btree (supplier_name) WHERE (deleted_at IS NULL);
CREATE INDEX idx_supplier_qualified ON public.supplier USING btree (is_qualified, qualification_expire) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX uk_supplier_code ON public.supplier USING btree (supplier_code) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX supplier_pkey ON public.supplier USING btree (id);

-- 表: supplier_certificate
-- 创建时间: 2026-04-29
CREATE TABLE supplier_certificate (
    id bigint NOT NULL DEFAULT nextval('supplier_certificate_id_seq'::regclass),
    supplier_id bigint NOT NULL,
    cert_type character varying(50) NOT NULL,
    cert_name character varying(200) NOT NULL,
    cert_no character varying(100),
    issue_date date,
    expire_date date,
    file_path character varying(500),
    remark text,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    updated_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    supplier_code character varying(20)
);

COMMENT ON TABLE supplier_certificate IS '供应商资质证书表';

COMMENT ON COLUMN supplier_certificate.id IS '证书ID，主键';
COMMENT ON COLUMN supplier_certificate.supplier_id IS '供应商ID';
COMMENT ON COLUMN supplier_certificate.cert_type IS '资质类型';
COMMENT ON COLUMN supplier_certificate.cert_name IS '资质名称';
COMMENT ON COLUMN supplier_certificate.cert_no IS '证书编号';
COMMENT ON COLUMN supplier_certificate.issue_date IS '发证日期';
COMMENT ON COLUMN supplier_certificate.expire_date IS '到期日期';
COMMENT ON COLUMN supplier_certificate.file_path IS '证书附件路径';
COMMENT ON COLUMN supplier_certificate.remark IS '备注';
COMMENT ON COLUMN supplier_certificate.created_at IS '创建时间';
COMMENT ON COLUMN supplier_certificate.updated_at IS '更新时间';

CREATE INDEX idx_supplier_cert_expire ON public.supplier_certificate USING btree (expire_date);
CREATE INDEX idx_supplier_cert_supplier ON public.supplier_certificate USING btree (supplier_id);
CREATE UNIQUE INDEX supplier_certificate_pkey ON public.supplier_certificate USING btree (id);

-- 表: sys_audit_log
-- 创建时间: 2026-04-29
CREATE TABLE sys_audit_log (
    id bigint NOT NULL DEFAULT nextval('sys_audit_log_id_seq'::regclass),
    user_id bigint NOT NULL,
    username character varying(50) NOT NULL,
    action character varying(50) NOT NULL,
    module character varying(50) NOT NULL,
    target_type character varying(50),
    target_id bigint,
    detail jsonb,
    ip_address character varying(50),
    user_agent character varying(500),
    created_at timestamp(6) with time zone NOT NULL DEFAULT now()
);

COMMENT ON TABLE sys_audit_log IS '操作审计日志表';

COMMENT ON COLUMN sys_audit_log.id IS '日志ID，主键';
COMMENT ON COLUMN sys_audit_log.user_id IS '操作用户ID';
COMMENT ON COLUMN sys_audit_log.username IS '操作用户名';
COMMENT ON COLUMN sys_audit_log.action IS '操作类型: CREATE/UPDATE/DELETE/LOGIN/LOGOUT/CONFIRM/APPROVE';
COMMENT ON COLUMN sys_audit_log.module IS '模块名称';
COMMENT ON COLUMN sys_audit_log.target_type IS '操作目标类型';
COMMENT ON COLUMN sys_audit_log.target_id IS '操作目标ID';
COMMENT ON COLUMN sys_audit_log.detail IS '操作详情（JSON格式，含变更前后对比）';
COMMENT ON COLUMN sys_audit_log.ip_address IS '客户端IP';
COMMENT ON COLUMN sys_audit_log.user_agent IS '客户端UserAgent';
COMMENT ON COLUMN sys_audit_log.created_at IS '操作时间';

CREATE INDEX idx_audit_log_created ON public.sys_audit_log USING btree (created_at);
CREATE INDEX idx_audit_log_module ON public.sys_audit_log USING btree (module);
CREATE INDEX idx_audit_log_user ON public.sys_audit_log USING btree (user_id);
CREATE UNIQUE INDEX sys_audit_log_pkey ON public.sys_audit_log USING btree (id);

-- 表: sys_batch_seq
-- 创建时间: 2026-04-29
CREATE TABLE sys_batch_seq (
    id bigint NOT NULL DEFAULT nextval('sys_batch_seq_id_seq'::regclass),
    date_str character varying(8) NOT NULL,
    current_seq integer NOT NULL DEFAULT 0,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

COMMENT ON TABLE sys_batch_seq IS '批次号序列表';

COMMENT ON COLUMN sys_batch_seq.date_str IS '日期字符串YYYYMMDD';
COMMENT ON COLUMN sys_batch_seq.current_seq IS '当前序号';

CREATE UNIQUE INDEX sys_batch_seq_pkey ON public.sys_batch_seq USING btree (id);
CREATE UNIQUE INDEX sys_batch_seq_date_str_key ON public.sys_batch_seq USING btree (date_str);

-- 表: sys_dict_data
-- 创建时间: 2026-04-29
CREATE TABLE sys_dict_data (
    id bigint NOT NULL DEFAULT nextval('sys_dict_data_id_seq'::regclass),
    dict_type character varying(50) NOT NULL,
    dict_value character varying(100) NOT NULL,
    dict_label character varying(200) NOT NULL,
    sort_order integer DEFAULT 0,
    is_default boolean NOT NULL DEFAULT false,
    remark text,
    status character varying(20) NOT NULL DEFAULT 'enabled'::character varying,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    updated_at timestamp(6) with time zone NOT NULL DEFAULT now()
);

COMMENT ON TABLE sys_dict_data IS '数据字典数据表';

COMMENT ON COLUMN sys_dict_data.id IS '字典数据ID，主键';
COMMENT ON COLUMN sys_dict_data.dict_type IS '字典类型编码';
COMMENT ON COLUMN sys_dict_data.dict_value IS '字典数据值';
COMMENT ON COLUMN sys_dict_data.dict_label IS '字典数据标签（显示名）';
COMMENT ON COLUMN sys_dict_data.sort_order IS '排序号';
COMMENT ON COLUMN sys_dict_data.is_default IS '是否默认值';
COMMENT ON COLUMN sys_dict_data.remark IS '备注';
COMMENT ON COLUMN sys_dict_data.status IS '状态: enabled-启用, disabled-禁用';
COMMENT ON COLUMN sys_dict_data.created_at IS '创建时间';
COMMENT ON COLUMN sys_dict_data.updated_at IS '更新时间';

CREATE INDEX idx_dict_data_type ON public.sys_dict_data USING btree (dict_type);
CREATE UNIQUE INDEX uk_dict_data_type_val ON public.sys_dict_data USING btree (dict_type, dict_value);
CREATE UNIQUE INDEX sys_dict_data_pkey ON public.sys_dict_data USING btree (id);

-- 表: sys_dict_type
-- 创建时间: 2026-04-29
CREATE TABLE sys_dict_type (
    id bigint NOT NULL DEFAULT nextval('sys_dict_type_id_seq'::regclass),
    dict_type character varying(50) NOT NULL,
    dict_name character varying(100) NOT NULL,
    remark text,
    status character varying(20) NOT NULL DEFAULT 'enabled'::character varying,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    updated_at timestamp(6) with time zone NOT NULL DEFAULT now()
);

COMMENT ON TABLE sys_dict_type IS '数据字典类型表';

COMMENT ON COLUMN sys_dict_type.id IS '字典类型ID，主键';
COMMENT ON COLUMN sys_dict_type.dict_type IS '字典类型编码，唯一';
COMMENT ON COLUMN sys_dict_type.dict_name IS '字典类型名称';
COMMENT ON COLUMN sys_dict_type.remark IS '备注';
COMMENT ON COLUMN sys_dict_type.status IS '状态: enabled-启用, disabled-禁用';
COMMENT ON COLUMN sys_dict_type.created_at IS '创建时间';
COMMENT ON COLUMN sys_dict_type.updated_at IS '更新时间';

CREATE UNIQUE INDEX sys_dict_type_dict_type_key ON public.sys_dict_type USING btree (dict_type);
CREATE UNIQUE INDEX sys_dict_type_pkey ON public.sys_dict_type USING btree (id);

-- 表: sys_material_code_seq
-- 创建时间: 2026-04-29
CREATE TABLE sys_material_code_seq (
    id bigint NOT NULL DEFAULT nextval('sys_material_code_seq_id_seq'::regclass),
    prefix character varying(10) NOT NULL,
    year integer NOT NULL,
    current_seq integer NOT NULL DEFAULT 0,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

COMMENT ON TABLE sys_material_code_seq IS '材料编码序列表';

COMMENT ON COLUMN sys_material_code_seq.prefix IS '编码前缀';
COMMENT ON COLUMN sys_material_code_seq.year IS '年度';
COMMENT ON COLUMN sys_material_code_seq.current_seq IS '当前序号';

CREATE UNIQUE INDEX sys_material_code_seq_pkey ON public.sys_material_code_seq USING btree (id);
CREATE UNIQUE INDEX sys_material_code_seq_prefix_year_key ON public.sys_material_code_seq USING btree (prefix, year);

-- 表: sys_notification
-- 创建时间: 2026-04-29
CREATE TABLE sys_notification (
    id bigint NOT NULL DEFAULT nextval('sys_notification_id_seq'::regclass),
    user_id bigint NOT NULL,
    title character varying(200) NOT NULL,
    content text NOT NULL,
    notify_type character varying(50) NOT NULL,
    ref_type character varying(50),
    ref_id bigint,
    is_read boolean NOT NULL DEFAULT false,
    read_at timestamp(6) with time zone,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now()
);

COMMENT ON TABLE sys_notification IS '系统通知表';

COMMENT ON COLUMN sys_notification.id IS '通知ID，主键';
COMMENT ON COLUMN sys_notification.user_id IS '接收用户ID';
COMMENT ON COLUMN sys_notification.title IS '通知标题';
COMMENT ON COLUMN sys_notification.content IS '通知内容';
COMMENT ON COLUMN sys_notification.notify_type IS '通知类型: stock_alert/approval/system';
COMMENT ON COLUMN sys_notification.ref_type IS '关联对象类型';
COMMENT ON COLUMN sys_notification.ref_id IS '关联对象ID';
COMMENT ON COLUMN sys_notification.is_read IS '是否已读';
COMMENT ON COLUMN sys_notification.read_at IS '已读时间';
COMMENT ON COLUMN sys_notification.created_at IS '创建时间';

CREATE INDEX idx_notification_created ON public.sys_notification USING btree (created_at);
CREATE INDEX idx_notification_user ON public.sys_notification USING btree (user_id, is_read);
CREATE UNIQUE INDEX sys_notification_pkey ON public.sys_notification USING btree (id);

-- 表: sys_permission
-- 创建时间: 2026-04-29
CREATE TABLE sys_permission (
    id bigint NOT NULL DEFAULT nextval('sys_permission_id_seq'::regclass),
    parent_id bigint DEFAULT 0,
    perm_code character varying(100) NOT NULL,
    perm_name character varying(100) NOT NULL,
    perm_type character varying(20) NOT NULL,
    path character varying(200),
    icon character varying(100),
    sort_order integer DEFAULT 0,
    status character varying(20) NOT NULL DEFAULT 'enabled'::character varying,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    updated_at timestamp(6) with time zone NOT NULL DEFAULT now()
);

COMMENT ON TABLE sys_permission IS '系统权限表';

COMMENT ON COLUMN sys_permission.id IS '权限ID，主键';
COMMENT ON COLUMN sys_permission.parent_id IS '父权限ID，0为顶级';
COMMENT ON COLUMN sys_permission.perm_code IS '权限编码，如 material:create';
COMMENT ON COLUMN sys_permission.perm_name IS '权限名称';
COMMENT ON COLUMN sys_permission.perm_type IS '权限类型: menu-菜单, button-按钮, api-接口';
COMMENT ON COLUMN sys_permission.path IS '前端路由路径';
COMMENT ON COLUMN sys_permission.icon IS '菜单图标';
COMMENT ON COLUMN sys_permission.sort_order IS '排序号';
COMMENT ON COLUMN sys_permission.status IS '状态: enabled-启用, disabled-禁用';
COMMENT ON COLUMN sys_permission.created_at IS '创建时间';
COMMENT ON COLUMN sys_permission.updated_at IS '更新时间';

CREATE UNIQUE INDEX sys_permission_perm_code_key ON public.sys_permission USING btree (perm_code);
CREATE UNIQUE INDEX sys_permission_pkey ON public.sys_permission USING btree (id);

-- 表: sys_role
-- 创建时间: 2026-04-29
CREATE TABLE sys_role (
    id bigint NOT NULL DEFAULT nextval('sys_role_id_seq'::regclass),
    role_code character varying(50) NOT NULL,
    role_name character varying(100) NOT NULL,
    description text,
    status character varying(20) NOT NULL DEFAULT 'enabled'::character varying,
    created_by bigint,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    updated_by bigint,
    updated_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp(6) with time zone
);

COMMENT ON TABLE sys_role IS '系统角色表';

COMMENT ON COLUMN sys_role.id IS '角色ID，主键';
COMMENT ON COLUMN sys_role.role_code IS '角色编码，唯一';
COMMENT ON COLUMN sys_role.role_name IS '角色名称';
COMMENT ON COLUMN sys_role.description IS '角色描述';
COMMENT ON COLUMN sys_role.status IS '状态: enabled-启用, disabled-禁用';
COMMENT ON COLUMN sys_role.created_by IS '创建人ID';
COMMENT ON COLUMN sys_role.created_at IS '创建时间';
COMMENT ON COLUMN sys_role.updated_by IS '更新人ID';
COMMENT ON COLUMN sys_role.updated_at IS '更新时间';
COMMENT ON COLUMN sys_role.deleted_at IS '删除时间（软删除标记）';

CREATE UNIQUE INDEX uk_sys_role_code ON public.sys_role USING btree (role_code) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX sys_role_pkey ON public.sys_role USING btree (id);

-- 表: sys_role_permission
-- 创建时间: 2026-04-29
CREATE TABLE sys_role_permission (
    id bigint NOT NULL DEFAULT nextval('sys_role_permission_id_seq'::regclass),
    role_id bigint NOT NULL,
    permission_id bigint NOT NULL,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now()
);

COMMENT ON TABLE sys_role_permission IS '角色权限关联表';

COMMENT ON COLUMN sys_role_permission.id IS '关联ID，主键';
COMMENT ON COLUMN sys_role_permission.role_id IS '角色ID';
COMMENT ON COLUMN sys_role_permission.permission_id IS '权限ID';
COMMENT ON COLUMN sys_role_permission.created_at IS '创建时间';

CREATE INDEX idx_sys_role_perm_role ON public.sys_role_permission USING btree (role_id);
CREATE UNIQUE INDEX sys_role_permission_role_id_permission_id_key ON public.sys_role_permission USING btree (role_id, permission_id);
CREATE UNIQUE INDEX sys_role_permission_pkey ON public.sys_role_permission USING btree (id);

-- 表: sys_serial_number
-- 创建时间: 2026-04-29
CREATE TABLE sys_serial_number (
    id bigint NOT NULL DEFAULT nextval('sys_serial_number_id_seq'::regclass),
    prefix character varying(10) NOT NULL,
    date_str character varying(8) NOT NULL,
    current_seq integer NOT NULL DEFAULT 1,
    updated_at timestamp(6) with time zone NOT NULL DEFAULT now()
);

COMMENT ON TABLE sys_serial_number IS '单据编号序列表（供 fn_generate_serial_no 使用）';

COMMENT ON COLUMN sys_serial_number.id IS '序列ID，主键';
COMMENT ON COLUMN sys_serial_number.prefix IS '编号前缀（如PO/SO/WI/WO）';
COMMENT ON COLUMN sys_serial_number.date_str IS '日期字符串（YYYYMMDD）';
COMMENT ON COLUMN sys_serial_number.current_seq IS '当前序号';
COMMENT ON COLUMN sys_serial_number.updated_at IS '更新时间';

CREATE UNIQUE INDEX sys_serial_number_prefix_date_str_key ON public.sys_serial_number USING btree (prefix, date_str);
CREATE UNIQUE INDEX sys_serial_number_pkey ON public.sys_serial_number USING btree (id);

-- 表: sys_sku_code_seq
-- 创建时间: 2026-04-29
CREATE TABLE sys_sku_code_seq (
    id bigint NOT NULL DEFAULT nextval('sys_sku_code_seq_id_seq'::regclass),
    material_id bigint NOT NULL,
    current_seq integer NOT NULL DEFAULT 0,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

COMMENT ON TABLE sys_sku_code_seq IS 'SKU编码序列表';

COMMENT ON COLUMN sys_sku_code_seq.material_id IS '物料ID';
COMMENT ON COLUMN sys_sku_code_seq.current_seq IS '当前序号';

CREATE UNIQUE INDEX sys_sku_code_seq_pkey ON public.sys_sku_code_seq USING btree (id);
CREATE UNIQUE INDEX sys_sku_code_seq_material_id_key ON public.sys_sku_code_seq USING btree (material_id);

-- 表: sys_user
-- 创建时间: 2026-04-29
CREATE TABLE sys_user (
    id bigint NOT NULL DEFAULT nextval('sys_user_id_seq'::regclass),
    username character varying(50) NOT NULL,
    password_hash character varying(255) NOT NULL,
    real_name character varying(50) NOT NULL,
    phone character varying(20),
    email character varying(100),
    department character varying(100),
    status character varying(20) NOT NULL DEFAULT 'enabled'::character varying,
    last_login_at timestamp(6) with time zone,
    last_login_ip character varying(50),
    created_by bigint,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    updated_by bigint,
    updated_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp(6) with time zone
);

COMMENT ON TABLE sys_user IS '系统用户表';

COMMENT ON COLUMN sys_user.id IS '用户ID，主键';
COMMENT ON COLUMN sys_user.username IS '登录用户名，唯一';
COMMENT ON COLUMN sys_user.password_hash IS '密码哈希值(bcrypt)';
COMMENT ON COLUMN sys_user.real_name IS '真实姓名';
COMMENT ON COLUMN sys_user.phone IS '手机号码';
COMMENT ON COLUMN sys_user.email IS '电子邮箱';
COMMENT ON COLUMN sys_user.department IS '所属部门';
COMMENT ON COLUMN sys_user.status IS '状态: enabled-启用, disabled-禁用';
COMMENT ON COLUMN sys_user.last_login_at IS '最后登录时间';
COMMENT ON COLUMN sys_user.last_login_ip IS '最后登录IP';
COMMENT ON COLUMN sys_user.created_by IS '创建人ID';
COMMENT ON COLUMN sys_user.created_at IS '创建时间';
COMMENT ON COLUMN sys_user.updated_by IS '更新人ID';
COMMENT ON COLUMN sys_user.updated_at IS '更新时间';
COMMENT ON COLUMN sys_user.deleted_at IS '删除时间（软删除标记）';

CREATE INDEX idx_sys_user_status ON public.sys_user USING btree (status) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX uk_sys_user_username ON public.sys_user USING btree (username) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX sys_user_pkey ON public.sys_user USING btree (id);

-- 表: sys_user_role
-- 创建时间: 2026-04-29
CREATE TABLE sys_user_role (
    id bigint NOT NULL DEFAULT nextval('sys_user_role_id_seq'::regclass),
    user_id bigint NOT NULL,
    role_id bigint NOT NULL,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now()
);

COMMENT ON TABLE sys_user_role IS '用户角色关联表';

COMMENT ON COLUMN sys_user_role.id IS '关联ID，主键';
COMMENT ON COLUMN sys_user_role.user_id IS '用户ID';
COMMENT ON COLUMN sys_user_role.role_id IS '角色ID';
COMMENT ON COLUMN sys_user_role.created_at IS '创建时间';

CREATE INDEX idx_sys_user_role_role ON public.sys_user_role USING btree (role_id);
CREATE INDEX idx_sys_user_role_user ON public.sys_user_role USING btree (user_id);
CREATE UNIQUE INDEX sys_user_role_user_id_role_id_key ON public.sys_user_role USING btree (user_id, role_id);
CREATE UNIQUE INDEX sys_user_role_pkey ON public.sys_user_role USING btree (id);

-- 表: warehouse
-- 创建时间: 2026-04-29
CREATE TABLE warehouse (
    id bigint NOT NULL DEFAULT nextval('warehouse_id_seq'::regclass),
    warehouse_code character varying(10) NOT NULL,
    warehouse_name character varying(50) NOT NULL,
    warehouse_type character varying(30) NOT NULL,
    manager_id bigint NOT NULL,
    status character varying(20) NOT NULL DEFAULT 'enabled'::character varying,
    created_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    updated_at timestamp(6) with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp(6) with time zone
);

COMMENT ON TABLE warehouse IS '仓库表';

COMMENT ON COLUMN warehouse.id IS '仓库ID，主键';
COMMENT ON COLUMN warehouse.warehouse_code IS '仓库编码，唯一';
COMMENT ON COLUMN warehouse.warehouse_name IS '仓库名称';
COMMENT ON COLUMN warehouse.warehouse_type IS '仓库类型: raw_material/welding/semi_finished/finished/scrap';
COMMENT ON COLUMN warehouse.manager_id IS '仓库负责人ID';
COMMENT ON COLUMN warehouse.status IS '状态: enabled-启用, disabled-禁用';
COMMENT ON COLUMN warehouse.created_at IS '创建时间';
COMMENT ON COLUMN warehouse.updated_at IS '更新时间';
COMMENT ON COLUMN warehouse.deleted_at IS '删除时间（软删除标记）';

CREATE UNIQUE INDEX uk_warehouse_code ON public.warehouse USING btree (warehouse_code) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX warehouse_pkey ON public.warehouse USING btree (id);

-- =====================================================
-- 视图定义
-- =====================================================

-- 视图: v_consumption_order_list
DROP VIEW IF EXISTS v_consumption_order_list;
CREATE VIEW v_consumption_order_list AS SELECT co.id,
    co.order_no,
    co.project_no,
    co.product_name,
    COALESCE(( SELECT coi2.warehouse_id
           FROM consumption_order_item coi2
          WHERE ((coi2.order_id = co.id) AND (coi2.warehouse_id IS NOT NULL))
          ORDER BY coi2.id DESC
         LIMIT 1), ( SELECT i.warehouse_id
           FROM (consumption_order_item coi2
             JOIN inventory i ON ((i.id = coi2.inventory_id)))
          WHERE (coi2.order_id = co.id)
          ORDER BY coi2.id
         LIMIT 1), (0)::bigint) AS warehouse_id,
    COALESCE(( SELECT coi2.warehouse_code
           FROM consumption_order_item coi2
          WHERE ((coi2.order_id = co.id) AND (coi2.warehouse_code IS NOT NULL) AND (btrim((coi2.warehouse_code)::text) <> ''::text))
          ORDER BY coi2.id DESC
         LIMIT 1), ( SELECT w2.warehouse_code
           FROM ((consumption_order_item coi2
             JOIN inventory i ON ((i.id = coi2.inventory_id)))
             JOIN warehouse w2 ON (((w2.id = i.warehouse_id) AND (w2.deleted_at IS NULL))))
          WHERE (coi2.order_id = co.id)
          ORDER BY coi2.id
         LIMIT 1), ''::character varying) AS warehouse_code,
    COALESCE(( SELECT w3.warehouse_name
           FROM (consumption_order_item coi2
             JOIN warehouse w3 ON (((w3.id = coi2.warehouse_id) AND (w3.deleted_at IS NULL))))
          WHERE ((coi2.order_id = co.id) AND (coi2.warehouse_id IS NOT NULL))
          ORDER BY coi2.id DESC
         LIMIT 1), ( SELECT w4.warehouse_name
           FROM ((consumption_order_item coi2
             JOIN inventory i ON ((i.id = coi2.inventory_id)))
             JOIN warehouse w4 ON (((w4.id = i.warehouse_id) AND (w4.deleted_at IS NULL))))
          WHERE (coi2.order_id = co.id)
          ORDER BY coi2.id
         LIMIT 1), ''::character varying) AS warehouse_name,
    co.order_date,
    co.designer_id,
    co.designer_name,
    co.status,
    co.stock_out_id,
    so.stock_out_no,
    co.remark,
    co.created_at,
    co.updated_at,
    count(coi.id) AS item_count,
    sum(coi.quantity) AS total_quantity
   FROM ((consumption_order co
     LEFT JOIN stock_out so ON ((so.id = co.stock_out_id)))
     LEFT JOIN consumption_order_item coi ON ((coi.order_id = co.id)))
  WHERE (co.deleted_at IS NULL)
  GROUP BY co.id, co.order_no, co.project_no, co.product_name, co.order_date, co.designer_id, co.designer_name, co.status, co.stock_out_id, so.stock_out_no, co.remark, co.created_at, co.updated_at;

-- 视图: v_customer_list
DROP VIEW IF EXISTS v_customer_list;
CREATE VIEW v_customer_list AS SELECT c.id,
    c.customer_code,
    c.customer_name,
    c.credit_code,
    c.customer_type,
    fn_dict_label('customer_type'::character varying, c.customer_type) AS customer_type_name,
    c.contact_person,
    c.contact_phone,
    c.address,
    c.sales_person_id,
    u.real_name AS sales_person_name,
    c.remark,
    c.status,
    fn_dict_label('common_status'::character varying, c.status) AS status_name,
    c.created_at,
    c.updated_at
   FROM (customer c
     LEFT JOIN sys_user u ON ((u.id = c.sales_person_id)))
  WHERE (c.deleted_at IS NULL);

-- 视图: v_dashboard_trend
DROP VIEW IF EXISTS v_dashboard_trend;
CREATE VIEW v_dashboard_trend AS SELECT d.trans_date,
    COALESCE(sum(
        CASE
            WHEN ((it.trans_type)::text = 'in'::text) THEN it.quantity
            ELSE (0)::numeric
        END), (0)::numeric) AS in_quantity,
    COALESCE(sum(
        CASE
            WHEN ((it.trans_type)::text = 'out'::text) THEN abs(it.quantity)
            ELSE (0)::numeric
        END), (0)::numeric) AS out_quantity,
    count(
        CASE
            WHEN ((it.trans_type)::text = 'in'::text) THEN 1
            ELSE NULL::integer
        END) AS in_count,
    count(
        CASE
            WHEN ((it.trans_type)::text = 'out'::text) THEN 1
            ELSE NULL::integer
        END) AS out_count
   FROM (generate_series((CURRENT_DATE - '29 days'::interval), (CURRENT_DATE)::timestamp without time zone, '1 day'::interval) d(trans_date)
     LEFT JOIN inventory_transaction it ON ((((it.created_at)::date = d.trans_date) AND ((it.trans_type)::text = ANY (ARRAY[('in'::character varying)::text, ('out'::character varying)::text])))))
  GROUP BY d.trans_date
  ORDER BY d.trans_date;

-- 视图: v_in_out_summary
DROP VIEW IF EXISTS v_in_out_summary;
CREATE VIEW v_in_out_summary AS SELECT (date_trunc('day'::text, created_at))::date AS trans_date,
    warehouse_id,
    material_id,
    sum(
        CASE
            WHEN ((trans_type)::text = 'in'::text) THEN quantity
            ELSE (0)::numeric
        END) AS in_qty,
    sum(
        CASE
            WHEN ((trans_type)::text = 'out'::text) THEN (- quantity)
            ELSE (0)::numeric
        END) AS out_qty
   FROM inventory_transaction
  GROUP BY ((date_trunc('day'::text, created_at))::date), warehouse_id, material_id;

-- 视图: v_inventory_alert
DROP VIEW IF EXISTS v_inventory_alert;
CREATE VIEW v_inventory_alert AS SELECT m.id AS material_id,
    m.material_code,
    m.material_name,
    m.unit,
    m.safety_stock,
    m.max_stock,
    COALESCE(sum(i.quantity), (0)::numeric) AS current_quantity,
    (m.safety_stock - COALESCE(sum(i.quantity), (0)::numeric)) AS shortage_quantity
   FROM (material m
     LEFT JOIN inventory i ON ((m.id = i.material_id)))
  WHERE (((m.status)::text = 'enabled'::text) AND (m.deleted_at IS NULL))
  GROUP BY m.id, m.material_code, m.material_name, m.unit, m.safety_stock, m.max_stock
 HAVING (COALESCE(sum(i.quantity), (0)::numeric) < m.safety_stock);

-- 视图: v_inventory_detail
DROP VIEW IF EXISTS v_inventory_detail;
CREATE VIEW v_inventory_detail AS SELECT i.id,
    i.material_id,
    m.material_code,
    m.material_name,
    m.is_code,
    m.unit,
    i.warehouse_id,
    w.warehouse_code,
    w.warehouse_name,
    i.sku_id,
    s.sku_code,
    s.sku_name,
    i.quantity,
    i.locked_quantity,
    (i.quantity - i.locked_quantity) AS available_quantity,
    i.unit_cost,
    i.stock_in_date,
    i.material_inspection_no,
    i.updated_at AS created_at,
    i.updated_at,
    NULL::text AS cert_urls
   FROM (((inventory i
     JOIN material m ON ((i.material_id = m.id)))
     JOIN warehouse w ON ((i.warehouse_id = w.id)))
     LEFT JOIN material_sku s ON ((i.sku_id = s.id)));

-- 视图: v_inventory_report
DROP VIEW IF EXISTS v_inventory_report;
CREATE VIEW v_inventory_report AS SELECT w.warehouse_name,
    mc.category_name,
    m.material_code,
    m.material_name,
    m.unit,
    sum(i.quantity) AS current_quantity,
    sum(i.locked_quantity) AS locked_quantity,
    (sum(i.quantity) - sum(i.locked_quantity)) AS available_quantity
   FROM (((inventory i
     JOIN material m ON ((i.material_id = m.id)))
     LEFT JOIN material_category mc ON ((m.category_id = mc.id)))
     JOIN warehouse w ON ((i.warehouse_id = w.id)))
  GROUP BY w.warehouse_name, mc.category_name, m.material_code, m.material_name, m.unit;

-- 视图: v_inventory_summary
DROP VIEW IF EXISTS v_inventory_summary;
CREATE VIEW v_inventory_summary AS SELECT m.id AS material_id,
    m.material_code,
    m.material_name,
    mc.category_name,
    m.unit,
    m.safety_stock,
    m.max_stock,
    COALESCE(sum(i.quantity), (0)::numeric) AS total_quantity,
    COALESCE(sum(i.locked_quantity), (0)::numeric) AS locked_quantity,
    COALESCE(sum(COALESCE(i.in_transit_quantity, (0)::numeric)), (0)::numeric) AS in_transit_quantity,
    COALESCE(sum(((i.quantity - i.locked_quantity) - COALESCE(i.in_transit_quantity, (0)::numeric))), (0)::numeric) AS available_quantity,
    COALESCE(count(i.id), (0)::bigint) AS batch_count
   FROM ((material m
     LEFT JOIN material_category mc ON ((mc.id = m.category_id)))
     LEFT JOIN inventory i ON (((i.material_id = m.id) AND (i.quantity > (0)::numeric))))
  WHERE (m.deleted_at IS NULL)
  GROUP BY m.id, m.material_code, m.material_name, mc.category_name, m.unit, m.safety_stock, m.max_stock;

-- 视图: v_material_list
DROP VIEW IF EXISTS v_material_list;
CREATE VIEW v_material_list AS SELECT m.id,
    m.category_id,
    c.category_name,
    m.material_code,
    m.material_name,
    m.unit,
    fn_dict_label('unit'::character varying, m.unit) AS unit_name,
    m.safety_stock,
    m.max_stock,
    m.is_code,
    m.sku_managed,
    m.custom_attributes,
    m.default_warehouse_id,
    w.warehouse_name AS default_warehouse_name,
    m.status,
    fn_dict_label('common_status'::character varying, m.status) AS status_name,
    m.remark,
    m.created_at,
    m.updated_at,
    ( SELECT count(1) AS count
           FROM material_sku s
          WHERE ((s.material_id = m.id) AND (s.deleted_at IS NULL))) AS sku_count
   FROM ((material m
     LEFT JOIN material_category c ON ((m.category_id = c.id)))
     LEFT JOIN warehouse w ON ((m.default_warehouse_id = w.id)))
  WHERE (m.deleted_at IS NULL);

-- 视图: v_purchase_order_list
DROP VIEW IF EXISTS v_purchase_order_list;
CREATE VIEW v_purchase_order_list AS SELECT po.id,
    po.order_no,
    po.request_id,
    po.supplier_code,
    s.supplier_name,
    po.buyer_id,
    u.real_name AS buyer_name,
    po.order_date,
    po.expected_date,
    po.payment_method,
    fn_dict_label('payment_method'::character varying, po.payment_method) AS payment_method_name,
    po.order_status,
    fn_dict_label('purchase_order_status'::character varying, po.order_status) AS order_status_name,
    po.total_amount,
    po.remark,
    po.created_at,
    po.updated_at
   FROM ((purchase_order po
     JOIN supplier s ON (((s.supplier_code)::text = (po.supplier_code)::text)))
     LEFT JOIN sys_user u ON ((u.id = po.buyer_id)))
  WHERE (po.deleted_at IS NULL);

-- 视图: v_purchase_order_tracking
DROP VIEW IF EXISTS v_purchase_order_tracking;
CREATE VIEW v_purchase_order_tracking AS SELECT po.id AS order_id,
    po.order_no,
    po.supplier_code,
    s.supplier_name,
    po.order_date,
    po.expected_date,
    po.order_status,
    po.total_amount,
    sum(poi.quantity) AS total_quantity,
    sum(poi.received_quantity) AS received_quantity,
    sum((poi.quantity - poi.received_quantity)) AS pending_quantity,
        CASE
            WHEN (sum(poi.received_quantity) = (0)::numeric) THEN '未到货'::text
            WHEN (sum(poi.received_quantity) < sum(poi.quantity)) THEN '部分到货'::text
            ELSE '全部到货'::text
        END AS receive_status
   FROM ((purchase_order po
     JOIN supplier s ON (((s.supplier_code)::text = (po.supplier_code)::text)))
     JOIN purchase_order_item poi ON ((po.id = poi.order_id)))
  WHERE (po.deleted_at IS NULL)
  GROUP BY po.id, po.order_no, po.supplier_code, s.supplier_name, po.order_date, po.expected_date, po.order_status, po.total_amount;

-- 视图: v_report_inventory_status
DROP VIEW IF EXISTS v_report_inventory_status;
CREATE VIEW v_report_inventory_status AS SELECT w.warehouse_name,
    mc.category_name,
    m.material_code,
    m.material_name,
    m.unit,
    sum(i.quantity) AS current_quantity,
    sum(i.locked_quantity) AS locked_quantity,
    sum((i.quantity - i.locked_quantity)) AS available_quantity
   FROM (((inventory i
     JOIN material m ON ((m.id = i.material_id)))
     JOIN material_category mc ON ((mc.id = m.category_id)))
     JOIN warehouse w ON ((w.id = i.warehouse_id)))
  GROUP BY w.warehouse_name, mc.category_name, m.material_code, m.material_name, m.unit;

-- 视图: v_report_stock_in_summary
DROP VIEW IF EXISTS v_report_stock_in_summary;
CREATE VIEW v_report_stock_in_summary AS SELECT to_char((si.stock_in_date)::timestamp with time zone, 'YYYY-MM'::text) AS report_month,
    COALESCE(s.supplier_name, ''::character varying) AS supplier_name,
    m.material_code,
    m.material_name,
    m.unit,
    count(DISTINCT si.id) AS order_count
   FROM (((stock_in si
     LEFT JOIN supplier s ON (((s.supplier_code)::text = (si.supplier_code)::text)))
     JOIN stock_in_item sii ON ((sii.stock_in_id = si.id)))
     JOIN material m ON ((m.id = sii.material_id)))
  WHERE ((si.deleted_at IS NULL) AND ((si.stock_in_status)::text = 'passed'::text))
  GROUP BY (to_char((si.stock_in_date)::timestamp with time zone, 'YYYY-MM'::text)), COALESCE(s.supplier_name, ''::character varying), m.material_code, m.material_name, m.unit;

-- 视图: v_report_stock_out_summary
DROP VIEW IF EXISTS v_report_stock_out_summary;
CREATE VIEW v_report_stock_out_summary AS SELECT to_char((so.stock_out_date)::timestamp with time zone, 'YYYY-MM'::text) AS report_month,
    so.out_type,
    m.material_code,
    m.material_name,
    m.unit,
    sum(soi.quantity) AS total_quantity,
    count(DISTINCT so.id) AS order_count
   FROM ((stock_out so
     JOIN stock_out_item soi ON ((soi.stock_out_id = so.id)))
     JOIN material m ON ((m.id = soi.material_id)))
  WHERE (so.deleted_at IS NULL)
  GROUP BY (to_char((so.stock_out_date)::timestamp with time zone, 'YYYY-MM'::text)), so.out_type, m.material_code, m.material_name, m.unit;

-- 视图: v_reversal_order_list
DROP VIEW IF EXISTS v_reversal_order_list;
CREATE VIEW v_reversal_order_list AS SELECT ro.id,
    ro.order_no,
    ro.project_no,
    ro.product_name,
    ro.warehouse_id,
    ro.warehouse_code,
    w.warehouse_name,
    ro.order_date,
    ro.designer_id,
    ro.designer_name,
    ro.status,
    ro.stock_in_id,
    si.stock_in_no,
    ro.remark,
    ro.created_at,
    ro.updated_at,
    count(roi.id) AS item_count,
    sum(roi.quantity) AS total_quantity
   FROM (((reversal_order ro
     LEFT JOIN warehouse w ON ((w.id = ro.warehouse_id)))
     LEFT JOIN stock_in si ON ((si.id = ro.stock_in_id)))
     LEFT JOIN reversal_order_item roi ON ((roi.order_id = ro.id)))
  WHERE (ro.deleted_at IS NULL)
  GROUP BY ro.id, ro.order_no, ro.project_no, ro.product_name, ro.warehouse_id, ro.warehouse_code, w.warehouse_name, ro.order_date, ro.designer_id, ro.designer_name, ro.status, ro.stock_in_id, si.stock_in_no, ro.remark, ro.created_at, ro.updated_at;

-- 视图: v_role_list
DROP VIEW IF EXISTS v_role_list;
CREATE VIEW v_role_list AS SELECT id,
    role_code,
    role_name,
    description,
    status,
    fn_dict_label('common_status'::character varying, status) AS status_name,
    ( SELECT count(*) AS count
           FROM (sys_user_role ur
             JOIN sys_user u ON (((u.id = ur.user_id) AND (u.deleted_at IS NULL))))
          WHERE (ur.role_id = r.id)) AS user_count,
    created_at,
    updated_at
   FROM sys_role r
  WHERE (deleted_at IS NULL);

-- 视图: v_sales_order_list
DROP VIEW IF EXISTS v_sales_order_list;
CREATE VIEW v_sales_order_list AS SELECT so.id,
    so.order_no,
    so.customer_code,
    c.customer_name,
    so.sales_person_id,
    u.real_name AS sales_person_name,
    so.order_date,
    so.delivery_date,
    so.order_status,
    fn_dict_label('sales_order_status'::character varying, so.order_status) AS order_status_name,
    so.total_amount,
    so.remark,
    so.created_at,
    so.updated_at
   FROM ((sales_order so
     JOIN customer c ON (((c.customer_code)::text = (so.customer_code)::text)))
     LEFT JOIN sys_user u ON ((u.id = so.sales_person_id)))
  WHERE (so.deleted_at IS NULL);

-- 视图: v_sales_order_tracking
DROP VIEW IF EXISTS v_sales_order_tracking;
CREATE VIEW v_sales_order_tracking AS SELECT so.id,
    so.order_no,
    so.order_date,
    c.customer_name,
    u.real_name AS sales_name,
    so.order_status,
    so.total_amount,
    so.delivery_date,
    count(soi.id) AS item_count,
    sum(soi.quantity) AS total_quantity,
    sum(soi.shipped_quantity) AS total_shipped
   FROM (((sales_order so
     JOIN customer c ON (((c.customer_code)::text = (so.customer_code)::text)))
     LEFT JOIN sys_user u ON ((u.id = so.sales_person_id)))
     LEFT JOIN sales_order_item soi ON ((soi.order_id = so.id)))
  WHERE (so.deleted_at IS NULL)
  GROUP BY so.id, so.order_no, so.order_date, c.customer_name, u.real_name, so.order_status, so.total_amount, so.delivery_date;

-- 视图: v_sku_list
DROP VIEW IF EXISTS v_sku_list;
CREATE VIEW v_sku_list AS SELECT s.id,
    s.material_id,
    m.material_code,
    m.material_name,
    m.unit,
    s.reference_price,
    mc.category_name,
    s.sku_code,
    s.sku_name,
    s.custom_attributes,
    ( SELECT string_agg(((((attr.value ->> 'attr_name'::text) || ':'::text) || (attr.value ->> 'attr_value'::text)) ||
                CASE
                    WHEN (((attr.value ->> 'attr_unit'::text) IS NOT NULL) AND ((attr.value ->> 'attr_unit'::text) <> ''::text)) THEN (attr.value ->> 'attr_unit'::text)
                    ELSE ''::text
                END), ' | '::text) AS string_agg
           FROM jsonb_array_elements(s.custom_attributes) attr(value)) AS attr_summary,
    s.status,
        CASE s.status
            WHEN 'enabled'::text THEN '启用'::text
            ELSE '禁用'::text
        END AS status_name,
    s.remark,
    s.created_by,
    u.real_name AS creator_name,
    s.created_at,
    s.updated_at
   FROM (((material_sku s
     JOIN material m ON ((m.id = s.material_id)))
     LEFT JOIN material_category mc ON ((mc.id = m.category_id)))
     LEFT JOIN sys_user u ON ((u.id = s.created_by)))
  WHERE ((s.deleted_at IS NULL) AND (m.deleted_at IS NULL));

-- 视图: v_stock_in_list
DROP VIEW IF EXISTS v_stock_in_list;
CREATE VIEW v_stock_in_list AS SELECT si.id,
    si.stock_in_no,
    si.purchase_order_id,
    po.order_no AS purchase_order_no,
    COALESCE(si.supplier_code, po.supplier_code) AS supplier_code,
    COALESCE(s.supplier_name, ''::character varying) AS supplier_name,
    si.warehouse_code,
    COALESCE(w.warehouse_name, ''::character varying) AS warehouse_name,
    si.stock_in_date,
    si.stock_in_status,
    si.remark,
    si.created_by,
    si.created_at,
    si.updated_at
   FROM (((stock_in si
     LEFT JOIN purchase_order po ON ((po.id = si.purchase_order_id)))
     LEFT JOIN supplier s ON (((s.supplier_code)::text = (COALESCE(si.supplier_code, po.supplier_code))::text)))
     LEFT JOIN warehouse w ON (((w.warehouse_code)::text = (si.warehouse_code)::text)))
  WHERE (si.deleted_at IS NULL);

-- 视图: v_stock_top_value
DROP VIEW IF EXISTS v_stock_top_value;
CREATE VIEW v_stock_top_value AS SELECT m.material_code,
    m.material_name,
    m.unit,
    sum(i.quantity) AS total_quantity,
    sum((i.quantity * i.unit_cost)) AS total_value
   FROM (inventory i
     JOIN material m ON ((m.id = i.material_id)))
  WHERE (i.quantity > (0)::numeric)
  GROUP BY m.material_code, m.material_name, m.unit
  ORDER BY (sum((i.quantity * i.unit_cost))) DESC;

-- 视图: v_supplier_list
DROP VIEW IF EXISTS v_supplier_list;
CREATE VIEW v_supplier_list AS SELECT id,
    supplier_code,
    supplier_name,
    credit_code,
    supplier_type,
    fn_dict_label('supplier_type'::character varying, supplier_type) AS supplier_type_name,
    contact_person,
    contact_phone,
    address,
    is_qualified,
    qualification_expire,
    supplier_rating,
    fn_dict_label('supplier_rating'::character varying, supplier_rating) AS supplier_rating_name,
    bank_name,
    bank_account,
    remark,
    status,
    fn_dict_label('common_status'::character varying, status) AS status_name,
    created_at,
    updated_at
   FROM supplier s
  WHERE (deleted_at IS NULL);

-- 视图: v_user_list
DROP VIEW IF EXISTS v_user_list;
CREATE VIEW v_user_list AS SELECT u.id,
    u.username,
    u.real_name,
    u.phone,
    u.email,
    u.department,
    u.status,
    fn_dict_label('common_status'::character varying, u.status) AS status_name,
    u.last_login_at,
    u.last_login_ip,
    u.created_at,
    u.updated_at,
    COALESCE(array_agg(r.role_name) FILTER (WHERE (r.role_name IS NOT NULL)), '{}'::character varying[]) AS role_names,
    COALESCE(array_agg(r.id) FILTER (WHERE (r.id IS NOT NULL)), '{}'::bigint[]) AS role_ids
   FROM ((sys_user u
     LEFT JOIN sys_user_role ur ON ((ur.user_id = u.id)))
     LEFT JOIN sys_role r ON (((r.id = ur.role_id) AND (r.deleted_at IS NULL))))
  WHERE (u.deleted_at IS NULL)
  GROUP BY u.id, u.username, u.real_name, u.phone, u.email, u.department, u.status, u.last_login_at, u.last_login_ip, u.created_at, u.updated_at;

-- 视图: v_warehouse_list
DROP VIEW IF EXISTS v_warehouse_list;
CREATE VIEW v_warehouse_list AS SELECT w.id,
    w.warehouse_code,
    w.warehouse_name,
    w.warehouse_type,
    fn_dict_label('warehouse_type'::character varying, w.warehouse_type) AS warehouse_type_name,
    w.manager_id,
    u.real_name AS manager_name,
    w.status,
    fn_dict_label('common_status'::character varying, w.status) AS status_name,
    w.created_at,
    w.updated_at
   FROM (warehouse w
     LEFT JOIN sys_user u ON ((u.id = w.manager_id)))
  WHERE (w.deleted_at IS NULL);

-- =====================================================
-- 函数定义
-- =====================================================

-- 函数: fn_calc_inventory_turnover
DROP FUNCTION IF EXISTS fn_calc_inventory_turnover CASCADE;
CREATE OR REPLACE FUNCTION public.fn_calc_inventory_turnover(p_start_date date, p_end_date date)
 RETURNS TABLE(material_id bigint, material_code character varying, material_name character varying, out_total numeric, avg_inventory numeric, turnover_rate numeric)
 LANGUAGE plpgsql
AS $function$
BEGIN
    RETURN QUERY
    WITH outs AS (
        SELECT it.material_id, SUM(-it.quantity) as total_out
        FROM inventory_transaction it
        WHERE it.trans_type = 'out' AND it.created_at::date BETWEEN p_start_date AND p_end_date
        GROUP BY it.material_id
    ),
    invs AS (
        -- 简化计算：平均库存 = (期初 + 期末) / 2
        -- 实际中可能需要更复杂的加权平均，此处演示核心逻辑
        SELECT i.material_id, AVG(i.quantity) as avg_qty
        FROM inventory i
        GROUP BY i.material_id
    )
    SELECT 
        m.id, m.material_code, m.material_name,
        COALESCE(outs.total_out, 0),
        COALESCE(invs.avg_qty, 0),
        CASE WHEN COALESCE(invs.avg_qty, 0) > 0 THEN COALESCE(outs.total_out, 0) / invs.avg_qty ELSE 0 END
    FROM material m
    LEFT JOIN outs ON outs.material_id = m.id
    LEFT JOIN invs ON invs.material_id = m.id
    WHERE m.deleted_at IS NULL;
END;
$function$;

-- 函数: fn_calc_inventory_value
DROP FUNCTION IF EXISTS fn_calc_inventory_value CASCADE;
CREATE OR REPLACE FUNCTION public.fn_calc_inventory_value(p_warehouse_id bigint DEFAULT NULL::bigint, p_category_id bigint DEFAULT NULL::bigint)
 RETURNS TABLE(total_value numeric, item_count bigint, material_count bigint)
 LANGUAGE plpgsql
AS $function$
BEGIN
    RETURN QUERY
    SELECT
        COALESCE(SUM(i.quantity * i.unit_cost), 0)::NUMERIC(18,2) AS total_value,
        COUNT(i.id) AS item_count,
        COUNT(DISTINCT i.material_id) AS material_count
    FROM inventory i
    JOIN material m ON m.id = i.material_id
    LEFT JOIN material_category mc ON mc.id = m.category_id
    WHERE (p_warehouse_id IS NULL OR i.warehouse_id = p_warehouse_id)
      AND (p_category_id IS NULL OR m.category_id = p_category_id)
      AND m.deleted_at IS NULL
      AND m.status = 'enabled';
END;
$function$;

-- 函数: fn_create_serial_codes_for_stock_in
DROP FUNCTION IF EXISTS fn_create_serial_codes_for_stock_in CASCADE;
CREATE OR REPLACE FUNCTION public.fn_create_serial_codes_for_stock_in(p_stock_in_id bigint, p_operator_id bigint)
 RETURNS TABLE(serial_code character varying, material_name character varying)
 LANGUAGE plpgsql
AS $function$
DECLARE
    v_item RECORD;
    v_i INT;
    v_serial_code VARCHAR;
    v_stock_in RECORD;
    v_operator_name VARCHAR;
    v_count INT;
BEGIN
    SELECT * INTO v_stock_in FROM stock_in WHERE id = p_stock_in_id;
    SELECT username INTO v_operator_name FROM sys_user WHERE id = p_operator_id;

    FOR v_item IN 
        SELECT sii.*, m.material_code, m.material_name, m.is_code
        FROM stock_in_item sii
        JOIN material m ON m.id = sii.material_id
        WHERE sii.stock_in_id = p_stock_in_id
    LOOP
        IF v_item.is_code THEN
            v_count := FLOOR(v_item.accepted_quantity)::INT;

            FOR v_i IN 1..v_count LOOP
                v_serial_code := fn_generate_serial_code(v_item.material_id);

                INSERT INTO sku_serial_code (
                    serial_code, material_id, material_code, material_name,
                    stock_in_id, stock_in_item_id, inventory_id,
                    warehouse_id, status
                ) VALUES (
                    v_serial_code, v_item.material_id, v_item.material_code, v_item.material_name,
                    p_stock_in_id, v_item.id, NULL,
                    v_stock_in.warehouse_id, 'in_stock'
                );

                INSERT INTO sku_serial_trace (
                    serial_code_id, serial_code, action,
                    ref_doc_type, ref_doc_no, ref_doc_id,
                    to_warehouse_id, operator_id, operator_name,
                    remark
                ) VALUES (
                    (SELECT id FROM sku_serial_code WHERE serial_code = v_serial_code),
                    v_serial_code, 'stock_in',
                    'stock_in', v_stock_in.stock_in_no, p_stock_in_id,
                    v_stock_in.warehouse_id, p_operator_id, v_operator_name,
                    '入库自动生成编码'
                );

                serial_code := v_serial_code;
                material_name := v_item.material_name;
                RETURN NEXT;
            END LOOP;
        END IF;
    END LOOP;
END;
$function$;

-- 函数: fn_dict_label
DROP FUNCTION IF EXISTS fn_dict_label CASCADE;
CREATE OR REPLACE FUNCTION public.fn_dict_label(p_dict_type character varying, p_dict_value character varying)
 RETURNS character varying
 LANGUAGE plpgsql
 STABLE
AS $function$
DECLARE
    v_label VARCHAR;
BEGIN
    IF p_dict_value IS NULL OR p_dict_value = '' THEN
        RETURN '';
    END IF;

    SELECT dict_label INTO v_label
    FROM sys_dict_data
    WHERE dict_type = p_dict_type
      AND dict_value = p_dict_value
    LIMIT 1;

    RETURN COALESCE(v_label, p_dict_value);
END;
$function$;

-- 函数: fn_generate_base_code
DROP FUNCTION IF EXISTS fn_generate_base_code CASCADE;
CREATE OR REPLACE FUNCTION public.fn_generate_base_code(p_prefix character varying)
 RETURNS character varying
 LANGUAGE plpgsql
AS $function$
DECLARE
    v_seq      INT;
    v_result   VARCHAR(10);
    v_lock_key BIGINT;
BEGIN
    v_lock_key := hashtext(p_prefix || 'BASE');
    PERFORM pg_advisory_xact_lock(v_lock_key);

    INSERT INTO sys_serial_number (prefix, date_str, current_seq, updated_at)
    VALUES (p_prefix, 'BASE', 1, NOW())
    ON CONFLICT (prefix, date_str) DO UPDATE
        SET current_seq = sys_serial_number.current_seq + 1,
            updated_at = NOW()
    RETURNING current_seq INTO v_seq;

    v_result := p_prefix || LPAD(v_seq::TEXT, 4, '0');
    RETURN v_result;
END;
$function$;

-- 函数: fn_generate_consumption_order_no
DROP FUNCTION IF EXISTS fn_generate_consumption_order_no CASCADE;
CREATE OR REPLACE FUNCTION public.fn_generate_consumption_order_no()
 RETURNS character varying
 LANGUAGE plpgsql
AS $function$
DECLARE
    v_date_str VARCHAR(8);
    v_seq_num INTEGER;
    v_order_no VARCHAR(20);
BEGIN
    v_date_str := TO_CHAR(CURRENT_DATE, 'YYYYMMDD');
    
    SELECT COALESCE(MAX(SUBSTRING(order_no FROM 11 FOR 3)::INTEGER), 0) + 1
    INTO v_seq_num
    FROM consumption_order
    WHERE order_no LIKE 'CO' || v_date_str || '%';
    
    v_order_no := 'CO' || v_date_str || LPAD(v_seq_num::TEXT, 3, '0');
    RETURN v_order_no;
END;
$function$;

-- 函数: fn_generate_material_code
DROP FUNCTION IF EXISTS fn_generate_material_code CASCADE;
CREATE OR REPLACE FUNCTION public.fn_generate_material_code()
 RETURNS character varying
 LANGUAGE plpgsql
AS $function$
DECLARE
    v_year     VARCHAR(4);
    v_seq      INT;
    v_result   VARCHAR(20);
    v_lock_key BIGINT;
BEGIN
    v_year := TO_CHAR(CURRENT_DATE, 'YYYY');
    v_lock_key := hashtext('MAT' || v_year);
    PERFORM pg_advisory_xact_lock(v_lock_key);

    INSERT INTO sys_serial_number (prefix, date_str, current_seq, updated_at)
    VALUES ('MAT', v_year, 1, NOW())
    ON CONFLICT (prefix, date_str) DO UPDATE
        SET current_seq = sys_serial_number.current_seq + 1,
            updated_at = NOW()
    RETURNING current_seq INTO v_seq;

    v_result := 'ZXAD' || v_year || '-' || LPAD(v_seq::TEXT, 5, '0');
    RETURN v_result;
END;
$function$;

-- 函数: fn_generate_material_code_new
DROP FUNCTION IF EXISTS fn_generate_material_code_new CASCADE;
CREATE OR REPLACE FUNCTION public.fn_generate_material_code_new(p_prefix character varying DEFAULT 'ZXAD'::character varying, p_year integer DEFAULT EXTRACT(year FROM CURRENT_DATE))
 RETURNS character varying
 LANGUAGE plpgsql
AS $function$
DECLARE
    v_year_str VARCHAR(4);
    v_seq INT;
    v_result VARCHAR;
    v_lock_key BIGINT;
BEGIN
    v_year_str := p_year::TEXT;
    
    -- 使用advisory lock保证并发安全
    v_lock_key := hashtext(p_prefix || v_year_str);
    PERFORM pg_advisory_xact_lock(v_lock_key);
    
    -- UPSERT序号
    INSERT INTO sys_material_code_seq (prefix, year, current_seq)
    VALUES (p_prefix, p_year, 1)
    ON CONFLICT (prefix, year) 
    DO UPDATE SET current_seq = sys_material_code_seq.current_seq + 1
    RETURNING current_seq INTO v_seq;
    
    -- 生成编码: 前缀 + 年度号 - 5位序号
    v_result := p_prefix || v_year_str || '-' || LPAD(v_seq::TEXT, 5, '0');
    RETURN v_result;
END;
$function$;

-- 函数: fn_generate_reversal_order_no
DROP FUNCTION IF EXISTS fn_generate_reversal_order_no CASCADE;
CREATE OR REPLACE FUNCTION public.fn_generate_reversal_order_no()
 RETURNS character varying
 LANGUAGE plpgsql
AS $function$
DECLARE
    v_date_str VARCHAR(8);
    v_seq_num INTEGER;
    v_order_no VARCHAR(20);
BEGIN
    v_date_str := TO_CHAR(CURRENT_DATE, 'YYYYMMDD');
    
    SELECT COALESCE(MAX(SUBSTRING(order_no FROM 11 FOR 3)::INTEGER), 0) + 1
    INTO v_seq_num
    FROM reversal_order
    WHERE order_no LIKE 'RO' || v_date_str || '%';
    
    v_order_no := 'RO' || v_date_str || LPAD(v_seq_num::TEXT, 3, '0');
    RETURN v_order_no;
END;
$function$;

-- 函数: fn_generate_serial_code
DROP FUNCTION IF EXISTS fn_generate_serial_code CASCADE;
CREATE OR REPLACE FUNCTION public.fn_generate_serial_code(p_material_id bigint)
 RETURNS character varying
 LANGUAGE plpgsql
AS $function$
DECLARE
    v_date_str VARCHAR(8);
    v_seq      INT;
    v_result   VARCHAR;
    v_lock_key BIGINT;
BEGIN
    v_date_str := TO_CHAR(CURRENT_DATE, 'YYYYMMDD');
    v_lock_key := hashtext('SERIAL' || v_date_str || p_material_id::TEXT);
    PERFORM pg_advisory_xact_lock(v_lock_key);

    INSERT INTO sys_serial_number (prefix, date_str, current_seq, updated_at)
    VALUES ('SN' || p_material_id::TEXT, v_date_str, 1, NOW())
    ON CONFLICT (prefix, date_str) DO UPDATE
        SET current_seq = sys_serial_number.current_seq + 1,
            updated_at = NOW()
    RETURNING current_seq INTO v_seq;

    v_result := v_date_str || '-' || LPAD(p_material_id::TEXT, 4, '0') || '-' || LPAD(v_seq::TEXT, 4, '0');
    RETURN v_result;
END;
$function$;

-- 函数: fn_generate_serial_no
DROP FUNCTION IF EXISTS fn_generate_serial_no CASCADE;
CREATE OR REPLACE FUNCTION public.fn_generate_serial_no(p_prefix character varying)
 RETURNS character varying
 LANGUAGE plpgsql
AS $function$
DECLARE
    v_date_str  VARCHAR(8);
    v_seq       INT;
    v_result    VARCHAR(20);
    v_lock_key  BIGINT;
BEGIN
    -- 当天日期
    v_date_str := TO_CHAR(CURRENT_DATE, 'YYYYMMDD');

    -- 使用 advisory lock 避免并发冲突
    -- lock key = prefix 的 hash
    v_lock_key := hashtext(p_prefix || v_date_str);
    PERFORM pg_advisory_xact_lock(v_lock_key);

    -- UPSERT: 存在则 +1，不存在则插入
    INSERT INTO sys_serial_number (prefix, date_str, current_seq, updated_at)
    VALUES (p_prefix, v_date_str, 1, NOW())
    ON CONFLICT (prefix, date_str) DO UPDATE
        SET current_seq = sys_serial_number.current_seq + 1,
            updated_at = NOW()
    RETURNING current_seq INTO v_seq;

    -- 拼接结果: 前缀 + 日期 + 3位流水号
    v_result := p_prefix || v_date_str || LPAD(v_seq::TEXT, 3, '0');

    RETURN v_result;
END;
$function$;

-- 函数: fn_generate_sku
DROP FUNCTION IF EXISTS fn_generate_sku CASCADE;
CREATE OR REPLACE FUNCTION public.fn_generate_sku(p_material_id bigint, p_attributes jsonb)
 RETURNS character varying
 LANGUAGE plpgsql
AS $function$
DECLARE
    v_material_code VARCHAR;
    v_attr_hash VARCHAR;
    v_sku VARCHAR;
BEGIN
    -- 获取物料编码
    SELECT material_code INTO v_material_code 
    FROM material WHERE id = p_material_id;
    
    -- 如果没有属性,直接返回物料编码
    IF p_attributes IS NULL OR jsonb_array_length(p_attributes) = 0 THEN
        RETURN v_material_code;
    END IF;
    
    -- 生成属性哈希(取前6位)
    v_attr_hash := LEFT(MD5(p_attributes::TEXT), 6);
    
    -- 组合SKU: 物料编码-属性哈希
    v_sku := v_material_code || '-' || UPPER(v_attr_hash);
    
    RETURN v_sku;
END;
$function$;

-- 函数: fn_generate_sku_code
DROP FUNCTION IF EXISTS fn_generate_sku_code CASCADE;
CREATE OR REPLACE FUNCTION public.fn_generate_sku_code(p_material_id bigint)
 RETURNS character varying
 LANGUAGE plpgsql
AS $function$
DECLARE
    v_material_code VARCHAR;
    v_seq INT;
    v_result VARCHAR;
    v_lock_key BIGINT;
BEGIN
    SELECT material_code INTO v_material_code
    FROM material
    WHERE id = p_material_id AND deleted_at IS NULL;

    IF v_material_code IS NULL THEN
        RAISE EXCEPTION '物料不存在 [id=%]', p_material_id;
    END IF;

    v_lock_key := hashtext('SKU' || p_material_id::TEXT);
    PERFORM pg_advisory_xact_lock(v_lock_key);

    INSERT INTO sys_sku_code_seq (material_id, current_seq)
    VALUES (p_material_id, 1)
    ON CONFLICT (material_id)
    DO UPDATE SET current_seq = sys_sku_code_seq.current_seq + 1,
                  updated_at = NOW()
    RETURNING current_seq INTO v_seq;

    v_result := v_material_code || '-' || LPAD(v_seq::TEXT, 3, '0');
    RETURN v_result;
END;
$function$;

-- 函数: fn_get_available_stock
DROP FUNCTION IF EXISTS fn_get_available_stock CASCADE;
CREATE OR REPLACE FUNCTION public.fn_get_available_stock(p_material_id bigint DEFAULT NULL::bigint, p_warehouse_id bigint DEFAULT NULL::bigint)
 RETURNS TABLE(material_id bigint, material_code character varying, material_name character varying, warehouse_id bigint, warehouse_code character varying, warehouse_name character varying, location_id bigint, location_code character varying, location_name character varying, quantity numeric, locked_quantity numeric, in_transit_quantity numeric, available_quantity numeric)
 LANGUAGE plpgsql
AS $function$
BEGIN
    RETURN QUERY
    SELECT 
        i.material_id,
        m.material_code,
        m.material_name,
        i.warehouse_id,
        w.warehouse_code,
        w.warehouse_name,
        i.location_id,
        l.location_code,
        l.location_name,
        i.quantity,
        i.locked_quantity,
        COALESCE(i.in_transit_quantity, 0) as in_transit_quantity,
        (i.quantity - i.locked_quantity) as available_quantity
    FROM inventory i
    JOIN material m ON m.id = i.material_id
    JOIN warehouse w ON w.id = i.warehouse_id
    LEFT JOIN location l ON l.id = i.location_id
    WHERE (p_material_id IS NULL OR i.material_id = p_material_id)
      AND (p_warehouse_id IS NULL OR i.warehouse_id = p_warehouse_id);
END;
$function$;

-- 函数: fn_get_category_tree
DROP FUNCTION IF EXISTS fn_get_category_tree CASCADE;
CREATE OR REPLACE FUNCTION public.fn_get_category_tree()
 RETURNS jsonb
 LANGUAGE sql
 STABLE
AS $function$
    WITH RECURSIVE tree AS (
        SELECT id, parent_id, category_code, category_name, sort_order, status
        FROM material_category
        WHERE parent_id = 0 AND deleted_at IS NULL
        UNION ALL
        SELECT c.id, c.parent_id, c.category_code, c.category_name, c.sort_order, c.status
        FROM material_category c
        INNER JOIN tree t ON c.parent_id = t.id
        WHERE c.deleted_at IS NULL
    ),
    -- 先找叶子节点（没有子分类的）
    leaf AS (
        SELECT t.id, t.parent_id, t.category_code, t.category_name, t.sort_order, t.status,
               '[]'::jsonb AS children
        FROM tree t
        WHERE NOT EXISTS (
            SELECT 1 FROM tree child WHERE child.parent_id = t.id
        )
    ),
    -- 再构建子节点 JSON
    nodes AS (
        SELECT t.id, t.parent_id, t.category_code, t.category_name, t.sort_order, t.status,
               COALESCE(
                   (SELECT jsonb_agg(
                       jsonb_build_object(
                           'id', sub.id,
                           'category_code', sub.category_code,
                           'category_name', sub.category_name,
                           'sort_order', sub.sort_order,
                           'status', sub.status,
                           'children', '[]'::jsonb
                       ) ORDER BY sub.sort_order
                   )
                   FROM tree sub WHERE sub.parent_id = t.id),
                   '[]'::jsonb
               ) AS children
        FROM tree t
        WHERE t.parent_id = 0
    )
    SELECT COALESCE(
        jsonb_agg(
            jsonb_build_object(
                'id', n.id,
                'category_code', n.category_code,
                'category_name', n.category_name,
                'sort_order', n.sort_order,
                'status', n.status,
                'children', n.children
            ) ORDER BY n.sort_order
        ),
        '[]'::jsonb
    )
    FROM nodes n;
$function$;

-- 函数: fn_get_user_permissions
DROP FUNCTION IF EXISTS fn_get_user_permissions CASCADE;
CREATE OR REPLACE FUNCTION public.fn_get_user_permissions(p_user_id bigint)
 RETURNS TABLE(perm_code character varying)
 LANGUAGE sql
 STABLE
AS $function$
    SELECT DISTINCT p.perm_code
    FROM sys_permission p
    INNER JOIN sys_role_permission rp ON rp.permission_id = p.id
    INNER JOIN sys_user_role ur ON ur.role_id = rp.role_id
    WHERE ur.user_id = p_user_id
      AND p.status = 'enabled';
$function$;

-- 函数: fn_trace_backward
DROP FUNCTION IF EXISTS fn_trace_backward CASCADE;
CREATE OR REPLACE FUNCTION public.fn_trace_backward(p_keyword character varying)
 RETURNS TABLE(trace_type character varying, doc_no character varying, doc_date date, material_code character varying, material_name character varying, quantity numeric, supplier_name character varying, cert_no character varying)
 LANGUAGE sql
 STABLE
AS $function$
	SELECT
		'出库'::varchar AS trace_type,
		so.stock_out_no AS doc_no,
		so.stock_out_date AS doc_date,
		m.material_code,
		m.material_name,
		soi.quantity,
		NULL::varchar AS supplier_name,
		NULL::varchar AS cert_no
	FROM stock_out_item soi
	JOIN stock_out so ON so.id = soi.stock_out_id AND so.deleted_at IS NULL
	JOIN material m ON m.id = soi.material_id
	LEFT JOIN sku_serial_code sc ON sc.inventory_id = soi.inventory_id
	WHERE sc.serial_code = p_keyword OR so.stock_out_no = p_keyword

	UNION ALL

	SELECT
		'入库'::varchar AS trace_type,
		si.stock_in_no AS doc_no,
		si.stock_in_date AS doc_date,
		m.material_code,
		m.material_name,
		sii.accepted_quantity AS quantity,
		s.supplier_name,
		NULL::varchar as cert_no
	FROM stock_in_item sii
	JOIN stock_in si ON si.id = sii.stock_in_id AND si.deleted_at IS NULL
	JOIN material m ON m.id = sii.material_id
	LEFT JOIN supplier s ON s.supplier_code = si.supplier_code
	LEFT JOIN sku_serial_code sc ON sc.stock_in_item_id = sii.id
	WHERE sc.serial_code = p_keyword OR si.stock_in_no = p_keyword;
$function$;

-- 函数: fn_trace_forward
DROP FUNCTION IF EXISTS fn_trace_forward CASCADE;
CREATE OR REPLACE FUNCTION public.fn_trace_forward(p_inspection_no character varying)
 RETURNS TABLE(trace_type character varying, doc_no character varying, doc_date date, material_code character varying, material_name character varying, inspection_no character varying, quantity numeric, warehouse character varying)
 LANGUAGE sql
 STABLE
AS $function$
	SELECT
		'入库'::varchar AS trace_type,
		si.stock_in_no AS doc_no,
		si.stock_in_date AS doc_date,
		m.material_code,
		m.material_name,
		i.material_inspection_no AS inspection_no,
		sii.accepted_quantity AS quantity,
		w.warehouse_name AS warehouse
	FROM stock_in_item sii
	JOIN stock_in si ON si.id = sii.stock_in_id AND si.deleted_at IS NULL
	JOIN material m ON m.id = sii.material_id
	LEFT JOIN warehouse w ON w.id = si.warehouse_id
    JOIN inventory i ON i.material_id = sii.material_id AND i.material_inspection_no = p_inspection_no
	WHERE i.material_inspection_no = p_inspection_no

	UNION ALL

	SELECT
		'出库'::varchar AS trace_type,
		so.stock_out_no AS doc_no,
		so.stock_out_date AS doc_date,
		m.material_code,
		m.material_name,
		i.material_inspection_no AS inspection_no,
		soi.quantity,
		w.warehouse_name AS warehouse
	FROM stock_out_item soi
	JOIN stock_out so ON so.id = soi.stock_out_id AND so.deleted_at IS NULL
	JOIN material m ON m.id = soi.material_id
    LEFT JOIN inventory i ON i.id = soi.inventory_id
	LEFT JOIN warehouse w ON w.id = i.warehouse_id
	WHERE i.material_inspection_no = p_inspection_no

	ORDER BY doc_date, trace_type;
$function$;

-- 函数: fn_trace_material_by_inspection
DROP FUNCTION IF EXISTS fn_trace_material_by_inspection CASCADE;
CREATE OR REPLACE FUNCTION public.fn_trace_material_by_inspection(p_material_code character varying DEFAULT NULL::character varying, p_inspection_no character varying DEFAULT NULL::character varying)
 RETURNS TABLE(serial_code character varying, material_code character varying, material_name character varying, inspection_no character varying, status character varying, status_label character varying, last_action character varying, last_action_label character varying, last_operator_name character varying, last_created_at timestamp with time zone, warehouse_name character varying)
 LANGUAGE plpgsql
AS $function$
BEGIN
    RETURN QUERY
    SELECT
        sc.serial_code,
        sc.material_code,
        sc.material_name,
        sc.serial_code as inspection_no,
        sc.status,
        CASE sc.status
            WHEN 'in_stock' THEN '在库'
            WHEN 'issued' THEN '已领用'
            WHEN 'returned' THEN '已退回'
            WHEN 'scrapped' THEN '已报废'
            ELSE sc.status
        END::VARCHAR AS status_label,
        t.action AS last_action,
        CASE t.action
            WHEN 'stock_in' THEN '入库'
            WHEN 'stock_out' THEN '出库领用'
            WHEN 'return' THEN '退料退回'
            WHEN 'transfer' THEN '调拨'
            WHEN 'scrap' THEN '报废'
            ELSE t.action
        END::VARCHAR AS last_action_label,
        t.operator_name AS last_operator_name,
        t.created_at AS last_created_at,
        w.warehouse_name
    FROM sku_serial_code sc
    LEFT JOIN LATERAL (
        SELECT action, operator_name, created_at
        FROM sku_serial_trace
        WHERE serial_code_id = sc.id
        ORDER BY created_at DESC
        LIMIT 1
    ) t ON TRUE
    LEFT JOIN warehouse w ON w.id = sc.warehouse_id
    WHERE (p_material_code IS NULL OR sc.material_code = p_material_code);
END;
$function$;

-- 函数: fn_trace_material_by_serial
DROP FUNCTION IF EXISTS fn_trace_material_by_serial CASCADE;
CREATE OR REPLACE FUNCTION public.fn_trace_material_by_serial(p_serial_code character varying)
 RETURNS TABLE(id bigint, serial_code character varying, material_code character varying, material_name character varying, action character varying, action_label character varying, ref_doc_type character varying, ref_doc_no character varying, ref_doc_id bigint, from_warehouse_name character varying, to_warehouse_name character varying, operator_name character varying, remark text, created_at timestamp with time zone)
 LANGUAGE plpgsql
AS $function$
BEGIN
    RETURN QUERY
    SELECT
        t.id,
        t.serial_code,
        sc.material_code,
        sc.material_name,
        t.action,
        CASE t.action
            WHEN 'stock_in' THEN '入库'
            WHEN 'stock_out' THEN '出库领用'
            WHEN 'return' THEN '退料退回'
            WHEN 'transfer' THEN '调拨'
            WHEN 'scrap' THEN '报废'
            ELSE t.action
        END::VARCHAR AS action_label,
        t.ref_doc_type,
        t.ref_doc_no,
        t.ref_doc_id,
        fw.warehouse_name AS from_warehouse_name,
        tw.warehouse_name AS to_warehouse_name,
        t.operator_name,
        t.remark,
        t.created_at
    FROM sku_serial_trace t
    INNER JOIN sku_serial_code sc ON sc.id = t.serial_code_id
    LEFT JOIN warehouse fw ON fw.id = t.from_warehouse_id
    LEFT JOIN warehouse tw ON tw.id = t.to_warehouse_id
    WHERE t.serial_code = p_serial_code
    ORDER BY t.created_at ASC;
END;
$function$;

-- 函数: trf_calc_item_amount
DROP FUNCTION IF EXISTS trf_calc_item_amount CASCADE;
CREATE OR REPLACE FUNCTION public.trf_calc_item_amount()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$
BEGIN
    -- 自动计算金额 = 数量 * 单价
    NEW.amount := COALESCE(NEW.quantity, 0) * COALESCE(NEW.unit_price, 0);
    RETURN NEW;
END;
$function$;

-- 函数: trf_sync_purchase_order_amount
DROP FUNCTION IF EXISTS trf_sync_purchase_order_amount CASCADE;
CREATE OR REPLACE FUNCTION public.trf_sync_purchase_order_amount()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$
DECLARE
    v_order_id BIGINT;
BEGIN
    v_order_id := COALESCE(NEW.order_id, OLD.order_id);
    
    UPDATE purchase_order
    SET total_amount = (
        SELECT COALESCE(SUM(amount), 0)
        FROM purchase_order_item
        WHERE order_id = v_order_id
    ),
    updated_at = NOW()
    WHERE id = v_order_id;
    
    RETURN COALESCE(NEW, OLD);
END;
$function$;

-- =====================================================
-- 存储过程定义
-- =====================================================

-- 存储过程: sp_approve_purchase_request
DROP PROCEDURE IF EXISTS sp_approve_purchase_request CASCADE;
CREATE OR REPLACE PROCEDURE public.sp_approve_purchase_request(IN p_request_id bigint, IN p_approver_id bigint, IN p_action character varying, IN p_remark text DEFAULT NULL::text)
 LANGUAGE plpgsql
AS $procedure$
DECLARE
    v_current_status VARCHAR(20);
BEGIN
    -- 查询当前状态并锁定
    SELECT approval_status INTO v_current_status
    FROM purchase_request
    WHERE id = p_request_id AND deleted_at IS NULL
    FOR UPDATE;

    IF v_current_status IS NULL THEN
        RAISE EXCEPTION '采购申请不存在或已删除 [ID=%]', p_request_id;
    END IF;

    IF v_current_status != 'pending' THEN
        RAISE EXCEPTION '采购申请当前状态[%]不允许审批', v_current_status;
    END IF;

    IF p_action = 'approve' THEN
        UPDATE purchase_request
        SET approval_status = 'approved',
            approved_by = p_approver_id,
            approved_at = NOW(),
            approval_remark = p_remark,
            updated_by = p_approver_id,
            updated_at = NOW()
        WHERE id = p_request_id;
    ELSIF p_action = 'reject' THEN
        UPDATE purchase_request
        SET approval_status = 'rejected',
            approved_by = p_approver_id,
            approved_at = NOW(),
            approval_remark = p_remark,
            updated_by = p_approver_id,
            updated_at = NOW()
        WHERE id = p_request_id;
    ELSE
        RAISE EXCEPTION '无效的审批操作: %', p_action;
    END IF;
END;
$procedure$;

-- 存储过程: sp_cancel_sales_order
DROP PROCEDURE IF EXISTS sp_cancel_sales_order CASCADE;
CREATE OR REPLACE PROCEDURE public.sp_cancel_sales_order(IN p_order_id bigint, IN p_operator_id bigint)
 LANGUAGE plpgsql
AS $procedure$
DECLARE
    v_item RECORD;
BEGIN
    IF NOT EXISTS(
        SELECT 1 FROM sales_order
        WHERE id = p_order_id AND order_status IN ('confirmed', 'preparing') AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION '销售订单不存在或状态不允许取消';
    END IF;

    -- 释放锁定库存
    FOR v_item IN
        SELECT soi.material_id, soi.quantity - soi.shipped_quantity AS remaining
        FROM sales_order_item soi
        WHERE soi.order_id = p_order_id AND soi.quantity > soi.shipped_quantity
    LOOP
        UPDATE inventory
        SET locked_quantity = GREATEST(locked_quantity - v_item.remaining, 0),
            updated_at = NOW()
        WHERE material_id = v_item.material_id
          AND locked_quantity > 0;
    END LOOP;

    UPDATE sales_order
    SET order_status = 'cancelled',
        updated_by = p_operator_id,
        updated_at = NOW()
    WHERE id = p_order_id;
END;
$procedure$;

-- 存储过程: sp_check_stock_alerts
DROP PROCEDURE IF EXISTS sp_check_stock_alerts CASCADE;
CREATE OR REPLACE PROCEDURE public.sp_check_stock_alerts()
 LANGUAGE plpgsql
AS $procedure$
DECLARE
    v_alert        RECORD;
    v_admin_id     BIGINT;
    v_title        VARCHAR(200);
    v_content      TEXT;
BEGIN
    -- 获取一个系统管理员或仓库管理员ID (此处简单取 ID=1)
    v_admin_id := 1; 

    FOR v_alert IN SELECT * FROM v_inventory_alert LOOP
        v_title := v_alert.alert_type || '预警: ' || v_alert.material_name;
        v_content := '物料[' || v_alert.material_code || ']当前库存[' || v_alert.total_quantity || 
                     '], ' || v_alert.alert_type || '余量[' || v_alert.alert_quantity || '].';

        -- 检查是否已存在未读的相同物料、相同类型的通知，避免重复发送
        IF NOT EXISTS (
            SELECT 1 FROM sys_notification 
            WHERE user_id = v_admin_id 
              AND ref_type = 'material' 
              AND ref_id = v_alert.material_id 
              AND title = v_title 
              AND is_read = FALSE
        ) THEN
            INSERT INTO sys_notification (
                user_id, title, content, notify_type, ref_type, ref_id, is_read, created_at
            )
            VALUES (
                v_admin_id, v_title, v_content, 'stock_alert', 'material', v_alert.material_id, FALSE, NOW()
            );
        END IF;
    END LOOP;
END;
$procedure$;

-- 存储过程: sp_confirm_consumption_order
DROP PROCEDURE IF EXISTS sp_confirm_consumption_order CASCADE;
CREATE OR REPLACE PROCEDURE public.sp_confirm_consumption_order(IN p_order_id bigint, IN p_operator_id bigint)
 LANGUAGE plpgsql
AS $procedure$
DECLARE
	v_order        RECORD;
	v_item         RECORD;
	v_stock_out_id BIGINT;
	v_stock_out_no VARCHAR(20);
BEGIN
	SELECT *
	INTO v_order
	FROM consumption_order
	WHERE id = p_order_id
	  AND deleted_at IS NULL
	FOR UPDATE;

	IF v_order IS NULL THEN
		RAISE EXCEPTION '领料订单不存在 [id=%]', p_order_id;
	END IF;

	IF v_order.status NOT IN ('pending', 'confirmed', 'completed') THEN
		RAISE EXCEPTION '单据状态[%]不允许确认', v_order.status;
	END IF;

	IF COALESCE(v_order.stock_out_id, 0) <> 0 THEN
		-- 已有关联出库单时只做状态兜底，不重复生成
		UPDATE consumption_order
		SET status = 'confirmed',
			updated_by = p_operator_id,
			updated_at = NOW()
		WHERE id = p_order_id;
		RETURN;
	END IF;

	v_stock_out_no := fn_generate_serial_no('SO');

	INSERT INTO stock_out (
		stock_out_no, out_type, ref_doc_type, ref_doc_id,
		stock_out_date, receiver, status, remark, created_by, created_at, updated_by, updated_at
	) VALUES (
		v_stock_out_no, 'consumption', 'consumption_order', p_order_id,
		COALESCE(v_order.order_date, CURRENT_DATE), COALESCE(v_order.designer_name, ''), 'pending',
		'领料订单自动生成', p_operator_id, NOW(), p_operator_id, NOW()
	)
	RETURNING id INTO v_stock_out_id;

	FOR v_item IN
		SELECT coi.*
		FROM consumption_order_item coi
		WHERE coi.order_id = p_order_id
		ORDER BY coi.id
	LOOP
		INSERT INTO stock_out_item (
			stock_out_id, material_id, inventory_id, quantity, unit, created_at
		) VALUES (
			v_stock_out_id, v_item.material_id, v_item.inventory_id,
			v_item.quantity, COALESCE(NULLIF(TRIM(v_item.unit), ''), '件'), NOW()
		);
	END LOOP;

	UPDATE consumption_order
	SET status = 'confirmed',
		stock_out_id = v_stock_out_id,
		updated_by = p_operator_id,
		updated_at = NOW()
	WHERE id = p_order_id;
END;
$procedure$;

-- 存储过程: sp_confirm_inventory_check
DROP PROCEDURE IF EXISTS sp_confirm_inventory_check CASCADE;
CREATE OR REPLACE PROCEDURE public.sp_confirm_inventory_check(IN p_check_id bigint, IN p_operator_id bigint)
 LANGUAGE plpgsql
AS $procedure$
DECLARE
    v_item RECORD;
    v_ch RECORD;
    v_balance NUMERIC;
BEGIN
    SELECT * INTO v_ch FROM inventory_check WHERE id = p_check_id;
    IF v_ch.status <> 'pending' THEN
        RAISE EXCEPTION '只有待确认状态的盘点单可以确认';
    END IF;

    FOR v_item IN 
        SELECT ici.*, m.unit
        FROM inventory_check_item ici
        INNER JOIN material m ON m.id = ici.material_id
        WHERE ici.check_id = p_check_id AND ici.diff_quantity <> 0
    LOOP
        IF v_item.diff_quantity > 0 THEN
            -- 盘盈：增加库存
            INSERT INTO inventory (
                material_id, warehouse_id, quantity, unit, updated_at
            )
            VALUES (
                v_item.material_id, v_ch.warehouse_id,
                v_item.diff_quantity, v_item.unit, NOW()
            )
            ON CONFLICT (material_id, warehouse_id, COALESCE(material_inspection_no, ''))
            DO UPDATE SET
                quantity = inventory.quantity + EXCLUDED.quantity,
                updated_at = NOW();

            -- 流水
            SELECT quantity INTO v_balance FROM inventory 
            WHERE material_id = v_item.material_id AND warehouse_id = v_ch.warehouse_id;

            INSERT INTO inventory_transaction (
                material_id, warehouse_id, trans_type, quantity,
                balance, ref_doc_type, ref_doc_no, ref_doc_id, operator_id,
                remark, created_at
            )
            VALUES (
                v_item.material_id, v_ch.warehouse_id,
                'in', v_item.diff_quantity, v_balance,
                'inventory_check', v_ch.check_no, p_check_id,
                p_operator_id, '盘点盘盈', NOW()
            );
        ELSE
            -- 盘亏：减少库存
            UPDATE inventory 
            SET quantity = quantity + v_item.diff_quantity, -- diff_quantity is negative
                updated_at = NOW()
            WHERE material_id = v_item.material_id
              AND warehouse_id = v_ch.warehouse_id;
            
            -- 流水
            SELECT quantity INTO v_balance FROM inventory 
            WHERE material_id = v_item.material_id AND warehouse_id = v_ch.warehouse_id;

            INSERT INTO inventory_transaction (
                material_id, warehouse_id, trans_type, quantity,
                balance, ref_doc_type, ref_doc_no, ref_doc_id, operator_id,
                remark, created_at
            )
            VALUES (
                v_item.material_id, v_ch.warehouse_id,
                'out', v_item.diff_quantity, v_balance,
                'inventory_check', v_ch.check_no, p_check_id,
                p_operator_id, '盘点盘亏', NOW()
            );
        END IF;
    END LOOP;

    UPDATE inventory_check 
    SET status = 'finished', updated_at = NOW(), updated_by = p_operator_id 
    WHERE id = p_check_id;
END;
$procedure$;

-- 存储过程: sp_confirm_purchase_order
DROP PROCEDURE IF EXISTS sp_confirm_purchase_order CASCADE;
CREATE OR REPLACE PROCEDURE public.sp_confirm_purchase_order(IN p_order_id bigint, IN p_operator_id bigint)
 LANGUAGE plpgsql
AS $procedure$
DECLARE
  v_order      RECORD;
  v_stock_in_id BIGINT;
  v_stock_in_no VARCHAR;
BEGIN
  SELECT * INTO v_order
  FROM purchase_order
  WHERE id = p_order_id AND deleted_at IS NULL
  FOR UPDATE;

  IF v_order IS NULL THEN
    RAISE EXCEPTION '采购订单不存在 [id=%]', p_order_id;
  END IF;

  IF v_order.order_status != 'draft' THEN
    RAISE EXCEPTION '只有草稿状态的订单可以确认 [当前状态=%]', v_order.order_status;
  END IF;

  UPDATE purchase_order
  SET order_status = 'ordered',
      updated_by = p_operator_id,
      updated_at = NOW()
  WHERE id = p_order_id;

  v_stock_in_no := 'SI' || TO_CHAR(NOW(), 'YYYYMMDD') || LPAD(nextval('stock_in_no_seq')::TEXT, 3, '0');

  INSERT INTO stock_in (
    stock_in_no, purchase_order_id, supplier_id, warehouse_id, warehouse_code, stock_in_date,
    stock_in_status, stock_in_type, created_by, created_at, updated_by, updated_at
  ) VALUES (
    v_stock_in_no, p_order_id, v_order.supplier_id, NULL, NULL, CURRENT_DATE,
    'preparing', 'purchase', p_operator_id, NOW(), p_operator_id, NOW()
  )
  RETURNING id INTO v_stock_in_id;

  INSERT INTO stock_in_item (
    stock_in_id, material_id, sku_id, custom_attributes,
    arrived_quantity, accepted_quantity, unit, unit_cost
  )
  SELECT
    v_stock_in_id,
    poi.material_id,
    poi.sku_id,
    COALESCE(ms.custom_attributes, '[]'::jsonb),
    poi.quantity,
    poi.quantity,
    COALESCE(m.unit, ''),
    COALESCE(poi.unit_price, 0)
  FROM purchase_order_item poi
  JOIN material m ON m.id = poi.material_id
  LEFT JOIN material_sku ms ON ms.id = poi.sku_id
  WHERE poi.order_id = p_order_id;

  INSERT INTO sys_audit_log (user_id, username, action, module, target_type, target_id, detail)
  SELECT p_operator_id, u.username, 'CREATE', 'stock_in', 'stock_in', v_stock_in_id,
         jsonb_build_object('stock_in_no', v_stock_in_no, 'from_order_no', v_order.order_no)
  FROM sys_user u WHERE u.id = p_operator_id;

  INSERT INTO sys_audit_log (user_id, username, action, module, target_type, target_id, detail)
  SELECT p_operator_id, u.username, 'CONFIRM', 'purchase_order', 'purchase_order', p_order_id,
         jsonb_build_object('order_no', v_order.order_no)
  FROM sys_user u WHERE u.id = p_operator_id;
END;
$procedure$;

-- 存储过程: sp_confirm_return_order
DROP PROCEDURE IF EXISTS sp_confirm_return_order CASCADE;
CREATE OR REPLACE PROCEDURE public.sp_confirm_return_order(IN p_return_id bigint, IN p_operator_id bigint)
 LANGUAGE plpgsql
AS $procedure$
DECLARE
    v_ro         RECORD;
    v_item       RECORD;
    v_balance    NUMERIC(18, 3);
    v_trans_type VARCHAR(20);
    v_inv_id     BIGINT;
BEGIN
    SELECT id, return_no, return_type, warehouse_id, return_status
    INTO v_ro
    FROM return_order
    WHERE id = p_return_id
      AND deleted_at IS NULL
    FOR UPDATE;

    IF v_ro IS NULL THEN
        RAISE EXCEPTION '退货单不存在';
    END IF;

    IF v_ro.return_status <> 'draft' THEN
        RAISE EXCEPTION '退货单当前状态[%]不允许确认', v_ro.return_status;
    END IF;

    FOR v_item IN
        SELECT roi.*, m.material_code
        FROM return_order_item roi
        INNER JOIN material m ON m.id = roi.material_id
        WHERE roi.return_id = p_return_id
    LOOP
        IF v_ro.return_type = 'sales_return' THEN
            v_trans_type := 'in';

            INSERT INTO inventory (material_id, warehouse_id, quantity, unit, updated_at)
            VALUES (v_item.material_id, v_ro.warehouse_id, v_item.quantity, v_item.unit, NOW())
            ON CONFLICT (material_id, warehouse_id, COALESCE(material_inspection_no, ''))
                DO UPDATE SET quantity   = inventory.quantity + EXCLUDED.quantity,
                              updated_at = NOW()
            RETURNING id INTO v_inv_id;

        ELSIF v_ro.return_type = 'purchase_return' THEN
            v_trans_type := 'out';

            -- 采购退货通常需要指定 inventory_id，如果没有指定，按 key 扣减
            IF v_item.inventory_id IS NOT NULL AND v_item.inventory_id > 0 THEN
                UPDATE inventory
                SET quantity   = quantity - v_item.quantity,
                    updated_at = NOW()
                WHERE id = v_item.inventory_id
                RETURNING id INTO v_inv_id;
            ELSE
                UPDATE inventory
                SET quantity   = quantity - v_item.quantity,
                    updated_at = NOW()
                WHERE material_id = v_item.material_id
                  AND warehouse_id = v_ro.warehouse_id
                  AND COALESCE(material_inspection_no, '') = ''
                RETURNING id INTO v_inv_id;
            END IF;

            IF NOT FOUND THEN
                RAISE EXCEPTION '库内物料[%]库存不足，无法进行采购退货', v_item.material_code;
            END IF;
        ELSE
            RAISE EXCEPTION '未知的退货类型: %', v_ro.return_type;
        END IF;

        SELECT quantity INTO v_balance FROM inventory WHERE id = v_inv_id;

        IF v_balance < 0 THEN
            RAISE EXCEPTION '退货导致库存为负 [物料=%]', v_item.material_code;
        END IF;

        INSERT INTO inventory_transaction (
            material_id, warehouse_id, trans_type, quantity, balance,
            ref_doc_type, ref_doc_no, ref_doc_id, operator_id, created_at
        )
        VALUES (
            v_item.material_id, v_ro.warehouse_id, v_trans_type,
            CASE WHEN v_trans_type = 'in' THEN v_item.quantity ELSE -v_item.quantity END,
            v_balance, 'return_order', v_ro.return_no, p_return_id, p_operator_id, NOW()
        );
    END LOOP;

    UPDATE return_order
    SET return_status = 'confirmed',
        updated_at    = NOW(),
        updated_by    = p_operator_id
    WHERE id = p_return_id;
END;
$procedure$;

-- 存储过程: sp_confirm_reversal_order
DROP PROCEDURE IF EXISTS sp_confirm_reversal_order CASCADE;
CREATE OR REPLACE PROCEDURE public.sp_confirm_reversal_order(IN p_order_id bigint, IN p_operator_id bigint)
 LANGUAGE plpgsql
AS $procedure$
DECLARE
	v_order       RECORD;
	v_item        RECORD;
	v_stock_in_id BIGINT;
	v_stock_in_no VARCHAR(20);
BEGIN
	SELECT *
	INTO v_order
	FROM reversal_order
	WHERE id = p_order_id
	  AND deleted_at IS NULL
	FOR UPDATE;

	IF v_order IS NULL THEN
		RAISE EXCEPTION '退料订单不存在 [id=%]', p_order_id;
	END IF;

	IF v_order.status NOT IN ('pending', 'confirmed', 'completed') THEN
		RAISE EXCEPTION '单据状态[%]不允许确认', v_order.status;
	END IF;

	IF COALESCE(v_order.stock_in_id, 0) <> 0 THEN
		UPDATE reversal_order
		SET status = 'confirmed',
			updated_by = p_operator_id,
			updated_at = NOW()
		WHERE id = p_order_id;
		RETURN;
	END IF;

	v_stock_in_no := fn_generate_serial_no('SI');

	INSERT INTO stock_in (
		stock_in_no, warehouse_id, warehouse_code, stock_in_date, stock_in_status, stock_in_type,
		remark, created_by, created_at, updated_by, updated_at
	) VALUES (
		v_stock_in_no, v_order.warehouse_id, v_order.warehouse_code, COALESCE(v_order.order_date, CURRENT_DATE),
		'preparing', 'reversal', '退料订单自动生成', p_operator_id, NOW(), p_operator_id, NOW()
	)
	RETURNING id INTO v_stock_in_id;

	FOR v_item IN
		SELECT roi.material_id,
		       roi.quantity,
		       roi.unit,
		       roi.inventory_id,
		       inv.sku_id,
		       COALESCE(inv.unit_cost, 0) AS unit_cost,
		       COALESCE(ms.custom_attributes, '[]'::jsonb) AS custom_attributes,
		       COALESCE(NULLIF(TRIM(inv.unit), ''), NULLIF(TRIM(roi.unit), ''), NULLIF(TRIM(m.unit), ''), '件') AS final_unit
		FROM reversal_order_item roi
		LEFT JOIN inventory inv ON inv.id = roi.inventory_id
		LEFT JOIN material_sku ms ON ms.id = inv.sku_id
		LEFT JOIN material m ON m.id = roi.material_id
		WHERE roi.order_id = p_order_id
		ORDER BY roi.id
	LOOP
		INSERT INTO stock_in_item (
			stock_in_id, material_id, sku_id, custom_attributes,
			arrived_quantity, accepted_quantity, unit, unit_cost, created_at
		) VALUES (
			v_stock_in_id, v_item.material_id, v_item.sku_id, v_item.custom_attributes,
			v_item.quantity, v_item.quantity, v_item.final_unit, v_item.unit_cost, NOW()
		);
	END LOOP;

	UPDATE reversal_order
	SET status = 'confirmed',
		stock_in_id = v_stock_in_id,
		updated_by = p_operator_id,
		updated_at = NOW()
	WHERE id = p_order_id;
END;
$procedure$;

-- 存储过程: sp_confirm_sales_order
DROP PROCEDURE IF EXISTS sp_confirm_sales_order CASCADE;
CREATE OR REPLACE PROCEDURE public.sp_confirm_sales_order(IN p_order_id bigint, IN p_operator_id bigint)
 LANGUAGE plpgsql
AS $procedure$
DECLARE
    v_item       RECORD;
    v_available  NUMERIC(18,3);
BEGIN
    -- 校验状态
    IF NOT EXISTS(
        SELECT 1 FROM sales_order
        WHERE id = p_order_id AND order_status = 'draft' AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION '销售订单不存在或状态不允许确认';
    END IF;

    -- 遍历明细，锁定库存
    FOR v_item IN
        SELECT soi.*, m.material_code, m.material_name
        FROM sales_order_item soi
        INNER JOIN material m ON m.id = soi.material_id
        WHERE soi.order_id = p_order_id
    LOOP
        -- 检查可用库存
        v_available := fn_get_available_stock(v_item.material_id);
        IF v_available < v_item.quantity THEN
            RAISE EXCEPTION '物料[%-%]可用库存不足，需要:%, 可用:%',
                v_item.material_code, v_item.material_name,
                v_item.quantity, v_available;
        END IF;

        -- 增加锁定数量（按 FIFO 锁定最早入库的批次）
        UPDATE inventory
        SET locked_quantity = locked_quantity + v_item.quantity,
            updated_at = NOW()
        WHERE id = (
            SELECT id FROM inventory
            WHERE material_id = v_item.material_id
              AND quantity - locked_quantity > 0
            ORDER BY stock_in_date ASC NULLS LAST
            LIMIT 1
        );
    END LOOP;

    -- 更新订单状态
    UPDATE sales_order
    SET order_status = 'confirmed',
        updated_by = p_operator_id,
        updated_at = NOW()
    WHERE id = p_order_id;
END;
$procedure$;

-- 存储过程: sp_confirm_stock_in
DROP PROCEDURE IF EXISTS sp_confirm_stock_in CASCADE;
CREATE OR REPLACE PROCEDURE public.sp_confirm_stock_in(IN p_stock_in_id bigint, IN p_operator_id bigint)
 LANGUAGE plpgsql
AS $procedure$
DECLARE
    v_stock_in RECORD;
    v_item RECORD;
    v_serial_code VARCHAR;
    v_inv_id BIGINT;
    v_operator_name VARCHAR;
    v_i INT;
BEGIN
    SELECT si.* INTO v_stock_in
    FROM stock_in si
    WHERE si.id = p_stock_in_id
      AND si.deleted_at IS NULL
    FOR UPDATE;

    IF v_stock_in IS NULL THEN
        RAISE EXCEPTION '入库单不存在 [id=%]', p_stock_in_id;
    END IF;

    IF v_stock_in.stock_in_status = 'passed' THEN
        RAISE EXCEPTION '入库单已全部入库，无需重复确认';
    END IF;

    SELECT u.real_name INTO v_operator_name
    FROM sys_user u
    WHERE u.id = p_operator_id;

    FOR v_item IN
        SELECT sii.*, m.is_code, m.material_code, m.material_name
        FROM stock_in_item sii
        JOIN material m ON m.id = sii.material_id
        WHERE sii.stock_in_id = p_stock_in_id
    LOOP
        IF COALESCE(v_item.accepted_quantity, 0) <= 0 THEN
            CONTINUE;
        END IF;

        INSERT INTO inventory (
            material_id, warehouse_id, quantity, locked_quantity, unit, stock_in_date, unit_cost, material_inspection_no, sku_id
        ) VALUES (
            v_item.material_id, v_stock_in.warehouse_id, v_item.accepted_quantity, 0, v_item.unit,
            v_stock_in.stock_in_date, COALESCE(v_item.unit_cost, 0), NULL,
            (SELECT poi.sku_id FROM purchase_order_item poi WHERE poi.order_id = v_stock_in.purchase_order_id AND poi.material_id = v_item.material_id LIMIT 1)
        )
        ON CONFLICT (material_id, warehouse_id, COALESCE(material_inspection_no, ''))
        DO UPDATE SET
            quantity = inventory.quantity + v_item.accepted_quantity,
            unit_cost = CASE WHEN v_item.unit_cost > 0 THEN v_item.unit_cost ELSE inventory.unit_cost END,
            updated_at = NOW()
        RETURNING id INTO v_inv_id;

        INSERT INTO inventory_transaction (
            material_id, warehouse_id, trans_type, quantity, balance, ref_doc_type, ref_doc_no, ref_doc_id, operator_id, created_at
        ) VALUES (
            v_item.material_id, v_stock_in.warehouse_id, 'in', v_item.accepted_quantity,
            (SELECT quantity FROM inventory WHERE id = v_inv_id), 'stock_in', v_stock_in.stock_in_no, p_stock_in_id, p_operator_id, NOW()
        );

        IF v_item.is_code THEN
            FOR v_i IN 1..floor(v_item.accepted_quantity) LOOP
                v_serial_code := fn_generate_serial_no('MAT');

                INSERT INTO sku_serial_code (
                    serial_code,
                    material_id, material_code, material_name,
                    stock_in_id, stock_in_item_id,
                    inventory_id, warehouse_id,
                    status, created_at, updated_at
                ) VALUES (
                    v_serial_code,
                    v_item.material_id, v_item.material_code, v_item.material_name,
                    p_stock_in_id, v_item.id,
                    v_inv_id, v_stock_in.warehouse_id,
                    'in_stock', NOW(), NOW()
                );

                INSERT INTO sku_serial_trace (
                    serial_code_id, serial_code, action, ref_doc_type, ref_doc_no, ref_doc_id,
                    to_warehouse_id, operator_id, operator_name, remark, created_at
                ) VALUES (
                    (SELECT id FROM sku_serial_code WHERE serial_code = v_serial_code),
                    v_serial_code,
                    'stock_in', 'stock_in', v_stock_in.stock_in_no, p_stock_in_id,
                    v_stock_in.warehouse_id, p_operator_id, COALESCE(v_operator_name, ''), '入库确认生成编码', NOW()
                );
            END LOOP;
        END IF;
    END LOOP;

    UPDATE stock_in
    SET stock_in_status = 'passed',
        updated_by = p_operator_id,
        updated_at = NOW()
    WHERE id = p_stock_in_id;
END;
$procedure$;

-- 存储过程: sp_confirm_stock_out
DROP PROCEDURE IF EXISTS sp_confirm_stock_out CASCADE;
CREATE OR REPLACE PROCEDURE public.sp_confirm_stock_out(IN p_stock_out_id bigint, IN p_operator_id bigint)
 LANGUAGE plpgsql
AS $procedure$
DECLARE
	v_so           RECORD;
	v_item         RECORD;
	v_inv          RECORD;
	v_balance      NUMERIC(18,3);
	v_operator_name VARCHAR;
	v_ref_wh_id    BIGINT;
	v_min_wh       BIGINT;
	v_max_wh       BIGINT;
    v_need         INTEGER;
    v_selected_cnt INTEGER;
BEGIN
	SELECT id, stock_out_no, stock_out_date, out_type, ref_doc_type, ref_doc_id, status
	INTO v_so
	FROM stock_out
	WHERE id = p_stock_out_id AND deleted_at IS NULL
	FOR UPDATE;

	IF v_so IS NULL THEN
		RAISE EXCEPTION '出库单不存在或已删除 [ID=%]', p_stock_out_id;
	END IF;

	IF COALESCE(v_so.status, 'draft') = 'confirmed' THEN
		RAISE EXCEPTION '出库单已完成，无需重复确认';
	END IF;

	SELECT u.real_name INTO v_operator_name FROM sys_user u WHERE u.id = p_operator_id;

	SELECT MIN(i.warehouse_id), MAX(i.warehouse_id)
	INTO v_min_wh, v_max_wh
	FROM stock_out_item soi
	INNER JOIN inventory i ON i.id = soi.inventory_id
	WHERE soi.stock_out_id = p_stock_out_id
	  AND COALESCE(soi.inventory_id, 0) <> 0;

	IF v_max_wh IS NOT NULL AND v_min_wh IS DISTINCT FROM v_max_wh THEN
		RAISE EXCEPTION '出库明细涉及多个仓库，请拆单后再确认';
	END IF;
	v_ref_wh_id := v_min_wh;

	FOR v_item IN
		SELECT soi.*, m.material_code, m.material_name, m.is_code
		FROM stock_out_item soi
		INNER JOIN material m ON m.id = soi.material_id
		WHERE soi.stock_out_id = p_stock_out_id
	LOOP
		IF COALESCE(v_item.inventory_id, 0) <> 0 THEN
			SELECT i.*
			INTO v_inv
			FROM inventory i
			WHERE i.id = v_item.inventory_id
			FOR UPDATE;

			IF v_inv IS NULL THEN
				RAISE EXCEPTION '库存不存在 [inventory_id=%]', v_item.inventory_id;
			END IF;

			IF (v_inv.quantity - v_inv.locked_quantity) < v_item.quantity THEN
				RAISE EXCEPTION '物料[%-%]库存不足，需要:%, 可用:%',
					v_item.material_code, v_item.material_name,
					v_item.quantity, (v_inv.quantity - v_inv.locked_quantity);
			END IF;

			UPDATE inventory
			SET quantity = quantity - v_item.quantity,
				in_transit_quantity = GREATEST(COALESCE(in_transit_quantity, 0) - v_item.quantity, 0),
				updated_at = NOW()
			WHERE id = v_item.inventory_id;

			SELECT quantity INTO v_balance FROM inventory WHERE id = v_item.inventory_id;

			INSERT INTO inventory_transaction (
				material_id, warehouse_id, trans_type, quantity,
				balance, ref_doc_type, ref_doc_no, ref_doc_id,
				operator_id, remark, created_at
			)
			VALUES (
				v_item.material_id, v_inv.warehouse_id, 'out', -v_item.quantity, v_balance,
				'stock_out', v_so.stock_out_no, p_stock_out_id,
				p_operator_id, '出库确认', NOW()
			);

			IF v_item.is_code THEN
				v_need := FLOOR(v_item.quantity)::INT;

				SELECT COUNT(*) INTO v_selected_cnt
				FROM stock_out_item_serial_selection
				WHERE stock_out_item_id = v_item.id;

				IF v_selected_cnt <> v_need THEN
					RAISE EXCEPTION '物料[%]编码未备齐，需:% 已备:%', v_item.material_name, v_need, v_selected_cnt;
				END IF;

				WITH picked AS (
					UPDATE sku_serial_code sc
					SET status = 'issued',
						updated_at = NOW()
					FROM stock_out_item_serial_selection s
					WHERE sc.id = s.serial_code_id
					  AND s.stock_out_item_id = v_item.id
					  AND sc.status = 'in_stock'
					RETURNING sc.id, sc.serial_code
				)
                INSERT INTO sku_serial_trace (
                    serial_code_id, serial_code, action,
                    ref_doc_type, ref_doc_no, ref_doc_id,
                    from_warehouse_id, operator_id, operator_name, remark
                )
                SELECT id, serial_code, 'stock_out', 'stock_out', v_so.stock_out_no, p_stock_out_id,
                       v_inv.warehouse_id, p_operator_id, v_operator_name, '出库确认-编码已领用(手动)'
                FROM picked;

				GET DIAGNOSTICS v_selected_cnt = ROW_COUNT;
				IF v_selected_cnt <> v_need THEN
					RAISE EXCEPTION '部分备货编码状态已变化，请重新备货';
				END IF;
			END IF;
		END IF;
	END LOOP;

	UPDATE stock_out
	SET status = 'confirmed',
		confirmed_at = NOW(),
		updated_by = p_operator_id,
		updated_at = NOW()
	WHERE id = p_stock_out_id;

	IF v_so.ref_doc_type = 'consumption_order' THEN
		UPDATE consumption_order SET status = 'completed', updated_at = NOW() WHERE id = v_so.ref_doc_id;
	END IF;

    DELETE FROM stock_out_item_serial_selection
    WHERE stock_out_item_id IN (SELECT id FROM stock_out_item WHERE stock_out_id = p_stock_out_id);
END;
$procedure$;

-- 存储过程: sp_confirm_transfer_in
DROP PROCEDURE IF EXISTS sp_confirm_transfer_in CASCADE;
CREATE OR REPLACE PROCEDURE public.sp_confirm_transfer_in(IN p_transfer_id bigint, IN p_operator_id bigint)
 LANGUAGE plpgsql
AS $procedure$
DECLARE
    v_tr      RECORD;
    v_item    RECORD;
    v_balance NUMERIC(18, 3);
    v_inv_id  BIGINT;
BEGIN
    SELECT id, transfer_no, to_warehouse_id, transfer_status
    INTO v_tr
    FROM stock_transfer
    WHERE id = p_transfer_id
      AND deleted_at IS NULL
    FOR UPDATE;

    IF v_tr IS NULL THEN
        RAISE EXCEPTION '调拨单不存在';
    END IF;

    IF v_tr.transfer_status <> 'in_transit' THEN
        RAISE EXCEPTION '调拨单当前状态[%]不允许确认为调入', v_tr.transfer_status;
    END IF;

    FOR v_item IN
        SELECT sti.*, i.unit, i.unit_cost, i.material_inspection_no
        FROM stock_transfer_item sti
        INNER JOIN inventory i ON i.id = sti.inventory_id
        WHERE sti.transfer_id = p_transfer_id
    LOOP
        INSERT INTO inventory (
            material_id, warehouse_id, quantity, unit, unit_cost,
            material_inspection_no, updated_at
        )
        VALUES (
            v_item.material_id, v_tr.to_warehouse_id, v_item.quantity, v_item.unit, v_item.unit_cost,
            v_item.material_inspection_no, NOW()
        )
        ON CONFLICT (material_id, warehouse_id, COALESCE(material_inspection_no, ''))
            DO UPDATE SET quantity   = inventory.quantity + EXCLUDED.quantity,
                          updated_at = NOW()
        RETURNING id INTO v_inv_id;

        SELECT quantity INTO v_balance FROM inventory WHERE id = v_inv_id;

        INSERT INTO inventory_transaction (
            material_id, warehouse_id, trans_type, quantity, balance, ref_doc_type, ref_doc_no,
            ref_doc_id, operator_id, remark, created_at
        )
        VALUES (
            v_item.material_id, v_tr.to_warehouse_id, 'in', v_item.quantity, v_balance, 'stock_transfer',
            v_tr.transfer_no, p_transfer_id, p_operator_id, '调拨调入', NOW());
    END LOOP;

    UPDATE stock_transfer
    SET transfer_status = 'completed',
        updated_at = NOW(),
        updated_by = p_operator_id
    WHERE id = p_transfer_id;
END;
$procedure$;

-- 存储过程: sp_confirm_transfer_out
DROP PROCEDURE IF EXISTS sp_confirm_transfer_out CASCADE;
CREATE OR REPLACE PROCEDURE public.sp_confirm_transfer_out(IN p_transfer_id bigint, IN p_operator_id bigint)
 LANGUAGE plpgsql
AS $procedure$
DECLARE
    v_tr           RECORD;
    v_item         RECORD;
    v_balance      NUMERIC(18,3);
    v_inv_id       BIGINT;
BEGIN
    SELECT id, transfer_no, from_warehouse_id, transfer_status
    INTO v_tr
    FROM stock_transfer
    WHERE id = p_transfer_id AND deleted_at IS NULL
    FOR UPDATE;

    IF v_tr IS NULL THEN
        RAISE EXCEPTION '调拨单不存在';
    END IF;

    IF v_tr.transfer_status <> 'draft' THEN
        RAISE EXCEPTION '调拨单当前状态[%]不允许确认为调出', v_tr.transfer_status;
    END IF;

    FOR v_item IN
        SELECT sti.*, m.material_code, m.material_name
        FROM stock_transfer_item sti
        INNER JOIN material m ON m.id = sti.material_id
        WHERE sti.transfer_id = p_transfer_id
    LOOP
        -- 调拨明细应当指定 inventory_id
        IF v_item.inventory_id IS NOT NULL AND v_item.inventory_id > 0 THEN
            UPDATE inventory
            SET quantity = quantity - v_item.quantity,
                updated_at = NOW()
            WHERE id = v_item.inventory_id
            RETURNING id INTO v_inv_id;
        ELSE
             UPDATE inventory
             SET quantity = quantity - v_item.quantity,
                 updated_at = NOW()
             WHERE material_id = v_item.material_id
               AND warehouse_id = v_tr.from_warehouse_id
               AND COALESCE(material_inspection_no, '') = ''
             RETURNING id INTO v_inv_id;
        END IF;

        IF NOT FOUND THEN
            RAISE EXCEPTION '仓库内无物料[%]的库存记录', v_item.material_code;
        END IF;

        SELECT quantity INTO v_balance FROM inventory WHERE id = v_inv_id;

        IF v_balance < 0 THEN
            RAISE EXCEPTION '调出数量超过库存余量 [物料=%]', v_item.material_code;
        END IF;

        INSERT INTO inventory_transaction (
            material_id, warehouse_id, trans_type, quantity,
            balance, ref_doc_type, ref_doc_no, ref_doc_id, operator_id,
            remark, created_at
        )
        VALUES (
            v_item.material_id, v_tr.from_warehouse_id,
            'out', -v_item.quantity, v_balance,
            'stock_transfer', v_tr.transfer_no, p_transfer_id,
            p_operator_id, '调拨调出', NOW()
        );
    END LOOP;

    UPDATE stock_transfer
    SET transfer_status = 'in_transit',
        updated_at = NOW(),
        updated_by = p_operator_id
    WHERE id = p_transfer_id;
END;
$procedure$;

-- 存储过程: sp_ship_sales_order_item
DROP PROCEDURE IF EXISTS sp_ship_sales_order_item CASCADE;
CREATE OR REPLACE PROCEDURE public.sp_ship_sales_order_item(IN p_order_id bigint, IN p_material_id bigint, IN p_ship_qty numeric, IN p_operator_id bigint)
 LANGUAGE plpgsql
AS $procedure$
DECLARE
    v_so            RECORD;
    v_soi           RECORD;
    v_inv           RECORD;
    v_remaining_ship NUMERIC(18,3) := p_ship_qty;
    v_pop_qty       NUMERIC(18,3);
    v_all_shipped   BOOLEAN;
BEGIN
    SELECT id, order_no, order_status
    INTO v_so
    FROM sales_order
    WHERE id = p_order_id AND deleted_at IS NULL
    FOR UPDATE;

    IF v_so IS NULL THEN
        RAISE EXCEPTION '销售订单不存在';
    END IF;

    IF v_so.order_status NOT IN ('confirmed', 'preparing') THEN
        RAISE EXCEPTION '订单当前状态[%]不允许发货', v_so.order_status;
    END IF;

    SELECT id, quantity, shipped_quantity
    INTO v_soi
    FROM sales_order_item
    WHERE order_id = p_order_id AND material_id = p_material_id
    FOR UPDATE;

    IF v_soi IS NULL THEN
        RAISE EXCEPTION '订单明细中不存在该物料';
    END IF;

    IF v_soi.shipped_quantity + p_ship_qty > v_soi.quantity THEN
        RAISE EXCEPTION '累计发货数量超过订单数量';
    END IF;

    FOR v_inv IN 
        SELECT id, warehouse_id, quantity, locked_quantity
        FROM inventory
        WHERE material_id = p_material_id AND quantity > 0
        ORDER BY stock_in_date ASC NULLS LAST
        FOR UPDATE
    LOOP
        EXIT WHEN v_remaining_ship <= 0;

        v_pop_qty := LEAST(v_inv.quantity, v_remaining_ship);
        
        UPDATE inventory
        SET quantity = quantity - v_pop_qty,
            locked_quantity = GREATEST(locked_quantity - v_pop_qty, 0),
            updated_at = NOW()
        WHERE id = v_inv.id;

        INSERT INTO inventory_transaction (
            material_id, warehouse_id, trans_type, quantity,
            ref_doc_type, ref_doc_no, ref_doc_id, operator_id, created_at
        )
        VALUES (
            p_material_id, v_inv.warehouse_id, 'out', -v_pop_qty,
            'sales_order', v_so.order_no, p_order_id, p_operator_id, NOW()
        );

        v_remaining_ship := v_remaining_ship - v_pop_qty;
    END LOOP;

    IF v_remaining_ship > 0 THEN
        RAISE EXCEPTION '实物库存不足，无法完成发货 [缺额:%]', v_remaining_ship;
    END IF;

    UPDATE sales_order_item
    SET shipped_quantity = shipped_quantity + p_ship_qty,
        updated_at = NOW()
    WHERE id = v_soi.id;

    SELECT NOT EXISTS (
        SELECT 1 FROM sales_order_item 
        WHERE order_id = p_order_id AND shipped_quantity < quantity
    ) INTO v_all_shipped;

    UPDATE sales_order
    SET order_status = CASE WHEN v_all_shipped THEN 'shipped' ELSE 'preparing' END,
        updated_at = NOW(),
        updated_by = p_operator_id
    WHERE id = p_order_id;
END;
$procedure$;

-- 存储过程: sp_submit_purchase_return
DROP PROCEDURE IF EXISTS sp_submit_purchase_return CASCADE;
CREATE OR REPLACE PROCEDURE public.sp_submit_purchase_return(IN p_return_id bigint, IN p_operator_id bigint)
 LANGUAGE plpgsql
AS $procedure$
DECLARE
    v_ro            RECORD;
    v_item          RECORD;
    v_inv           RECORD;
    v_supplier_name VARCHAR(200);
    v_available     NUMERIC(18, 3);
    v_stock_out_id  BIGINT;
    v_stock_out_no  VARCHAR(20);
BEGIN
    SELECT ro.*
    INTO v_ro
    FROM return_order ro
    WHERE ro.id = p_return_id
      AND ro.deleted_at IS NULL
    FOR UPDATE;

    IF v_ro IS NULL THEN
        RAISE EXCEPTION '退货单不存在';
    END IF;

    IF v_ro.return_type <> 'purchase_return' THEN
        RAISE EXCEPTION '非采购退货单不允许走采购退货提交流程';
    END IF;

    IF v_ro.return_status <> 'draft' THEN
        RAISE EXCEPTION '退货单当前状态[%]不允许提交', v_ro.return_status;
    END IF;

    IF COALESCE(v_ro.supplier_code, '') = '' THEN
        RAISE EXCEPTION '采购退货必须选择供应商';
    END IF;

    IF v_ro.stock_out_id IS NOT NULL THEN
        RAISE EXCEPTION '该退货单已生成出库单';
    END IF;

    SELECT COALESCE(s.supplier_name, '')
    INTO v_supplier_name
    FROM supplier s
    WHERE s.supplier_code = v_ro.supplier_code
      AND s.deleted_at IS NULL;

    v_stock_out_no := fn_generate_serial_no('SO');

    INSERT INTO stock_out (
        stock_out_no, stock_out_date, out_type, ref_doc_type, ref_doc_id,
        receiver, status, remark, created_by
    )
    VALUES (
        v_stock_out_no, v_ro.return_date, 'purchase_return', 'purchase_return', p_return_id,
        v_supplier_name, 'pending', v_ro.remark, p_operator_id
    )
    RETURNING id INTO v_stock_out_id;

    FOR v_item IN
        SELECT roi.*
        FROM return_order_item roi
        WHERE roi.return_id = p_return_id
    LOOP
        IF v_item.inventory_id IS NULL OR v_item.inventory_id = 0 THEN
            RAISE EXCEPTION '采购退货明细必须指定库存批次(inventory_id)';
        END IF;

        SELECT i.id, i.warehouse_id, i.material_id,
               i.quantity, i.locked_quantity, COALESCE(i.in_transit_quantity, 0) AS in_transit_quantity
        INTO v_inv
        FROM inventory i
        WHERE i.id = v_item.inventory_id
        FOR UPDATE;

        IF v_inv IS NULL THEN
            RAISE EXCEPTION '库存不存在 [inventory_id=%]', v_item.inventory_id;
        END IF;

        IF v_item.warehouse_id IS NOT NULL AND v_item.warehouse_id <> v_inv.warehouse_id THEN
            RAISE EXCEPTION '退货明细仓库与库存批次不一致';
        END IF;

        v_available := (v_inv.quantity - v_inv.locked_quantity - v_inv.in_transit_quantity);
        IF v_available < v_item.quantity THEN
            RAISE EXCEPTION '库存不足，需:% 可用:% [inventory_id=%]', v_item.quantity, v_available, v_item.inventory_id;
        END IF;

        UPDATE inventory
        SET in_transit_quantity = COALESCE(in_transit_quantity, 0) + v_item.quantity,
            updated_at = NOW()
        WHERE id = v_item.inventory_id;

        INSERT INTO stock_out_item (
            stock_out_id, material_id, inventory_id, quantity, unit
        )
        VALUES (
            v_stock_out_id, v_item.material_id, v_item.inventory_id, v_item.quantity, v_item.unit
        );
    END LOOP;

    UPDATE return_order
    SET return_status = 'pending_out',
        stock_out_id  = v_stock_out_id,
        updated_at    = NOW(),
        updated_by    = p_operator_id
    WHERE id = p_return_id;
END;
$procedure$;

-- 存储过程: sp_write_audit_log
DROP PROCEDURE IF EXISTS sp_write_audit_log CASCADE;
CREATE OR REPLACE PROCEDURE public.sp_write_audit_log(IN p_user_id bigint, IN p_username character varying, IN p_action character varying, IN p_module character varying, IN p_target_type character varying DEFAULT NULL::character varying, IN p_target_id bigint DEFAULT NULL::bigint, IN p_detail jsonb DEFAULT NULL::jsonb)
 LANGUAGE plpgsql
AS $procedure$
BEGIN
    INSERT INTO sys_audit_log (user_id, username, action, module, target_type, target_id, detail, created_at)
    VALUES (p_user_id, p_username, p_action, p_module, p_target_type, p_target_id, p_detail, NOW());
END;
$procedure$;

-- =====================================================
-- 触发器定义
-- =====================================================

-- 触发器: trg_purchase_order_item_amount
DROP TRIGGER IF EXISTS trg_purchase_order_item_amount ON purchase_order_item;
CREATE TRIGGER trg_purchase_order_item_amount BEFORE INSERT ON purchase_order_item FOR EACH ROW EXECUTE FUNCTION trf_calc_item_amount();

-- 触发器: trg_sync_po_amount
DROP TRIGGER IF EXISTS trg_sync_po_amount ON purchase_order_item;
CREATE TRIGGER trg_sync_po_amount BEFORE UPDATE ON purchase_order_item FOR EACH ROW EXECUTE FUNCTION trf_sync_purchase_order_amount();

