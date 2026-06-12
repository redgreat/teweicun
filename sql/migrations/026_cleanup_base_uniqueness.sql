-- MIGRATION_ID: 026_cleanup_base_uniqueness
-- MIGRATION_APPLIED: pending
-- 功能：清理重复基础数据与测试单据，并补齐基础资料名称唯一约束和仓库白名单

UPDATE warehouse
SET warehouse_name = CASE warehouse_code
    WHEN 'WH001' THEN '主材料库'
    WHEN 'W0001' THEN '成品库'
    WHEN 'W0002' THEN '焊材库'
    WHEN 'W0003' THEN '辅材库'
    ELSE warehouse_name
END,
warehouse_type = CASE warehouse_code
    WHEN 'WH001' THEN 'raw_material'
    WHEN 'W0001' THEN 'finished'
    WHEN 'W0002' THEN 'welding'
    WHEN 'W0003' THEN 'raw_material'
    ELSE warehouse_type
END,
status = 'enabled',
deleted_at = NULL,
updated_at = NOW()
WHERE warehouse_code IN ('WH001', 'W0001', 'W0002', 'W0003');

CREATE TEMP TABLE cleanup_warehouse_ids (id bigint PRIMARY KEY) ON COMMIT DROP;
INSERT INTO cleanup_warehouse_ids (id)
SELECT id
FROM warehouse
WHERE deleted_at IS NULL
  AND warehouse_code NOT IN ('WH001', 'W0001', 'W0002', 'W0003');

CREATE TEMP TABLE cleanup_category_ids (id bigint PRIMARY KEY) ON COMMIT DROP;
INSERT INTO cleanup_category_ids (id)
SELECT mc.id
FROM material_category mc
WHERE mc.deleted_at IS NULL
  AND (
      mc.category_name = '测试分类'
      OR mc.id <> (
          SELECT min(mc2.id)
          FROM material_category mc2
          WHERE mc2.deleted_at IS NULL
            AND mc2.category_name = mc.category_name
      )
  );

CREATE TEMP TABLE cleanup_supplier_ids (id bigint PRIMARY KEY) ON COMMIT DROP;
INSERT INTO cleanup_supplier_ids (id)
SELECT s.id
FROM supplier s
WHERE s.deleted_at IS NULL
  AND (
      s.supplier_name = '华钢'
      OR s.id <> (
          SELECT min(s2.id)
          FROM supplier s2
          WHERE s2.deleted_at IS NULL
            AND s2.supplier_name = s.supplier_name
      )
  );

CREATE TEMP TABLE cleanup_customer_ids (id bigint PRIMARY KEY) ON COMMIT DROP;
INSERT INTO cleanup_customer_ids (id)
SELECT c.id
FROM customer c
WHERE c.deleted_at IS NULL
  AND (
      c.customer_name IN ('中容', '测试客户')
      OR c.id <> (
          SELECT min(c2.id)
          FROM customer c2
          WHERE c2.deleted_at IS NULL
            AND c2.customer_name = c.customer_name
      )
  );

CREATE TEMP TABLE cleanup_material_ids (id bigint PRIMARY KEY) ON COMMIT DROP;
INSERT INTO cleanup_material_ids (id)
SELECT m.id
FROM material m
WHERE m.deleted_at IS NULL
  AND (
      m.category_id IN (SELECT id FROM cleanup_category_ids)
      OR m.default_warehouse_id IN (SELECT id FROM cleanup_warehouse_ids)
      OR m.material_name IN (
          '封头（材质：锰钢，规格：一米，工序：压制）',
          '并发测试件',
          '焊丝（牌号：焊材，直径：一点二，包装：盘装）',
          '编码成品',
          '编码钢板',
          '钢板（材质：锰钢，标准：国标，厚度：十六，用途：筒体）'
      )
      OR m.id <> (
          SELECT min(m2.id)
          FROM material m2
          WHERE m2.deleted_at IS NULL
            AND m2.material_name = m.material_name
      )
  );

CREATE TEMP TABLE cleanup_user_ids (id bigint PRIMARY KEY) ON COMMIT DROP;
INSERT INTO cleanup_user_ids (id)
SELECT u.id
FROM sys_user u
WHERE u.deleted_at IS NULL
  AND (
      u.username ILIKE 'E2E%'
      OR u.username ILIKE 'TEST%'
      OR u.id <> (
          SELECT min(u2.id)
          FROM sys_user u2
          WHERE u2.deleted_at IS NULL
            AND u2.real_name = u.real_name
      )
  );

