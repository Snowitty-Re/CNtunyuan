package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/model"
	"github.com/Snowitty-Re/CNtunyuan/internal/repository"
	"github.com/google/uuid"
)

// WorkflowService 工作流服务
type WorkflowService struct {
	workflowRepo *repository.WorkflowRepository
	userRepo     *repository.UserRepository
}

// NewWorkflowService 创建工作流服务
func NewWorkflowService(
	workflowRepo *repository.WorkflowRepository,
	userRepo *repository.UserRepository,
) *WorkflowService {
	return &WorkflowService{
		workflowRepo: workflowRepo,
		userRepo:     userRepo,
	}
}

// CreateWorkflowRequest 创建工作流请求
type CreateWorkflowRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type" binding:"required"`
}

// CreateWorkflow 创建工作流
func (s *WorkflowService) CreateWorkflow(ctx context.Context, creatorID uuid.UUID, req *CreateWorkflowRequest) (*model.Workflow, error) {
	// 检查编码是否已存在
	_, err := s.workflowRepo.GetByCode(ctx, req.Code)
	if err == nil {
		return nil, fmt.Errorf("工作流编码已存在")
	}

	workflow := &model.Workflow{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Type:        req.Type,
		Status:      model.WorkflowStatusDraft,
		CreatorID:   creatorID,
	}

	if err := s.workflowRepo.Create(ctx, workflow); err != nil {
		return nil, fmt.Errorf("创建工作流失败: %w", err)
	}

	return workflow, nil
}

// GetWorkflow 获取工作流详情
func (s *WorkflowService) GetWorkflow(ctx context.Context, id uuid.UUID) (*model.Workflow, error) {
	workflow, err := s.workflowRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("工作流不存在")
	}
	return workflow, nil
}

// ListWorkflows 获取工作流列表
type ListWorkflowsRequest struct {
	Status   string `form:"status"`
	Type     string `form:"type"`
	Keyword  string `form:"keyword"`
	Page     int    `form:"page" default:"1"`
	PageSize int    `form:"page_size" default:"20"`
}

func (s *WorkflowService) ListWorkflows(ctx context.Context, req *ListWorkflowsRequest) ([]model.Workflow, int64, error) {
	params := map[string]interface{}{
		"page":      req.Page,
		"page_size": req.PageSize,
	}
	if req.Status != "" {
		params["status"] = req.Status
	}
	if req.Type != "" {
		params["type"] = req.Type
	}
	if req.Keyword != "" {
		params["keyword"] = req.Keyword
	}

	return s.workflowRepo.List(ctx, params)
}

// UpdateWorkflow 更新工作流
type UpdateWorkflowRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	IsDefault   *bool  `json:"is_default"`
}

func (s *WorkflowService) UpdateWorkflow(ctx context.Context, id uuid.UUID, req *UpdateWorkflowRequest) (*model.Workflow, error) {
	workflow, err := s.workflowRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("工作流不存在")
	}

	// 更新字段
	if req.Name != "" {
		workflow.Name = req.Name
	}
	if req.Description != "" {
		workflow.Description = req.Description
	}
	if req.Type != "" {
		workflow.Type = req.Type
	}
	if req.Status != "" {
		workflow.Status = req.Status
	}
	if req.IsDefault != nil {
		workflow.IsDefault = *req.IsDefault
	}

	if err := s.workflowRepo.Update(ctx, workflow); err != nil {
		return nil, fmt.Errorf("更新工作流失败: %w", err)
	}

	return workflow, nil
}

// DeleteWorkflow 删除工作流
func (s *WorkflowService) DeleteWorkflow(ctx context.Context, id uuid.UUID) error {
	_, err := s.workflowRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("工作流不存在")
	}

	return s.workflowRepo.Delete(ctx, id)
}

// CreateStepRequest 创建步骤请求
type CreateStepRequest struct {
	Name          string                 `json:"name" binding:"required"`
	Description   string                 `json:"description"`
	StepType      string                 `json:"step_type" binding:"required"`
	AssigneeType  string                 `json:"assignee_type"`
	AssigneeRole  string                 `json:"assignee_role"`
	Duration      int                    `json:"duration"`
	IsTimeoutSkip bool                   `json:"is_timeout_skip"`
	FormConfig    map[string]interface{} `json:"form_config"`
	Conditions    map[string]interface{} `json:"conditions"`
	Actions       map[string]interface{} `json:"actions"`
}

