package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/gin-gonic/gin"
)

// AuditMiddleware 审计中间件
type AuditMiddleware struct {
	auditService *service.AuditService
}

// NewAuditMiddleware 创建审计中间件
func NewAuditMiddleware(auditService *service.AuditService) *AuditMiddleware {
	return &AuditMiddleware{
		auditService: auditService,
	}
}

// Audit 审计中间件
func (m *AuditMiddleware) Audit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否需要审计
		if !m.auditService.ShouldAudit(c.Request.URL.Path) {
			c.Next()
			return
		}
		
		start := time.Now()
		
		// 读取请求体
		var requestBody []byte
		if c.Request.Body != nil && c.Request.ContentLength > 0 {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}
		
		// 包装响应写入器以捕获响应
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw
		
		// 继续处理请求
		c.Next()
		
		// 计算执行时间
		duration := time.Since(start).Milliseconds()
		
		// 获取用户信息
		userID, _ := c.Get("user_id")
		orgID, _ := c.Get("org_id")
		username, _ := c.Get("username")
		
		userIDStr := ""
		if id, ok := userID.(string); ok {
			userIDStr = id
		}
		
		orgIDStr := ""
		if id, ok := orgID.(string); ok {
			orgIDStr = id
		}
		
		usernameStr := ""
		if name, ok := username.(string); ok {
			usernameStr = name
		}
		
		// 获取追踪ID
		traceID, _ := c.Get("trace_id")
		traceIDStr := ""
		if id, ok := traceID.(string); ok {
			traceIDStr = id
		}
		
		// 解析操作类型和资源
		action := parseAction(c.Request.Method)
		resource := parseResource(c.Request.URL.Path)
		
		// 创建审计日志
		log := entity.NewAuditLog(userIDStr, orgIDStr, action, resource).
			SetRequestInfo(c.Request.Method, c.Request.URL.String(), c.ClientIP(), c.Request.UserAgent()).
			SetTraceID(traceIDStr).
			SetStatus(c.Writer.Status()).
			SetDuration(duration).
			SetUsername(usernameStr)
		
		// 如果是错误响应，记录错误信息
		if c.Writer.Status() >= 400 {
			if len(c.Errors) > 0 {
				log.SetError(c.Errors.String())
			} else if blw.body.Len() > 0 {
				// 限制错误信息长度
				errorMsg := blw.body.String()
				if len(errorMsg) > 500 {
					errorMsg = errorMsg[:500] + "..."
				}
				log.SetError(errorMsg)
			}
		}
		
		// 对于敏感操作，记录请求体
		if shouldLogRequestBody(c.Request.Method, resource) && len(requestBody) > 0 {
			var requestData map[string]interface{}
			if err := c.ShouldBindBodyWith(&requestData, nil); err == nil {
				log.NewValues = requestData
			}
		}
		
		// 异步记录审计日志
		go func() {
			if err := m.auditService.Log(c.Request.Context(), log); err != nil {
				logger.Error("Failed to log audit", logger.Err(err))
			}
		}()
	}
}

// bodyLogWriter 用于捕获响应体的写入器
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 写入响应体
func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// WriteString 写入字符串
func (w *bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// parseAction 解析HTTP方法为审计操作
func parseAction(method string) entity.AuditAction {
	switch method {
	case http.MethodPost:
		return entity.AuditActionCreate
	case http.MethodPut, http.MethodPatch:
		return entity.AuditActionUpdate
	case http.MethodDelete:
		return entity.AuditActionDelete
	case http.MethodGet:
		return entity.AuditActionQuery
	default:
		return entity.AuditActionOther
	}
}

// parseResource 从URL路径解析资源类型
func parseResource(path string) string {
	// 移除 /api/v1/ 前缀
	if len(path) > 8 && path[:8] == "/api/v1/" {
		path = path[8:]
	}
	
	// 获取第一个路径段
	for i, c := range path {
		if c == '/' {
			return path[:i]
		}
	}
	
	return path
}

// shouldLogRequestBody 检查是否应该记录请求体
func shouldLogRequestBody(method, resource string) bool {
	// 只记录写操作
	if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch {
		return false
	}
	
	// 排除敏感资源
	sensitiveResources := []string{"auth", "login", "register", "password", "token"}
	for _, sr := range sensitiveResources {
		if resource == sr {
			return false
		}
	}
	
	return true
}

// AuditAction 记录特定动作的审计日志
func (m *AuditMiddleware) AuditAction(action entity.AuditAction, resource, description string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户信息
		userID, _ := c.Get("user_id")
		orgID, _ := c.Get("org_id")
		username, _ := c.Get("username")
		
		userIDStr := ""
		if id, ok := userID.(string); ok {
			userIDStr = id
		}
		
		orgIDStr := ""
		if id, ok := orgID.(string); ok {
			orgIDStr = id
		}
		
		usernameStr := ""
		if name, ok := username.(string); ok {
			usernameStr = name
		}
		
		// 获取资源ID
		resourceID := c.Param("id")
		
		// 创建审计日志
		log := entity.NewAuditLog(userIDStr, orgIDStr, action, resource).
			SetResourceID(resourceID).
			SetDescription(description).
			SetRequestInfo(c.Request.Method, c.Request.URL.String(), c.ClientIP(), c.Request.UserAgent()).
			SetUsername(usernameStr)
		
		// 继续处理请求
		c.Next()
		
		// 更新状态
		log.SetStatus(c.Writer.Status())
		
		// 异步记录
		go func() {
			if err := m.auditService.Log(c.Request.Context(), log); err != nil {
				logger.Error("Failed to log audit action", logger.Err(err))
			}
		}()
	}
}
