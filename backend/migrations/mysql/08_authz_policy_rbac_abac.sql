-- ============================================================
-- 08_authz_policy_rbac_abac.sql
-- 说明：权限系统第二阶段，新增 RBAC/ABAC 策略表
-- ============================================================

-- 1) 角色权限映射（RBAC）
CREATE TABLE IF NOT EXISTS ty_role_permissions (
    id CHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,

    role VARCHAR(20) NOT NULL,
    permission_code VARCHAR(100) NOT NULL,
    effect VARCHAR(10) NOT NULL DEFAULT 'allow',
    enabled TINYINT(1) NOT NULL DEFAULT 1,
    priority INT NOT NULL DEFAULT 100,

    CONSTRAINT chk_role_permissions_role CHECK (role IN ('super_admin', 'admin', 'manager', 'volunteer')),
    CONSTRAINT chk_role_permissions_effect CHECK (effect IN ('allow', 'deny'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色权限映射表';

CREATE UNIQUE INDEX uk_role_permission_active ON ty_role_permissions(role, permission_code, deleted_at);
CREATE INDEX idx_role_permissions_role_enabled ON ty_role_permissions(role, enabled, deleted_at);

-- 2) 策略规则（ABAC）
CREATE TABLE IF NOT EXISTS ty_policy_rules (
    id CHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,

    permission_code VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    scope_rule VARCHAR(50) NOT NULL,
    condition_json JSON NULL,
    effect VARCHAR(10) NOT NULL DEFAULT 'allow',
    priority INT NOT NULL DEFAULT 100,
    enabled TINYINT(1) NOT NULL DEFAULT 1,

    CONSTRAINT chk_policy_rules_effect CHECK (effect IN ('allow', 'deny'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='ABAC策略规则表';

CREATE INDEX idx_policy_rules_permission ON ty_policy_rules(permission_code, deleted_at);
CREATE INDEX idx_policy_rules_resource ON ty_policy_rules(resource_type, deleted_at);

-- 3) RBAC 种子（幂等）
INSERT IGNORE INTO ty_role_permissions (id, role, permission_code, effect, enabled, priority) VALUES
    -- super_admin
    (UUID(), 'super_admin', 'user:create', 'allow', 1, 100),
    (UUID(), 'super_admin', 'user:modify', 'allow', 1, 100),
    (UUID(), 'super_admin', 'user:view', 'allow', 1, 100),
    (UUID(), 'super_admin', 'task:manage', 'allow', 1, 100),
    (UUID(), 'super_admin', 'task:view', 'allow', 1, 100),
    (UUID(), 'super_admin', 'task:edit', 'allow', 1, 100),
    (UUID(), 'super_admin', 'task:execute', 'allow', 1, 100),
    (UUID(), 'super_admin', 'missing:modify', 'allow', 1, 100),
    (UUID(), 'super_admin', 'missing:manage', 'allow', 1, 100),
    (UUID(), 'super_admin', 'dialect:modify', 'allow', 1, 100),
    (UUID(), 'super_admin', 'dialect:manage', 'allow', 1, 100),

    -- admin
    (UUID(), 'admin', 'user:create', 'allow', 1, 100),
    (UUID(), 'admin', 'user:modify', 'allow', 1, 100),
    (UUID(), 'admin', 'user:view', 'allow', 1, 100),
    (UUID(), 'admin', 'task:manage', 'allow', 1, 100),
    (UUID(), 'admin', 'task:view', 'allow', 1, 100),
    (UUID(), 'admin', 'task:edit', 'allow', 1, 100),
    (UUID(), 'admin', 'task:execute', 'allow', 1, 100),
    (UUID(), 'admin', 'missing:modify', 'allow', 1, 100),
    (UUID(), 'admin', 'missing:manage', 'allow', 1, 100),
    (UUID(), 'admin', 'dialect:modify', 'allow', 1, 100),
    (UUID(), 'admin', 'dialect:manage', 'allow', 1, 100),

    -- manager
    (UUID(), 'manager', 'user:view', 'allow', 1, 100),
    (UUID(), 'manager', 'user:modify', 'allow', 1, 100),
    (UUID(), 'manager', 'task:manage', 'allow', 1, 100),
    (UUID(), 'manager', 'task:view', 'allow', 1, 100),
    (UUID(), 'manager', 'task:edit', 'allow', 1, 100),
    (UUID(), 'manager', 'task:execute', 'allow', 1, 100),
    (UUID(), 'manager', 'missing:modify', 'allow', 1, 100),
    (UUID(), 'manager', 'missing:manage', 'allow', 1, 100),
    (UUID(), 'manager', 'dialect:modify', 'allow', 1, 100),
    (UUID(), 'manager', 'dialect:manage', 'allow', 1, 100),

    -- volunteer
    (UUID(), 'volunteer', 'user:view', 'allow', 1, 100),
    (UUID(), 'volunteer', 'user:modify', 'allow', 1, 100),
    (UUID(), 'volunteer', 'task:view', 'allow', 1, 100),
    (UUID(), 'volunteer', 'task:edit', 'allow', 1, 100),
    (UUID(), 'volunteer', 'task:execute', 'allow', 1, 100),
    (UUID(), 'volunteer', 'missing:modify', 'allow', 1, 100),
    (UUID(), 'volunteer', 'dialect:modify', 'allow', 1, 100);

-- 4) ABAC 默认规则（幂等）
INSERT IGNORE INTO ty_policy_rules (id, permission_code, resource_type, scope_rule, effect, priority, enabled) VALUES
    (UUID(), 'user:create', 'user', 'ORG_DESCENDANT', 'allow', 100, 1),
    (UUID(), 'user:modify', 'user', 'SELF', 'allow', 100, 1),
    (UUID(), 'user:modify', 'user', 'ORG_DESCENDANT', 'allow', 110, 1),
    (UUID(), 'user:view', 'user', 'SELF', 'allow', 100, 1),
    (UUID(), 'user:view', 'user', 'ORG_DESCENDANT', 'allow', 110, 1),

    (UUID(), 'task:manage', 'task', 'ORG_DESCENDANT', 'allow', 100, 1),
    (UUID(), 'task:view', 'task', 'CREATOR', 'allow', 100, 1),
    (UUID(), 'task:view', 'task', 'ASSIGNEE', 'allow', 110, 1),
    (UUID(), 'task:view', 'task', 'ORG_DESCENDANT', 'allow', 120, 1),
    (UUID(), 'task:edit', 'task', 'CREATOR', 'allow', 100, 1),
    (UUID(), 'task:edit', 'task', 'ASSIGNEE', 'allow', 110, 1),
    (UUID(), 'task:edit', 'task', 'ORG_DESCENDANT', 'allow', 120, 1),
    (UUID(), 'task:execute', 'task', 'ASSIGNEE', 'allow', 100, 1),
    (UUID(), 'task:execute', 'task', 'ORG_DESCENDANT', 'allow', 110, 1),

    (UUID(), 'missing:modify', 'missing_person', 'REPORTER', 'allow', 100, 1),
    (UUID(), 'missing:modify', 'missing_person', 'ORG_DESCENDANT', 'allow', 110, 1),
    (UUID(), 'missing:manage', 'missing_person', 'ORG_DESCENDANT', 'allow', 100, 1),

    (UUID(), 'dialect:modify', 'dialect', 'OWNER', 'allow', 100, 1),
    (UUID(), 'dialect:modify', 'dialect', 'ORG_DESCENDANT', 'allow', 110, 1),
    (UUID(), 'dialect:manage', 'dialect', 'ORG_DESCENDANT', 'allow', 100, 1);
