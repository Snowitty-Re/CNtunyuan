package handler

import (
	"errors"
	"strconv"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/application/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
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
		dialects.GET("/stats", authMiddleware.Required(), middleware.RequireManager(), h.GetStats)
		dialects.GET("/:id", authMiddleware.Required(), h.GetByID)
		dialects.GET("/:id/comments", authMiddleware.Required(), h.GetComments)
		dialects.POST("/:id/play", authMiddleware.Required(), h.Play)
		dialects.POST("/:id/like", authMiddleware.Required(), h.Like)
		dialects.DELETE("/:id/like", authMiddleware.Required(), h.Unlike)
		dialects.POST("/:id/comments", authMiddleware.Required(), h.AddComment)

		// 需要上传权限
		dialects.POST("", authMiddleware.Required(), h.Create)
		dialects.POST("/batch", authMiddleware.Required(), h.CreateBatch)
		dialects.PUT("/:id", authMiddleware.Required(), h.Update)
		dialects.DELETE("/:id", authMiddleware.Required(), h.Delete)

		// 需要管理员权限（RequireRole 检查依赖 authMiddleware.Required 设置的 userRole）
		dialects.PUT("/:id/status", authMiddleware.Required(), middleware.RequireManager(), h.UpdateStatus)
		dialects.POST("/:id/feature", authMiddleware.Required(), middleware.RequireAdmin(), h.Feature)
		dialects.DELETE("/:id/feature", authMiddleware.Required(), middleware.RequireAdmin(), h.Unfeature)
	}

	dialectCards := router.Group("/dialect-cards")
	{
		dialectCards.GET("/template", authMiddleware.Required(), h.ListCardTemplate)

		dialectCards.GET("/groups", authMiddleware.Required(), middleware.RequireManager(), h.ListCardGroups)
		dialectCards.POST("/groups", authMiddleware.Required(), middleware.RequireManager(), h.CreateCardGroup)
		dialectCards.PUT("/groups/:id", authMiddleware.Required(), middleware.RequireManager(), h.UpdateCardGroup)
		dialectCards.DELETE("/groups/:id", authMiddleware.Required(), middleware.RequireManager(), h.DeleteCardGroup)

		dialectCards.GET("", authMiddleware.Required(), middleware.RequireManager(), h.ListCards)
		dialectCards.POST("", authMiddleware.Required(), middleware.RequireManager(), h.CreateCard)
		dialectCards.PUT("/:id", authMiddleware.Required(), middleware.RequireManager(), h.UpdateCard)
		dialectCards.DELETE("/:id", authMiddleware.Required(), middleware.RequireManager(), h.DeleteCard)
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
		response.Error(c, err)
		return
	}

	response.Created(c, dialect)
}

