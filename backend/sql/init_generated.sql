-- 团圆寻亲志愿者系统 - 数据库初始化脚本
-- 生成时间: 2026-02-27 21:26:54
-- 超级管理员: 13800138000
-- 默认密码: admin123

BEGIN;

-- ============================================
-- 1. 根组织 (团圆志愿者总部)
-- ============================================
INSERT INTO ty_organizations (
    id, name, code, type, level, parent_id, leader_id,
    province, city, district, street, address,
    contact, phone, email, description, sort, status,
    volunteer_count, case_count, created_at, updated_at
) VALUES (
    '00000000-0000-0000-0000-000000000001',
    '团圆志愿者总部',
    'ROOT',
    'root',
    1,
    NULL,
    NULL,
    '全国',
    '',
    '',
    '',
    '',
    '',
    '',
    '',
    '团圆寻亲志愿者系统总部',
    0,
    'active',
    0,
    0,
    NOW(),
    NOW()
) ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    updated_at = NOW();

-- ============================================
-- 2. 超级管理员
-- ============================================
-- 先删除已存在的超级管理员（如果存在）
DELETE FROM ty_users WHERE phone = '13800138000' OR role = 'super_admin';

-- 插入超级管理员
INSERT INTO ty_users (
    id, union_id, open_id, nickname, avatar, phone, email,
    real_name, id_card, password, role, status, org_id,
    last_login, login_ip, created_at, updated_at
) VALUES (
    '00000000-0000-0000-0000-000000000002',
    '',
    '',
    '超级管理员',
    '',
    '13800138000',
    'admin@cntunyuan.com',
    '系统管理员',
    '',
    '$2a$10$qU2L2VrILKwGWHoVYuZMFuS.3yKCYxPhyLCMj9x654g918V.Rp8nW',
    'super_admin',
    'active',
    '00000000-0000-0000-0000-000000000001',
    NULL,
    '',
    NOW(),
    NOW()
);

-- ============================================
-- 3. 示例省级组织 (可选)
-- ============================================
INSERT INTO ty_organizations (
    id, name, code, type, level, parent_id, leader_id,
    province, city, district, street, address,
    contact, phone, email, description, sort, status,
    volunteer_count, case_count, created_at, updated_at
) VALUES 
('10000000-0000-0000-0000-000000000001', '北京志愿者协会', 'BJ-001', 'province', 2, '00000000-0000-0000-0000-000000000001', NULL, '北京市', '', '', '', '', '', '', '', '北京市团圆志愿者协会', 1, 'active', 0, 0, NOW(), NOW()),
('10000000-0000-0000-0000-000000000002', '上海志愿者协会', 'SH-001', 'province', 2, '00000000-0000-0000-0000-000000000001', NULL, '上海市', '', '', '', '', '', '', '', '上海市团圆志愿者协会', 2, 'active', 0, 0, NOW(), NOW()),
('10000000-0000-0000-0000-000000000003', '广东志愿者协会', 'GD-001', 'province', 2, '00000000-0000-0000-0000-000000000001', NULL, '广东省', '', '', '', '', '', '', '', '广东省团圆志愿者协会', 3, 'active', 0, 0, NOW(), NOW())
ON CONFLICT (code) DO NOTHING;

COMMIT;

-- 验证数据
SELECT '初始化完成' as status;
SELECT id, name, code, type, status FROM ty_organizations WHERE code = 'ROOT';
SELECT id, nickname, phone, email, role, status FROM ty_users WHERE role = 'super_admin';
