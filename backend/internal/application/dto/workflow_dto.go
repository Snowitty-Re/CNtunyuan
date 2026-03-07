package dto

import (
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
)

// CreateWorkflowDefinitionRequest 创建工作流定义请求
type CreateWorkflowDefinitionRequest struct {
	Name        string                     `json:"name" binding:"required"`
	Key         string                     `json:"key" binding:"required"`
	Description string                     `json:"description,omitempty"`
	Category    string                     `json:"category,omitempty"`
	OrgID       string                     `json:"org_id" binding:"required"`
	Config      entity.JSONMap             `json:"config,omitempty"`
	FormSchema  entity.JSONMap             `json:"form_schema,omitempty"`
	Nodes       []CreateWorkflowNodeRequest `json:"nodes" binding:"required,min=2"`
}

// CreateWorkflowNodeRequest 创建节点请求
type CreateWorkflowNodeRequest struct {
	ID            string           `json:"id" binding:"required"`
	Name          string           `json:"name" binding:"required"`
	Type          string           `json:"type" binding:"required"`
	Config        entity.JSONMap   `json:"config,omitempty"`
	AssigneeType  string           `json:"assignee_type,omitempty"`
	Assignees     entity.JSONMap   `json:"assignees,omitempty"`
	ApprovalMode  string           `json:"approval_mode,omitempty"`
	RequiredCount int              `json:"required_count,omitempty"`
	AllowTransfer bool             `json:"allow_transfer,omitempty"`
	AllowDelegate bool             `json:"allow_delegate,omitempty"`
	AutoPass      bool             `json:"auto_pass,omitempty"`
	Conditions    entity.JSONMap   `json:"conditions,omitempty"`
	PositionX     float64          `json:"position_x,omitempty"`
	PositionY     float64          `json:"position_y,omitempty"`
}

