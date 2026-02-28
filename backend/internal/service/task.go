package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/model"
	"github.com/Snowitty-Re/CNtunyuan/internal/repository"
	"github.com/google/uuid"
)

// TaskService 任务服务
type TaskService struct {
	taskRepo    *repository.TaskRepository
	userRepo    *repository.UserRepository
	mpRepo      *repository.MissingPersonRepository
	orgRepo     *repository.OrganizationRepository
}

// NewTaskService 创建任务服务
func NewTaskService(
	taskRepo *repository.TaskRepository,
	userRepo *repository.UserRepository,
	mpRepo *repository.MissingPersonRepository,
	orgRepo *repository.OrganizationRepository,
) *TaskService {
	return &TaskService{
		taskRepo:    taskRepo,
		userRepo:    userRepo,
		mpRepo:      mpRepo,
		orgRepo:     orgRepo,
	}
}

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	Title           string    `json:"title" binding:"required"`
	Description     string    `json:"description"`
	Type            string    `json:"type" binding:"required"`
	Priority        string    `json:"priority" default:"normal"`
	MissingPersonID *string   `json:"missing_person_id"`
	AssigneeID      *string   `json:"assignee_id"`
	OrgID           string    `json:"org_id" binding:"required"`
	StartTime       *time.Time `json:"start_time"`
	Deadline        *time.Time `json:"deadline"`
	EstimatedHours  int       `json:"estimated_hours"`
	Location        string    `json:"location"`
	Longitude       float64   `json:"longitude"`
	Latitude        float64   `json:"latitude"`
	Address         string    `json:"address"`
	Requirements    string    `json:"requirements"`
	Materials       []string  `json:"materials"`
	Notes           string    `json:"notes"`
}

// CreateTask 创建任务
func (s *TaskService) CreateTask(ctx context.Context, creatorID uuid.UUID, req *CreateTaskRequest) (*model.Task, error) {
	// 验证组织
	orgID, err := uuid.Parse(req.OrgID)
	if err != nil {
		return nil, fmt.Errorf("无效的组织ID")
	}
	_, err = s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("组织不存在")
	}

	// 构建任务
	task := &model.Task{
		Title:          req.Title,
		Description:    req.Description,
		Type:           req.Type,
		Priority:       req.Priority,
		CreatorID:      creatorID,
		OrgID:          orgID,
		EstimatedHours: req.EstimatedHours,
		Location:       req.Location,
		Longitude:      req.Longitude,
		Latitude:       req.Latitude,
		Address:        req.Address,
		Requirements:   req.Requirements,
		Materials:      req.Materials,
		Notes:          req.Notes,
	}

	// 设置优先级默认值
	if task.Priority == "" {
		task.Priority = model.TaskPriorityNormal
	}

	// 关联走失人员
	if req.MissingPersonID != nil && *req.MissingPersonID != "" {
		mpID, err := uuid.Parse(*req.MissingPersonID)
		if err != nil {
			return nil, fmt.Errorf("无效的走失人员ID")
		}
		task.MissingPersonID = &mpID
	}

	// 设置时间
	if req.StartTime != nil {
		task.StartTime = req.StartTime
	}
	if req.Deadline != nil {
		task.Deadline = req.Deadline
	}

	// 如果指定了执行人，直接分配
	if req.AssigneeID != nil && *req.AssigneeID != "" {
		assigneeID, err := uuid.Parse(*req.AssigneeID)
		if err != nil {
			return nil, fmt.Errorf("无效的执行人ID")
		}
		// 验证执行人
		_, err = s.userRepo.GetByID(ctx, assigneeID)
		if err != nil {
			return nil, fmt.Errorf("执行人不存在")
		}
		task.AssigneeID = &assigneeID
		task.Status = model.TaskStatusAssigned
		task.StartTime = &[]time.Time{time.Now()}[0]
	} else {
		task.Status = model.TaskStatusPending
	}

	// 创建任务
	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("创建任务失败: %w", err)
	}

	// 创建日志
	s.createLog(ctx, task.ID, creatorID, "create", "", task.Status, "创建任务")

	return task, nil
}

// GetTask 获取任务详情
func (s *TaskService) GetTask(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("任务不存在")
	}
	return task, nil
}

