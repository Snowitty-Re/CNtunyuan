-- ============================================================
-- 12_authz_policy_change_rollback_ref.sql
-- 说明：权限系统第六阶段，策略变更日志增加 rollback_of_id 与回滚幂等约束
-- ============================================================

ALTER TABLE ty_authz_policy_changes
    ADD COLUMN IF NOT EXISTS rollback_of_id UUID;

CREATE INDEX IF NOT EXISTS idx_authz_policy_changes_rollback_of_id
    ON ty_authz_policy_changes(rollback_of_id)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uk_authz_policy_changes_rollback_once
    ON ty_authz_policy_changes(rollback_of_id)
    WHERE operation = 'rollback' AND rollback_of_id IS NOT NULL AND deleted_at IS NULL;
