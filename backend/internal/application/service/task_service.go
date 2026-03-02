package service

import (
	"context"
	"errors"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
)

var (
	ErrTaskNotFound    = errors.New("task not found")
	ErrTaskInvalidStatus = errors.New("invalid status transition")
	ErrAlreadyAssigned = errors.New("task already assigned")
)

// TaskAppService 任务应用服务
type TaskAppService struct {
	taskRepo repository.TaskRepository
}

// NewTaskAppService 创建任务应用服务
func NewTaskAppService(taskRepo repository.TaskRepository) *TaskAppService {
	return &TaskAppService{taskRepo: taskRepo}
}

// Create 创建任务
func (s *TaskAppService) Create(ctx context.Context, req *dto.CreateTaskRequest, creatorID string, orgID string) (*dto.TaskResponse, error) {
	task := &entity.Task{
		Title:           req.Title,
		Description:     req.Description,
		Type:            entity.TaskType(req.Type),
		Priority:        entity.TaskPriority(req.Priority),
		Deadline:        &req.Deadline,
		CreatorID:       creatorID,
		OrgID:           orgID,
		MissingPersonID: &req.MissingPersonID,
		Location:        req.Location,
		Province:        req.Province,
		City:            req.City,
		District:        req.District,
		Address:         req.Address,
		Lat:             req.Lat,
		Lng:             req.Lng,
		Status:          entity.TaskStatusDraft,
	}

	if req.Priority == "" {
		task.Priority = entity.TaskPriorityMedium
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		logger.Error("Failed to create task", logger.Err(err))
		return nil, err
	}

	logger.Info("Task created", logger.String("task_id", task.ID))

	resp := dto.ToTaskResponse(task)
	return &resp, nil
}

// GetByID 根据ID获取
func (s *TaskAppService) GetByID(ctx context.Context, id string) (*dto.TaskResponse, error) {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrTaskNotFound
	}

	resp := dto.ToTaskResponse(task)
	return &resp, nil
}

// List 列表查询
func (s *TaskAppService) List(ctx context.Context, req *dto.TaskListRequest) (*dto.TaskListResponse, error) {
	query := repository.NewTaskQuery()
	query.Page = req.Page
	query.PageSize = req.PageSize
	query.Keyword = req.Keyword
	query.Type = entity.TaskType(req.Type)
	query.Status = entity.TaskStatus(req.Status)
	query.Priority = entity.TaskPriority(req.Priority)
	query.AssigneeID = req.AssigneeID
	query.MissingPersonID = req.MissingPersonID
	query.IsOverdue = req.IsOverdue

	result, err := s.taskRepo.List(ctx, query)
	if err != nil {
		return nil, err
	}

	list := make([]dto.TaskResponse, len(result.List))
	for i, task := range result.List {
		list[i] = dto.ToTaskResponse(&task)
	}

	resp := dto.NewTaskListResponse(list, result.Total, result.Page, result.PageSize)
	return &resp, nil
}

// Update 更新
func (s *TaskAppService) Update(ctx context.Context, id string, req *dto.UpdateTaskRequest) (*dto.TaskResponse, error) {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrTaskNotFound
	}

	if !task.CanUpdate() {
		return nil, errors.New("cannot update completed or cancelled task")
	}

	if req.Title != "" {
		task.Title = req.Title
	}
	if req.Description != "" {
		task.Description = req.Description
	}
	if req.Type != "" {
		task.Type = entity.TaskType(req.Type)
	}
	if req.Priority != "" {
		task.Priority = entity.TaskPriority(req.Priority)
	}
	if !req.Deadline.IsZero() {
		task.Deadline = &req.Deadline
	}
	if req.Location != "" {
		task.Location = req.Location
	}
	if req.Province != "" {
		task.Province = req.Province
	}
	if req.City != "" {
		task.City = req.City
	}
	if req.District != "" {
		task.District = req.District
	}
	if req.Address != "" {
		task.Address = req.Address
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		logger.Error("Failed to update task", logger.Err(err))
		return nil, err
	}

	resp := dto.ToTaskResponse(task)
	return &resp, nil
}

// Delete 删除
func (s *TaskAppService) Delete(ctx context.Context, id string) error {
	if err := s.taskRepo.SoftDelete(ctx, id); err != nil {
		logger.Error("Failed to delete task", logger.Err(err))
		return err
	}
	return nil
}

// Assign 分配任务
func (s *TaskAppService) Assign(ctx context.Context, id string, assigneeID string) error {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return ErrTaskNotFound
	}

	if err := task.Assign(assigneeID); err != nil {
		return err
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return err
	}

	// 记录日志
	log := &entity.TaskLog{
		TaskID:    id,
		UserID:    assigneeID,
		Action:    "assign",
		NewStatus: string(task.Status),
		Content:   "Task assigned",
	}
	s.taskRepo.AddLog(ctx, log)

	return nil
}

