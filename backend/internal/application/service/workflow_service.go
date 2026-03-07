package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/google/uuid"
)

var (
	ErrWorkflowNotFound       = errors.New("workflow definition not found")
	ErrInstanceNotFound       = errors.New("workflow instance not found")
	ErrWorkflowTaskNotFound   = errors.New("workflow task not found")
	ErrWorkflowInvalidStatus  = errors.New("invalid workflow status")
	ErrInvalidTransition      = errors.New("invalid status transition")
	ErrNoActiveDefinition     = errors.New("no active workflow definition found")
	ErrInstanceAlreadyStarted = errors.New("instance already started")
	ErrCannotCancel           = errors.New("instance cannot be cancelled")
)

// WorkflowEngine 工作流引擎
type WorkflowEngine interface {
	// Definition management
	CreateDefinition(ctx context.Context, req *dto.CreateWorkflowDefinitionRequest) (*dto.WorkflowDefinitionResponse, error)
	PublishDefinition(ctx context.Context, id string) error
	ArchiveDefinition(ctx context.Context, id string) error
	GetDefinition(ctx context.Context, id string) (*dto.WorkflowDefinitionResponse, error)
	GetDefinitionByKey(ctx context.Context, key string, version int) (*dto.WorkflowDefinitionResponse, error)
	ListDefinitions(ctx context.Context, req *dto.ListWorkflowDefinitionsRequest) (*dto.ListWorkflowDefinitionsResponse, error)
	
	// Instance management
	StartInstance(ctx context.Context, req *dto.StartWorkflowInstanceRequest) (*dto.WorkflowInstanceResponse, error)
	CancelInstance(ctx context.Context, id string, reason string) error
	GetInstance(ctx context.Context, id string) (*dto.WorkflowInstanceResponse, error)
	ListInstances(ctx context.Context, req *dto.ListWorkflowInstancesRequest) (*dto.ListWorkflowInstancesResponse, error)
	ListMyInstances(ctx context.Context, userID string, page, pageSize int) (*dto.ListWorkflowInstancesResponse, error)
	ListMyTodoInstances(ctx context.Context, userID string, page, pageSize int) (*dto.ListWorkflowInstancesResponse, error)
	
	// Task operations
	ApproveTask(ctx context.Context, taskID string, req *dto.ApproveTaskRequest, userID string) error
	RejectTask(ctx context.Context, taskID string, req *dto.RejectTaskRequest, userID string) error
	TransferTask(ctx context.Context, taskID string, toUserID string, userID string) error
	DelegateTask(ctx context.Context, taskID string, toUserID string, req *dto.DelegateTaskRequest, userID string) error
	ReturnTask(ctx context.Context, taskID string, toNodeID string, comment string, userID string) error
	GetTask(ctx context.Context, id string) (*dto.WorkflowTaskResponse, error)
	ListTodoTasks(ctx context.Context, userID string, page, pageSize int) (*dto.ListWorkflowTasksResponse, error)
	ListDoneTasks(ctx context.Context, userID string, page, pageSize int) (*dto.ListWorkflowTasksResponse, error)
	
	// Reminder
	SendReminder(ctx context.Context, taskID string, userID string) error
	
	// Stats
	GetStats(ctx context.Context, orgID string) (*dto.WorkflowStatsResponse, error)
}

// WorkflowAppService 工作流应用服务
type WorkflowAppService struct {
	workflowRepo repository.WorkflowRepository
	taskRepo     repository.WorkflowTaskRepository
	userRepo     repository.UserRepository
}

// NewWorkflowAppService 创建工作流服务
func NewWorkflowAppService(
	workflowRepo repository.WorkflowRepository,
	taskRepo repository.WorkflowTaskRepository,
	userRepo repository.UserRepository,
) *WorkflowAppService {
	return &WorkflowAppService{
		workflowRepo: workflowRepo,
		taskRepo:     taskRepo,
		userRepo:     userRepo,
	}
}

