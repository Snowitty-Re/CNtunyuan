-- ============================================================
-- Phase 2 OA 工作流引擎
-- 版本: 2.0.0
-- 说明: 工作流相关表结构
-- ============================================================

-- ============================================================
-- 1. 工作流定义表
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_workflow_definitions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    name VARCHAR(100) NOT NULL,
    key VARCHAR(50) NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    description TEXT,
    category VARCHAR(50),
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'archived')),
    start_node_id UUID,
    org_id UUID NOT NULL,
    config JSONB,
    form_schema JSONB,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    
    UNIQUE(key, version),
    CONSTRAINT fk_wfd_org FOREIGN KEY (org_id) REFERENCES ty_organizations(id) ON DELETE RESTRICT ON UPDATE CASCADE
);

COMMENT ON TABLE ty_workflow_definitions IS '工作流定义表';
COMMENT ON COLUMN ty_workflow_definitions.key IS '流程标识，如: leave_request';
COMMENT ON COLUMN ty_workflow_definitions.status IS '状态: draft-草稿, active-激活, archived-归档';

-- 索引
CREATE INDEX idx_workflow_defs_org ON ty_workflow_definitions(org_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_workflow_defs_key ON ty_workflow_definitions(key) WHERE deleted_at IS NULL;
CREATE INDEX idx_workflow_defs_status ON ty_workflow_definitions(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_workflow_defs_category ON ty_workflow_definitions(category) WHERE deleted_at IS NULL;

-- ============================================================
-- 2. 工作流节点表
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_workflow_nodes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    workflow_id UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('start', 'end', 'approval', 'task', 'branch', 'parallel', 'condition')),
    config JSONB,
    assignee_type VARCHAR(20) CHECK (assignee_type IN ('user', 'role', 'dept', 'expression', 'self', 'starter')),
    assignees JSONB,
    approval_mode VARCHAR(20) CHECK (approval_mode IN ('sequential', 'parallel')),
    required_count INTEGER NOT NULL DEFAULT 1,
    allow_transfer BOOLEAN NOT NULL DEFAULT TRUE,
    allow_delegate BOOLEAN NOT NULL DEFAULT TRUE,
    auto_pass BOOLEAN NOT NULL DEFAULT FALSE,
    conditions JSONB,
    position_x DOUBLE PRECISION DEFAULT 0,
    position_y DOUBLE PRECISION DEFAULT 0,
    order_index INTEGER NOT NULL DEFAULT 0,
    
    CONSTRAINT fk_wfn_workflow FOREIGN KEY (workflow_id) REFERENCES ty_workflow_definitions(id) ON DELETE CASCADE ON UPDATE CASCADE
);

COMMENT ON TABLE ty_workflow_nodes IS '工作流节点表';
COMMENT ON COLUMN ty_workflow_nodes.type IS '节点类型: start-开始, end-结束, approval-审批, task-任务, branch-分支, parallel-并行, condition-条件';
COMMENT ON COLUMN ty_workflow_nodes.approval_mode IS '审批模式: sequential-顺序, parallel-并行';

-- 索引
CREATE INDEX idx_workflow_nodes_workflow ON ty_workflow_nodes(workflow_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_workflow_nodes_type ON ty_workflow_nodes(type) WHERE deleted_at IS NULL;

-- ============================================================
-- 3. 流程实例表
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_workflow_instances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    definition_id UUID NOT NULL,
    business_key VARCHAR(100),
    business_id UUID,
    title VARCHAR(200) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'pending', 'processing', 'approved', 'rejected', 'cancelled', 'returned')),
    current_node_id UUID,
    started_by UUID NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    variables JSONB,
    result VARCHAR(20) CHECK (result IN ('approved', 'rejected')),
    comment TEXT,
    org_id UUID NOT NULL,
    
    CONSTRAINT fk_wfi_definition FOREIGN KEY (definition_id) REFERENCES ty_workflow_definitions(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT fk_wfi_org FOREIGN KEY (org_id) REFERENCES ty_organizations(id) ON DELETE RESTRICT ON UPDATE CASCADE
);

COMMENT ON TABLE ty_workflow_instances IS '流程实例表';
COMMENT ON COLUMN ty_workflow_instances.status IS '状态: draft-草稿, pending-待审批, processing-审批中, approved-已通过, rejected-已拒绝, cancelled-已取消, returned-已退回';
COMMENT ON COLUMN ty_workflow_instances.result IS '结果: approved-通过, rejected-拒绝';