CREATE TEMP TABLE cleanup_role_ids (id bigint PRIMARY KEY) ON COMMIT DROP;
INSERT INTO cleanup_role_ids (id)
SELECT r.id
FROM sys_role r
WHERE r.deleted_at IS NULL
  AND (
      r.role_code ILIKE 'E2E%'
      OR r.role_code ILIKE 'TEST%'
      OR r.id <> (
          SELECT min(r2.id)
          FROM sys_role r2
          WHERE r2.deleted_at IS NULL
            AND r2.role_name = r.role_name
      )
  );

CREATE TEMP TABLE cleanup_inventory_ids (id bigint PRIMARY KEY) ON COMMIT DROP;
INSERT INTO cleanup_inventory_ids (id)
SELECT i.id
FROM inventory i
WHERE i.material_id IN (SELECT id FROM cleanup_material_ids)
   OR i.warehouse_id IN (SELECT id FROM cleanup_warehouse_ids);

CREATE TEMP TABLE cleanup_purchase_order_ids (id bigint PRIMARY KEY) ON COMMIT DROP;
INSERT INTO cleanup_purchase_order_ids (id)
SELECT DISTINCT po.id
FROM purchase_order po
LEFT JOIN purchase_order_item poi ON poi.order_id = po.id
WHERE po.deleted_at IS NULL
  AND (
      po.supplier_id IN (SELECT id FROM cleanup_supplier_ids)
      OR po.supplier_code IN (SELECT supplier_code FROM supplier WHERE id IN (SELECT id FROM cleanup_supplier_ids))
      OR po.created_by IN (SELECT id FROM cleanup_user_ids)
      OR poi.material_id IN (SELECT id FROM cleanup_material_ids)
  );

CREATE TEMP TABLE cleanup_sales_order_ids (id bigint PRIMARY KEY) ON COMMIT DROP;
INSERT INTO cleanup_sales_order_ids (id)
SELECT DISTINCT so.id
FROM sales_order so
LEFT JOIN sales_order_item soi ON soi.order_id = so.id
WHERE so.deleted_at IS NULL
  AND (
      so.customer_id IN (SELECT id FROM cleanup_customer_ids)
      OR so.customer_code IN (SELECT customer_code FROM customer WHERE id IN (SELECT id FROM cleanup_customer_ids))
      OR so.created_by IN (SELECT id FROM cleanup_user_ids)
      OR soi.material_id IN (SELECT id FROM cleanup_material_ids)
  );

CREATE TEMP TABLE cleanup_stock_in_ids (id bigint PRIMARY KEY) ON COMMIT DROP;
INSERT INTO cleanup_stock_in_ids (id)
SELECT DISTINCT si.id
FROM stock_in si
LEFT JOIN stock_in_item sii ON sii.stock_in_id = si.id
WHERE si.deleted_at IS NULL
  AND (
      si.purchase_order_id IN (SELECT id FROM cleanup_purchase_order_ids)
      OR si.supplier_id IN (SELECT id FROM cleanup_supplier_ids)
      OR si.supplier_code IN (SELECT supplier_code FROM supplier WHERE id IN (SELECT id FROM cleanup_supplier_ids))
      OR si.warehouse_id IN (SELECT id FROM cleanup_warehouse_ids)
      OR si.warehouse_code IN (SELECT warehouse_code FROM warehouse WHERE id IN (SELECT id FROM cleanup_warehouse_ids))
      OR si.created_by IN (SELECT id FROM cleanup_user_ids)
      OR sii.material_id IN (SELECT id FROM cleanup_material_ids)
  );

CREATE TEMP TABLE cleanup_stock_out_ids (id bigint PRIMARY KEY) ON COMMIT DROP;
INSERT INTO cleanup_stock_out_ids (id)
SELECT DISTINCT so.id
FROM stock_out so
LEFT JOIN stock_out_item soi ON soi.stock_out_id = so.id
WHERE so.deleted_at IS NULL
  AND (
      so.created_by IN (SELECT id FROM cleanup_user_ids)
      OR soi.material_id IN (SELECT id FROM cleanup_material_ids)
      OR soi.inventory_id IN (SELECT id FROM cleanup_inventory_ids)
      OR (so.ref_doc_type = 'sales_order' AND so.ref_doc_id IN (SELECT id FROM cleanup_sales_order_ids))
  );