// CreateDefinition 创建工作流定义
func (s *WorkflowAppService) CreateDefinition(ctx context.Context, req *dto.CreateWorkflowDefinitionRequest) (*dto.WorkflowDefinitionResponse, error) {
	// 检查key是否已存在
	latestVersion, err := s.workflowRepo.GetLatestVersion(ctx, req.Key)
	if err != nil {
		logger.Error("Failed to get latest version", logger.Err(err))
		return nil, err
	}
	
	def := entity.NewWorkflowDefinition(req.Name, req.Key, req.OrgID)
	def.Description = req.Description
	def.Category = req.Category
	def.Version = latestVersion + 1
	def.Config = req.Config
	def.FormSchema = req.FormSchema
	
	if err := s.workflowRepo.CreateDefinition(ctx, def); err != nil {
		logger.Error("Failed to create workflow definition", logger.Err(err))
		return nil, err
	}
	
	// 创建节点
	for i, nodeReq := range req.Nodes {
		node := &entity.WorkflowNode{
			BaseEntity: entity.BaseEntity{
				ID: nodeReq.ID,
			},
			WorkflowID:    def.ID,
			Name:          nodeReq.Name,
			Type:          entity.WorkflowNodeType(nodeReq.Type),
			Config:        nodeReq.Config,
			AssigneeType:  entity.AssigneeType(nodeReq.AssigneeType),
			Assignees:     nodeReq.Assignees,
			ApprovalMode:  entity.ApprovalMode(nodeReq.ApprovalMode),
			RequiredCount: nodeReq.RequiredCount,
			AllowTransfer: nodeReq.AllowTransfer,
			AllowDelegate: nodeReq.AllowDelegate,
			AutoPass:      nodeReq.AutoPass,
			Conditions:    nodeReq.Conditions,
			PositionX:     nodeReq.PositionX,
			PositionY:     nodeReq.PositionY,
			OrderIndex:    i,
		}
		
		if err := s.workflowRepo.CreateNode(ctx, node); err != nil {
			logger.Error("Failed to create workflow node", logger.Err(err))
			return nil, err
		}
		
		// 设置开始节点
		if nodeReq.Type == string(entity.NodeTypeStart) {
			def.StartNodeID = &node.ID
		}
	}
	
	// 更新开始节点
	if def.StartNodeID != nil {
		if err := s.workflowRepo.UpdateDefinition(ctx, def); err != nil {
			logger.Error("Failed to update workflow definition", logger.Err(err))
			return nil, err
		}
	}
	
	logger.Info("Workflow definition created", logger.String("id", def.ID), logger.String("key", def.Key))
	
	resp := dto.ToWorkflowDefinitionResponse(def)
	return &resp, nil
}

// PublishDefinition 发布工作流定义
func (s *WorkflowAppService) PublishDefinition(ctx context.Context, id string) error {
	def, err := s.workflowRepo.FindDefinitionByID(ctx, id)
	if err != nil {
		logger.Error("Failed to find workflow definition", logger.Err(err))
		return err
	}
	if def == nil {
		return ErrWorkflowNotFound
	}
	
	if !def.CanPublish() {
		return errors.New("workflow definition cannot be published")
	}
	
	if def.StartNodeID == nil {
		return errors.New("workflow definition has no start node")
	}
	
	def.Status = entity.WorkflowStatusActive
	if err := s.workflowRepo.UpdateDefinition(ctx, def); err != nil {
		logger.Error("Failed to publish workflow definition", logger.Err(err))
		return err
	}
	
	logger.Info("Workflow definition published", logger.String("id", id))
	return nil
}

// GetDefinition 获取工作流定义
func (s *WorkflowAppService) GetDefinition(ctx context.Context, id string) (*dto.WorkflowDefinitionResponse, error) {
	def, err := s.workflowRepo.FindDefinitionByID(ctx, id)
	if err != nil {
		logger.Error("Failed to find workflow definition", logger.Err(err))
		return nil, err
	}
	if def == nil {
		return nil, ErrWorkflowNotFound
	}
	
	resp := dto.ToWorkflowDefinitionResponse(def)
	return &resp, nil
}

// GetDefinitionByKey 根据Key获取工作流定义
func (s *WorkflowAppService) GetDefinitionByKey(ctx context.Context, key string, version int) (*dto.WorkflowDefinitionResponse, error) {
	var def *entity.WorkflowDefinition
	var err error
	
	if version == 0 {
		def, err = s.workflowRepo.FindActiveDefinitionByKey(ctx, key)
	} else {
		def, err = s.workflowRepo.FindDefinitionByKey(ctx, key, version)
	}
	
	if err != nil {
		logger.Error("Failed to find workflow definition", logger.Err(err))
		return nil, err
	}
	if def == nil {
		return nil, ErrWorkflowNotFound
	}
	
	resp := dto.ToWorkflowDefinitionResponse(def)
	return &resp, nil
}

