-- Migration: RBAC Permission System Enhancement
-- Date: 2026-03-07
-- Description: Add comprehensive permission system with roles, role-permissions, user-roles

-- ============================================
-- 1. Roles Table
-- ============================================
CREATE TABLE IF NOT EXISTS ty_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL,
    description VARCHAR(255),
    type VARCHAR(20) NOT NULL DEFAULT 'custom',
    org_id UUID NOT NULL REFERENCES ty_organizations(id) ON DELETE CASCADE,
    data_scope VARCHAR(20) NOT NULL DEFAULT 'org_only',
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by UUID,
    updated_by UUID,
    
    UNIQUE(code, org_id)
);

-- Indexes for roles
CREATE INDEX IF NOT EXISTS idx_roles_org_id ON ty_roles(org_id);
CREATE INDEX IF NOT EXISTS idx_roles_type ON ty_roles(type);
CREATE INDEX IF NOT EXISTS idx_roles_status ON ty_roles(status);
CREATE INDEX IF NOT EXISTS idx_roles_deleted_at ON ty_roles(deleted_at) WHERE deleted_at IS NULL;

-- ============================================
-- 2. Permissions Table (Enhanced)
-- ============================================
-- Note: ty_permissions table already exists from initial migration
-- Adding new columns if they don't exist

DO $$
BEGIN
    -- Add resource column if not exists
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'ty_permissions' AND column_name = 'resource') THEN
        ALTER TABLE ty_permissions ADD COLUMN resource VARCHAR(50);
    END IF;

    -- Add action column if not exists
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'ty_permissions' AND column_name = 'action') THEN
        ALTER TABLE ty_permissions ADD COLUMN action VARCHAR(50);
    END IF;

    -- Add category column if not exists
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'ty_permissions' AND column_name = 'category') THEN
        ALTER TABLE ty_permissions ADD COLUMN category VARCHAR(50);
    END IF;

    -- Add is_system column if not exists
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'ty_permissions' AND column_name = 'is_system') THEN
        ALTER TABLE ty_permissions ADD COLUMN is_system BOOLEAN DEFAULT FALSE;
    END IF;

    -- Add sort_order column if not exists
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'ty_permissions' AND column_name = 'sort_order') THEN
        ALTER TABLE ty_permissions ADD COLUMN sort_order INTEGER DEFAULT 0;
    END IF;

    -- Add parent_id column if not exists
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'ty_permissions' AND column_name = 'parent_id') THEN
        ALTER TABLE ty_permissions ADD COLUMN parent_id UUID;
    END IF;
END $$;

-- Add index for new columns
CREATE INDEX IF NOT EXISTS idx_permissions_resource ON ty_permissions(resource);
CREATE INDEX IF NOT EXISTS idx_permissions_action ON ty_permissions(action);
CREATE INDEX IF NOT EXISTS idx_permissions_category ON ty_permissions(category);

-- ============================================
-- 3. Role Permissions Junction Table
-- ============================================
CREATE TABLE IF NOT EXISTS ty_role_permissions (
    role_id UUID NOT NULL REFERENCES ty_roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES ty_permissions(id) ON DELETE CASCADE,
    granted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    granted_by UUID,
    
    PRIMARY KEY (role_id, permission_id)
);

-- Indexes for role_permissions
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON ty_role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON ty_role_permissions(permission_id);

-- ============================================
-- 4. User Roles Junction Table
-- ============================================
CREATE TABLE IF NOT EXISTS ty_user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES ty_users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES ty_roles(id) ON DELETE CASCADE,
    org_id UUID NOT NULL REFERENCES ty_organizations(id) ON DELETE CASCADE,
    assigned_by UUID,
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    UNIQUE(user_id, role_id, org_id)
);

