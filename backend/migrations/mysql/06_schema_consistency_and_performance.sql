-- ============================================================
-- 助力团圆志愿者系统 - 结构一致性与性能优化（MySQL 8+）
-- 目标：
-- 1) 文件软删除语义统一为 deleted_at
-- 2) 补齐经纬度约束
-- 3) 补齐审计日志 user_id 外键
-- 4) 补齐高频查询复合索引
-- ============================================================

START TRANSACTION;

-- 1) 统一 ty_files 软删除语义
ALTER TABLE ty_files DROP COLUMN IF EXISTS is_deleted;

-- 2) 经纬度约束补齐（轨迹）
ALTER TABLE ty_missing_person_tracks
  ADD CONSTRAINT chk_mpt_lat CHECK (lat IS NULL OR (lat BETWEEN -90 AND 90)),
  ADD CONSTRAINT chk_mpt_lng CHECK (lng IS NULL OR (lng BETWEEN -180 AND 180));

-- 3) 经纬度约束补齐（任务）
ALTER TABLE ty_tasks
  ADD CONSTRAINT chk_task_lat CHECK (lat IS NULL OR (lat BETWEEN -90 AND 90)),
  ADD CONSTRAINT chk_task_lng CHECK (lng IS NULL OR (lng BETWEEN -180 AND 180));

-- 4) 审计日志用户外键
ALTER TABLE ty_audit_logs
  ADD CONSTRAINT fk_audit_user FOREIGN KEY (user_id) REFERENCES ty_users(id) ON DELETE SET NULL ON UPDATE CASCADE;

-- 5) 高频查询复合索引
CREATE INDEX idx_users_org_status_created_at ON ty_users(org_id, status, created_at);
CREATE INDEX idx_users_role_status_created_at ON ty_users(role, status, created_at);
CREATE INDEX idx_missing_persons_status_urgency_created_at ON ty_missing_persons(status, urgency, created_at);
CREATE INDEX idx_tasks_status_created_at ON ty_tasks(status, created_at);
CREATE INDEX idx_tasks_org_status_created_at ON ty_tasks(org_id, status, created_at);
CREATE INDEX idx_tasks_assignee_status_created_at ON ty_tasks(assignee_id, status, created_at);

COMMIT;