// ListDefinitions 列表查询
func (s *WorkflowAppService) ListDefinitions(ctx context.Context, req *dto.ListWorkflowDefinitionsRequest) (*dto.ListWorkflowDefinitionsResponse, error) {
	defs, total, err := s.workflowRepo.ListDefinitions(ctx, req.OrgID, req.Category, req.Page, req.PageSize)
	if err != nil {
		logger.Error("Failed to list workflow definitions", logger.Err(err))
		return nil, err
	}
	
	items := make([]dto.WorkflowDefinitionResponse, len(defs))
	for i, def := range defs {
		items[i] = dto.ToWorkflowDefinitionResponse(def)
	}
	
	return &dto.ListWorkflowDefinitionsResponse{
		List:     items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// StartInstance 启动流程实例
func (s *WorkflowAppService) StartInstance(ctx context.Context, req *dto.StartWorkflowInstanceRequest) (*dto.WorkflowInstanceResponse, error) {
	// 查找工作流定义
	var def *entity.WorkflowDefinition
	var err error
	
	if req.DefinitionID != "" {
		def, err = s.workflowRepo.FindDefinitionByID(ctx, req.DefinitionID)
	} else if req.WorkflowKey != "" {
		def, err = s.workflowRepo.FindActiveDefinitionByKey(ctx, req.WorkflowKey)
	}
	
	if err != nil {
		logger.Error("Failed to find workflow definition", logger.Err(err))
		return nil, err
	}
	if def == nil {
		return nil, ErrWorkflowNotFound
	}
	
	if !def.IsActive() {
		return nil, errors.New("workflow definition is not active")
	}
	
	// 创建实例
	instance := entity.NewWorkflowInstance(def.ID, req.Title, req.StartedBy, req.OrgID)
	instance.BusinessKey = req.BusinessKey
	instance.BusinessID = req.BusinessID
	instance.Variables = req.Variables
	
	if err := s.workflowRepo.CreateInstance(ctx, instance); err != nil {
		logger.Error("Failed to create workflow instance", logger.Err(err))
		return nil, err
	}
	
	// 启动实例
	if err := instance.Start(); err != nil {
		logger.Error("Failed to start workflow instance", logger.Err(err))
		return nil, err
	}
	
	if err := s.workflowRepo.UpdateInstance(ctx, instance); err != nil {
		logger.Error("Failed to update workflow instance", logger.Err(err))
		return nil, err
	}
	
	// 创建第一个节点的任务
	if def.StartNodeID != nil {
		startNode := def.GetNode(*def.StartNodeID)
		if startNode != nil {
			if err := s.createTasksForNode(ctx, instance, startNode, req.StartedBy); err != nil {
				logger.Error("Failed to create tasks for start node", logger.Err(err))
				return nil, err
			}
			instance.CurrentNodeID = def.StartNodeID
			if err := s.workflowRepo.UpdateInstance(ctx, instance); err != nil {
				logger.Error("Failed to update instance current node", logger.Err(err))
				return nil, err
			}
		}
	}
	
	// 记录转换
	transition := &entity.WorkflowTransition{
		BaseEntity: entity.BaseEntity{
			ID: uuid.New().String(),
		},
		InstanceID: instance.ID,
		ToNodeID:   *instance.CurrentNodeID,
		Action:     "start",
		UserID:     req.StartedBy,
		Comment:    "流程启动",
		Variables:  req.Variables,
	}
	if err := s.workflowRepo.CreateTransition(ctx, transition); err != nil {
		logger.Error("Failed to create transition", logger.Err(err))
	}
	
	logger.Info("Workflow instance started", logger.String("id", instance.ID))
	
	resp := dto.ToWorkflowInstanceResponse(instance)
	return &resp, nil
}

// createTasksForNode 为节点创建任务
func (s *WorkflowAppService) createTasksForNode(ctx context.Context, instance *entity.WorkflowInstance, node *entity.WorkflowNode, starterID string) error {
	assigneeIDs, err := s.resolveAssignees(ctx, node, instance, starterID)
	if err != nil {
		return err
	}
	
	for i, assigneeID := range assigneeIDs {
		task := &entity.WorkflowTask{
			Task: entity.Task{
				BaseEntity: entity.BaseEntity{
					ID: uuid.New().String(),
				},
				Title:       node.Name,
				Description: fmt.Sprintf("%s - %s", instance.Title, node.Name),
				Type:        entity.TaskTypeVerify,
				Status:      entity.TaskStatusPending,
				CreatorID:   instance.StartedBy,
				OrgID:       instance.OrgID,
			},
			WorkflowInstanceID: instance.ID,
			WorkflowNodeID:     node.ID,
			WorkflowNodeType:   node.Type,
			SequentialIndex:    i,
		}
		
		// 根据处理人类型设置处理人
		if assigneeID != "" {
			task.AssigneeID = &assigneeID
		}
		
		// 设置截止时间
		if node.Config != nil {
			if timeoutConfig, ok := node.Config["timeout"].(map[string]interface{}); ok {
				duration := 0
				unit := "hour"
				if d, ok := timeoutConfig["duration"].(int); ok {
					duration = d
				}
				if u, ok := timeoutConfig["unit"].(string); ok {
					unit = u
				}
				
				if duration > 0 {
					dueTime := time.Now()
					if unit == "day" {
						dueTime = dueTime.AddDate(0, 0, duration)
					} else {
						dueTime = dueTime.Add(time.Duration(duration) * time.Hour)
					}
					task.DueTime = &dueTime
				}
			}
		}
		
		if err := s.taskRepo.Create(ctx, task); err != nil {
			return err
		}
	}
	
	return nil
}

// resolveAssignees 解析处理人
func (s *WorkflowAppService) resolveAssignees(ctx context.Context, node *entity.WorkflowNode, instance *entity.WorkflowInstance, starterID string) ([]string, error) {
	switch node.AssigneeType {
	case entity.AssigneeTypeStarter:
		return []string{starterID}, nil
		
	case entity.AssigneeTypeUser:
		if node.Assignees != nil {
			if users, ok := node.Assignees["users"].([]string); ok {
				return users, nil
			}
			if users, ok := node.Assignees["users"].([]interface{}); ok {
				result := make([]string, len(users))
				for i, u := range users {
					result[i] = fmt.Sprint(u)
				}
				return result, nil
			}
		}
		return nil, errors.New("no users specified in assignees")
		
	case entity.AssigneeTypeRole:
		// 查询角色下的用户
		if node.Assignees != nil {
			if roles, ok := node.Assignees["roles"].([]interface{}); ok && len(roles) > 0 {
				_ = fmt.Sprint(roles[0]) // roleName reserved for future
				// 这里应该根据角色查询用户列表
				// 简化处理，返回空让用户自己去查询
				return []string{""}, nil
			}
		}
		return nil, errors.New("no roles specified in assignees")
		
	default:
		return []string{""}, nil
	}
}

// ApproveTask 审批通过
func (s *WorkflowAppService) ApproveTask(ctx context.Context, taskID string, req *dto.ApproveTaskRequest, userID string) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		logger.Error("Failed to find task", logger.Err(err))
		return err
	}
	if task == nil {
		return ErrTaskNotFound
	}
	
	if !task.CanApprove() {
		return errors.New("task cannot be approved")
	}
	
	// 检查权限
	if task.AssigneeID == nil || *task.AssigneeID != userID {
		// 检查是否有委托
		delegation, err := s.taskRepo.FindActiveDelegation(ctx, taskID)
		if err != nil {
			return err
		}
		if delegation == nil || delegation.ToUserID != userID {
			return errors.New("you are not the assignee of this task")
		}
	}
	
	// 完成任务
	now := time.Now()
	task.Status = entity.TaskStatusCompleted
	task.CompletedAt = &now
	task.ApprovalAction = "approve"
	task.ApprovalComment = req.Comment
	task.Result = "approved"
	
	if err := s.taskRepo.Update(ctx, task); err != nil {
		logger.Error("Failed to update task", logger.Err(err))
		return err
	}
	
	// 记录转换
	transition := &entity.WorkflowTransition{
		BaseEntity: entity.BaseEntity{
			ID: uuid.New().String(),
		},
		InstanceID: task.WorkflowInstanceID,
		ToNodeID:   task.WorkflowNodeID,
		Action:     "approve",
		UserID:     userID,
		Comment:    req.Comment,
		Variables:  req.Variables,
	}
	if err := s.workflowRepo.CreateTransition(ctx, transition); err != nil {
		logger.Error("Failed to create transition", logger.Err(err))
	}
	
	// 检查是否需要推进到下一个节点
	if err := s.processNodeCompletion(ctx, task); err != nil {
		logger.Error("Failed to process node completion", logger.Err(err))
		return err
	}
	
	logger.Info("Task approved", logger.String("task_id", taskID), logger.String("user_id", userID))
	return nil
}

