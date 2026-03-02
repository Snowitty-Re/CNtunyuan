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
		dashboard.GET("/stats", h.GetStats)
		dashboard.GET("/overview", h.GetOverview)
		dashboard.GET("/trend", h.GetTrend)
	}
}

// GetStats 获取仪表盘统计
func (h *DashboardHandler) GetStats(c *gin.Context) {
	stats, err := h.dashboardService.GetDashboardStats(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "failed to get dashboard stats")
		return
	}

	response.Success(c, stats)
}

// GetOverview 获取概览数据
func (h *DashboardHandler) GetOverview(c *gin.Context) {
	overview, err := h.dashboardService.GetOverview(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "failed to get overview")
		return
	}

	response.Success(c, overview)
}

// GetTrend 获取趋势数据
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
