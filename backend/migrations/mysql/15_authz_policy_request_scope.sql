-- 15_authz_policy_request_scope.sql
-- 权限策略申请：组织作用域审批

ALTER TABLE ty_authz_policy_change_requests
    ADD COLUMN IF NOT EXISTS scope_type VARCHAR(20) NOT NULL DEFAULT 'global' AFTER status,
    ADD COLUMN IF NOT EXISTS target_org_id CHAR(36) NULL AFTER scope_type;

-- MySQL 8.0.16+ supports CHECK
ALTER TABLE ty_authz_policy_change_requests
    ADD CONSTRAINT chk_authz_policy_change_requests_scope_type
    CHECK (scope_type IN ('global', 'org'));

CREATE INDEX idx_authz_policy_change_requests_scope_type
    ON ty_authz_policy_change_requests(scope_type, deleted_at);

CREATE INDEX idx_authz_policy_change_requests_target_org_id
    ON ty_authz_policy_change_requests(target_org_id, deleted_at);