// ListTasks 获取任务列表
type ListTasksRequest struct {
	Status          string `form:"status"`
	AssigneeID      string `form:"assignee_id"`
	CreatorID       string `form:"creator_id"`
	OrgID           string `form:"org_id"`
	Type            string `form:"type"`
	Priority        string `form:"priority"`
	MissingPersonID string `form:"missing_person_id"`
	Keyword         string `form:"keyword"`
	Page            int    `form:"page" default:"1"`
	PageSize        int    `form:"page_size" default:"20"`
}

func (s *TaskService) ListTasks(ctx context.Context, req *ListTasksRequest) ([]model.Task, int64, error) {
	params := map[string]interface{}{
		"page":     req.Page,
		"page_size": req.PageSize,
	}
	if req.Status != "" {
		params["status"] = req.Status
	}
	if req.AssigneeID != "" {
		params["assignee_id"] = req.AssigneeID
	}
	if req.CreatorID != "" {
		params["creator_id"] = req.CreatorID
	}
	if req.OrgID != "" {
		params["org_id"] = req.OrgID
	}
	if req.Type != "" {
		params["type"] = req.Type
	}
	if req.Priority != "" {
		params["priority"] = req.Priority
	}
	if req.MissingPersonID != "" {
		params["missing_person_id"] = req.MissingPersonID
	}
	if req.Keyword != "" {
		params["keyword"] = req.Keyword
	}

	return s.taskRepo.List(ctx, params)
}

// UpdateTask 更新任务
type UpdateTaskRequest struct {
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Type            string    `json:"type"`
	Priority        string    `json:"priority"`
	MissingPersonID *string   `json:"missing_person_id"`
	Deadline        *time.Time `json:"deadline"`
	EstimatedHours  int       `json:"estimated_hours"`
	Location        string    `json:"location"`
	Longitude       float64   `json:"longitude"`
	Latitude        float64   `json:"latitude"`
	Address         string    `json:"address"`
	Requirements    string    `json:"requirements"`
	Materials       []string  `json:"materials"`
	Notes           string    `json:"notes"`
}

func (s *TaskService) UpdateTask(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *UpdateTaskRequest) (*model.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("任务不存在")
	}

	// 检查是否可以编辑
	if !task.CanEdit() {
		return nil, fmt.Errorf("当前状态不允许编辑")
	}

	// 更新字段
	if req.Title != "" {
		task.Title = req.Title
	}
	if req.Description != "" {
		task.Description = req.Description
	}
	if req.Type != "" {
		task.Type = req.Type
	}
	if req.Priority != "" {
		task.Priority = req.Priority
	}
	if req.Deadline != nil {
		task.Deadline = req.Deadline
	}
	if req.EstimatedHours > 0 {
		task.EstimatedHours = req.EstimatedHours
	}
	if req.Location != "" {
		task.Location = req.Location
	}
	if req.Longitude != 0 {
		task.Longitude = req.Longitude
	}
	if req.Latitude != 0 {
		task.Latitude = req.Latitude
	}
	if req.Address != "" {
		task.Address = req.Address
	}
	if req.Requirements != "" {
		task.Requirements = req.Requirements
	}
	if req.Materials != nil {
		task.Materials = req.Materials
	}
	if req.Notes != "" {
		task.Notes = req.Notes
	}

	// 关联走失人员
	if req.MissingPersonID != nil {
		if *req.MissingPersonID == "" {
			task.MissingPersonID = nil
		} else {
			mpID, err := uuid.Parse(*req.MissingPersonID)
			if err != nil {
				return nil, fmt.Errorf("无效的走失人员ID")
			}
			task.MissingPersonID = &mpID
		}
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("更新任务失败: %w", err)
	}

	// 创建日志
	s.createLog(ctx, task.ID, userID, "update", "", task.Status, "更新任务信息")

	return task, nil
}

// DeleteTask 删除任务
func (s *TaskService) DeleteTask(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("任务不存在")
	}

	// 只有创建者或管理员可以删除
	if task.CreatorID != userID {
		// 检查是否为管理员
		user, err := s.userRepo.GetByID(ctx, userID)
		if err != nil || !user.IsAdmin() {
			return fmt.Errorf("无权删除此任务")
		}
	}

	// 只有草稿或待分配状态可以删除
	if task.Status != model.TaskStatusDraft && task.Status != model.TaskStatusPending {
		return fmt.Errorf("当前状态不允许删除")
	}

	if err := s.taskRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("删除任务失败: %w", err)
	}

	return nil
}

