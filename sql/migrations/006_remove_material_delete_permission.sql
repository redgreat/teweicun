-- 功能：移除物料删除权限，物料仅允许启停用
-- 创建时间：2026-05-09
-- 创建人：GPT-5.3-Codex
-- MIGRATION_ID: 006_remove_material_delete_permission
-- MIGRATION_APPLIED: applied_20260510T110445

BEGIN;

DO $$
BEGIN
    IF to_regclass('public.sys_role_permission') IS NOT NULL THEN
        DELETE FROM sys_role_permission
        WHERE permission_id IN (
            SELECT id FROM sys_permission WHERE perm_code = 'material:delete'
        );
    END IF;
END $$;

DELETE FROM sys_permission
WHERE perm_code = 'material:delete';

COMMIT;
