-- ============================================================
-- 13_authz_policy_change_request_workflow.sql
-- 说明：权限系统第七阶段，新增策略变更审批申请表
-- ============================================================

CREATE TABLE IF NOT EXISTS ty_authz_policy_change_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    request_type VARCHAR(30) NOT NULL CHECK (request_type IN ('role_permissions', 'policy_rules', 'rollback')),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    target_key VARCHAR(120) NOT NULL,
    payload_json TEXT,
    preview_json TEXT,
    request_note TEXT,
    requested_by UUID,
    requested_by_role VARCHAR(20),
    review_note TEXT,
    reviewed_by UUID,
    reviewed_at TIMESTAMP WITH TIME ZONE,
    executed BOOLEAN NOT NULL DEFAULT FALSE,
    executed_at TIMESTAMP WITH TIME ZONE,
    executed_log_id UUID,
    trace_id VARCHAR(100)
);

CREATE INDEX IF NOT EXISTS idx_authz_policy_change_requests_type_status
    ON ty_authz_policy_change_requests(request_type, status)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_authz_policy_change_requests_requested_by
    ON ty_authz_policy_change_requests(requested_by, requested_by_role)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_authz_policy_change_requests_target_key
    ON ty_authz_policy_change_requests(target_key)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_authz_policy_change_requests_created_at
    ON ty_authz_policy_change_requests(created_at DESC)
    WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS update_authz_policy_change_requests_updated_at ON ty_authz_policy_change_requests;
CREATE TRIGGER update_authz_policy_change_requests_updated_at BEFORE UPDATE ON ty_authz_policy_change_requests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
