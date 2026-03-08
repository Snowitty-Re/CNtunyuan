// Package repository 审计日志仓储实现
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"gorm.io/gorm"
)

// AuditLogRepositoryImpl 审计日志仓储实现
type AuditLogRepositoryImpl struct {
	db *gorm.DB
}

// NewAuditLogRepository 创建审计日志仓储
func NewAuditLogRepository(db *gorm.DB) repository.AuditLogRepository {
	return &AuditLogRepositoryImpl{db: db}
}

// Create 创建审计日志
func (r *AuditLogRepositoryImpl) Create(ctx context.Context, log *entity.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// FindByID 根据ID查找
func (r *AuditLogRepositoryImpl) FindByID(ctx context.Context, id string) (*entity.AuditLog, error) {
	var log entity.AuditLog
	if err := r.db.WithContext(ctx).First(&log, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("audit log not found")
		}
		return nil, err
	}
	return &log, nil
}

// List 查询列表
func (r *AuditLogRepositoryImpl) List(ctx context.Context, query *entity.AuditLogQuery) (*repository.PaginatedResult, error) {
	db := r.db.WithContext(ctx).Model(&entity.AuditLog{})

	// 应用过滤条件
	if query.UserID != "" {
		db = db.Where("user_id = ?", query.UserID)
	}
	if query.Module != "" {
		db = db.Where("module = ?", query.Module)
	}
	if query.Action != "" {
		db = db.Where("action = ?", query.Action)
	}
	if query.Type != "" {
		db = db.Where("type = ?", query.Type)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.RequestIP != "" {
		db = db.Where("request_ip = ?", query.RequestIP)
	}
	if query.StartTime != nil {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("created_at <= ?", query.EndTime)
	}
	if query.Keyword != "" {
		db = db.Where("description LIKE ? OR username LIKE ? OR action LIKE ?",
			"%"+query.Keyword+"%", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}

	// 获取总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	var logs []entity.AuditLog
	offset := (query.Page - 1) * query.PageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&logs).Error; err != nil {
		return nil, err
	}

	return &repository.PaginatedResult{
		List:     logs,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

// GetStats 获取统计信息
func (r *AuditLogRepositoryImpl) GetStats(ctx context.Context, startTime, endTime *time.Time) (*entity.AuditLogStats, error) {
	stats := &entity.AuditLogStats{}

	db := r.db.WithContext(ctx).Model(&entity.AuditLog{})
	if startTime != nil {
		db = db.Where("created_at >= ?", startTime)
	}
	if endTime != nil {
		db = db.Where("created_at <= ?", endTime)
	}

	// 总数量
	if err := db.Count(&stats.TotalCount).Error; err != nil {
		return nil, err
	}

	// 今日数量
	today := time.Now().Truncate(24 * time.Hour)
	if err := r.db.WithContext(ctx).Model(&entity.AuditLog{}).
		Where("created_at >= ?", today).Count(&stats.TodayCount).Error; err != nil {
		return nil, err
	}

	// 登录数量
	if err := db.Where("type = ?", entity.AuditLogTypeLogin).Count(&stats.LoginCount).Error; err != nil {
		return nil, err
	}

	// 操作数量（排除登录登出）
	if err := db.Where("type NOT IN ?", []entity.AuditLogType{entity.AuditLogTypeLogin, entity.AuditLogTypeLogout, entity.AuditLogTypeQuery}).
		Count(&stats.OperationCount).Error; err != nil {
		return nil, err
	}

	// 错误数量
	if err := db.Where("status = ?", entity.AuditLogStatusFailure).Count(&stats.ErrorCount).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// GetUserActivity 获取用户活动统计
func (r *AuditLogRepositoryImpl) GetUserActivity(ctx context.Context, userID string, days int) ([]*repository.UserActivityItem, error) {
	var results []*repository.UserActivityItem

	// 按日期分组统计
	err := r.db.WithContext(ctx).Raw(`
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM ty_audit_logs
		WHERE user_id = ? AND created_at >= ?
		GROUP BY DATE(created_at)
		ORDER BY date DESC
	`, userID, time.Now().AddDate(0, 0, -days)).Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetModuleStats 获取模块统计
func (r *AuditLogRepositoryImpl) GetModuleStats(ctx context.Context, startTime, endTime *time.Time) ([]*repository.ModuleStatItem, error) {
	var results []*repository.ModuleStatItem

	db := r.db.WithContext(ctx).Model(&entity.AuditLog{}).
		Select("module, COUNT(*) as count").
		Group("module").
		Order("count DESC")

	if startTime != nil {
		db = db.Where("created_at >= ?", startTime)
	}
	if endTime != nil {
		db = db.Where("created_at <= ?", endTime)
	}

	if err := db.Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

// CleanupOldLogs 清理旧日志
func (r *AuditLogRepositoryImpl) CleanupOldLogs(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("created_at < ?", before).Delete(&entity.AuditLog{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// GetRecentLogs 获取最近的日志
func (r *AuditLogRepositoryImpl) GetRecentLogs(ctx context.Context, limit int) ([]entity.AuditLog, error) {
	var logs []entity.AuditLog
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
