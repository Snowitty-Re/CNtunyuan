package entity

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// WorkflowStatus 工作流定义状态
type WorkflowStatus string

const (
	WorkflowStatusDraft     WorkflowStatus = "draft"
	WorkflowStatusActive    WorkflowStatus = "active"
	WorkflowStatusArchived  WorkflowStatus = "archived"
)

// WorkflowInstanceStatus 流程实例状态
type WorkflowInstanceStatus string

const (
	InstanceStatusDraft      WorkflowInstanceStatus = "draft"
	InstanceStatusPending    WorkflowInstanceStatus = "pending"
	InstanceStatusProcessing WorkflowInstanceStatus = "processing"
	InstanceStatusApproved   WorkflowInstanceStatus = "approved"
	InstanceStatusRejected   WorkflowInstanceStatus = "rejected"
	InstanceStatusCancelled  WorkflowInstanceStatus = "cancelled"
	InstanceStatusReturned   WorkflowInstanceStatus = "returned"
)

// WorkflowNodeType 节点类型
type WorkflowNodeType string

const (
	NodeTypeStart       WorkflowNodeType = "start"
	NodeTypeEnd         WorkflowNodeType = "end"
	NodeTypeApproval    WorkflowNodeType = "approval"
	NodeTypeTask        WorkflowNodeType = "task"
	NodeTypeBranch      WorkflowNodeType = "branch"
	NodeTypeParallel    WorkflowNodeType = "parallel"
	NodeTypeCondition   WorkflowNodeType = "condition"
)

// ApprovalMode 审批模式
type ApprovalMode string

const (
	ApprovalModeSequential ApprovalMode = "sequential" // 顺序审批
	ApprovalModeParallel   ApprovalMode = "parallel"   // 并行审批
)

// AssigneeType 处理人类型
type AssigneeType string

const (
	AssigneeTypeUser       AssigneeType = "user"
	AssigneeTypeRole       AssigneeType = "role"
	AssigneeTypeDept       AssigneeType = "dept"
	AssigneeTypeExpression AssigneeType = "expression"
	AssigneeTypeSelf       AssigneeType = "self"
	AssigneeTypeStarter    AssigneeType = "starter"
)

// WorkflowDefinition 工作流定义
type WorkflowDefinition struct {
	BaseEntity
	Name          string          `gorm:"size:100;not null" json:"name"`
	Key           string          `gorm:"size:50;not null;uniqueIndex:idx_wf_key_version" json:"key"`
	Version       int             `gorm:"not null;default:1;uniqueIndex:idx_wf_key_version" json:"version"`
	Description   string          `gorm:"type:text" json:"description,omitempty"`
	Category      string          `gorm:"size:50" json:"category,omitempty"`
	Status        WorkflowStatus  `gorm:"size:20;not null;default:'draft'" json:"status"`
	StartNodeID   *string         `gorm:"type:uuid" json:"start_node_id,omitempty"`
	OrgID         string          `gorm:"type:uuid;not null;index" json:"org_id"`
	Config        JSONMap         `gorm:"type:jsonb" json:"config,omitempty"`
	FormSchema    JSONMap         `gorm:"type:jsonb" json:"form_schema,omitempty"`
	IsSystem      bool            `gorm:"default:false" json:"is_system"`
	
	Nodes []WorkflowNode `gorm:"foreignKey:WorkflowID" json:"nodes,omitempty"`
}

// TableName 表名
func (WorkflowDefinition) TableName() string {
	return "ty_workflow_definitions"
}

// IsActive 是否激活
func (w *WorkflowDefinition) IsActive() bool {
	return w.Status == WorkflowStatusActive
}

// CanPublish 是否可以发布
func (w *WorkflowDefinition) CanPublish() bool {
	return w.Status == WorkflowStatusDraft
}

// CanArchive 是否可以归档
func (w *WorkflowDefinition) CanArchive() bool {
	return w.Status == WorkflowStatusActive
}