CREATE TEMP TABLE cleanup_consumption_order_ids (id bigint PRIMARY KEY) ON COMMIT DROP;
INSERT INTO cleanup_consumption_order_ids (id)
SELECT DISTINCT co.id
FROM consumption_order co
LEFT JOIN consumption_order_item coi ON coi.order_id = co.id
WHERE co.deleted_at IS NULL
  AND (
      co.created_by IN (SELECT id FROM cleanup_user_ids)
      OR co.designer_id IN (SELECT id FROM cleanup_user_ids)
      OR co.stock_out_id IN (SELECT id FROM cleanup_stock_out_ids)
      OR coi.material_id IN (SELECT id FROM cleanup_material_ids)
      OR coi.inventory_id IN (SELECT id FROM cleanup_inventory_ids)
      OR coi.warehouse_id IN (SELECT id FROM cleanup_warehouse_ids)
      OR coi.warehouse_code IN (SELECT warehouse_code FROM warehouse WHERE id IN (SELECT id FROM cleanup_warehouse_ids))
  );

INSERT INTO cleanup_stock_out_ids (id)
SELECT stock_out_id
FROM consumption_order
WHERE id IN (SELECT id FROM cleanup_consumption_order_ids)
  AND stock_out_id IS NOT NULL
ON CONFLICT DO NOTHING;

CREATE TEMP TABLE cleanup_reversal_order_ids (id bigint PRIMARY KEY) ON COMMIT DROP;
INSERT INTO cleanup_reversal_order_ids (id)
SELECT DISTINCT ro.id
FROM reversal_order ro
LEFT JOIN reversal_order_item roi ON roi.order_id = ro.id
WHERE ro.deleted_at IS NULL
  AND (
      ro.created_by IN (SELECT id FROM cleanup_user_ids)
      OR ro.designer_id IN (SELECT id FROM cleanup_user_ids)
      OR ro.stock_in_id IN (SELECT id FROM cleanup_stock_in_ids)
      OR ro.warehouse_id IN (SELECT id FROM cleanup_warehouse_ids)
      OR ro.warehouse_code IN (SELECT warehouse_code FROM warehouse WHERE id IN (SELECT id FROM cleanup_warehouse_ids))
      OR roi.material_id IN (SELECT id FROM cleanup_material_ids)
      OR roi.inventory_id IN (SELECT id FROM cleanup_inventory_ids)
  );

INSERT INTO cleanup_stock_in_ids (id)
SELECT stock_in_id
FROM reversal_order
WHERE id IN (SELECT id FROM cleanup_reversal_order_ids)
  AND stock_in_id IS NOT NULL
ON CONFLICT DO NOTHING;

CREATE TEMP TABLE cleanup_return_order_ids (id bigint PRIMARY KEY) ON COMMIT DROP;
INSERT INTO cleanup_return_order_ids (id)
SELECT DISTINCT ro.id
FROM return_order ro
LEFT JOIN return_order_item roi ON roi.return_id = ro.id
WHERE ro.deleted_at IS NULL
  AND (
      ro.created_by IN (SELECT id FROM cleanup_user_ids)
      OR ro.stock_out_id IN (SELECT id FROM cleanup_stock_out_ids)
      OR ro.warehouse_id IN (SELECT id FROM cleanup_warehouse_ids)
      OR ro.warehouse_code IN (SELECT warehouse_code FROM warehouse WHERE id IN (SELECT id FROM cleanup_warehouse_ids))
      OR ro.supplier_code IN (SELECT supplier_code FROM supplier WHERE id IN (SELECT id FROM cleanup_supplier_ids))
      OR ro.ref_doc_id IN (SELECT id FROM cleanup_purchase_order_ids)
      OR ro.ref_doc_id IN (SELECT id FROM cleanup_sales_order_ids)
      OR roi.material_id IN (SELECT id FROM cleanup_material_ids)
      OR roi.inventory_id IN (SELECT id FROM cleanup_inventory_ids)
      OR roi.warehouse_id IN (SELECT id FROM cleanup_warehouse_ids)
      OR roi.warehouse_code IN (SELECT warehouse_code FROM warehouse WHERE id IN (SELECT id FROM cleanup_warehouse_ids))
  );

INSERT INTO cleanup_stock_out_ids (id)
SELECT stock_out_id
FROM return_order
WHERE id IN (SELECT id FROM cleanup_return_order_ids)
  AND stock_out_id IS NOT NULL
ON CONFLICT DO NOTHING;

CREATE TEMP TABLE cleanup_fund_payment_ids (id bigint PRIMARY KEY) ON COMMIT DROP;
INSERT INTO cleanup_fund_payment_ids (id)
SELECT DISTINCT fp.id
FROM fund_payment fp
LEFT JOIN fund_payment_item fpi ON fpi.statement_id = fp.id
WHERE fp.deleted_at IS NULL
  AND (
      fp.supplier_id IN (SELECT id FROM cleanup_supplier_ids)
      OR fpi.source_order_id IN (SELECT id FROM cleanup_purchase_order_ids)
      OR fpi.source_order_id IN (SELECT id FROM cleanup_return_order_ids)
  );

