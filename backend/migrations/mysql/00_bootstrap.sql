-- ============================================================
-- 助力团圆志愿者系统 - MySQL 一键初始化（Schema + Alignment + Workflow + Seed）
-- 用途：全新数据库首次初始化
-- ============================================================

-- ============================================================
-- 助力团圆志愿者系统 - MySQL 数据库结构
-- 版本: 1.0.0
-- 说明: 此脚本创建所有表结构、索引和外键约束
-- 字符集: utf8mb4 (支持中文和 Emoji)
-- ============================================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================================
-- 1. 组织表 (ty_organizations)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_organizations (
    id CHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    name VARCHAR(100) NOT NULL COMMENT '组织名称',
    code VARCHAR(50) NOT NULL COMMENT '组织编码',
    type VARCHAR(20) NOT NULL COMMENT '组织类型: root-总部, province-省级, city-市级, district-区级, street-街道, community-社区, team-团队',
    level INT NOT NULL DEFAULT 1 COMMENT '层级',
    parent_id CHAR(36) NULL DEFAULT NULL COMMENT '父组织ID',
    description TEXT COMMENT '描述',
    address VARCHAR(255) COMMENT '地址',
    contact_name VARCHAR(50) COMMENT '联系人',
    contact_phone VARCHAR(20) COMMENT '联系电话',
    status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '状态: active-活跃, inactive-禁用',
    logo VARCHAR(255) COMMENT 'Logo',
    sort_order INT NOT NULL DEFAULT 0 COMMENT '排序',
    
    CONSTRAINT fk_org_parent FOREIGN KEY (parent_id) REFERENCES ty_organizations(id) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT chk_org_type CHECK (type IN ('root', 'province', 'city', 'district', 'street', 'community', 'team')),
    CONSTRAINT chk_org_status CHECK (status IN ('active', 'inactive')),
    UNIQUE KEY uk_org_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='组织表';

-- 组织表索引
CREATE INDEX idx_organizations_parent_id ON ty_organizations(parent_id);
CREATE INDEX idx_organizations_type ON ty_organizations(type);
CREATE INDEX idx_organizations_status ON ty_organizations(status);
CREATE INDEX idx_organizations_deleted_at ON ty_organizations(deleted_at);

-- ============================================================
-- 2. 用户表 (ty_users)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_users (
    id CHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    nickname VARCHAR(100) NOT NULL COMMENT '昵称',
    phone VARCHAR(20) NOT NULL COMMENT '手机号',
    email VARCHAR(100) COMMENT '邮箱',
    password VARCHAR(255) NOT NULL COMMENT '密码',
    role VARCHAR(20) NOT NULL DEFAULT 'volunteer' COMMENT '角色: super_admin-超级管理员, admin-管理员, manager-管理者, volunteer-志愿者',
    status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '状态: active-活跃, inactive-禁用, banned-封禁',
    org_id CHAR(36) NOT NULL COMMENT '组织ID',
    avatar VARCHAR(255) COMMENT '头像',
    last_login_at TIMESTAMP NULL DEFAULT NULL COMMENT '最后登录时间',
    last_login_ip VARCHAR(50) COMMENT '最后登录IP',
    real_name VARCHAR(50) COMMENT '真实姓名',
    id_card VARCHAR(18) COMMENT '身份证号',
    gender VARCHAR(10) COMMENT '性别',
    address VARCHAR(255) COMMENT '地址',
    emergency VARCHAR(50) COMMENT '紧急联系人',
    emergency_tel VARCHAR(20) COMMENT '紧急联系电话',
    introduction TEXT COMMENT '个人介绍',
    wx_openid VARCHAR(100) COMMENT '微信OpenID',
    wx_unionid VARCHAR(100) COMMENT '微信UnionID',
    
    CONSTRAINT fk_user_org FOREIGN KEY (org_id) REFERENCES ty_organizations(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT chk_user_role CHECK (role IN ('super_admin', 'admin', 'manager', 'volunteer')),
    CONSTRAINT chk_user_status CHECK (status IN ('active', 'inactive', 'banned')),
    UNIQUE KEY uk_user_phone (phone),
    UNIQUE KEY uk_user_email (email),
    UNIQUE KEY uk_user_wx_openid (wx_openid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 用户表索引
CREATE INDEX idx_users_org_id ON ty_users(org_id);
CREATE INDEX idx_users_role ON ty_users(role);
CREATE INDEX idx_users_status ON ty_users(status);
CREATE INDEX idx_users_wx_openid ON ty_users(wx_openid);
CREATE INDEX idx_users_deleted_at ON ty_users(deleted_at);

-- ============================================================
-- 3. 权限表 (ty_permissions)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_permissions (
    id CHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    name VARCHAR(100) NOT NULL COMMENT '权限名称',
    code VARCHAR(100) NOT NULL COMMENT '权限代码',
    description VARCHAR(255) COMMENT '描述',
    resource VARCHAR(100) NOT NULL COMMENT '资源',
    action VARCHAR(50) NOT NULL COMMENT '操作',
    
    UNIQUE KEY uk_perm_name (name),
    UNIQUE KEY uk_perm_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限表';

-- 权限表索引
CREATE INDEX idx_permissions_code ON ty_permissions(code);
CREATE INDEX idx_permissions_deleted_at ON ty_permissions(deleted_at);

-- ============================================================
-- 4. 用户权限关联表 (ty_user_permissions)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_user_permissions (
    user_id CHAR(36) NOT NULL COMMENT '用户ID',
    permission_id CHAR(36) NOT NULL COMMENT '权限ID',
    granted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '授权时间',
    granted_by CHAR(36) COMMENT '授权人',
    
    PRIMARY KEY (user_id, permission_id),
    CONSTRAINT fk_up_user FOREIGN KEY (user_id) REFERENCES ty_users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_up_permission FOREIGN KEY (permission_id) REFERENCES ty_permissions(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_up_granted_by FOREIGN KEY (granted_by) REFERENCES ty_users(id) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户权限关联表';

-- ============================================================
-- 5. 组织统计表 (ty_org_stats)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_org_stats (
    id CHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    org_id CHAR(36) NOT NULL COMMENT '组织ID',
    total_volunteers INT NOT NULL DEFAULT 0 COMMENT '志愿者总数',
    active_volunteers INT NOT NULL DEFAULT 0 COMMENT '活跃志愿者数',
    total_cases INT NOT NULL DEFAULT 0 COMMENT '案件总数',
    active_cases INT NOT NULL DEFAULT 0 COMMENT '活跃案件数',
    completed_cases INT NOT NULL DEFAULT 0 COMMENT '已完成案件数',
    total_tasks INT NOT NULL DEFAULT 0 COMMENT '任务总数',
    pending_tasks INT NOT NULL DEFAULT 0 COMMENT '待处理任务数',
    
    CONSTRAINT fk_stats_org FOREIGN KEY (org_id) REFERENCES ty_organizations(id) ON DELETE CASCADE ON UPDATE CASCADE,
    UNIQUE KEY uk_stats_org (org_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='组织统计表';

-- ============================================================
-- 6. 走失人员表 (ty_missing_persons)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_missing_persons (
    id CHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    case_no VARCHAR(50) UNIQUE COMMENT '案件编号',
    name VARCHAR(50) NOT NULL COMMENT '姓名',
    gender VARCHAR(10) NOT NULL COMMENT '性别',
    birth_date DATE COMMENT '出生日期',
    age INT COMMENT '年龄',
    height INT COMMENT '身高(cm)',
    weight INT COMMENT '体重(kg)',
    description TEXT COMMENT '描述',
    case_type VARCHAR(20) NOT NULL DEFAULT 'other' COMMENT '案件类型: elderly-老人, child-儿童, adult-成人, disability-残障, other-其他',
    photo_url VARCHAR(255) COMMENT '照片URL',
    
    missing_time TIMESTAMP NOT NULL COMMENT '走失时间',
    province VARCHAR(50) COMMENT '省',
    city VARCHAR(50) COMMENT '市',
    district VARCHAR(50) COMMENT '区',
    address VARCHAR(255) COMMENT '详细地址',
    missing_latitude DOUBLE NOT NULL DEFAULT 0 COMMENT '走失地点纬度',
    missing_longitude DOUBLE NOT NULL DEFAULT 0 COMMENT '走失地点经度',
    clothes TEXT COMMENT '衣着特征',
    features TEXT COMMENT '体貌特征',
    
    contact_name VARCHAR(50) NOT NULL COMMENT '联系人姓名',
    contact_phone VARCHAR(20) NOT NULL COMMENT '联系人电话',
    contact_rel VARCHAR(20) NOT NULL COMMENT '联系人关系',
    alt_contact VARCHAR(20) COMMENT '备用联系人',
    
    status VARCHAR(20) NOT NULL DEFAULT 'missing' COMMENT '状态: missing-待寻找, searching-寻找中, found-已找到, reunited-已团聚, closed-已关闭',
    urgency VARCHAR(20) NOT NULL DEFAULT 'medium' COMMENT '紧急程度: critical-紧急, high-高, medium-中, low-低',
    views INT NOT NULL DEFAULT 0 COMMENT '浏览次数',
    share_count INT NOT NULL DEFAULT 0 COMMENT '分享次数',
    
    reporter_id CHAR(36) NOT NULL COMMENT '报告人ID',
    org_id CHAR(36) NOT NULL COMMENT '组织ID',
    assigned_to CHAR(36) COMMENT '分配给',
    
    found_time TIMESTAMP NULL DEFAULT NULL COMMENT '找到时间',
    found_location VARCHAR(255) COMMENT '找到地点',
    found_note TEXT COMMENT '找到备注',
    
    CONSTRAINT fk_mp_reporter FOREIGN KEY (reporter_id) REFERENCES ty_users(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT fk_mp_org FOREIGN KEY (org_id) REFERENCES ty_organizations(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT fk_mp_assigned FOREIGN KEY (assigned_to) REFERENCES ty_users(id) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT chk_mp_status CHECK (status IN ('missing', 'searching', 'found', 'reunited', 'closed')),
    CONSTRAINT chk_mp_urgency CHECK (urgency IN ('critical', 'high', 'medium', 'low')),
    CONSTRAINT chk_mp_case_type CHECK (case_type IN ('elderly', 'child', 'adult', 'disability', 'other')),
    CONSTRAINT chk_mp_missing_latitude CHECK (missing_latitude BETWEEN -90 AND 90),
    CONSTRAINT chk_mp_missing_longitude CHECK (missing_longitude BETWEEN -180 AND 180)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='走失人员表';

-- 走失人员表索引
CREATE INDEX idx_missing_persons_status ON ty_missing_persons(status);
CREATE INDEX idx_missing_persons_urgency ON ty_missing_persons(urgency);
CREATE INDEX idx_missing_persons_case_type ON ty_missing_persons(case_type);
CREATE INDEX idx_missing_persons_reporter ON ty_missing_persons(reporter_id);
CREATE INDEX idx_missing_persons_org ON ty_missing_persons(org_id);
CREATE INDEX idx_missing_persons_assigned ON ty_missing_persons(assigned_to);
CREATE INDEX idx_missing_persons_missing_time ON ty_missing_persons(missing_time);
CREATE INDEX idx_missing_persons_location ON ty_missing_persons(province, city, district);
CREATE INDEX idx_missing_persons_geo ON ty_missing_persons(missing_latitude, missing_longitude);
CREATE INDEX idx_missing_persons_deleted_at ON ty_missing_persons(deleted_at);

-- ============================================================
-- 7. 走失人员轨迹表 (ty_missing_person_tracks)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_missing_person_tracks (
    id CHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    missing_person_id CHAR(36) NOT NULL COMMENT '走失人员ID',
    reporter_id CHAR(36) NOT NULL COMMENT '报告人ID',
    location VARCHAR(255) COMMENT '位置',
    province VARCHAR(50) COMMENT '省',
    city VARCHAR(50) COMMENT '市',
    district VARCHAR(50) COMMENT '区',
    address VARCHAR(255) COMMENT '详细地址',
    time TIMESTAMP NOT NULL COMMENT '时间',
    description TEXT NOT NULL COMMENT '描述',
    photos JSON COMMENT '照片JSON数组',
    video_url VARCHAR(255) COMMENT '视频URL',
    audio_url VARCHAR(255) COMMENT '音频URL',
    lat DOUBLE COMMENT '纬度',
    lng DOUBLE COMMENT '经度',
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '状态: pending-待确认, confirmed-已确认, rejected-已拒绝',
    is_key_point TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否关键点',
    
    CONSTRAINT fk_mpt_missing_person FOREIGN KEY (missing_person_id) REFERENCES ty_missing_persons(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_mpt_reporter FOREIGN KEY (reporter_id) REFERENCES ty_users(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT chk_mpt_status CHECK (status IN ('pending', 'confirmed', 'rejected'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='走失人员轨迹表';

-- 轨迹表索引
CREATE INDEX idx_tracks_missing_person ON ty_missing_person_tracks(missing_person_id);
CREATE INDEX idx_tracks_reporter ON ty_missing_person_tracks(reporter_id);
CREATE INDEX idx_tracks_time ON ty_missing_person_tracks(time);
CREATE INDEX idx_tracks_status ON ty_missing_person_tracks(status);
CREATE INDEX idx_tracks_key_point ON ty_missing_person_tracks(is_key_point);
CREATE INDEX idx_tracks_location ON ty_missing_person_tracks(province, city, district);
CREATE INDEX idx_tracks_deleted_at ON ty_missing_person_tracks(deleted_at);

-- ============================================================
-- 8. 走失人员照片表 (ty_missing_photos)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_missing_photos (
    id CHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    missing_person_id CHAR(36) NOT NULL COMMENT '走失人员ID',
    url VARCHAR(500) NOT NULL COMMENT '照片URL',
    type VARCHAR(20) NOT NULL DEFAULT 'normal' COMMENT '照片类型: normal-普通, simulated-模拟, feature-特征',
    description TEXT COMMENT '描述',
    is_primary TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否主照片',
    
    CONSTRAINT fk_mp_photos_missing_person FOREIGN KEY (missing_person_id) REFERENCES ty_missing_persons(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT chk_mp_photos_type CHECK (type IN ('normal', 'simulated', 'feature'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='走失人员照片表';

-- 照片表索引
CREATE INDEX idx_photos_missing_person ON ty_missing_photos(missing_person_id);
CREATE INDEX idx_photos_primary ON ty_missing_photos(missing_person_id, is_primary);
CREATE INDEX idx_photos_deleted_at ON ty_missing_photos(deleted_at);

-- ============================================================
-- 9. 任务表 (ty_tasks)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_tasks (
    id CHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    title VARCHAR(200) NOT NULL COMMENT '标题',
    description TEXT COMMENT '描述',
    type VARCHAR(20) NOT NULL COMMENT '类型: search-搜索, verify-核实, assist-协助, follow-跟进, interview-寻访, other-其他',
    priority VARCHAR(20) NOT NULL DEFAULT 'medium' COMMENT '优先级: low-低, medium-中, high-高, urgent-紧急',
    status VARCHAR(20) NOT NULL DEFAULT 'draft' COMMENT '状态: draft-草稿, pending-待分配, assigned-已分配, processing-进行中, completed-已完成, cancelled-已取消, overdue-已逾期',
    
    deadline TIMESTAMP NULL DEFAULT NULL COMMENT '截止时间',
    started_at TIMESTAMP NULL DEFAULT NULL COMMENT '开始时间',
    completed_at TIMESTAMP NULL DEFAULT NULL COMMENT '完成时间',
    
    creator_id CHAR(36) NOT NULL COMMENT '创建人ID',
    assignee_id CHAR(36) COMMENT '分配人ID',
    org_id CHAR(36) NOT NULL COMMENT '组织ID',
    missing_person_id CHAR(36) COMMENT '关联走失人员ID',
    
    location VARCHAR(255) COMMENT '位置',
    province VARCHAR(50) COMMENT '省',
    city VARCHAR(50) COMMENT '市',
    district VARCHAR(50) COMMENT '区',
    address VARCHAR(255) COMMENT '详细地址',
    lat DOUBLE COMMENT '纬度',
    lng DOUBLE COMMENT '经度',
    
    result TEXT COMMENT '结果',
    result_photos JSON COMMENT '结果照片JSON数组',
    feedback TEXT COMMENT '反馈',
    progress INT NOT NULL DEFAULT 0 COMMENT '进度(0-100)',
    view_count INT NOT NULL DEFAULT 0 COMMENT '浏览次数',
    
    CONSTRAINT fk_task_creator FOREIGN KEY (creator_id) REFERENCES ty_users(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT fk_task_assignee FOREIGN KEY (assignee_id) REFERENCES ty_users(id) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT fk_task_org FOREIGN KEY (org_id) REFERENCES ty_organizations(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT fk_task_missing_person FOREIGN KEY (missing_person_id) REFERENCES ty_missing_persons(id) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT chk_task_type CHECK (type IN ('search', 'verify', 'assist', 'follow', 'interview', 'other')),
    CONSTRAINT chk_task_priority CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    CONSTRAINT chk_task_status CHECK (status IN ('draft', 'pending', 'assigned', 'processing', 'completed', 'cancelled', 'overdue')),
    CONSTRAINT chk_task_progress CHECK (progress >= 0 AND progress <= 100)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务表';

-- 任务表索引
CREATE INDEX idx_tasks_status ON ty_tasks(status);
CREATE INDEX idx_tasks_type ON ty_tasks(type);
CREATE INDEX idx_tasks_priority ON ty_tasks(priority);
CREATE INDEX idx_tasks_creator ON ty_tasks(creator_id);
CREATE INDEX idx_tasks_assignee ON ty_tasks(assignee_id);
CREATE INDEX idx_tasks_org ON ty_tasks(org_id);
CREATE INDEX idx_tasks_missing_person ON ty_tasks(missing_person_id);
CREATE INDEX idx_tasks_deadline ON ty_tasks(deadline);
CREATE INDEX idx_tasks_deleted_at ON ty_tasks(deleted_at);

-- ============================================================
-- 9. 任务附件表 (ty_task_attachments)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_task_attachments (
    id CHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    task_id CHAR(36) NOT NULL COMMENT '任务ID',
    file_name VARCHAR(255) NOT NULL COMMENT '文件名',
    file_url VARCHAR(255) NOT NULL COMMENT '文件URL',
    file_type VARCHAR(50) COMMENT '文件类型',
    file_size BIGINT NOT NULL DEFAULT 0 COMMENT '文件大小',
    description TEXT COMMENT '描述',
    uploaded_by CHAR(36) NOT NULL COMMENT '上传人',
    
    CONSTRAINT fk_ta_task FOREIGN KEY (task_id) REFERENCES ty_tasks(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_ta_uploader FOREIGN KEY (uploaded_by) REFERENCES ty_users(id) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务附件表';

-- 任务附件表索引
CREATE INDEX idx_task_attachments_task ON ty_task_attachments(task_id);
CREATE INDEX idx_task_attachments_deleted_at ON ty_task_attachments(deleted_at);

-- ============================================================
-- 10. 任务日志表 (ty_task_logs)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_task_logs (
    id CHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    task_id CHAR(36) NOT NULL COMMENT '任务ID',
    user_id CHAR(36) NOT NULL COMMENT '用户ID',
    action VARCHAR(50) NOT NULL COMMENT '操作',
    old_status VARCHAR(20) COMMENT '旧状态',
    new_status VARCHAR(20) COMMENT '新状态',
    content TEXT COMMENT '内容',
    
    CONSTRAINT fk_tl_task FOREIGN KEY (task_id) REFERENCES ty_tasks(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_tl_user FOREIGN KEY (user_id) REFERENCES ty_users(id) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务日志表';

-- 任务日志表索引
CREATE INDEX idx_task_logs_task ON ty_task_logs(task_id);
CREATE INDEX idx_task_logs_user ON ty_task_logs(user_id);
CREATE INDEX idx_task_logs_created ON ty_task_logs(created_at);
CREATE INDEX idx_task_logs_deleted_at ON ty_task_logs(deleted_at);

-- ============================================================
-- 11. 任务评论表 (ty_task_comments)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_task_comments (
    id CHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    task_id CHAR(36) NOT NULL COMMENT '任务ID',
    user_id CHAR(36) NOT NULL COMMENT '用户ID',
    content TEXT NOT NULL COMMENT '内容',
    parent_id CHAR(36) COMMENT '父评论ID',
    
    CONSTRAINT fk_tc_task FOREIGN KEY (task_id) REFERENCES ty_tasks(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_tc_user FOREIGN KEY (user_id) REFERENCES ty_users(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT fk_tc_parent FOREIGN KEY (parent_id) REFERENCES ty_task_comments(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务评论表';

-- 任务评论表索引
CREATE INDEX idx_task_comments_task ON ty_task_comments(task_id);
CREATE INDEX idx_task_comments_user ON ty_task_comments(user_id);
CREATE INDEX idx_task_comments_parent ON ty_task_comments(parent_id);
CREATE INDEX idx_task_comments_deleted_at ON ty_task_comments(deleted_at);

-- ============================================================
-- 12. 方言表 (ty_dialects)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_dialects (
    id CHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    title VARCHAR(100) NOT NULL COMMENT '标题',
    content TEXT COMMENT '内容',
    region VARCHAR(100) NOT NULL COMMENT '地区',
    province VARCHAR(50) COMMENT '省',
    city VARCHAR(50) COMMENT '市',
    dialect_type VARCHAR(20) NOT NULL DEFAULT 'phrase' COMMENT '类型: phrase-短语, story-故事, song-歌曲, daily-日常用语, other-其他',
    audio_url VARCHAR(255) NOT NULL COMMENT '音频URL',
    duration INT NOT NULL DEFAULT 0 COMMENT '时长(秒)',
    file_size INT NOT NULL DEFAULT 0 COMMENT '文件大小(字节)',
    format VARCHAR(10) COMMENT '格式',
    status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '状态: active-活跃, inactive-禁用, pending-待审核',
    is_featured TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否精选',
    play_count INT NOT NULL DEFAULT 0 COMMENT '播放次数',
    like_count INT NOT NULL DEFAULT 0 COMMENT '点赞数',
    comment_count INT NOT NULL DEFAULT 0 COMMENT '评论数',
    tags JSON COMMENT '标签JSON',
    description TEXT COMMENT '描述',
    uploader_id CHAR(36) NOT NULL COMMENT '上传人ID',
    org_id CHAR(36) NOT NULL COMMENT '组织ID',
    
    CONSTRAINT fk_dialect_uploader FOREIGN KEY (uploader_id) REFERENCES ty_users(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT fk_dialect_org FOREIGN KEY (org_id) REFERENCES ty_organizations(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT chk_dialect_type CHECK (dialect_type IN ('phrase', 'story', 'song', 'daily', 'other')),
    CONSTRAINT chk_dialect_status CHECK (status IN ('active', 'inactive', 'pending'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='方言表';

-- 方言表索引
CREATE INDEX idx_dialects_status ON ty_dialects(status);
CREATE INDEX idx_dialects_type ON ty_dialects(dialect_type);
CREATE INDEX idx_dialects_region ON ty_dialects(region);
CREATE INDEX idx_dialects_uploader ON ty_dialects(uploader_id);
CREATE INDEX idx_dialects_org ON ty_dialects(org_id);
CREATE INDEX idx_dialects_featured ON ty_dialects(is_featured);
CREATE INDEX idx_dialects_deleted_at ON ty_dialects(deleted_at);

-- ============================================================
-- 13. 方言评论表 (ty_dialect_comments)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_dialect_comments (
    id CHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    dialect_id CHAR(36) NOT NULL COMMENT '方言ID',
    user_id CHAR(36) NOT NULL COMMENT '用户ID',
    content TEXT NOT NULL COMMENT '内容',
    parent_id CHAR(36) COMMENT '父评论ID',
    reply_count INT NOT NULL DEFAULT 0 COMMENT '回复数',
    like_count INT NOT NULL DEFAULT 0 COMMENT '点赞数',
    
    CONSTRAINT fk_dc_dialect FOREIGN KEY (dialect_id) REFERENCES ty_dialects(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_dc_user FOREIGN KEY (user_id) REFERENCES ty_users(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT fk_dc_parent FOREIGN KEY (parent_id) REFERENCES ty_dialect_comments(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='方言评论表';

-- 方言评论表索引
CREATE INDEX idx_dialect_comments_dialect ON ty_dialect_comments(dialect_id);
CREATE INDEX idx_dialect_comments_user ON ty_dialect_comments(user_id);
CREATE INDEX idx_dialect_comments_parent ON ty_dialect_comments(parent_id);
CREATE INDEX idx_dialect_comments_deleted_at ON ty_dialect_comments(deleted_at);

-- ============================================================
-- 14. 方言点赞表 (ty_dialect_likes)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_dialect_likes (
    id CHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    dialect_id CHAR(36) NOT NULL COMMENT '方言ID',
    user_id CHAR(36) NOT NULL COMMENT '用户ID',
    
    UNIQUE KEY uk_dialect_user (dialect_id, user_id),
    CONSTRAINT fk_dl_dialect FOREIGN KEY (dialect_id) REFERENCES ty_dialects(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_dl_user FOREIGN KEY (user_id) REFERENCES ty_users(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='方言点赞表';

-- 方言点赞表索引
CREATE INDEX idx_dialect_likes_dialect ON ty_dialect_likes(dialect_id);
CREATE INDEX idx_dialect_likes_user ON ty_dialect_likes(user_id);

-- ============================================================
-- 15. 方言播放记录表 (ty_dialect_play_logs)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_dialect_play_logs (
    id CHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    dialect_id CHAR(36) NOT NULL COMMENT '方言ID',
    user_id CHAR(36) COMMENT '用户ID',
    ip VARCHAR(50) COMMENT 'IP地址',
    user_agent VARCHAR(255) COMMENT 'User-Agent',
    duration INT NOT NULL DEFAULT 0 COMMENT '播放时长(秒)',
    
    CONSTRAINT fk_dpl_dialect FOREIGN KEY (dialect_id) REFERENCES ty_dialects(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_dpl_user FOREIGN KEY (user_id) REFERENCES ty_users(id) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='方言播放记录表';

-- 方言播放记录表索引
CREATE INDEX idx_dialect_play_logs_dialect ON ty_dialect_play_logs(dialect_id);
CREATE INDEX idx_dialect_play_logs_user ON ty_dialect_play_logs(user_id);
CREATE INDEX idx_dialect_play_logs_created ON ty_dialect_play_logs(created_at);

-- ============================================================
-- 16. 文件表 (ty_files)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_files (
    id CHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    file_name VARCHAR(255) NOT NULL COMMENT '文件名',
    original_name VARCHAR(255) NOT NULL COMMENT '原始文件名',
    file_type VARCHAR(20) NOT NULL COMMENT '文件类型: image-图片, audio-音频, video-视频, document-文档',
    mime_type VARCHAR(100) COMMENT 'MIME类型',
    size BIGINT NOT NULL DEFAULT 0 COMMENT '文件大小',
    path VARCHAR(500) NOT NULL COMMENT '存储路径',
    url VARCHAR(500) COMMENT '访问URL',
    storage_type VARCHAR(20) NOT NULL COMMENT '存储类型: local-本地, oss-阿里云OSS, cos-腾讯云COS',
    uploader_id CHAR(36) COMMENT '上传人ID',
    entity_type VARCHAR(50) COMMENT '关联实体类型',
    entity_id CHAR(36) COMMENT '关联实体ID',
    description TEXT COMMENT '描述',
    is_deleted TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已删除',
    
    CONSTRAINT fk_file_uploader FOREIGN KEY (uploader_id) REFERENCES ty_users(id) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT chk_file_type CHECK (file_type IN ('image', 'audio', 'video', 'document')),
    CONSTRAINT chk_storage_type CHECK (storage_type IN ('local', 'oss', 'cos'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件表';

-- 文件表索引
CREATE INDEX idx_files_type ON ty_files(file_type);
CREATE INDEX idx_files_uploader ON ty_files(uploader_id);
CREATE INDEX idx_files_entity ON ty_files(entity_type, entity_id);
CREATE INDEX idx_files_deleted ON ty_files(is_deleted);
CREATE INDEX idx_files_deleted_at ON ty_files(deleted_at);

-- ============================================================
-- 17. 审计日志表 (ty_audit_logs)
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_audit_logs (
    id CHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    user_id CHAR(36) COMMENT '用户ID',
    username VARCHAR(100) COMMENT '用户名',
    user_role VARCHAR(20) COMMENT '用户角色',
    module VARCHAR(50) NOT NULL COMMENT '模块',
    action VARCHAR(50) NOT NULL COMMENT '操作',
    type VARCHAR(20) NOT NULL COMMENT '类型: login-登录, logout-登出, create-创建, update-更新, delete-删除, query-查询, export-导出, upload-上传, download-下载, other-其他',
    description TEXT COMMENT '操作描述',
    request_method VARCHAR(10) COMMENT 'HTTP方法',
    request_url VARCHAR(500) COMMENT '请求URL',
    request_ip VARCHAR(50) COMMENT '请求IP',
    request_body TEXT COMMENT '请求体',
    response_code INT COMMENT '响应状态码',
    response_body TEXT COMMENT '响应体',
    user_agent VARCHAR(500) COMMENT '用户代理',
    duration_ms BIGINT COMMENT '执行时长(毫秒)',
    status VARCHAR(20) NOT NULL COMMENT '状态: success-成功, failure-失败',
    error_message TEXT COMMENT '错误信息',
    trace_id VARCHAR(100) COMMENT '追踪ID',
    
    CONSTRAINT chk_audit_type CHECK (type IN ('login', 'logout', 'create', 'update', 'delete', 'query', 'export', 'upload', 'download', 'other')),
    CONSTRAINT chk_audit_status CHECK (status IN ('success', 'failure'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审计日志表';

-- 审计日志表索引
CREATE INDEX idx_audit_logs_user_id ON ty_audit_logs(user_id);
CREATE INDEX idx_audit_logs_module ON ty_audit_logs(module);
CREATE INDEX idx_audit_logs_type ON ty_audit_logs(type);
CREATE INDEX idx_audit_logs_status ON ty_audit_logs(status);
CREATE INDEX idx_audit_logs_created_at ON ty_audit_logs(created_at);
CREATE INDEX idx_audit_logs_trace_id ON ty_audit_logs(trace_id);
CREATE INDEX idx_audit_logs_request_ip ON ty_audit_logs(request_ip);

SET FOREIGN_KEY_CHECKS = 1;

-- ============================================================
-- 完成
-- ============================================================

-- ============================================================
-- 助力团圆志愿者系统 - 数据生命周期与约束对齐（MySQL 8+）
-- 目标：
-- 1) 统一 ty_files 软删除语义为 deleted_at
-- 2) 让 ty_users 唯一约束与软删除兼容（仅未删除记录唯一）
-- ============================================================

START TRANSACTION;

-- 1. ty_files 历史数据回填：is_deleted=1 的记录补齐 deleted_at
UPDATE ty_files
SET deleted_at = IFNULL(deleted_at, CURRENT_TIMESTAMP),
    is_deleted = 1
WHERE is_deleted = 1
   OR deleted_at IS NOT NULL;

-- 2. ty_files 状态对齐：未软删记录统一标记为 is_deleted=0
UPDATE ty_files
SET is_deleted = 0
WHERE deleted_at IS NULL;

-- 3. ty_users 唯一约束改为“仅未删除记录唯一”
ALTER TABLE ty_users
  DROP INDEX uk_user_phone,
  DROP INDEX uk_user_email,
  DROP INDEX uk_user_wx_openid;

ALTER TABLE ty_users
  ADD COLUMN active_phone VARCHAR(20)
    GENERATED ALWAYS AS (CASE WHEN deleted_at IS NULL THEN phone ELSE NULL END) STORED,
  ADD COLUMN active_email VARCHAR(100)
    GENERATED ALWAYS AS (CASE WHEN deleted_at IS NULL THEN email ELSE NULL END) STORED,
  ADD COLUMN active_wx_openid VARCHAR(100)
    GENERATED ALWAYS AS (CASE WHEN deleted_at IS NULL THEN wx_openid ELSE NULL END) STORED;

ALTER TABLE ty_users
  ADD UNIQUE KEY uk_user_phone_active (active_phone),
  ADD UNIQUE KEY uk_user_email_active (active_email),
  ADD UNIQUE KEY uk_user_wx_openid_active (active_wx_openid);

COMMIT;


-- 任务跟进工作流：跟进记录 + 评论 + 审核

CREATE TABLE IF NOT EXISTS ty_task_follow_ups (
  id CHAR(36) PRIMARY KEY,
  task_id CHAR(36) NOT NULL,
  user_id CHAR(36) NOT NULL,
  content TEXT NOT NULL,
  progress INT NULL CHECK (progress >= 0 AND progress <= 100),
  attachments JSON NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'pending',
  review_remark TEXT NULL,
  reviewed_by CHAR(36) NULL,
  reviewed_at TIMESTAMP NULL,
  comment_count INT NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  CONSTRAINT chk_task_follow_up_status CHECK (status IN ('pending', 'approved', 'rejected')),
  INDEX idx_task_follow_ups_task_id (task_id),
  INDEX idx_task_follow_ups_user_id (user_id),
  INDEX idx_task_follow_ups_status (status),
  INDEX idx_task_follow_ups_created_at (created_at),
  CONSTRAINT fk_task_follow_ups_task FOREIGN KEY (task_id) REFERENCES ty_tasks(id) ON DELETE CASCADE,
  CONSTRAINT fk_task_follow_ups_user FOREIGN KEY (user_id) REFERENCES ty_users(id),
  CONSTRAINT fk_task_follow_ups_reviewer FOREIGN KEY (reviewed_by) REFERENCES ty_users(id)
);

CREATE TABLE IF NOT EXISTS ty_task_follow_up_comments (
  id CHAR(36) PRIMARY KEY,
  task_id CHAR(36) NOT NULL,
  follow_up_id CHAR(36) NOT NULL,
  user_id CHAR(36) NOT NULL,
  content TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  INDEX idx_task_follow_up_comments_follow_up_id (follow_up_id),
  INDEX idx_task_follow_up_comments_task_id (task_id),
  INDEX idx_task_follow_up_comments_created_at (created_at),
  CONSTRAINT fk_task_follow_up_comments_task FOREIGN KEY (task_id) REFERENCES ty_tasks(id) ON DELETE CASCADE,
  CONSTRAINT fk_task_follow_up_comments_follow_up FOREIGN KEY (follow_up_id) REFERENCES ty_task_follow_ups(id) ON DELETE CASCADE,
  CONSTRAINT fk_task_follow_up_comments_user FOREIGN KEY (user_id) REFERENCES ty_users(id)
);

-- ============================================================
-- 助力团圆志愿者系统 - MySQL 种子数据
-- 版本: 1.0.0
-- 说明: 初始化超级管理员、根组织和基础权限数据
-- ============================================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================================
-- 1. 创建根组织
-- ============================================================
INSERT IGNORE INTO ty_organizations (
    id, created_at, updated_at, name, code, type, level, parent_id, 
    description, address, contact_name, contact_phone, status, logo, sort_order
) VALUES (
    '00000000-0000-0000-0000-000000000000',
    NOW(),
    NOW(),
    '助力团圆志愿者协会',
    'ROOT',
    'root',
    1,
    NULL,
    '助力团圆志愿者系统根组织，负责统筹全国志愿者工作',
    '中国',
    '系统管理员',
    '13800000000',
    'active',
    NULL,
    0
);

-- ============================================================
-- 2. 创建超级管理员用户
-- 密码: admin123 (bcrypt 加密)
-- ============================================================
INSERT IGNORE INTO ty_users (
    id, created_at, updated_at, nickname, phone, email, password, 
    role, status, org_id, avatar, real_name, gender, address, introduction
) VALUES (
    '00000000-0000-0000-0000-000000000001',
    NOW(),
    NOW(),
    '超级管理员',
    '13800138000',
    'admin@cntuanyuan.org',
    '$2a$10$0P4.9amnrR6959b1e6WeouNswfG.6rOz796S/Jm2SEOeiu7lc/duG',
    'super_admin',
    'active',
    '00000000-0000-0000-0000-000000000000',
    NULL,
    '系统管理员',
    'male',
    '中国',
    '助力团圆志愿者系统超级管理员'
);

-- ============================================================
-- 3. 初始化组织统计
-- ============================================================
INSERT IGNORE INTO ty_org_stats (
    id, created_at, updated_at, org_id, total_volunteers, active_volunteers,
    total_cases, active_cases, completed_cases, total_tasks, pending_tasks
) VALUES (
    UUID(),
    NOW(),
    NOW(),
    '00000000-0000-0000-0000-000000000000',
    1,
    1,
    0,
    0,
    0,
    0,
    0
);

-- ============================================================
-- 4. 创建基础权限数据
-- ============================================================

-- 用户管理权限
INSERT IGNORE INTO ty_permissions (id, created_at, updated_at, name, code, description, resource, action) VALUES
    (UUID(), NOW(), NOW(), '查看用户', 'user:view', '查看用户列表和详情', 'user', 'view'),
    (UUID(), NOW(), NOW(), '创建用户', 'user:create', '创建新用户', 'user', 'create'),
    (UUID(), NOW(), NOW(), '编辑用户', 'user:edit', '编辑用户信息', 'user', 'edit'),
    (UUID(), NOW(), NOW(), '删除用户', 'user:delete', '删除用户', 'user', 'delete');

-- 组织管理权限
INSERT IGNORE INTO ty_permissions (id, created_at, updated_at, name, code, description, resource, action) VALUES
    (UUID(), NOW(), NOW(), '查看组织', 'org:view', '查看组织列表和详情', 'organization', 'view'),
    (UUID(), NOW(), NOW(), '创建组织', 'org:create', '创建新组织', 'organization', 'create'),
    (UUID(), NOW(), NOW(), '编辑组织', 'org:edit', '编辑组织信息', 'organization', 'edit'),
    (UUID(), NOW(), NOW(), '删除组织', 'org:delete', '删除组织', 'organization', 'delete');

-- 走失人员管理权限
INSERT IGNORE INTO ty_permissions (id, created_at, updated_at, name, code, description, resource, action) VALUES
    (UUID(), NOW(), NOW(), '查看走失人员', 'missing:view', '查看走失人员列表和详情', 'missing_person', 'view'),
    (UUID(), NOW(), NOW(), '创建走失人员', 'missing:create', '登记走失人员', 'missing_person', 'create'),
    (UUID(), NOW(), NOW(), '编辑走失人员', 'missing:edit', '编辑走失人员信息', 'missing_person', 'edit'),
    (UUID(), NOW(), NOW(), '删除走失人员', 'missing:delete', '删除走失人员记录', 'missing_person', 'delete');

-- 任务管理权限
INSERT IGNORE INTO ty_permissions (id, created_at, updated_at, name, code, description, resource, action) VALUES
    (UUID(), NOW(), NOW(), '查看任务', 'task:view', '查看任务列表和详情', 'task', 'view'),
    (UUID(), NOW(), NOW(), '创建任务', 'task:create', '创建新任务', 'task', 'create'),
    (UUID(), NOW(), NOW(), '编辑任务', 'task:edit', '编辑任务信息', 'task', 'edit'),
    (UUID(), NOW(), NOW(), '删除任务', 'task:delete', '删除任务', 'task', 'delete'),
    (UUID(), NOW(), NOW(), '分配任务', 'task:assign', '分配任务给志愿者', 'task', 'assign');

-- 方言管理权限
INSERT IGNORE INTO ty_permissions (id, created_at, updated_at, name, code, description, resource, action) VALUES
    (UUID(), NOW(), NOW(), '查看方言', 'dialect:view', '查看方言列表和详情', 'dialect', 'view'),
    (UUID(), NOW(), NOW(), '上传方言', 'dialect:upload', '上传方言语音', 'dialect', 'upload'),
    (UUID(), NOW(), NOW(), '审核方言', 'dialect:review', '审核方言内容', 'dialect', 'review'),
    (UUID(), NOW(), NOW(), '删除方言', 'dialect:delete', '删除方言记录', 'dialect', 'delete');

-- 系统管理权限
INSERT IGNORE INTO ty_permissions (id, created_at, updated_at, name, code, description, resource, action) VALUES
    (UUID(), NOW(), NOW(), '系统设置', 'system:config', '管理系统配置', 'system', 'config'),
    (UUID(), NOW(), NOW(), '查看日志', 'system:log', '查看系统日志', 'system', 'log'),
    (UUID(), NOW(), NOW(), '数据统计', 'system:stats', '查看数据统计', 'system', 'stats');

-- ============================================================
-- 5. 为超级管理员分配所有权限
-- ============================================================
INSERT IGNORE INTO ty_user_permissions (user_id, permission_id, granted_at, granted_by)
SELECT 
    '00000000-0000-0000-0000-000000000001',
    id,
    NOW(),
    '00000000-0000-0000-0000-000000000001'
FROM ty_permissions;

SET FOREIGN_KEY_CHECKS = 1;

-- ============================================================
-- 完成
-- ============================================================
