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

// MissingPersonHandler 走失人员处理器
type MissingPersonHandler struct {
	mpService *service.MissingPersonAppService
}

// NewMissingPersonHandler 创建走失人员处理器
func NewMissingPersonHandler(mpService *service.MissingPersonAppService) *MissingPersonHandler {
	return &MissingPersonHandler{mpService: mpService}
}

// RegisterRoutes 注册路由
func (h *MissingPersonHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	mps := router.Group("/missing-persons")
	mps.Use(authMiddleware.Required())
	{
		mps.GET("", h.List)
		mps.GET("/search", h.Search)
		mps.GET("/stats", h.GetStats)
		mps.GET("/:id", h.GetByID)
		mps.GET("/:id/tracks", h.GetTracks)
		mps.POST("", h.Create)
		mps.PUT("/:id", h.Update)
		mps.DELETE("/:id", middleware.RequireManager(), h.Delete)
		mps.PUT("/:id/status", middleware.RequireManager(), h.UpdateStatus)
		mps.POST("/:id/found", middleware.RequireManager(), h.MarkFound)
		mps.POST("/:id/reunited", middleware.RequireManager(), h.MarkReunited)
		mps.POST("/:id/tracks", h.AddTrack)
	}
}

// Create 创建走失人员
func (h *MissingPersonHandler) Create(c *gin.Context) {
	var req dto.CreateMissingPersonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	orgID := middleware.GetOrgID(c)

	mp, err := h.mpService.Create(c.Request.Context(), &req, userID, orgID)
	if err != nil {
		logger.Error("Failed to create missing person", logger.Err(err))
		response.InternalServerError(c, "failed to create missing person")
		return
	}

	response.Created(c, mp)
}

// GetByID 获取详情
func (h *MissingPersonHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "missing person id is required")
		return
	}

	mp, err := h.mpService.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrMissingPersonNotFound {
			response.NotFound(c, "missing person not found")
			return
		}
		response.InternalServerError(c, "failed to get missing person")
		return
	}

	response.Success(c, mp)
}

// List 列表查询
func (h *MissingPersonHandler) List(c *gin.Context) {
	var req dto.MissingPersonListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	mps, err := h.mpService.List(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "failed to get list")
		return
	}

	response.Success(c, mps)
}

// Update 更新
func (h *MissingPersonHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "missing person id is required")
		return
	}

	var req dto.UpdateMissingPersonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	mp, err := h.mpService.Update(c.Request.Context(), id, &req)
	if err != nil {
		switch err {
		case service.ErrMissingPersonNotFound:
			response.NotFound(c, "missing person not found")
		default:
			logger.Error("Failed to update missing person", logger.Err(err))
			response.InternalServerError(c, "failed to update")
		}
		return
	}

	response.Success(c, mp)
}

// Delete 删除
func (h *MissingPersonHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "missing person id is required")
		return
	}

	if err := h.mpService.Delete(c.Request.Context(), id); err != nil {
		logger.Error("Failed to delete missing person", logger.Err(err))
		response.InternalServerError(c, "failed to delete")
		return
	}

	response.NoContent(c)
}

// UpdateStatus 更新状态
func (h *MissingPersonHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "missing person id is required")
		return
	}

	var req dto.UpdateMissingPersonStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.mpService.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		switch err {
		case service.ErrMissingPersonNotFound:
			response.NotFound(c, "missing person not found")
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.Success(c, nil)
}

// MarkFound 标记找到
func (h *MissingPersonHandler) MarkFound(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "missing person id is required")
		return
	}

	var req dto.MarkFoundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.mpService.MarkFound(c.Request.Context(), id, &req); err != nil {
		switch err {
		case service.ErrMissingPersonNotFound:
			response.NotFound(c, "missing person not found")
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.Success(c, nil)
}

// MarkReunited 标记团聚
func (h *MissingPersonHandler) MarkReunited(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "missing person id is required")
		return
	}

	if err := h.mpService.MarkReunited(c.Request.Context(), id); err != nil {
		switch err {
		case service.ErrMissingPersonNotFound:
			response.NotFound(c, "missing person not found")
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.Success(c, nil)
}

// AddTrack 添加轨迹
func (h *MissingPersonHandler) AddTrack(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "missing person id is required")
		return
	}

	var req dto.CreateMissingPersonTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)

	track, err := h.mpService.AddTrack(c.Request.Context(), id, &req, userID)
	if err != nil {
		switch err {
		case service.ErrMissingPersonNotFound:
			response.NotFound(c, "missing person not found")
		default:
			response.InternalServerError(c, "failed to add track")
		}
		return
	}

	response.Created(c, track)
}

// GetTracks 获取轨迹
func (h *MissingPersonHandler) GetTracks(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "missing person id is required")
		return
	}

	tracks, err := h.mpService.GetTracks(c.Request.Context(), id)
	if err != nil {
		response.InternalServerError(c, "failed to get tracks")
		return
	}

	response.Success(c, tracks)
}

// GetStats 获取统计
func (h *MissingPersonHandler) GetStats(c *gin.Context) {
	stats, err := h.mpService.GetStats(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "failed to get stats")
		return
	}

	response.Success(c, stats)
}

// Search 搜索
func (h *MissingPersonHandler) Search(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		response.BadRequest(c, "keyword is required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	result, err := h.mpService.Search(c.Request.Context(), keyword, page, pageSize)
	if err != nil {
		response.InternalServerError(c, "search failed")
		return
	}

	response.Success(c, result)
}
