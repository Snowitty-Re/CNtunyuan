package api

import (
	"strconv"

	"github.com/Snowitty-Re/CNtunyuan/internal/service"
	"github.com/Snowitty-Re/CNtunyuan/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TaskHandler 任务处理器
type TaskHandler struct {
	taskService *service.TaskService
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

// CreateTask 创建任务
// @Summary 创建任务
// @Description 创建新的任务
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body service.CreateTaskRequest true "任务信息"
// @Success 200 {object} utils.Response{data=model.Task}
// @Router /tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req service.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	task, err := h.taskService.CreateTask(c.Request.Context(), uid, &req)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, task)
}

// GetTask 获取任务详情
// @Summary 获取任务详情
// @Description 根据ID获取任务详情
// @Tags 任务管理
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "任务ID"
// @Success 200 {object} utils.Response{data=model.Task}
// @Router /tasks/{id} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}

	task, err := h.taskService.GetTask(c.Request.Context(), id)
	if err != nil {
		utils.NotFound(c, "任务不存在")
		return
	}

	utils.Success(c, task)
}

// ListTasks 获取任务列表
// @Summary 获取任务列表
// @Description 获取任务列表，支持分页和筛选
// @Tags 任务管理
// @Produce json
// @Security ApiKeyAuth
// @Param status query string false "状态"
// @Param assignee_id query string false "执行人ID"
// @Param creator_id query string false "创建人ID"
// @Param org_id query string false "组织ID"
// @Param type query string false "任务类型"
// @Param priority query string false "优先级"
// @Param keyword query string false "关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} utils.Response{data=map[string]interface{}}
// @Router /tasks [get]
func (h *TaskHandler) ListTasks(c *gin.Context) {
	var req service.ListTasksRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	// 默认分页
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	tasks, total, err := h.taskService.ListTasks(c.Request.Context(), &req)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"list":       tasks,
		"total":      total,
		"page":       req.Page,
		"page_size":  req.PageSize,
		"total_pages": (total + int64(req.PageSize) - 1) / int64(req.PageSize),
	})
}

// UpdateTask 更新任务
// @Summary 更新任务
// @Description 更新任务信息
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "任务ID"
// @Param body body service.UpdateTaskRequest true "任务信息"
// @Success 200 {object} utils.Response{data=model.Task}
// @Router /tasks/{id} [put]
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}

	var req service.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	task, err := h.taskService.UpdateTask(c.Request.Context(), id, uid, &req)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, task)
}

// DeleteTask 删除任务
// @Summary 删除任务
// @Description 删除指定任务
// @Tags 任务管理
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "任务ID"
// @Success 200 {object} utils.Response
// @Router /tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}

	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	if err := h.taskService.DeleteTask(c.Request.Context(), id, uid); err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// AssignTask 分配任务
// @Summary 分配任务
// @Description 将任务分配给指定用户
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "任务ID"
// @Param body body service.AssignTaskRequest true "分配信息"
// @Success 200 {object} utils.Response
// @Router /tasks/{id}/assign [post]
func (h *TaskHandler) AssignTask(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}

	var req service.AssignTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	if err := h.taskService.AssignTask(c.Request.Context(), id, uid, &req); err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// UnassignTask 取消分配
// @Summary 取消任务分配
// @Description 取消任务的分配
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "任务ID"
// @Param reason body string false "取消原因"
// @Success 200 {object} utils.Response
// @Router /tasks/{id}/unassign [post]
func (h *TaskHandler) UnassignTask(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	if err := h.taskService.UnassignTask(c.Request.Context(), id, uid, req.Reason); err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// TransferTask 转派任务
// @Summary 转派任务
// @Description 将任务转派给其他用户
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "任务ID"
// @Param body body service.TransferTaskRequest true "转派信息"
// @Success 200 {object} utils.Response
// @Router /tasks/{id}/transfer [post]
func (h *TaskHandler) TransferTask(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}

	var req service.TransferTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	if err := h.taskService.TransferTask(c.Request.Context(), id, uid, &req); err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// CompleteTask 完成任务
// @Summary 完成任务
// @Description 标记任务为已完成
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "任务ID"
// @Param body body service.CompleteTaskRequest true "完成信息"
// @Success 200 {object} utils.Response{data=model.Task}
// @Router /tasks/{id}/complete [post]
func (h *TaskHandler) CompleteTask(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}

	var req service.CompleteTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	task, err := h.taskService.CompleteTask(c.Request.Context(), id, uid, &req)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, task)
}

