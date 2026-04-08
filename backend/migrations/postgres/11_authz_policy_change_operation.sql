-- ============================================================
-- 11_authz_policy_change_operation.sql
-- 说明：权限系统第五阶段，策略变更日志增加 operation 字段
-- ============================================================

ALTER TABLE ty_authz_policy_changes
    ADD COLUMN IF NOT EXISTS operation VARCHAR(20) NOT NULL DEFAULT 'apply';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'chk_authz_policy_changes_operation'
    ) THEN
        ALTER TABLE ty_authz_policy_changes
            ADD CONSTRAINT chk_authz_policy_changes_operation
            CHECK (operation IN ('apply', 'rollback'));
    END IF;
END $$;

UPDATE ty_authz_policy_changes
SET operation = 'apply'
WHERE operation IS NULL OR operation = '';

CREATE INDEX IF NOT EXISTS idx_authz_policy_changes_operation
    ON ty_authz_policy_changes(operation)
    WHERE deleted_at IS NULL;
