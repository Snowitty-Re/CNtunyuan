package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 工作流状态
const (
	WorkflowStatusDraft    = "draft"
	WorkflowStatusActive   = "active"
	WorkflowStatusInactive = "inactive"
)

// 工作流实例状态
const (
	WorkflowInstanceStatusRunning   = "running"
	WorkflowInstanceStatusCompleted = "completed"
	WorkflowInstanceStatusCancelled = "cancelled"
	WorkflowInstanceStatusRejected  = "rejected"
)

// Workflow 工作流定义
type Workflow struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"size:100;not null;index:idx_workflow_name;comment:工作流名称" json:"name"`
	Code        string         `gorm:"size:50;uniqueIndex:idx_workflow_code;comment:工作流编码" json:"code"`
	Description string         `gorm:"type:text;comment:描述" json:"description"`
	Type        string         `gorm:"size:30;index:idx_workflow_type;comment:类型" json:"type"`
	Status      string         `gorm:"size:20;default:draft;index:idx_workflow_status;comment:状态" json:"status"`
	Version     int            `gorm:"default:1;comment:版本" json:"version"`
	IsDefault   bool           `gorm:"default:false;comment:是否默认" json:"is_default"`
	CreatorID   uuid.UUID      `gorm:"type:uuid;index:idx_workflow_creator;comment:创建人ID" json:"creator_id"`
	Creator     User           `gorm:"foreignKey:CreatorID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"creator,omitempty"`
	Steps       []WorkflowStep `gorm:"foreignKey:WorkflowID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"steps,omitempty"`
	CreatedAt   time.Time      `gorm:"index:idx_workflow_created" json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// WorkflowStep 工作流步骤
type WorkflowStep struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	WorkflowID    uuid.UUID `gorm:"type:uuid;index:idx_wfstep_workflow;not null" json:"workflow_id"`
	Workflow      Workflow  `gorm:"foreignKey:WorkflowID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Name          string    `gorm:"size:100;not null;comment:步骤名称" json:"name"`
	Description   string    `gorm:"type:text;comment:描述" json:"description"`
	StepOrder     int       `gorm:"not null;index:idx_wfstep_order;comment:步骤顺序" json:"step_order"`
	StepType      string    `gorm:"size:30;comment:步骤类型" json:"step_type"`

	// 执行人配置
	AssigneeType string `gorm:"size:30;comment:分配类型(auto/manual/role)" json:"assignee_type"`
	AssigneeRole string `gorm:"size:30;comment:分配角色" json:"assignee_role"`

	// 时间配置
	Duration        int  `gorm:"comment:预计时长(小时)" json:"duration"`
	IsTimeoutSkip   bool `gorm:"default:false;comment:超时时是否跳过" json:"is_timeout_skip"`

	// 表单配置
	FormConfig string `gorm:"type:jsonb;comment:表单配置" json:"form_config"`

	// 条件配置
	Conditions string `gorm:"type:jsonb;comment:流转条件" json:"conditions"`

	// 关联
	NextStepID *uuid.UUID `gorm:"comment:下一步ID" json:"next_step_id"`
	PrevStepID *uuid.UUID `gorm:"comment:上一步ID" json:"prev_step_id"`

	// 操作配置
	Actions string `gorm:"type:jsonb;comment:可执行操作" json:"actions"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// WorkflowInstance 工作流实例
type WorkflowInstance struct {
	ID            uuid.UUID         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	WorkflowID    uuid.UUID         `gorm:"type:uuid;index:idx_wfi_workflow;not null" json:"workflow_id"`
	Workflow      Workflow          `gorm:"foreignKey:WorkflowID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"workflow,omitempty"`
	BusinessID    uuid.UUID         `gorm:"type:uuid;index:idx_wfi_business;not null;comment:业务ID" json:"business_id"`
	BusinessType  string            `gorm:"size:30;index:idx_wfi_business_type;comment:业务类型" json:"business_type"`
	Title         string            `gorm:"size:200;comment:标题" json:"title"`
	Status        string            `gorm:"size:20;default:running;index:idx_wfi_status;comment:状态" json:"status"`
	CurrentStepID *uuid.UUID        `gorm:"comment:当前步骤ID" json:"current_step_id"`
	CurrentStep   *WorkflowStep     `gorm:"foreignKey:CurrentStepID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"current_step,omitempty"`
	StarterID     uuid.UUID         `gorm:"type:uuid;index:idx_wfi_starter;comment:发起人ID" json:"starter_id"`
	Starter       User              `gorm:"foreignKey:StarterID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"starter,omitempty"`
	History       []WorkflowHistory `gorm:"foreignKey:InstanceID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"history,omitempty"`
	StartTime     time.Time         `json:"start_time"`
	EndTime       *time.Time        `json:"end_time"`
	CreatedAt     time.Time         `gorm:"index:idx_wfi_created" json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// WorkflowHistory 工作流历史
type WorkflowHistory struct {
	ID         uuid.UUID     `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	InstanceID uuid.UUID     `gorm:"type:uuid;index:idx_wfh_instance;not null" json:"instance_id"`
	Instance   WorkflowInstance `gorm:"foreignKey:InstanceID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	StepID     uuid.UUID     `gorm:"type:uuid;index:idx_wfh_step;not null" json:"step_id"`
	Step       WorkflowStep  `gorm:"foreignKey:StepID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
	StepName   string        `gorm:"size:100;comment:步骤名称" json:"step_name"`
	OperatorID uuid.UUID     `gorm:"type:uuid;index:idx_wfh_operator;comment:操作人ID" json:"operator_id"`
	Operator   User          `gorm:"foreignKey:OperatorID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"operator,omitempty"`
	Action     string        `gorm:"size:50;index:idx_wfh_action;comment:操作" json:"action"`
	Comment    string        `gorm:"type:text;comment:意见" json:"comment"`
	FormData   string        `gorm:"type:jsonb;comment:表单数据" json:"form_data"`
	Attachments []string     `gorm:"type:jsonb;comment:附件" json:"attachments"`
	Duration   int           `gorm:"comment:耗时(分钟)" json:"duration"`
	StartTime  time.Time     `json:"start_time"`
	EndTime    *time.Time    `json:"end_time"`
	CreatedAt  time.Time     `gorm:"index:idx_wfh_created" json:"created_at"`
}

// TableName 指定表名
func (Workflow) TableName() string {
	return "workflows"
}

func (WorkflowStep) TableName() string {
	return "workflow_steps"
}

func (WorkflowInstance) TableName() string {
	return "workflow_instances"
}

func (WorkflowHistory) TableName() string {
	return "workflow_histories"
}

// IsCompleted 检查是否完成
func (w *WorkflowInstance) IsCompleted() bool {
	return w.Status == WorkflowInstanceStatusCompleted || 
		w.Status == WorkflowInstanceStatusCancelled || 
		w.Status == WorkflowInstanceStatusRejected
}

// IsRunning 检查是否进行中
func (w *WorkflowInstance) IsRunning() bool {
	return w.Status == WorkflowInstanceStatusRunning
}

// Complete 完成工作流实例
func (w *WorkflowInstance) Complete() {
	w.Status = WorkflowInstanceStatusCompleted
	now := time.Now()
	w.EndTime = &now
}

// Cancel 取消工作流实例
func (w *WorkflowInstance) Cancel() {
	w.Status = WorkflowInstanceStatusCancelled
	now := time.Now()
	w.EndTime = &now
}

// Reject 驳回工作流实例
func (w *WorkflowInstance) Reject() {
	w.Status = WorkflowInstanceStatusRejected
	now := time.Now()
	w.EndTime = &now
}
