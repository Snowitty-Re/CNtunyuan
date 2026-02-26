package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/Snowitty-Re/CNtunyuan/internal/utils"
	"github.com/gin-gonic/gin"
)

// Recovery 自定义恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录堆栈信息
				stack := debug.Stack()
				
				// 记录错误日志
				fmt.Printf("[PANIC] %v\n%s\n", err, stack)

				// 返回500错误
				c.JSON(http.StatusInternalServerError, utils.Response{
					Code:    utils.CodeServerError,
					Message: "服务器内部错误",
					Data:    nil,
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
