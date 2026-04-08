-- ============================================================
-- 12_authz_policy_change_rollback_ref.sql
-- 说明：权限系统第六阶段，策略变更日志增加 rollback_of_id 与回滚幂等约束
-- ============================================================

ALTER TABLE ty_authz_policy_changes
    ADD COLUMN IF NOT EXISTS rollback_of_id CHAR(36) NULL AFTER target_key;

CREATE INDEX idx_authz_policy_changes_rollback_of_id ON ty_authz_policy_changes(rollback_of_id, deleted_at);

-- MySQL 下 UNIQUE 对 NULL 值不冲突；rollback 记录会写入非空 rollback_of_id，因此可用于防重复回滚
CREATE UNIQUE INDEX uk_authz_policy_changes_rollback_once
    ON ty_authz_policy_changes(operation, rollback_of_id);
