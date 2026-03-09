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

// DialectHandler 方言处理器
type DialectHandler struct {
	dialectService *service.DialectAppService
}

// NewDialectHandler 创建方言处理器
func NewDialectHandler(dialectService *service.DialectAppService) *DialectHandler {
	return &DialectHandler{dialectService: dialectService}
}

// RegisterRoutes 注册路由
func (h *DialectHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	dialects := router.Group("/dialects")
	{
		// 公开接口（需要登录但不需要特殊权限）
		dialects.GET("", authMiddleware.Required(), h.List)
		dialects.GET("/featured", authMiddleware.Required(), h.GetFeatured)
		dialects.GET("/stats", authMiddleware.Required(), h.GetStats)
		dialects.GET("/:id", authMiddleware.Required(), h.GetByID)
		dialects.GET("/:id/comments", authMiddleware.Required(), h.GetComments)
		dialects.POST("/:id/play", authMiddleware.Required(), h.Play)
		dialects.POST("/:id/like", authMiddleware.Required(), h.Like)
		dialects.DELETE("/:id/like", authMiddleware.Required(), h.Unlike)
		dialects.POST("/:id/comments", authMiddleware.Required(), h.AddComment)

		// 需要上传权限
		dialects.POST("", authMiddleware.Required(), h.Create)
		dialects.PUT("/:id", authMiddleware.Required(), h.Update)
		dialects.DELETE("/:id", authMiddleware.Required(), h.Delete)

		// 需要管理员权限（RequireRole 检查依赖 authMiddleware.Required 设置的 userRole）
		dialects.PUT("/:id/status", authMiddleware.Required(), middleware.RequireManager(), h.UpdateStatus)
		dialects.POST("/:id/feature", authMiddleware.Required(), middleware.RequireAdmin(), h.Feature)
		dialects.DELETE("/:id/feature", authMiddleware.Required(), middleware.RequireAdmin(), h.Unfeature)
	}
}

// Create 创建方言
// @Summary 创建方言
// @Description 创建新的方言语音记录，需要登录权限
// @Tags 方言管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.CreateDialectRequest true "创建方言请求"
// @Success 201 {object} response.Response{data=dto.DialectResponse} "创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /dialects [post]
func (h *DialectHandler) Create(c *gin.Context) {
	var req dto.CreateDialectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	orgID := middleware.GetOrgID(c)

	dialect, err := h.dialectService.Create(c.Request.Context(), &req, userID, orgID)
	if err != nil {
		logger.Error("Failed to create dialect", logger.Err(err))
		response.InternalServerError(c, "failed to create dialect")
		return
	}

	response.Created(c, dialect)
}

// GetByID 获取方言详情
// @Summary 获取方言详情
// @Description 根据ID获取方言详细信息，需要登录权限
// @Tags 方言管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "方言ID"
// @Success 200 {object} response.Response{data=dto.DialectResponse} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "方言不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /dialects/{id} [get]
func (h *DialectHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "dialect id is required")
		return
	}

	dialect, err := h.dialectService.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrDialectNotFound {
			response.NotFound(c, "dialect not found")
			return
		}
		response.InternalServerError(c, "failed to get dialect")
		return
	}

	response.Success(c, dialect)
}

// List 获取方言列表
// @Summary 获取方言列表
// @Description 分页获取方言列表，支持关键词、地区、状态等筛选，需要登录权限
// @Tags 方言管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码，默认为1" default(1)
// @Param page_size query int false "每页数量，默认为10" default(10)
// @Param keyword query string false "关键词搜索"
// @Param region query string false "地区筛选"
// @Param province query string false "省份筛选"
// @Param city query string false "城市筛选"
// @Param type query string false "方言类型"
// @Param status query string false "状态筛选"
// @Param sort_by query string false "排序字段"
// @Param sort_order query string false "排序方式(asc/desc)"
// @Success 200 {object} response.Response{data=dto.DialectListResponse} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /dialects [get]
func (h *DialectHandler) List(c *gin.Context) {
	var req dto.DialectListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	dialects, err := h.dialectService.List(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "failed to get list")
		return
	}

	response.Success(c, dialects)
}

