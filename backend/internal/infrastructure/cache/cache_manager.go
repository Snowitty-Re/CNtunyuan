package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/pkg/errors"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// CacheManager 缓存管理器接口
type CacheManager interface {
	// Get 获取缓存
	Get(ctx context.Context, key string, dest interface{}) error
	
	// Set 设置缓存
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	
	// SetNX 仅当key不存在时才设置（用于分布式锁等场景）
	SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error)
	
	// Delete 删除缓存
	Delete(ctx context.Context, key string) error
	
	// DeleteByPattern 根据模式删除缓存
	DeleteByPattern(ctx context.Context, pattern string) error
	
	// Exists 检查key是否存在
	Exists(ctx context.Context, key string) (bool, error)
	
	// TTL 获取key的剩余生存时间
	TTL(ctx context.Context, key string) (time.Duration, error)
	
	// Expire 设置key的过期时间
	Expire(ctx context.Context, key string, ttl time.Duration) error
	
	// Incr 原子递增
	Incr(ctx context.Context, key string) (int64, error)
	
	// Decr 原子递减
	Decr(ctx context.Context, key string) (int64, error)
	
	// GetSet 获取旧值并设置新值
	GetSet(ctx context.Context, key string, value interface{}, ttl time.Duration) (interface{}, error)
	
	// Close 关闭缓存连接
	Close() error
	
	// Ping 检查连接
	Ping(ctx context.Context) error
}

// CacheConfig 缓存配置
type CacheConfig struct {
	// 默认TTL
	DefaultTTL time.Duration
	
	// 空值缓存TTL（用于防止缓存穿透）
	NullTTL time.Duration
	
	// 是否启用缓存
	Enabled bool
	
	// 最大key长度
	MaxKeyLength int
	
	// 最大value大小（字节）
	MaxValueSize int
	
	// 前缀
	KeyPrefix string
}

// DefaultCacheConfig 默认缓存配置
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		DefaultTTL:   5 * time.Minute,
		NullTTL:      1 * time.Minute,
		Enabled:      true,
		MaxKeyLength: 250,
		MaxValueSize: 1024 * 1024, // 1MB
		KeyPrefix:    "cntuanyuan:",
	}
}

// RedisCacheManager Redis缓存管理器
type RedisCacheManager struct {
	client *redis.Client
	config *CacheConfig
}

// NewRedisCacheManager 创建Redis缓存管理器
func NewRedisCacheManager(client *redis.Client, config *CacheConfig) CacheManager {
	if config == nil {
		config = DefaultCacheConfig()
	}
	return &RedisCacheManager{
		client: client,
		config: config,
	}
}

// makeKey 生成带前缀的key
func (c *RedisCacheManager) makeKey(key string) string {
	if c.config.KeyPrefix != "" {
		return c.config.KeyPrefix + key
	}
	return key
}

// serialize 序列化值
func (c *RedisCacheManager) serialize(value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	default:
		return json.Marshal(value)
	}
}

// deserialize 反序列化值
func (c *RedisCacheManager) deserialize(data []byte, dest interface{}) error {
	if dest == nil {
		return nil
	}
	
	switch d := dest.(type) {
	case *string:
		*d = string(data)
		return nil
	case *[]byte:
		*d = data
		return nil
	default:
		return json.Unmarshal(data, dest)
	}
}

// Get 获取缓存
func (c *RedisCacheManager) Get(ctx context.Context, key string, dest interface{}) error {
	if !c.config.Enabled {
		return errors.ErrCacheMiss
	}
	
	fullKey := c.makeKey(key)
	
	val, err := c.client.Get(ctx, fullKey).Result()
	if err == redis.Nil {
		return errors.ErrCacheMiss
	}
	if err != nil {
		return errors.Wrap(err, errors.CodeCacheError, "cache get failed")
	}
	
	// 检查是否是空值标记（缓存穿透防护）
	if val == "__NULL__" {
		return errors.ErrCacheMiss
	}
	
	return c.deserialize([]byte(val), dest)
}

// Set 设置缓存
func (c *RedisCacheManager) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !c.config.Enabled {
		return nil
	}
	
	if ttl <= 0 {
		ttl = c.config.DefaultTTL
	}
	
	fullKey := c.makeKey(key)
	
	data, err := c.serialize(value)
	if err != nil {
		return errors.Wrap(err, errors.CodeCacheError, "cache serialize failed")
	}
	
	// 检查value大小
	if len(data) > c.config.MaxValueSize {
		logger.Warn("Cache value too large, skipping",
			logger.String("key", key),
			logger.Int("size", len(data)),
			logger.Int("max_size", c.config.MaxValueSize),
		)
		return nil
	}
	
	return c.client.Set(ctx, fullKey, data, ttl).Err()
}

