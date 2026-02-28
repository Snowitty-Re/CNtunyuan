package api

import (
	"github.com/Snowitty-Re/CNtunyuan/internal/service"
	"github.com/Snowitty-Re/CNtunyuan/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// WorkflowHandler 工作流处理器
type WorkflowHandler struct {
	workflowService *service.WorkflowService
}

// NewWorkflowHandler 创建工作流处理器
func NewWorkflowHandler(workflowService *service.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{workflowService: workflowService}
}

// CreateWorkflow 创建工作流
// @Summary 创建工作流
// @Description 创建新的工作流定义
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body service.CreateWorkflowRequest true "工作流信息"
// @Success 200 {object} utils.Response{data=model.Workflow}
// @Router /workflows [post]
func (h *WorkflowHandler) CreateWorkflow(c *gin.Context) {
	var req service.CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	workflow, err := h.workflowService.CreateWorkflow(c.Request.Context(), uid, &req)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, workflow)
}

// GetWorkflow 获取工作流详情
// @Summary 获取工作流详情
// @Description 根据ID获取工作流详情
// @Tags 工作流管理
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "工作流ID"
// @Success 200 {object} utils.Response{data=model.Workflow}
// @Router /workflows/{id} [get]
func (h *WorkflowHandler) GetWorkflow(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的工作流ID")
		return
	}

	workflow, err := h.workflowService.GetWorkflow(c.Request.Context(), id)
	if err != nil {
		utils.NotFound(c, "工作流不存在")
		return
	}

	utils.Success(c, workflow)
}

// ListWorkflows 获取工作流列表
// @Summary 获取工作流列表
// @Description 获取工作流列表，支持分页和筛选
// @Tags 工作流管理
// @Produce json
// @Security ApiKeyAuth
// @Param status query string false "状态"
// @Param type query string false "类型"
// @Param keyword query string false "关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} utils.Response{data=map[string]interface{}}
// @Router /workflows [get]
func (h *WorkflowHandler) ListWorkflows(c *gin.Context) {
	var req service.ListWorkflowsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	workflows, total, err := h.workflowService.ListWorkflows(c.Request.Context(), &req)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"list":        workflows,
		"total":       total,
		"page":        req.Page,
		"page_size":   req.PageSize,
		"total_pages": (total + int64(req.PageSize) - 1) / int64(req.PageSize),
	})
}

// UpdateWorkflow 更新工作流
// @Summary 更新工作流
// @Description 更新工作流信息
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "工作流ID"
// @Param body body service.UpdateWorkflowRequest true "工作流信息"
// @Success 200 {object} utils.Response{data=model.Workflow}
// @Router /workflows/{id} [put]
func (h *WorkflowHandler) UpdateWorkflow(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的工作流ID")
		return
	}

	var req service.UpdateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	workflow, err := h.workflowService.UpdateWorkflow(c.Request.Context(), id, &req)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, workflow)
}

// DeleteWorkflow 删除工作流
// @Summary 删除工作流
// @Description 删除指定工作流
// @Tags 工作流管理
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "工作流ID"
// @Success 200 {object} utils.Response
// @Router /workflows/{id} [delete]
func (h *WorkflowHandler) DeleteWorkflow(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的工作流ID")
		return
	}

	if err := h.workflowService.DeleteWorkflow(c.Request.Context(), id); err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// CreateStep 创建工作流步骤
// @Summary 创建工作流步骤
// @Description 为工作流添加步骤
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "工作流ID"
// @Param body body service.CreateStepRequest true "步骤信息"
// @Success 200 {object} utils.Response{data=model.WorkflowStep}
// @Router /workflows/{id}/steps [post]
func (h *WorkflowHandler) CreateStep(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的工作流ID")
		return
	}

	var req service.CreateStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	step, err := h.workflowService.CreateStep(c.Request.Context(), workflowID, &req)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, step)
}

// UpdateStep 更新步骤
// @Summary 更新工作流步骤
// @Description 更新工作流步骤信息
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "工作流ID"
// @Param step_id path string true "步骤ID"
// @Param body body service.CreateStepRequest true "步骤信息"
// @Success 200 {object} utils.Response{data=model.WorkflowStep}
// @Router /workflows/{id}/steps/{step_id} [put]
func (h *WorkflowHandler) UpdateStep(c *gin.Context) {
	stepID, err := uuid.Parse(c.Param("step_id"))
	if err != nil {
		utils.BadRequest(c, "无效的步骤ID")
		return
	}

	var req service.CreateStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	step, err := h.workflowService.UpdateStep(c.Request.Context(), stepID, &req)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, step)
}

