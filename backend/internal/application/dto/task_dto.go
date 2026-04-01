package dto

import (
	"encoding/json"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
)

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	Title           string    `json:"title" binding:"required"`
	Description     string    `json:"description"`
	Type            string    `json:"type" binding:"required"`
	Priority        string    `json:"priority"`
	Deadline        time.Time `json:"deadline"`
	MissingPersonID string    `json:"missing_person_id"`
	AssigneeID      string    `json:"assignee_id"`
	Location        string    `json:"location"`
	Province        string    `json:"province"`
	City            string    `json:"city"`
	District        string    `json:"district"`
	Address         string    `json:"address"`
	Lat             float64   `json:"lat"`
	Lng             float64   `json:"lng"`
}

// UpdateTaskRequest 更新任务请求
type UpdateTaskRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Priority    string    `json:"priority"`
	Deadline    time.Time `json:"deadline"`
	Location    string    `json:"location"`
	Province    string    `json:"province"`
	City        string    `json:"city"`
	District    string    `json:"district"`
	Address     string    `json:"address"`
	Lat         float64   `json:"lat"`
	Lng         float64   `json:"lng"`
}

// TaskResponse 任务响应
type TaskResponse struct {
	ID              string                 `json:"id"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Type            string                 `json:"type"`
	Priority        string                 `json:"priority"`
	Status          string                 `json:"status"`
	Deadline        *time.Time             `json:"deadline,omitempty"`
	StartedAt       *time.Time             `json:"started_at,omitempty"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`
	CreatorID       string                 `json:"creator_id"`
	AssigneeID      *string                `json:"assignee_id,omitempty"`
	OrgID           string                 `json:"org_id"`
	MissingPersonID *string                `json:"missing_person_id,omitempty"`
	Location        string                 `json:"location"`
	Province        string                 `json:"province"`
	City            string                 `json:"city"`
	District        string                 `json:"district"`
	Address         string                 `json:"address"`
	Lat             float64                `json:"lat"`
	Lng             float64                `json:"lng"`
	Result          string                 `json:"result"`
	Feedback        string                 `json:"feedback"`
	Progress        int                    `json:"progress"`
	ViewCount       int                    `json:"view_count"`
	Creator         *UserResponse          `json:"creator,omitempty"`
	Assignee        *UserResponse          `json:"assignee,omitempty"`
	MissingPerson   *MissingPersonResponse `json:"missing_person,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
}

// TaskListRequest 任务列表请求
type TaskListRequest struct {
	Page            int    `form:"page,default=1" binding:"min=1"`
	PageSize        int    `form:"page_size,default=10" binding:"min=1,max=100"`
	Keyword         string `form:"keyword"`
	Type            string `form:"type"`
	Status          string `form:"status"`
	Priority        string `form:"priority"`
	AssigneeID      string `form:"assignee_id"`
	MissingPersonID string `form:"missing_person_id"`
	IsOverdue       *bool  `form:"is_overdue,omitempty"`
}

// TaskListResponse 任务列表响应
type TaskListResponse = PageResult[TaskResponse]

// AssignTaskRequest 分配任务请求
type AssignTaskRequest struct {
	AssigneeID string `json:"assignee_id" binding:"required"`
}

// UpdateTaskStatusRequest 更新状态请求
type UpdateTaskStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateTaskProgressRequest 更新进度请求
type UpdateTaskProgressRequest struct {
	Progress int    `json:"progress" binding:"required,min=0,max=100"`
	Remark   string `json:"remark"`
}

// CompleteTaskRequest 完成任务请求
type CompleteTaskRequest struct {
	Result      string   `json:"result"`
	Feedback    string   `json:"feedback"`
	Attachments []string `json:"attachments"`
}

// CancelTaskRequest 取消任务请求
type CancelTaskRequest struct {
	Reason string `json:"reason"`
}

// CreateTaskCommentRequest 创建评论请求
type CreateTaskCommentRequest struct {
	Content  string `json:"content" binding:"required"`
	ParentID string `json:"parent_id"`
}

// TaskCommentResponse 任务评论响应
type TaskCommentResponse struct {
	ID        string        `json:"id"`
	TaskID    string        `json:"task_id"`
	UserID    string        `json:"user_id"`
	Content   string        `json:"content"`
	ParentID  *string       `json:"parent_id,omitempty"`
	User      *UserResponse `json:"user,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
}

// CreateTaskFollowUpRequest 创建任务跟进请求
type CreateTaskFollowUpRequest struct {
	Content     string   `json:"content" binding:"required"`
	Progress    *int     `json:"progress,omitempty" binding:"omitempty,min=0,max=100"`
	Attachments []string `json:"attachments"`
}

// ReviewTaskFollowUpRequest 审核任务跟进请求
type ReviewTaskFollowUpRequest struct {
	Approve bool   `json:"approve"`
	Remark  string `json:"remark"`
}

// CreateTaskFollowUpCommentRequest 创建任务跟进评论请求
type CreateTaskFollowUpCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

// TaskFollowUpResponse 任务跟进响应
type TaskFollowUpResponse struct {
	ID           string        `json:"id"`
	TaskID       string        `json:"task_id"`
	UserID       string        `json:"user_id"`
	Content      string        `json:"content"`
	Progress     *int          `json:"progress,omitempty"`
	Attachments  []string      `json:"attachments"`
	Status       string        `json:"status"`
	ReviewRemark string        `json:"review_remark,omitempty"`
	ReviewedBy   *string       `json:"reviewed_by,omitempty"`
	ReviewedAt   *time.Time    `json:"reviewed_at,omitempty"`
	CommentCount int           `json:"comment_count"`
	User         *UserResponse `json:"user,omitempty"`
	Reviewer     *UserResponse `json:"reviewer,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
}