// SetNX 仅当key不存在时才设置
func (c *RedisCacheManager) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	if !c.config.Enabled {
		return false, nil
	}
	
	if ttl <= 0 {
		ttl = c.config.DefaultTTL
	}
	
	fullKey := c.makeKey(key)
	
	data, err := c.serialize(value)
	if err != nil {
		return false, errors.Wrap(err, errors.CodeCacheError, "cache serialize failed")
	}
	
	return c.client.SetNX(ctx, fullKey, data, ttl).Result()
}

// Delete 删除缓存
func (c *RedisCacheManager) Delete(ctx context.Context, key string) error {
	if !c.config.Enabled {
		return nil
	}
	
	fullKey := c.makeKey(key)
	return c.client.Del(ctx, fullKey).Err()
}

// DeleteByPattern 根据模式删除缓存
func (c *RedisCacheManager) DeleteByPattern(ctx context.Context, pattern string) error {
	if !c.config.Enabled {
		return nil
	}
	
	fullPattern := c.makeKey(pattern)
	
	// 使用SCAN避免阻塞Redis
	var cursor uint64
	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, fullPattern, 100).Result()
		if err != nil {
			return errors.Wrap(err, errors.CodeCacheError, "cache scan failed")
		}
		
		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				logger.Error("Failed to delete cache keys", logger.Err(err))
			}
		}
		
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	
	return nil
}

// Exists 检查key是否存在
func (c *RedisCacheManager) Exists(ctx context.Context, key string) (bool, error) {
	if !c.config.Enabled {
		return false, nil
	}
	
	fullKey := c.makeKey(key)
	n, err := c.client.Exists(ctx, fullKey).Result()
	return n > 0, err
}

// TTL 获取key的剩余生存时间
func (c *RedisCacheManager) TTL(ctx context.Context, key string) (time.Duration, error) {
	fullKey := c.makeKey(key)
	return c.client.TTL(ctx, fullKey).Result()
}

// Expire 设置key的过期时间
func (c *RedisCacheManager) Expire(ctx context.Context, key string, ttl time.Duration) error {
	fullKey := c.makeKey(key)
	return c.client.Expire(ctx, fullKey, ttl).Err()
}

// Incr 原子递增
func (c *RedisCacheManager) Incr(ctx context.Context, key string) (int64, error) {
	if !c.config.Enabled {
		return 0, nil
	}
	
	fullKey := c.makeKey(key)
	return c.client.Incr(ctx, fullKey).Result()
}

// Decr 原子递减
func (c *RedisCacheManager) Decr(ctx context.Context, key string) (int64, error) {
	if !c.config.Enabled {
		return 0, nil
	}
	
	fullKey := c.makeKey(key)
	return c.client.Decr(ctx, fullKey).Result()
}

// GetSet 获取旧值并设置新值
func (c *RedisCacheManager) GetSet(ctx context.Context, key string, value interface{}, ttl time.Duration) (interface{}, error) {
	if !c.config.Enabled {
		return nil, errors.ErrCacheMiss
	}
	
	fullKey := c.makeKey(key)
	
	data, err := c.serialize(value)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeCacheError, "cache serialize failed")
	}
	
	oldVal, err := c.client.GetSet(ctx, fullKey, data).Result()
	if err == redis.Nil {
		// 设置过期时间
		if ttl > 0 {
			c.client.Expire(ctx, fullKey, ttl)
		}
		return nil, errors.ErrCacheMiss
	}
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeCacheError, "cache getset failed")
	}
	
	// 设置过期时间
	if ttl > 0 {
		c.client.Expire(ctx, fullKey, ttl)
	}
	
	return oldVal, nil
}

// Close 关闭缓存连接
func (c *RedisCacheManager) Close() error {
	return c.client.Close()
}

// Ping 检查连接
func (c *RedisCacheManager) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// CacheAside 旁路缓存模式
type CacheAside struct {
	cache CacheManager
}

// NewCacheAside 创建旁路缓存
func NewCacheAside(cache CacheManager) *CacheAside {
	return &CacheAside{cache: cache}
}

