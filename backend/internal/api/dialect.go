package api

import (
	"strconv"

	"github.com/Snowitty-Re/CNtunyuan/internal/middleware"
	"github.com/Snowitty-Re/CNtunyuan/internal/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DialectHandler 方言处理器
type DialectHandler struct {
	dialectService *service.DialectService
}

// NewDialectHandler 创建方言处理器
func NewDialectHandler(dialectService *service.DialectService) *DialectHandler {
	return &DialectHandler{dialectService: dialectService}
}

// Create 创建方言记录
// @Summary 创建方言语音记录
// @Description 上传并创建新的方言语音记录
// @Tags 方言管理
// @Accept json
// @Produce json
// @Param body body service.CreateDialectRequest true "方言信息"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /dialects [post]
func (h *DialectHandler) Create(c *gin.Context) {
	var req service.CreateDialectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	collectorID := middleware.GetUserID(c)

	dialect, err := h.dialectService.Create(c.Request.Context(), &req, uuid.MustParse(collectorID))
	if err != nil {
		utils.ServerError(c, "创建失败: "+err.Error())
		return
	}

	utils.Created(c, dialect)
}

// Get 获取详情
// @Summary 获取方言详情
// @Description 根据ID获取方言详细信息
// @Tags 方言管理
// @Produce json
// @Param id path string true "记录ID"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /dialects/{id} [get]
func (h *DialectHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	dialect, err := h.dialectService.GetByID(c.Request.Context(), id)
	if err != nil {
		utils.NotFound(c, "记录不存在")
		return
	}

	utils.Success(c, dialect)
}

// Update 更新
// @Summary 更新方言记录
// @Description 更新方言语音记录
// @Tags 方言管理
// @Accept json
// @Produce json
// @Param id path string true "记录ID"
// @Param body body service.CreateDialectRequest true "方言信息"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /dialects/{id} [put]
func (h *DialectHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	var req service.CreateDialectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	dialect, err := h.dialectService.Update(c.Request.Context(), id, &req)
	if err != nil {
		utils.ServerError(c, "更新失败: "+err.Error())
		return
	}

	utils.Success(c, dialect)
}

// Delete 删除
// @Summary 删除方言记录
// @Description 删除方言语音记录
// @Tags 方言管理
// @Produce json
// @Param id path string true "记录ID"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /dialects/{id} [delete]
func (h *DialectHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	if err := h.dialectService.Delete(c.Request.Context(), id); err != nil {
		utils.ServerError(c, "删除失败: "+err.Error())
		return
	}

	utils.Success(c, nil)
}

// List 列表
// @Summary 获取方言列表
// @Description 分页获取方言语音记录列表
// @Tags 方言管理
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param province query string false "省份"
// @Param city query string false "城市"
// @Param keyword query string false "关键词"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /dialects [get]
func (h *DialectHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	filters := map[string]interface{}{}
	if province := c.Query("province"); province != "" {
		filters["province"] = province
	}
	if city := c.Query("city"); city != "" {
		filters["city"] = city
	}
	if district := c.Query("district"); district != "" {
		filters["district"] = district
	}
	if keyword := c.Query("keyword"); keyword != "" {
		filters["title"] = keyword
	}

	dialects, total, err := h.dialectService.List(c.Request.Context(), page, pageSize, filters)
	if err != nil {
		utils.ServerError(c, "获取列表失败")
		return
	}

	utils.PageSuccess(c, dialects, total, page, pageSize)
}

// Play 播放
// @Summary 播放方言
// @Description 记录方言播放
// @Tags 方言管理
// @Produce json
// @Param id path string true "记录ID"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /dialects/{id}/play [post]
func (h *DialectHandler) Play(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	userID := middleware.GetUserID(c)
	duration, _ := strconv.Atoi(c.DefaultQuery("duration", "0"))

	if err := h.dialectService.Play(c.Request.Context(), id, uuid.MustParse(userID), c.ClientIP(), duration); err != nil {
		utils.ServerError(c, "记录播放失败")
		return
	}

	utils.Success(c, nil)
}

// Like 点赞
// @Summary 点赞方言
// @Description 对方言语音记录点赞
// @Tags 方言管理
// @Produce json
// @Param id path string true "记录ID"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /dialects/{id}/like [post]
func (h *DialectHandler) Like(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	userID := middleware.GetUserID(c)

	if err := h.dialectService.Like(c.Request.Context(), id, uuid.MustParse(userID)); err != nil {
		utils.ServerError(c, "点赞失败")
		return
	}

	utils.Success(c, nil)
}

// Unlike 取消点赞
// @Summary 取消点赞
// @Description 取消对方言语音记录的点赞
// @Tags 方言管理
// @Produce json
// @Param id path string true "记录ID"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /dialects/{id}/unlike [post]
func (h *DialectHandler) Unlike(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	userID := middleware.GetUserID(c)

	if err := h.dialectService.Unlike(c.Request.Context(), id, uuid.MustParse(userID)); err != nil {
		utils.ServerError(c, "取消点赞失败")
		return
	}

	utils.Success(c, nil)
}

// GetNearby 获取附近方言
// @Summary 获取附近方言
// @Description 获取指定位置附近的方言记录
// @Tags 方言管理
// @Produce json
// @Param lat query number true "纬度"
// @Param lng query number true "经度"
// @Param radius query number false "半径(km)"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /dialects/nearby [get]
func (h *DialectHandler) GetNearby(c *gin.Context) {
	lat, err := strconv.ParseFloat(c.Query("lat"), 64)
	if err != nil {
		utils.BadRequest(c, "纬度参数错误")
		return
	}

	lng, err := strconv.ParseFloat(c.Query("lng"), 64)
	if err != nil {
		utils.BadRequest(c, "经度参数错误")
		return
	}

	radius, _ := strconv.ParseFloat(c.DefaultQuery("radius", "10"), 64)

	dialects, err := h.dialectService.GetNearbyDialects(c.Request.Context(), lat, lng, radius)
	if err != nil {
		utils.ServerError(c, "获取附近方言失败")
		return
	}

	utils.Success(c, dialects)
}

// GetStatistics 获取统计
// @Summary 获取方言统计
// @Description 获取方言相关统计数据
// @Tags 方言管理
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /dialects/statistics [get]
func (h *DialectHandler) GetStatistics(c *gin.Context) {
	stats, err := h.dialectService.GetStatistics(c.Request.Context())
	if err != nil {
		utils.ServerError(c, "获取统计失败")
		return
	}

	utils.Success(c, stats)
}
