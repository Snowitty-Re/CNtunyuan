CREATE TABLE IF NOT EXISTS ty_system_settings (
    id CHAR(36) PRIMARY KEY,
    category VARCHAR(50) NOT NULL,
    setting_key VARCHAR(120) NOT NULL,
    setting_value TEXT NOT NULL,
    value_type VARCHAR(20) NOT NULL DEFAULT 'string',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    UNIQUE KEY uk_system_settings_key_active (setting_key, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_system_settings_category_active ON ty_system_settings(category, deleted_at);
CREATE INDEX idx_system_settings_deleted_at ON ty_system_settings(deleted_at);
