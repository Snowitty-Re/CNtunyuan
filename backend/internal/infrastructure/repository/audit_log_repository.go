package repository

import (
	"context"
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

// CreateBatch 批量创建审计日志
func (r *AuditLogRepositoryImpl) CreateBatch(ctx context.Context, logs []*entity.AuditLog) error {
	if len(logs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(logs, 100).Error
}

// FindByID 根据ID查询
func (r *AuditLogRepositoryImpl) FindByID(ctx context.Context, id string) (*entity.AuditLog, error) {
	var log entity.AuditLog
	err := r.db.WithContext(ctx).First(&log, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &log, nil
}

// List 查询审计日志列表
func (r *AuditLogRepositoryImpl) List(ctx context.Context, query *entity.AuditLogQuery) ([]*entity.AuditLog, int64, error) {
	db := r.db.WithContext(ctx).Model(&entity.AuditLog{})
	
	// 应用查询条件
	if query.UserID != "" {
		db = db.Where("user_id = ?", query.UserID)
	}
	if query.OrgID != "" {
		db = db.Where("org_id = ?", query.OrgID)
	}
	if query.Action != "" {
		db = db.Where("action = ?", query.Action)
	}
	if query.Resource != "" {
		db = db.Where("resource = ?", query.Resource)
	}
	if query.ResourceID != "" {
		db = db.Where("resource_id = ?", query.ResourceID)
	}
	if query.IPAddress != "" {
		db = db.Where("ip_address = ?", query.IPAddress)
	}
	if query.TraceID != "" {
		db = db.Where("trace_id = ?", query.TraceID)
	}
	if query.StartTime != nil {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("created_at <= ?", query.EndTime)
	}
	
	// 统计总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 分页查询
	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	
	var logs []*entity.AuditLog
	err := db.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error
	
	return logs, total, err
}

// ListByUser 查询用户的审计日志
func (r *AuditLogRepositoryImpl) ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*entity.AuditLog, int64, error) {
	query := &entity.AuditLogQuery{
		UserID:   userID,
		Page:     page,
		PageSize: pageSize,
	}
	return r.List(ctx, query)
}

// ListByResource 查询资源的审计日志
func (r *AuditLogRepositoryImpl) ListByResource(ctx context.Context, resource string, resourceID string, page, pageSize int) ([]*entity.AuditLog, int64, error) {
	query := &entity.AuditLogQuery{
		Resource:   resource,
		ResourceID: resourceID,
		Page:       page,
		PageSize:   pageSize,
	}
	return r.List(ctx, query)
}

// ListByTimeRange 查询时间范围内的审计日志
func (r *AuditLogRepositoryImpl) ListByTimeRange(ctx context.Context, startTime, endTime time.Time, page, pageSize int) ([]*entity.AuditLog, int64, error) {
	query := &entity.AuditLogQuery{
		StartTime: &startTime,
		EndTime:   &endTime,
		Page:      page,
		PageSize:  pageSize,
	}
	return r.List(ctx, query)
}

// GetStats 获取统计信息
func (r *AuditLogRepositoryImpl) GetStats(ctx context.Context, orgID string, startTime, endTime *time.Time) (*entity.AuditLogStats, error) {
	db := r.db.WithContext(ctx).Model(&entity.AuditLog{})
	
	if orgID != "" {
		db = db.Where("org_id = ?", orgID)
	}
	if startTime != nil {
		db = db.Where("created_at >= ?", startTime)
	}
	if endTime != nil {
		db = db.Where("created_at <= ?", endTime)
	}
	
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}
	
	// 今日统计
	today := time.Now().Truncate(24 * time.Hour)
	var todayCount int64
	todayDB := r.db.WithContext(ctx).Model(&entity.AuditLog{}).Where("created_at >= ?", today)
	if orgID != "" {
		todayDB = todayDB.Where("org_id = ?", orgID)
	}
	if err := todayDB.Count(&todayCount).Error; err != nil {
		return nil, err
	}
	
	stats := &entity.AuditLogStats{
		TotalCount:  total,
		TodayCount:  todayCount,
		ActionStats: make(map[string]int64),
		ResourceStats: make(map[string]int64),
	}
	
	return stats, nil
}

// GetActionStats 获取操作类型统计
func (r *AuditLogRepositoryImpl) GetActionStats(ctx context.Context, orgID string, startTime, endTime *time.Time) (map[string]int64, error) {
	db := r.db.WithContext(ctx).Model(&entity.AuditLog{})
	
	if orgID != "" {
		db = db.Where("org_id = ?", orgID)
	}
	if startTime != nil {
		db = db.Where("created_at >= ?", startTime)
	}
	if endTime != nil {
		db = db.Where("created_at <= ?", endTime)
	}
	
	var results []struct {
		Action string
		Count  int64
	}
	
	err := db.Select("action, COUNT(*) as count").
		Group("action").
		Find(&results).Error
	
	if err != nil {
		return nil, err
	}
	
	stats := make(map[string]int64)
	for _, r := range results {
		stats[r.Action] = r.Count
	}
	
	return stats, nil
}

// GetResourceStats 获取资源类型统计
func (r *AuditLogRepositoryImpl) GetResourceStats(ctx context.Context, orgID string, startTime, endTime *time.Time) (map[string]int64, error) {
	db := r.db.WithContext(ctx).Model(&entity.AuditLog{})
	
	if orgID != "" {
		db = db.Where("org_id = ?", orgID)
	}
	if startTime != nil {
		db = db.Where("created_at >= ?", startTime)
	}
	if endTime != nil {
		db = db.Where("created_at <= ?", endTime)
	}
	
	var results []struct {
		Resource string
		Count    int64
	}
	
	err := db.Select("resource, COUNT(*) as count").
		Group("resource").
		Find(&results).Error
	
	if err != nil {
		return nil, err
	}
	
	stats := make(map[string]int64)
	for _, r := range results {
		stats[r.Resource] = r.Count
	}
	
	return stats, nil
}

// GetUserActivityStats 获取用户活动统计
func (r *AuditLogRepositoryImpl) GetUserActivityStats(ctx context.Context, userID string, days int) (map[string]interface{}, error) {
	if days <= 0 {
		days = 7
	}
	
	startTime := time.Now().AddDate(0, 0, -days)
	
	// 总操作数
	var totalCount int64
	err := r.db.WithContext(ctx).
		Model(&entity.AuditLog{}).
		Where("user_id = ?", userID).
		Where("created_at >= ?", startTime).
		Count(&totalCount).Error
	if err != nil {
		return nil, err
	}
	
	// 按操作类型统计
	actionStats, err := r.getUserActionStats(ctx, userID, startTime)
	if err != nil {
		return nil, err
	}
	
	// 按资源类型统计
	resourceStats, err := r.getUserResourceStats(ctx, userID, startTime)
	if err != nil {
		return nil, err
	}
	
	// 按天统计
	dailyStats, err := r.getUserDailyStats(ctx, userID, days)
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"total_count":    totalCount,
		"action_stats":   actionStats,
		"resource_stats": resourceStats,
		"daily_stats":    dailyStats,
		"period_days":    days,
	}, nil
}

