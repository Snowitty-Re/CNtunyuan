-- ============================================================
-- 13_authz_policy_change_request_workflow.sql
-- 说明：权限系统第七阶段，新增策略变更审批申请表
-- ============================================================

CREATE TABLE IF NOT EXISTS ty_authz_policy_change_requests (
    id CHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    request_type VARCHAR(30) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    target_key VARCHAR(120) NOT NULL,
    payload_json TEXT NULL,
    preview_json TEXT NULL,
    request_note TEXT NULL,
    requested_by CHAR(36) NULL,
    requested_by_role VARCHAR(20) NULL,
    review_note TEXT NULL,
    reviewed_by CHAR(36) NULL,
    reviewed_at TIMESTAMP NULL DEFAULT NULL,
    executed TINYINT(1) NOT NULL DEFAULT 0,
    executed_at TIMESTAMP NULL DEFAULT NULL,
    executed_log_id CHAR(36) NULL,
    trace_id VARCHAR(100) NULL,
    CONSTRAINT chk_authz_policy_change_requests_type CHECK (request_type IN ('role_permissions', 'policy_rules', 'rollback')),
    CONSTRAINT chk_authz_policy_change_requests_status CHECK (status IN ('pending', 'approved', 'rejected'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限策略变更审批申请表';

CREATE INDEX idx_authz_policy_change_requests_type_status ON ty_authz_policy_change_requests(request_type, status, deleted_at);
CREATE INDEX idx_authz_policy_change_requests_requested_by ON ty_authz_policy_change_requests(requested_by, requested_by_role, deleted_at);
CREATE INDEX idx_authz_policy_change_requests_target_key ON ty_authz_policy_change_requests(target_key, deleted_at);
CREATE INDEX idx_authz_policy_change_requests_created_at ON ty_authz_policy_change_requests(created_at, deleted_at);