// RejectTask 审批拒绝
func (s *WorkflowAppService) RejectTask(ctx context.Context, taskID string, req *dto.RejectTaskRequest, userID string) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		logger.Error("Failed to find task", logger.Err(err))
		return err
	}
	if task == nil {
		return ErrTaskNotFound
	}
	
	if !task.CanApprove() {
		return errors.New("task cannot be rejected")
	}
	
	// 完成任务
	now := time.Now()
	task.Status = entity.TaskStatusCompleted
	task.CompletedAt = &now
	task.ApprovalAction = "reject"
	task.ApprovalComment = req.Comment
	task.Result = "rejected"
	
	if err := s.taskRepo.Update(ctx, task); err != nil {
		logger.Error("Failed to update task", logger.Err(err))
		return err
	}
	
	// 更新流程实例为拒绝
	instance, err := s.workflowRepo.FindInstanceByID(ctx, task.WorkflowInstanceID)
	if err != nil {
		return err
	}
	if instance != nil {
		instance.Complete("rejected")
		if err := s.workflowRepo.UpdateInstance(ctx, instance); err != nil {
			return err
		}
	}
	
	// 记录转换
	transition := &entity.WorkflowTransition{
		BaseEntity: entity.BaseEntity{
			ID: uuid.New().String(),
		},
		InstanceID: task.WorkflowInstanceID,
		ToNodeID:   task.WorkflowNodeID,
		Action:     "reject",
		UserID:     userID,
		Comment:    req.Comment,
	}
	if err := s.workflowRepo.CreateTransition(ctx, transition); err != nil {
		logger.Error("Failed to create transition", logger.Err(err))
	}
	
	logger.Info("Task rejected", logger.String("task_id", taskID))
	return nil
}

