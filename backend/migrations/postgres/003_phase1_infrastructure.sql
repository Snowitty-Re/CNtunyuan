-- ============================================================
-- Phase 1 基础设施强化
-- 版本: 2.0.0
-- 说明: 审计日志表、数据权限索引等
-- ============================================================

-- ============================================================
-- 1. 审计日志表
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- 用户信息
    user_id UUID NOT NULL,
    username VARCHAR(100),
    org_id UUID NOT NULL,
    
    -- 操作信息
    action VARCHAR(20) NOT NULL CHECK (action IN ('CREATE', 'UPDATE', 'DELETE', 'LOGIN', 'LOGOUT', 'QUERY', 'EXPORT', 'IMPORT', 'APPROVE', 'REJECT', 'OTHER')),
    resource VARCHAR(50) NOT NULL,
    resource_id UUID,
    resource_name VARCHAR(200),
    description TEXT,
    
    -- 数据变更
    old_values JSONB,
    new_values JSONB,
    delta JSONB,
    
    -- 请求信息
    ip_address VARCHAR(50),
    user_agent TEXT,
    request_url TEXT,
    request_method VARCHAR(10),
    trace_id VARCHAR(50),
    status INTEGER DEFAULT 200,
    duration BIGINT,
    error TEXT,
    extra JSONB
);

-- 审计日志表索引
CREATE INDEX idx_audit_logs_user_id ON ty_audit_logs(user_id);
CREATE INDEX idx_audit_logs_org_id ON ty_audit_logs(org_id);
CREATE INDEX idx_audit_logs_action ON ty_audit_logs(action);
CREATE INDEX idx_audit_logs_resource ON ty_audit_logs(resource);
CREATE INDEX idx_audit_logs_resource_id ON ty_audit_logs(resource_id);
CREATE INDEX idx_audit_logs_created_at ON ty_audit_logs(created_at);
CREATE INDEX idx_audit_logs_trace_id ON ty_audit_logs(trace_id);
CREATE INDEX idx_audit_logs_ip_address ON ty_audit_logs(ip_address);

-- 复合索引（常用查询优化）
CREATE INDEX idx_audit_logs_user_time ON ty_audit_logs(user_id, created_at);
CREATE INDEX idx_audit_logs_org_time ON ty_audit_logs(org_id, created_at);
CREATE INDEX idx_audit_logs_resource_time ON ty_audit_logs(resource, created_at);

-- ============================================================
-- 2. 组织层级索引优化（用于数据权限）
-- ============================================================
CREATE INDEX idx_organizations_parent_path ON ty_organizations(parent_id, level);

-- ============================================================
-- 3. 用户角色索引优化
-- ============================================================
CREATE INDEX idx_users_org_role ON ty_users(org_id, role);

-- ============================================================
-- 4. 任务状态索引优化
-- ============================================================
CREATE INDEX idx_tasks_status_assignee ON ty_tasks(status, assignee_id);
CREATE INDEX idx_tasks_status_org ON ty_tasks(status, org_id);
CREATE INDEX idx_tasks_deadline_status ON ty_tasks(deadline, status) WHERE status NOT IN ('completed', 'cancelled');

-- ============================================================
-- 5. 创建审计日志分区表（按月份分区）
-- ============================================================
-- 注意：分区表功能需要 PostgreSQL 10+
-- 如果需要，可以取消下面的注释启用分区

/*
-- 创建分区表
CREATE TABLE ty_audit_logs_partitioned (
    id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    user_id UUID NOT NULL,
    username VARCHAR(100),
    org_id UUID NOT NULL,
    action VARCHAR(20) NOT NULL,
    resource VARCHAR(50) NOT NULL,
    resource_id UUID,
    resource_name VARCHAR(200),
    description TEXT,
    old_values JSONB,
    new_values JSONB,
    delta JSONB,
    ip_address VARCHAR(50),
    user_agent TEXT,
    request_url TEXT,
    request_method VARCHAR(10),
    trace_id VARCHAR(50),
    status INTEGER DEFAULT 200,
    duration BIGINT,
    error TEXT,
    extra JSONB
) PARTITION BY RANGE (created_at);

-- 创建分区（按月）
CREATE TABLE ty_audit_logs_2024_01 PARTITION OF ty_audit_logs_partitioned
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
CREATE TABLE ty_audit_logs_2024_02 PARTITION OF ty_audit_logs_partitioned
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');
-- 更多分区...
*/

-- ============================================================
-- 6. 创建审计日志清理函数
-- ============================================================
CREATE OR REPLACE FUNCTION cleanup_old_audit_logs(retention_days INTEGER)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
    cutoff_date TIMESTAMP WITH TIME ZONE;
BEGIN
    cutoff_date := CURRENT_TIMESTAMP - (retention_days || ' days')::INTERVAL;
    
    DELETE FROM ty_audit_logs
    WHERE created_at < cutoff_date;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- 7. 表注释
-- ============================================================
COMMENT ON TABLE ty_audit_logs IS '审计日志表，记录所有关键操作';
COMMENT ON COLUMN ty_audit_logs.action IS '操作类型: CREATE-创建, UPDATE-更新, DELETE-删除, LOGIN-登录, LOGOUT-登出, QUERY-查询, EXPORT-导出, IMPORT-导入, APPROVE-审批通过, REJECT-审批拒绝, OTHER-其他';
COMMENT ON COLUMN ty_audit_logs.delta IS '变更字段对比，记录old和new值';
COMMENT ON COLUMN ty_audit_logs.trace_id IS '请求追踪ID，用于关联请求链路';
COMMENT ON COLUMN ty_audit_logs.duration IS '操作执行时间（毫秒）';

-- ============================================================
-- 完成
-- ============================================================