-- Indexes for user_roles
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON ty_user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON ty_user_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_org_id ON ty_user_roles(org_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_expires_at ON ty_user_roles(expires_at);
CREATE INDEX IF NOT EXISTS idx_user_roles_deleted_at ON ty_user_roles(deleted_at) WHERE deleted_at IS NULL;

-- ============================================
-- 5. Field Permissions Table
-- ============================================
CREATE TABLE IF NOT EXISTS ty_field_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource VARCHAR(50) NOT NULL,
    field VARCHAR(50) NOT NULL,
    role_id UUID NOT NULL REFERENCES ty_roles(id) ON DELETE CASCADE,
    permission VARCHAR(20) NOT NULL DEFAULT 'hidden', -- hidden, readonly, writeable
    condition JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    UNIQUE(resource, field, role_id)
);

-- Indexes for field_permissions
CREATE INDEX IF NOT EXISTS idx_field_permissions_resource ON ty_field_permissions(resource);
CREATE INDEX IF NOT EXISTS idx_field_permissions_role_id ON ty_field_permissions(role_id);

-- ============================================
-- 6. Data Permission Rules Table
-- ============================================
CREATE TABLE IF NOT EXISTS ty_data_permission_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description VARCHAR(255),
    resource VARCHAR(50) NOT NULL,
    role_id UUID REFERENCES ty_roles(id) ON DELETE CASCADE,
    org_id UUID REFERENCES ty_organizations(id) ON DELETE CASCADE,
    rule_type VARCHAR(20) NOT NULL, -- field_filter, org_scope, custom
    rule_config JSONB NOT NULL,
    priority INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for data_permission_rules
CREATE INDEX IF NOT EXISTS idx_data_perm_rules_resource ON ty_data_permission_rules(resource);
CREATE INDEX IF NOT EXISTS idx_data_perm_rules_role_id ON ty_data_permission_rules(role_id);
CREATE INDEX IF NOT EXISTS idx_data_perm_rules_org_id ON ty_data_permission_rules(org_id);
CREATE INDEX IF NOT EXISTS idx_data_perm_rules_active ON ty_data_permission_rules(is_active);

-- ============================================
-- 7. Insert Default System Roles
-- ============================================

-- Note: Using fixed UUID for system-wide roles
INSERT INTO ty_roles (id, name, code, description, type, org_id, data_scope, is_system, status, sort_order)
VALUES 
    ('00000000-0000-0000-0000-000000000001', '超级管理员', 'super_admin', '系统超级管理员，拥有所有权限', 'system', '00000000-0000-0000-0000-000000000000', 'all', TRUE, 'active', 0),
    ('00000000-0000-0000-0000-000000000002', '管理员', 'admin', '组织管理员', 'system', '00000000-0000-0000-0000-000000000000', 'org_and_sub', TRUE, 'active', 1),
    ('00000000-0000-0000-0000-000000000003', '经理', 'manager', '部门经理', 'system', '00000000-0000-0000-0000-000000000000', 'org_only', TRUE, 'active', 2),
    ('00000000-0000-0000-0000-000000000004', '志愿者', 'volunteer', '普通志愿者', 'system', '00000000-0000-0000-0000-000000000000', 'self', TRUE, 'active', 3)
ON CONFLICT (code, org_id) DO NOTHING;

-- ============================================
-- 8. Insert Default Permissions
-- ============================================

-- User Management Permissions
INSERT INTO ty_permissions (id, name, code, description, resource, action, category, is_system, sort_order)
VALUES 
    (gen_random_uuid(), '用户创建', 'user:create', '创建用户', 'user', 'create', 'user', TRUE, 1),
    (gen_random_uuid(), '用户查看', 'user:read', '查看用户详情', 'user', 'read', 'user', TRUE, 2),
    (gen_random_uuid(), '用户更新', 'user:update', '更新用户信息', 'user', 'update', 'user', TRUE, 3),
    (gen_random_uuid(), '用户删除', 'user:delete', '删除用户', 'user', 'delete', 'user', TRUE, 4),
    (gen_random_uuid(), '用户列表', 'user:list', '查看用户列表', 'user', 'list', 'user', TRUE, 5)
