package handler

import (
	"github.com/Snowitty-Re/CNtunyuan/internal/application/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/middleware"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"
	"github.com/gin-gonic/gin"
	"strconv"
)

// DashboardHandler 仪表盘处理器
type DashboardHandler struct {
	dashboardService *service.DashboardService
}

// NewDashboardHandler 创建仪表盘处理器
func NewDashboardHandler(dashboardService *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

// RegisterRoutes 注册路由
func (h *DashboardHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	dashboard := router.Group("/dashboard")
	dashboard.Use(authMiddleware.Required())
	{
		dashboard.GET("/stats", middleware.RequireManager(), h.GetStats)
		dashboard.GET("/overview", middleware.RequireManager(), h.GetOverview)
		dashboard.GET("/trend", middleware.RequireManager(), h.GetTrend)
	}
}

// GetStats 获取仪表盘统计
// @Summary 获取仪表盘统计
// @Description 获取仪表盘各项统计数据，包括用户、组织、走失人员、任务、方言、文件等统计信息
// @Tags 仪表盘
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=service.DashboardStats} "成功"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /dashboard/stats [get]
func (h *DashboardHandler) GetStats(c *gin.Context) {
	stats, err := h.dashboardService.GetDashboardStats(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "failed to get dashboard stats")
		return
	}

	response.Success(c, stats)
}

// GetOverview 获取概览数据
// @Summary 获取概览数据
// @Description 获取系统关键数据概览，包括总用户数、案件数、活跃案件数、已解决案件数等
// @Tags 仪表盘
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=map[string]interface{}} "成功"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /dashboard/overview [get]
func (h *DashboardHandler) GetOverview(c *gin.Context) {
	overview, err := h.dashboardService.GetOverview(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "failed to get overview")
		return
	}

	response.Success(c, overview)
}

// GetTrend 获取趋势数据
// @Summary 获取趋势数据
// @Description 获取指定天数内的趋势数据，包括新增案件、已解决案件、新增任务、已完成任务、新增用户等趋势
// @Tags 仪表盘
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param days query int false "天数范围，默认7天，最大30天" default(7) minimum(1) maximum(30)
// @Success 200 {object} response.Response{data=[]service.TrendData} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /dashboard/trend [get]
func (h *DashboardHandler) GetTrend(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	if days <= 0 {
		days = 7
	}
	if days > 30 {
		days = 30
	}

	trend, err := h.dashboardService.GetTrendData(c.Request.Context(), days)
	if err != nil {
		response.InternalServerError(c, "failed to get trend data")
		return
	}

	response.Success(c, trend)
}