// AssignTask 分配任务
type AssignTaskRequest struct {
	AssigneeID string `json:"assignee_id" binding:"required"`
	Comment    string `json:"comment"`
}

func (s *TaskService) AssignTask(ctx context.Context, id uuid.UUID, operatorID uuid.UUID, req *AssignTaskRequest) error {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("任务不存在")
	}

	// 检查是否可以分配
	if !task.CanAssign() {
		return fmt.Errorf("当前状态不允许分配")
	}

	// 验证执行人
	assigneeID, err := uuid.Parse(req.AssigneeID)
	if err != nil {
		return fmt.Errorf("无效的执行人ID")
	}
	_, err = s.userRepo.GetByID(ctx, assigneeID)
	if err != nil {
		return fmt.Errorf("执行人不存在")
	}

	oldStatus := task.Status
	if err := s.taskRepo.Assign(ctx, id, assigneeID); err != nil {
		return fmt.Errorf("分配任务失败: %w", err)
	}

	// 创建日志
	content := fmt.Sprintf("分配给用户: %s", req.AssigneeID)
	if req.Comment != "" {
		content += fmt.Sprintf("，备注: %s", req.Comment)
	}
	s.createLog(ctx, id, operatorID, "assign", oldStatus, model.TaskStatusAssigned, content)

	return nil
}

// UnassignTask 取消分配
func (s *TaskService) UnassignTask(ctx context.Context, id uuid.UUID, operatorID uuid.UUID, reason string) error {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("任务不存在")
	}

	if task.Status != model.TaskStatusAssigned {
		return fmt.Errorf("当前状态不允许取消分配")
	}

	oldStatus := task.Status
	if err := s.taskRepo.Unassign(ctx, id); err != nil {
		return fmt.Errorf("取消分配失败: %w", err)
	}

	// 创建日志
	content := "取消任务分配"
	if reason != "" {
		content += fmt.Sprintf("，原因: %s", reason)
	}
	s.createLog(ctx, id, operatorID, "unassign", oldStatus, model.TaskStatusPending, content)

	return nil
}

// TransferTask 转派任务
type TransferTaskRequest struct {
	ToUserID string `json:"to_user_id" binding:"required"`
	Reason   string `json:"reason"`
}

func (s *TaskService) TransferTask(ctx context.Context, id uuid.UUID, fromUserID uuid.UUID, req *TransferTaskRequest) error {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("任务不存在")
	}

	// 只有当前执行人可以转派（或管理员）
	if task.AssigneeID == nil || *task.AssigneeID != fromUserID {
		user, err := s.userRepo.GetByID(ctx, fromUserID)
		if err != nil || !user.IsManager() {
			return fmt.Errorf("无权转派此任务")
		}
	}

	toUserID, err := uuid.Parse(req.ToUserID)
	if err != nil {
		return fmt.Errorf("无效的目标用户ID")
	}

	// 验证目标用户
	_, err = s.userRepo.GetByID(ctx, toUserID)
	if err != nil {
		return fmt.Errorf("目标用户不存在")
	}

	if err := s.taskRepo.Transfer(ctx, id, fromUserID, toUserID, req.Reason); err != nil {
		return fmt.Errorf("转派任务失败: %w", err)
	}

	// 创建日志
	content := fmt.Sprintf("任务转派给用户: %s", req.ToUserID)
	if req.Reason != "" {
		content += fmt.Sprintf("，原因: %s", req.Reason)
	}
	s.createLog(ctx, id, fromUserID, "transfer", task.Status, task.Status, content)

	return nil
}

// CompleteTask 完成任务
type CompleteTaskRequest struct {
	Feedback    string   `json:"feedback" binding:"required"`
	Result      string   `json:"result"`
	Attachments []string `json:"attachments"`
	ActualHours int      `json:"actual_hours"`
}