ON CONFLICT (code) DO NOTHING;

-- Organization Management Permissions
INSERT INTO ty_permissions (id, name, code, description, resource, action, category, is_system, sort_order)
VALUES 
    (gen_random_uuid(), '组织创建', 'organization:create', '创建组织', 'organization', 'create', 'organization', TRUE, 10),
    (gen_random_uuid(), '组织查看', 'organization:read', '查看组织详情', 'organization', 'read', 'organization', TRUE, 11),
    (gen_random_uuid(), '组织更新', 'organization:update', '更新组织信息', 'organization', 'update', 'organization', TRUE, 12),
    (gen_random_uuid(), '组织删除', 'organization:delete', '删除组织', 'organization', 'delete', 'organization', TRUE, 13),
    (gen_random_uuid(), '组织列表', 'organization:list', '查看组织列表', 'organization', 'list', 'organization', TRUE, 14)
ON CONFLICT (code) DO NOTHING;

-- Task Management Permissions
INSERT INTO ty_permissions (id, name, code, description, resource, action, category, is_system, sort_order)
VALUES 
    (gen_random_uuid(), '任务创建', 'task:create', '创建任务', 'task', 'create', 'task', TRUE, 20),
    (gen_random_uuid(), '任务查看', 'task:read', '查看任务详情', 'task', 'read', 'task', TRUE, 21),
    (gen_random_uuid(), '任务更新', 'task:update', '更新任务信息', 'task', 'update', 'task', TRUE, 22),
    (gen_random_uuid(), '任务删除', 'task:delete', '删除任务', 'task', 'delete', 'task', TRUE, 23),
    (gen_random_uuid(), '任务列表', 'task:list', '查看任务列表', 'task', 'list', 'task', TRUE, 24),
    (gen_random_uuid(), '任务分配', 'task:assign', '分配任务', 'task', 'assign', 'task', TRUE, 25)
ON CONFLICT (code) DO NOTHING;

-- Missing Person Management Permissions
INSERT INTO ty_permissions (id, name, code, description, resource, action, category, is_system, sort_order)
VALUES 
    (gen_random_uuid(), '寻人启事创建', 'missing_person:create', '创建寻人启事', 'missing_person', 'create', 'missing_person', TRUE, 30),
    (gen_random_uuid(), '寻人启事查看', 'missing_person:read', '查看寻人启事', 'missing_person', 'read', 'missing_person', TRUE, 31),
    (gen_random_uuid(), '寻人启事更新', 'missing_person:update', '更新寻人启事', 'missing_person', 'update', 'missing_person', TRUE, 32),
    (gen_random_uuid(), '寻人启事删除', 'missing_person:delete', '删除寻人启事', 'missing_person', 'delete', 'missing_person', TRUE, 33),
    (gen_random_uuid(), '寻人启事列表', 'missing_person:list', '查看寻人启事列表', 'missing_person', 'list', 'missing_person', TRUE, 34)
ON CONFLICT (code) DO NOTHING;

-- Workflow Permissions
INSERT INTO ty_permissions (id, name, code, description, resource, action, category, is_system, sort_order)
VALUES 
    (gen_random_uuid(), '流程定义创建', 'workflow:create', '创建流程定义', 'workflow', 'create', 'workflow', TRUE, 40),
    (gen_random_uuid(), '流程定义查看', 'workflow:read', '查看流程定义', 'workflow', 'read', 'workflow', TRUE, 41),
    (gen_random_uuid(), '流程定义更新', 'workflow:update', '更新流程定义', 'workflow', 'update', 'workflow', TRUE, 42),
    (gen_random_uuid(), '流程定义删除', 'workflow:delete', '删除流程定义', 'workflow', 'delete', 'workflow', TRUE, 43),
    (gen_random_uuid(), '流程审批', 'workflow:approve', '审批流程实例', 'workflow', 'approve', 'workflow', TRUE, 44),
    (gen_random_uuid(), '流程委托', 'workflow:delegate', '委托流程任务', 'workflow', 'delegate', 'workflow', TRUE, 45)
