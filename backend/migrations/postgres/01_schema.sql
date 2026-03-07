-- ============================================================
-- 团圆寻亲志愿者系统 - PostgreSQL 数据库结构
-- 版本: 1.0.0
-- 说明: 此脚本创建所有表结构、索引和外键约束
-- ============================================================

-- 启用 UUID 扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
-- 1. 组织表 (ty_organizations)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('root', 'province', 'city', 'district', 'street', 'community', 'team')),
    level INTEGER NOT NULL DEFAULT 1,
    parent_id UUID,
    description TEXT,
    address VARCHAR(255),
    contact_name VARCHAR(50),
    contact_phone VARCHAR(20),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    logo VARCHAR(255),
    sort_order INTEGER NOT NULL DEFAULT 0,
    
    CONSTRAINT fk_org_parent FOREIGN KEY (parent_id) REFERENCES ty_organizations(id) ON DELETE SET NULL ON UPDATE CASCADE
);

COMMENT ON TABLE ty_organizations IS '组织表';
COMMENT ON COLUMN ty_organizations.type IS '组织类型: root-总部, province-省级, city-市级, district-区级, street-街道, community-社区, team-团队';

-- 组织表索引
CREATE INDEX idx_organizations_parent_id ON ty_organizations(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_organizations_type ON ty_organizations(type) WHERE deleted_at IS NULL;
CREATE INDEX idx_organizations_status ON ty_organizations(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_organizations_deleted_at ON ty_organizations(deleted_at) WHERE deleted_at IS NOT NULL;

-- ============================================================
-- 2. 用户表 (ty_users)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    nickname VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL UNIQUE,
    email VARCHAR(100) UNIQUE,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'volunteer' CHECK (role IN ('super_admin', 'admin', 'manager', 'volunteer')),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'banned')),
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
    wx_openid VARCHAR(100) UNIQUE,
    wx_unionid VARCHAR(100),
    
    CONSTRAINT fk_user_org FOREIGN KEY (org_id) REFERENCES ty_organizations(id) ON DELETE RESTRICT ON UPDATE CASCADE
);

COMMENT ON TABLE ty_users IS '用户表';
COMMENT ON COLUMN ty_users.role IS '角色: super_admin-超级管理员, admin-管理员, manager-管理者, volunteer-志愿者';
COMMENT ON COLUMN ty_users.status IS '状态: active-活跃, inactive-禁用, banned-封禁';