func (s *TaskService) CompleteTask(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *CompleteTaskRequest) (*model.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("任务不存在")
	}

	// 检查是否可以完成
	if !task.CanComplete() {
		return nil, fmt.Errorf("当前状态不允许完成")
	}

	// 只有执行人可以完成任务（或管理员）
	if task.AssigneeID == nil || *task.AssigneeID != userID {
		user, err := s.userRepo.GetByID(ctx, userID)
		if err != nil || !user.IsManager() {
			return nil, fmt.Errorf("无权完成此任务")
		}
	}

	oldStatus := task.Status
	if err := s.taskRepo.Complete(ctx, id, req.Feedback, req.Result, req.Attachments); err != nil {
		return nil, fmt.Errorf("完成任务失败: %w", err)
	}

	// 更新实际工时
	if req.ActualHours > 0 {
		task.ActualHours = req.ActualHours
		s.taskRepo.Update(ctx, task)
	}

	// 创建日志
	s.createLog(ctx, id, userID, "complete", oldStatus, model.TaskStatusCompleted, "完成任务")

	// 重新获取任务
	return s.taskRepo.GetByID(ctx, id)
}

// CancelTask 取消任务
func (s *TaskService) CancelTask(ctx context.Context, id uuid.UUID, userID uuid.UUID, reason string) error {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("任务不存在")
	}

	// 只有创建者或管理员可以取消
	if task.CreatorID != userID {
		user, err := s.userRepo.GetByID(ctx, userID)
		if err != nil || !user.IsAdmin() {
			return fmt.Errorf("无权取消此任务")
		}
	}

	// 已完成的任务不能取消
	if task.Status == model.TaskStatusCompleted {
		return fmt.Errorf("已完成的任务不能取消")
	}

	oldStatus := task.Status
	if err := s.taskRepo.Cancel(ctx, id, reason); err != nil {
		return fmt.Errorf("取消任务失败: %w", err)
	}

	// 创建日志
	content := "取消任务"
	if reason != "" {
		content += fmt.Sprintf("，原因: %s", reason)
	}
	s.createLog(ctx, id, userID, "cancel", oldStatus, model.TaskStatusCancelled, content)

	return nil
}

// UpdateProgress 更新进度
func (s *TaskService) UpdateProgress(ctx context.Context, id uuid.UUID, userID uuid.UUID, progress int) error {
	if progress < 0 || progress > 100 {
		return fmt.Errorf("进度必须在0-100之间")
	}

	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("任务不存在")
	}

	// 只有执行人可以更新进度
	if task.AssigneeID == nil || *task.AssigneeID != userID {
		return fmt.Errorf("无权更新此任务进度")
	}

	if err := s.taskRepo.UpdateProgress(ctx, id, progress); err != nil {
		return fmt.Errorf("更新进度失败: %w", err)
	}

	// 创建日志
	content := fmt.Sprintf("更新进度为 %d%%", progress)
	s.createLog(ctx, id, userID, "progress", task.Status, task.Status, content)

	return nil
}

// GetMyTasks 获取我的任务
func (s *TaskService) GetMyTasks(ctx context.Context, userID uuid.UUID, status string) ([]model.Task, error) {
	return s.taskRepo.GetMyTasks(ctx, userID, status)
}

// GetCreatedTasks 获取我创建的任务
func (s *TaskService) GetCreatedTasks(ctx context.Context, userID uuid.UUID) ([]model.Task, error) {
	return s.taskRepo.GetCreatedTasks(ctx, userID)
}

// GetTaskStatistics 获取任务统计
func (s *TaskService) GetTaskStatistics(ctx context.Context, orgID string) (map[string]interface{}, error) {
	return s.taskRepo.GetStatistics(ctx, orgID)
}

// GetTaskLogs 获取任务日志
func (s *TaskService) GetTaskLogs(ctx context.Context, taskID uuid.UUID) ([]model.TaskLog, error) {
	return s.taskRepo.GetLogs(ctx, taskID)
}

// AddComment 添加评论
type AddCommentRequest struct {
	Content     string   `json:"content" binding:"required"`
	Attachments []string `json:"attachments"`
	ParentID    *string  `json:"parent_id"`
}