CREATE TEMP TABLE cleanup_fund_collection_ids (id bigint PRIMARY KEY) ON COMMIT DROP;
INSERT INTO cleanup_fund_collection_ids (id)
SELECT DISTINCT fc.id
FROM fund_collection fc
LEFT JOIN fund_collection_item fci ON fci.statement_id = fc.id
WHERE fc.deleted_at IS NULL
  AND (
      fc.customer_id IN (SELECT id FROM cleanup_customer_ids)
      OR fci.source_order_id IN (SELECT id FROM cleanup_sales_order_ids)
      OR fci.source_order_id IN (SELECT id FROM cleanup_return_order_ids)
  );

UPDATE fund_payment
SET deleted_at = NOW(), updated_at = NOW(), status = 'cancelled'
WHERE id IN (SELECT id FROM cleanup_fund_payment_ids);

UPDATE fund_collection
SET deleted_at = NOW(), updated_at = NOW(), status = 'cancelled'
WHERE id IN (SELECT id FROM cleanup_fund_collection_ids);

UPDATE stock_out
SET deleted_at = NOW(), updated_at = NOW(), status = 'cancelled'
WHERE id IN (SELECT id FROM cleanup_stock_out_ids);

UPDATE stock_in
SET deleted_at = NOW(), updated_at = NOW(), stock_in_status = 'cancelled'
WHERE id IN (SELECT id FROM cleanup_stock_in_ids);

UPDATE purchase_order
SET deleted_at = NOW(), updated_at = NOW(), order_status = 'cancelled'
WHERE id IN (SELECT id FROM cleanup_purchase_order_ids);

UPDATE sales_order
SET deleted_at = NOW(), updated_at = NOW(), order_status = 'cancelled'
WHERE id IN (SELECT id FROM cleanup_sales_order_ids);

UPDATE consumption_order
SET deleted_at = NOW(), updated_at = NOW(), status = 'cancelled'
WHERE id IN (SELECT id FROM cleanup_consumption_order_ids);

UPDATE reversal_order
SET deleted_at = NOW(), updated_at = NOW(), status = 'cancelled'
WHERE id IN (SELECT id FROM cleanup_reversal_order_ids);

UPDATE return_order
SET deleted_at = NOW(), updated_at = NOW(), return_status = 'closed'
WHERE id IN (SELECT id FROM cleanup_return_order_ids);

UPDATE inventory
SET quantity = 0,
    locked_quantity = 0,
    in_transit_quantity = 0,
    updated_at = NOW()
WHERE id IN (SELECT id FROM cleanup_inventory_ids);

DO $$
BEGIN
    IF to_regclass('public.material_attribute_value') IS NOT NULL THEN
        EXECUTE 'UPDATE material_attribute_value
                 SET deleted_at = NOW(), updated_at = NOW()
                 WHERE material_id IN (SELECT id FROM cleanup_material_ids)
                   AND deleted_at IS NULL';
    END IF;
END $$;

UPDATE material
SET deleted_at = NOW(), updated_at = NOW(), status = 'disabled'
WHERE id IN (SELECT id FROM cleanup_material_ids);

UPDATE material_category
SET deleted_at = NOW(), updated_at = NOW(), status = 'disabled'
WHERE id IN (SELECT id FROM cleanup_category_ids);

UPDATE supplier
SET deleted_at = NOW(), updated_at = NOW(), status = 'disabled'
WHERE id IN (SELECT id FROM cleanup_supplier_ids);

UPDATE customer
SET deleted_at = NOW(), updated_at = NOW(), status = 'disabled'
WHERE id IN (SELECT id FROM cleanup_customer_ids);

UPDATE warehouse
SET deleted_at = NOW(), updated_at = NOW(), status = 'disabled'
WHERE id IN (SELECT id FROM cleanup_warehouse_ids);

DELETE FROM sys_user_role
WHERE user_id IN (SELECT id FROM cleanup_user_ids)
   OR role_id IN (SELECT id FROM cleanup_role_ids);

DELETE FROM sys_role_permission
WHERE role_id IN (SELECT id FROM cleanup_role_ids);

UPDATE sys_user
SET deleted_at = NOW(), updated_at = NOW(), status = 'disabled'
WHERE id IN (SELECT id FROM cleanup_user_ids);

UPDATE sys_role
SET deleted_at = NOW(), updated_at = NOW(), status = 'disabled'
WHERE id IN (SELECT id FROM cleanup_role_ids);