// Update 更新方言
// @Summary 更新方言
// @Description 更新指定ID的方言信息，需要登录权限
// @Tags 方言管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "方言ID"
// @Param request body dto.UpdateDialectRequest true "更新方言请求"
// @Success 200 {object} response.Response{data=dto.DialectResponse} "更新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "方言不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /dialects/{id} [put]
func (h *DialectHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "dialect id is required")
		return
	}

	var req dto.UpdateDialectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	dialect, err := h.dialectService.Update(c.Request.Context(), id, &req)
	if err != nil {
		switch err {
		case service.ErrDialectNotFound:
			response.NotFound(c, "dialect not found")
		default:
			logger.Error("Failed to update dialect", logger.Err(err))
			response.InternalServerError(c, "failed to update")
		}
		return
	}

	response.Success(c, dialect)
}

// Delete 删除方言
// @Summary 删除方言
// @Description 删除指定ID的方言，需要登录权限
// @Tags 方言管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "方言ID"
// @Success 204 "删除成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /dialects/{id} [delete]
func (h *DialectHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "dialect id is required")
		return
	}

	if err := h.dialectService.Delete(c.Request.Context(), id); err != nil {
		logger.Error("Failed to delete dialect", logger.Err(err))
		response.InternalServerError(c, "failed to delete")
		return
	}

	response.NoContent(c)
}

// UpdateStatus 更新方言状态
// @Summary 更新方言状态
// @Description 更新方言的审核状态，需要管理员或经理权限
// @Tags 方言管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "方言ID"
// @Param request body dto.UpdateDialectStatusRequest true "状态更新请求"
// @Success 200 {object} response.Response "更新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 404 {object} response.Response "方言不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /dialects/{id}/status [put]
func (h *DialectHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "dialect id is required")
		return
	}

	var req dto.UpdateDialectStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.dialectService.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		switch err {
		case service.ErrDialectNotFound:
			response.NotFound(c, "dialect not found")
		default:
			response.InternalServerError(c, "failed to update status")
		}
		return
	}

	response.Success(c, nil)
}

// Feature 设为精选
// @Summary 将方言设为精选
// @Description 将指定方言标记为精选内容，需要管理员权限
// @Tags 方言管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "方言ID"
// @Success 200 {object} response.Response "设置成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 404 {object} response.Response "方言不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /dialects/{id}/feature [post]
func (h *DialectHandler) Feature(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "dialect id is required")
		return
	}

	if err := h.dialectService.Feature(c.Request.Context(), id); err != nil {
		switch err {
		case service.ErrDialectNotFound:
			response.NotFound(c, "dialect not found")
		default:
			response.InternalServerError(c, "failed to feature dialect")
		}
		return
	}

	response.Success(c, nil)
}

// Unfeature 取消精选
// @Summary 取消方言精选
// @Description 取消指定方言的精选标记，需要管理员权限
// @Tags 方言管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "方言ID"
// @Success 200 {object} response.Response "取消成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 404 {object} response.Response "方言不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /dialects/{id}/feature [delete]
func (h *DialectHandler) Unfeature(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "dialect id is required")
		return
	}

	if err := h.dialectService.Unfeature(c.Request.Context(), id); err != nil {
		switch err {
		case service.ErrDialectNotFound:
			response.NotFound(c, "dialect not found")
		default:
			response.InternalServerError(c, "failed to unfeature dialect")
		}
		return
	}

	response.Success(c, nil)
}

// Play 记录播放
// @Summary 记录方言播放
// @Description 记录用户播放方言的次数，需要登录权限
// @Tags 方言管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "方言ID"
// @Success 200 {object} response.Response "记录成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /dialects/{id}/play [post]
func (h *DialectHandler) Play(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "dialect id is required")
		return
	}

	if err := h.dialectService.IncrementPlayCount(c.Request.Context(), id); err != nil {
		response.InternalServerError(c, "failed to record play")
		return
	}

	response.Success(c, nil)
}

