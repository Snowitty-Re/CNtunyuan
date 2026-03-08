// Package handler 审计日志处理器
package handler

import (
	"strconv"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/middleware"
	"github.com/Snowitty-Re/CNtunyuan/pkg/errors"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"
	"github.com/gin-gonic/gin"
)

// AuditHandler 审计日志处理器
type AuditHandler struct {
	auditService *service.AuditLogService
}

// NewAuditHandler 创建审计日志处理器
func NewAuditHandler(auditService *service.AuditLogService) *AuditHandler {
	return &AuditHandler{auditService: auditService}
}

// RegisterRoutes 注册路由
func (h *AuditHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	audit := router.Group("/audit")
	audit.Use(authMiddleware.Required())
	{
		audit.GET("/logs", h.List)
		audit.GET("/logs/:id", h.GetByID)
		audit.GET("/stats", h.GetStats)
		audit.GET("/user-activity/:userId", h.GetUserActivity)
		audit.GET("/module-stats", h.GetModuleStats)
		audit.POST("/cleanup", h.Cleanup)
	}
}

// ListRequest 列表查询请求
type ListRequest struct {
	Page      int    `form:"page" binding:"min=1"`
	PageSize  int    `form:"page_size" binding:"min=1,max=100"`
	UserID    string `form:"user_id"`
	Module    string `form:"module"`
	Action    string `form:"action"`
	Type      string `form:"type"`
	Status    string `form:"status"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
	Keyword   string `form:"keyword"`
	RequestIP string `form:"request_ip"`
}

// List 获取审计日志列表
func (h *AuditHandler) List(c *gin.Context) {
	var req ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.New(errors.CodeInvalidParam, "参数错误"))
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	// 解析时间
	query := entity.NewAuditLogQuery()
	query.Page = req.Page
	query.PageSize = req.PageSize
	query.UserID = req.UserID
	query.Module = req.Module
	query.Action = req.Action
	query.Keyword = req.Keyword
	query.RequestIP = req.RequestIP

	if req.Type != "" {
		query.Type = entity.AuditLogType(req.Type)
	}
	if req.Status != "" {
		query.Status = entity.AuditLogStatus(req.Status)
	}

	if req.StartTime != "" {
		if t, err := time.Parse("2006-01-02", req.StartTime); err == nil {
			query.StartTime = &t
		}
	}
	if req.EndTime != "" {
		if t, err := time.Parse("2006-01-02", req.EndTime); err == nil {
			endOfDay := t.Add(24*time.Hour - time.Second)
			query.EndTime = &endOfDay
		}
	}

	result, err := h.auditService.List(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to get audit logs", logger.Err(err))
		response.Error(c, errors.ErrInternal)
		return
	}

	response.Success(c, result)
}

// GetByID 根据ID获取审计日志
func (h *AuditHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, errors.New(errors.CodeInvalidParam, "ID不能为空"))
		return
	}

	log, err := h.auditService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, errors.ErrNotFound)
		return
	}

	response.Success(c, log)
}

// GetStats 获取统计信息
func (h *AuditHandler) GetStats(c *gin.Context) {
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	var startTime, endTime *time.Time

	if startTimeStr != "" {
		if t, err := time.Parse("2006-01-02", startTimeStr); err == nil {
			startTime = &t
		}
	}
	if endTimeStr != "" {
		if t, err := time.Parse("2006-01-02", endTimeStr); err == nil {
			endOfDay := t.Add(24*time.Hour - time.Second)
			endTime = &endOfDay
		}
	}

	stats, err := h.auditService.GetStats(c.Request.Context(), startTime, endTime)
	if err != nil {
		logger.Error("Failed to get audit stats", logger.Err(err))
		response.Error(c, errors.ErrInternal)
		return
	}

	response.Success(c, stats)
}

// GetUserActivity 获取用户活动统计
func (h *AuditHandler) GetUserActivity(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		response.Error(c, errors.New(errors.CodeInvalidParam, "用户ID不能为空"))
		return
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	if days <= 0 || days > 90 {
		days = 7
	}

	activity, err := h.auditService.GetUserActivity(c.Request.Context(), userID, days)
	if err != nil {
		logger.Error("Failed to get user activity", logger.Err(err))
		response.Error(c, errors.ErrInternal)
		return
	}

	response.Success(c, activity)
}

// GetModuleStats 获取模块统计
func (h *AuditHandler) GetModuleStats(c *gin.Context) {
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	var startTime, endTime *time.Time

	if startTimeStr != "" {
		if t, err := time.Parse("2006-01-02", startTimeStr); err == nil {
			startTime = &t
		}
	}
	if endTimeStr != "" {
		if t, err := time.Parse("2006-01-02", endTimeStr); err == nil {
			endOfDay := t.Add(24*time.Hour - time.Second)
			endTime = &endOfDay
		}
	}

	stats, err := h.auditService.GetModuleStats(c.Request.Context(), startTime, endTime)
	if err != nil {
		logger.Error("Failed to get module stats", logger.Err(err))
		response.Error(c, errors.ErrInternal)
		return
	}

	response.Success(c, stats)
}

// CleanupRequest 清理请求
type CleanupRequest struct {
	Days int `json:"days" binding:"required,min=1,max=365"`
}

// Cleanup 清理旧日志
func (h *AuditHandler) Cleanup(c *gin.Context) {
	// 只有超级管理员可以清理日志
	if !middleware.IsSuperAdmin(c) {
		response.Error(c, errors.ErrForbidden)
		return
	}

	var req CleanupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.CodeInvalidParam, "参数错误"))
		return
	}

	before := time.Now().AddDate(0, 0, -req.Days)
	count, err := h.auditService.CleanupOldLogs(c.Request.Context(), before)
	if err != nil {
		logger.Error("Failed to cleanup audit logs", logger.Err(err))
		response.Error(c, errors.ErrInternal)
		return
	}

	response.Success(c, gin.H{
		"deleted_count": count,
		"before":        before.Format("2006-01-02 15:04:05"),
	})
}
