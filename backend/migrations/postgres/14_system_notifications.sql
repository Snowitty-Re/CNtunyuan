-- 14_system_notifications.sql
-- 站内通知中心（权限审批等系统事件）

CREATE TABLE IF NOT EXISTS ty_system_notifications (
    id UUID PRIMARY KEY,
    category VARCHAR(40) NOT NULL DEFAULT 'authz',
    title VARCHAR(200) NOT NULL,
    content TEXT,
    recipient_id UUID,
    recipient_role VARCHAR(20),
    status VARCHAR(20) NOT NULL DEFAULT 'unread',
    related_type VARCHAR(60),
    related_id VARCHAR(80),
    operator_id UUID,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_system_notifications_status CHECK (status IN ('unread', 'read'))
);

CREATE INDEX IF NOT EXISTS idx_system_notifications_recipient_id_status
    ON ty_system_notifications(recipient_id, status)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_system_notifications_recipient_role_status
    ON ty_system_notifications(recipient_role, status)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_system_notifications_category_created
    ON ty_system_notifications(category, created_at DESC)
    WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS update_system_notifications_updated_at ON ty_system_notifications;
CREATE TRIGGER update_system_notifications_updated_at BEFORE UPDATE ON ty_system_notifications
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
