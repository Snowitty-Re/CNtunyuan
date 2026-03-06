// Package middleware 提供 HTTP 中间件
package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/pkg/errors"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/Snowitty-Re/CNtunyuan/pkg/metrics"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// TraceIDKey 追踪 ID 的上下文键
type TraceIDKey struct{}

// GetTraceID 从上下文获取追踪 ID
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey{}).(string); ok {
		return traceID
	}
	return ""
}

// TraceIDMiddleware 追踪 ID 中间件
func TraceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取或生成新的 trace_id
		traceID := c.GetHeader("X-Request-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}
		
		// 设置到上下文
		ctx := context.WithValue(c.Request.Context(), TraceIDKey{}, traceID)
		c.Request = c.Request.WithContext(ctx)
		
		// 设置到响应头
		c.Header("X-Request-ID", traceID)
		c.Set("trace_id", traceID)
		
		c.Next()
	}
}

// LoggingMiddleware 结构化日志中间件
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		
		// 读取请求体（用于日志记录）
		var bodyBytes []byte
		if c.Request.Body != nil && c.Request.Method != http.MethodGet {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
		
		c.Next()
		
		// 计算延迟
		latency := time.Since(start)
		
		// 获取状态码
		status := c.Writer.Status()
		
		// 获取错误
		var err error
		if len(c.Errors) > 0 {
			err = c.Errors.Last()
		}
		
		// 获取 trace_id
		traceID, _ := c.Get("trace_id")
		
		// 构建日志字段
		fields := []zap.Field{
			zap.String("trace_id", traceID.(string)),
			zap.String("client_ip", c.ClientIP()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", raw),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("referer", c.Request.Referer()),
		}
		
		// 添加请求体（非敏感接口）
		if len(bodyBytes) > 0 && len(bodyBytes) < 1024 {
			if !isSensitivePath(path) {
				fields = append(fields, zap.String("body", string(bodyBytes)))
			}
		}
		
		// 添加错误信息
		if err != nil {
			fields = append(fields, zap.Error(err))
		}
		
		// 根据状态码记录日志级别
		if status >= 500 {
			logger.Error("HTTP Request", fields...)
		} else if status >= 400 {
			logger.Warn("HTTP Request", fields...)
		} else {
			logger.Info("HTTP Request", fields...)
		}

		// 记录 Prometheus 指标
		metrics.RecordHTTPRequest(
			c.Request.Method,
			path,
			strconv.Itoa(status),
			latency.Seconds(),
			c.Request.ContentLength,
			int64(c.Writer.Size()),
		)
	}
}

// isSensitivePath 检查是否是敏感路径
func isSensitivePath(path string) bool {
	sensitivePaths := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/refresh",
	}
	for _, p := range sensitivePaths {
		if strings.Contains(path, p) {
			return true
		}
	}
	return false
}

// RecoveryMiddleware 恢复中间件（捕获 panic）
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// 获取 trace_id
				traceID, _ := c.Get("trace_id")
				
				// 记录错误日志
				logger.Error("Panic recovered",
					zap.String("trace_id", traceID.(string)),
					zap.Any("panic", r),
					zap.Stack("stack"),
				)
				
				// 返回 500 错误
				response.InternalServerError(c, fmt.Sprintf("internal server error (trace_id: %s)", traceID))
				c.Abort()
			}
		}()
		c.Next()
	}
}

// SecurityHeadersMiddleware 安全响应头中间件
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// XSS 防护
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// CSP
		c.Header("Content-Security-Policy", "default-src 'self'")
		
		// HSTS (仅在 HTTPS 环境下启用)
		// c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		
		//  Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Permissions Policy
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		
		c.Next()
	}
}

// RateLimiter 限流器
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	limit    rate.Limit
	burst    int
}

// NewRateLimiter 创建限流器
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		limit:    rate.Limit(rps),
		burst:    burst,
	}
}

// getLimiter 获取或创建限流器
func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	if limiter, ok := rl.limiters[key]; ok {
		return limiter
	}
	limiter := rate.NewLimiter(rl.limit, rl.burst)
	rl.limiters[key] = limiter
	return limiter
}

// RateLimitMiddleware IP 限流中间件
func RateLimitMiddleware(rps float64, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(rps, burst)
	
	return func(c *gin.Context) {
		key := c.ClientIP()
		if !limiter.getLimiter(key).Allow() {
			response.TooManyRequests(c, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}
		c.Next()
	}
}

// UserRateLimitMiddleware 用户限流中间件（需要登录）
func UserRateLimitMiddleware(rps float64, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(rps, burst)
	
	return func(c *gin.Context) {
		// 获取用户 ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}
		
		key := fmt.Sprintf("user:%v", userID)
		if !limiter.getLimiter(key).Allow() {
			response.TooManyRequests(c, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}
		c.Next()
	}
}

// CORSMiddleware CORS 中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// 允许的源（生产环境应该配置具体域名）
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:5173",
		}
		
		allowed := false
		for _, o := range allowedOrigins {
			if origin == o {
				allowed = true
				break
			}
		}
		
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}

// ErrorHandlerMiddleware 统一错误处理中间件
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		// 处理错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			response.Error(c, err)
		}
	}
}

// RequestSizeMiddleware 请求大小限制中间件
func RequestSizeMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			response.ErrorCodeWithMessage(c, errors.CodeInvalidParam, 
				fmt.Sprintf("请求体过大，最大允许 %d MB", maxSize/1024/1024))
			c.Abort()
			return
		}
		c.Next()
	}
}
