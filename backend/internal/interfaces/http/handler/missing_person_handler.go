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
// @Summary 创建走失人员
// @Description 创建新的走失人员档案
// @Tags 走失人员
// @Accept json
// @Produce json
// @Param request body dto.CreateMissingPersonRequest true "创建走失人员请求参数"
// @Success 201 {object} response.Response{data=dto.MissingPersonResponse} "创建成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /missing-persons [post]
// @Security Bearer
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
// @Summary 获取走失人员详情
// @Description 根据ID获取走失人员详细信息
// @Tags 走失人员
// @Accept json
// @Produce json
// @Param id path string true "走失人员ID"
// @Success 200 {object} response.Response{data=dto.MissingPersonResponse} "获取成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "走失人员不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /missing-persons/{id} [get]
// @Security Bearer
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
// @Summary 获取走失人员列表
// @Description 分页获取走失人员列表，支持多种筛选条件
// @Tags 走失人员
// @Accept json
// @Produce json
// @Param page query int false "页码，默认1" default(1)
// @Param page_size query int false "每页数量，默认10" default(10)
// @Param keyword query string false "关键词搜索"
// @Param status query string false "状态筛选：missing(走失中), searching(搜寻中), found(已找到), reunited(已团聚), closed(已关闭)"
// @Param gender query string false "性别：male(男), female(女), other(其他)"
// @Param age_min query int false "最小年龄"
// @Param age_max query int false "最大年龄"
// @Param province query string false "省份"
// @Param city query string false "城市"
// @Param district query string false "区县"
// @Param urgency_level query string false "紧急程度：critical(紧急), high(高), medium(中), low(低)"
// @Param case_type query string false "案件类型：elderly(老人), child(儿童), adult(成人), disability(残障), other(其他)"
// @Success 200 {object} response.Response{data=dto.MissingPersonListResponse} "获取成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /missing-persons [get]
// @Security Bearer
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
// @Summary 更新走失人员信息
// @Description 更新指定走失人员的详细信息
// @Tags 走失人员
// @Accept json
// @Produce json
// @Param id path string true "走失人员ID"
// @Param request body dto.UpdateMissingPersonRequest true "更新走失人员请求参数"
// @Success 200 {object} response.Response{data=dto.MissingPersonResponse} "更新成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "走失人员不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /missing-persons/{id} [put]
// @Security Bearer
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

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	isManager := entity.GetRoleLevel(role) >= entity.RoleLevelManager

	mp, err := h.mpService.Update(c.Request.Context(), id, &req, userID, isManager)
	if err != nil {
		switch err {
		case service.ErrMissingPersonNotFound:
			response.NotFound(c, "missing person not found")
		case service.ErrMissingPersonForbidden:
			response.Forbidden(c, "no permission to modify this case")
		default:
			logger.Error("Failed to update missing person", logger.Err(err))
			response.InternalServerError(c, "failed to update")
		}
		return
	}

	response.Success(c, mp)
}

// Delete 删除
// @Summary 删除走失人员
// @Description 删除指定的走失人员档案（需要管理员权限）
// @Tags 走失人员
// @Accept json
// @Produce json
// @Param id path string true "走失人员ID"
// @Success 204 "删除成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足，需要管理员权限"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /missing-persons/{id} [delete]
// @Security Bearer
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
// @Summary 更新走失人员状态
// @Description 更新走失人员的状态（需要管理员权限）
// @Tags 走失人员
// @Accept json
// @Produce json
// @Param id path string true "走失人员ID"
// @Param request body dto.UpdateMissingPersonStatusRequest true "状态更新请求参数"
// @Success 200 {object} response.Response "更新成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足，需要管理员权限"
// @Failure 404 {object} response.Response "走失人员不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /missing-persons/{id}/status [put]
// @Security Bearer
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
// @Summary 标记走失人员已找到
// @Description 标记指定走失人员已被找到，可填写发现地点和备注（需要管理员权限）
// @Tags 走失人员
// @Accept json
// @Produce json
// @Param id path string true "走失人员ID"
// @Param request body dto.MarkFoundRequest false "标记找到请求参数"
// @Success 200 {object} response.Response "标记成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足，需要管理员权限"
// @Failure 404 {object} response.Response "走失人员不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /missing-persons/{id}/found [post]
// @Security Bearer
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
// @Summary 标记走失人员已团聚
// @Description 标记指定走失人员已与家属团聚（需要管理员权限）
// @Tags 走失人员
// @Accept json
// @Produce json
// @Param id path string true "走失人员ID"
// @Success 200 {object} response.Response "标记成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足，需要管理员权限"
// @Failure 404 {object} response.Response "走失人员不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /missing-persons/{id}/reunited [post]
// @Security Bearer
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
// @Summary 添加轨迹记录
// @Description 为指定走失人员添加新的轨迹/线索记录
// @Tags 走失人员
// @Accept json
// @Produce json
// @Param id path string true "走失人员ID"
// @Param request body dto.CreateMissingPersonTrackRequest true "创建轨迹请求参数"
// @Success 201 {object} response.Response{data=dto.MissingPersonTrackResponse} "创建成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "走失人员不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /missing-persons/{id}/tracks [post]
// @Security Bearer
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
// @Summary 获取轨迹记录列表
// @Description 获取指定走失人员的所有轨迹/线索记录
// @Tags 走失人员
// @Accept json
// @Produce json
// @Param id path string true "走失人员ID"
// @Success 200 {object} response.Response{data=[]dto.MissingPersonTrackResponse} "获取成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /missing-persons/{id}/tracks [get]
// @Security Bearer
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
// @Summary 获取走失人员统计数据
// @Description 获取走失人员的统计信息，包括总数、各状态数量等
// @Tags 走失人员
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=dto.MissingPersonStatsResponse} "获取成功"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /missing-persons/stats [get]
// @Security Bearer
func (h *MissingPersonHandler) GetStats(c *gin.Context) {
	stats, err := h.mpService.GetStats(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "failed to get stats")
		return
	}

	response.Success(c, stats)
}

// Search 搜索
// @Summary 搜索走失人员
// @Description 根据关键词搜索走失人员，支持分页
// @Tags 走失人员
// @Accept json
// @Produce json
// @Param keyword query string true "搜索关键词"
// @Param page query int false "页码，默认1" default(1)
// @Param page_size query int false "每页数量，默认10" default(10)
// @Success 200 {object} response.Response{data=dto.MissingPersonListResponse} "搜索成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /missing-persons/search [get]
// @Security Bearer
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
