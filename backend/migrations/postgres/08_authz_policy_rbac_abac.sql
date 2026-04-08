-- ============================================================
-- 08_authz_policy_rbac_abac.sql
-- 说明：权限系统第二阶段，新增 RBAC/ABAC 策略表
-- ============================================================

-- 1) 角色权限映射（RBAC）
CREATE TABLE IF NOT EXISTS ty_role_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    role VARCHAR(20) NOT NULL CHECK (role IN ('super_admin', 'admin', 'manager', 'volunteer')),
    permission_code VARCHAR(100) NOT NULL,
    effect VARCHAR(10) NOT NULL DEFAULT 'allow' CHECK (effect IN ('allow', 'deny')),
    enabled BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER NOT NULL DEFAULT 100
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_role_permission_active
    ON ty_role_permissions(role, permission_code)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_enabled
    ON ty_role_permissions(role, enabled)
    WHERE deleted_at IS NULL;

-- 2) 策略规则（ABAC）
CREATE TABLE IF NOT EXISTS ty_policy_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    permission_code VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    scope_rule VARCHAR(50) NOT NULL,
    condition_json JSONB,
    effect VARCHAR(10) NOT NULL DEFAULT 'allow' CHECK (effect IN ('allow', 'deny')),
    priority INTEGER NOT NULL DEFAULT 100,
    enabled BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX IF NOT EXISTS idx_policy_rules_permission
    ON ty_policy_rules(permission_code)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_policy_rules_resource
    ON ty_policy_rules(resource_type)
    WHERE deleted_at IS NULL;
DROP TRIGGER IF EXISTS update_role_permissions_updated_at ON ty_role_permissions;
CREATE TRIGGER update_role_permissions_updated_at BEFORE UPDATE ON ty_role_permissions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_policy_rules_updated_at ON ty_policy_rules;
CREATE TRIGGER update_policy_rules_updated_at BEFORE UPDATE ON ty_policy_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 3) RBAC 种子（幂等）
INSERT INTO ty_role_permissions (id, role, permission_code, effect, enabled, priority)
VALUES
    -- super_admin
    (uuid_generate_v4(), 'super_admin', 'user:create', 'allow', true, 100),
    (uuid_generate_v4(), 'super_admin', 'user:modify', 'allow', true, 100),
    (uuid_generate_v4(), 'super_admin', 'user:view', 'allow', true, 100),
    (uuid_generate_v4(), 'super_admin', 'task:manage', 'allow', true, 100),
    (uuid_generate_v4(), 'super_admin', 'task:view', 'allow', true, 100),
    (uuid_generate_v4(), 'super_admin', 'task:edit', 'allow', true, 100),
    (uuid_generate_v4(), 'super_admin', 'task:execute', 'allow', true, 100),
    (uuid_generate_v4(), 'super_admin', 'missing:modify', 'allow', true, 100),
    (uuid_generate_v4(), 'super_admin', 'missing:manage', 'allow', true, 100),
    (uuid_generate_v4(), 'super_admin', 'dialect:modify', 'allow', true, 100),
    (uuid_generate_v4(), 'super_admin', 'dialect:manage', 'allow', true, 100),

    -- admin
    (uuid_generate_v4(), 'admin', 'user:create', 'allow', true, 100),
    (uuid_generate_v4(), 'admin', 'user:modify', 'allow', true, 100),
    (uuid_generate_v4(), 'admin', 'user:view', 'allow', true, 100),
    (uuid_generate_v4(), 'admin', 'task:manage', 'allow', true, 100),
    (uuid_generate_v4(), 'admin', 'task:view', 'allow', true, 100),
    (uuid_generate_v4(), 'admin', 'task:edit', 'allow', true, 100),
    (uuid_generate_v4(), 'admin', 'task:execute', 'allow', true, 100),
    (uuid_generate_v4(), 'admin', 'missing:modify', 'allow', true, 100),
    (uuid_generate_v4(), 'admin', 'missing:manage', 'allow', true, 100),
    (uuid_generate_v4(), 'admin', 'dialect:modify', 'allow', true, 100),
    (uuid_generate_v4(), 'admin', 'dialect:manage', 'allow', true, 100),

    -- manager
    (uuid_generate_v4(), 'manager', 'user:view', 'allow', true, 100),
    (uuid_generate_v4(), 'manager', 'user:modify', 'allow', true, 100),
    (uuid_generate_v4(), 'manager', 'task:manage', 'allow', true, 100),
    (uuid_generate_v4(), 'manager', 'task:view', 'allow', true, 100),
    (uuid_generate_v4(), 'manager', 'task:edit', 'allow', true, 100),
    (uuid_generate_v4(), 'manager', 'task:execute', 'allow', true, 100),
    (uuid_generate_v4(), 'manager', 'missing:modify', 'allow', true, 100),
    (uuid_generate_v4(), 'manager', 'missing:manage', 'allow', true, 100),
    (uuid_generate_v4(), 'manager', 'dialect:modify', 'allow', true, 100),
    (uuid_generate_v4(), 'manager', 'dialect:manage', 'allow', true, 100),

    -- volunteer
    (uuid_generate_v4(), 'volunteer', 'user:view', 'allow', true, 100),
    (uuid_generate_v4(), 'volunteer', 'user:modify', 'allow', true, 100),
    (uuid_generate_v4(), 'volunteer', 'task:view', 'allow', true, 100),
    (uuid_generate_v4(), 'volunteer', 'task:edit', 'allow', true, 100),
    (uuid_generate_v4(), 'volunteer', 'task:execute', 'allow', true, 100),
    (uuid_generate_v4(), 'volunteer', 'missing:modify', 'allow', true, 100),
    (uuid_generate_v4(), 'volunteer', 'dialect:modify', 'allow', true, 100)
