package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/Snowitty-Re/CNtunyuan/pkg/validator"
)

var (
	ErrTaskNotFound          = errors.New("task not found")
	ErrTaskInvalidStatus     = errors.New("invalid status transition")
	ErrAlreadyAssigned       = errors.New("task already assigned")
	ErrTaskNotAssignedToUser = errors.New("task not assigned to this user")
	ErrTaskForbidden         = errors.New("no permission to modify this task")
	ErrTaskFollowUpNotFound  = errors.New("task follow up not found")
	ErrTaskFollowUpReviewed  = errors.New("task follow up already reviewed")
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
		Title:       req.Title,
		Description: req.Description,
		Type:        entity.TaskType(req.Type),
		CreatorID:   creatorID,
		OrgID:       orgID,
		Location:    req.Location,
		Province:    req.Province,
		City:        req.City,
		District:    req.District,
		Address:     req.Address,
		Lat:         req.Lat,
		Lng:         req.Lng,
		Status:      entity.TaskStatusPending,
	}

	// 设置优先级
	if req.Priority != "" {
		task.Priority = entity.TaskPriority(req.Priority)
	} else {
		task.Priority = entity.TaskPriorityMedium
	}

	// 设置截止日期（如果不是零值）
	if !req.Deadline.IsZero() {
		task.Deadline = &req.Deadline
	}

	// 设置关联的走失人员（如果提供）
	if req.MissingPersonID != "" {
		task.MissingPersonID = &req.MissingPersonID
	}

	// 创建时允许直接分配执行人：有 assignee 则状态进入 assigned
	if req.AssigneeID != "" {
		if err := task.Assign(req.AssigneeID); err != nil {
			return nil, err
		}
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
	req.Page, req.PageSize = validator.SanitizePagination(req.Page, req.PageSize)

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
func (s *TaskAppService) Update(ctx context.Context, id string, req *dto.UpdateTaskRequest, userID string, isManager bool) (*dto.TaskResponse, error) {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrTaskNotFound
	}

	// 只有任务创建者、任务执行人或管理员可以更新任务
	isCreator := task.CreatorID == userID
	isAssignee := task.AssigneeID != nil && *task.AssigneeID == userID
	if !isManager && !isCreator && !isAssignee {
		return nil, ErrTaskForbidden
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

	// 记录日志（非关键操作，失败仅记录警告）
	log := &entity.TaskLog{
		TaskID:    id,
		UserID:    assigneeID,
		Action:    "assign",
		NewStatus: string(task.Status),
		Content:   "Task assigned",
	}
	if err := s.taskRepo.AddLog(ctx, log); err != nil {
		logger.Warn("Failed to add task assign log", logger.String("task_id", id), logger.Err(err))
	}

	return nil
}

// Start 开始任务（只有被分配的执行人才能开始任务）
func (s *TaskAppService) Start(ctx context.Context, id string, userID string) error {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return ErrTaskNotFound
	}

	// 只有被分配的执行人才能开始任务
	if task.AssigneeID == nil || *task.AssigneeID != userID {
		return ErrTaskNotAssignedToUser
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
	if err := s.taskRepo.AddLog(ctx, log); err != nil {
		logger.Warn("Failed to add task start log", logger.String("task_id", id), logger.Err(err))
	}

	return nil
}

// Complete 完成任务（只有被分配的执行人才能完成任务）
func (s *TaskAppService) Complete(ctx context.Context, id string, req *dto.CompleteTaskRequest, userID string) error {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return ErrTaskNotFound
	}

	// 只有被分配的执行人才能完成任务
	if task.AssigneeID == nil || *task.AssigneeID != userID {
		return ErrTaskNotAssignedToUser
	}

	if err := task.Complete(req.Result); err != nil {
		return err
	}

	if req.Feedback != "" {
		task.Feedback = req.Feedback
	}
	if len(req.Attachments) > 0 {
		if b, err := json.Marshal(req.Attachments); err == nil {
			task.ResultPhotos = string(b)
		}
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return err
	}

	logContent := req.Result
	if req.Feedback != "" {
		logContent = req.Feedback
	}
	log := &entity.TaskLog{
		TaskID:    id,
		UserID:    userID,
		Action:    "complete",
		NewStatus: string(task.Status),
		Content:   logContent,
	}
	if err := s.taskRepo.AddLog(ctx, log); err != nil {
		logger.Warn("Failed to add task complete log", logger.String("task_id", id), logger.Err(err))
	}

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
	if err := s.taskRepo.AddLog(ctx, log); err != nil {
		logger.Warn("Failed to add task cancel log", logger.String("task_id", id), logger.Err(err))
	}

	return nil
}