// processNodeCompletion 处理节点完成
func (s *WorkflowAppService) processNodeCompletion(ctx context.Context, completedTask *entity.WorkflowTask) error {
	instance, err := s.workflowRepo.FindInstanceByID(ctx, completedTask.WorkflowInstanceID)
	if err != nil {
		return err
	}
	if instance == nil {
		return ErrInstanceNotFound
	}
	
	// 如果是并行审批，检查是否所有人都完成了
	node := instance.Definition.GetNode(completedTask.WorkflowNodeID)
	if node == nil {
		return errors.New("node not found")
	}
	
	// 查找所有活跃任务
	activeTasks, err := s.taskRepo.FindActiveTasksByInstance(ctx, instance.ID)
	if err != nil {
		return err
	}
	
	// 如果还有活跃任务，不推进
	if len(activeTasks) > 0 {
		return nil
	}
	
	// 根据节点配置决定下一步
	nextNode := s.findNextNode(instance.Definition, node)
	if nextNode != nil {
		// 创建下一个节点的任务
		if err := s.createTasksForNode(ctx, instance, nextNode, instance.StartedBy); err != nil {
			return err
		}
		
		instance.CurrentNodeID = &nextNode.ID
		instance.Status = entity.InstanceStatusProcessing
		if err := s.workflowRepo.UpdateInstance(ctx, instance); err != nil {
			return err
		}
		
		// 记录转换
		assigneeID := ""
		if completedTask.AssigneeID != nil {
			assigneeID = *completedTask.AssigneeID
		}
		transition := &entity.WorkflowTransition{
			BaseEntity: entity.BaseEntity{
				ID: uuid.New().String(),
			},
			InstanceID: instance.ID,
			FromNodeID: &node.ID,
			ToNodeID:   nextNode.ID,
			Action:     "next",
			UserID:     assigneeID,
			Variables:  make(entity.JSONMap),
		}
		if err := s.workflowRepo.CreateTransition(ctx, transition); err != nil {
			logger.Error("Failed to create transition", logger.Err(err))
		}
	} else {
		// 没有下一个节点，流程结束
		instance.Complete("approved")
		if err := s.workflowRepo.UpdateInstance(ctx, instance); err != nil {
			return err
		}
	}
	
	return nil
}