// CreateStep 创建工作流步骤
func (s *WorkflowService) CreateStep(ctx context.Context, workflowID uuid.UUID, req *CreateStepRequest) (*model.WorkflowStep, error) {
	// 验证工作流
	workflow, err := s.workflowRepo.GetByID(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("工作流不存在")
	}

	// 获取当前最大顺序
	steps, err := s.workflowRepo.GetStepsByWorkflowID(ctx, workflowID)
	if err != nil {
		return nil, err
	}
	maxOrder := 0
	for _, step := range steps {
		if step.StepOrder > maxOrder {
			maxOrder = step.StepOrder
		}
	}

	step := &model.WorkflowStep{
		WorkflowID:    workflowID,
		Name:          req.Name,
		Description:   req.Description,
		StepOrder:     maxOrder + 1,
		StepType:      req.StepType,
		AssigneeType:  req.AssigneeType,
		AssigneeRole:  req.AssigneeRole,
		Duration:      req.Duration,
		IsTimeoutSkip: req.IsTimeoutSkip,
	}

	// 序列化JSON字段
	if req.FormConfig != nil {
		formConfigJSON, _ := json.Marshal(req.FormConfig)
		step.FormConfig = string(formConfigJSON)
	}
	if req.Conditions != nil {
		conditionsJSON, _ := json.Marshal(req.Conditions)
		step.Conditions = string(conditionsJSON)
	}
	if req.Actions != nil {
		actionsJSON, _ := json.Marshal(req.Actions)
		step.Actions = string(actionsJSON)
	}

	if err := s.workflowRepo.CreateStep(ctx, step); err != nil {
		return nil, fmt.Errorf("创建步骤失败: %w", err)
	}

	// 更新工作流版本
	workflow.Version++
	s.workflowRepo.Update(ctx, workflow)

	return step, nil
}

// UpdateStep 更新步骤
func (s *WorkflowService) UpdateStep(ctx context.Context, stepID uuid.UUID, req *CreateStepRequest) (*model.WorkflowStep, error) {
	step, err := s.workflowRepo.GetStepByID(ctx, stepID)
	if err != nil {
		return nil, fmt.Errorf("步骤不存在")
	}

	// 更新字段
	if req.Name != "" {
		step.Name = req.Name
	}
	if req.Description != "" {
		step.Description = req.Description
	}
	if req.StepType != "" {
		step.StepType = req.StepType
	}
	if req.AssigneeType != "" {
		step.AssigneeType = req.AssigneeType
	}
	if req.AssigneeRole != "" {
		step.AssigneeRole = req.AssigneeRole
	}
	step.Duration = req.Duration
	step.IsTimeoutSkip = req.IsTimeoutSkip

	// 序列化JSON字段
	if req.FormConfig != nil {
		formConfigJSON, _ := json.Marshal(req.FormConfig)
		step.FormConfig = string(formConfigJSON)
	}
	if req.Conditions != nil {
		conditionsJSON, _ := json.Marshal(req.Conditions)
		step.Conditions = string(conditionsJSON)
	}
	if req.Actions != nil {
		actionsJSON, _ := json.Marshal(req.Actions)
		step.Actions = string(actionsJSON)
	}

	if err := s.workflowRepo.UpdateStep(ctx, step); err != nil {
		return nil, fmt.Errorf("更新步骤失败: %w", err)
	}

	return step, nil
}

// DeleteStep 删除步骤
func (s *WorkflowService) DeleteStep(ctx context.Context, stepID uuid.UUID) error {
	_, err := s.workflowRepo.GetStepByID(ctx, stepID)
	if err != nil {
		return fmt.Errorf("步骤不存在")
	}

	return s.workflowRepo.DeleteStep(ctx, stepID)
}

// ReorderSteps 重新排序步骤
func (s *WorkflowService) ReorderSteps(ctx context.Context, workflowID uuid.UUID, stepIDs []string) error {
	// 验证工作流
	_, err := s.workflowRepo.GetByID(ctx, workflowID)
	if err != nil {
		return fmt.Errorf("工作流不存在")
	}

	// 转换ID
	ids := make([]uuid.UUID, 0, len(stepIDs))
	for _, idStr := range stepIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}

	return s.workflowRepo.ReorderSteps(ctx, workflowID, ids)
}

