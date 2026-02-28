-- 团圆寻亲志愿者系统 - 数据库结构脚本
-- 仅用于查看结构，实际迁移使用 GORM AutoMigrate
-- 此文件可用于文档参考或手动建表

-- 用户表
CREATE TABLE IF NOT EXISTS ty_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    union_id VARCHAR(100) UNIQUE,
    open_id VARCHAR(100) UNIQUE,
    nickname VARCHAR(100),
    avatar VARCHAR(500),
    phone VARCHAR(20),
    email VARCHAR(100),
    real_name VARCHAR(50),
    id_card VARCHAR(18),
    password VARCHAR(255),  -- bcrypt 哈希
    role VARCHAR(20) DEFAULT 'volunteer',
    status VARCHAR(20) DEFAULT 'active',
    org_id UUID,
    last_login TIMESTAMP,
    login_ip VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- 组织表
CREATE TABLE IF NOT EXISTS ty_organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    type VARCHAR(20) NOT NULL,
    level INTEGER NOT NULL,
    parent_id UUID,
    leader_id UUID,
    province VARCHAR(50),
    city VARCHAR(50),
    district VARCHAR(50),
    street VARCHAR(100),
    address TEXT,
    contact VARCHAR(50),
    phone VARCHAR(20),
    email VARCHAR(100),
    description TEXT,
    sort INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    volunteer_count INTEGER DEFAULT 0,
    case_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- 其他表的创建由 GORM AutoMigrate 处理
-- 此文件仅作为参考