// CreateBatch 批量创建方言（按模板卡片分段）
func (h *DialectHandler) CreateBatch(c *gin.Context) {
	var req dto.CreateDialectBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	orgID := middleware.GetOrgID(c)

	result, err := h.dialectService.CreateBatch(c.Request.Context(), &req, userID, orgID)
	if err != nil {
		logger.Error("Failed to create dialect batch", logger.Err(err))
		response.Error(c, err)
		return
	}
	response.Created(c, result)
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

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	isManager := entity.GetRoleLevel(role) >= entity.RoleLevelManager
	dialect, err := h.dialectService.GetByID(c.Request.Context(), id, userID, isManager)
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

	operator := &entity.User{
		BaseEntity: entity.BaseEntity{ID: middleware.GetUserID(c)},
		Role:       middleware.GetUserRole(c),
		OrgID:      middleware.GetOrgID(c),
	}

	dialects, err := h.dialectService.List(c.Request.Context(), &req, operator)
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

	operator := &entity.User{
		BaseEntity: entity.BaseEntity{ID: middleware.GetUserID(c)},
		Role:       middleware.GetUserRole(c),
		OrgID:      middleware.GetOrgID(c),
	}

	dialect, err := h.dialectService.Update(c.Request.Context(), id, &req, operator)
	if err != nil {
		switch err {
		case service.ErrDialectNotFound:
			response.NotFound(c, "dialect not found")
		case service.ErrDialectForbidden:
			response.Forbidden(c, "no permission to modify this dialect")
		default:
			logger.Error("Failed to update dialect", logger.Err(err))
			response.Error(c, err)
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

	operator := &entity.User{
		BaseEntity: entity.BaseEntity{ID: middleware.GetUserID(c)},
		Role:       middleware.GetUserRole(c),
		OrgID:      middleware.GetOrgID(c),
	}

	if err := h.dialectService.Delete(c.Request.Context(), id, operator); err != nil {
		switch err {
		case service.ErrDialectNotFound:
			response.NotFound(c, "dialect not found")
		case service.ErrDialectForbidden:
			response.Forbidden(c, "no permission to delete this dialect")
		default:
			logger.Error("Failed to delete dialect", logger.Err(err))
			response.InternalServerError(c, "failed to delete")
		}
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

	operator := &entity.User{
		BaseEntity: entity.BaseEntity{ID: middleware.GetUserID(c)},
		Role:       middleware.GetUserRole(c),
		OrgID:      middleware.GetOrgID(c),
	}

	if err := h.dialectService.UpdateStatus(c.Request.Context(), id, req.Status, operator); err != nil {
		switch {
		case err == service.ErrDialectNotFound:
			response.NotFound(c, "dialect not found")
		case errors.Is(err, service.ErrDialectInvalidStatus):
			response.BadRequest(c, err.Error())
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

	operator := &entity.User{
		BaseEntity: entity.BaseEntity{ID: middleware.GetUserID(c)},
		Role:       middleware.GetUserRole(c),
		OrgID:      middleware.GetOrgID(c),
	}

	if err := h.dialectService.Feature(c.Request.Context(), id, operator); err != nil {
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

	operator := &entity.User{
		BaseEntity: entity.BaseEntity{ID: middleware.GetUserID(c)},
		Role:       middleware.GetUserRole(c),
		OrgID:      middleware.GetOrgID(c),
	}

	if err := h.dialectService.Unfeature(c.Request.Context(), id, operator); err != nil {
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

	role := middleware.GetUserRole(c)
	isManager := entity.GetRoleLevel(role) >= entity.RoleLevelManager

	if err := h.dialectService.IncrementPlayCount(c.Request.Context(), id, isManager); err != nil {
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
	role := middleware.GetUserRole(c)
	isManager := entity.GetRoleLevel(role) >= entity.RoleLevelManager

	if err := h.dialectService.Like(c.Request.Context(), id, userID, isManager); err != nil {
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
	role := middleware.GetUserRole(c)
	isManager := entity.GetRoleLevel(role) >= entity.RoleLevelManager

	if err := h.dialectService.Unlike(c.Request.Context(), id, userID, isManager); err != nil {
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
	role := middleware.GetUserRole(c)
	isManager := entity.GetRoleLevel(role) >= entity.RoleLevelManager

	comment, err := h.dialectService.AddComment(c.Request.Context(), id, &req, userID, isManager)
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
// @Success 200 {object} response.Response{data=dto.DialectCommentListResponse} "获取成功"
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
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	role := middleware.GetUserRole(c)
	isManager := entity.GetRoleLevel(role) >= entity.RoleLevelManager

	comments, err := h.dialectService.GetComments(c.Request.Context(), id, page, pageSize, isManager)
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
// @Success 200 {object} response.Response{data=dto.DialectListResponse} "获取成功"
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
	role := middleware.GetUserRole(c)
	isManager := entity.GetRoleLevel(role) >= entity.RoleLevelManager

	stats, err := h.dialectService.GetStats(c.Request.Context(), isManager)
	if err != nil {
		response.InternalServerError(c, "failed to get stats")
		return
	}

	response.Success(c, stats)
}

// ListCardTemplate 获取方言卡片模板（分组 + 卡片）
func (h *DialectHandler) ListCardTemplate(c *gin.Context) {
	includeInactive := c.Query("include_inactive") == "1" || c.Query("include_inactive") == "true"
	role := middleware.GetUserRole(c)
	if entity.GetRoleLevel(role) < entity.RoleLevelManager {
		includeInactive = false
	}
	groups, err := h.dialectService.ListCardTemplate(c.Request.Context(), includeInactive)
	if err != nil {
		response.InternalServerError(c, "failed to get card template")
		return
	}
	response.Success(c, gin.H{"groups": groups})
}

// ListCardGroups 获取卡片分组列表（管理）
func (h *DialectHandler) ListCardGroups(c *gin.Context) {
	groups, err := h.dialectService.ListCardTemplate(c.Request.Context(), true)
	if err != nil {
		response.InternalServerError(c, "failed to get card groups")
		return
	}
	response.Success(c, gin.H{"groups": groups})
}

// CreateCardGroup 创建卡片分组
func (h *DialectHandler) CreateCardGroup(c *gin.Context) {
	var req dto.CreateDialectCardGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	operator := &entity.User{
		BaseEntity: entity.BaseEntity{ID: middleware.GetUserID(c)},
		Role:       middleware.GetUserRole(c),
		OrgID:      middleware.GetOrgID(c),
	}
	group, err := h.dialectService.CreateCardGroup(c.Request.Context(), &req, operator)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, group)
}

// UpdateCardGroup 更新卡片分组
func (h *DialectHandler) UpdateCardGroup(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "group id is required")
		return
	}
	var req dto.UpdateDialectCardGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	group, err := h.dialectService.UpdateCardGroup(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, group)
}

// DeleteCardGroup 删除卡片分组
func (h *DialectHandler) DeleteCardGroup(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "group id is required")
		return
	}
	if err := h.dialectService.DeleteCardGroup(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// ListCards 获取卡片列表（管理）
func (h *DialectHandler) ListCards(c *gin.Context) {
	groupID := c.Query("group_id")
	groups, err := h.dialectService.ListCardTemplate(c.Request.Context(), true)
	if err != nil {
		response.InternalServerError(c, "failed to get cards")
		return
	}
	if groupID == "" {
		response.Success(c, gin.H{"groups": groups})
		return
	}
	filtered := make([]dto.DialectCardGroupResponse, 0, len(groups))
	for _, group := range groups {
		if group.ID == groupID {
			filtered = append(filtered, group)
			break
		}
	}
	response.Success(c, gin.H{"groups": filtered})
}

// CreateCard 创建卡片
func (h *DialectHandler) CreateCard(c *gin.Context) {
	var req dto.CreateDialectCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	operator := &entity.User{
		BaseEntity: entity.BaseEntity{ID: middleware.GetUserID(c)},
		Role:       middleware.GetUserRole(c),
		OrgID:      middleware.GetOrgID(c),
	}
	card, err := h.dialectService.CreateCard(c.Request.Context(), &req, operator)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, card)
}

// UpdateCard 更新卡片
func (h *DialectHandler) UpdateCard(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "card id is required")
		return
	}
	var req dto.UpdateDialectCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	card, err := h.dialectService.UpdateCard(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, card)
}

// DeleteCard 删除卡片
func (h *DialectHandler) DeleteCard(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "card id is required")
		return
	}
	if err := h.dialectService.DeleteCard(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}
