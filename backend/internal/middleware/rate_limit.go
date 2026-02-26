package middleware

import (
	"net/http"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

// RateLimiter 限流器
type RateLimiter struct {
	redis      *redis.Client
	limit      int
	window     time.Duration
	keyPrefix  string
}

// NewRateLimiter 创建限流器
func NewRateLimiter(redis *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		redis:     redis,
		limit:     limit,
		window:    window,
		keyPrefix: "rate_limit:",
	}
}

// Limit 限流中间件
func (r *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果Redis未配置，直接放行
		if r.redis == nil {
			c.Next()
			return
		}
		
		key := r.keyPrefix + c.ClientIP()
		ctx := context.Background()

		// 使用Redis计数
		pipe := r.redis.Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, r.window)
		_, err := pipe.Exec(ctx)
		
		if err != nil {
			c.Next()
			return
		}

		count := incr.Val()
		
		// 设置限流头信息
		c.Header("X-RateLimit-Limit", string(rune(r.limit)))
		c.Header("X-RateLimit-Remaining", string(rune(r.limit-int(count))))
		
		if count > int64(r.limit) {
			c.Header("X-RateLimit-Retry-After", string(rune(r.window/time.Second)))
			c.JSON(http.StatusTooManyRequests, utils.Response{
				Code:    utils.CodeServerError,
				Message: "请求过于频繁，请稍后再试",
				Data:    nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// IPBasedRateLimit 基于IP的限流
func IPBasedRateLimit(redis *redis.Client) gin.HandlerFunc {
	limiter := NewRateLimiter(redis, 100, time.Minute)
	return limiter.Limit()
}

// UserBasedRateLimit 基于用户的限流
func UserBasedRateLimit(redis *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == "" {
			c.Next()
			return
		}

		limiter := &RateLimiter{
			redis:     redis,
			limit:     1000,
			window:    time.Minute,
			keyPrefix: "rate_limit:user:",
		}

		key := limiter.keyPrefix + userID
		ctx := context.Background()

		pipe := redis.Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, limiter.window)
		_, err := pipe.Exec(ctx)
		
		if err != nil {
			c.Next()
			return
		}

		if incr.Val() > int64(limiter.limit) {
			c.JSON(http.StatusTooManyRequests, utils.Response{
				Code:    utils.CodeServerError,
				Message: "请求过于频繁，请稍后再试",
				Data:    nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
