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

// TaskHandler 任务处理器
type TaskHandler struct {
	taskService *service.TaskAppService
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(taskService *service.TaskAppService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

// RegisterRoutes 注册路由
func (h *TaskHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	tasks := router.Group("/tasks")
	tasks.Use(authMiddleware.Required())
	{
		tasks.GET("", h.List)
		tasks.GET("/my", h.GetMyTasks)
		tasks.GET("/pending", h.GetPendingTasks)
		tasks.GET("/overdue", middleware.RequireManager(), h.GetOverdueTasks)
		tasks.GET("/stats", h.GetStats)
		tasks.GET("/:id", h.GetByID)
		tasks.GET("/:id/logs", h.GetLogs)
		tasks.GET("/:id/follow-ups", h.GetFollowUps)
		tasks.POST("/:id/follow-ups", h.CreateFollowUp)
		tasks.POST("/:id/follow-ups/:follow_up_id/review", middleware.RequireManager(), h.ReviewFollowUp)
		tasks.GET("/:id/follow-ups/:follow_up_id/comments", h.GetFollowUpComments)
		tasks.POST("/:id/follow-ups/:follow_up_id/comments", h.AddFollowUpComment)
		tasks.POST("", h.Create)
		tasks.PUT("/:id", h.Update)
		tasks.DELETE("/:id", middleware.RequireManager(), h.Delete)
		tasks.POST("/:id/assign", middleware.RequireManager(), h.Assign)
		tasks.POST("/:id/start", h.Start)
		tasks.POST("/:id/complete", h.Complete)
		tasks.POST("/:id/cancel", middleware.RequireManager(), h.Cancel)
		tasks.PUT("/:id/progress", h.UpdateProgress)
	}
}

// Create 创建任务
// @Summary      创建任务
// @Description  创建新的寻人任务，需要登录
// @Tags         任务管理
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreateTaskRequest  true  "创建任务请求参数"
// @Success      201      {object}  response.Response{data=dto.TaskResponse}  "创建成功"
// @Failure      400      {object}  response.Response  "请求参数错误"
// @Failure      401      {object}  response.Response  "未授权"
// @Failure      500      {object}  response.Response  "服务器内部错误"
// @Router       /tasks [post]
// @Security     Bearer
func (h *TaskHandler) Create(c *gin.Context) {
	var req dto.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	orgID := middleware.GetOrgID(c)

	task, err := h.taskService.Create(c.Request.Context(), &req, userID, orgID)
	if err != nil {
		logger.Error("Failed to create task", logger.Err(err))
		response.InternalServerError(c, "failed to create task")
		return
	}

	response.Created(c, task)
}

// GetByID 获取任务详情
// @Summary      获取任务详情
// @Description  根据任务ID获取任务详细信息
// @Tags         任务管理
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "任务ID"
// @Success      200  {object}  response.Response{data=dto.TaskResponse}  "获取成功"
// @Failure      400  {object}  response.Response  "任务ID不能为空"
// @Failure      401  {object}  response.Response  "未授权"
// @Failure      404  {object}  response.Response  "任务不存在"
// @Failure      500  {object}  response.Response  "服务器内部错误"
// @Router       /tasks/{id} [get]
// @Security     Bearer
func (h *TaskHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "task id is required")
		return
	}

	task, err := h.taskService.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrTaskNotFound {
			response.NotFound(c, "task not found")
			return
		}
		response.InternalServerError(c, "failed to get task")
		return
	}

	response.Success(c, task)
}

