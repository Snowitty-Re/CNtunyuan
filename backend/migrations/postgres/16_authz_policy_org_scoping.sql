-- 16_authz_policy_org_scoping.sql
-- 权限策略按组织作用域生效（org_id 为空表示全局）

ALTER TABLE ty_role_permissions
    ADD COLUMN IF NOT EXISTS org_id UUID;

ALTER TABLE ty_policy_rules
    ADD COLUMN IF NOT EXISTS org_id UUID;

DROP INDEX IF EXISTS uk_role_permission_active;
CREATE UNIQUE INDEX IF NOT EXISTS uk_role_permission_active
    ON ty_role_permissions(role, permission_code, org_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_role_permissions_org_enabled
    ON ty_role_permissions(org_id, enabled)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_policy_rules_permission_org
    ON ty_policy_rules(permission_code, org_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_policy_rules_org_priority
    ON ty_policy_rules(org_id, priority)
    WHERE deleted_at IS NULL;
