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
	
	// CreateBatch 批量创建审计日志
	CreateBatch(ctx context.Context, logs []*entity.AuditLog) error
	
	// FindByID 根据ID查询
	FindByID(ctx context.Context, id string) (*entity.AuditLog, error)
	
	// List 查询审计日志列表
	List(ctx context.Context, query *entity.AuditLogQuery) ([]*entity.AuditLog, int64, error)
	
	// ListByUser 查询用户的审计日志
	ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*entity.AuditLog, int64, error)
	
	// ListByResource 查询资源的审计日志
	ListByResource(ctx context.Context, resource string, resourceID string, page, pageSize int) ([]*entity.AuditLog, int64, error)
	
	// ListByTimeRange 查询时间范围内的审计日志
	ListByTimeRange(ctx context.Context, startTime, endTime time.Time, page, pageSize int) ([]*entity.AuditLog, int64, error)
	
	// GetStats 获取统计信息
	GetStats(ctx context.Context, orgID string, startTime, endTime *time.Time) (*entity.AuditLogStats, error)
	
	// GetActionStats 获取操作类型统计
	GetActionStats(ctx context.Context, orgID string, startTime, endTime *time.Time) (map[string]int64, error)
	
	// GetResourceStats 获取资源类型统计
	GetResourceStats(ctx context.Context, orgID string, startTime, endTime *time.Time) (map[string]int64, error)
	
	// GetUserActivityStats 获取用户活动统计
	GetUserActivityStats(ctx context.Context, userID string, days int) (map[string]interface{}, error)
	
	// CleanupOldLogs 清理过期日志
	CleanupOldLogs(ctx context.Context, before time.Time) (int64, error)
	
	// GetRetentionStats 获取保留统计
	GetRetentionStats(ctx context.Context, days int) (map[string]interface{}, error)
}
