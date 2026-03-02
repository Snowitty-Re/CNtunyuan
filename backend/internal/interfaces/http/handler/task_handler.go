package handler

import (
	"strconv"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/application/service"
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

// GetByID 获取详情
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

// List 列表查询
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

// Update 更新
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

	task, err := h.taskService.Update(c.Request.Context(), id, &req)
	if err != nil {
		switch err {
		case service.ErrTaskNotFound:
			response.NotFound(c, "task not found")
		default:
			logger.Error("Failed to update task", logger.Err(err))
			response.InternalServerError(c, "failed to update")
		}
		return
	}

	response.Success(c, task)
}

// Delete 删除
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
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.Success(c, nil)
}

// Complete 完成任务
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
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.Success(c, nil)
}

// Cancel 取消任务
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

// UpdateProgress 更新进度
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

	if err := h.taskService.UpdateProgress(c.Request.Context(), id, req.Progress, userID); err != nil {
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

// GetMyTasks 获取我的任务
func (h *TaskHandler) GetMyTasks(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	tasks, err := h.taskService.GetMyTasks(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.InternalServerError(c, "failed to get my tasks")
		return
	}

	response.Success(c, tasks)
}

// GetPendingTasks 获取待分配任务
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

// GetStats 获取统计
func (h *TaskHandler) GetStats(c *gin.Context) {
	userID := middleware.GetUserID(c)

	stats, err := h.taskService.GetStats(c.Request.Context(), userID)
	if err != nil {
		response.InternalServerError(c, "failed to get stats")
		return
	}

	response.Success(c, stats)
}