// StartInstanceRequest 启动实例请求
type StartInstanceRequest struct {
	WorkflowID   string `json:"workflow_id" binding:"required"`
	BusinessID   string `json:"business_id" binding:"required"`
	BusinessType string `json:"business_type" binding:"required"`
	Title        string `json:"title" binding:"required"`
	FormData     map[string]interface{} `json:"form_data"`
}

// StartInstance 启动工作流实例
func (s *WorkflowService) StartInstance(ctx context.Context, starterID uuid.UUID, req *StartInstanceRequest) (*model.WorkflowInstance, error) {
	// 验证工作流
	workflowID, err := uuid.Parse(req.WorkflowID)
	if err != nil {
		return nil, fmt.Errorf("无效的工作流ID")
	}

	workflow, err := s.workflowRepo.GetByID(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("工作流不存在")
	}

	if workflow.Status != model.WorkflowStatusActive {
		return nil, fmt.Errorf("工作流未激活")
	}

	// 检查是否有步骤
	if len(workflow.Steps) == 0 {
		return nil, fmt.Errorf("工作流没有定义步骤")
	}

	// 解析业务ID
	businessID, err := uuid.Parse(req.BusinessID)
	if err != nil {
		return nil, fmt.Errorf("无效的业务ID")
	}

	// 创建实例
	instance := &model.WorkflowInstance{
		WorkflowID:   workflowID,
		BusinessID:   businessID,
		BusinessType: req.BusinessType,
		Title:        req.Title,
		Status:       "running",
		CurrentStepID: &workflow.Steps[0].ID,
		StarterID:    starterID,
		StartTime:    time.Now(),
	}

	if err := s.workflowRepo.CreateInstance(ctx, instance); err != nil {
		return nil, fmt.Errorf("创建工作流实例失败: %w", err)
	}

	// 创建第一条历史记录
	formDataJSON, _ := json.Marshal(req.FormData)
	history := &model.WorkflowHistory{
		InstanceID: instance.ID,
		StepID:     workflow.Steps[0].ID,
		StepName:   workflow.Steps[0].Name,
		OperatorID: starterID,
		Action:     "start",
		Comment:    "启动工作流",
		FormData:   string(formDataJSON),
		StartTime:  time.Now(),
	}
	s.workflowRepo.CreateHistory(ctx, history)

	return instance, nil
}

// GetInstance 获取实例详情
func (s *WorkflowService) GetInstance(ctx context.Context, id uuid.UUID) (*model.WorkflowInstance, error) {
	instance, err := s.workflowRepo.GetInstanceByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("实例不存在")
	}
	return instance, nil
}

// ListInstances 获取实例列表
type ListInstancesRequest struct {
	Status       string `form:"status"`
	WorkflowID   string `form:"workflow_id"`
	StarterID    string `form:"starter_id"`
	BusinessType string `form:"business_type"`
	Page         int    `form:"page" default:"1"`
	PageSize     int    `form:"page_size" default:"20"`
}

func (s *WorkflowService) ListInstances(ctx context.Context, req *ListInstancesRequest) ([]model.WorkflowInstance, int64, error) {
	params := map[string]interface{}{
		"page":      req.Page,
		"page_size": req.PageSize,
	}
	if req.Status != "" {
		params["status"] = req.Status
	}
	if req.WorkflowID != "" {
		params["workflow_id"] = req.WorkflowID
	}
	if req.StarterID != "" {
		params["starter_id"] = req.StarterID
	}
	if req.BusinessType != "" {
		params["business_type"] = req.BusinessType
	}

	return s.workflowRepo.ListInstances(ctx, params)
}

// ApproveRequest 审批请求
type ApproveRequest struct {
	Action      string                 `json:"action" binding:"required"` // approve, reject, transfer
	Comment     string                 `json:"comment"`
	FormData    map[string]interface{} `json:"form_data"`
	NextStepID  *string                `json:"next_step_id"`
	TransferTo  *string                `json:"transfer_to"`
}