// WorkflowDefinitionResponse 工作流定义响应
type WorkflowDefinitionResponse struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Key         string                   `json:"key"`
	Version     int                      `json:"version"`
	Description string                   `json:"description,omitempty"`
	Category    string                   `json:"category,omitempty"`
	Status      string                   `json:"status"`
	StartNodeID *string                  `json:"start_node_id,omitempty"`
	OrgID       string                   `json:"org_id"`
	Config      entity.JSONMap           `json:"config,omitempty"`
	FormSchema  entity.JSONMap           `json:"form_schema,omitempty"`
	IsSystem    bool                     `json:"is_system"`
	Nodes       []WorkflowNodeResponse   `json:"nodes,omitempty"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
}

// WorkflowNodeResponse 节点响应
type WorkflowNodeResponse struct {
	ID            string         `json:"id"`
	WorkflowID    string         `json:"workflow_id"`
	Name          string         `json:"name"`
	Type          string         `json:"type"`
	Config        entity.JSONMap `json:"config,omitempty"`
	AssigneeType  string         `json:"assignee_type,omitempty"`
	Assignees     entity.JSONMap `json:"assignees,omitempty"`
	ApprovalMode  string         `json:"approval_mode,omitempty"`
	RequiredCount int            `json:"required_count,omitempty"`
	AllowTransfer bool           `json:"allow_transfer,omitempty"`
	AllowDelegate bool           `json:"allow_delegate,omitempty"`
	AutoPass      bool           `json:"auto_pass,omitempty"`
	Conditions    entity.JSONMap `json:"conditions,omitempty"`
	PositionX     float64        `json:"position_x,omitempty"`
	PositionY     float64        `json:"position_y,omitempty"`
	OrderIndex    int            `json:"order_index"`
	CreatedAt     time.Time      `json:"created_at"`
}

// ListWorkflowDefinitionsRequest 列表查询请求
type ListWorkflowDefinitionsRequest struct {
	OrgID    string `json:"org_id,omitempty" form:"org_id"`
	Category string `json:"category,omitempty" form:"category"`
	Page     int    `json:"page,omitempty" form:"page"`
	PageSize int    `json:"page_size,omitempty" form:"page_size"`
}

// ListWorkflowDefinitionsResponse 列表响应
type ListWorkflowDefinitionsResponse struct {
	List     []WorkflowDefinitionResponse `json:"list"`
	Total    int64                        `json:"total"`
	Page     int                          `json:"page"`
	PageSize int                          `json:"page_size"`
}

// StartWorkflowInstanceRequest 启动流程实例请求
type StartWorkflowInstanceRequest struct {
	DefinitionID string         `json:"definition_id,omitempty"`
	WorkflowKey  string         `json:"workflow_key,omitempty"`
	BusinessKey  string         `json:"business_key,omitempty"`
	BusinessID   string         `json:"business_id,omitempty"`
	Title        string         `json:"title" binding:"required"`
	Variables    entity.JSONMap `json:"variables,omitempty"`
	StartedBy    string         `json:"started_by"`
	OrgID        string         `json:"org_id" binding:"required"`
}

// WorkflowInstanceResponse 流程实例响应
type WorkflowInstanceResponse struct {
	ID            string                  `json:"id"`
	DefinitionID  string                  `json:"definition_id"`
	Definition    *WorkflowDefinitionRef  `json:"definition,omitempty"`
	BusinessKey   string                  `json:"business_key,omitempty"`
	BusinessID    string                  `json:"business_id,omitempty"`
	Title         string                  `json:"title"`
	Status        string                  `json:"status"`
	CurrentNodeID *string                 `json:"current_node_id,omitempty"`
	CurrentNode   *WorkflowNodeRef        `json:"current_node,omitempty"`
	StartedBy     string                  `json:"started_by"`
	StartedAt     *time.Time              `json:"started_at,omitempty"`
	CompletedAt   *time.Time              `json:"completed_at,omitempty"`
	Variables     entity.JSONMap          `json:"variables,omitempty"`
	Result        string                  `json:"result,omitempty"`
	Comment       string                  `json:"comment,omitempty"`
	OrgID         string                  `json:"org_id"`
	CreatedAt     time.Time               `json:"created_at"`
	UpdatedAt     time.Time               `json:"updated_at"`
}

// WorkflowDefinitionRef 流程定义引用
type WorkflowDefinitionRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

// WorkflowNodeRef 节点引用
type WorkflowNodeRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// ListWorkflowInstancesRequest 列表查询请求
type ListWorkflowInstancesRequest struct {
	DefinitionID string `json:"definition_id,omitempty" form:"definition_id"`
	Status       string `json:"status,omitempty" form:"status"`
	StartedBy    string `json:"started_by,omitempty" form:"started_by"`
	OrgID        string `json:"org_id,omitempty" form:"org_id"`
	Page         int    `json:"page,omitempty" form:"page"`
	PageSize     int    `json:"page_size,omitempty" form:"page_size"`
}

// ListWorkflowInstancesResponse 列表响应
type ListWorkflowInstancesResponse struct {
	List     []WorkflowInstanceResponse `json:"list"`
	Total    int64                      `json:"total"`
	Page     int                        `json:"page"`
	PageSize int                        `json:"page_size"`
}

// ApproveTaskRequest 审批通过请求
type ApproveTaskRequest struct {
	Comment   string         `json:"comment,omitempty"`
	Variables entity.JSONMap `json:"variables,omitempty"`
}

// RejectTaskRequest 审批拒绝请求
type RejectTaskRequest struct {
	Comment   string         `json:"comment,omitempty"`
	Variables entity.JSONMap `json:"variables,omitempty"`
}

// DelegateTaskRequest 委托任务请求
type DelegateTaskRequest struct {
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Reason    string     `json:"reason,omitempty"`
}

// WorkflowTaskResponse 工作流任务响应
type WorkflowTaskResponse struct {
	ID                 string                 `json:"id"`
	Title              string                 `json:"title"`
	Description        string                 `json:"description,omitempty"`
	Status             string                 `json:"status"`
	AssigneeID         *string                `json:"assignee_id,omitempty"`
	CreatorID          string                 `json:"creator_id"`
	OrgID              string                 `json:"org_id"`
	WorkflowInstanceID string                 `json:"workflow_instance_id"`
	WorkflowNodeID     string                 `json:"workflow_node_id"`
	WorkflowNodeType   string                 `json:"workflow_node_type,omitempty"`
	Instance           *WorkflowInstanceRef   `json:"instance,omitempty"`
	ApprovalAction     string                 `json:"approval_action,omitempty"`
	ApprovalComment    string                 `json:"approval_comment,omitempty"`
	DueTime            *time.Time             `json:"due_time,omitempty"`
	RemindedAt         *time.Time             `json:"reminded_at,omitempty"`
	RemindCount        int                    `json:"remind_count"`
	DelegateFrom       *string                `json:"delegate_from,omitempty"`
	TransferredFrom    *string                `json:"transferred_from,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	CompletedAt        *time.Time             `json:"completed_at,omitempty"`
}

// WorkflowInstanceRef 流程实例引用
type WorkflowInstanceRef struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

