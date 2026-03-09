// Package middleware 审计日志中间件
package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/gin-gonic/gin"
)

// AuditMiddleware 审计日志中间件
type AuditMiddleware struct {
	auditRepo repository.AuditLogRepository
}

// NewAuditMiddleware 创建审计日志中间件
func NewAuditMiddleware(auditRepo repository.AuditLogRepository) *AuditMiddleware {
	return &AuditMiddleware{auditRepo: auditRepo}
}

// AuditOptions 审计选项
type AuditOptions struct {
	Module      string             // 模块名称
	Action      string             // 操作名称
	Type        entity.AuditLogType // 操作类型
	IgnoreBody  bool               // 是否忽略请求体
	Async       bool               // 是否异步记录
}

// Audit 通用审计中间件
func (m *AuditMiddleware) Audit(opts AuditOptions) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果不是审计路径，跳过
		if isAuditPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		start := time.Now()
		
		// 创建审计日志
		auditLog := entity.NewAuditLog(opts.Module, opts.Action, opts.Type)
		auditLog.RequestMethod = c.Request.Method
		auditLog.RequestURL = c.Request.URL.String()
		auditLog.RequestIP = c.ClientIP()
		auditLog.UserAgent = c.Request.UserAgent()
		
		// 获取追踪ID
		if traceID, exists := c.Get("trace_id"); exists {
			if id, ok := traceID.(string); ok {
				auditLog.TraceID = id
			}
		}

		// 获取用户信息
		if userID := GetUserID(c); userID != "" {
			auditLog.UserID = userID
			auditLog.UserRole = string(GetUserRole(c))
		}

		// 读取请求体
		var requestBody string
		if !opts.IgnoreBody && c.Request.Body != nil && c.Request.Method != http.MethodGet {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			requestBody = string(bodyBytes)
			
			// 敏感信息脱敏
			auditLog.RequestBody = maskSensitiveData(requestBody)
		}

		// 执行请求
		c.Next()

		// 记录响应信息
		auditLog.ResponseCode = c.Writer.Status()
		auditLog.Duration = time.Since(start).Milliseconds()

		// 判断状态
		if c.Writer.Status() >= 400 {
			auditLog.MarkFailure(getErrorMessage(c))
		} else {
			auditLog.MarkSuccess()
		}

		// 生成描述
		auditLog.Description = generateDescription(auditLog, opts)

		// 保存日志
		if opts.Async {
			go m.saveAuditLog(auditLog)
		} else {
			m.saveAuditLog(auditLog)
		}
	}
}

// AutoAudit 自动审计（根据URL路径自动识别）
func (m *AuditMiddleware) AutoAudit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过审计路径
		if isAuditPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		start := time.Now()
		
		// 解析路径获取模块和动作
		module, action, logType := parsePath(c.Request.URL.Path, c.Request.Method)
		
		// 创建审计日志
		auditLog := entity.NewAuditLog(module, action, logType)
		auditLog.RequestMethod = c.Request.Method
		auditLog.RequestURL = c.Request.URL.String()
		auditLog.RequestIP = c.ClientIP()
		auditLog.UserAgent = c.Request.UserAgent()
		
		if traceID, exists := c.Get("trace_id"); exists {
			if id, ok := traceID.(string); ok {
				auditLog.TraceID = id
			}
		}

		// 获取用户信息
		if userID := GetUserID(c); userID != "" {
			auditLog.UserID = userID
			auditLog.UserRole = string(GetUserRole(c))
		}

		// 敏感操作记录请求体
		if isSensitiveOperation(module, action) && c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			auditLog.RequestBody = maskSensitiveData(string(bodyBytes))
		}

		// 执行请求
		c.Next()

		// 记录响应
		auditLog.ResponseCode = c.Writer.Status()
		auditLog.Duration = time.Since(start).Milliseconds()

		if c.Writer.Status() >= 400 {
			auditLog.MarkFailure(getErrorMessage(c))
		} else {
			auditLog.MarkSuccess()
		}

		auditLog.Description = generateDescription(auditLog, AuditOptions{
			Module: module,
			Action: action,
			Type:   logType,
		})

		// 异步保存
		go m.saveAuditLog(auditLog)
	}
}

// saveAuditLog 保存审计日志
func (m *AuditMiddleware) saveAuditLog(log *entity.AuditLog) {
	if m.auditRepo == nil {
		return
	}

	// 使用后台 context
	ctx := context.Background()
	if err := m.auditRepo.Create(ctx, log); err != nil {
		logger.Error("Failed to save audit log", logger.Err(err))
	}
}