// Start 开始任务
func (s *TaskAppService) Start(ctx context.Context, id string, userID string) error {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return ErrTaskNotFound
	}

	if err := task.Start(); err != nil {
		return err
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return err
	}

	log := &entity.TaskLog{
		TaskID:    id,
		UserID:    userID,
		Action:    "start",
		NewStatus: string(task.Status),
		Content:   "Task started",
	}
	s.taskRepo.AddLog(ctx, log)

	return nil
}

// Complete 完成任务
func (s *TaskAppService) Complete(ctx context.Context, id string, req *dto.CompleteTaskRequest, userID string) error {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return ErrTaskNotFound
	}

	if err := task.Complete(req.Result); err != nil {
		return err
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return err
	}

	log := &entity.TaskLog{
		TaskID:    id,
		UserID:    userID,
		Action:    "complete",
		NewStatus: string(task.Status),
		Content:   req.Result,
	}
	s.taskRepo.AddLog(ctx, log)

	return nil
}

// Cancel 取消任务
func (s *TaskAppService) Cancel(ctx context.Context, id string, req *dto.CancelTaskRequest, userID string) error {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return ErrTaskNotFound
	}

	if err := task.Cancel(req.Reason); err != nil {
		return err
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return err
	}

	log := &entity.TaskLog{
		TaskID:    id,
		UserID:    userID,
		Action:    "cancel",
		NewStatus: string(task.Status),
		Content:   req.Reason,
	}
	s.taskRepo.AddLog(ctx, log)

	return nil
}

// UpdateProgress 更新进度
func (s *TaskAppService) UpdateProgress(ctx context.Context, id string, progress int, userID string) error {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return ErrTaskNotFound
	}

	if err := task.UpdateProgress(progress); err != nil {
		return err
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return err
	}

	return nil
}

// GetMyTasks 获取我的任务
func (s *TaskAppService) GetMyTasks(ctx context.Context, userID string, page, pageSize int) (*dto.TaskListResponse, error) {
	pagination := repository.Pagination{Page: page, PageSize: pageSize}
	result, err := s.taskRepo.FindByAssignee(ctx, userID, pagination)
	if err != nil {
		return nil, err
	}

	list := make([]dto.TaskResponse, len(result.List))
	for i, task := range result.List {
		list[i] = dto.ToTaskResponse(&task)
	}

	resp := dto.NewTaskListResponse(list, result.Total, result.Page, result.PageSize)
	return &resp, nil
}

// GetPendingTasks 获取待分配任务
func (s *TaskAppService) GetPendingTasks(ctx context.Context, page, pageSize int) (*dto.TaskListResponse, error) {
	pagination := repository.Pagination{Page: page, PageSize: pageSize}
	result, err := s.taskRepo.FindPending(ctx, pagination)
	if err != nil {
		return nil, err
	}

	list := make([]dto.TaskResponse, len(result.List))
	for i, task := range result.List {
		list[i] = dto.ToTaskResponse(&task)
	}

	resp := dto.NewTaskListResponse(list, result.Total, result.Page, result.PageSize)
	return &resp, nil
}

// GetOverdueTasks 获取逾期任务
func (s *TaskAppService) GetOverdueTasks(ctx context.Context, page, pageSize int) (*dto.TaskListResponse, error) {
	pagination := repository.Pagination{Page: page, PageSize: pageSize}
	result, err := s.taskRepo.FindOverdue(ctx, pagination)
	if err != nil {
		return nil, err
	}

	list := make([]dto.TaskResponse, len(result.List))
	for i, task := range result.List {
		list[i] = dto.ToTaskResponse(&task)
	}

	resp := dto.NewTaskListResponse(list, result.Total, result.Page, result.PageSize)
	return &resp, nil
}

// GetLogs 获取任务日志
func (s *TaskAppService) GetLogs(ctx context.Context, taskID string) ([]dto.TaskLogResponse, error) {
	logs, err := s.taskRepo.GetLogs(ctx, taskID)
	if err != nil {
		return nil, err
	}

	list := make([]dto.TaskLogResponse, len(logs))
	for i, log := range logs {
		list[i] = dto.ToTaskLogResponse(&log)
	}

	return list, nil
}

// GetStats 获取统计
func (s *TaskAppService) GetStats(ctx context.Context, userID string) (*dto.TaskStatsResponse, error) {
	stats, err := s.taskRepo.GetStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &dto.TaskStatsResponse{
		Total:       stats.Total,
		Draft:       stats.Draft,
		Pending:     stats.Pending,
		Assigned:    stats.Assigned,
		Processing:  stats.Processing,
		Completed:   stats.Completed,
		Cancelled:   stats.Cancelled,
		Overdue:     stats.Overdue,
		MyTasks:     stats.MyTasks,
		MyPending:   stats.MyPending,
		MyCompleted: stats.MyCompleted,
	}, nil
}
