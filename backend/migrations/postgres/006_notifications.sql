-- Migration: Notification System
-- Date: 2026-03-07
-- Description: Add notification system with WebSocket support

-- ============================================
-- 1. Notifications Table
-- ============================================
CREATE TABLE IF NOT EXISTS ty_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(200) NOT NULL,
    content TEXT,
    type VARCHAR(20) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    priority VARCHAR(10) NOT NULL DEFAULT 'normal',
    status VARCHAR(20) NOT NULL DEFAULT 'unread',
    to_user_id UUID NOT NULL REFERENCES ty_users(id) ON DELETE CASCADE,
    from_user_id UUID REFERENCES ty_users(id) ON DELETE SET NULL,
    org_id UUID NOT NULL REFERENCES ty_organizations(id) ON DELETE CASCADE,
    
    -- 业务关联
    business_type VARCHAR(50),
    business_id UUID,
    
    -- 扩展数据
    data JSONB,
    action_url VARCHAR(500),
    action_text VARCHAR(50),
    
    -- 时间追踪
    read_at TIMESTAMP WITH TIME ZONE,
    archived_at TIMESTAMP WITH TIME ZONE,
    expire_at TIMESTAMP WITH TIME ZONE,
    
    -- 发送追踪
    sent_at TIMESTAMP WITH TIME ZONE,
    sent_success BOOLEAN NOT NULL DEFAULT FALSE,
    error_msg VARCHAR(500),
    retry_count INTEGER NOT NULL DEFAULT 0,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for notifications
CREATE INDEX IF NOT EXISTS idx_notifications_to_user_id ON ty_notifications(to_user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_org_id ON ty_notifications(org_id);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON ty_notifications(type);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON ty_notifications(status);
CREATE INDEX IF NOT EXISTS idx_notifications_business ON ty_notifications(business_type, business_id);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON ty_notifications(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_unread ON ty_notifications(to_user_id, status) WHERE status = 'unread';
CREATE INDEX IF NOT EXISTS idx_notifications_deleted_at ON ty_notifications(deleted_at) WHERE deleted_at IS NULL;

-- ============================================
-- 2. Notification Settings Table
-- ============================================
CREATE TABLE IF NOT EXISTS ty_notification_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES ty_users(id) ON DELETE CASCADE,
    org_id UUID NOT NULL REFERENCES ty_organizations(id) ON DELETE CASCADE,
    
    -- 渠道开关
    web_socket_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    push_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sms_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    email_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    in_app_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- 类型开关
    system_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    task_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    workflow_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    alert_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- 免打扰设置
    do_not_disturb BOOLEAN NOT NULL DEFAULT FALSE,
    do_not_disturb_from TIME,
    do_not_disturb_to TIME,
    
    -- 摘要设置
    daily_digest_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    weekly_digest_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    UNIQUE(user_id)
);

-- Indexes for notification settings
CREATE INDEX IF NOT EXISTS idx_notification_settings_user_id ON ty_notification_settings(user_id);
CREATE INDEX IF NOT EXISTS idx_notification_settings_org_id ON ty_notification_settings(org_id);

-- ============================================
-- 3. Message Templates Table
-- ============================================
CREATE TABLE IF NOT EXISTS ty_message_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    type VARCHAR(20) NOT NULL,
    
    -- 模板内容
    subject VARCHAR(200),
    content TEXT NOT NULL,
    content_sms VARCHAR(500),
    
    -- 变量定义
    variables JSONB,
    variables_desc JSONB,
    
    -- 模板状态
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    version INTEGER NOT NULL DEFAULT 1,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- 示例数据
    example_data JSONB,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    UNIQUE(code, channel)
);

-- Indexes for message templates
CREATE INDEX IF NOT EXISTS idx_message_templates_code ON ty_message_templates(code);
CREATE INDEX IF NOT EXISTS idx_message_templates_channel ON ty_message_templates(channel);
CREATE INDEX IF NOT EXISTS idx_message_templates_type ON ty_message_templates(type);
CREATE INDEX IF NOT EXISTS idx_message_templates_status ON ty_message_templates(status);
CREATE INDEX IF NOT EXISTS idx_message_templates_deleted_at ON ty_message_templates(deleted_at) WHERE deleted_at IS NULL;

-- ============================================
-- 4. Insert Default Message Templates
-- ============================================

-- Task Assigned Template
INSERT INTO ty_message_templates (code, name, channel, type, subject, content, variables, variables_desc, is_system)
VALUES (
    'task_assigned',
    '任务分配通知',
    'websocket',
    'task',
    '新任务分配',
    '您有一个新的{{.TaskType}}任务「{{.TaskTitle}}」需要处理，截止日期：{{.Deadline}}',
    '{"task_type": "string", "task_title": "string", "deadline": "string"}'::jsonb,
    '{"task_type": "任务类型", "task_title": "任务标题", "deadline": "截止日期"}'::jsonb,
    TRUE
)
ON CONFLICT (code, channel) DO NOTHING;

