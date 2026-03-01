package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 任务状态
const (
	TaskStatusDraft      = "draft"      // 草稿
	TaskStatusPending    = "pending"    // 待分配
	TaskStatusAssigned   = "assigned"   // 已分配
	TaskStatusProcessing = "processing" // 进行中
	TaskStatusCompleted  = "completed"  // 已完成
	TaskStatusCancelled  = "cancelled"  // 已取消
	TaskStatusTimeout    = "timeout"    // 已超时
)

// 任务类型
const (
	TaskTypeSearch        = "search"         // 实地寻访
	TaskTypeCall          = "call"           // 电话核实
	TaskTypeInfoCollect   = "info_collect"   // 信息收集
	TaskTypeDialectRecord = "dialect_record" // 方言录制
	TaskTypeCoordination  = "coordination"   // 协调沟通
	TaskTypeOther         = "other"          // 其他
)

// 任务优先级
const (
	TaskPriorityUrgent = "urgent" // 紧急
	TaskPriorityHigh   = "high"   // 高
	TaskPriorityNormal = "normal" // 普通
	TaskPriorityLow    = "low"    // 低
)

// Task 任务模型
type Task struct {
	ID              uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TaskNo          string          `gorm:"size:50;uniqueIndex:idx_task_no;comment:任务编号" json:"task_no"`
	Title           string          `gorm:"size:200;not null;index:idx_task_title;comment:任务标题" json:"title"`
	Description     string          `gorm:"type:text;comment:任务描述" json:"description"`
	Type            string          `gorm:"size:30;not null;index:idx_task_type;comment:任务类型" json:"type"`
	Priority        string          `gorm:"size:20;default:normal;index:idx_task_priority;comment:优先级" json:"priority"`
	Status          string          `gorm:"size:20;default:draft;index:idx_task_status;comment:状态" json:"status"`

	// 关联案件
	MissingPersonID *uuid.UUID     `gorm:"type:uuid;index:idx_task_mp;comment:关联走失人员ID" json:"missing_person_id"`
	MissingPerson   *MissingPerson `gorm:"foreignKey:MissingPersonID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"missing_person,omitempty"`

	// 创建人/负责人
	CreatorID  uuid.UUID    `gorm:"type:uuid;index:idx_task_creator;not null" json:"creator_id"`
	Creator    User         `gorm:"foreignKey:CreatorID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"creator,omitempty"`
	AssigneeID *uuid.UUID   `gorm:"type:uuid;index:idx_task_assignee;comment:执行人ID" json:"assignee_id"`
	Assignee   *User        `gorm:"foreignKey:AssigneeID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"assignee,omitempty"`
	OrgID      uuid.UUID    `gorm:"type:uuid;index:idx_task_org;comment:所属组织ID" json:"org_id"`
	Org        Organization `gorm:"foreignKey:OrgID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"org,omitempty"`

	// 时间
	StartTime      *time.Time `gorm:"index:idx_task_start;comment:开始时间" json:"start_time"`
	Deadline       *time.Time `gorm:"index:idx_task_deadline;comment:截止时间" json:"deadline"`
	CompletedTime  *time.Time `gorm:"index:idx_task_completed;comment:完成时间" json:"completed_time"`
	EstimatedHours int        `gorm:"comment:预计工时" json:"estimated_hours"`
	ActualHours    int        `gorm:"comment:实际工时" json:"actual_hours"`

	// 地点
	Location  string  `gorm:"size:200;comment:任务地点" json:"location"`
	Longitude float64 `gorm:"comment:经度" json:"longitude"`
	Latitude  float64 `gorm:"comment:纬度" json:"latitude"`
	Address   string  `gorm:"size:200;comment:详细地址" json:"address"`

	// 任务要求
	Requirements string   `gorm:"type:text;comment:任务要求" json:"requirements"`
	Materials    []string `gorm:"type:jsonb;comment:所需材料" json:"materials"`
	Notes        string   `gorm:"type:text;comment:备注" json:"notes"`

	// 工作流
	WorkflowID  *uuid.UUID `gorm:"type:uuid;index:idx_task_workflow;comment:工作流ID" json:"workflow_id"`
	Workflow    *Workflow  `gorm:"foreignKey:WorkflowID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"workflow,omitempty"`
	CurrentStep int        `gorm:"default:0;comment:当前步骤" json:"current_step"`

	// 反馈
	Feedback    string           `gorm:"type:text;comment:反馈内容" json:"feedback"`
	Result      string           `gorm:"type:text;comment:任务结果" json:"result"`
	Attachments []TaskAttachment `gorm:"foreignKey:TaskID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"attachments,omitempty"`

	// 统计
	Progress     int `gorm:"default:0;comment:进度百分比" json:"progress"`
	ViewCount    int `gorm:"default:0;comment:查看次数" json:"view_count"`
	CommentCount int `gorm:"default:0;comment:评论数" json:"comment_count"`

	CreatedAt time.Time      `gorm:"index:idx_task_created" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TaskAttachment 任务附件
type TaskAttachment struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TaskID     uuid.UUID      `gorm:"type:uuid;index:idx_taskatt_task;not null" json:"task_id"`
	Task       Task           `gorm:"foreignKey:TaskID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Name       string         `gorm:"size:100;comment:文件名" json:"name"`
	URL        string         `gorm:"size:500;not null;comment:文件URL" json:"url"`
	Type       string         `gorm:"size:50;comment:文件类型" json:"type"`
	Size       int            `gorm:"comment:文件大小" json:"size"`
	UploaderID uuid.UUID      `gorm:"type:uuid;comment:上传人ID" json:"uploader_id"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// TaskLog 任务日志
type TaskLog struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TaskID    uuid.UUID      `gorm:"type:uuid;index:idx_tasklog_task;not null" json:"task_id"`
	Task      Task           `gorm:"foreignKey:TaskID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	UserID    uuid.UUID      `gorm:"type:uuid;index:idx_tasklog_user;not null" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"user,omitempty"`
	Action    string         `gorm:"size:50;not null;index:idx_tasklog_action;comment:操作类型" json:"action"`
	OldStatus string         `gorm:"size:20;comment:原状态" json:"old_status"`
	NewStatus string         `gorm:"size:20;comment:新状态" json:"new_status"`
	Content   string         `gorm:"type:text;comment:内容" json:"content"`
	IP        string         `gorm:"size:50;comment:IP地址" json:"ip"`
	CreatedAt time.Time      `gorm:"index:idx_tasklog_created" json:"created_at"`
}