// List 获取任务列表
// @Summary      获取任务列表
// @Description  分页获取任务列表，支持多种筛选条件
// @Tags         任务管理
// @Accept       json
// @Produce      json
// @Param        page              query     int     false  "页码，默认1"
// @Param        page_size         query     int     false  "每页数量，默认10，最大100"
// @Param        keyword           query     string  false  "关键词搜索"
// @Param        type              query     string  false  "任务类型"
// @Param        status            query     string  false  "任务状态: draft, pending, assigned, processing, completed, cancelled"
// @Param        priority          query     string  false  "优先级: low, normal, high, urgent"
// @Param        assignee_id       query     string  false  "执行人ID"
// @Param        missing_person_id query     string  false  "走失人员ID"
// @Param        is_overdue        query     bool    false  "是否逾期"
// @Success      200  {object}  response.Response{data=dto.TaskListResponse}  "获取成功"
// @Failure      400  {object}  response.Response  "请求参数错误"
// @Failure      401  {object}  response.Response  "未授权"
// @Failure      500  {object}  response.Response  "服务器内部错误"
// @Router       /tasks [get]
// @Security     Bearer
func (h *TaskHandler) List(c *gin.Context) {
	var req dto.TaskListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tasks, err := h.taskService.List(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "failed to get list")
		return
	}

	response.Success(c, tasks)
}

// Update 更新任务
// @Summary      更新任务
// @Description  更新指定任务的详细信息
// @Tags         任务管理
// @Accept       json
// @Produce      json
// @Param        id       path      string                 true  "任务ID"
// @Param        request  body      dto.UpdateTaskRequest  true  "更新任务请求参数"
// @Success      200      {object}  response.Response{data=dto.TaskResponse}  "更新成功"
// @Failure      400      {object}  response.Response  "请求参数错误"
// @Failure      401      {object}  response.Response  "未授权"
// @Failure      404      {object}  response.Response  "任务不存在"
// @Failure      500      {object}  response.Response  "服务器内部错误"
// @Router       /tasks/{id} [put]
// @Security     Bearer
func (h *TaskHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "task id is required")
		return
	}

	var req dto.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	isManager := entity.GetRoleLevel(role) >= entity.RoleLevelManager

	task, err := h.taskService.Update(c.Request.Context(), id, &req, userID, isManager)
	if err != nil {
		switch err {
		case service.ErrTaskNotFound:
			response.NotFound(c, "task not found")
		case service.ErrTaskForbidden:
			response.Forbidden(c, "no permission to modify this task")
		default:
			logger.Error("Failed to update task", logger.Err(err))
			response.InternalServerError(c, "failed to update")
		}
		return
	}

	response.Success(c, task)
}

// Delete 删除任务
// @Summary      删除任务
// @Description  删除指定任务，需要管理员权限(manager+)
// @Tags         任务管理
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "任务ID"
// @Success      204  {object}  nil  "删除成功"
// @Failure      400  {object}  response.Response  "任务ID不能为空"
// @Failure      401  {object}  response.Response  "未授权"
// @Failure      403  {object}  response.Response  "权限不足"
// @Failure      500  {object}  response.Response  "服务器内部错误"
// @Router       /tasks/{id} [delete]
// @Security     Bearer
func (h *TaskHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "task id is required")
		return
	}

	if err := h.taskService.Delete(c.Request.Context(), id); err != nil {
		logger.Error("Failed to delete task", logger.Err(err))
		response.InternalServerError(c, "failed to delete")
		return
	}

	response.NoContent(c)
}