// findNextNode 查找下一个节点
func (s *WorkflowAppService) findNextNode(def *entity.WorkflowDefinition, currentNode *entity.WorkflowNode) *entity.WorkflowNode {
	// 简化实现：按顺序查找下一个节点
	// 实际应该根据节点配置和条件判断
	
	for i, node := range def.Nodes {
		if node.ID == currentNode.ID {
			if i+1 < len(def.Nodes) {
				nextNode := def.Nodes[i+1]
				if nextNode.Type != entity.NodeTypeEnd {
					return &nextNode
				}
			}
			break
		}
	}
	
	return nil
}

// ListTodoTasks 查询待办任务
func (s *WorkflowAppService) ListTodoTasks(ctx context.Context, userID string, page, pageSize int) (*dto.ListWorkflowTasksResponse, error) {
	tasks, total, err := s.taskRepo.FindTodoTasks(ctx, userID, page, pageSize)
	if err != nil {
		logger.Error("Failed to find todo tasks", logger.Err(err))
		return nil, err
	}
	
	items := make([]dto.WorkflowTaskResponse, len(tasks))
	for i, task := range tasks {
		items[i] = dto.ToWorkflowTaskResponse(task)
	}
	
	return &dto.ListWorkflowTasksResponse{
		List:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ListDoneTasks 查询已办任务
func (s *WorkflowAppService) ListDoneTasks(ctx context.Context, userID string, page, pageSize int) (*dto.ListWorkflowTasksResponse, error) {
	tasks, total, err := s.taskRepo.FindDoneTasks(ctx, userID, page, pageSize)
	if err != nil {
		logger.Error("Failed to find done tasks", logger.Err(err))
		return nil, err
	}
	
	items := make([]dto.WorkflowTaskResponse, len(tasks))
	for i, task := range tasks {
		items[i] = dto.ToWorkflowTaskResponse(task)
	}
	
	return &dto.ListWorkflowTasksResponse{
		List:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetInstance 获取流程实例
func (s *WorkflowAppService) GetInstance(ctx context.Context, id string) (*dto.WorkflowInstanceResponse, error) {
	instance, err := s.workflowRepo.FindInstanceByID(ctx, id)
	if err != nil {
		logger.Error("Failed to find instance", logger.Err(err))
		return nil, err
	}
	if instance == nil {
		return nil, ErrInstanceNotFound
	}
	
	resp := dto.ToWorkflowInstanceResponse(instance)
	return &resp, nil
}

// ListInstances 列表查询流程实例
func (s *WorkflowAppService) ListInstances(ctx context.Context, req *dto.ListWorkflowInstancesRequest) (*dto.ListWorkflowInstancesResponse, error) {
	query := &entity.WorkflowQuery{
		DefinitionID: req.DefinitionID,
		Status:       entity.WorkflowInstanceStatus(req.Status),
		StartedBy:    req.StartedBy,
		OrgID:        req.OrgID,
		Page:         req.Page,
		PageSize:     req.PageSize,
	}
	
	instances, total, err := s.workflowRepo.ListInstances(ctx, query)
	if err != nil {
		logger.Error("Failed to list instances", logger.Err(err))
		return nil, err
	}
	
	items := make([]dto.WorkflowInstanceResponse, len(instances))
	for i, instance := range instances {
		items[i] = dto.ToWorkflowInstanceResponse(instance)
	}
	
	return &dto.ListWorkflowInstancesResponse{
		List:     items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetStats 获取统计
func (s *WorkflowAppService) GetStats(ctx context.Context, orgID string) (*dto.WorkflowStatsResponse, error) {
	stats, err := s.workflowRepo.GetInstanceStats(ctx, orgID)
	if err != nil {
		logger.Error("Failed to get workflow stats", logger.Err(err))
		return nil, err
	}
	
	return &dto.WorkflowStatsResponse{
		TotalInstances:  stats.TotalInstances,
		PendingCount:    stats.PendingCount,
		ProcessingCount: stats.ProcessingCount,
		ApprovedCount:   stats.ApprovedCount,
		RejectedCount:   stats.RejectedCount,
		CancelledCount:  stats.CancelledCount,
	}, nil
}
// GetTask 获取任务
func (s *WorkflowAppService) GetTask(ctx context.Context, id string) (*dto.WorkflowTaskResponse, error) {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		logger.Error("Failed to find task", logger.Err(err))
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}
	
	resp := dto.ToWorkflowTaskResponse(task)
	return &resp, nil
}

// ArchiveDefinition 归档工作流定义
func (s *WorkflowAppService) ArchiveDefinition(ctx context.Context, id string) error {
	def, err := s.workflowRepo.FindDefinitionByID(ctx, id)
	if err != nil {
		logger.Error("Failed to find workflow definition", logger.Err(err))
		return err
	}
	if def == nil {
		return ErrWorkflowNotFound
	}
	
	if !def.CanArchive() {
		return errors.New("workflow definition cannot be archived")
	}
	
	def.Status = entity.WorkflowStatusArchived
	if err := s.workflowRepo.UpdateDefinition(ctx, def); err != nil {
		logger.Error("Failed to archive workflow definition", logger.Err(err))
		return err
	}
	
	logger.Info("Workflow definition archived", logger.String("id", id))
	return nil
}

// CancelInstance 取消流程实例
func (s *WorkflowAppService) CancelInstance(ctx context.Context, id string, reason string) error {
	instance, err := s.workflowRepo.FindInstanceByID(ctx, id)
	if err != nil {
		logger.Error("Failed to find instance", logger.Err(err))
		return err
	}
	if instance == nil {
		return ErrInstanceNotFound
	}
	
	if err := instance.Cancel(reason); err != nil {
		return err
	}
	
	if err := s.workflowRepo.UpdateInstance(ctx, instance); err != nil {
		logger.Error("Failed to cancel instance", logger.Err(err))
		return err
	}
	
	logger.Info("Workflow instance cancelled", logger.String("id", id))
	return nil
}

// TransferTask 转办任务
func (s *WorkflowAppService) TransferTask(ctx context.Context, taskID string, toUserID string, userID string) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		logger.Error("Failed to find task", logger.Err(err))
		return err
	}
	if task == nil {
		return ErrTaskNotFound
	}
	
	if !task.CanApprove() {
		return errors.New("task cannot be transferred")
	}
	
	// 检查权限
	if task.AssigneeID == nil || *task.AssigneeID != userID {
		return errors.New("you are not the assignee of this task")
	}
	
	// 记录转办
	task.TransferredFrom = task.AssigneeID
	task.AssigneeID = &toUserID
	task.ApprovalAction = "transfer"
	
	if err := s.taskRepo.Update(ctx, task); err != nil {
		logger.Error("Failed to transfer task", logger.Err(err))
		return err
	}
	
	logger.Info("Task transferred", logger.String("task_id", taskID), logger.String("to_user_id", toUserID))
	return nil
}

// DelegateTask 委托任务
func (s *WorkflowAppService) DelegateTask(ctx context.Context, taskID string, toUserID string, req *dto.DelegateTaskRequest, userID string) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		logger.Error("Failed to find task", logger.Err(err))
		return err
	}
	if task == nil {
		return ErrTaskNotFound
	}
	
	if !task.CanApprove() {
		return errors.New("task cannot be delegated")
	}
	
	// 检查权限
	if task.AssigneeID == nil || *task.AssigneeID != userID {
		return errors.New("you are not the assignee of this task")
	}
	
	// 创建委托记录
	delegation := &entity.WorkflowDelegation{
		BaseEntity: entity.BaseEntity{
			ID: uuid.New().String(),
		},
		TaskID:     taskID,
		FromUserID: userID,
		ToUserID:   toUserID,
		StartTime:  time.Now(),
		EndTime:    req.EndTime,
		Reason:     req.Reason,
		Status:     "active",
	}
	
	if err := s.taskRepo.CreateDelegation(ctx, delegation); err != nil {
		logger.Error("Failed to create delegation", logger.Err(err))
		return err
	}
	
	// 更新任务
	task.DelegateFrom = &userID
	task.AssigneeID = &toUserID
	task.ApprovalAction = "delegate"
	
	if err := s.taskRepo.Update(ctx, task); err != nil {
		logger.Error("Failed to update task", logger.Err(err))
		return err
	}
	
	logger.Info("Task delegated", logger.String("task_id", taskID), logger.String("to_user_id", toUserID))
	return nil
}

// ReturnTask 退回任务
func (s *WorkflowAppService) ReturnTask(ctx context.Context, taskID string, toNodeID string, comment string, userID string) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		logger.Error("Failed to find task", logger.Err(err))
		return err
	}
	if task == nil {
		return ErrTaskNotFound
	}
	
	if !task.CanApprove() {
		return errors.New("task cannot be returned")
	}
	
	instance, err := s.workflowRepo.FindInstanceByID(ctx, task.WorkflowInstanceID)
	if err != nil {
		return err
	}
	if instance == nil {
		return ErrInstanceNotFound
	}
	
	// 完成当前任务
	task.Status = entity.TaskStatusCompleted
	task.ApprovalAction = "return"
	task.ApprovalComment = comment
	now := time.Now()
	task.CompletedAt = &now
	
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return err
	}
	
	// 更新实例状态
	instance.Status = entity.InstanceStatusReturned
	if err := s.workflowRepo.UpdateInstance(ctx, instance); err != nil {
		return err
	}
	
	// 记录转换
	transition := &entity.WorkflowTransition{
		BaseEntity: entity.BaseEntity{
			ID: uuid.New().String(),
		},
		InstanceID: instance.ID,
		FromNodeID: &task.WorkflowNodeID,
		ToNodeID:   toNodeID,
		Action:     "return",
		UserID:     userID,
		Comment:    comment,
	}
	if err := s.workflowRepo.CreateTransition(ctx, transition); err != nil {
		logger.Error("Failed to create transition", logger.Err(err))
	}
	
	logger.Info("Task returned", logger.String("task_id", taskID))
	return nil
}