UPDATE sys_dict_data
SET status = 'disabled', updated_at = NOW()
WHERE dict_type = 'warehouse_type'
  AND dict_value NOT IN ('raw_material', 'welding', 'finished');

DROP VIEW IF EXISTS v_inventory_detail;
CREATE VIEW v_inventory_detail AS
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
    i.quantity,
    i.locked_quantity,
    i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0) AS available_quantity,
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
WHERE m.deleted_at IS NULL
  AND w.deleted_at IS NULL;

DROP VIEW IF EXISTS v_inventory_report;
CREATE VIEW v_inventory_report AS
SELECT
    w.warehouse_name,
    mc.category_name,
    m.material_code,
    m.material_name,
    m.unit,
    SUM(i.quantity) AS current_quantity,
    SUM(i.locked_quantity) AS locked_quantity,
    SUM(COALESCE(i.in_transit_quantity, 0)) AS in_transit_quantity,
    SUM(i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0)) AS available_quantity
FROM inventory i
JOIN material m ON i.material_id = m.id
LEFT JOIN material_category mc ON m.category_id = mc.id AND mc.deleted_at IS NULL
JOIN warehouse w ON i.warehouse_id = w.id
WHERE m.deleted_at IS NULL
  AND w.deleted_at IS NULL
GROUP BY w.warehouse_name, mc.category_name, m.material_code, m.material_name, m.unit;

DROP VIEW IF EXISTS v_inventory_summary;
CREATE VIEW v_inventory_summary AS
SELECT
    m.id AS material_id,
    m.material_code,
    m.material_name,
    mc.category_name,
    m.unit,
    m.safety_stock,
    m.max_stock,
    COALESCE(SUM(i.quantity), 0) AS total_quantity,
    COALESCE(SUM(i.locked_quantity), 0) AS locked_quantity,
    COALESCE(SUM(COALESCE(i.in_transit_quantity, 0)), 0) AS in_transit_quantity,
    COALESCE(SUM(i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0)), 0) AS available_quantity,
    COALESCE(COUNT(i.id), 0) AS batch_count
FROM material m
LEFT JOIN material_category mc ON mc.id = m.category_id AND mc.deleted_at IS NULL
LEFT JOIN inventory i ON i.material_id = m.id
    AND i.quantity > 0
    AND EXISTS (
        SELECT 1
        FROM warehouse w
        WHERE w.id = i.warehouse_id
          AND w.deleted_at IS NULL
    )
WHERE m.deleted_at IS NULL
GROUP BY m.id, m.material_code, m.material_name, mc.category_name, m.unit, m.safety_stock, m.max_stock;

CREATE UNIQUE INDEX IF NOT EXISTS uk_material_name_active
    ON material (material_name)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uk_material_category_name_active
    ON material_category (category_name)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uk_supplier_name_active
    ON supplier (supplier_name)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uk_customer_name_active
    ON customer (customer_name)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uk_warehouse_name_active
    ON warehouse (warehouse_name)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_user_real_name_active
    ON sys_user (real_name)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_role_name_active
    ON sys_role (role_name)
    WHERE deleted_at IS NULL;

ALTER TABLE warehouse DROP CONSTRAINT IF EXISTS ck_warehouse_allowed_names;
ALTER TABLE warehouse ADD CONSTRAINT ck_warehouse_allowed_names
    CHECK (deleted_at IS NOT NULL OR warehouse_name IN ('主材料库', '成品库', '焊材库', '辅材库'));

ALTER TABLE warehouse DROP CONSTRAINT IF EXISTS ck_warehouse_allowed_types;
ALTER TABLE warehouse ADD CONSTRAINT ck_warehouse_allowed_types
    CHECK (deleted_at IS NOT NULL OR warehouse_type IN ('raw_material', 'welding', 'finished'));

COMMENT ON INDEX uk_material_name_active IS '物料名称在未删除数据中唯一';
COMMENT ON INDEX uk_material_category_name_active IS '物料分类名称在未删除数据中唯一';
COMMENT ON INDEX uk_supplier_name_active IS '供应商名称在未删除数据中唯一';
COMMENT ON INDEX uk_customer_name_active IS '客户名称在未删除数据中唯一';
COMMENT ON INDEX uk_warehouse_name_active IS '仓库名称在未删除数据中唯一';
COMMENT ON INDEX uk_sys_user_real_name_active IS '用户姓名在未删除数据中唯一';
COMMENT ON INDEX uk_sys_role_name_active IS '角色名称在未删除数据中唯一';
