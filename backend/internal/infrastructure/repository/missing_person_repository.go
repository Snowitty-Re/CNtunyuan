package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"gorm.io/gorm"
)

// MissingPersonRepositoryImpl 走失人员仓储实现
type MissingPersonRepositoryImpl struct {
	*BaseRepository[entity.MissingPerson]
}

// NewMissingPersonRepository 创建走失人员仓储
func NewMissingPersonRepository(db *gorm.DB) repository.MissingPersonRepository {
	return &MissingPersonRepositoryImpl{
		BaseRepository: NewBaseRepository[entity.MissingPerson](db),
	}
}

// List 分页查询
func (r *MissingPersonRepositoryImpl) List(ctx context.Context, query *repository.MissingPersonQuery) (*repository.PageResult[entity.MissingPerson], error) {
	var persons []entity.MissingPerson
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.MissingPerson{})

	// 关键词搜索
	if query.Keyword != "" {
		db = db.Where("name LIKE ? OR contact_name LIKE ? OR contact_phone LIKE ?",
			"%"+query.Keyword+"%", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}

	// 状态筛选
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	// 性别筛选
	if query.Gender != "" {
		db = db.Where("gender = ?", query.Gender)
	}

	// 年龄范围
	if query.AgeMin > 0 {
		db = db.Where("age >= ?", query.AgeMin)
	}
	if query.AgeMax > 0 {
		db = db.Where("age <= ?", query.AgeMax)
	}

	// 地区筛选
	if query.Province != "" {
		db = db.Where("province = ?", query.Province)
	}
	if query.City != "" {
		db = db.Where("city = ?", query.City)
	}
	if query.District != "" {
		db = db.Where("district = ?", query.District)
	}

	// 紧急程度
	if query.UrgencyLevel != "" {
		db = db.Where("urgency = ?", query.UrgencyLevel)
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	// 排序
	order := query.SortField + " " + query.SortOrder
	if query.SortField == "" {
		order = "created_at DESC"
	}

	// 分页查询
	if err := db.Order(order).
		Preload("Reporter").
		Preload("Assignee").
		Preload("Photos").
		Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).
		Find(&persons).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(persons, total, query.Page, query.PageSize), nil
}

