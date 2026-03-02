package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusDraft      TaskStatus = "draft"      // 草稿
	TaskStatusPending    TaskStatus = "pending"    // 待分配
	TaskStatusAssigned   TaskStatus = "assigned"   // 已分配
	TaskStatusProcessing TaskStatus = "processing" // 进行中
	TaskStatusCompleted  TaskStatus = "completed"  // 已完成
	TaskStatusCancelled  TaskStatus = "cancelled"  // 已取消
	TaskStatusOverdue    TaskStatus = "overdue"    // 已逾期
)

// TaskPriority 任务优先级
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityUrgent TaskPriority = "urgent"
)

// TaskType 任务类型
type TaskType string

const (
	TaskTypeSearch    TaskType = "search"    // 搜索任务
	TaskTypeVerify    TaskType = "verify"    // 核实任务
	TaskTypeAssist    TaskType = "assist"    // 协助任务
	TaskTypeFollow    TaskType = "follow"    // 跟进任务
	TaskTypeInterview TaskType = "interview" // 寻访任务
	TaskTypeOther     TaskType = "other"     // 其他任务
)

// Task 任务领域实体
type Task struct {
	BaseEntity
	Title        string        `gorm:"size:200;not null" json:"title"`
	Description  string        `gorm:"type:text" json:"description,omitempty"`
	Type         TaskType      `gorm:"size:20;not null" json:"type"`
	Priority     TaskPriority  `gorm:"size:20;default:'medium'" json:"priority"`
	Status       TaskStatus    `gorm:"size:20;default:'draft'" json:"status"`
	
	// 时间
	Deadline     *time.Time    `json:"deadline,omitempty"`
	StartedAt    *time.Time    `json:"started_at,omitempty"`
	CompletedAt  *time.Time    `json:"completed_at,omitempty"`
	
	// 关联
	CreatorID    string        `gorm:"type:uuid;not null;index" json:"creator_id"`
	AssigneeID   *string       `gorm:"type:uuid;index" json:"assignee_id,omitempty"`
	OrgID        string        `gorm:"type:uuid;not null;index" json:"org_id"`
	MissingPersonID *string    `gorm:"type:uuid;index" json:"missing_person_id,omitempty"`
	
	// 地点
	Location     string        `gorm:"size:255" json:"location,omitempty"`
	Province     string        `gorm:"size:50" json:"province,omitempty"`
	City         string        `gorm:"size:50" json:"city,omitempty"`
	District     string        `gorm:"size:50" json:"district,omitempty"`
	Address      string        `gorm:"size:255" json:"address,omitempty"`
	Lat          float64       `json:"lat,omitempty"`
	Lng          float64       `json:"lng,omitempty"`
	
	// 结果
	Result       string        `gorm:"type:text" json:"result,omitempty"`
	ResultPhotos string        `gorm:"type:json" json:"result_photos,omitempty"`
	Feedback     string        `gorm:"type:text" json:"feedback,omitempty"`
	
	// 统计
	Progress     int           `gorm:"default:0" json:"progress"`
	ViewCount    int           `gorm:"default:0" json:"view_count"`
	
	// 关联实体
	Creator      *User            `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	Assignee     *User            `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	Org          *Organization    `gorm:"foreignKey:OrgID" json:"org,omitempty"`
	MissingPerson *MissingPerson  `gorm:"foreignKey:MissingPersonID" json:"missing_person,omitempty"`
}

// TableName 表名
func (Task) TableName() string {
	return "ty_tasks"
}

// Validate 验证
func (t *Task) Validate() error {
	if t.Title == "" {
		return errors.New("任务标题不能为空")
	}
	if t.Type == "" {
		return errors.New("任务类型不能为空")
	}
	if !isValidTaskType(t.Type) {
		return errors.New("无效的任务类型")
	}
	return nil
}

// IsActive 是否活跃
func (t *Task) IsActive() bool {
	return t.Status != TaskStatusCompleted && 
	       t.Status != TaskStatusCancelled &&
	       t.Status != TaskStatusOverdue
}

// CanAssign 是否可以分配
func (t *Task) CanAssign() bool {
	return t.Status == TaskStatusDraft || t.Status == TaskStatusPending
}

// CanStart 是否可以开始
func (t *Task) CanStart() bool {
	return t.Status == TaskStatusAssigned
}

// CanUpdate 是否可以更新
func (t *Task) CanUpdate() bool {
	return t.Status != TaskStatusCompleted && t.Status != TaskStatusCancelled
}

// CanComplete 是否可以完成
func (t *Task) CanComplete() bool {
	return t.Status == TaskStatusProcessing || t.Status == TaskStatusAssigned
}

// CanCancel 是否可以取消
func (t *Task) CanCancel() bool {
	return t.Status != TaskStatusCompleted && t.Status != TaskStatusCancelled
}

// Assign 分配任务
func (t *Task) Assign(assigneeID string) error {
	if !t.CanAssign() {
		return errors.New("当前状态不能分配任务")
	}
	t.AssigneeID = &assigneeID
	t.Status = TaskStatusAssigned
	return nil
}

