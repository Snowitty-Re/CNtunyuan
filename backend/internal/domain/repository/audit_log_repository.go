// Package repository 审计日志仓储接口
package repository

import (
	"context"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
)

// AuditLogRepository 审计日志仓储接口
type AuditLogRepository interface {
	// Create 创建审计日志
	Create(ctx context.Context, log *entity.AuditLog) error
	
	// FindByID 根据ID查找
	FindByID(ctx context.Context, id string) (*entity.AuditLog, error)
	
	// List 查询列表
	List(ctx context.Context, query *entity.AuditLogQuery) (*PaginatedResult, error)
	
	// GetStats 获取统计信息
	GetStats(ctx context.Context, startTime, endTime *time.Time) (*entity.AuditLogStats, error)
	
	// GetUserActivity 获取用户活动统计
	GetUserActivity(ctx context.Context, userID string, days int) ([]*UserActivityItem, error)
	
	// GetModuleStats 获取模块统计
	GetModuleStats(ctx context.Context, startTime, endTime *time.Time) ([]*ModuleStatItem, error)
	
	// CleanupOldLogs 清理旧日志
	CleanupOldLogs(ctx context.Context, before time.Time) (int64, error)
	
	// GetRecentLogs 获取最近的日志
	GetRecentLogs(ctx context.Context, limit int) ([]entity.AuditLog, error)
}

// PaginatedResult 分页结果
type PaginatedResult struct {
	List     []entity.AuditLog `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// UserActivityItem 用户活动项
type UserActivityItem struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// ModuleStatItem 模块统计项
type ModuleStatItem struct {
	Module string `json:"module"`
	Count  int64  `json:"count"`
}