// CancelTask 取消任务
// @Summary 取消任务
// @Description 取消指定任务
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "任务ID"
// @Param reason body string false "取消原因"
// @Success 200 {object} utils.Response
// @Router /tasks/{id}/cancel [post]
func (h *TaskHandler) CancelTask(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	if err := h.taskService.CancelTask(c.Request.Context(), id, uid, req.Reason); err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// UpdateProgress 更新进度
// @Summary 更新任务进度
// @Description 更新任务完成进度
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "任务ID"
// @Param progress body int true "进度(0-100)"
// @Success 200 {object} utils.Response
// @Router /tasks/{id}/progress [put]
func (h *TaskHandler) UpdateProgress(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}

	var req struct {
		Progress int `json:"progress" binding:"required,min=0,max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: 进度必须在0-100之间")
		return
	}

	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	if err := h.taskService.UpdateProgress(c.Request.Context(), id, uid, req.Progress); err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// GetMyTasks 获取我的任务
// @Summary 获取我的任务
// @Description 获取当前登录用户的任务列表
// @Tags 任务管理
// @Produce json
// @Security ApiKeyAuth
// @Param status query string false "状态筛选"
// @Success 200 {object} utils.Response{data=[]model.Task}
// @Router /tasks/my [get]
func (h *TaskHandler) GetMyTasks(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	status := c.Query("status")

	tasks, err := h.taskService.GetMyTasks(c.Request.Context(), uid, status)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, tasks)
}

// GetCreatedTasks 获取我创建的任务
// @Summary 获取我创建的任务
// @Description 获取当前登录用户创建的任务列表
// @Tags 任务管理
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response{data=[]model.Task}
// @Router /tasks/created [get]
func (h *TaskHandler) GetCreatedTasks(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	tasks, err := h.taskService.GetCreatedTasks(c.Request.Context(), uid)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, tasks)
}

// GetTaskStatistics 获取任务统计
// @Summary 获取任务统计
// @Description 获取任务统计数据
// @Tags 任务管理
// @Produce json
// @Security ApiKeyAuth
// @Param org_id query string false "组织ID"
// @Success 200 {object} utils.Response{data=map[string]interface{}}
// @Router /tasks/statistics [get]
func (h *TaskHandler) GetTaskStatistics(c *gin.Context) {
	orgID := c.Query("org_id")

	stats, err := h.taskService.GetTaskStatistics(c.Request.Context(), orgID)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, stats)
}

// GetTaskLogs 获取任务日志
// @Summary 获取任务日志
// @Description 获取任务的变更日志
// @Tags 任务管理
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "任务ID"
// @Success 200 {object} utils.Response{data=[]model.TaskLog}
// @Router /tasks/{id}/logs [get]
func (h *TaskHandler) GetTaskLogs(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}

	logs, err := h.taskService.GetTaskLogs(c.Request.Context(), id)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, logs)
}

// AddComment 添加评论
// @Summary 添加任务评论
// @Description 为任务添加评论
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "任务ID"
// @Param body body service.AddCommentRequest true "评论内容"
// @Success 200 {object} utils.Response{data=model.TaskComment}
// @Router /tasks/{id}/comments [post]
func (h *TaskHandler) AddComment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}

	var req service.AddCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	comment, err := h.taskService.AddComment(c.Request.Context(), id, uid, &req)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, comment)
}

// GetComments 获取评论列表
// @Summary 获取任务评论
// @Description 获取任务的评论列表
// @Tags 任务管理
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "任务ID"
// @Success 200 {object} utils.Response{data=[]model.TaskComment}
// @Router /tasks/{id}/comments [get]
func (h *TaskHandler) GetComments(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}

	comments, err := h.taskService.GetComments(c.Request.Context(), id)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, comments)
}

// BatchAssign 批量分配任务
// @Summary 批量分配任务
// @Description 批量将任务分配给指定用户
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body service.BatchAssignRequest true "批量分配信息"
// @Success 200 {object} utils.Response
// @Router /tasks/batch-assign [post]
func (h *TaskHandler) BatchAssign(c *gin.Context) {
	var req service.BatchAssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	uid, _ := uuid.Parse(userID.(string))

	if err := h.taskService.BatchAssign(c.Request.Context(), uid, &req); err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// AutoAssign 自动分配任务
// @Summary 自动分配任务
// @Description 根据负载自动分配待分配的任务
// @Tags 任务管理
// @Produce json
// @Security ApiKeyAuth
// @Param org_id query string false "组织ID"
// @Param limit query int false "限制数量" default(10)
// @Success 200 {object} utils.Response
// @Router /tasks/auto-assign [post]
func (h *TaskHandler) AutoAssign(c *gin.Context) {
	orgID := c.Query("org_id")
	
	limit := 10
	if l := c.DefaultQuery("limit", "10"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	if err := h.taskService.AutoAssignTasks(c.Request.Context(), orgID, limit); err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, nil)
}
