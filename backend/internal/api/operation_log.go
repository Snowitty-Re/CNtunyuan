package api

import (
	"strconv"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/middleware"
	"github.com/Snowitty-Re/CNtunyuan/internal/model"
	"github.com/Snowitty-Re/CNtunyuan/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// OperationLogHandler 操作日志处理器
type OperationLogHandler struct {
	db *gorm.DB
}

// NewOperationLogHandler 创建操作日志处理器
func NewOperationLogHandler(db *gorm.DB) *OperationLogHandler {
	return &OperationLogHandler{db: db}
}

// List 获取操作日志列表
// @Summary 获取操作日志列表
// @Description 获取系统操作日志列表，仅超级管理员可访问
// @Tags 操作日志
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param user_id query string false "用户ID"
// @Param role query string false "角色"
// @Param org_id query string false "组织ID"
// @Param module query string false "模块"
// @Param action query string false "操作"
// @Param status query int false "状态码"
// @Param start_time query string false "开始时间(ISO8601)"
// @Param end_time query string false "结束时间(ISO8601)"
// @Param keyword query string false "关键词搜索"
// @Success 200 {object} utils.Response
// @Router /api/v1/operation-logs [get]
func (h *OperationLogHandler) List(c *gin.Context) {
	var query middleware.OperationLogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	// 解析时间
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			query.StartTime = &t
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			query.EndTime = &t
		}
	}

	logs, total, err := middleware.GetOperationLogs(h.db, &query)
	if err != nil {
		utils.ServerError(c, "查询失败")
		return
	}

	utils.Success(c, gin.H{
		"list":  logs,
		"total": total,
		"page":  query.Page,
		"size":  query.PageSize,
	})
}

// Cleanup 清理旧日志
// @Summary 清理旧日志
// @Description 清理指定天数之前的操作日志，仅超级管理员可访问
// @Tags 操作日志
// @Accept json
// @Produce json
// @Param days query int false "保留天数" default(30)
// @Success 200 {object} utils.Response
// @Router /api/v1/operation-logs/cleanup [delete]
func (h *OperationLogHandler) Cleanup(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 30
	}

	err = middleware.DeleteOldLogs(h.db, days)
	if err != nil {
		utils.ServerError(c, "清理失败")
		return
	}

	utils.Success(c, gin.H{
		"message": "已清理 " + daysStr + " 天前的日志",
	})
}

// GetUserStats 获取用户活动统计
// @Summary 获取用户活动统计
// @Description 获取指定用户的操作统计，仅超级管理员可访问
// @Tags 操作日志
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Param days query int false "统计天数" default(7)
// @Success 200 {object} utils.Response
// @Router /api/v1/operation-logs/stats/user/{user_id} [get]
func (h *OperationLogHandler) GetUserStats(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		utils.BadRequest(c, "用户ID不能为空")
		return
	}

	daysStr := c.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 7
	}

	stats, err := middleware.GetUserActivityStats(h.db, userID, days)
	if err != nil {
		utils.ServerError(c, "查询失败")
		return
	}

	utils.Success(c, stats)
}

// GetSummary 获取日志统计摘要
// @Summary 获取日志统计摘要
// @Description 获取操作日志的整体统计，仅超级管理员可访问
// @Tags 操作日志
// @Accept json
// @Produce json
// @Param days query int false "统计天数" default(7)
// @Success 200 {object} utils.Response
// @Router /api/v1/operation-logs/stats/summary [get]
func (h *OperationLogHandler) GetSummary(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 7
	}
	if days > 90 {
		days = 90
	}

	db := h.db

	var result struct {
		TotalRequests   int64 `json:"total_requests"`
		SuccessCount    int64 `json:"success_count"`
		ErrorCount      int64 `json:"error_count"`
		UniqueUsers     int64 `json:"unique_users"`
		AvgDuration     int   `json:"avg_duration"`
		TopModules      []struct {
			Module string `json:"module"`
			Count  int64  `json:"count"`
		} `json:"top_modules"`
		TopActions []struct {
			Action string `json:"action"`
			Count  int64  `json:"count"`
		} `json:"top_actions"`
		DailyStats []struct {
			Date        string `json:"date"`
			Count       int64  `json:"count"`
			ErrorCount  int64  `json:"error_count"`
			AvgDuration int    `json:"avg_duration"`
		} `json:"daily_stats"`
	}

	startTime := time.Now().AddDate(0, 0, -days)

	// 总请求数
	db.Model(&model.OperationLog{}).Where("created_at >= ?", startTime).Count(&result.TotalRequests)

	// 成功/失败数
	db.Model(&model.OperationLog{}).Where("created_at >= ? AND status_code < 400", startTime).Count(&result.SuccessCount)
	db.Model(&model.OperationLog{}).Where("created_at >= ? AND status_code >= 400", startTime).Count(&result.ErrorCount)

	// 独立用户数
	db.Model(&model.OperationLog{}).Where("created_at >= ?", startTime).Select("COUNT(DISTINCT user_id)").Scan(&result.UniqueUsers)

	// 平均响应时间
	db.Model(&model.OperationLog{}).Where("created_at >= ?", startTime).Select("COALESCE(AVG(duration), 0)").Scan(&result.AvgDuration)

	// 热门模块
	db.Model(&model.OperationLog{}).Where("created_at >= ?", startTime).
		Select("module, COUNT(*) as count").
		Group("module").Order("count DESC").Limit(5).
		Scan(&result.TopModules)

	// 热门操作
	db.Model(&model.OperationLog{}).Where("created_at >= ?", startTime).
		Select("action, COUNT(*) as count").
		Group("action").Order("count DESC").Limit(5).
		Scan(&result.TopActions)

	// 每日统计
	db.Model(&model.OperationLog{}).Where("created_at >= ?", startTime).
		Select("DATE(created_at) as date, COUNT(*) as count, COUNT(CASE WHEN status_code >= 400 THEN 1 END) as error_count, COALESCE(AVG(duration), 0) as avg_duration").
		Group("DATE(created_at)").Order("date ASC").
		Scan(&result.DailyStats)

	utils.Success(c, result)
}
