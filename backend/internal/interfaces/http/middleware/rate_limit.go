package middleware

import (
	"fmt"
	"time"

	domainService "github.com/Snowitty-Re/CNtunyuan/internal/domain/service"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"
	"github.com/gin-gonic/gin"
)

// DistributedRateLimitMiddleware 基于 Redis 的全局 IP 限流。
// 使用固定窗口计数：window 内单 IP 超过 limit 次请求将被拒绝。
func DistributedRateLimitMiddleware(cache domainService.Cache, limit int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 缓存不可用时不阻断请求，降级交由其他限流手段处理。
		if cache == nil {
			c.Next()
			return
		}

		slot := time.Now().Unix() / int64(window.Seconds())
		key := fmt.Sprintf("rate_limit:ip:%s:%d", c.ClientIP(), slot)
		count, err := cache.IncrWithTTL(c.Request.Context(), key, window)
		if err != nil {
			logger.Warn("Distributed rate limit degraded", logger.Err(err))
			c.Next()
			return
		}
		if count > limit {
			response.TooManyRequests(c, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}