// GetNode 获取指定节点
func (w *WorkflowDefinition) GetNode(nodeID string) *WorkflowNode {
	for i := range w.Nodes {
		if w.Nodes[i].ID == nodeID {
			return &w.Nodes[i]
		}
	}
	return nil
}

// GetStartNode 获取开始节点
func (w *WorkflowDefinition) GetStartNode() *WorkflowNode {
	if w.StartNodeID == nil {
		return nil
	}
	return w.GetNode(*w.StartNodeID)
}

// WorkflowNode 工作流节点
type WorkflowNode struct {
	BaseEntity
	WorkflowID    string           `gorm:"type:uuid;not null;index" json:"workflow_id"`
	Name          string           `gorm:"size:100;not null" json:"name"`
	Type          WorkflowNodeType `gorm:"size:20;not null" json:"type"`
	Config        JSONMap          `gorm:"type:jsonb" json:"config,omitempty"`
	
	// 处理人配置
	AssigneeType  AssigneeType     `gorm:"size:20" json:"assignee_type,omitempty"`
	Assignees     JSONMap          `gorm:"type:jsonb" json:"assignees,omitempty"`
	
	// 审批配置
	ApprovalMode  ApprovalMode     `gorm:"size:20" json:"approval_mode,omitempty"`
	RequiredCount int              `gorm:"default:1" json:"required_count"`
	AllowTransfer bool             `gorm:"default:true" json:"allow_transfer"`
	AllowDelegate bool             `gorm:"default:true" json:"allow_delegate"`
	AutoPass      bool             `gorm:"default:false" json:"auto_pass"`
	
	// 条件配置
	Conditions    JSONMap          `gorm:"type:jsonb" json:"conditions,omitempty"`
	
	// UI 配置
	PositionX     float64          `json:"position_x,omitempty"`
	PositionY     float64          `json:"position_y,omitempty"`
	OrderIndex    int              `json:"order_index"`
}

// TableName 表名
func (WorkflowNode) TableName() string {
	return "ty_workflow_nodes"
}

// IsApproval 是否是审批节点
func (n *WorkflowNode) IsApproval() bool {
	return n.Type == NodeTypeApproval
}

// IsTask 是否是任务节点
func (n *WorkflowNode) IsTask() bool {
	return n.Type == NodeTypeTask
}

// IsEnd 是否是结束节点
func (n *WorkflowNode) IsEnd() bool {
	return n.Type == NodeTypeEnd
}

// WorkflowInstance 流程实例
type WorkflowInstance struct {
	BaseEntity
	DefinitionID   string                  `gorm:"type:uuid;not null;index" json:"definition_id"`
	BusinessKey    string                  `gorm:"size:100;index" json:"business_key,omitempty"`
	BusinessID     string                  `gorm:"type:uuid;index" json:"business_id,omitempty"`
	Title          string                  `gorm:"size:200;not null" json:"title"`
	Status         WorkflowInstanceStatus  `gorm:"size:20;not null;default:'draft';index" json:"status"`
	CurrentNodeID  *string                 `gorm:"type:uuid" json:"current_node_id,omitempty"`
	StartedBy      string                  `gorm:"type:uuid;not null" json:"started_by"`
	StartedAt      *time.Time              `json:"started_at,omitempty"`
	CompletedAt    *time.Time              `json:"completed_at,omitempty"`
	Variables      JSONMap                 `gorm:"type:jsonb" json:"variables,omitempty"`
	Result         string                  `gorm:"size:20" json:"result,omitempty"` // approved/rejected
	Comment        string                  `gorm:"type:text" json:"comment,omitempty"`
	OrgID          string                  `gorm:"type:uuid;not null;index" json:"org_id"`
	
	Definition     *WorkflowDefinition     `gorm:"foreignKey:DefinitionID" json:"definition,omitempty"`
}

// TableName 表名
func (WorkflowInstance) TableName() string {
	return "ty_workflow_instances"
}