// Approve 审批
func (s *WorkflowService) Approve(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID, req *ApproveRequest) error {
	// 获取实例
	instance, err := s.workflowRepo.GetInstanceByID(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("实例不存在")
	}

	if instance.IsCompleted() {
		return fmt.Errorf("工作流已结束")
	}

	// 获取工作流
	workflow, err := s.workflowRepo.GetByID(ctx, instance.WorkflowID)
	if err != nil {
		return fmt.Errorf("工作流不存在")
	}

	// 获取当前步骤
	currentStep := instance.CurrentStep
	if currentStep == nil {
		return fmt.Errorf("当前步骤不存在")
	}

	// 验证操作权限（简化版，实际应该根据角色判断）
	// TODO: 验证操作人是否有权限审批当前步骤

	// 创建历史记录
	formDataJSON, _ := json.Marshal(req.FormData)
	history := &model.WorkflowHistory{
		InstanceID: instance.ID,
		StepID:     currentStep.ID,
		StepName:   currentStep.Name,
		OperatorID: operatorID,
		Action:     req.Action,
		Comment:    req.Comment,
		FormData:   string(formDataJSON),
		EndTime:    &[]time.Time{time.Now()}[0],
	}

	now := time.Now()
	history.EndTime = &now

	// 根据操作处理
	switch req.Action {
	case "approve":
		// 找到下一步
		var nextStep *model.WorkflowStep
		if req.NextStepID != nil {
			nextStepID, _ := uuid.Parse(*req.NextStepID)
			for _, step := range workflow.Steps {
				if step.ID == nextStepID {
					nextStep = &step
					break
				}
			}
		} else {
			// 默认下一步
			for _, step := range workflow.Steps {
				if step.StepOrder > currentStep.StepOrder {
					if nextStep == nil || step.StepOrder < nextStep.StepOrder {
						nextStep = &step
					}
				}
			}
		}

		if nextStep != nil {
			instance.CurrentStepID = &nextStep.ID
			history.Action = "approve"
		} else {
			// 没有下一步，完成工作流
			instance.Status = "completed"
			instance.EndTime = &now
			history.Action = "complete"
		}

	case "reject":
		instance.Status = "rejected"
		instance.EndTime = &now
		history.Action = "reject"

	case "transfer":
		// 转派给其他人
		if req.TransferTo == nil {
			return fmt.Errorf("转派目标不能为空")
		}
		// TODO: 实现转派逻辑
		history.Action = "transfer"

	default:
		return fmt.Errorf("无效的操作")
	}

	// 保存历史
	if err := s.workflowRepo.CreateHistory(ctx, history); err != nil {
		return fmt.Errorf("保存历史记录失败: %w", err)
	}

	// 更新实例
	if err := s.workflowRepo.UpdateInstance(ctx, instance); err != nil {
		return fmt.Errorf("更新实例失败: %w", err)
	}

	return nil
}

// GetInstanceHistory 获取实例历史
func (s *WorkflowService) GetInstanceHistory(ctx context.Context, instanceID uuid.UUID) ([]model.WorkflowHistory, error) {
	return s.workflowRepo.GetInstanceHistory(ctx, instanceID)
}

// CancelInstance 取消实例
func (s *WorkflowService) CancelInstance(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID, reason string) error {
	instance, err := s.workflowRepo.GetInstanceByID(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("实例不存在")
	}

	if instance.IsCompleted() {
		return fmt.Errorf("工作流已结束")
	}

	now := time.Now()
	instance.Status = "cancelled"
	instance.EndTime = &now

	// 创建历史记录
	history := &model.WorkflowHistory{
		InstanceID: instance.ID,
		StepID:     *instance.CurrentStepID,
		StepName:   instance.CurrentStep.Name,
		OperatorID: operatorID,
		Action:     "cancel",
		Comment:    reason,
		EndTime:    &now,
	}

	if err := s.workflowRepo.CreateHistory(ctx, history); err != nil {
		return fmt.Errorf("保存历史记录失败: %w", err)
	}

	return s.workflowRepo.UpdateInstance(ctx, instance)
}

// GetMyInstances 获取我的待办
func (s *WorkflowService) GetMyInstances(ctx context.Context, userID uuid.UUID) ([]model.WorkflowInstance, error) {
	// 获取所有进行中的实例
	params := map[string]interface{}{
		"status": "running",
		"page":   1,
		"page_size": 100,
	}
	instances, _, err := s.workflowRepo.ListInstances(ctx, params)
	if err != nil {
		return nil, err
	}

	// 过滤出需要当前用户处理的实例（简化版）
	var myInstances []model.WorkflowInstance
	for _, instance := range instances {
		if instance.CurrentStep != nil {
			// 检查当前用户是否有权限处理此步骤
			// TODO: 根据角色和分配类型判断是否属于当前用户
			_ = userID
			myInstances = append(myInstances, instance)
		}
	}

	return myInstances, nil
}