// getUserActionStats 获取用户操作类型统计
func (r *AuditLogRepositoryImpl) getUserActionStats(ctx context.Context, userID string, startTime time.Time) (map[string]int64, error) {
	var results []struct {
		Action string
		Count  int64
	}
	
	err := r.db.WithContext(ctx).
		Model(&entity.AuditLog{}).
		Where("user_id = ?", userID).
		Where("created_at >= ?", startTime).
		Select("action, COUNT(*) as count").
		Group("action").
		Find(&results).Error
	
	if err != nil {
		return nil, err
	}
	
	stats := make(map[string]int64)
	for _, r := range results {
		stats[r.Action] = r.Count
	}
	return stats, nil
}

// getUserResourceStats 获取用户资源类型统计
func (r *AuditLogRepositoryImpl) getUserResourceStats(ctx context.Context, userID string, startTime time.Time) (map[string]int64, error) {
	var results []struct {
		Resource string
		Count    int64
	}
	
	err := r.db.WithContext(ctx).
		Model(&entity.AuditLog{}).
		Where("user_id = ?", userID).
		Where("created_at >= ?", startTime).
		Select("resource, COUNT(*) as count").
		Group("resource").
		Find(&results).Error
	
	if err != nil {
		return nil, err
	}
	
	stats := make(map[string]int64)
	for _, r := range results {
		stats[r.Resource] = r.Count
	}
	return stats, nil
}

