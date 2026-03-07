package handler

import (
	"strconv"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/application/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/middleware"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"
	"github.com/gin-gonic/gin"
)

// WorkflowHandler 工作流处理器
type WorkflowHandler struct {
	workflowService *service.WorkflowAppService
}

// NewWorkflowHandler 创建工作流处理器
func NewWorkflowHandler(workflowService *service.WorkflowAppService) *WorkflowHandler {
	return &WorkflowHandler{
		workflowService: workflowService,
	}
}

// RegisterRoutes 注册路由
func (h *WorkflowHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	// 流程定义（管理员）
	definitions := router.Group("/workflow-definitions")
	definitions.Use(authMiddleware.Required())
	{
		definitions.GET("", h.ListDefinitions)
		definitions.GET("/:id", h.GetDefinition)
		definitions.POST("", middleware.RequireManager(), h.CreateDefinition)
		definitions.POST("/:id/publish", middleware.RequireManager(), h.PublishDefinition)
	}
	
	// 流程实例
	instances := router.Group("/workflow-instances")
	instances.Use(authMiddleware.Required())
	{
		instances.GET("", h.ListInstances)
		instances.GET("/my", h.GetMyInstances)
		instances.GET("/todo", h.GetMyTodoInstances)
		instances.GET("/done", h.GetMyDoneInstances)
		instances.GET("/:id", h.GetInstance)
		instances.POST("", h.StartInstance)
		instances.POST("/:id/cancel", h.CancelInstance)
	}
	
	// 工作流任务
	tasks := router.Group("/workflow-tasks")
	tasks.Use(authMiddleware.Required())
	{
		tasks.GET("/todo", h.GetTodoTasks)
		tasks.GET("/done", h.GetDoneTasks)
		tasks.GET("/:id", h.GetTask)
		tasks.POST("/:id/approve", h.ApproveTask)
		tasks.POST("/:id/reject", h.RejectTask)
		tasks.POST("/:id/transfer", h.TransferTask)
		tasks.POST("/:id/delegate", h.DelegateTask)
		tasks.POST("/:id/return", h.ReturnTask)
		tasks.POST("/:id/remind", h.SendReminder)
	}
	
	// 统计
	router.GET("/workflow-stats", authMiddleware.Required(), h.GetStats)
}