-- 索引
CREATE INDEX idx_workflow_instances_def ON ty_workflow_instances(definition_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_workflow_instances_status ON ty_workflow_instances(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_workflow_instances_business_key ON ty_workflow_instances(business_key) WHERE deleted_at IS NULL;
CREATE INDEX idx_workflow_instances_started_by ON ty_workflow_instances(started_by) WHERE deleted_at IS NULL;
CREATE INDEX idx_workflow_instances_org ON ty_workflow_instances(org_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_workflow_instances_created ON ty_workflow_instances(created_at) WHERE deleted_at IS NULL;

-- 复合索引
CREATE INDEX idx_workflow_instances_org_status ON ty_workflow_instances(org_id, status) WHERE deleted_at IS NULL;

-- ============================================================
-- 4. 工作流任务表（增强任务表）
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_workflow_tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- 基础任务字段
    title VARCHAR(200) NOT NULL,
    description TEXT,
    type VARCHAR(20) NOT NULL DEFAULT 'verify',
    priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    creator_id UUID NOT NULL,
    assignee_id UUID,
    org_id UUID NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    result VARCHAR(20),
    
    -- 工作流特有字段
    workflow_instance_id UUID,
    workflow_node_id UUID,
    workflow_node_type VARCHAR(20),
    approval_action VARCHAR(20),
    approval_comment TEXT,
    due_time TIMESTAMP WITH TIME ZONE,
    reminded_at TIMESTAMP WITH TIME ZONE,
    remind_count INTEGER NOT NULL DEFAULT 0,
    delegate_from UUID,
    transferred_from UUID,
    sequential_index INTEGER NOT NULL DEFAULT 0,
    
    CONSTRAINT fk_wft_instance FOREIGN KEY (workflow_instance_id) REFERENCES ty_workflow_instances(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_wft_org FOREIGN KEY (org_id) REFERENCES ty_organizations(id) ON DELETE RESTRICT ON UPDATE CASCADE
);

COMMENT ON TABLE ty_workflow_tasks IS '工作流任务表';

-- 索引
CREATE INDEX idx_workflow_tasks_instance ON ty_workflow_tasks(workflow_instance_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_workflow_tasks_node ON ty_workflow_tasks(workflow_node_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_workflow_tasks_assignee ON ty_workflow_tasks(assignee_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_workflow_tasks_status ON ty_workflow_tasks(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_workflow_tasks_due ON ty_workflow_tasks(due_time) WHERE deleted_at IS NULL AND due_time IS NOT NULL;

-- 复合索引
CREATE INDEX idx_workflow_tasks_assignee_status ON ty_workflow_tasks(assignee_id, status) WHERE deleted_at IS NULL;

-- ============================================================
-- 5. 流程转换记录表
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_workflow_transitions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    instance_id UUID NOT NULL,
    from_node_id UUID,
    to_node_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    user_id UUID NOT NULL,
    comment TEXT,
    variables JSONB,
    
    CONSTRAINT fk_wftt_instance FOREIGN KEY (instance_id) REFERENCES ty_workflow_instances(id) ON DELETE CASCADE ON UPDATE CASCADE
);

COMMENT ON TABLE ty_workflow_transitions IS '流程转换记录表';

-- 索引
CREATE INDEX idx_workflow_transitions_instance ON ty_workflow_transitions(instance_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_workflow_transitions_user ON ty_workflow_transitions(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_workflow_transitions_created ON ty_workflow_transitions(created_at) WHERE deleted_at IS NULL;

-- ============================================================
-- 6. 任务委托表
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_workflow_delegations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    task_id UUID NOT NULL,
    from_user_id UUID NOT NULL,
    to_user_id UUID NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    reason TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'cancelled', 'expired')),
    
    CONSTRAINT fk_wfd_task FOREIGN KEY (task_id) REFERENCES ty_workflow_tasks(id) ON DELETE CASCADE ON UPDATE CASCADE
);

COMMENT ON TABLE ty_workflow_delegations IS '任务委托表';

-- 索引
CREATE INDEX idx_workflow_delegations_task ON ty_workflow_delegations(task_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_workflow_delegations_from ON ty_workflow_delegations(from_user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_workflow_delegations_to ON ty_workflow_delegations(to_user_id) WHERE deleted_at IS NULL;

-- ============================================================
-- 7. 任务催办表
-- ============================================================
CREATE TABLE IF NOT EXISTS ty_workflow_reminders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    task_id UUID NOT NULL,
    reminder_type VARCHAR(20) NOT NULL DEFAULT 'system' CHECK (reminder_type IN ('system', 'manual')),
    remind_count INTEGER NOT NULL DEFAULT 0,
    last_remind_at TIMESTAMP WITH TIME ZONE,
    next_remind_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed')),
    
    CONSTRAINT fk_wfr_task FOREIGN KEY (task_id) REFERENCES ty_workflow_tasks(id) ON DELETE CASCADE ON UPDATE CASCADE
);

COMMENT ON TABLE ty_workflow_reminders IS '任务催办表';

-- 索引
CREATE INDEX idx_workflow_reminders_task ON ty_workflow_reminders(task_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_workflow_reminders_next ON ty_workflow_reminders(next_remind_at) WHERE deleted_at IS NULL AND next_remind_at IS NOT NULL;

-- ============================================================
-- 8. 创建更新时间触发器
-- ============================================================
CREATE TRIGGER update_workflow_definitions_updated_at BEFORE UPDATE ON ty_workflow_definitions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_workflow_nodes_updated_at BEFORE UPDATE ON ty_workflow_nodes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_workflow_instances_updated_at BEFORE UPDATE ON ty_workflow_instances
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_workflow_tasks_updated_at BEFORE UPDATE ON ty_workflow_tasks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_workflow_transitions_updated_at BEFORE UPDATE ON ty_workflow_transitions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_workflow_delegations_updated_at BEFORE UPDATE ON ty_workflow_delegations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_workflow_reminders_updated_at BEFORE UPDATE ON ty_workflow_reminders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- 9. 插入示例工作流定义
-- ============================================================
-- 请假流程示例
INSERT INTO ty_workflow_definitions (id, name, key, version, description, category, status, org_id, is_system)
VALUES (
    '00000000-0000-0000-0000-000000000100',
    '请假审批流程',
    'leave_request',
    1,
    '员工请假申请审批流程',
    'hr',
    'active',
    '00000000-0000-0000-0000-000000000000',
    true
)
ON CONFLICT (key, version) DO NOTHING;

-- ============================================================
-- 完成
-- ============================================================