// GetOrSet 获取或设置缓存（支持缓存穿透防护）
func (ca *CacheAside) GetOrSet(
	ctx context.Context,
	key string,
	dest interface{},
	ttl time.Duration,
	getter func() (interface{}, error),
) error {
	// 1. 尝试从缓存获取
	err := ca.cache.Get(ctx, key, dest)
	if err == nil {
		return nil
	}
	
	// 2. 缓存未命中，从数据源获取
	data, err := getter()
	if err != nil {
		return err
	}
	
	// 3. 写入缓存
	if data != nil {
		if err := ca.cache.Set(ctx, key, data, ttl); err != nil {
			logger.Warn("Failed to set cache", logger.String("key", key), logger.Err(err))
		}
	}
	
	// 4. 设置到目标
	if dest != nil && data != nil {
		// 使用json序列化再反序列化来复制数据
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		return json.Unmarshal(jsonData, dest)
	}
	
	return nil
}

// GetOrSetWithNullProtection 带空值防护的获取或设置（防止缓存穿透）
func (ca *CacheAside) GetOrSetWithNullProtection(
	ctx context.Context,
	key string,
	dest interface{},
	ttl time.Duration,
	nullTTL time.Duration,
	getter func() (interface{}, error),
) error {
	// 1. 尝试从缓存获取
	err := ca.cache.Get(ctx, key, dest)
	if err == nil {
		return nil
	}
	
	// 2. 检查是否是空值标记（已经通过Get方法检查了__NULL__）
	
	// 3. 使用分布式锁防止缓存击穿
	lockKey := key + ":lock"
	locked, err := ca.cache.SetNX(ctx, lockKey, "1", 10*time.Second)
	if err != nil || !locked {
		// 获取锁失败，等待后重试
		time.Sleep(100 * time.Millisecond)
		return ca.cache.Get(ctx, key, dest)
	}
	
	// 4. 释放锁
	defer ca.cache.Delete(ctx, lockKey)
	
	// 5. 从数据源获取
	result, err := getter()
	if err != nil {
		// 记录错误但继续返回
		return err
	}
	
	// 6. 缓存空值防止穿透
	if result == nil {
		ca.cache.Set(ctx, key, "__NULL__", nullTTL)
		return errors.ErrNotFound
	}
	
	// 7. 写入缓存
	if err := ca.cache.Set(ctx, key, result, ttl); err != nil {
		logger.Warn("Failed to set cache", logger.String("key", key), logger.Err(err))
	}
	
	// 8. 设置到目标
	if dest != nil {
		jsonData, err := json.Marshal(result)
		if err != nil {
			return err
		}
		return json.Unmarshal(jsonData, dest)
	}
	
	return nil
}

// DeleteCache 删除缓存
func (ca *CacheAside) DeleteCache(ctx context.Context, key string) error {
	return ca.cache.Delete(ctx, key)
}

// DeleteCacheByPattern 根据模式删除缓存
func (ca *CacheAside) DeleteCacheByPattern(ctx context.Context, pattern string) error {
	return ca.cache.DeleteByPattern(ctx, pattern)
}

// CacheKeyBuilder 缓存key构建器
type CacheKeyBuilder struct {
	prefix string
}

// NewCacheKeyBuilder 创建缓存key构建器
func NewCacheKeyBuilder(prefix string) *CacheKeyBuilder {
	return &CacheKeyBuilder{prefix: prefix}
}

// Build 构建缓存key
func (b *CacheKeyBuilder) Build(parts ...string) string {
	key := b.prefix
	for _, part := range parts {
		if key != "" && !endsWith(key, ":") && !startsWith(part, ":") {
			key += ":"
		}
		key += part
	}
	return key
}

// BuildWithID 构建带ID的缓存key
func (b *CacheKeyBuilder) BuildWithID(resource string, id string) string {
	return b.Build(resource, id)
}

// BuildList 构建列表缓存key
func (b *CacheKeyBuilder) BuildList(resource string, params map[string]string) string {
	key := b.Build(resource, "list")
	if len(params) > 0 {
		key += ":"
		for k, v := range params {
			key += fmt.Sprintf("%s=%s:", k, v)
		}
		key = key[:len(key)-1] // 移除最后的冒号
	}
	return key
}

// endsWith 检查字符串是否以指定后缀结尾
func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

// startsWith 检查字符串是否以指定前缀开头
func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