// FindByReporterID 根据报告人查找
func (r *MissingPersonRepositoryImpl) FindByReporterID(ctx context.Context, reporterID string, pagination repository.Pagination) (*repository.PageResult[entity.MissingPerson], error) {
	var persons []entity.MissingPerson
	var total int64

	db := r.db.WithContext(ctx).Where("reporter_id = ?", reporterID)

	if err := db.Model(&entity.MissingPerson{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.Paginate(db, pagination).Order("created_at DESC").Find(&persons).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(persons, total, pagination.Page, pagination.PageSize), nil
}

// FindByStatus 根据状态查找
func (r *MissingPersonRepositoryImpl) FindByStatus(ctx context.Context, status entity.MissingStatus, pagination repository.Pagination) (*repository.PageResult[entity.MissingPerson], error) {
	var persons []entity.MissingPerson
	var total int64

	db := r.db.WithContext(ctx).Where("status = ?", status)

	if err := db.Model(&entity.MissingPerson{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.Paginate(db, pagination).Order("created_at DESC").Find(&persons).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(persons, total, pagination.Page, pagination.PageSize), nil
}

// FindByRegion 根据地区查找
func (r *MissingPersonRepositoryImpl) FindByRegion(ctx context.Context, province, city, district string, pagination repository.Pagination) (*repository.PageResult[entity.MissingPerson], error) {
	var persons []entity.MissingPerson
	var total int64

	db := r.db.WithContext(ctx)
	if province != "" {
		db = db.Where("province = ?", province)
	}
	if city != "" {
		db = db.Where("city = ?", city)
	}
	if district != "" {
		db = db.Where("district = ?", district)
	}

	if err := db.Model(&entity.MissingPerson{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.Paginate(db, pagination).Order("created_at DESC").Find(&persons).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(persons, total, pagination.Page, pagination.PageSize), nil
}

// UpdateStatus 更新状态
func (r *MissingPersonRepositoryImpl) UpdateStatus(ctx context.Context, id string, status entity.MissingStatus) error {
	return r.db.WithContext(ctx).
		Model(&entity.MissingPerson{}).
		Where("id = ?", id).
		Update("status", status).
		Error
}

// AddTrack 添加轨迹
func (r *MissingPersonRepositoryImpl) AddTrack(ctx context.Context, track *entity.MissingPersonTrack) error {
	return r.db.WithContext(ctx).Create(track).Error
}

// GetTracks 获取轨迹
func (r *MissingPersonRepositoryImpl) GetTracks(ctx context.Context, personID string) ([]entity.MissingPersonTrack, error) {
	var tracks []entity.MissingPersonTrack
	err := r.db.WithContext(ctx).
		Where("missing_person_id = ?", personID).
		Order("time DESC").
		Preload("Reporter").
		Find(&tracks).Error
	return tracks, err
}

// GetStats 获取统计
func (r *MissingPersonRepositoryImpl) GetStats(ctx context.Context) (*entity.MissingPersonStats, error) {
	stats := &entity.MissingPersonStats{}

	db := r.db.WithContext(ctx).Model(&entity.MissingPerson{})

	// 总数
	if err := db.Count(&stats.Total).Error; err != nil {
		return nil, err
	}

	// 各状态统计
	for _, status := range []entity.MissingStatus{
		entity.MissingStatusMissing,
		entity.MissingStatusSearching,
		entity.MissingStatusFound,
		entity.MissingStatusReunited,
		entity.MissingStatusClosed,
	} {
		var count int64
		if err := db.Where("status = ?", status).Count(&count).Error; err != nil {
			return nil, err
		}
		switch status {
		case entity.MissingStatusMissing:
			stats.Missing = count
		case entity.MissingStatusSearching:
			stats.Searching = count
		case entity.MissingStatusFound:
			stats.Found = count
		case entity.MissingStatusReunited:
			stats.Reunited = count
		case entity.MissingStatusClosed:
			stats.Closed = count
		}
	}

	// 今日新增
	today := time.Now().Format("2006-01-02")
	if err := db.Where("DATE(created_at) = ?", today).Count(&stats.TodayNew).Error; err != nil {
		return nil, err
	}

	// 本周新增
	weekStart := time.Now().AddDate(0, 0, -int(time.Now().Weekday())).Format("2006-01-02")
	if err := db.Where("DATE(created_at) >= ?", weekStart).Count(&stats.ThisWeekNew).Error; err != nil {
		return nil, err
	}

	// 本月新增
	monthStart := time.Now().Format("2006-01") + "-01"
	if err := db.Where("DATE(created_at) >= ?", monthStart).Count(&stats.ThisMonthNew).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// Search 全文搜索
func (r *MissingPersonRepositoryImpl) Search(ctx context.Context, keyword string, pagination repository.Pagination) (*repository.PageResult[entity.MissingPerson], error) {
	var persons []entity.MissingPerson
	var total int64

	db := r.db.WithContext(ctx).Where(
		"name LIKE ? OR description LIKE ? OR features LIKE ? OR clothes LIKE ? OR address LIKE ?",
		"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%",
	)

	if err := db.Model(&entity.MissingPerson{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.Paginate(db, pagination).Order("created_at DESC").Find(&persons).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(persons, total, pagination.Page, pagination.PageSize), nil
}

// CountByStatus 按状态统计
func (r *MissingPersonRepositoryImpl) CountByStatus(ctx context.Context, status entity.MissingStatus) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.MissingPerson{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// CountByDateRange 按日期范围统计
func (r *MissingPersonRepositoryImpl) CountByDateRange(ctx context.Context, start, end string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.MissingPerson{}).
		Where("created_at >= ? AND created_at <= ?", start, end).
		Count(&count).Error
	return count, err
}

// IncrementViews 增加浏览次数
func (r *MissingPersonRepositoryImpl) IncrementViews(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&entity.MissingPerson{}).Where("id = ?", id).UpdateColumn("views", gorm.Expr("views + 1")).Error
}

// FindByID 根据ID查找
func (r *MissingPersonRepositoryImpl) FindByID(ctx context.Context, id string) (*entity.MissingPerson, error) {
	var person entity.MissingPerson
	err := r.db.WithContext(ctx).Preload("Reporter").Preload("Assignee").Preload("Photos").First(&person, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("missing person not found")
		}
		return nil, err
	}
	return &person, nil
}