// DeleteStep 删除步骤
// @Summary 删除工作流步骤
// @Description 删除指定工作流步骤
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "工作流ID"
// @Param step_id path string true "步骤ID"
// @Success 200 {object} utils.Response
// @Router /workflows/{id}/steps/{step_id} [delete]
func (h *WorkflowHandler) DeleteStep(c *gin.Context) {
	stepID, err := uuid.Parse(c.Param("step_id"))
	if err != nil {
		utils.BadRequest(c, "无效的步骤ID")
		return
	}

	if err := h.workflowService.DeleteStep(c.Request.Context(), stepID); err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// ReorderSteps 重新排序步骤
// @Summary 重新排序步骤
// @Description 重新排序工作流步骤
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "工作流ID"
// @Param body body []string true "步骤ID数组"
// @Success 200 {object} utils.Response
// @Router /workflows/{id}/steps/reorder [post]
func (h *WorkflowHandler) ReorderSteps(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的工作流ID")
		return
	}

	var req struct {
		StepIDs []string `json:"step_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.workflowService.ReorderSteps(c.Request.Context(), workflowID, req.StepIDs); err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// StartInstance 启动工作流实例
// @Summary 启动工作流实例
// @Description 启动一个新的工作流实例
// @Tags 工作流实例
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body service.StartInstanceRequest true "启动信息"
// @Success 200 {object} utils.Response{data=model.WorkflowInstance}
// @Router /workflow-instances [post]
func (h *WorkflowHandler) StartInstance(c *gin.Context) {
	var req service.StartInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	instance, err := h.workflowService.StartInstance(c.Request.Context(), uid, &req)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, instance)
}

// GetInstance 获取实例详情
// @Summary 获取工作流实例详情
// @Description 根据ID获取工作流实例详情
// @Tags 工作流实例
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "实例ID"
// @Success 200 {object} utils.Response{data=model.WorkflowInstance}
// @Router /workflow-instances/{id} [get]
func (h *WorkflowHandler) GetInstance(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的实例ID")
		return
	}

	instance, err := h.workflowService.GetInstance(c.Request.Context(), id)
	if err != nil {
		utils.NotFound(c, "实例不存在")
		return
	}

	utils.Success(c, instance)
}

// ListInstances 获取实例列表
// @Summary 获取工作流实例列表
// @Description 获取工作流实例列表，支持分页和筛选
// @Tags 工作流实例
// @Produce json
// @Security ApiKeyAuth
// @Param status query string false "状态"
// @Param workflow_id query string false "工作流ID"
// @Param starter_id query string false "发起人ID"
// @Param business_type query string false "业务类型"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} utils.Response{data=map[string]interface{}}
// @Router /workflow-instances [get]
func (h *WorkflowHandler) ListInstances(c *gin.Context) {
	var req service.ListInstancesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	instances, total, err := h.workflowService.ListInstances(c.Request.Context(), &req)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"list":        instances,
		"total":       total,
		"page":        req.Page,
		"page_size":   req.PageSize,
		"total_pages": (total + int64(req.PageSize) - 1) / int64(req.PageSize),
	})
}

// Approve 审批
// @Summary 审批工作流实例
// @Description 审批当前工作流实例步骤
// @Tags 工作流实例
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "实例ID"
// @Param body body service.ApproveRequest true "审批信息"
// @Success 200 {object} utils.Response
// @Router /workflow-instances/{id}/approve [post]
func (h *WorkflowHandler) Approve(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的实例ID")
		return
	}

	var req service.ApproveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	if err := h.workflowService.Approve(c.Request.Context(), id, uid, &req); err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// GetInstanceHistory 获取实例历史
// @Summary 获取工作流实例历史
// @Description 获取工作流实例的审批历史
// @Tags 工作流实例
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "实例ID"
// @Success 200 {object} utils.Response{data=[]model.WorkflowHistory}
// @Router /workflow-instances/{id}/history [get]
func (h *WorkflowHandler) GetInstanceHistory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的实例ID")
		return
	}

	history, err := h.workflowService.GetInstanceHistory(c.Request.Context(), id)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, history)
}

// CancelInstance 取消实例
// @Summary 取消工作流实例
// @Description 取消指定工作流实例
// @Tags 工作流实例
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "实例ID"
// @Param reason body string false "取消原因"
// @Success 200 {object} utils.Response
// @Router /workflow-instances/{id}/cancel [post]
func (h *WorkflowHandler) CancelInstance(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的实例ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	if err := h.workflowService.CancelInstance(c.Request.Context(), id, uid, req.Reason); err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// GetMyInstances 获取我的待办
// @Summary 获取我的待办
// @Description 获取当前用户需要处理的工作流实例
// @Tags 工作流实例
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response{data=[]model.WorkflowInstance}
// @Router /workflow-instances/my [get]
func (h *WorkflowHandler) GetMyInstances(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	instances, err := h.workflowService.GetMyInstances(c.Request.Context(), uid)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, instances)
}
