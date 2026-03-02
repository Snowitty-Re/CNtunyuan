package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// Cache 缓存接口
type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, key string) (bool, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	Close() error
}

// RedisCache Redis 缓存实现
type RedisCache struct {
	client *redis.Client
	prefix string
}

// NewRedis 创建 Redis 缓存
func NewRedis(cfg *config.RedisConfig) (*RedisCache, error) {
	if cfg.Host == "" {
		logger.Info("Redis not configured, cache will be disabled")
		return nil, nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Warn("Redis connection failed, cache will be disabled", logger.Err(err))
		return nil, err
	}

	logger.Info("Redis connected successfully",
		logger.String("addr", cfg.Host),
		logger.Int("db", cfg.DB),
	)

	return &RedisCache{
		client: client,
		prefix: "cntuanyuan:",
	}, nil
}

// Get 获取缓存
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("cache not available")
	}

	data, err := c.client.Get(ctx, c.prefix+key).Bytes()
	if err == redis.Nil {
		return fmt.Errorf("key not found")
	}
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

// Set 设置缓存
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if c == nil || c.client == nil {
		return nil // 缓存不可用时不报错
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, c.prefix+key, data, expiration).Err()
}

// Delete 删除缓存
func (c *RedisCache) Delete(ctx context.Context, keys ...string) error {
	if c == nil || c.client == nil {
		return nil
	}

	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = c.prefix + key
	}

	return c.client.Del(ctx, prefixedKeys...).Err()
}

// Exists 检查 key 是否存在
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	if c == nil || c.client == nil {
		return false, nil
	}

	n, err := c.client.Exists(ctx, c.prefix+key).Result()
	return n > 0, err
}

// TTL 获取 key 的过期时间
func (c *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	if c == nil || c.client == nil {
		return 0, nil
	}

	return c.client.TTL(ctx, c.prefix+key).Result()
}

// Expire 设置过期时间
func (c *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}

	return c.client.Expire(ctx, c.prefix+key, expiration).Err()
}

// Close 关闭连接
func (c *RedisCache) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}

// GetClient 获取原始 Redis 客户端
func (c *RedisCache) GetClient() *redis.Client {
	if c == nil {
		return nil
	}
	return c.client
}

// CacheKey 生成缓存 key
func CacheKey(parts ...string) string {
	key := ""
	for i, part := range parts {
		if i > 0 {
			key += ":"
		}
		key += part
	}
	return key
}

// DefaultExpiration 默认过期时间
const DefaultExpiration = 5 * time.Minute

// LongExpiration 长过期时间
const LongExpiration = 1 * time.Hour

// ShortExpiration 短过期时间
const ShortExpiration = 1 * time.Minute
