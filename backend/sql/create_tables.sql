-- ============================================
-- 团圆寻亲系统 - 表结构创建脚本
-- 编码: UTF-8
-- 数据库: PostgreSQL 14+
-- ============================================

-- 确保使用 UTF-8 编码
SET client_encoding = 'UTF8';

-- ============================================
-- 1. 用户表
-- ============================================
CREATE TABLE IF NOT EXISTS ty_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nickname VARCHAR(100) NOT NULL,
    phone VARCHAR(20) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'volunteer',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    org_id UUID NOT NULL,
    avatar VARCHAR(255),
    last_login_at TIMESTAMP WITH TIME ZONE,
    last_login_ip VARCHAR(50),
    real_name VARCHAR(50),
    id_card VARCHAR(18),
    gender VARCHAR(10),
    address VARCHAR(255),
    emergency VARCHAR(50),
    emergency_tel VARCHAR(20),
    introduction TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_users_phone ON ty_users(phone);
CREATE INDEX idx_users_email ON ty_users(email);
CREATE INDEX idx_users_org_id ON ty_users(org_id);
CREATE INDEX idx_users_role ON ty_users(role);
CREATE INDEX idx_users_status ON ty_users(status);

-- ============================================
-- 2. 权限表
-- ============================================
CREATE TABLE IF NOT EXISTS ty_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    code VARCHAR(100) UNIQUE NOT NULL,
    description VARCHAR(255),
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- 3. 用户权限关联表
-- ============================================
CREATE TABLE IF NOT EXISTS ty_user_permissions (
    user_id UUID NOT NULL,
    permission_id UUID NOT NULL,
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    granted_by UUID,
    PRIMARY KEY (user_id, permission_id)
);

-- ============================================
-- 4. 组织表
-- ============================================
CREATE TABLE IF NOT EXISTS ty_organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    type VARCHAR(20) NOT NULL,
    level INT NOT NULL DEFAULT 1,
    parent_id UUID,
    description TEXT,
    address VARCHAR(255),
    contact_name VARCHAR(50),
    contact_phone VARCHAR(20),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    logo VARCHAR(255),
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_organizations_code ON ty_organizations(code);
CREATE INDEX idx_organizations_parent_id ON ty_organizations(parent_id);
CREATE INDEX idx_organizations_type ON ty_organizations(type);
CREATE INDEX idx_organizations_status ON ty_organizations(status);

-- ============================================
-- 5. 组织统计表
-- ============================================
CREATE TABLE IF NOT EXISTS ty_org_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID UNIQUE NOT NULL,
    total_volunteers INT DEFAULT 0,
    active_volunteers INT DEFAULT 0,
    total_cases INT DEFAULT 0,
    active_cases INT DEFAULT 0,
    completed_cases INT DEFAULT 0,
    total_tasks INT DEFAULT 0,
    pending_tasks INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_org_stats_org_id ON ty_org_stats(org_id);

