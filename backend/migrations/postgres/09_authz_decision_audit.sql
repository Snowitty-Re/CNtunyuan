-- ============================================================
-- 09_authz_decision_audit.sql
-- 说明：权限系统第三阶段，新增授权决策审计日志
-- ============================================================

CREATE TABLE IF NOT EXISTS ty_authz_decisions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    operator_id UUID,
    operator_role VARCHAR(20),
    operator_org_id UUID,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id VARCHAR(100),
    allowed BOOLEAN NOT NULL,
    reason VARCHAR(50) NOT NULL,
    trace_id VARCHAR(100)
);

CREATE INDEX IF NOT EXISTS idx_authz_decisions_created_at
    ON ty_authz_decisions(created_at DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_authz_decisions_action_allowed
    ON ty_authz_decisions(action, allowed)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_authz_decisions_operator
    ON ty_authz_decisions(operator_id, operator_role)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_authz_decisions_resource
    ON ty_authz_decisions(resource_type, resource_id)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_authz_decisions_reason
    ON ty_authz_decisions(reason)
    WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS update_authz_decisions_updated_at ON ty_authz_decisions;
CREATE TRIGGER update_authz_decisions_updated_at BEFORE UPDATE ON ty_authz_decisions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
