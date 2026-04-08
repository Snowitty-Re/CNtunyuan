-- ============================================================
-- 09_authz_decision_audit.sql
-- 说明：权限系统第三阶段，新增授权决策审计日志
-- ============================================================

CREATE TABLE IF NOT EXISTS ty_authz_decisions (
    id CHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,

    operator_id CHAR(36) NULL,
    operator_role VARCHAR(20) NULL,
    operator_org_id CHAR(36) NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id VARCHAR(100) NULL,
    allowed TINYINT(1) NOT NULL,
    reason VARCHAR(50) NOT NULL,
    trace_id VARCHAR(100) NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='授权决策审计日志';

CREATE INDEX idx_authz_decisions_created_at ON ty_authz_decisions(created_at, deleted_at);
CREATE INDEX idx_authz_decisions_action_allowed ON ty_authz_decisions(action, allowed, deleted_at);
CREATE INDEX idx_authz_decisions_operator ON ty_authz_decisions(operator_id, operator_role, deleted_at);
CREATE INDEX idx_authz_decisions_resource ON ty_authz_decisions(resource_type, resource_id, deleted_at);
CREATE INDEX idx_authz_decisions_reason ON ty_authz_decisions(reason, deleted_at);