// CreateDefinition 创建工作流定义
func (h *WorkflowHandler) CreateDefinition(c *gin.Context) {
	var req dto.CreateWorkflowDefinitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	
	// 设置当前用户组织
	req.OrgID = middleware.GetOrgID(c)
	
	resp, err := h.workflowService.CreateDefinition(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to create workflow definition", logger.Err(err))
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Created(c, resp)
}

// PublishDefinition 发布工作流定义
func (h *WorkflowHandler) PublishDefinition(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}
	
	if err := h.workflowService.PublishDefinition(c.Request.Context(), id); err != nil {
		logger.Error("Failed to publish workflow definition", logger.Err(err))
		if err == service.ErrWorkflowNotFound {
			response.NotFound(c, "workflow definition not found")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, nil)
}

// GetDefinition 获取工作流定义
func (h *WorkflowHandler) GetDefinition(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}
	
	resp, err := h.workflowService.GetDefinition(c.Request.Context(), id)
	if err != nil {
		logger.Error("Failed to get workflow definition", logger.Err(err))
		if err == service.ErrWorkflowNotFound {
			response.NotFound(c, "workflow definition not found")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, resp)
}

// ListDefinitions 列表查询
func (h *WorkflowHandler) ListDefinitions(c *gin.Context) {
	var req dto.ListWorkflowDefinitionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	
	req.OrgID = middleware.GetOrgID(c)
	
	resp, err := h.workflowService.ListDefinitions(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to list workflow definitions", logger.Err(err))
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, resp)
}

// StartInstance 启动流程实例
func (h *WorkflowHandler) StartInstance(c *gin.Context) {
	var req dto.StartWorkflowInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	
	req.StartedBy = middleware.GetUserID(c)
	if req.OrgID == "" {
		req.OrgID = middleware.GetOrgID(c)
	}
	
	resp, err := h.workflowService.StartInstance(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to start workflow instance", logger.Err(err))
		if err == service.ErrWorkflowNotFound {
			response.NotFound(c, "workflow definition not found")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Created(c, resp)
}

// GetInstance 获取流程实例
func (h *WorkflowHandler) GetInstance(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}
	
	resp, err := h.workflowService.GetInstance(c.Request.Context(), id)
	if err != nil {
		logger.Error("Failed to get workflow instance", logger.Err(err))
		if err == service.ErrInstanceNotFound {
			response.NotFound(c, "workflow instance not found")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, resp)
}

// ListInstances 列表查询
func (h *WorkflowHandler) ListInstances(c *gin.Context) {
	var req dto.ListWorkflowInstancesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	
	req.OrgID = middleware.GetOrgID(c)
	
	// 普通用户只能看到自己的
	if middleware.GetUserRole(c) == string(entity.RoleVolunteer) {
		req.StartedBy = middleware.GetUserID(c)
	}
	
	resp, err := h.workflowService.ListInstances(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to list workflow instances", logger.Err(err))
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, resp)
}

// GetMyInstances 获取我的流程
func (h *WorkflowHandler) GetMyInstances(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	
	resp, err := h.workflowService.ListInstances(c.Request.Context(), &dto.ListWorkflowInstancesRequest{
		StartedBy: userID,
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		logger.Error("Failed to get my instances", logger.Err(err))
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, resp)
}

// GetMyTodoInstances 获取我的待办流程
func (h *WorkflowHandler) GetMyTodoInstances(c *gin.Context) {
	_ = middleware.GetUserID(c) // reserved for future
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	
	// 这里需要查询用户有待办任务的流程
	// 简化处理，返回空列表
	response.Success(c, dto.ListWorkflowInstancesResponse{
		List:     []dto.WorkflowInstanceResponse{},
		Total:    0,
		Page:     page,
		PageSize: pageSize,
	})
}

// GetMyDoneInstances 获取我的已办流程
func (h *WorkflowHandler) GetMyDoneInstances(c *gin.Context) {
	_ = middleware.GetUserID(c) // reserved for future
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	
	// 简化处理
	response.Success(c, dto.ListWorkflowInstancesResponse{
		List:     []dto.WorkflowInstanceResponse{},
		Total:    0,
		Page:     page,
		PageSize: pageSize,
	})
}

// CancelInstance 取消流程实例
func (h *WorkflowHandler) CancelInstance(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}
	
	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)
	
	// TODO: 实现取消流程
	response.Success(c, nil)
}

// GetTodoTasks 获取待办任务
func (h *WorkflowHandler) GetTodoTasks(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	
	resp, err := h.workflowService.ListTodoTasks(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		logger.Error("Failed to get todo tasks", logger.Err(err))
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, resp)
}

// GetDoneTasks 获取已办任务
func (h *WorkflowHandler) GetDoneTasks(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	
	resp, err := h.workflowService.ListDoneTasks(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		logger.Error("Failed to get done tasks", logger.Err(err))
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, resp)
}

// GetTask 获取任务详情
func (h *WorkflowHandler) GetTask(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}
	
	resp, err := h.workflowService.GetTask(c.Request.Context(), id)
	if err != nil {
		logger.Error("Failed to get task", logger.Err(err))
		if err == service.ErrTaskNotFound {
			response.NotFound(c, "task not found")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, resp)
}

// ApproveTask 审批通过
func (h *WorkflowHandler) ApproveTask(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}
	
	var req dto.ApproveTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	
	userID := middleware.GetUserID(c)
	
	if err := h.workflowService.ApproveTask(c.Request.Context(), id, &req, userID); err != nil {
		logger.Error("Failed to approve task", logger.Err(err))
		if err == service.ErrTaskNotFound {
			response.NotFound(c, "task not found")
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	
	response.Success(c, nil)
}

// RejectTask 审批拒绝
func (h *WorkflowHandler) RejectTask(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}
	
	var req dto.RejectTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	
	userID := middleware.GetUserID(c)
	
	if err := h.workflowService.RejectTask(c.Request.Context(), id, &req, userID); err != nil {
		logger.Error("Failed to reject task", logger.Err(err))
		if err == service.ErrTaskNotFound {
			response.NotFound(c, "task not found")
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	
	response.Success(c, nil)
}

// TransferTask 转办任务
func (h *WorkflowHandler) TransferTask(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}
	
	var req struct {
		ToUserID string `json:"to_user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	
	userID := middleware.GetUserID(c)
	
	if err := h.workflowService.TransferTask(c.Request.Context(), id, req.ToUserID, userID); err != nil {
		logger.Error("Failed to transfer task", logger.Err(err))
		response.BadRequest(c, err.Error())
		return
	}
	
	response.Success(c, nil)
}

// DelegateTask 委托任务
func (h *WorkflowHandler) DelegateTask(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}
	
	var req struct {
		ToUserID string                   `json:"to_user_id" binding:"required"`
		Delegate dto.DelegateTaskRequest `json:"delegate"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	
	userID := middleware.GetUserID(c)
	
	if err := h.workflowService.DelegateTask(c.Request.Context(), id, req.ToUserID, &req.Delegate, userID); err != nil {
		logger.Error("Failed to delegate task", logger.Err(err))
		response.BadRequest(c, err.Error())
		return
	}
	
	response.Success(c, nil)
}

// ReturnTask 退回任务
func (h *WorkflowHandler) ReturnTask(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}
	
	var req struct {
		ToNodeID string `json:"to_node_id" binding:"required"`
		Comment  string `json:"comment,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	
	userID := middleware.GetUserID(c)
	
	if err := h.workflowService.ReturnTask(c.Request.Context(), id, req.ToNodeID, req.Comment, userID); err != nil {
		logger.Error("Failed to return task", logger.Err(err))
		response.BadRequest(c, err.Error())
		return
	}
	
	response.Success(c, nil)
}

// SendReminder 发送催办
func (h *WorkflowHandler) SendReminder(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}
	
	userID := middleware.GetUserID(c)
	
	if err := h.workflowService.SendReminder(c.Request.Context(), id, userID); err != nil {
		logger.Error("Failed to send reminder", logger.Err(err))
		response.BadRequest(c, err.Error())
		return
	}
	
	response.Success(c, nil)
}

// GetStats 获取统计
func (h *WorkflowHandler) GetStats(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	
	resp, err := h.workflowService.GetStats(c.Request.Context(), orgID)
	if err != nil {
		logger.Error("Failed to get workflow stats", logger.Err(err))
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, resp)
}