// IsActive 是否活跃
func (i *WorkflowInstance) IsActive() bool {
	return i.Status == InstanceStatusPending || i.Status == InstanceStatusProcessing
}

// IsCompleted 是否已完成
func (i *WorkflowInstance) IsCompleted() bool {
	return i.Status == InstanceStatusApproved || i.Status == InstanceStatusRejected
}

// CanApprove 是否可以审批
func (i *WorkflowInstance) CanApprove() bool {
	return i.Status == InstanceStatusPending || i.Status == InstanceStatusProcessing
}

// CanCancel 是否可以取消
func (i *WorkflowInstance) CanCancel() bool {
	return i.Status == InstanceStatusDraft || i.Status == InstanceStatusPending || i.Status == InstanceStatusProcessing
}

// Start 启动流程
func (i *WorkflowInstance) Start() error {
	if i.Status != InstanceStatusDraft {
		return errors.New("instance can only be started from draft status")
	}
	
	now := time.Now()
	i.StartedAt = &now
	i.Status = InstanceStatusPending
	return nil
}

// Complete 完成流程
func (i *WorkflowInstance) Complete(result string) error {
	if i.IsCompleted() {
		return errors.New("instance is already completed")
	}
	
	now := time.Now()
	i.CompletedAt = &now
	i.Result = result
	
	switch result {
	case "approved":
		i.Status = InstanceStatusApproved
	case "rejected":
		i.Status = InstanceStatusRejected
	default:
		return fmt.Errorf("invalid result: %s", result)
	}
	
	return nil
}

// Cancel 取消流程
func (i *WorkflowInstance) Cancel(reason string) error {
	if !i.CanCancel() {
		return errors.New("instance cannot be cancelled")
	}
	
	now := time.Now()
	i.CompletedAt = &now
	i.Status = InstanceStatusCancelled
	i.Comment = reason
	return nil
}

// WorkflowTask 工作流任务（增强现有 Task）
type WorkflowTask struct {
	Task
	WorkflowInstanceID string              `gorm:"type:uuid;index" json:"workflow_instance_id,omitempty"`
	WorkflowNodeID     string              `gorm:"type:uuid;index" json:"workflow_node_id,omitempty"`
	WorkflowNodeType   WorkflowNodeType    `json:"workflow_node_type,omitempty"`
	ApprovalAction     string              `gorm:"size:20" json:"approval_action,omitempty"` // approve/reject/transfer/delegate
	ApprovalComment    string              `gorm:"type:text" json:"approval_comment,omitempty"`
	DueTime            *time.Time          `json:"due_time,omitempty"`
	RemindedAt         *time.Time          `json:"reminded_at,omitempty"`
	RemindCount        int                 `gorm:"default:0" json:"remind_count"`
	DelegateFrom       *string             `gorm:"type:uuid" json:"delegate_from,omitempty"`
	TransferredFrom    *string             `gorm:"type:uuid" json:"transferred_from,omitempty"`
	SequentialIndex    int                 `gorm:"default:0" json:"sequential_index"` // 顺序审批序号
	
	Instance           *WorkflowInstance   `gorm:"foreignKey:WorkflowInstanceID" json:"instance,omitempty"`
}

// TableName 表名
func (WorkflowTask) TableName() string {
	return "ty_workflow_tasks"
}

// CanApprove 是否可以审批
func (t *WorkflowTask) CanApprove() bool {
	return t.Status == TaskStatusPending || t.Status == TaskStatusProcessing
}

// IsOverdue 是否逾期
func (t *WorkflowTask) IsOverdue() bool {
	if t.DueTime == nil {
		return false
	}
	return time.Now().After(*t.DueTime)
}