ON CONFLICT (code) DO NOTHING;

-- System Management Permissions
INSERT INTO ty_permissions (id, name, code, description, resource, action, category, is_system, sort_order)
VALUES 
    (gen_random_uuid(), '角色管理', 'system:role_manage', '管理角色', 'system', 'manage', 'system', TRUE, 50),
    (gen_random_uuid(), '权限管理', 'system:permission_manage', '管理权限', 'system', 'permission', 'system', TRUE, 51),
    (gen_random_uuid(), '系统设置', 'system:settings', '系统设置', 'system', 'settings', 'system', TRUE, 52),
    (gen_random_uuid(), '审计日志查看', 'system:audit_read', '查看审计日志', 'system', 'audit', 'system', TRUE, 53)
ON CONFLICT (code) DO NOTHING;

-- ============================================
-- 9. Grant All Permissions to Super Admin
-- ============================================

INSERT INTO ty_role_permissions (role_id, permission_id, granted_at)
SELECT '00000000-0000-0000-0000-000000000001', id, NOW()
FROM ty_permissions
ON CONFLICT DO NOTHING;

-- ============================================
-- 10. Grant User/Org/Task Permissions to Admin
-- ============================================

INSERT INTO ty_role_permissions (role_id, permission_id, granted_at)
SELECT '00000000-0000-0000-0000-000000000002', id, NOW()
FROM ty_permissions
WHERE category IN ('user', 'organization', 'task', 'missing_person')
ON CONFLICT DO NOTHING;

-- ============================================
-- 11. Grant Task/Workflow Read to Manager
-- ============================================

INSERT INTO ty_role_permissions (role_id, permission_id, granted_at)
SELECT '00000000-0000-0000-0000-000000000003', id, NOW()
FROM ty_permissions
WHERE (category IN ('task', 'workflow') AND action IN ('read', 'list'))
   OR (category = 'missing_person' AND action IN ('read', 'list', 'create', 'update'))
ON CONFLICT DO NOTHING;

-- ============================================
-- 12. Grant Basic Read to Volunteer
-- ============================================

INSERT INTO ty_role_permissions (role_id, permission_id, granted_at)
SELECT '00000000-0000-0000-0000-000000000004', id, NOW()
FROM ty_permissions
WHERE action IN ('read', 'list') 
  AND category IN ('task', 'missing_person', 'organization')
ON CONFLICT DO NOTHING;

-- ============================================
-- 13. Add Foreign Key Constraint (if user_permissions exists)
-- ============================================

-- Note: If there's an existing user_permissions table, we keep it for backward compatibility
-- The new system uses ty_user_roles and ty_role_permissions instead

-- ============================================
-- 14. Create Updated At Trigger Function
-- ============================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for new tables
DO $$
BEGIN
    -- ty_roles trigger
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_ty_roles_updated_at') THEN
        CREATE TRIGGER update_ty_roles_updated_at
            BEFORE UPDATE ON ty_roles
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;

    -- ty_user_roles trigger
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_ty_user_roles_updated_at') THEN
        CREATE TRIGGER update_ty_user_roles_updated_at
            BEFORE UPDATE ON ty_user_roles
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;

    -- ty_field_permissions trigger
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_ty_field_permissions_updated_at') THEN
        CREATE TRIGGER update_ty_field_permissions_updated_at
            BEFORE UPDATE ON ty_field_permissions
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;

    -- ty_data_permission_rules trigger
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_ty_data_permission_rules_updated_at') THEN
        CREATE TRIGGER update_ty_data_permission_rules_updated_at
            BEFORE UPDATE ON ty_data_permission_rules
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;

-- ============================================
-- Migration Complete
-- ============================================
