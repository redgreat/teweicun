-- MIGRATION_ID: 027_fix_warehouse_type_scope
-- MIGRATION_APPLIED: pending
-- 功能：修正仓库规则为四类仓库类型唯一，仓库名称仅做唯一不做白名单

ALTER TABLE warehouse DROP CONSTRAINT IF EXISTS ck_warehouse_allowed_names;
ALTER TABLE warehouse DROP CONSTRAINT IF EXISTS ck_warehouse_allowed_types;

UPDATE warehouse
SET warehouse_name = '主材料库',
    warehouse_type = 'main_material',
    status = 'enabled',
    deleted_at = NULL,
    updated_at = NOW()
WHERE warehouse_code = 'WH001';

UPDATE warehouse
SET warehouse_name = '成品库',
    warehouse_type = 'finished',
    status = 'enabled',
    deleted_at = NULL,
    updated_at = NOW()
WHERE warehouse_code = 'W0001';

UPDATE warehouse
SET warehouse_name = '焊材库',
    warehouse_type = 'welding',
    status = 'enabled',
    deleted_at = NULL,
    updated_at = NOW()
WHERE warehouse_code = 'W0002';

UPDATE warehouse
SET warehouse_name = '辅材库',
    warehouse_type = 'auxiliary',
    status = 'enabled',
    deleted_at = NULL,
    updated_at = NOW()
WHERE warehouse_code = 'W0003';

UPDATE sys_dict_data
SET status = CASE
        WHEN dict_value IN ('main_material', 'finished', 'welding', 'auxiliary') THEN 'enabled'
        ELSE 'disabled'
    END,
    dict_label = CASE dict_value
        WHEN 'main_material' THEN '主材料库'
        WHEN 'finished' THEN '成品库'
        WHEN 'welding' THEN '焊材库'
        WHEN 'auxiliary' THEN '辅材库'
        ELSE dict_label
    END,
    updated_at = NOW()
WHERE dict_type = 'warehouse_type';

INSERT INTO sys_dict_data (dict_type, dict_value, dict_label, sort_order, is_default, remark, status)
VALUES
    ('warehouse_type', 'main_material', '主材料库', 1, true, '仓库类型', 'enabled'),
    ('warehouse_type', 'finished', '成品库', 2, false, '仓库类型', 'enabled'),
    ('warehouse_type', 'welding', '焊材库', 3, false, '仓库类型', 'enabled'),
    ('warehouse_type', 'auxiliary', '辅材库', 4, false, '仓库类型', 'enabled')
ON CONFLICT (dict_type, dict_value) DO UPDATE
SET dict_label = EXCLUDED.dict_label,
    sort_order = EXCLUDED.sort_order,
    status = 'enabled',
    updated_at = NOW();

ALTER TABLE warehouse DROP CONSTRAINT IF EXISTS ck_warehouse_allowed_types;
ALTER TABLE warehouse ADD CONSTRAINT ck_warehouse_allowed_types
    CHECK (deleted_at IS NOT NULL OR warehouse_type IN ('main_material', 'finished', 'welding', 'auxiliary'));

CREATE UNIQUE INDEX IF NOT EXISTS uk_warehouse_type_active
    ON warehouse (warehouse_type)
    WHERE deleted_at IS NULL;

COMMENT ON INDEX uk_warehouse_type_active IS '仓库类型在未删除数据中唯一';