-- 用户表索引
CREATE INDEX idx_users_org_id ON ty_users(org_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role ON ty_users(role) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_status ON ty_users(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_wx_openid ON ty_users(wx_openid) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_deleted_at ON ty_users(deleted_at) WHERE deleted_at IS NOT NULL;

-- ============================================================
-- 3. 权限表 (ty_permissions)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    name VARCHAR(100) NOT NULL UNIQUE,
    code VARCHAR(100) NOT NULL UNIQUE,
    description VARCHAR(255),
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL
);

COMMENT ON TABLE ty_permissions IS '权限表';

-- 权限表索引
CREATE INDEX idx_permissions_code ON ty_permissions(code) WHERE deleted_at IS NULL;
CREATE INDEX idx_permissions_deleted_at ON ty_permissions(deleted_at) WHERE deleted_at IS NOT NULL;

-- ============================================================
-- 4. 用户权限关联表 (ty_user_permissions)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_user_permissions (
    user_id UUID NOT NULL,
    permission_id UUID NOT NULL,
    granted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    granted_by UUID,
    
    PRIMARY KEY (user_id, permission_id),
    CONSTRAINT fk_up_user FOREIGN KEY (user_id) REFERENCES ty_users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_up_permission FOREIGN KEY (permission_id) REFERENCES ty_permissions(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_up_granted_by FOREIGN KEY (granted_by) REFERENCES ty_users(id) ON DELETE SET NULL ON UPDATE CASCADE
);

COMMENT ON TABLE ty_user_permissions IS '用户权限关联表';

-- ============================================================
-- 5. 组织统计表 (ty_org_stats)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_org_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    org_id UUID NOT NULL UNIQUE,
    total_volunteers INTEGER NOT NULL DEFAULT 0,
    active_volunteers INTEGER NOT NULL DEFAULT 0,
    total_cases INTEGER NOT NULL DEFAULT 0,
    active_cases INTEGER NOT NULL DEFAULT 0,
    completed_cases INTEGER NOT NULL DEFAULT 0,
    total_tasks INTEGER NOT NULL DEFAULT 0,
    pending_tasks INTEGER NOT NULL DEFAULT 0,
    
    CONSTRAINT fk_stats_org FOREIGN KEY (org_id) REFERENCES ty_organizations(id) ON DELETE CASCADE ON UPDATE CASCADE
);

COMMENT ON TABLE ty_org_stats IS '组织统计表';

-- ============================================================
-- 6. 走失人员表 (ty_missing_persons)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_missing_persons (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    case_no VARCHAR(50) UNIQUE,
    name VARCHAR(50) NOT NULL,
    gender VARCHAR(10) NOT NULL,
    birth_date DATE,
    age INTEGER,
    height INTEGER,
    weight INTEGER,
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
    contact_rel VARCHAR(20) NOT NULL,
    alt_contact VARCHAR(20),
    
    status VARCHAR(20) NOT NULL DEFAULT 'missing' CHECK (status IN ('missing', 'searching', 'found', 'reunited', 'closed')),
    urgency VARCHAR(20) NOT NULL DEFAULT 'medium' CHECK (urgency IN ('critical', 'high', 'medium', 'low')),
    views INTEGER NOT NULL DEFAULT 0,
    share_count INTEGER NOT NULL DEFAULT 0,
    
    reporter_id UUID NOT NULL,
    org_id UUID NOT NULL,
    assigned_to UUID,
    
    found_time TIMESTAMP WITH TIME ZONE,
    found_location VARCHAR(255),
    found_note TEXT,
    
    CONSTRAINT fk_mp_reporter FOREIGN KEY (reporter_id) REFERENCES ty_users(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT fk_mp_org FOREIGN KEY (org_id) REFERENCES ty_organizations(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT fk_mp_assigned FOREIGN KEY (assigned_to) REFERENCES ty_users(id) ON DELETE SET NULL ON UPDATE CASCADE
);

COMMENT ON TABLE ty_missing_persons IS '走失人员表';
COMMENT ON COLUMN ty_missing_persons.status IS '状态: missing-待寻找, searching-寻找中, found-已找到, reunited-已团聚, closed-已关闭';
COMMENT ON COLUMN ty_missing_persons.urgency IS '紧急程度: critical-紧急, high-高, medium-中, low-低';

-- 走失人员表索引
CREATE INDEX idx_missing_persons_status ON ty_missing_persons(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_missing_persons_urgency ON ty_missing_persons(urgency) WHERE deleted_at IS NULL;
CREATE INDEX idx_missing_persons_reporter ON ty_missing_persons(reporter_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_missing_persons_org ON ty_missing_persons(org_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_missing_persons_assigned ON ty_missing_persons(assigned_to) WHERE deleted_at IS NULL;
CREATE INDEX idx_missing_persons_missing_time ON ty_missing_persons(missing_time) WHERE deleted_at IS NULL;
CREATE INDEX idx_missing_persons_location ON ty_missing_persons(province, city, district) WHERE deleted_at IS NULL;
CREATE INDEX idx_missing_persons_deleted_at ON ty_missing_persons(deleted_at) WHERE deleted_at IS NOT NULL;

-- ============================================================
-- 7. 走失人员轨迹表 (ty_missing_person_tracks)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_missing_person_tracks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    missing_person_id UUID NOT NULL,
    reporter_id UUID NOT NULL,
    location VARCHAR(255),
    province VARCHAR(50),
    city VARCHAR(50),
    district VARCHAR(50),
    address VARCHAR(255),
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    description TEXT NOT NULL,
    photos JSONB,
    video_url VARCHAR(255),
    audio_url VARCHAR(255),
    lat DOUBLE PRECISION,
    lng DOUBLE PRECISION,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'rejected')),
    is_key_point BOOLEAN NOT NULL DEFAULT FALSE,
    
    CONSTRAINT fk_mpt_missing_person FOREIGN KEY (missing_person_id) REFERENCES ty_missing_persons(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_mpt_reporter FOREIGN KEY (reporter_id) REFERENCES ty_users(id) ON DELETE RESTRICT ON UPDATE CASCADE
);

COMMENT ON TABLE ty_missing_person_tracks IS '走失人员轨迹表';

-- 轨迹表索引
CREATE INDEX idx_tracks_missing_person ON ty_missing_person_tracks(missing_person_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tracks_reporter ON ty_missing_person_tracks(reporter_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tracks_time ON ty_missing_person_tracks(time) WHERE deleted_at IS NULL;
CREATE INDEX idx_tracks_status ON ty_missing_person_tracks(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_tracks_key_point ON ty_missing_person_tracks(is_key_point) WHERE is_key_point = TRUE AND deleted_at IS NULL;
CREATE INDEX idx_tracks_location ON ty_missing_person_tracks(province, city, district) WHERE deleted_at IS NULL;
CREATE INDEX idx_tracks_deleted_at ON ty_missing_person_tracks(deleted_at) WHERE deleted_at IS NOT NULL;

-- ============================================================
-- 8. 走失人员照片表 (ty_missing_photos)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_missing_photos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    missing_person_id UUID NOT NULL,
    url VARCHAR(500) NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'normal' CHECK (type IN ('normal', 'simulated', 'feature')),
    description TEXT,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    
    CONSTRAINT fk_mp_photos_missing_person FOREIGN KEY (missing_person_id) REFERENCES ty_missing_persons(id) ON DELETE CASCADE ON UPDATE CASCADE
);

COMMENT ON TABLE ty_missing_photos IS '走失人员照片表';
COMMENT ON COLUMN ty_missing_photos.type IS '照片类型: normal-普通照片, simulated-模拟照片, feature-特征照片';

-- 照片表索引
CREATE INDEX idx_photos_missing_person ON ty_missing_photos(missing_person_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_photos_primary ON ty_missing_photos(missing_person_id, is_primary) WHERE is_primary = TRUE AND deleted_at IS NULL;
CREATE INDEX idx_photos_deleted_at ON ty_missing_photos(deleted_at) WHERE deleted_at IS NOT NULL;

-- ============================================================
-- 9. 任务表 (ty_tasks)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    title VARCHAR(200) NOT NULL,
    description TEXT,
    type VARCHAR(20) NOT NULL CHECK (type IN ('search', 'verify', 'assist', 'follow', 'interview', 'other')),
    priority VARCHAR(20) NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'pending', 'assigned', 'processing', 'completed', 'cancelled', 'overdue')),
    
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
    result_photos JSONB,
    feedback TEXT,
    progress INTEGER NOT NULL DEFAULT 0 CHECK (progress >= 0 AND progress <= 100),
    view_count INTEGER NOT NULL DEFAULT 0,
    
    CONSTRAINT fk_task_creator FOREIGN KEY (creator_id) REFERENCES ty_users(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT fk_task_assignee FOREIGN KEY (assignee_id) REFERENCES ty_users(id) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT fk_task_org FOREIGN KEY (org_id) REFERENCES ty_organizations(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT fk_task_missing_person FOREIGN KEY (missing_person_id) REFERENCES ty_missing_persons(id) ON DELETE SET NULL ON UPDATE CASCADE
);

COMMENT ON TABLE ty_tasks IS '任务表';
COMMENT ON COLUMN ty_tasks.type IS '任务类型: search-搜索, verify-核实, assist-协助, follow-跟进, interview-寻访, other-其他';
COMMENT ON COLUMN ty_tasks.priority IS '优先级: low-低, medium-中, high-高, urgent-紧急';
COMMENT ON COLUMN ty_tasks.status IS '状态: draft-草稿, pending-待分配, assigned-已分配, processing-进行中, completed-已完成, cancelled-已取消, overdue-已逾期';

-- 任务表索引
CREATE INDEX idx_tasks_status ON ty_tasks(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_type ON ty_tasks(type) WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_priority ON ty_tasks(priority) WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_creator ON ty_tasks(creator_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_assignee ON ty_tasks(assignee_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_org ON ty_tasks(org_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_missing_person ON ty_tasks(missing_person_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_deadline ON ty_tasks(deadline) WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_deleted_at ON ty_tasks(deleted_at) WHERE deleted_at IS NOT NULL;

-- ============================================================
-- 9. 任务附件表 (ty_task_attachments)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_task_attachments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    task_id UUID NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_url VARCHAR(255) NOT NULL,
    file_type VARCHAR(50),
    file_size BIGINT NOT NULL DEFAULT 0,
    description TEXT,
    uploaded_by UUID NOT NULL,
    
    CONSTRAINT fk_ta_task FOREIGN KEY (task_id) REFERENCES ty_tasks(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_ta_uploader FOREIGN KEY (uploaded_by) REFERENCES ty_users(id) ON DELETE RESTRICT ON UPDATE CASCADE
);

COMMENT ON TABLE ty_task_attachments IS '任务附件表';

-- 任务附件表索引
CREATE INDEX idx_task_attachments_task ON ty_task_attachments(task_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_task_attachments_deleted_at ON ty_task_attachments(deleted_at) WHERE deleted_at IS NOT NULL;

-- ============================================================
-- 10. 任务日志表 (ty_task_logs)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_task_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    task_id UUID NOT NULL,
    user_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    old_status VARCHAR(20),
    new_status VARCHAR(20),
    content TEXT,
    
    CONSTRAINT fk_tl_task FOREIGN KEY (task_id) REFERENCES ty_tasks(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_tl_user FOREIGN KEY (user_id) REFERENCES ty_users(id) ON DELETE RESTRICT ON UPDATE CASCADE
);

COMMENT ON TABLE ty_task_logs IS '任务日志表';

-- 任务日志表索引
CREATE INDEX idx_task_logs_task ON ty_task_logs(task_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_task_logs_user ON ty_task_logs(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_task_logs_created ON ty_task_logs(created_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_task_logs_deleted_at ON ty_task_logs(deleted_at) WHERE deleted_at IS NOT NULL;

-- ============================================================
-- 11. 任务评论表 (ty_task_comments)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_task_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    task_id UUID NOT NULL,
    user_id UUID NOT NULL,
    content TEXT NOT NULL,
    parent_id UUID,
    
    CONSTRAINT fk_tc_task FOREIGN KEY (task_id) REFERENCES ty_tasks(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_tc_user FOREIGN KEY (user_id) REFERENCES ty_users(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT fk_tc_parent FOREIGN KEY (parent_id) REFERENCES ty_task_comments(id) ON DELETE CASCADE ON UPDATE CASCADE
);

COMMENT ON TABLE ty_task_comments IS '任务评论表';

-- 任务评论表索引
CREATE INDEX idx_task_comments_task ON ty_task_comments(task_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_task_comments_user ON ty_task_comments(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_task_comments_parent ON ty_task_comments(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_task_comments_deleted_at ON ty_task_comments(deleted_at) WHERE deleted_at IS NOT NULL;

-- ============================================================
-- 12. 方言表 (ty_dialects)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_dialects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    title VARCHAR(100) NOT NULL,
    content TEXT,
    region VARCHAR(100) NOT NULL,
    province VARCHAR(50),
    city VARCHAR(50),
    dialect_type VARCHAR(20) NOT NULL DEFAULT 'phrase' CHECK (dialect_type IN ('phrase', 'story', 'song', 'daily', 'other')),
    audio_url VARCHAR(255) NOT NULL,
    duration INTEGER NOT NULL DEFAULT 0,
    file_size INTEGER NOT NULL DEFAULT 0,
    format VARCHAR(10),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'pending')),
    is_featured BOOLEAN NOT NULL DEFAULT FALSE,
    play_count INTEGER NOT NULL DEFAULT 0,
    like_count INTEGER NOT NULL DEFAULT 0,
    comment_count INTEGER NOT NULL DEFAULT 0,
    tags JSONB,
    description TEXT,
    uploader_id UUID NOT NULL,
    org_id UUID NOT NULL,
    
    CONSTRAINT fk_dialect_uploader FOREIGN KEY (uploader_id) REFERENCES ty_users(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT fk_dialect_org FOREIGN KEY (org_id) REFERENCES ty_organizations(id) ON DELETE RESTRICT ON UPDATE CASCADE
);

COMMENT ON TABLE ty_dialects IS '方言表';
COMMENT ON COLUMN ty_dialects.dialect_type IS '方言类型: phrase-短语, story-故事, song-歌曲, daily-日常用语, other-其他';
COMMENT ON COLUMN ty_dialects.status IS '状态: active-活跃, inactive-禁用, pending-待审核';

-- 方言表索引
CREATE INDEX idx_dialects_status ON ty_dialects(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_dialects_type ON ty_dialects(dialect_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_dialects_region ON ty_dialects(region) WHERE deleted_at IS NULL;
CREATE INDEX idx_dialects_uploader ON ty_dialects(uploader_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_dialects_org ON ty_dialects(org_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_dialects_featured ON ty_dialects(is_featured) WHERE is_featured = TRUE AND deleted_at IS NULL;
CREATE INDEX idx_dialects_deleted_at ON ty_dialects(deleted_at) WHERE deleted_at IS NOT NULL;

-- ============================================================
-- 13. 方言评论表 (ty_dialect_comments)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_dialect_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    dialect_id UUID NOT NULL,
    user_id UUID NOT NULL,
    content TEXT NOT NULL,
    parent_id UUID,
    reply_count INTEGER NOT NULL DEFAULT 0,
    like_count INTEGER NOT NULL DEFAULT 0,
    
    CONSTRAINT fk_dc_dialect FOREIGN KEY (dialect_id) REFERENCES ty_dialects(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_dc_user FOREIGN KEY (user_id) REFERENCES ty_users(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT fk_dc_parent FOREIGN KEY (parent_id) REFERENCES ty_dialect_comments(id) ON DELETE CASCADE ON UPDATE CASCADE
);

COMMENT ON TABLE ty_dialect_comments IS '方言评论表';

-- 方言评论表索引
CREATE INDEX idx_dialect_comments_dialect ON ty_dialect_comments(dialect_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_dialect_comments_user ON ty_dialect_comments(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_dialect_comments_parent ON ty_dialect_comments(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_dialect_comments_deleted_at ON ty_dialect_comments(deleted_at) WHERE deleted_at IS NOT NULL;

-- ============================================================
-- 14. 方言点赞表 (ty_dialect_likes)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_dialect_likes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    dialect_id UUID NOT NULL,
    user_id UUID NOT NULL,
    
    UNIQUE (dialect_id, user_id),
    CONSTRAINT fk_dl_dialect FOREIGN KEY (dialect_id) REFERENCES ty_dialects(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_dl_user FOREIGN KEY (user_id) REFERENCES ty_users(id) ON DELETE CASCADE ON UPDATE CASCADE
);

COMMENT ON TABLE ty_dialect_likes IS '方言点赞表';

-- 方言点赞表索引
CREATE INDEX idx_dialect_likes_dialect ON ty_dialect_likes(dialect_id);
CREATE INDEX idx_dialect_likes_user ON ty_dialect_likes(user_id);

-- ============================================================
-- 15. 方言播放记录表 (ty_dialect_play_logs)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_dialect_play_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    dialect_id UUID NOT NULL,
    user_id UUID,
    ip VARCHAR(50),
    user_agent VARCHAR(255),
    duration INTEGER NOT NULL DEFAULT 0,
    
    CONSTRAINT fk_dpl_dialect FOREIGN KEY (dialect_id) REFERENCES ty_dialects(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_dpl_user FOREIGN KEY (user_id) REFERENCES ty_users(id) ON DELETE SET NULL ON UPDATE CASCADE
);

COMMENT ON TABLE ty_dialect_play_logs IS '方言播放记录表';

-- 方言播放记录表索引
CREATE INDEX idx_dialect_play_logs_dialect ON ty_dialect_play_logs(dialect_id);
CREATE INDEX idx_dialect_play_logs_user ON ty_dialect_play_logs(user_id);
CREATE INDEX idx_dialect_play_logs_created ON ty_dialect_play_logs(created_at);

-- ============================================================
-- 16. 文件表 (ty_files)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    file_name VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    file_type VARCHAR(20) NOT NULL CHECK (file_type IN ('image', 'audio', 'video', 'document')),
    mime_type VARCHAR(100),
    size BIGINT NOT NULL DEFAULT 0,
    path VARCHAR(500) NOT NULL,
    url VARCHAR(500),
    storage_type VARCHAR(20) NOT NULL CHECK (storage_type IN ('local', 'oss', 'cos')),
    uploader_id UUID,
    entity_type VARCHAR(50),
    entity_id UUID,
    description TEXT,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    
    CONSTRAINT fk_file_uploader FOREIGN KEY (uploader_id) REFERENCES ty_users(id) ON DELETE SET NULL ON UPDATE CASCADE
);

COMMENT ON TABLE ty_files IS '文件表';
COMMENT ON COLUMN ty_files.file_type IS '文件类型: image-图片, audio-音频, video-视频, document-文档';
COMMENT ON COLUMN ty_files.storage_type IS '存储类型: local-本地, oss-阿里云OSS, cos-腾讯云COS';

-- 文件表索引
CREATE INDEX idx_files_type ON ty_files(file_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_files_uploader ON ty_files(uploader_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_files_entity ON ty_files(entity_type, entity_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_files_deleted ON ty_files(is_deleted) WHERE is_deleted = TRUE AND deleted_at IS NULL;
CREATE INDEX idx_files_deleted_at ON ty_files(deleted_at) WHERE deleted_at IS NOT NULL;

-- ============================================================
-- 17. 创建更新时间触发器函数
-- ============================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 为所有表创建更新时间触发器
CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON ty_organizations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON ty_users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_permissions_updated_at BEFORE UPDATE ON ty_permissions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_org_stats_updated_at BEFORE UPDATE ON ty_org_stats
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_missing_persons_updated_at BEFORE UPDATE ON ty_missing_persons
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_missing_person_tracks_updated_at BEFORE UPDATE ON ty_missing_person_tracks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_missing_photos_updated_at BEFORE UPDATE ON ty_missing_photos
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_tasks_updated_at BEFORE UPDATE ON ty_tasks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_task_attachments_updated_at BEFORE UPDATE ON ty_task_attachments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_task_logs_updated_at BEFORE UPDATE ON ty_task_logs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_task_comments_updated_at BEFORE UPDATE ON ty_task_comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_dialects_updated_at BEFORE UPDATE ON ty_dialects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_dialect_comments_updated_at BEFORE UPDATE ON ty_dialect_comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_files_updated_at BEFORE UPDATE ON ty_files
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- 完成
-- ============================================================