// TaskComment 任务评论
type TaskComment struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TaskID      uuid.UUID      `gorm:"type:uuid;index:idx_taskcmt_task;not null" json:"task_id"`
	Task        Task           `gorm:"foreignKey:TaskID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	UserID      uuid.UUID      `gorm:"type:uuid;index:idx_taskcmt_user;not null" json:"user_id"`
	User        User           `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"user,omitempty"`
	Content     string         `gorm:"type:text;not null;comment:内容" json:"content"`
	Attachments []string       `gorm:"type:jsonb;comment:附件" json:"attachments"`
	ParentID    *uuid.UUID     `gorm:"type:uuid;index:idx_taskcmt_parent;comment:父评论ID" json:"parent_id"`
	CreatedAt   time.Time      `gorm:"index:idx_taskcmt_created" json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Task) TableName() string {
	return "tasks"
}

func (TaskAttachment) TableName() string {
	return "task_attachments"
}

func (TaskLog) TableName() string {
	return "task_logs"
}

func (TaskComment) TableName() string {
	return "task_comments"
}

// BeforeCreate 创建前钩子
func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.TaskNo == "" {
		t.TaskNo = generateTaskNo()
	}
	return nil
}

func generateTaskNo() string {
	return "TK" + time.Now().Format("20060102") + uuid.New().String()[:6]
}

// IsOverdue 检查是否逾期
func (t *Task) IsOverdue() bool {
	if t.Deadline == nil {
		return false
	}
	return time.Now().After(*t.Deadline) && t.Status != TaskStatusCompleted && t.Status != TaskStatusCancelled
}

// CanEdit 检查是否可以编辑
func (t *Task) CanEdit() bool {
	return t.Status == TaskStatusDraft || t.Status == TaskStatusPending
}

// CanAssign 检查是否可以分配
func (t *Task) CanAssign() bool {
	return t.Status == TaskStatusDraft || t.Status == TaskStatusPending
}

// CanComplete 检查是否可以完成
func (t *Task) CanComplete() bool {
	return t.Status == TaskStatusProcessing || t.Status == TaskStatusAssigned
}

// CanCancel 检查是否可以取消
func (t *Task) CanCancel() bool {
	return t.Status != TaskStatusCompleted && t.Status != TaskStatusCancelled
}

// UpdateProgress 更新进度
func (t *Task) UpdateProgress(progress int) {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	t.Progress = progress
	if progress == 100 && t.Status != TaskStatusCompleted {
		t.Status = TaskStatusCompleted
		now := time.Now()
		t.CompletedTime = &now
	}
}
