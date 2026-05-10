-- 功能：删除物料属性定义与属性值表，仅保留 material.custom_attributes JSON
-- 创建时间：2026-05-09
-- 创建人：GPT-5.3-Codex
-- MIGRATION_ID: 005_drop_material_attribute_objects
-- MIGRATION_APPLIED: applied_20260510T110440

BEGIN;

DO $$
BEGIN
    IF to_regclass('public.sys_role_permission') IS NOT NULL THEN
        DELETE FROM sys_role_permission
        WHERE permission_id IN (
            SELECT id
            FROM sys_permission
            WHERE perm_code IN (
                'base:material-attribute',
                'material-attr:list',
                'material-attr:create',
                'material-attr:update',
                'material-attr:delete'
            )
        );
    END IF;
END $$;

DELETE FROM sys_permission
WHERE perm_code IN (
    'base:material-attribute',
    'material-attr:list',
    'material-attr:create',
    'material-attr:update',
    'material-attr:delete'
);

DROP TABLE IF EXISTS material_attribute_value;
DROP TABLE IF EXISTS material_attribute_def;

DROP SEQUENCE IF EXISTS material_attribute_value_id_seq;
DROP SEQUENCE IF EXISTS material_attribute_def_id_seq;

COMMIT;
