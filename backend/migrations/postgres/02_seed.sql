-- ============================================================
-- 团圆寻亲志愿者系统 - PostgreSQL 种子数据
-- 版本: 1.0.0
-- 说明: 初始化超级管理员、根组织和基础权限数据
-- ============================================================

-- ============================================================
-- 1. 创建根组织
-- ============================================================
INSERT INTO ty_organizations (
    id, created_at, updated_at, name, code, type, level, parent_id, 
    description, address, contact_name, contact_phone, status, logo, sort_order
) VALUES (
    '00000000-0000-0000-0000-000000000000',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    '团圆寻亲志愿者协会',
    'ROOT',
    'root',
    1,
    NULL,
    '团圆寻亲志愿者系统根组织，负责统筹全国志愿者工作',
    '中国',
    '系统管理员',
    '13800000000',
    'active',
    NULL,
    0
) ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- 2. 创建超级管理员用户
-- 密码: admin123 (bcrypt 加密)
-- ============================================================
INSERT INTO ty_users (
    id, created_at, updated_at, nickname, phone, email, password, 
    role, status, org_id, avatar, real_name, gender, address, introduction
) VALUES (
    '00000000-0000-0000-0000-000000000001',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    '超级管理员',
    '13800138000',
    'admin@cntuanyuan.org',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjXAgwkLzhxDfmAg5r3eRmFfbwkBDeq',
    'super_admin',
    'active',
    '00000000-0000-0000-0000-000000000000',
    NULL,
    '系统管理员',
    'male',
    '中国',
    '团圆寻亲志愿者系统超级管理员'
) ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- 3. 初始化组织统计
-- ============================================================
INSERT INTO ty_org_stats (
    id, created_at, updated_at, org_id, total_volunteers, active_volunteers,
    total_cases, active_cases, completed_cases, total_tasks, pending_tasks
) VALUES (
    uuid_generate_v4(),
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    '00000000-0000-0000-0000-000000000000',
    1,
    1,
    0,
    0,
    0,
    0,
    0
) ON CONFLICT (org_id) DO NOTHING;

-- ============================================================
-- 4. 创建基础权限数据
-- ============================================================

-- 用户管理权限
INSERT INTO ty_permissions (id, created_at, updated_at, name, code, description, resource, action) VALUES
    (uuid_generate_v4(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '查看用户', 'user:view', '查看用户列表和详情', 'user', 'view'),
    (uuid_generate_v4(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '创建用户', 'user:create', '创建新用户', 'user', 'create'),
    (uuid_generate_v4(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '编辑用户', 'user:edit', '编辑用户信息', 'user', 'edit'),
    (uuid_generate_v4(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '删除用户', 'user:delete', '删除用户', 'user', 'delete')
ON CONFLICT (code) DO NOTHING;

-- 组织管理权限
INSERT INTO ty_permissions (id, created_at, updated_at, name, code, description, resource, action) VALUES
    (uuid_generate_v4(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '查看组织', 'org:view', '查看组织列表和详情', 'organization', 'view'),
    (uuid_generate_v4(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '创建组织', 'org:create', '创建新组织', 'organization', 'create'),
    (uuid_generate_v4(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '编辑组织', 'org:edit', '编辑组织信息', 'organization', 'edit'),
    (uuid_generate_v4(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '删除组织', 'org:delete', '删除组织', 'organization', 'delete')
ON CONFLICT (code) DO NOTHING;

-- 走失人员管理权限
INSERT INTO ty_permissions (id, created_at, updated_at, name, code, description, resource, action) VALUES
    (uuid_generate_v4(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '查看走失人员', 'missing:view', '查看走失人员列表和详情', 'missing_person', 'view'),
    (uuid_generate_v4(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '创建走失人员', 'missing:create', '登记走失人员', 'missing_person', 'create'),
    (uuid_generate_v4(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '编辑走失人员', 'missing:edit', '编辑走失人员信息', 'missing_person', 'edit'),
    (uuid_generate_v4(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '删除走失人员', 'missing:delete', '删除走失人员记录', 'missing_person', 'delete')
ON CONFLICT (code) DO NOTHING;

-- 任务管理权限
INSERT INTO ty_permissions (id, created_at, updated_at, name, code, description, resource, action) VALUES
    (uuid_generate_v4(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '查看任务', 'task:view', '查看任务列表和详情', 'task', 'view'),
    (uuid_generate_v4(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '创建任务', 'task:create', '创建新任务', 'task', 'create'),
    (uuid_generate_v4(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '编辑任务', 'task:edit', '编辑任务信息', 'task', 'edit'),
    (uuid_generate_v4(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '删除任务', 'task:delete', '删除任务', 'task', 'delete'),
    (uuid_generate_v4(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '分配任务', 'task:assign', '分配任务给志愿者', 'task', 'assign')
ON CONFLICT (code) DO NOTHING;

-- 方言管理权限
INSERT INTO ty_permissions (id, created_at, updated_at, name, code, description, resource, action) VALUES
    (uuid_generate_v4(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '查看方言', 'dialect:view', '查看方言列表和详情', 'dialect', 'view'),
    (uuid_generate_v4(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '上传方言', 'dialect:upload', '上传方言语音', 'dialect', 'upload'),
    (uuid_generate_v4(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '审核方言', 'dialect:review', '审核方言内容', 'dialect', 'review'),
    (uuid_generate_v4(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '删除方言', 'dialect:delete', '删除方言记录', 'dialect', 'delete')
ON CONFLICT (code) DO NOTHING;

-- 系统管理权限
INSERT INTO ty_permissions (id, created_at, updated_at, name, code, description, resource, action) VALUES
    (uuid_generate_v4(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '系统设置', 'system:config', '管理系统配置', 'system', 'config'),
    (uuid_generate_v4(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '查看日志', 'system:log', '查看系统日志', 'system', 'log'),
    (uuid_generate_v4(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '数据统计', 'system:stats', '查看数据统计', 'system', 'stats')
ON CONFLICT (code) DO NOTHING;

-- ============================================================
-- 5. 为超级管理员分配所有权限
-- ============================================================
INSERT INTO ty_user_permissions (user_id, permission_id, granted_at, granted_by)
SELECT 
    '00000000-0000-0000-0000-000000000001',
    id,
    CURRENT_TIMESTAMP,
    '00000000-0000-0000-0000-000000000001'
FROM ty_permissions
ON CONFLICT (user_id, permission_id) DO NOTHING;

-- ============================================================
-- 完成
-- ============================================================