// ListWorkflowTasksResponse 任务列表响应
type ListWorkflowTasksResponse struct {
	List     []WorkflowTaskResponse `json:"list"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

// WorkflowStatsResponse 统计响应
type WorkflowStatsResponse struct {
	TotalInstances  int64            `json:"total_instances"`
	PendingCount    int64            `json:"pending_count"`
	ProcessingCount int64            `json:"processing_count"`
	ApprovedCount   int64            `json:"approved_count"`
	RejectedCount   int64            `json:"rejected_count"`
	CancelledCount  int64            `json:"cancelled_count"`
	MyPendingCount  int64            `json:"my_pending_count"`
	MyDoneCount     int64            `json:"my_done_count"`
	DefinitionStats map[string]int64 `json:"definition_stats"`
}

// ToWorkflowDefinitionResponse 转换为响应
func ToWorkflowDefinitionResponse(def *entity.WorkflowDefinition) WorkflowDefinitionResponse {
	resp := WorkflowDefinitionResponse{
		ID:          def.ID,
		Name:        def.Name,
		Key:         def.Key,
		Version:     def.Version,
		Description: def.Description,
		Category:    def.Category,
		Status:      string(def.Status),
		StartNodeID: def.StartNodeID,
		OrgID:       def.OrgID,
		Config:      def.Config,
		FormSchema:  def.FormSchema,
		IsSystem:    def.IsSystem,
		CreatedAt:   def.CreatedAt,
		UpdatedAt:   def.UpdatedAt,
	}
	
	if len(def.Nodes) > 0 {
		resp.Nodes = make([]WorkflowNodeResponse, len(def.Nodes))
		for i, node := range def.Nodes {
			resp.Nodes[i] = ToWorkflowNodeResponse(&node)
		}
	}
	
	return resp
}

// ToWorkflowNodeResponse 转换为响应
func ToWorkflowNodeResponse(node *entity.WorkflowNode) WorkflowNodeResponse {
	return WorkflowNodeResponse{
		ID:            node.ID,
		WorkflowID:    node.WorkflowID,
		Name:          node.Name,
		Type:          string(node.Type),
		Config:        node.Config,
		AssigneeType:  string(node.AssigneeType),
		Assignees:     node.Assignees,
		ApprovalMode:  string(node.ApprovalMode),
		RequiredCount: node.RequiredCount,
		AllowTransfer: node.AllowTransfer,
		AllowDelegate: node.AllowDelegate,
		AutoPass:      node.AutoPass,
		Conditions:    node.Conditions,
		PositionX:     node.PositionX,
		PositionY:     node.PositionY,
		OrderIndex:    node.OrderIndex,
		CreatedAt:     node.CreatedAt,
	}
}

// ToWorkflowInstanceResponse 转换为响应
func ToWorkflowInstanceResponse(instance *entity.WorkflowInstance) WorkflowInstanceResponse {
	resp := WorkflowInstanceResponse{
		ID:            instance.ID,
		DefinitionID:  instance.DefinitionID,
		BusinessKey:   instance.BusinessKey,
		BusinessID:    instance.BusinessID,
		Title:         instance.Title,
		Status:        string(instance.Status),
		CurrentNodeID: instance.CurrentNodeID,
		StartedBy:     instance.StartedBy,
		StartedAt:     instance.StartedAt,
		CompletedAt:   instance.CompletedAt,
		Variables:     instance.Variables,
		Result:        instance.Result,
		Comment:       instance.Comment,
		OrgID:         instance.OrgID,
		CreatedAt:     instance.CreatedAt,
		UpdatedAt:     instance.UpdatedAt,
	}
	
	if instance.Definition != nil {
		resp.Definition = &WorkflowDefinitionRef{
			ID:   instance.Definition.ID,
			Name: instance.Definition.Name,
			Key:  instance.Definition.Key,
		}
		
		if instance.CurrentNodeID != nil {
			for _, node := range instance.Definition.Nodes {
				if node.ID == *instance.CurrentNodeID {
					resp.CurrentNode = &WorkflowNodeRef{
						ID:   node.ID,
						Name: node.Name,
						Type: string(node.Type),
					}
					break
				}
			}
		}
	}
	
	return resp
}

// ToWorkflowTaskResponse 转换为响应
func ToWorkflowTaskResponse(task *entity.WorkflowTask) WorkflowTaskResponse {
	resp := WorkflowTaskResponse{
		ID:                 task.ID,
		Title:              task.Title,
		Description:        task.Description,
		Status:             string(task.Status),
		AssigneeID:         task.AssigneeID,
		CreatorID:          task.CreatorID,
		OrgID:              task.OrgID,
		WorkflowInstanceID: task.WorkflowInstanceID,
		WorkflowNodeID:     task.WorkflowNodeID,
		WorkflowNodeType:   string(task.WorkflowNodeType),
		ApprovalAction:     task.ApprovalAction,
		ApprovalComment:    task.ApprovalComment,
		DueTime:            task.DueTime,
		RemindedAt:         task.RemindedAt,
		RemindCount:        task.RemindCount,
		DelegateFrom:       task.DelegateFrom,
		TransferredFrom:    task.TransferredFrom,
		CreatedAt:          task.CreatedAt,
		CompletedAt:        task.CompletedAt,
	}
	
	if task.Instance != nil {
		resp.Instance = &WorkflowInstanceRef{
			ID:     task.Instance.ID,
			Title:  task.Instance.Title,
			Status: string(task.Instance.Status),
		}
	}
	
	return resp
}