// UpdateProgress 更新进度
func (s *TaskAppService) UpdateProgress(ctx context.Context, id string, progress int, remark string, userID string) error {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return ErrTaskNotFound
	}

	// 只有被分配的执行人才能更新进度
	if task.AssigneeID == nil || *task.AssigneeID != userID {
		return ErrTaskNotAssignedToUser
	}

	if err := task.UpdateProgress(progress); err != nil {
		return err
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return err
	}

	content := fmt.Sprintf("进度更新至 %d%%", progress)
	if remark != "" {
		content += "：" + remark
	}
	log := &entity.TaskLog{
		TaskID:  id,
		UserID:  userID,
		Action:  "update_progress",
		Content: content,
	}
	if err := s.taskRepo.AddLog(ctx, log); err != nil {
		logger.Warn("Failed to add progress log", logger.Err(err), logger.String("task_id", id))
	}

	return nil
}

// CreateFollowUp 创建任务跟进记录
func (s *TaskAppService) CreateFollowUp(ctx context.Context, taskID string, req *dto.CreateTaskFollowUpRequest, userID string, isManager bool) (*dto.TaskFollowUpResponse, error) {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, ErrTaskNotFound
	}
	if task.AssigneeID == nil || (*task.AssigneeID != userID && !isManager) {
		return nil, ErrTaskForbidden
	}
	if task.Status != entity.TaskStatusAssigned && task.Status != entity.TaskStatusProcessing {
		return nil, errors.New("当前状态不能提交跟进")
	}

	attachmentsJSON := "[]"
	if len(req.Attachments) > 0 {
		if b, err := json.Marshal(req.Attachments); err == nil {
			attachmentsJSON = string(b)
		}
	}

	followUp := &entity.TaskFollowUp{
		TaskID:      taskID,
		UserID:      userID,
		Content:     req.Content,
		Attachments: attachmentsJSON,
		Status:      entity.TaskFollowUpStatusPending,
	}
	if req.Progress != nil {
		followUp.Progress = req.Progress
		if err := task.UpdateProgress(*req.Progress); err == nil {
			_ = s.taskRepo.Update(ctx, task)
		}
	}

	if err := s.taskRepo.AddFollowUp(ctx, followUp); err != nil {
		return nil, err
	}

	log := &entity.TaskLog{
		TaskID:  taskID,
		UserID:  userID,
		Action:  "follow_up_create",
		Content: "提交任务跟进",
	}
	if err := s.taskRepo.AddLog(ctx, log); err != nil {
		logger.Warn("Failed to add follow up create log", logger.Err(err), logger.String("task_id", taskID))
	}

	resp := dto.ToTaskFollowUpResponse(followUp)
	return &resp, nil
}

// ListFollowUps 获取任务跟进列表
func (s *TaskAppService) ListFollowUps(ctx context.Context, taskID string, page, pageSize int) (*dto.PageResult[dto.TaskFollowUpResponse], error) {
	if _, err := s.taskRepo.FindByID(ctx, taskID); err != nil {
		return nil, ErrTaskNotFound
	}

	page, pageSize = validator.SanitizePagination(page, pageSize)
	pagination := repository.Pagination{Page: page, PageSize: pageSize}
	result, err := s.taskRepo.GetFollowUps(ctx, taskID, pagination)
	if err != nil {
		return nil, err
	}

	list := make([]dto.TaskFollowUpResponse, len(result.List))
	for i, item := range result.List {
		list[i] = dto.ToTaskFollowUpResponse(&item)
	}

	totalPages := int(result.Total) / pageSize
	if int(result.Total)%pageSize > 0 {
		totalPages++
	}
	return &dto.PageResult[dto.TaskFollowUpResponse]{
		List:       list,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetFollowUpByID 获取任务跟进详情
func (s *TaskAppService) GetFollowUpByID(ctx context.Context, taskID, followUpID, userID string, isManager bool) (*dto.TaskFollowUpResponse, error) {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, ErrTaskNotFound
	}

	isCreator := task.CreatorID == userID
	isAssignee := task.AssigneeID != nil && *task.AssigneeID == userID
	if !isManager && !isCreator && !isAssignee {
		return nil, ErrTaskForbidden
	}

	followUp, err := s.taskRepo.FindFollowUpByID(ctx, taskID, followUpID)
	if err != nil {
		return nil, ErrTaskFollowUpNotFound
	}

	resp := dto.ToTaskFollowUpResponse(followUp)
	return &resp, nil
}

// ReviewFollowUp 审核任务跟进
func (s *TaskAppService) ReviewFollowUp(ctx context.Context, taskID, followUpID string, req *dto.ReviewTaskFollowUpRequest, reviewerID string, isManager bool) error {
	if !isManager {
		return ErrTaskForbidden
	}
	if _, err := s.taskRepo.FindByID(ctx, taskID); err != nil {
		return ErrTaskNotFound
	}

	followUp, err := s.taskRepo.FindFollowUpByID(ctx, taskID, followUpID)
	if err != nil {
		return ErrTaskFollowUpNotFound
	}
	if followUp.Status != entity.TaskFollowUpStatusPending {
		return ErrTaskFollowUpReviewed
	}

	now := time.Now()
	followUp.ReviewRemark = req.Remark
	followUp.ReviewedBy = &reviewerID
	followUp.ReviewedAt = &now
	if req.Approve {
		followUp.Status = entity.TaskFollowUpStatusApproved
	} else {
		followUp.Status = entity.TaskFollowUpStatusRejected
	}

	if err := s.taskRepo.UpdateFollowUp(ctx, followUp); err != nil {
		return err
	}

	action := "follow_up_approve"
	if !req.Approve {
		action = "follow_up_reject"
	}
	log := &entity.TaskLog{
		TaskID:  taskID,
		UserID:  reviewerID,
		Action:  action,
		Content: req.Remark,
	}
	if err := s.taskRepo.AddLog(ctx, log); err != nil {
		logger.Warn("Failed to add follow up review log", logger.Err(err), logger.String("task_id", taskID))
	}
	return nil
}

// AddFollowUpComment 添加任务跟进评论
func (s *TaskAppService) AddFollowUpComment(ctx context.Context, taskID, followUpID string, req *dto.CreateTaskFollowUpCommentRequest, userID string, isManager bool) (*dto.TaskFollowUpCommentResponse, error) {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, ErrTaskNotFound
	}
	isCreator := task.CreatorID == userID
	isAssignee := task.AssigneeID != nil && *task.AssigneeID == userID
	if !isManager && !isCreator && !isAssignee {
		return nil, ErrTaskForbidden
	}

	if _, err := s.taskRepo.FindFollowUpByID(ctx, taskID, followUpID); err != nil {
		return nil, ErrTaskFollowUpNotFound
	}

	comment := &entity.TaskFollowUpComment{
		TaskID:     taskID,
		FollowUpID: followUpID,
		UserID:     userID,
		Content:    req.Content,
	}
	if err := s.taskRepo.AddFollowUpComment(ctx, comment); err != nil {
		return nil, err
	}

	log := &entity.TaskLog{
		TaskID:  taskID,
		UserID:  userID,
		Action:  "follow_up_comment",
		Content: "任务跟进评论",
	}
	if err := s.taskRepo.AddLog(ctx, log); err != nil {
		logger.Warn("Failed to add follow up comment log", logger.Err(err), logger.String("task_id", taskID))
	}

	resp := dto.ToTaskFollowUpCommentResponse(comment)
	return &resp, nil
}