// Start 开始任务
func (t *Task) Start() error {
	if !t.CanStart() {
		return errors.New("当前状态不能开始任务")
	}
	now := time.Now()
	t.StartedAt = &now
	t.Status = TaskStatusProcessing
	return nil
}

// Complete 完成任务
func (t *Task) Complete(result string) error {
	if !t.CanComplete() {
		return errors.New("当前状态不能完成任务")
	}
	now := time.Now()
	t.CompletedAt = &now
	t.Status = TaskStatusCompleted
	t.Result = result
	t.Progress = 100
	return nil
}

// Cancel 取消任务
func (t *Task) Cancel(reason string) error {
	if !t.CanCancel() {
		return errors.New("当前状态不能取消任务")
	}
	t.Status = TaskStatusCancelled
	t.Feedback = reason
	return nil
}

// UpdateProgress 更新进度
func (t *Task) UpdateProgress(progress int) error {
	if progress < 0 || progress > 100 {
		return errors.New("进度必须在0-100之间")
	}
	t.Progress = progress
	return nil
}

// IsOverdue 是否逾期
func (t *Task) IsOverdue() bool {
	if t.Deadline == nil || t.Status == TaskStatusCompleted || t.Status == TaskStatusCancelled {
		return false
	}
	return time.Now().After(*t.Deadline)
}

// CheckAndUpdateOverdue 检查并更新逾期状态
func (t *Task) CheckAndUpdateOverdue() bool {
	if t.IsOverdue() && t.Status != TaskStatusOverdue {
		t.Status = TaskStatusOverdue
		return true
	}
	return false
}

// GetDuration 获取任务持续时间
func (t *Task) GetDuration() time.Duration {
	start := t.CreatedAt
	if t.StartedAt != nil {
		start = *t.StartedAt
	}
	
	end := time.Now()
	if t.CompletedAt != nil {
		end = *t.CompletedAt
	}
	
	return end.Sub(start)
}

// TaskAttachment 任务附件
type TaskAttachment struct {
	BaseEntity
	TaskID      string `gorm:"type:uuid;not null;index" json:"task_id"`
	FileName    string `gorm:"size:255;not null" json:"file_name"`
	FileUrl     string `gorm:"size:255;not null" json:"file_url"`
	FileType    string `gorm:"size:50" json:"file_type"`
	FileSize    int64  `json:"file_size"`
	Description string `gorm:"type:text" json:"description,omitempty"`
	UploadedBy  string `gorm:"type:uuid;not null" json:"uploaded_by"`
}

// TableName 表名
func (TaskAttachment) TableName() string {
	return "ty_task_attachments"
}

// TaskLog 任务日志
type TaskLog struct {
	BaseEntity
	TaskID      string    `gorm:"type:uuid;not null;index" json:"task_id"`
	UserID      string    `gorm:"type:uuid;not null" json:"user_id"`
	Action      string    `gorm:"size:50;not null" json:"action"`
	OldStatus   string    `gorm:"size:20" json:"old_status,omitempty"`
	NewStatus   string    `gorm:"size:20" json:"new_status,omitempty"`
	Content     string    `gorm:"type:text" json:"content,omitempty"`
	
	User        *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 表名
func (TaskLog) TableName() string {
	return "ty_task_logs"
}

// TaskComment 任务评论
type TaskComment struct {
	BaseEntity
	TaskID      string    `gorm:"type:uuid;not null;index" json:"task_id"`
	UserID      string    `gorm:"type:uuid;not null" json:"user_id"`
	Content     string    `gorm:"type:text;not null" json:"content"`
	ParentID    *string   `gorm:"type:uuid;index" json:"parent_id,omitempty"`
	
	User        *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 表名
func (TaskComment) TableName() string {
	return "ty_task_comments"
}

// TaskStats 任务统计
type TaskStats struct {
	Total        int64 `json:"total"`
	Draft        int64 `json:"draft"`
	Pending      int64 `json:"pending"`
	Assigned     int64 `json:"assigned"`
	Processing   int64 `json:"processing"`
	Completed    int64 `json:"completed"`
	Cancelled    int64 `json:"cancelled"`
	Overdue      int64 `json:"overdue"`
	MyTasks      int64 `json:"my_tasks"`
	MyPending    int64 `json:"my_pending"`
	MyCompleted  int64 `json:"my_completed"`
}

// isValidTaskType 验证任务类型
func isValidTaskType(t TaskType) bool {
	switch t {
	case TaskTypeSearch, TaskTypeVerify, TaskTypeAssist, TaskTypeFollow, TaskTypeInterview, TaskTypeOther:
		return true
	default:
		return false
	}
}

// NewTask 创建新任务
func NewTask(title string, taskType TaskType, creatorID, orgID string) (*Task, error) {
	task := &Task{
		BaseEntity: BaseEntity{
			ID: uuid.New().String(),
		},
		Title:    title,
		Type:     taskType,
		Status:   TaskStatusDraft,
		Priority: TaskPriorityMedium,
		CreatorID: creatorID,
		OrgID:    orgID,
	}

	if err := task.Validate(); err != nil {
		return nil, err
	}

	return task, nil
}
