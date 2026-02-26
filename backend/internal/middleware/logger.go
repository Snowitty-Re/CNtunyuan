package middleware

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// 读取请求体
		var requestBody []byte
		if method == "POST" || method == "PUT" || method == "PATCH" {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 包装响应写入器以捕获响应
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		// 计算耗时
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// 记录日志
		userID := GetUserID(c)
		if userID == "" {
			userID = "anonymous"
		}

		// 异步记录操作日志
		go func() {
			log := model.OperationLog{
				UserID:       uuid.MustParse(userID),
				Module:       getModule(path),
				Action:       method,
				Method:       method,
				Path:         path,
				IP:           clientIP,
				UserAgent:    userAgent,
				RequestBody:  string(requestBody),
				ResponseBody: blw.body.String(),
				StatusCode:   statusCode,
				Duration:     int(latency.Milliseconds()),
			}
			model.DB.Create(&log)
		}()

		// 控制台输出
		fullPath := path
		if query != "" {
			fullPath = path + "?" + query
		}
		fmt.Printf("[GIN] %s | %3d | %13v | %15s | %-7s %s\n",
			time.Now().Format("2006/01/02 - 15:04:05"),
			statusCode,
			latency,
			clientIP,
			method,
			fullPath,
		)
	}
}

// bodyLogWriter 用于捕获响应体
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// getModule 根据路径获取模块名
func getModule(path string) string {
	if len(path) < 5 {
		return "unknown"
	}
	// 去掉 /api/ 前缀
	if len(path) > 5 && path[:5] == "/api/" {
		path = path[5:]
	}
	// 获取第一个段
	for i, c := range path {
		if c == '/' {
			return path[:i]
		}
	}
	return path
}