// Like 点赞方言
// @Summary 点赞方言
// @Description 为指定方言点赞，需要登录权限
// @Tags 方言管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "方言ID"
// @Success 200 {object} response.Response "点赞成功"
// @Failure 400 {object} response.Response "请求参数错误或已点赞"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "方言不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /dialects/{id}/like [post]
func (h *DialectHandler) Like(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "dialect id is required")
		return
	}

	userID := middleware.GetUserID(c)

	if err := h.dialectService.Like(c.Request.Context(), id, userID); err != nil {
		switch err {
		case service.ErrAlreadyLiked:
			response.BadRequest(c, "already liked")
		case service.ErrDialectNotFound:
			response.NotFound(c, "dialect not found")
		default:
			response.InternalServerError(c, "failed to like")
		}
		return
	}

	response.Success(c, nil)
}

// Unlike 取消点赞
// @Summary 取消点赞方言
// @Description 取消对指定方言的点赞，需要登录权限
// @Tags 方言管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "方言ID"
// @Success 200 {object} response.Response "取消点赞成功"
// @Failure 400 {object} response.Response "请求参数错误或未点赞"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "方言不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /dialects/{id}/like [delete]
func (h *DialectHandler) Unlike(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "dialect id is required")
		return
	}

	userID := middleware.GetUserID(c)

	if err := h.dialectService.Unlike(c.Request.Context(), id, userID); err != nil {
		switch err {
		case service.ErrNotLiked:
			response.BadRequest(c, "not liked yet")
		case service.ErrDialectNotFound:
			response.NotFound(c, "dialect not found")
		default:
			response.InternalServerError(c, "failed to unlike")
		}
		return
	}

	response.Success(c, nil)
}

// AddComment 添加评论
// @Summary 添加方言评论
// @Description 为指定方言添加评论，需要登录权限
// @Tags 方言管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "方言ID"
// @Param request body dto.CreateDialectCommentRequest true "评论请求"
// @Success 201 {object} response.Response{data=dto.DialectCommentResponse} "评论成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "方言不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /dialects/{id}/comments [post]
func (h *DialectHandler) AddComment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "dialect id is required")
		return
	}

	var req dto.CreateDialectCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)

	comment, err := h.dialectService.AddComment(c.Request.Context(), id, &req, userID)
	if err != nil {
		switch err {
		case service.ErrDialectNotFound:
			response.NotFound(c, "dialect not found")
		default:
			response.InternalServerError(c, "failed to add comment")
		}
		return
	}

	response.Created(c, comment)
}

// GetComments 获取评论列表
// @Summary 获取方言评论列表
// @Description 获取指定方言的评论列表，支持分页，需要登录权限
// @Tags 方言管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "方言ID"
// @Param page query int false "页码，默认为1" default(1)
// @Param page_size query int false "每页数量，默认为10" default(10)
// @Success 200 {object} response.Response{data=repository.PageResult[dto.DialectCommentResponse]} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /dialects/{id}/comments [get]
func (h *DialectHandler) GetComments(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "dialect id is required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	comments, err := h.dialectService.GetComments(c.Request.Context(), id, page, pageSize)
	if err != nil {
		response.InternalServerError(c, "failed to get comments")
		return
	}

	response.Success(c, comments)
}

// GetFeatured 获取精选列表
// @Summary 获取精选方言列表
// @Description 分页获取标记为精选的方言列表，需要登录权限
// @Tags 方言管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码，默认为1" default(1)
// @Param page_size query int false "每页数量，默认为10" default(10)
// @Success 200 {object} response.Response{data=repository.PageResult[dto.DialectResponse]} "获取成功"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /dialects/featured [get]
func (h *DialectHandler) GetFeatured(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	dialects, err := h.dialectService.GetFeatured(c.Request.Context(), page, pageSize)
	if err != nil {
		response.InternalServerError(c, "failed to get featured dialects")
		return
	}

	response.Success(c, dialects)
}

// GetStats 获取方言统计
// @Summary 获取方言统计数据
// @Description 获取方言相关的统计数据（总数、活跃数、待审核数、精选数、总播放量、总点赞数），需要登录权限
// @Tags 方言管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=dto.DialectStatsResponse} "获取成功"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /dialects/stats [get]
func (h *DialectHandler) GetStats(c *gin.Context) {
	stats, err := h.dialectService.GetStats(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "failed to get stats")
		return
	}

	response.Success(c, stats)
}
