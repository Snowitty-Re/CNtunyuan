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

// AuditHandler 审计日志处理器
type AuditHandler struct {
	auditService *service.AuditService
}

// NewAuditHandler 创建审计日志处理器
func NewAuditHandler(auditService *service.AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// RegisterRoutes 注册路由
func (h *AuditHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	logs := router.Group("/audit-logs")
	logs.Use(authMiddleware.Required())
	{
		logs.GET("", h.List)
		logs.GET("/stats", h.GetStats)
		logs.GET("/my", h.GetMyLogs)
		logs.GET("/users/:user_id/activity", middleware.RequireManager(), h.GetUserActivity)
		logs.POST("/cleanup", middleware.RequireAdmin(), h.Cleanup)
		logs.GET("/retention", middleware.RequireAdmin(), h.GetRetentionStats)
	}
}

// List 获取审计日志列表
func (h *AuditHandler) List(c *gin.Context) {
	var req dto.AuditLogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	
	// 设置分页默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}
	
	// 普通用户只能查看本组织的日志
	userRole := middleware.GetUserRole(c)
	if userRole == entity.RoleVolunteer {
		req.OrgID = middleware.GetOrgID(c)
	}
	
	logs, err := h.auditService.List(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to list audit logs", logger.Err(err))
		response.InternalServerError(c, "failed to get audit logs")
		return
	}
	
	response.Success(c, logs)
}

// GetStats 获取审计统计
func (h *AuditHandler) GetStats(c *gin.Context) {
	// 获取查询天数，默认7天
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	if days <= 0 || days > 90 {
		days = 7
	}
	
	// 普通用户只能查看本组织的统计
	orgID := ""
	userRole := middleware.GetUserRole(c)
	if userRole == entity.RoleVolunteer || userRole == entity.RoleManager {
		orgID = middleware.GetOrgID(c)
	}
	
	stats, err := h.auditService.GetStats(c.Request.Context(), orgID, days)
	if err != nil {
		logger.Error("Failed to get audit stats", logger.Err(err))
		response.InternalServerError(c, "failed to get audit stats")
		return
	}
	
	response.Success(c, stats)
}

// GetMyLogs 获取当前用户的审计日志
func (h *AuditHandler) GetMyLogs(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	
	req := &dto.AuditLogListRequest{
		UserID:   userID,
		Page:     page,
		PageSize: pageSize,
	}
	
	logs, err := h.auditService.List(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to get my audit logs", logger.Err(err))
		response.InternalServerError(c, "failed to get audit logs")
		return
	}
	
	response.Success(c, logs)
}

// GetUserActivity 获取用户活动统计
func (h *AuditHandler) GetUserActivity(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		response.BadRequest(c, "user_id is required")
		return
	}
	
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	if days <= 0 || days > 90 {
		days = 7
	}
	
	activity, err := h.auditService.GetUserActivity(c.Request.Context(), userID, days)
	if err != nil {
		logger.Error("Failed to get user activity", logger.String("user_id", userID), logger.Err(err))
		response.InternalServerError(c, "failed to get user activity")
		return
	}
	
	response.Success(c, activity)
}

// Cleanup 清理过期日志
func (h *AuditHandler) Cleanup(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "365"))
	if days <= 0 {
		days = 365
	}
	if days < 30 {
		response.BadRequest(c, "retention days must be at least 30")
		return
	}
	
	deleted, err := h.auditService.CleanupOldLogs(c.Request.Context(), days)
	if err != nil {
		logger.Error("Failed to cleanup audit logs", logger.Err(err))
		response.InternalServerError(c, "failed to cleanup audit logs")
		return
	}
	
	response.Success(c, gin.H{
		"deleted_count": deleted,
		"retention_days": days,
	})
}

// GetRetentionStats 获取保留统计
func (h *AuditHandler) GetRetentionStats(c *gin.Context) {
	stats, err := h.auditService.GetRetentionStats(c.Request.Context())
	if err != nil {
		logger.Error("Failed to get retention stats", logger.Err(err))
		response.InternalServerError(c, "failed to get retention stats")
		return
	}
	
	response.Success(c, stats)
}
