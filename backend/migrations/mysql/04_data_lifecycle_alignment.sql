-- ============================================================
-- 团圆寻亲志愿者系统 - 数据生命周期与约束对齐（MySQL 8+）
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