// Assign 分配任务
// @Summary      分配任务
// @Description  将任务分配给指定执行人，需要管理员权限(manager+)
// @Tags         任务管理
// @Accept       json
// @Produce      json
// @Param        id       path      string                   true  "任务ID"
// @Param        request  body      dto.AssignTaskRequest    true  "分配任务请求参数"
// @Success      200      {object}  response.Response        "分配成功"
// @Failure      400      {object}  response.Response        "请求参数错误或任务已分配"
// @Failure      401      {object}  response.Response        "未授权"
// @Failure      403      {object}  response.Response        "权限不足"
// @Failure      404      {object}  response.Response        "任务不存在"
// @Router       /tasks/{id}/assign [post]
// @Security     Bearer
func (h *TaskHandler) Assign(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "task id is required")
		return
	}

	var req dto.AssignTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.taskService.Assign(c.Request.Context(), id, req.AssigneeID); err != nil {
		switch err {
		case service.ErrTaskNotFound:
			response.NotFound(c, "task not found")
		case service.ErrAlreadyAssigned:
			response.BadRequest(c, "task already assigned")
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.Success(c, nil)
}

// Start 开始任务
// @Summary      开始任务
// @Description  开始执行指定的任务
// @Tags         任务管理
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "任务ID"
// @Success      200  {object}  response.Response  "开始任务成功"
// @Failure      400  {object}  response.Response  "任务ID不能为空或状态不允许"
// @Failure      401  {object}  response.Response  "未授权"
// @Failure      404  {object}  response.Response  "任务不存在"
// @Router       /tasks/{id}/start [post]
// @Security     Bearer
func (h *TaskHandler) Start(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "task id is required")
		return
	}

	userID := middleware.GetUserID(c)

	if err := h.taskService.Start(c.Request.Context(), id, userID); err != nil {
		switch err {
		case service.ErrTaskNotFound:
			response.NotFound(c, "task not found")
		case service.ErrTaskNotAssignedToUser:
			response.Forbidden(c, "task not assigned to you")
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.Success(c, nil)
}

// Complete 完成任务
// @Summary      完成任务
// @Description  标记任务为已完成，并填写任务结果
// @Tags         任务管理
// @Accept       json
// @Produce      json
// @Param        id       path      string                     true  "任务ID"
// @Param        request  body      dto.CompleteTaskRequest    true  "完成任务请求参数"
// @Success      200      {object}  response.Response          "完成任务成功"
// @Failure      400      {object}  response.Response          "请求参数错误或状态不允许"
// @Failure      401      {object}  response.Response          "未授权"
// @Failure      404      {object}  response.Response          "任务不存在"
// @Router       /tasks/{id}/complete [post]
// @Security     Bearer
func (h *TaskHandler) Complete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "task id is required")
		return
	}

	var req dto.CompleteTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)

	if err := h.taskService.Complete(c.Request.Context(), id, &req, userID); err != nil {
		switch err {
		case service.ErrTaskNotFound:
			response.NotFound(c, "task not found")
		case service.ErrTaskNotAssignedToUser:
			response.Forbidden(c, "task not assigned to you")
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.Success(c, nil)
}

// Cancel 取消任务
// @Summary      取消任务
// @Description  取消指定的任务，需要管理员权限(manager+)
// @Tags         任务管理
// @Accept       json
// @Produce      json
// @Param        id       path      string                   true  "任务ID"
// @Param        request  body      dto.CancelTaskRequest    true  "取消任务请求参数"
// @Success      200      {object}  response.Response        "取消任务成功"
// @Failure      400      {object}  response.Response        "请求参数错误或状态不允许"
// @Failure      401      {object}  response.Response        "未授权"
// @Failure      403      {object}  response.Response        "权限不足"
// @Failure      404      {object}  response.Response        "任务不存在"
// @Router       /tasks/{id}/cancel [post]
// @Security     Bearer
func (h *TaskHandler) Cancel(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "task id is required")
		return
	}

	var req dto.CancelTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)

	if err := h.taskService.Cancel(c.Request.Context(), id, &req, userID); err != nil {
		switch err {
		case service.ErrTaskNotFound:
			response.NotFound(c, "task not found")
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.Success(c, nil)
}

// UpdateProgress 更新任务进度
// @Summary      更新任务进度
// @Description  更新任务的执行进度(0-100)
// @Tags         任务管理
// @Accept       json
// @Produce      json
// @Param        id       path      string                          true  "任务ID"
// @Param        request  body      dto.UpdateTaskProgressRequest   true  "更新进度请求参数"
// @Success      200      {object}  response.Response               "更新进度成功"
// @Failure      400      {object}  response.Response               "请求参数错误或进度值无效"
// @Failure      401      {object}  response.Response               "未授权"
// @Failure      404      {object}  response.Response               "任务不存在"
// @Router       /tasks/{id}/progress [put]
// @Security     Bearer
func (h *TaskHandler) UpdateProgress(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "task id is required")
		return
	}

	var req dto.UpdateTaskProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)

	if err := h.taskService.UpdateProgress(c.Request.Context(), id, req.Progress, req.Remark, userID); err != nil {
		switch err {
		case service.ErrTaskNotFound:
			response.NotFound(c, "task not found")
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.Success(c, nil)
}

// CreateFollowUp 创建任务跟进
func (h *TaskHandler) CreateFollowUp(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		response.BadRequest(c, "task id is required")
		return
	}

	var req dto.CreateTaskFollowUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	isManager := entity.GetRoleLevel(role) >= entity.RoleLevelManager

	followUp, err := h.taskService.CreateFollowUp(c.Request.Context(), taskID, &req, userID, isManager)
	if err != nil {
		switch err {
		case service.ErrTaskNotFound:
			response.NotFound(c, "task not found")
		case service.ErrTaskForbidden:
			response.Forbidden(c, "no permission to add follow up")
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.Created(c, followUp)
}

// GetFollowUps 获取任务跟进列表
func (h *TaskHandler) GetFollowUps(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		response.BadRequest(c, "task id is required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	result, err := h.taskService.ListFollowUps(c.Request.Context(), taskID, page, pageSize)
	if err != nil {
		switch err {
		case service.ErrTaskNotFound:
			response.NotFound(c, "task not found")
		default:
			response.InternalServerError(c, "failed to get follow ups")
		}
		return
	}
	response.Success(c, result)
}

// ReviewFollowUp 审核任务跟进
func (h *TaskHandler) ReviewFollowUp(c *gin.Context) {
	taskID := c.Param("id")
	followUpID := c.Param("follow_up_id")
	if taskID == "" || followUpID == "" {
		response.BadRequest(c, "task id and follow up id are required")
		return
	}

	var req dto.ReviewTaskFollowUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	isManager := entity.GetRoleLevel(role) >= entity.RoleLevelManager

	if err := h.taskService.ReviewFollowUp(c.Request.Context(), taskID, followUpID, &req, userID, isManager); err != nil {
		switch err {
		case service.ErrTaskNotFound:
			response.NotFound(c, "task not found")
		case service.ErrTaskFollowUpNotFound:
			response.NotFound(c, "follow up not found")
		case service.ErrTaskForbidden:
			response.Forbidden(c, "no permission to review follow up")
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.Success(c, nil)
}

// AddFollowUpComment 添加任务跟进评论
func (h *TaskHandler) AddFollowUpComment(c *gin.Context) {
	taskID := c.Param("id")
	followUpID := c.Param("follow_up_id")
	if taskID == "" || followUpID == "" {
		response.BadRequest(c, "task id and follow up id are required")
		return
	}

	var req dto.CreateTaskFollowUpCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	isManager := entity.GetRoleLevel(role) >= entity.RoleLevelManager

	comment, err := h.taskService.AddFollowUpComment(c.Request.Context(), taskID, followUpID, &req, userID, isManager)
	if err != nil {
		switch err {
		case service.ErrTaskNotFound:
			response.NotFound(c, "task not found")
		case service.ErrTaskFollowUpNotFound:
			response.NotFound(c, "follow up not found")
		case service.ErrTaskForbidden:
			response.Forbidden(c, "no permission to comment")
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}
	response.Created(c, comment)
}

// GetFollowUpComments 获取任务跟进评论
func (h *TaskHandler) GetFollowUpComments(c *gin.Context) {
	taskID := c.Param("id")
	followUpID := c.Param("follow_up_id")
	if taskID == "" || followUpID == "" {
		response.BadRequest(c, "task id and follow up id are required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	comments, err := h.taskService.GetFollowUpComments(c.Request.Context(), taskID, followUpID, page, pageSize)
	if err != nil {
		switch err {
		case service.ErrTaskNotFound:
			response.NotFound(c, "task not found")
		case service.ErrTaskFollowUpNotFound:
			response.NotFound(c, "follow up not found")
		default:
			response.InternalServerError(c, "failed to get follow up comments")
		}
		return
	}
	response.Success(c, comments)
}

// GetMyTasks 获取我的任务
// @Summary      获取我的任务
// @Description  获取当前登录用户的任务列表
// @Tags         任务管理
// @Accept       json
// @Produce      json
// @Param        page       query     int  false  "页码，默认1"
// @Param        page_size  query     int  false  "每页数量，默认10"
// @Success      200        {object}  response.Response{data=dto.TaskListResponse}  "获取成功"
// @Failure      401        {object}  response.Response  "未授权"
// @Failure      500        {object}  response.Response  "服务器内部错误"
// @Router       /tasks/my [get]
// @Security     Bearer
func (h *TaskHandler) GetMyTasks(c *gin.Context) {
	userID := middleware.GetUserID(c)
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	tasks, err := h.taskService.GetMyTasks(c.Request.Context(), userID, status, page, pageSize)
	if err != nil {
		response.InternalServerError(c, "failed to get my tasks")
		return
	}

	response.Success(c, tasks)
}

// GetPendingTasks 获取待分配任务
// @Summary      获取待分配任务
// @Description  获取状态为待分配的任务列表
// @Tags         任务管理
// @Accept       json
// @Produce      json
// @Param        page       query     int  false  "页码，默认1"
// @Param        page_size  query     int  false  "每页数量，默认10"
// @Success      200  {object}  response.Response{data=dto.TaskListResponse}  "获取成功"
// @Failure      401  {object}  response.Response  "未授权"
// @Failure      500  {object}  response.Response  "服务器内部错误"
// @Router       /tasks/pending [get]
// @Security     Bearer
func (h *TaskHandler) GetPendingTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	tasks, err := h.taskService.GetPendingTasks(c.Request.Context(), page, pageSize)
	if err != nil {
		response.InternalServerError(c, "failed to get pending tasks")
		return
	}

	response.Success(c, tasks)
}

// GetOverdueTasks 获取逾期任务
// @Summary      获取逾期任务
// @Description  获取已逾期的任务列表，需要管理员权限(manager+)
// @Tags         任务管理
// @Accept       json
// @Produce      json
// @Param        page       query     int  false  "页码，默认1"
// @Param        page_size  query     int  false  "每页数量，默认10"
// @Success      200  {object}  response.Response{data=dto.TaskListResponse}  "获取成功"
// @Failure      401  {object}  response.Response  "未授权"
// @Failure      403  {object}  response.Response  "权限不足"
// @Failure      500  {object}  response.Response  "服务器内部错误"
// @Router       /tasks/overdue [get]
// @Security     Bearer
func (h *TaskHandler) GetOverdueTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	tasks, err := h.taskService.GetOverdueTasks(c.Request.Context(), page, pageSize)
	if err != nil {
		response.InternalServerError(c, "failed to get overdue tasks")
		return
	}

	response.Success(c, tasks)
}

// GetLogs 获取任务日志
// @Summary      获取任务日志
// @Description  获取指定任务的操作日志记录
// @Tags         任务管理
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "任务ID"
// @Success      200  {object}  response.Response{data=[]dto.TaskLogResponse}  "获取成功"
// @Failure      400  {object}  response.Response  "任务ID不能为空"
// @Failure      401  {object}  response.Response  "未授权"
// @Failure      500  {object}  response.Response  "服务器内部错误"
// @Router       /tasks/{id}/logs [get]
// @Security     Bearer
func (h *TaskHandler) GetLogs(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "task id is required")
		return
	}

	logs, err := h.taskService.GetLogs(c.Request.Context(), id)
	if err != nil {
		response.InternalServerError(c, "failed to get logs")
		return
	}

	response.Success(c, logs)
}

// GetStats 获取任务统计
// @Summary      获取任务统计
// @Description  获取当前用户的任务统计数据
// @Tags         任务管理
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.Response{data=dto.TaskStatsResponse}  "获取成功"
// @Failure      401  {object}  response.Response  "未授权"
// @Failure      500  {object}  response.Response  "服务器内部错误"
// @Router       /tasks/stats [get]
// @Security     Bearer
func (h *TaskHandler) GetStats(c *gin.Context) {
	userID := middleware.GetUserID(c)

	stats, err := h.taskService.GetStats(c.Request.Context(), userID)
	if err != nil {
		response.InternalServerError(c, "failed to get stats")
		return
	}

	response.Success(c, stats)
}
