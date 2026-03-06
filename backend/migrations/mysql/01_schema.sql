-- ============================================================
-- 团圆寻亲志愿者系统 - MySQL 数据库结构
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
    
    name VARCHAR(50) NOT NULL COMMENT '姓名',
    gender VARCHAR(10) NOT NULL COMMENT '性别',
    birth_date DATE COMMENT '出生日期',
    age INT COMMENT '年龄',
    height INT COMMENT '身高(cm)',
    weight INT COMMENT '体重(kg)',
    description TEXT COMMENT '描述',
    photo_url VARCHAR(255) COMMENT '照片URL',
    
    missing_time TIMESTAMP NOT NULL COMMENT '走失时间',
    province VARCHAR(50) COMMENT '省',
    city VARCHAR(50) COMMENT '市',
    district VARCHAR(50) COMMENT '区',
    address VARCHAR(255) COMMENT '详细地址',
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
    CONSTRAINT chk_mp_urgency CHECK (urgency IN ('critical', 'high', 'medium', 'low'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='走失人员表';

-- 走失人员表索引
CREATE INDEX idx_missing_persons_status ON ty_missing_persons(status);
CREATE INDEX idx_missing_persons_urgency ON ty_missing_persons(urgency);
CREATE INDEX idx_missing_persons_reporter ON ty_missing_persons(reporter_id);
CREATE INDEX idx_missing_persons_org ON ty_missing_persons(org_id);
CREATE INDEX idx_missing_persons_assigned ON ty_missing_persons(assigned_to);
CREATE INDEX idx_missing_persons_missing_time ON ty_missing_persons(missing_time);
CREATE INDEX idx_missing_persons_location ON ty_missing_persons(province, city, district);
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
-- 8. 任务表 (ty_tasks)
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

SET FOREIGN_KEY_CHECKS = 1;

-- ============================================================
-- 完成
-- ============================================================
