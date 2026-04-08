-- ============================================================
-- 10_authz_policy_change_audit.sql
-- 说明：权限系统第四阶段，新增策略变更审计日志
-- ============================================================

CREATE TABLE IF NOT EXISTS ty_authz_policy_changes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    operator_id UUID,
    operator_role VARCHAR(20),
    operation VARCHAR(20) NOT NULL DEFAULT 'apply' CHECK (operation IN ('apply', 'rollback')),
    change_type VARCHAR(30) NOT NULL CHECK (change_type IN ('role_permissions', 'policy_rules')),
    target_key VARCHAR(120) NOT NULL,
    rollback_of_id UUID,
    before_json TEXT,
    after_json TEXT,
    trace_id VARCHAR(100)
);

CREATE INDEX IF NOT EXISTS idx_authz_policy_changes_created_at
    ON ty_authz_policy_changes(created_at DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_authz_policy_changes_type_target
    ON ty_authz_policy_changes(change_type, target_key)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_authz_policy_changes_operator
    ON ty_authz_policy_changes(operator_id, operator_role)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_authz_policy_changes_rollback_of_id
    ON ty_authz_policy_changes(rollback_of_id)
    WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS update_authz_policy_changes_updated_at ON ty_authz_policy_changes;
CREATE TRIGGER update_authz_policy_changes_updated_at BEFORE UPDATE ON ty_authz_policy_changes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