// TaskFollowUpCommentResponse 任务跟进评论响应
type TaskFollowUpCommentResponse struct {
	ID         string        `json:"id"`
	TaskID     string        `json:"task_id"`
	FollowUpID string        `json:"follow_up_id"`
	UserID     string        `json:"user_id"`
	Content    string        `json:"content"`
	User       *UserResponse `json:"user,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
}

// TaskLogResponse 任务日志响应
type TaskLogResponse struct {
	ID        string        `json:"id"`
	TaskID    string        `json:"task_id"`
	UserID    string        `json:"user_id"`
	Action    string        `json:"action"`
	OldStatus string        `json:"old_status"`
	NewStatus string        `json:"new_status"`
	Content   string        `json:"content"`
	User      *UserResponse `json:"user,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
}

// TaskStatsResponse 任务统计响应
type TaskStatsResponse struct {
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
	MyProcessing int64 `json:"my_processing"`
	MyCompleted  int64 `json:"my_completed"`
}

// ToTaskResponse 转换为任务响应
func ToTaskResponse(t *entity.Task) TaskResponse {
	resp := TaskResponse{
		ID:              t.ID,
		Title:           t.Title,
		Description:     t.Description,
		Type:            string(t.Type),
		Priority:        string(t.Priority),
		Status:          string(t.Status),
		Deadline:        t.Deadline,
		StartedAt:       t.StartedAt,
		CompletedAt:     t.CompletedAt,
		CreatorID:       t.CreatorID,
		AssigneeID:      t.AssigneeID,
		OrgID:           t.OrgID,
		MissingPersonID: t.MissingPersonID,
		Location:        t.Location,
		Province:        t.Province,
		City:            t.City,
		District:        t.District,
		Address:         t.Address,
		Lat:             t.Lat,
		Lng:             t.Lng,
		Result:          t.Result,
		Feedback:        t.Feedback,
		Progress:        t.Progress,
		ViewCount:       t.ViewCount,
		CreatedAt:       t.CreatedAt,
	}

	if t.Creator != nil {
		creator := ToUserResponse(t.Creator)
		resp.Creator = &creator
	}
	if t.Assignee != nil {
		assignee := ToUserResponse(t.Assignee)
		resp.Assignee = &assignee
	}
	if t.MissingPerson != nil {
		mp := ToMissingPersonResponse(t.MissingPerson)
		resp.MissingPerson = &mp
	}

	return resp
}

// ToTaskLogResponse 转换为任务日志响应
func ToTaskLogResponse(log *entity.TaskLog) TaskLogResponse {
	resp := TaskLogResponse{
		ID:        log.ID,
		TaskID:    log.TaskID,
		UserID:    log.UserID,
		Action:    log.Action,
		OldStatus: log.OldStatus,
		NewStatus: log.NewStatus,
		Content:   log.Content,
		CreatedAt: log.CreatedAt,
	}

	if log.User != nil {
		user := ToUserResponse(log.User)
		resp.User = &user
	}

	return resp
}

// ToTaskFollowUpResponse 转换为跟进响应
func ToTaskFollowUpResponse(f *entity.TaskFollowUp) TaskFollowUpResponse {
	resp := TaskFollowUpResponse{
		ID:           f.ID,
		TaskID:       f.TaskID,
		UserID:       f.UserID,
		Content:      f.Content,
		Progress:     f.Progress,
		Status:       string(f.Status),
		ReviewRemark: f.ReviewRemark,
		ReviewedBy:   f.ReviewedBy,
		ReviewedAt:   f.ReviewedAt,
		CommentCount: f.CommentCount,
		CreatedAt:    f.CreatedAt,
		Attachments:  []string{},
	}
	if f.Attachments != "" {
		_ = json.Unmarshal([]byte(f.Attachments), &resp.Attachments)
	}
	if f.User != nil {
		user := ToUserResponse(f.User)
		resp.User = &user
	}
	if f.Reviewer != nil {
		reviewer := ToUserResponse(f.Reviewer)
		resp.Reviewer = &reviewer
	}
	return resp
}

// ToTaskFollowUpCommentResponse 转换为跟进评论响应
func ToTaskFollowUpCommentResponse(c *entity.TaskFollowUpComment) TaskFollowUpCommentResponse {
	resp := TaskFollowUpCommentResponse{
		ID:         c.ID,
		TaskID:     c.TaskID,
		FollowUpID: c.FollowUpID,
		UserID:     c.UserID,
		Content:    c.Content,
		CreatedAt:  c.CreatedAt,
	}
	if c.User != nil {
		user := ToUserResponse(c.User)
		resp.User = &user
	}
	return resp
}

// NewTaskListResponse 创建任务列表响应
func NewTaskListResponse(list []TaskResponse, total int64, page, pageSize int) TaskListResponse {
	if pageSize <= 0 {
		pageSize = 10
	}
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return TaskListResponse{
		List:       list,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