// getUserDailyStats 获取用户每日统计
func (r *AuditLogRepositoryImpl) getUserDailyStats(ctx context.Context, userID string, days int) (map[string]int64, error) {
	stats := make(map[string]int64)
	
	for i := 0; i < days; i++ {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		startOfDay := time.Now().AddDate(0, 0, -i).Truncate(24 * time.Hour)
		endOfDay := startOfDay.Add(24 * time.Hour)
		
		var count int64
		err := r.db.WithContext(ctx).
			Model(&entity.AuditLog{}).
			Where("user_id = ?", userID).
			Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).
			Count(&count).Error
		if err != nil {
			return nil, err
		}
		
		stats[date] = count
	}
	
	return stats, nil
}

// CleanupOldLogs 清理过期日志
func (r *AuditLogRepositoryImpl) CleanupOldLogs(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("created_at < ?", before).
		Delete(&entity.AuditLog{})
	
	return result.RowsAffected, result.Error
}

// GetRetentionStats 获取保留统计
func (r *AuditLogRepositoryImpl) GetRetentionStats(ctx context.Context, days int) (map[string]interface{}, error) {
	if days <= 0 {
		days = 365
	}
	
	cutoffDate := time.Now().AddDate(0, 0, -days)
	
	// 总日志数
	var totalCount int64
	if err := r.db.WithContext(ctx).Model(&entity.AuditLog{}).Count(&totalCount).Error; err != nil {
		return nil, err
	}
	
	// 需要清理的日志数
	var oldCount int64
	if err := r.db.WithContext(ctx).
		Model(&entity.AuditLog{}).
		Where("created_at < ?", cutoffDate).
		Count(&oldCount).Error; err != nil {
		return nil, err
	}
	
	// 最旧的日志时间
	var oldestLog time.Time
	row := r.db.WithContext(ctx).
		Model(&entity.AuditLog{}).
		Select("MIN(created_at)").
		Row()
	row.Scan(&oldestLog)
	
	// 按月的日志数量
	var monthlyStats []struct {
		Month string
		Count int64
	}
	
	err := r.db.WithContext(ctx).
		Model(&entity.AuditLog{}).
		Select("TO_CHAR(created_at, 'YYYY-MM') as month, COUNT(*) as count").
		Group("TO_CHAR(created_at, 'YYYY-MM')").
		Order("month DESC").
		Limit(12).
		Find(&monthlyStats).Error
	
	if err != nil {
		// 如果不是PostgreSQL，使用不同的语法
		err = r.db.WithContext(ctx).
			Raw("SELECT DATE_FORMAT(created_at, '%Y-%m') as month, COUNT(*) as count FROM ty_audit_logs GROUP BY DATE_FORMAT(created_at, '%Y-%m') ORDER BY month DESC LIMIT 12").
			Scan(&monthlyStats).Error
	}
	
	monthlyMap := make(map[string]int64)
	for _, m := range monthlyStats {
		monthlyMap[m.Month] = m.Count
	}
	
	return map[string]interface{}{
		"total_count":      totalCount,
		"retention_days":   days,
		"logs_to_cleanup":  oldCount,
		"oldest_log_time":  oldestLog,
		"monthly_stats":    monthlyMap,
	}, nil
}
