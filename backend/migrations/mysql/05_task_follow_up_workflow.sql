-- 任务跟进工作流：跟进记录 + 评论 + 审核

CREATE TABLE IF NOT EXISTS ty_task_follow_ups (
  id CHAR(36) PRIMARY KEY,
  task_id CHAR(36) NOT NULL,
  user_id CHAR(36) NOT NULL,
  content TEXT NOT NULL,
  progress INT NULL CHECK (progress >= 0 AND progress <= 100),
  attachments JSON NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'pending',
  review_remark TEXT NULL,
  reviewed_by CHAR(36) NULL,
  reviewed_at TIMESTAMP NULL,
  comment_count INT NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  CONSTRAINT chk_task_follow_up_status CHECK (status IN ('pending', 'approved', 'rejected')),
  INDEX idx_task_follow_ups_task_id (task_id),
  INDEX idx_task_follow_ups_user_id (user_id),
  INDEX idx_task_follow_ups_status (status),
  INDEX idx_task_follow_ups_created_at (created_at),
  CONSTRAINT fk_task_follow_ups_task FOREIGN KEY (task_id) REFERENCES ty_tasks(id) ON DELETE CASCADE,
  CONSTRAINT fk_task_follow_ups_user FOREIGN KEY (user_id) REFERENCES ty_users(id),
  CONSTRAINT fk_task_follow_ups_reviewer FOREIGN KEY (reviewed_by) REFERENCES ty_users(id)
);

CREATE TABLE IF NOT EXISTS ty_task_follow_up_comments (
  id CHAR(36) PRIMARY KEY,
  task_id CHAR(36) NOT NULL,
  follow_up_id CHAR(36) NOT NULL,
  user_id CHAR(36) NOT NULL,
  content TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  INDEX idx_task_follow_up_comments_follow_up_id (follow_up_id),
  INDEX idx_task_follow_up_comments_task_id (task_id),
  INDEX idx_task_follow_up_comments_created_at (created_at),
  CONSTRAINT fk_task_follow_up_comments_task FOREIGN KEY (task_id) REFERENCES ty_tasks(id) ON DELETE CASCADE,
  CONSTRAINT fk_task_follow_up_comments_follow_up FOREIGN KEY (follow_up_id) REFERENCES ty_task_follow_ups(id) ON DELETE CASCADE,
  CONSTRAINT fk_task_follow_up_comments_user FOREIGN KEY (user_id) REFERENCES ty_users(id)
);
