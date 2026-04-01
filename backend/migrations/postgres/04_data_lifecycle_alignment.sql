-- ============================================================
-- 团圆寻亲志愿者系统 - 数据生命周期与约束对齐（PostgreSQL）
-- 目标：
-- 1) 统一 ty_files 软删除语义为 deleted_at
-- 2) 让 ty_users 唯一约束与软删除兼容（仅未删除记录唯一）
-- ============================================================

BEGIN;

-- 1. ty_files 历史数据回填：is_deleted=true 的记录补齐 deleted_at
UPDATE ty_files
SET deleted_at = COALESCE(deleted_at, CURRENT_TIMESTAMP),
    is_deleted = TRUE
WHERE is_deleted = TRUE
   OR deleted_at IS NOT NULL;

-- 2. ty_files 状态对齐：未软删记录统一标记为 is_deleted=false
UPDATE ty_files
SET is_deleted = FALSE
WHERE deleted_at IS NULL;

-- 3. ty_users 唯一约束改为“仅未删除记录唯一”
ALTER TABLE ty_users DROP CONSTRAINT IF EXISTS ty_users_phone_key;
ALTER TABLE ty_users DROP CONSTRAINT IF EXISTS ty_users_email_key;
ALTER TABLE ty_users DROP CONSTRAINT IF EXISTS ty_users_wx_openid_key;

CREATE UNIQUE INDEX IF NOT EXISTS uk_users_phone_active
ON ty_users(phone)
WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uk_users_email_active
ON ty_users(email)
WHERE deleted_at IS NULL AND email IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uk_users_wx_openid_active
ON ty_users(wx_openid)
WHERE deleted_at IS NULL AND wx_openid IS NOT NULL;

COMMIT;

