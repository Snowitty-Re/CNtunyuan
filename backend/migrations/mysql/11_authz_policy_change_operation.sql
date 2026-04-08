-- ============================================================
-- 11_authz_policy_change_operation.sql
-- 说明：权限系统第五阶段，策略变更日志增加 operation 字段
-- ============================================================

ALTER TABLE ty_authz_policy_changes
    ADD COLUMN IF NOT EXISTS operation VARCHAR(20) NOT NULL DEFAULT 'apply' AFTER operator_role;

UPDATE ty_authz_policy_changes
SET operation = 'apply'
WHERE operation IS NULL OR operation = '';

-- 说明：operation 约束与索引已在新版本 schema/bootstrap 中内置。
-- 历史库升级仅补字段与数据归一化，避免重复添加约束/索引导致迁移失败。
