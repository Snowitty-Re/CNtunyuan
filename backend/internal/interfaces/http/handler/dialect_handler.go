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

		// 需要管理员权限
		dialects.PUT("/:id/status", middleware.RequireManager(), h.UpdateStatus)
		dialects.POST("/:id/feature", middleware.RequireAdmin(), h.Feature)
		dialects.DELETE("/:id/feature", middleware.RequireAdmin(), h.Unfeature)
	}
}

// Create 创建方言
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

// GetByID 获取详情
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

// List 列表查询
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

// Update 更新
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

// Delete 删除
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

// UpdateStatus 更新状态
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

// Play 播放记录
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

// Like 点赞
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

// GetComments 获取评论
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

// GetFeatured 获取精选
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

// GetStats 获取统计
func (h *DialectHandler) GetStats(c *gin.Context) {
	stats, err := h.dialectService.GetStats(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "failed to get stats")
		return
	}

	response.Success(c, stats)
}