-- ============================================
-- 6. 走失人员表
-- ============================================
CREATE TABLE IF NOT EXISTS ty_missing_persons (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL,
    gender VARCHAR(10) NOT NULL,
    birth_date DATE,
    age INT,
    height INT,
    weight INT,
    description TEXT,
    photo_url VARCHAR(255),
    missing_time TIMESTAMP WITH TIME ZONE NOT NULL,
    province VARCHAR(50),
    city VARCHAR(50),
    district VARCHAR(50),
    address VARCHAR(255),
    clothes TEXT,
    features TEXT,
    contact_name VARCHAR(50) NOT NULL,
    contact_phone VARCHAR(20) NOT NULL,
    contact_rel VARCHAR(20),
    alt_contact VARCHAR(20),
    status VARCHAR(20) NOT NULL DEFAULT 'missing',
    urgency VARCHAR(20) DEFAULT 'medium',
    views INT DEFAULT 0,
    share_count INT DEFAULT 0,
    reporter_id UUID NOT NULL,
    org_id UUID NOT NULL,
    assigned_to UUID,
    found_time TIMESTAMP WITH TIME ZONE,
    found_location VARCHAR(255),
    found_note TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_missing_persons_status ON ty_missing_persons(status);
CREATE INDEX idx_missing_persons_urgency ON ty_missing_persons(urgency);
CREATE INDEX idx_missing_persons_reporter_id ON ty_missing_persons(reporter_id);
CREATE INDEX idx_missing_persons_org_id ON ty_missing_persons(org_id);
CREATE INDEX idx_missing_persons_assigned_to ON ty_missing_persons(assigned_to);
CREATE INDEX idx_missing_persons_missing_time ON ty_missing_persons(missing_time);

-- ============================================
-- 7. 走失人员轨迹表
-- ============================================
CREATE TABLE IF NOT EXISTS ty_missing_person_tracks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    missing_person_id UUID NOT NULL,
    reporter_id UUID NOT NULL,
    location VARCHAR(255),
    province VARCHAR(50),
    city VARCHAR(50),
    district VARCHAR(50),
    address VARCHAR(255),
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    description TEXT NOT NULL,
    photos JSON,
    video_url VARCHAR(255),
    audio_url VARCHAR(255),
    lat DOUBLE PRECISION,
    lng DOUBLE PRECISION,
    status VARCHAR(20) DEFAULT 'pending',
    is_key_point BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tracks_missing_person_id ON ty_missing_person_tracks(missing_person_id);
CREATE INDEX idx_tracks_reporter_id ON ty_missing_person_tracks(reporter_id);
CREATE INDEX idx_tracks_time ON ty_missing_person_tracks(time);

-- ============================================
-- 8. 方言表
-- ============================================
CREATE TABLE IF NOT EXISTS ty_dialects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(100) NOT NULL,
    content TEXT,
    region VARCHAR(100) NOT NULL,
    province VARCHAR(50),
    city VARCHAR(50),
    dialect_type VARCHAR(20) DEFAULT 'phrase',
    audio_url VARCHAR(255) NOT NULL,
    duration INT NOT NULL,
    file_size INT,
    format VARCHAR(10),
    status VARCHAR(20) DEFAULT 'active',
    is_featured BOOLEAN DEFAULT FALSE,
    play_count INT DEFAULT 0,
    like_count INT DEFAULT 0,
    comment_count INT DEFAULT 0,
    tags JSON,
    description TEXT,
    uploader_id UUID NOT NULL,
    org_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_dialects_status ON ty_dialects(status);
CREATE INDEX idx_dialects_uploader_id ON ty_dialects(uploader_id);
CREATE INDEX idx_dialects_org_id ON ty_dialects(org_id);
CREATE INDEX idx_dialects_region ON ty_dialects(region);
CREATE INDEX idx_dialects_is_featured ON ty_dialects(is_featured);

-- ============================================
-- 9. 方言评论表
-- ============================================
CREATE TABLE IF NOT EXISTS ty_dialect_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    dialect_id UUID NOT NULL,
    user_id UUID NOT NULL,
    content TEXT NOT NULL,
    parent_id UUID,
    reply_count INT DEFAULT 0,
    like_count INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_dialect_comments_dialect_id ON ty_dialect_comments(dialect_id);
CREATE INDEX idx_dialect_comments_user_id ON ty_dialect_comments(user_id);
CREATE INDEX idx_dialect_comments_parent_id ON ty_dialect_comments(parent_id);

-- ============================================
-- 10. 方言点赞表
-- ============================================
CREATE TABLE IF NOT EXISTS ty_dialect_likes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    dialect_id UUID NOT NULL,
    user_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(dialect_id, user_id)
);

CREATE INDEX idx_dialect_likes_dialect_id ON ty_dialect_likes(dialect_id);
CREATE INDEX idx_dialect_likes_user_id ON ty_dialect_likes(user_id);

-- ============================================
-- 11. 方言播放记录表
-- ============================================
CREATE TABLE IF NOT EXISTS ty_dialect_play_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    dialect_id UUID NOT NULL,
    user_id UUID,
    ip VARCHAR(50),
    user_agent VARCHAR(255),
    duration INT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_dialect_play_logs_dialect_id ON ty_dialect_play_logs(dialect_id);
