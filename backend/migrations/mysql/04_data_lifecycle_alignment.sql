-- ============================================================
-- 助力团圆志愿者系统 - 数据生命周期与约束对齐（MySQL 8+）
-- 目标：
-- 1) 统一 ty_files 软删除语义为 deleted_at
-- 2) 让 ty_users 唯一约束与软删除兼容（仅未删除记录唯一）
-- ============================================================

START TRANSACTION;

-- 1. ty_files 软删除统一使用 deleted_at
ALTER TABLE ty_files DROP COLUMN IF EXISTS is_deleted;

-- 2. ty_users 唯一约束改为“仅未删除记录唯一”
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