func (s *TaskService) AddComment(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, req *AddCommentRequest) (*model.TaskComment, error) {
	// 验证任务
	_, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("任务不存在")
	}

	comment := &model.TaskComment{
		TaskID:      taskID,
		UserID:      userID,
		Content:     req.Content,
		Attachments: req.Attachments,
	}

	if req.ParentID != nil && *req.ParentID != "" {
		parentID, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("无效的父评论ID")
		}
		comment.ParentID = &parentID
	}

	if err := s.taskRepo.CreateComment(ctx, comment); err != nil {
		return nil, fmt.Errorf("添加评论失败: %w", err)
	}

	return comment, nil
}

// GetComments 获取评论列表
func (s *TaskService) GetComments(ctx context.Context, taskID uuid.UUID) ([]model.TaskComment, error) {
	return s.taskRepo.GetComments(ctx, taskID)
}

// BatchAssign 批量分配任务
type BatchAssignRequest struct {
	TaskIDs    []string `json:"task_ids" binding:"required"`
	AssigneeID string   `json:"assignee_id" binding:"required"`
	Comment    string   `json:"comment"`
}

func (s *TaskService) BatchAssign(ctx context.Context, operatorID uuid.UUID, req *BatchAssignRequest) error {
	assigneeID, err := uuid.Parse(req.AssigneeID)
	if err != nil {
		return fmt.Errorf("无效的执行人ID")
	}

	// 验证执行人
	_, err = s.userRepo.GetByID(ctx, assigneeID)
	if err != nil {
		return fmt.Errorf("执行人不存在")
	}

	// 转换任务ID
	taskIDs := make([]uuid.UUID, 0, len(req.TaskIDs))
	for _, idStr := range req.TaskIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		taskIDs = append(taskIDs, id)
	}

	if len(taskIDs) == 0 {
		return fmt.Errorf("没有有效的任务ID")
	}

	if err := s.taskRepo.BatchAssign(ctx, taskIDs, assigneeID); err != nil {
		return fmt.Errorf("批量分配失败: %w", err)
	}

	// 为每个任务创建日志
	for _, taskID := range taskIDs {
		content := fmt.Sprintf("批量分配给: %s", req.AssigneeID)
		if req.Comment != "" {
			content += fmt.Sprintf("，备注: %s", req.Comment)
		}
		s.createLog(ctx, taskID, operatorID, "batch_assign", "", model.TaskStatusAssigned, content)
	}

	return nil
}

// AutoAssignTasks 自动分配任务
func (s *TaskService) AutoAssignTasks(ctx context.Context, orgID string, limit int) error {
	// 获取待分配任务
	tasks, err := s.taskRepo.GetPendingTasks(ctx, limit)
	if err != nil {
		return fmt.Errorf("获取待分配任务失败: %w", err)
	}

	// 获取组织下的志愿者列表
	orgUUID, _ := uuid.Parse(orgID)
	users, err := s.userRepo.GetByOrgID(ctx, orgUUID)
	if err != nil {
		return fmt.Errorf("获取用户列表失败: %w", err)
	}

	if len(users) == 0 {
		return fmt.Errorf("没有可用的志愿者")
	}

	// 简单的负载均衡算法：分配给任务数最少的用户
	for _, task := range tasks {
		var minWorkload int64 = -1
		var selectedUser *model.User

		for _, user := range users {
			workload, err := s.taskRepo.GetUserWorkload(ctx, user.ID)
			if err != nil {
				continue
			}
			if minWorkload == -1 || workload < minWorkload {
				minWorkload = workload
				selectedUser = user
			}
		}

		if selectedUser != nil {
			if err := s.taskRepo.Assign(ctx, task.ID, selectedUser.ID); err != nil {
				continue
			}
			// 创建日志
			content := fmt.Sprintf("系统自动分配给: %s", selectedUser.Nickname)
			s.createLog(ctx, task.ID, selectedUser.ID, "auto_assign", model.TaskStatusPending, model.TaskStatusAssigned, content)
		}
	}

	return nil
}

// createLog 创建任务日志
func (s *TaskService) createLog(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, action, oldStatus, newStatus, content string) {
	log := &model.TaskLog{
		TaskID:    taskID,
		UserID:    userID,
		Action:    action,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		Content:   content,
		IP:        "",
	}
	s.taskRepo.CreateLog(ctx, log)
}
