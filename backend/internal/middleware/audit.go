package middleware

import (
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditLogger 审计日志中间件
type AuditLogger struct {
	DB *gorm.DB
}

// NewAuditLogger 创建审计日志中间件
func NewAuditLogger(db *gorm.DB) *AuditLogger {
	return &AuditLogger{DB: db}
}

// LogConfig 日志配置
type LogConfig struct {
	SkipPaths   []string // 跳过的路径
	MaxBodySize int      // 最大记录body大小
}

// DefaultLogConfig 默认配置
var DefaultLogConfig = LogConfig{
	SkipPaths: []string{
		"/api/v1/auth/wechat-login",
		"/api/v1/auth/admin-login",
		"/api/v1/auth/refresh",
		"/health",
		"/swagger",
	},
	MaxBodySize: 1024 * 10, // 10KB
}

// Logger 操作日志中间件
func (a *AuditLogger) Logger(config ...LogConfig) gin.HandlerFunc {
	cfg := DefaultLogConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *gin.Context) {
		// 检查是否需要跳过
		if shouldSkipPath(c.Request.URL.Path, cfg.SkipPaths) {
			c.Next()
			return
		}

		// 开始时间
		start := time.Now()

		// 读取请求body
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// 截断body
		bodyStr := ""
		if len(bodyBytes) > 0 {
			if len(bodyBytes) > cfg.MaxBodySize {
				bodyStr = string(bodyBytes[:cfg.MaxBodySize]) + "... [truncated]"
			} else {
				bodyStr = maskSensitiveFields(string(bodyBytes))
			}
		}

		// 使用自定义的ResponseWriter捕获响应
		blw := &auditResponseWriter{
			body:           bytes.NewBufferString(""),
			ResponseWriter: c.Writer,
		}
		c.Writer = blw

		// 执行请求
		c.Next()

		// 计算执行时间
		duration := time.Since(start)

		// 获取用户信息
		userID := GetUserID(c)
		role := GetRole(c)

		// 获取客户端信息
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// 解析模块和操作
		module, action := parseOperation(c.Request.URL.Path, c.Request.Method)

		// 获取响应状态
		statusCode := c.Writer.Status()

		// 记录日志
		go func() {
			log := &model.OperationLog{
				UserID:      parseUUID(userID),
				Username:    role,
				Module:      module,
				Action:      action,
				Method:      c.Request.Method,
				Path:        c.Request.URL.Path,
				IP:          clientIP,
				UserAgent:   userAgent,
				RequestBody: bodyStr,
				StatusCode:  statusCode,
				Duration:    int(duration.Milliseconds()),
				CreatedAt:   time.Now(),
			}

			// 如果响应错误，记录错误信息
			if statusCode >= 400 {
				errorMsg := blw.body.String()
				if len(errorMsg) > 2000 {
					errorMsg = errorMsg[:2000] + "..."
				}
				log.ErrorMsg = errorMsg
			}

			// 保存到数据库
			a.saveLog(log)
		}()
	}
}

// auditResponseWriter 响应写入器
type auditResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w auditResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// shouldSkipPath 检查路径是否应该跳过
func shouldSkipPath(path string, skipPaths []string) bool {
	for _, skip := range skipPaths {
		if strings.HasPrefix(path, skip) {
			return true
		}
	}
	return false
}

// maskSensitiveFields 脱敏敏感字段
func maskSensitiveFields(body string) string {
	sensitiveFields := []string{"password", "token", "secret", "key", "credential"}

	// 简单的字符串替换，实际项目中可能需要更完善的JSON解析处理
	result := body
	for _, field := range sensitiveFields {
		// 匹配 "field": "value"
		idx := strings.Index(strings.ToLower(result), "\""+field+"\"")
		if idx != -1 {
			// 找到值的位置
			valueStart := idx + len(field) + 2
			if valueStart < len(result) {
				// 找到下一个引号
				for valueStart < len(result) && (result[valueStart] == '"' || result[valueStart] == ' ' || result[valueStart] == ':') {
					valueStart++
				}
				valueEnd := valueStart
				for valueEnd < len(result) && result[valueEnd] != '"' {
					valueEnd++
				}
				if valueEnd < len(result) {
					result = result[:valueStart] + "***" + result[valueEnd:]
				}
			}
		}
	}
	return result
}

