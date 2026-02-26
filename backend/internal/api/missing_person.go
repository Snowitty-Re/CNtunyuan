package api

import (
	"strconv"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/middleware"
	"github.com/Snowitty-Re/CNtunyuan/internal/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MissingPersonHandler 走失人员处理器
type MissingPersonHandler struct {
	mpService *service.MissingPersonService
}

// NewMissingPersonHandler 创建处理器
func NewMissingPersonHandler(mpService *service.MissingPersonService) *MissingPersonHandler {
	return &MissingPersonHandler{mpService: mpService}
}

// Create 创建走失人员记录
// @Summary 创建走失人员记录
// @Description 登记新的走失人员信息
// @Tags 走失人员管理
// @Accept json
// @Produce json
// @Param body body service.CreateMissingPersonRequest true "走失人员信息"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /missing-persons [post]
func (h *MissingPersonHandler) Create(c *gin.Context) {
	var req service.CreateMissingPersonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	reporterID := middleware.GetUserID(c)
	if reporterID == "" {
		utils.Unauthorized(c, "未登录")
		return
	}

	mp, err := h.mpService.Create(c.Request.Context(), &req, uuid.MustParse(reporterID))
	if err != nil {
		utils.ServerError(c, "创建失败: "+err.Error())
		return
	}

	utils.Created(c, mp)
}

// Get 获取详情
// @Summary 获取走失人员详情
// @Description 根据ID获取走失人员详细信息
// @Tags 走失人员管理
// @Produce json
// @Param id path string true "记录ID"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /missing-persons/{id} [get]
func (h *MissingPersonHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	mp, err := h.mpService.GetByID(c.Request.Context(), id)
	if err != nil {
		utils.NotFound(c, "记录不存在")
		return
	}

	utils.Success(c, mp)
}

// List 获取列表
// @Summary 获取走失人员列表
// @Description 分页获取走失人员列表
// @Tags 走失人员管理
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param status query string false "状态"
// @Param case_type query string false "案件类型"
// @Param keyword query string false "关键词"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /missing-persons [get]
func (h *MissingPersonHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	filters := map[string]interface{}{}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if caseType := c.Query("case_type"); caseType != "" {
		filters["case_type"] = caseType
	}
	if keyword := c.Query("keyword"); keyword != "" {
		filters["name"] = keyword
	}
	if orgID := c.Query("org_id"); orgID != "" {
		filters["org_id"] = orgID
	}

	mps, total, err := h.mpService.List(c.Request.Context(), page, pageSize, filters)
	if err != nil {
		utils.ServerError(c, "获取列表失败")
		return
	}

	utils.PageSuccess(c, mps, total, page, pageSize)
}

// UpdateStatus 更新状态
// @Summary 更新案件状态
// @Description 更新走失人员案件状态
// @Tags 走失人员管理
// @Accept json
// @Produce json
// @Param id path string true "记录ID"
// @Param status query string true "状态"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /missing-persons/{id}/status [put]
func (h *MissingPersonHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	status := c.Query("status")
	if status == "" {
		utils.BadRequest(c, "状态不能为空")
		return
	}

	foundDetail := c.PostForm("found_detail")

	if err := h.mpService.UpdateStatus(c.Request.Context(), id, status, foundDetail); err != nil {
		utils.ServerError(c, "更新失败: "+err.Error())
		return
	}

	utils.Success(c, nil)
}

// GetNearby 获取附近案件
// @Summary 获取附近案件
// @Description 获取指定位置附近的走失人员案件
// @Tags 走失人员管理
// @Produce json
// @Param lat query number true "纬度"
// @Param lng query number true "经度"
// @Param radius query number false "半径(km)"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /missing-persons/nearby [get]
func (h *MissingPersonHandler) GetNearby(c *gin.Context) {
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

	mps, err := h.mpService.GetNearbyCases(c.Request.Context(), lat, lng, radius)
	if err != nil {
		utils.ServerError(c, "获取附近案件失败")
		return
	}

	utils.Success(c, mps)
}

// AddTrack 添加轨迹
// @Summary 添加轨迹记录
// @Description 添加走失人员轨迹记录
// @Tags 走失人员管理
// @Accept json
// @Produce json
// @Param id path string true "记录ID"
// @Param body body AddTrackRequest true "轨迹信息"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /missing-persons/{id}/tracks [post]
func (h *MissingPersonHandler) AddTrack(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	var req AddTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	reporterID := middleware.GetUserID(c)

	trackTime, _ := time.Parse(time.RFC3339, req.TrackTime)
	if err := h.mpService.AddTrack(c.Request.Context(), id, trackTime, req.Location, req.Longitude, req.Latitude, req.Description, req.Photos, uuid.MustParse(reporterID)); err != nil {
		utils.ServerError(c, "添加轨迹失败: "+err.Error())
		return
	}

	utils.Success(c, nil)
}

// AddTrackRequest 添加轨迹请求
type AddTrackRequest struct {
	TrackTime   string   `json:"track_time"`
	Location    string   `json:"location"`
	Longitude   float64  `json:"longitude"`
	Latitude    float64  `json:"latitude"`
	Description string   `json:"description"`
	Photos      []string `json:"photos"`
}

// GetTracks 获取轨迹列表
// @Summary 获取轨迹列表
// @Description 获取走失人员轨迹记录列表
// @Tags 走失人员管理
// @Produce json
// @Param id path string true "记录ID"
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /missing-persons/{id}/tracks [get]
func (h *MissingPersonHandler) GetTracks(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	tracks, err := h.mpService.GetTracks(c.Request.Context(), id)
	if err != nil {
		utils.ServerError(c, "获取轨迹失败")
		return
	}

	utils.Success(c, tracks)
}

// GetStatistics 获取统计
// @Summary 获取统计
// @Description 获取走失人员统计信息
// @Tags 走失人员管理
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /missing-persons/statistics [get]
func (h *MissingPersonHandler) GetStatistics(c *gin.Context) {
	orgID := uuid.Nil
	if oid := c.Query("org_id"); oid != "" {
		orgID, _ = uuid.Parse(oid)
	}

	stats, err := h.mpService.GetStatistics(c.Request.Context(), orgID)
	if err != nil {
		utils.ServerError(c, "获取统计失败")
		return
	}

	utils.Success(c, stats)
}