// isAuditPath 是否是审计路径（避免循环记录）
func isAuditPath(path string) bool {
	auditPaths := []string{
		"/api/v1/audit",
		"/api/v1/health",
		"/api/v1/metrics",
	}
	
	for _, p := range auditPaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// parsePath 解析路径获取模块和动作
func parsePath(path, method string) (module, action string, logType entity.AuditLogType) {
	// 移除 /api/v1/ 前缀
	path = strings.TrimPrefix(path, "/api/v1/")
	parts := strings.Split(path, "/")
	
	if len(parts) == 0 {
		return "unknown", "unknown", entity.AuditLogTypeOther
	}

	// 第一个部分作为模块
	module = parts[0]
	
	// 根据HTTP方法判断动作
	switch method {
	case http.MethodGet:
		action = "查询"
		logType = entity.AuditLogTypeQuery
	case http.MethodPost:
		action = "创建"
		logType = entity.AuditLogTypeCreate
	case http.MethodPut, http.MethodPatch:
		action = "更新"
		logType = entity.AuditLogTypeUpdate
	case http.MethodDelete:
		action = "删除"
		logType = entity.AuditLogTypeDelete
	default:
		action = "其他"
		logType = entity.AuditLogTypeOther
	}

	// 特殊处理
	switch module {
	case "auth":
		if strings.Contains(path, "login") {
			action = "登录"
			logType = entity.AuditLogTypeLogin
		} else if strings.Contains(path, "logout") {
			action = "登出"
			logType = entity.AuditLogTypeLogout
		}
	case "upload":
		action = "上传"
		logType = entity.AuditLogTypeUpload
	}

	// 转换模块名称为中文
	moduleNames := map[string]string{
		"auth":            "认证",
		"users":           "用户管理",
		"organizations":   "组织管理",
		"missing-persons": "案件管理",
		"tasks":           "任务管理",
		"dialects":        "方言管理",
		"dashboard":       "仪表盘",
		"upload":          "文件管理",
	}
	
	if name, ok := moduleNames[module]; ok {
		module = name
	}

	return
}

// isSensitiveOperation 是否是敏感操作
func isSensitiveOperation(module, action string) bool {
	// 登录、删除、密码修改等敏感操作
	sensitiveActions := []string{"删除", "登录", "修改密码", "重置密码"}
	for _, sa := range sensitiveActions {
		if strings.Contains(action, sa) {
			return true
		}
	}
	
	// 敏感模块的写操作
	sensitiveModules := []string{"用户管理", "组织管理"}
	for _, sm := range sensitiveModules {
		if module == sm && (action == "创建" || action == "更新" || action == "删除") {
			return true
		}
	}
	
	return false
}

// maskSensitiveData 敏感数据脱敏
func maskSensitiveData(data string) string {
	if data == "" {
		return ""
	}

	// 密码脱敏
	passwordPattern := regexp.MustCompile(`"password"\s*:\s*"[^"]*"`)
	data = passwordPattern.ReplaceAllString(data, `"password":"***"`)

	// 身份证号脱敏
	idCardPattern := regexp.MustCompile(`(\d{6})\d{8}(\d{4})`)
	data = idCardPattern.ReplaceAllString(data, "${1}********${2}")

	// 手机号脱敏
	phonePattern := regexp.MustCompile(`(1\d{2})\d{4}(\d{4})`)
	data = phonePattern.ReplaceAllString(data, "${1}****${2}")

	// Token脱敏
	tokenPattern := regexp.MustCompile(`"(token|access_token|refresh_token)"\s*:\s*"[^"]{10,}"`)
	data = tokenPattern.ReplaceAllString(data, `"${1}":"***"`)

	return data
}

// getErrorMessage 获取错误信息
func getErrorMessage(c *gin.Context) string {
	if len(c.Errors) > 0 {
		return c.Errors.Last().Error()
	}
	return http.StatusText(c.Writer.Status())
}

// generateDescription 生成操作描述
func generateDescription(log *entity.AuditLog, opts AuditOptions) string {
	var desc strings.Builder
	
	if log.Username != "" {
		desc.WriteString(log.Username)
		desc.WriteString(" ")
	}
	
	desc.WriteString(opts.Action)
	
	if opts.Module != "" {
		desc.WriteString(" ")
		desc.WriteString(opts.Module)
	}
	
	if log.RequestURL != "" {
		desc.WriteString(" (")
		desc.WriteString(log.RequestURL)
		desc.WriteString(")")
	}
	
	return desc.String()
}
