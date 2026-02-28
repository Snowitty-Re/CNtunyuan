-- 团圆寻亲志愿者系统 - 数据库初始化脚本
-- 执行方式: psql -U postgres -d cntuanyuan -f sql/init.sql

-- 注意: 超级管理员密码需要通过工具生成 bcrypt 哈希后替换 :ADMIN_PASSWORD_HASH
-- 或者使用后端提供的命令行工具自动完成

BEGIN;

-- ============================================
-- 1. 根组织 (团圆志愿者总部)
-- ============================================
INSERT INTO ty_organizations (
    id, name, code, type, level, parent_id, leader_id,
    province, city, district, street, address,
    contact, phone, email, description, sort, status,
    created_at, updated_at
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
    NOW(),
    NOW()
) ON CONFLICT (code) DO NOTHING;

-- ============================================
-- 2. 超级管理员
-- 注意: 密码哈希需要替换为实际生成的 bcrypt 哈希
-- 默认密码: admin123
-- ============================================
-- 先删除已存在的超级管理员（如果存在）
DELETE FROM ty_users WHERE phone = :ADMIN_PHONE;

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
    :ADMIN_PHONE,
    :ADMIN_EMAIL,
    '系统管理员',
    '',
    :ADMIN_PASSWORD_HASH,
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
    created_at, updated_at
) VALUES 
('10000000-0000-0000-0000-000000000001', '北京志愿者协会', 'BJ-001', 'province', 2, '00000000-0000-0000-0000-000000000001', NULL, '北京市', '', '', '', '', '', '', '', '北京市团圆志愿者协会', 1, 'active', NOW(), NOW()),
('10000000-0000-0000-0000-000000000002', '上海志愿者协会', 'SH-001', 'province', 2, '00000000-0000-0000-0000-000000000001', NULL, '上海市', '', '', '', '', '', '', '', '上海市团圆志愿者协会', 2, 'active', NOW(), NOW()),
('10000000-0000-0000-0000-000000000003', '广东志愿者协会', 'GD-001', 'province', 2, '00000000-0000-0000-0000-000000000001', NULL, '广东省', '', '', '', '', '', '', '', '广东省团圆志愿者协会', 3, 'active', NOW(), NOW())
ON CONFLICT (code) DO NOTHING;

-- ============================================
-- 4. 初始化完成
-- ============================================
COMMIT;

-- 验证数据
SELECT '根组织信息:' as info;
SELECT id, name, code, type, status FROM ty_organizations WHERE code = 'ROOT';

SELECT '超级管理员信息:' as info;
SELECT id, nickname, phone, email, role, status FROM ty_users WHERE role = 'super_admin';

SELECT '省级组织数量:' as info, COUNT(*) as count FROM ty_organizations WHERE type = 'province';