// SendReminder 发送催办
func (s *WorkflowAppService) SendReminder(ctx context.Context, taskID string, userID string) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		logger.Error("Failed to find task", logger.Err(err))
		return err
	}
	if task == nil {
		return ErrTaskNotFound
	}
	
	// 查找催办记录
	reminder, err := s.taskRepo.FindReminderByTaskID(ctx, taskID)
	if err != nil {
		return err
	}
	
	now := time.Now()
	if reminder == nil {
		reminder = &entity.WorkflowReminder{
			BaseEntity: entity.BaseEntity{
				ID: uuid.New().String(),
			},
			TaskID:       taskID,
			ReminderType: "manual",
			RemindCount:  1,
			LastRemindAt: &now,
			Status:       "pending",
		}
		if err := s.taskRepo.CreateReminder(ctx, reminder); err != nil {
			return err
		}
	} else {
		reminder.RemindCount++
		reminder.LastRemindAt = &now
		if err := s.taskRepo.UpdateReminder(ctx, reminder); err != nil {
			return err
		}
	}
	
	// 更新任务催办信息
	task.RemindedAt = &now
	task.RemindCount = reminder.RemindCount
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return err
	}
	
	logger.Info("Reminder sent", logger.String("task_id", taskID))
	return nil
}

// ListMyInstances 查询我的流程
func (s *WorkflowAppService) ListMyInstances(ctx context.Context, userID string, page, pageSize int) (*dto.ListWorkflowInstancesResponse, error) {
	req := &dto.ListWorkflowInstancesRequest{
		StartedBy: userID,
		Page:      page,
		PageSize:  pageSize,
	}
	return s.ListInstances(ctx, req)
}

// ListMyTodoInstances 查询我的待办流程
func (s *WorkflowAppService) ListMyTodoInstances(ctx context.Context, userID string, page, pageSize int) (*dto.ListWorkflowInstancesResponse, error) {
	instances, total, err := s.workflowRepo.ListMyTodoInstances(ctx, userID, page, pageSize)
	if err != nil {
		logger.Error("Failed to list my todo instances", logger.Err(err))
		return nil, err
	}
	
	items := make([]dto.WorkflowInstanceResponse, len(instances))
	for i, instance := range instances {
		items[i] = dto.ToWorkflowInstanceResponse(instance)
	}
	
	return &dto.ListWorkflowInstancesResponse{
		List:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
