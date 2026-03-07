package repository

import (
	"context"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
)

// TaskRepository 任务仓储接口
type TaskRepository interface {
	Repository[entity.Task]

	// List 分页查询
	List(ctx context.Context, query *TaskQuery) (*PageResult[entity.Task], error)

	// FindByAssignee 根据执行人查找
	FindByAssignee(ctx context.Context, assigneeID string, pagination Pagination) (*PageResult[entity.Task], error)

	// FindByCreator 根据创建人查找
	FindByCreator(ctx context.Context, creatorID string, pagination Pagination) (*PageResult[entity.Task], error)

	// FindByStatus 根据状态查找
	FindByStatus(ctx context.Context, status entity.TaskStatus, pagination Pagination) (*PageResult[entity.Task], error)

	// FindByOrg 根据组织查找
	FindByOrg(ctx context.Context, orgID string, pagination Pagination) (*PageResult[entity.Task], error)

	// FindByMissingPerson 根据走失人员查找
	FindByMissingPerson(ctx context.Context, missingPersonID string) ([]entity.Task, error)

	// FindPending 查找待分配任务
	FindPending(ctx context.Context, pagination Pagination) (*PageResult[entity.Task], error)

	// FindOverdue 查找逾期任务
	FindOverdue(ctx context.Context, pagination Pagination) (*PageResult[entity.Task], error)

	// UpdateStatus 更新状态
	UpdateStatus(ctx context.Context, id string, status entity.TaskStatus) error

	// UpdateProgress 更新进度
	UpdateProgress(ctx context.Context, id string, progress int) error

	// AddLog 添加日志
	AddLog(ctx context.Context, log *entity.TaskLog) error

	// GetLogs 获取日志
	GetLogs(ctx context.Context, taskID string) ([]entity.TaskLog, error)

	// AddAttachment 添加附件
	AddAttachment(ctx context.Context, attachment *entity.TaskAttachment) error

	// GetAttachments 获取附件
	GetAttachments(ctx context.Context, taskID string) ([]entity.TaskAttachment, error)

	// GetStats 获取统计
	GetStats(ctx context.Context, userID string) (*entity.TaskStats, error)

	// CountByStatus 按状态统计
	CountByStatus(ctx context.Context, status entity.TaskStatus) (int64, error)

	// CountByAssignee 按执行人统计
	CountByAssignee(ctx context.Context, assigneeID string) (int64, error)

	// CountOverdue 统计逾期任务
	CountOverdue(ctx context.Context) (int64, error)
}

// TaskQuery 任务查询参数
type TaskQuery struct {
	Pagination
	Keyword         string              `json:"keyword"`
	Type            entity.TaskType     `json:"type"`
	Status          entity.TaskStatus   `json:"status"`
	Priority        entity.TaskPriority `json:"priority"`
	CreatorID       string              `json:"creator_id"`
	AssigneeID      string              `json:"assignee_id"`
	OrgID           string              `json:"org_id"`
	MissingPersonID string              `json:"missing_person_id"`
	Province        string              `json:"province"`
	City            string              `json:"city"`
	StartDate       string              `json:"start_date"`
	EndDate         string              `json:"end_date"`
	IsOverdue       *bool               `json:"is_overdue,omitempty"`
	SortField       string              `json:"sort_field"`
	SortOrder       string              `json:"sort_order"`
}

// NewTaskQuery 创建默认任务查询
func NewTaskQuery() *TaskQuery {
	return &TaskQuery{
		Pagination: Pagination{
			Page:     1,
			PageSize: 10,
		},
		SortField: "created_at",
		SortOrder: "desc",
	}
}
