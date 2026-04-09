BEGIN;

CREATE TABLE IF NOT EXISTS ty_system_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    category VARCHAR(50) NOT NULL,
    setting_key VARCHAR(120) NOT NULL,
    setting_value TEXT NOT NULL,
    value_type VARCHAR(20) NOT NULL DEFAULT 'string',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_system_settings_key_active ON ty_system_settings(setting_key) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_system_settings_category_active ON ty_system_settings(category) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_system_settings_deleted_at ON ty_system_settings(deleted_at) WHERE deleted_at IS NOT NULL;

DROP TRIGGER IF EXISTS update_system_settings_updated_at ON ty_system_settings;
CREATE TRIGGER update_system_settings_updated_at BEFORE UPDATE ON ty_system_settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMIT;