// parseOperation 解析操作模块和动作
func parseOperation(path, method string) (string, string) {
	// 去除/api/v1前缀
	path = strings.TrimPrefix(path, "/api/v1")

	// 分割路径
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return "unknown", "unknown"
	}

	module := parts[0]

	// 根据HTTP方法和路径确定操作
	var action string
	switch method {
	case "GET":
		if len(parts) > 1 && parts[1] != "" {
			action = "view"
		} else {
			action = "list"
		}
	case "POST":
		action = "create"
	case "PUT", "PATCH":
		action = "update"
	case "DELETE":
		action = "delete"
	default:
		action = "unknown"
	}

	return module, action
}

// saveLog 保存日志到数据库
func (a *AuditLogger) saveLog(log *model.OperationLog) {
	// 异步保存，不阻塞主流程
	if a.DB == nil {
		return
	}

	// 使用Create忽略错误，避免影响主流程
	a.DB.Create(log)
}

// GetOperationLogs 获取操作日志列表（管理员功能）
func GetOperationLogs(db *gorm.DB, query *OperationLogQuery) ([]model.OperationLog, int64, error) {
	var logs []model.OperationLog
	var total int64

	dbConn := db.Model(&model.OperationLog{})

	// 应用过滤条件
	if query.UserID != "" {
		dbConn = dbConn.Where("user_id = ?", query.UserID)
	}
	if query.Module != "" {
		dbConn = dbConn.Where("module = ?", query.Module)
	}
	if query.Action != "" {
		dbConn = dbConn.Where("action = ?", query.Action)
	}
	if query.Status > 0 {
		dbConn = dbConn.Where("status_code = ?", query.Status)
	}
	if query.StartTime != nil {
		dbConn = dbConn.Where("created_at >= ?", query.StartTime)
	}
	if query.EndTime != nil {
		dbConn = dbConn.Where("created_at <= ?", query.EndTime)
	}
	if query.Keyword != "" {
		dbConn = dbConn.Where("path ILIKE ? OR request_body ILIKE ?",
			"%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}

	// 统计总数
	dbConn.Count(&total)

	// 分页查询
	offset := (query.Page - 1) * query.PageSize
	err := dbConn.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&logs).Error

	return logs, total, err
}

// OperationLogQuery 操作日志查询参数
type OperationLogQuery struct {
	Page      int        `form:"page" json:"page"`
	PageSize  int        `form:"page_size" json:"page_size"`
	UserID    string     `form:"user_id" json:"user_id"`
	Role      string     `form:"role" json:"role"`
	OrgID     string     `form:"org_id" json:"org_id"`
	Module    string     `form:"module" json:"module"`
	Action    string     `form:"action" json:"action"`
	Status    int        `form:"status" json:"status"`
	StartTime *time.Time `form:"start_time" json:"start_time"`
	EndTime   *time.Time `form:"end_time" json:"end_time"`
	Keyword   string     `form:"keyword" json:"keyword"`
}

// DeleteOldLogs 删除旧日志
func DeleteOldLogs(db *gorm.DB, days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	return db.Where("created_at < ?", cutoff).Delete(&model.OperationLog{}).Error
}

// parseUUID 将字符串解析为UUID
func parseUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}

// GetUserActivityStats 获取用户活动统计
func GetUserActivityStats(db *gorm.DB, userID string, days int) (*UserActivityStats, error) {
	startTime := time.Now().AddDate(0, 0, -days)

	var stats UserActivityStats
	stats.UserID = userID
	stats.Days = days

	userUUID := parseUUID(userID)
	dbConn := db.Model(&model.OperationLog{}).Where("user_id = ? AND created_at >= ?", userUUID, startTime)

	// 总操作次数
	var totalCount int64
	dbConn.Count(&totalCount)
	stats.TotalOperations = int(totalCount)

	// 按模块统计
	type ModuleStat struct {
		Module string
		Count  int64
	}
	var moduleStats []ModuleStat
	dbConn.Select("module, COUNT(*) as count").Group("module").Scan(&moduleStats)
	stats.ModuleStats = make(map[string]int)
	for _, ms := range moduleStats {
		stats.ModuleStats[ms.Module] = int(ms.Count)
	}

	// 按操作类型统计
	type ActionStat struct {
		Action string
		Count  int64
	}
	var actionStats []ActionStat
	dbConn.Select("action, COUNT(*) as count").Group("action").Scan(&actionStats)
	stats.ActionStats = make(map[string]int)
	for _, as := range actionStats {
		stats.ActionStats[as.Action] = int(as.Count)
	}

	return &stats, nil
}

// UserActivityStats 用户活动统计
type UserActivityStats struct {
	UserID          string         `json:"user_id"`
	Days            int            `json:"days"`
	TotalOperations int            `json:"total_operations"`
	ModuleStats     map[string]int `json:"module_stats"`
	ActionStats     map[string]int `json:"action_stats"`
}
