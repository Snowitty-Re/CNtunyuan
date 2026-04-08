-- ============================================================
-- 10_authz_policy_change_audit.sql
-- 说明：权限系统第四阶段，新增策略变更审计日志
-- ============================================================

CREATE TABLE IF NOT EXISTS ty_authz_policy_changes (
    id CHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,

    operator_id CHAR(36) NULL,
    operator_role VARCHAR(20) NULL,
    operation VARCHAR(20) NOT NULL DEFAULT 'apply',
    change_type VARCHAR(30) NOT NULL,
    target_key VARCHAR(120) NOT NULL,
    rollback_of_id CHAR(36) NULL,
    before_json TEXT NULL,
    after_json TEXT NULL,
    trace_id VARCHAR(100) NULL,

    CONSTRAINT chk_authz_policy_changes_type CHECK (change_type IN ('role_permissions', 'policy_rules')),
    CONSTRAINT chk_authz_policy_changes_operation CHECK (operation IN ('apply', 'rollback'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限策略变更审计日志';

CREATE INDEX idx_authz_policy_changes_created_at ON ty_authz_policy_changes(created_at, deleted_at);
CREATE INDEX idx_authz_policy_changes_type_target ON ty_authz_policy_changes(change_type, target_key, deleted_at);
CREATE INDEX idx_authz_policy_changes_operator ON ty_authz_policy_changes(operator_id, operator_role, deleted_at);
CREATE INDEX idx_authz_policy_changes_rollback_of_id ON ty_authz_policy_changes(rollback_of_id, deleted_at);