// WorkflowTransition 工作流转换记录
type WorkflowTransition struct {
	BaseEntity
	InstanceID     string                  `gorm:"type:uuid;not null;index" json:"instance_id"`
	FromNodeID     *string                 `gorm:"type:uuid" json:"from_node_id,omitempty"`
	ToNodeID       string                  `gorm:"type:uuid;not null" json:"to_node_id"`
	Action         string                  `gorm:"size:50;not null" json:"action"`
	UserID         string                  `gorm:"type:uuid;not null" json:"user_id"`
	Comment        string                  `gorm:"type:text" json:"comment,omitempty"`
	Variables      JSONMap                 `gorm:"type:jsonb" json:"variables,omitempty"`
}

// TableName 表名
func (WorkflowTransition) TableName() string {
	return "ty_workflow_transitions"
}

// WorkflowDelegation 任务委托
type WorkflowDelegation struct {
	BaseEntity
	TaskID       string     `gorm:"type:uuid;not null;index" json:"task_id"`
	FromUserID   string     `gorm:"type:uuid;not null" json:"from_user_id"`
	ToUserID     string     `gorm:"type:uuid;not null" json:"to_user_id"`
	StartTime    time.Time  `gorm:"not null" json:"start_time"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	Reason       string     `gorm:"type:text" json:"reason,omitempty"`
	Status       string     `gorm:"size:20;default:'active'" json:"status"` // active/cancelled/expired
}

// TableName 表名
func (WorkflowDelegation) TableName() string {
	return "ty_workflow_delegations"
}

// IsActive 是否有效
func (d *WorkflowDelegation) IsActive() bool {
	if d.Status != "active" {
		return false
	}
	if d.EndTime != nil && time.Now().After(*d.EndTime) {
		return false
	}
	return true
}

// WorkflowReminder 催办记录
type WorkflowReminder struct {
	BaseEntity
	TaskID       string     `gorm:"type:uuid;not null;index" json:"task_id"`
	ReminderType string     `gorm:"size:20;not null" json:"reminder_type"` // system/manual
	RemindCount  int        `gorm:"default:0" json:"remind_count"`
	LastRemindAt *time.Time `json:"last_remind_at,omitempty"`
	NextRemindAt *time.Time `json:"next_remind_at,omitempty"`
	Status       string     `gorm:"size:20;default:'pending'" json:"status"` // pending/completed
}

// TableName 表名
func (WorkflowReminder) TableName() string {
	return "ty_workflow_reminders"
}

// WorkflowQuery 工作流查询条件
type WorkflowQuery struct {
	DefinitionID   string
	InstanceID     string
	BusinessKey    string
	BusinessID     string
	Status         WorkflowInstanceStatus
	StartedBy      string
	CurrentNodeID  string
	OrgID          string
	Page           int
	PageSize       int
}

// WorkflowTaskQuery 工作流任务查询
type WorkflowTaskQuery struct {
	InstanceID         string
	AssigneeID         string
	Status             TaskStatus
	WorkflowInstanceID string
	IsOverdue          bool
	Page               int
	PageSize           int
}

// NewWorkflowDefinition 创建工作流定义
func NewWorkflowDefinition(name, key, orgID string) *WorkflowDefinition {
	return &WorkflowDefinition{
		BaseEntity: BaseEntity{
			ID: uuid.New().String(),
		},
		Name:     name,
		Key:      key,
		Version:  1,
		Status:   WorkflowStatusDraft,
		OrgID:    orgID,
		Config:   make(JSONMap),
	}
}

// NewWorkflowInstance 创建流程实例
func NewWorkflowInstance(definitionID, title, startedBy, orgID string) *WorkflowInstance {
	return &WorkflowInstance{
		BaseEntity: BaseEntity{
			ID: uuid.New().String(),
		},
		DefinitionID: definitionID,
		Title:        title,
		StartedBy:    startedBy,
		OrgID:        orgID,
		Status:       InstanceStatusDraft,
		Variables:    make(JSONMap),
	}
}

// StateTransition 状态转换
type StateTransition struct {
	From      WorkflowInstanceStatus
	To        WorkflowInstanceStatus
	Event     string
	CanTransition func(instance *WorkflowInstance) bool
}

// GetWorkflowTransitions 获取工作流状态转换规则
func GetWorkflowTransitions() []StateTransition {
	return []StateTransition{
		{InstanceStatusDraft, InstanceStatusPending, "start", nil},
		{InstanceStatusPending, InstanceStatusProcessing, "approve", nil},
		{InstanceStatusPending, InstanceStatusApproved, "approve", nil},
		{InstanceStatusPending, InstanceStatusRejected, "reject", nil},
		{InstanceStatusProcessing, InstanceStatusApproved, "approve", nil},
		{InstanceStatusProcessing, InstanceStatusRejected, "reject", nil},
		{InstanceStatusProcessing, InstanceStatusReturned, "return", nil},
		{InstanceStatusPending, InstanceStatusCancelled, "cancel", nil},
		{InstanceStatusProcessing, InstanceStatusCancelled, "cancel", nil},
		{InstanceStatusReturned, InstanceStatusPending, "resubmit", nil},
	}
}

// CanTransition 检查是否可以状态转换
func CanTransition(from WorkflowInstanceStatus, event string) (WorkflowInstanceStatus, bool) {
	transitions := GetWorkflowTransitions()
	for _, t := range transitions {
		if t.From == from && t.Event == event {
			return t.To, true
		}
	}
	return from, false
}

// WorkflowStats 工作流统计
type WorkflowStats struct {
	TotalInstances   int64            `json:"total_instances"`
	PendingCount     int64            `json:"pending_count"`
	ProcessingCount  int64            `json:"processing_count"`
	ApprovedCount    int64            `json:"approved_count"`
	RejectedCount    int64            `json:"rejected_count"`
	CancelledCount   int64            `json:"cancelled_count"`
	MyPendingCount   int64            `json:"my_pending_count"`
	MyCompletedCount int64            `json:"my_completed_count"`
	AvgProcessTime   int64            `json:"avg_process_time"` // 平均处理时间(秒)
	DefinitionStats  map[string]int64 `json:"definition_stats"`
}

// WorkflowNodeConfig 节点配置
type WorkflowNodeConfig struct {
	Timeout            *WorkflowTimeoutConfig    `json:"timeout,omitempty"`
	Reminder           *WorkflowReminderConfig   `json:"reminder,omitempty"`
	FormFields         []WorkflowFormField       `json:"form_fields,omitempty"`
	Buttons            []WorkflowButton          `json:"buttons,omitempty"`
}

// WorkflowTimeoutConfig 超时配置
type WorkflowTimeoutConfig struct {
	Duration     int    `json:"duration"`      // 时长
	Unit         string `json:"unit"`          // hour/day
	Action       string `json:"action"`        // remind/auto_approve/auto_reject
	EscalationTo string `json:"escalation_to"` // 升级处理人
}

// WorkflowReminderConfig 催办配置
type WorkflowReminderConfig struct {
	Intervals    []int  `json:"intervals"`     // 催办间隔(小时)
	MaxCount     int    `json:"max_count"`     // 最大催办次数
	AutoEscalate bool   `json:"auto_escalate"` // 是否自动升级
}

// WorkflowFormField 表单字段
type WorkflowFormField struct {
	Key          string      `json:"key"`
	Label        string      `json:"label"`
	Type         string      `json:"type"`
	Required     bool        `json:"required"`
	DefaultValue interface{} `json:"default_value,omitempty"`
	Options      []Option    `json:"options,omitempty"`
}

// Option 选项
type Option struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// WorkflowButton 按钮配置
type WorkflowButton struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Type    string `json:"type"`    // primary/danger/default
	Action  string `json:"action"`  // approve/reject/transfer/return
}

// MarshalJSON JSON序列化
func (w *WorkflowDefinition) MarshalJSON() ([]byte, error) {
	type Alias WorkflowDefinition
	return json.Marshal(&struct {
		*Alias
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}{
		Alias:     (*Alias)(w),
		CreatedAt: w.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: w.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}