// GetFollowUpComments 获取任务跟进评论列表
func (s *TaskAppService) GetFollowUpComments(ctx context.Context, taskID, followUpID string, page, pageSize int) (*dto.PageResult[dto.TaskFollowUpCommentResponse], error) {
	if _, err := s.taskRepo.FindByID(ctx, taskID); err != nil {
		return nil, ErrTaskNotFound
	}
	if _, err := s.taskRepo.FindFollowUpByID(ctx, taskID, followUpID); err != nil {
		return nil, ErrTaskFollowUpNotFound
	}

	page, pageSize = validator.SanitizePagination(page, pageSize)
	pagination := repository.Pagination{Page: page, PageSize: pageSize}
	result, err := s.taskRepo.GetFollowUpComments(ctx, taskID, followUpID, pagination)
	if err != nil {
		return nil, err
	}

	list := make([]dto.TaskFollowUpCommentResponse, len(result.List))
	for i, item := range result.List {
		list[i] = dto.ToTaskFollowUpCommentResponse(&item)
	}

	totalPages := int(result.Total) / pageSize
	if int(result.Total)%pageSize > 0 {
		totalPages++
	}
	return &dto.PageResult[dto.TaskFollowUpCommentResponse]{
		List:       list,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetMyTasks 获取我的任务
func (s *TaskAppService) GetMyTasks(ctx context.Context, userID string, status string, page, pageSize int) (*dto.TaskListResponse, error) {
	page, pageSize = validator.SanitizePagination(page, pageSize)
	query := repository.NewTaskQuery()
	query.Page = page
	query.PageSize = pageSize
	query.AssigneeID = userID
	query.Status = entity.TaskStatus(status)
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

// GetPendingTasks 获取待分配任务
func (s *TaskAppService) GetPendingTasks(ctx context.Context, page, pageSize int) (*dto.TaskListResponse, error) {
	page, pageSize = validator.SanitizePagination(page, pageSize)
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
	page, pageSize = validator.SanitizePagination(page, pageSize)
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
		Total:        stats.Total,
		Draft:        stats.Draft,
		Pending:      stats.Pending,
		Assigned:     stats.Assigned,
		Processing:   stats.Processing,
		Completed:    stats.Completed,
		Cancelled:    stats.Cancelled,
		Overdue:      stats.Overdue,
		MyTasks:      stats.MyTasks,
		MyPending:    stats.MyPending,
		MyProcessing: stats.MyProcessing,
		MyCompleted:  stats.MyCompleted,
	}, nil
}