-- Task Reminder Template
INSERT INTO ty_message_templates (code, name, channel, type, subject, content, variables, variables_desc, is_system)
VALUES (
    'task_reminder',
    '任务提醒',
    'websocket',
    'task',
    '任务即将到期',
    '您的任务「{{.TaskTitle}}」即将到期，请尽快处理',
    '{"task_title": "string"}'::jsonb,
    '{"task_title": "任务标题"}'::jsonb,
    TRUE
)
ON CONFLICT (code, channel) DO NOTHING;

-- Workflow Approved Template
INSERT INTO ty_message_templates (code, name, channel, type, subject, content, variables, variables_desc, is_system)
VALUES (
    'workflow_approved',
    '审批通过通知',
    'websocket',
    'workflow',
    '审批通过',
    '您的「{{.WorkflowName}}」申请已通过{{.ApprovedBy}}的审批',
    '{"workflow_name": "string", "approved_by": "string"}'::jsonb,
    '{"workflow_name": "流程名称", "approved_by": "审批人"}'::jsonb,
    TRUE
)
ON CONFLICT (code, channel) DO NOTHING;

-- Workflow Rejected Template
INSERT INTO ty_message_templates (code, name, channel, type, subject, content, variables, variables_desc, is_system)
VALUES (
    'workflow_rejected',
    '审批拒绝通知',
    'websocket',
    'workflow',
    '审批被拒绝',
    '您的「{{.WorkflowName}}」申请被{{.RejectedBy}}拒绝，原因：{{.Reason}}',
    '{"workflow_name": "string", "rejected_by": "string", "reason": "string"}'::jsonb,
    '{"workflow_name": "流程名称", "rejected_by": "审批人", "reason": "拒绝原因"}'::jsonb,
    TRUE
)
ON CONFLICT (code, channel) DO NOTHING;

-- Workflow Pending Template
INSERT INTO ty_message_templates (code, name, channel, type, subject, content, variables, variables_desc, is_system)
VALUES (
    'workflow_pending',
    '待审批通知',
    'websocket',
    'workflow',
    '有新的审批待处理',
    '{{.Submitter}}提交了「{{.WorkflowName}}」，需要您的审批',
    '{"submitter": "string", "workflow_name": "string"}'::jsonb,
    '{"submitter": "提交人", "workflow_name": "流程名称"}'::jsonb,
    TRUE
)
ON CONFLICT (code, channel) DO NOTHING;

-- System Announcement Template
INSERT INTO ty_message_templates (code, name, channel, type, subject, content, variables, variables_desc, is_system)
VALUES (
    'system_announcement',
    '系统公告',
    'websocket',
    'system',
    '{{.Title}}',
    '{{.Content}}',
    '{"title": "string", "content": "string"}'::jsonb,
    '{"title": "公告标题", "content": "公告内容"}'::jsonb,
    TRUE
)
ON CONFLICT (code, channel) DO NOTHING;

-- Alert Template
INSERT INTO ty_message_templates (code, name, channel, type, subject, content, variables, variables_desc, is_system)
VALUES (
    'system_alert',
    '系统告警',
    'websocket',
    'alert',
    '{{.AlertTitle}}',
    '{{.AlertContent}}',
    '{"alert_title": "string", "alert_content": "string"}'::jsonb,
    '{"alert_title": "告警标题", "alert_content": "告警内容"}'::jsonb,
    TRUE
)
ON CONFLICT (code, channel) DO NOTHING;

-- SMS Templates
INSERT INTO ty_message_templates (code, name, channel, type, content, content_sms, is_system)
VALUES (
    'verify_code',
    '验证码',
    'sms',
    'system',
    '您的验证码是：{{.Code}}，{{.Expire}}分钟内有效。如非本人操作，请忽略。',
    '您的验证码是：{{.Code}}，{{.Expire}}分钟内有效。',
    TRUE
)
ON CONFLICT (code, channel) DO NOTHING;

-- Email Templates
INSERT INTO ty_message_templates (code, name, channel, type, subject, content, is_system)
VALUES (
    'welcome_email',
    '欢迎邮件',
    'email',
    'system',
    '欢迎使用团圆寻亲志愿者系统',
    '<h1>欢迎，{{.UserName}}！</h1><p>感谢您注册团圆寻亲志愿者系统...</p>',
    TRUE
)
ON CONFLICT (code, channel) DO NOTHING;

-- ============================================
-- 5. Create Updated At Triggers
-- ============================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ty_notifications trigger
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_ty_notifications_updated_at') THEN
        CREATE TRIGGER update_ty_notifications_updated_at
            BEFORE UPDATE ON ty_notifications
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;

-- ty_notification_settings trigger
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_ty_notification_settings_updated_at') THEN
        CREATE TRIGGER update_ty_notification_settings_updated_at
            BEFORE UPDATE ON ty_notification_settings
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;

-- ty_message_templates trigger
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_ty_message_templates_updated_at') THEN
        CREATE TRIGGER update_ty_message_templates_updated_at
            BEFORE UPDATE ON ty_message_templates
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;

-- ============================================
-- Migration Complete
-- ============================================
