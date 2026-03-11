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

// AuditPaginatedResult 审计日志分页结果（Swagger用）
// @Description 审计日志分页查询结果
type AuditPaginatedResult struct {
	Items    []entity.AuditLog `json:"items"`     // 日志列表
	Total    int64             `json:"total"`      // 总数
	Page     int               `json:"page"`       // 当前页
	PageSize int               `json:"page_size"`  // 每页数量
}

// AuditUserActivityItem 用户活动项（Swagger用）
// @Description 用户活动统计项
type AuditUserActivityItem struct {
	Date  string `json:"date" example:"2024-01-01"`  // 日期
	Count int64  `json:"count" example:"15"`          // 操作次数
}

// AuditModuleStatItem 模块统计项（Swagger用）
// @Description 模块操作统计项
type AuditModuleStatItem struct {
	Module string `json:"module" example:"用户管理"` // 模块名称
	Count  int64  `json:"count" example:"42"`       // 操作次数
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
// @Summary 获取审计日志列表
// @Description 分页查询审计日志列表，支持按用户、模块、操作类型、状态、时间范围等条件筛选
// @Tags 审计日志
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码，默认1" default(1) minimum(1)
// @Param page_size query int false "每页数量，默认20" default(20) minimum(1) maximum(100)
// @Param user_id query string false "用户ID"
// @Param module query string false "模块名称，如：用户管理、案件管理等"
// @Param action query string false "操作名称，如：创建、更新、删除等"
// @Param type query string false "日志类型：login/logout/create/update/delete/query/export/upload/download/other"
// @Param status query string false "状态：success/failure"
// @Param start_time query string false "开始时间，格式：2006-01-02"
// @Param end_time query string false "结束时间，格式：2006-01-02"
// @Param keyword query string false "关键词搜索"
// @Param request_ip query string false "请求IP地址"
// @Success 200 {object} response.Response{data=AuditPaginatedResult} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /audit/logs [get]
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
// @Summary 根据ID获取审计日志
// @Description 根据审计日志ID获取单条日志详情
// @Tags 审计日志
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "审计日志ID"
// @Success 200 {object} response.Response{data=entity.AuditLog} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "日志不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /audit/logs/{id} [get]
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

// GetStats 获取审计日志统计
// @Summary 获取审计日志统计
// @Description 获取指定时间范围内的审计日志统计信息，包括总数、今日数量、登录次数、操作次数、错误次数等
// @Tags 审计日志
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param start_time query string false "开始时间，格式：2006-01-02"
// @Param end_time query string false "结束时间，格式：2006-01-02"
// @Success 200 {object} response.Response{data=entity.AuditLogStats} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /audit/stats [get]
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
// @Summary 获取用户活动统计
// @Description 获取指定用户在指定天数内的活动统计，按日期展示操作次数
// @Tags 审计日志
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param userId path string true "用户ID"
// @Param days query int false "天数范围，默认7天，最大90天" default(7) minimum(1) maximum(90)
// @Success 200 {object} response.Response{data=[]AuditUserActivityItem} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /audit/user-activity/{userId} [get]
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
// @Summary 获取模块统计
// @Description 获取指定时间范围内各模块的操作次数统计
// @Tags 审计日志
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param start_time query string false "开始时间，格式：2006-01-02"
// @Param end_time query string false "结束时间，格式：2006-01-02"
// @Success 200 {object} response.Response{data=[]AuditModuleStatItem} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /audit/module-stats [get]
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

// CleanupResponse 清理响应
type CleanupResponse struct {
	DeletedCount int64  `json:"deleted_count"`
	Before       string `json:"before"`
}

// Cleanup 清理旧日志
// @Summary 清理旧日志
// @Description 清理指定天数之前的审计日志，只有超级管理员可以执行此操作
// @Tags 审计日志
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body CleanupRequest true "清理请求参数"
// @Success 200 {object} response.Response{data=CleanupResponse} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "禁止访问，需要超级管理员权限"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /audit/cleanup [post]
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
