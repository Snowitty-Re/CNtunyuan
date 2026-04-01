-- 任务跟进工作流：跟进记录 + 评论 + 审核

CREATE TABLE IF NOT EXISTS ty_task_follow_ups (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  task_id UUID NOT NULL REFERENCES ty_tasks(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES ty_users(id),
  content TEXT NOT NULL,
  progress INT NULL CHECK (progress >= 0 AND progress <= 100),
  attachments JSONB NULL DEFAULT '[]'::jsonb,
  status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
  review_remark TEXT NULL,
  reviewed_by UUID NULL REFERENCES ty_users(id),
  reviewed_at TIMESTAMPTZ NULL,
  comment_count INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_task_follow_ups_task_id ON ty_task_follow_ups(task_id);
CREATE INDEX IF NOT EXISTS idx_task_follow_ups_user_id ON ty_task_follow_ups(user_id);
CREATE INDEX IF NOT EXISTS idx_task_follow_ups_status ON ty_task_follow_ups(status);
CREATE INDEX IF NOT EXISTS idx_task_follow_ups_created_at ON ty_task_follow_ups(created_at DESC);

CREATE TABLE IF NOT EXISTS ty_task_follow_up_comments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  task_id UUID NOT NULL REFERENCES ty_tasks(id) ON DELETE CASCADE,
  follow_up_id UUID NOT NULL REFERENCES ty_task_follow_ups(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES ty_users(id),
  content TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_task_follow_up_comments_follow_up_id ON ty_task_follow_up_comments(follow_up_id);
CREATE INDEX IF NOT EXISTS idx_task_follow_up_comments_task_id ON ty_task_follow_up_comments(task_id);
CREATE INDEX IF NOT EXISTS idx_task_follow_up_comments_created_at ON ty_task_follow_up_comments(created_at DESC);
