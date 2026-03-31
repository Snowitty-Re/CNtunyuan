// Package service 健康检查服务
package service

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"time"

	domainService "github.com/Snowitty-Re/CNtunyuan/internal/domain/service"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"gorm.io/gorm"
)

// HealthStatus 健康状态
type HealthStatus string

const (
	HealthStatusUP   HealthStatus = "UP"
	HealthStatusDOWN HealthStatus = "DOWN"
)

// HealthCheck 健康检查结果
type HealthCheck struct {
	Status    HealthStatus `json:"status"`
	Message   string       `json:"message,omitempty"`
	Response  int64        `json:"response_ms"`
	Timestamp time.Time    `json:"timestamp"`
}

// HealthStatus 整体健康状态
type HealthStatusResult struct {
	Status    HealthStatus           `json:"status"`
	Version   string                 `json:"version"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]HealthCheck `json:"checks"`
}

// HealthService 健康检查服务
type HealthService struct {
	db    *gorm.DB
	cache domainService.Cache
}

// NewHealthService 创建健康检查服务
func NewHealthService(db *gorm.DB, cache domainService.Cache) *HealthService {
	return &HealthService{
		db:    db,
		cache: cache,
	}
}

// CheckHealth 检查健康状态
func (s *HealthService) CheckHealth(ctx context.Context) *HealthStatusResult {
	result := &HealthStatusResult{
		Status:    HealthStatusUP,
		Version:   "2.0.0",
		Timestamp: time.Now(),
		Checks:    make(map[string]HealthCheck),
	}

	// 检查数据库
	dbCheck := s.checkDatabase(ctx)
	result.Checks["database"] = dbCheck
	if dbCheck.Status != HealthStatusUP {
		result.Status = HealthStatusDOWN
	}

	// 检查缓存
	cacheCheck := s.checkCache(ctx)
	result.Checks["cache"] = cacheCheck
	if cacheCheck.Status != HealthStatusUP && cacheCheck.Status != "" {
		// 缓存不是关键组件，不标记为整体DOWN
		logger.Warn("Cache health check failed", logger.String("status", string(cacheCheck.Status)))
	}

	// 检查系统资源
	sysCheck := s.checkSystemResources(ctx)
	result.Checks["system"] = sysCheck

	return result
}

// checkDatabase 检查数据库连接
func (s *HealthService) checkDatabase(ctx context.Context) HealthCheck {
	start := time.Now()
	check := HealthCheck{
		Timestamp: time.Now(),
	}

	if s.db == nil {
		check.Status = HealthStatusDOWN
		check.Message = "database not configured"
		check.Response = time.Since(start).Milliseconds()
		return check
	}

	sqlDB, err := s.db.DB()
	if err != nil {
		check.Status = HealthStatusDOWN
		check.Message = fmt.Sprintf("failed to get sql.DB: %v", err)
		check.Response = time.Since(start).Milliseconds()
		return check
	}

	// 使用超时上下文检查连接
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		check.Status = HealthStatusDOWN
		check.Message = fmt.Sprintf("database ping failed: %v", err)
		check.Response = time.Since(start).Milliseconds()
		return check
	}

	// 检查连接池状态
	stats := sqlDB.Stats()
	if float64(stats.InUse) > float64(stats.MaxOpenConnections)*0.9 {
		check.Status = HealthStatusDOWN
		check.Message = fmt.Sprintf("database connections near limit: %d/%d", stats.InUse, stats.MaxOpenConnections)
		check.Response = time.Since(start).Milliseconds()
		return check
	}

	check.Status = HealthStatusUP
	check.Message = fmt.Sprintf("connections: %d/%d, idle: %d", stats.OpenConnections, stats.MaxOpenConnections, stats.Idle)
	check.Response = time.Since(start).Milliseconds()
	return check
}

// checkCache 检查缓存连接
func (s *HealthService) checkCache(ctx context.Context) HealthCheck {
	start := time.Now()
	check := HealthCheck{
		Timestamp: time.Now(),
	}

	if s.cache == nil {
		check.Status = "" // 缓存未配置，不返回状态
		check.Message = "cache not configured"
		check.Response = time.Since(start).Milliseconds()
		return check
	}

	// 尝试一个简单的操作
	testKey := "health:check"
	testValue := time.Now().Unix()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := s.cache.Set(ctx, testKey, testValue, 10*time.Second); err != nil {
		check.Status = HealthStatusDOWN
		check.Message = fmt.Sprintf("cache write failed: %v", err)
		check.Response = time.Since(start).Milliseconds()
		return check
	}

	var readValue int64
	if err := s.cache.Get(ctx, testKey, &readValue); err != nil {
		check.Status = HealthStatusDOWN
		check.Message = fmt.Sprintf("cache read failed: %v", err)
		check.Response = time.Since(start).Milliseconds()
		return check
	}

	if readValue != testValue {
		check.Status = HealthStatusDOWN
		check.Message = "cache value mismatch"
		check.Response = time.Since(start).Milliseconds()
		return check
	}

	check.Status = HealthStatusUP
	check.Message = "cache read/write ok"
	check.Response = time.Since(start).Milliseconds()
	return check
}

// checkSystemResources 检查系统资源
func (s *HealthService) checkSystemResources(ctx context.Context) HealthCheck {
	start := time.Now()
	check := HealthCheck{
		Timestamp: time.Now(),
	}

	// 获取内存统计
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 获取goroutine数量
	goroutines := runtime.NumGoroutine()

	// 获取GC暂停时间（纳秒转毫秒）
	gcPause := m.PauseNs[(m.NumGC+255)%256]

	check.Status = HealthStatusUP
	check.Message = fmt.Sprintf("goroutines: %d, alloc: %dMB, gc_pause: %dms",
		goroutines, m.Alloc/1024/1024, gcPause/1000000)
	check.Response = time.Since(start).Milliseconds()

	// 如果goroutine数量过多，发出警告但不标记为DOWN
	if goroutines > 10000 {
		check.Status = "WARNING"
		check.Message += " (high goroutine count)"
	}

	return check
}

// GetDBStats 获取数据库统计
func (s *HealthService) GetDBStats() *sql.DBStats {
	if s.db == nil {
		return nil
	}

	sqlDB, err := s.db.DB()
	if err != nil {
		return nil
	}

	stats := sqlDB.Stats()
	return &stats
}