CREATE INDEX idx_dialect_play_logs_user_id ON ty_dialect_play_logs(user_id);
CREATE INDEX idx_dialect_play_logs_created_at ON ty_dialect_play_logs(created_at);

-- ============================================
-- 12. 任务表
-- ============================================
CREATE TABLE IF NOT EXISTS ty_tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(200) NOT NULL,
    description TEXT,
    type VARCHAR(20) NOT NULL,
    priority VARCHAR(20) DEFAULT 'medium',
    status VARCHAR(20) DEFAULT 'draft',
    deadline TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    creator_id UUID NOT NULL,
    assignee_id UUID,
    org_id UUID NOT NULL,
    missing_person_id UUID,
    location VARCHAR(255),
    province VARCHAR(50),
    city VARCHAR(50),
    district VARCHAR(50),
    address VARCHAR(255),
    lat DOUBLE PRECISION,
    lng DOUBLE PRECISION,
    result TEXT,
    result_photos JSON,
    feedback TEXT,
    progress INT DEFAULT 0,
    view_count INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_tasks_status ON ty_tasks(status);
CREATE INDEX idx_tasks_type ON ty_tasks(type);
CREATE INDEX idx_tasks_priority ON ty_tasks(priority);
CREATE INDEX idx_tasks_creator_id ON ty_tasks(creator_id);
CREATE INDEX idx_tasks_assignee_id ON ty_tasks(assignee_id);
CREATE INDEX idx_tasks_org_id ON ty_tasks(org_id);
CREATE INDEX idx_tasks_missing_person_id ON ty_tasks(missing_person_id);
CREATE INDEX idx_tasks_deadline ON ty_tasks(deadline);

-- ============================================
-- 13. 任务附件表
-- ============================================
CREATE TABLE IF NOT EXISTS ty_task_attachments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_url VARCHAR(255) NOT NULL,
    file_type VARCHAR(50),
    file_size BIGINT,
    description TEXT,
    uploaded_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_task_attachments_task_id ON ty_task_attachments(task_id);

-- ============================================
-- 14. 任务日志表
-- ============================================
CREATE TABLE IF NOT EXISTS ty_task_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL,
    user_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    old_status VARCHAR(20),
    new_status VARCHAR(20),
    content TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_task_logs_task_id ON ty_task_logs(task_id);
CREATE INDEX idx_task_logs_user_id ON ty_task_logs(user_id);
CREATE INDEX idx_task_logs_created_at ON ty_task_logs(created_at);

-- ============================================
-- 15. 任务评论表
-- ============================================
CREATE TABLE IF NOT EXISTS ty_task_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL,
    user_id UUID NOT NULL,
    content TEXT NOT NULL,
    parent_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_task_comments_task_id ON ty_task_comments(task_id);
CREATE INDEX idx_task_comments_user_id ON ty_task_comments(user_id);
CREATE INDEX idx_task_comments_parent_id ON ty_task_comments(parent_id);

-- ============================================
-- 16. 文件表
-- ============================================
CREATE TABLE IF NOT EXISTS ty_files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_name VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    file_type VARCHAR(20) NOT NULL,
    mime_type VARCHAR(100),
    size BIGINT NOT NULL,
    path VARCHAR(500) NOT NULL,
    url VARCHAR(500),
    storage_type VARCHAR(20) NOT NULL,
    uploader_id UUID,
    entity_type VARCHAR(50),
    entity_id UUID,
    description TEXT,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_files_file_type ON ty_files(file_type);
CREATE INDEX idx_files_storage_type ON ty_files(storage_type);
CREATE INDEX idx_files_uploader_id ON ty_files(uploader_id);
CREATE INDEX idx_files_entity ON ty_files(entity_type, entity_id);

-- ============================================
-- 完成
-- ============================================
SELECT 'All tables created successfully!' AS status;
