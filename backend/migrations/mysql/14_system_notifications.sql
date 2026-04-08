-- 14_system_notifications.sql
-- 站内通知中心（权限审批等系统事件）

CREATE TABLE IF NOT EXISTS ty_system_notifications (
    id CHAR(36) PRIMARY KEY,
    category VARCHAR(40) NOT NULL DEFAULT 'authz',
    title VARCHAR(200) NOT NULL,
    content TEXT,
    recipient_id CHAR(36) NULL,
    recipient_role VARCHAR(20) NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'unread',
    related_type VARCHAR(60) NULL,
    related_id VARCHAR(80) NULL,
    operator_id CHAR(36) NULL,
    read_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    CONSTRAINT chk_system_notifications_status CHECK (status IN ('unread', 'read'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_system_notifications_recipient_id_status ON ty_system_notifications(recipient_id, status, deleted_at);
CREATE INDEX idx_system_notifications_recipient_role_status ON ty_system_notifications(recipient_role, status, deleted_at);
CREATE INDEX idx_system_notifications_category_created ON ty_system_notifications(category, created_at, deleted_at);