ON CONFLICT DO NOTHING;

-- 4) ABAC 默认规则（幂等）
INSERT INTO ty_policy_rules (id, permission_code, resource_type, scope_rule, effect, priority, enabled)
VALUES
    (uuid_generate_v4(), 'user:create', 'user', 'ORG_DESCENDANT', 'allow', 100, true),
    (uuid_generate_v4(), 'user:modify', 'user', 'SELF', 'allow', 100, true),
    (uuid_generate_v4(), 'user:modify', 'user', 'ORG_DESCENDANT', 'allow', 110, true),
    (uuid_generate_v4(), 'user:view', 'user', 'SELF', 'allow', 100, true),
    (uuid_generate_v4(), 'user:view', 'user', 'ORG_DESCENDANT', 'allow', 110, true),

    (uuid_generate_v4(), 'task:manage', 'task', 'ORG_DESCENDANT', 'allow', 100, true),
    (uuid_generate_v4(), 'task:view', 'task', 'CREATOR', 'allow', 100, true),
    (uuid_generate_v4(), 'task:view', 'task', 'ASSIGNEE', 'allow', 110, true),
    (uuid_generate_v4(), 'task:view', 'task', 'ORG_DESCENDANT', 'allow', 120, true),
    (uuid_generate_v4(), 'task:edit', 'task', 'CREATOR', 'allow', 100, true),
    (uuid_generate_v4(), 'task:edit', 'task', 'ASSIGNEE', 'allow', 110, true),
    (uuid_generate_v4(), 'task:edit', 'task', 'ORG_DESCENDANT', 'allow', 120, true),
    (uuid_generate_v4(), 'task:execute', 'task', 'ASSIGNEE', 'allow', 100, true),
    (uuid_generate_v4(), 'task:execute', 'task', 'ORG_DESCENDANT', 'allow', 110, true),

    (uuid_generate_v4(), 'missing:modify', 'missing_person', 'REPORTER', 'allow', 100, true),
    (uuid_generate_v4(), 'missing:modify', 'missing_person', 'ORG_DESCENDANT', 'allow', 110, true),
    (uuid_generate_v4(), 'missing:manage', 'missing_person', 'ORG_DESCENDANT', 'allow', 100, true),

    (uuid_generate_v4(), 'dialect:modify', 'dialect', 'OWNER', 'allow', 100, true),
    (uuid_generate_v4(), 'dialect:modify', 'dialect', 'ORG_DESCENDANT', 'allow', 110, true),
    (uuid_generate_v4(), 'dialect:manage', 'dialect', 'ORG_DESCENDANT', 'allow', 100, true)
ON CONFLICT DO NOTHING;
