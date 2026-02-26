package repository

import (
	"context"

	"github.com/Snowitty-Re/CNtunyuan/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DialectRepository 方言仓库
type DialectRepository struct {
	db *gorm.DB
}

// NewDialectRepository 创建方言仓库
func NewDialectRepository(db *gorm.DB) *DialectRepository {
	return &DialectRepository{db: db}
}

// Create 创建
func (r *DialectRepository) Create(ctx context.Context, dialect *model.Dialect) error {
	return r.db.WithContext(ctx).Create(dialect).Error
}

// GetByID 根据ID获取
func (r *DialectRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Dialect, error) {
	var dialect model.Dialect
	err := r.db.WithContext(ctx).Preload("Collector").Preload("Org").Preload("Tags").First(&dialect, id).Error
	if err != nil {
		return nil, err
	}
	return &dialect, nil
}

// Update 更新
func (r *DialectRepository) Update(ctx context.Context, dialect *model.Dialect) error {
	return r.db.WithContext(ctx).Save(dialect).Error
}

// Delete 删除
func (r *DialectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Dialect{}, id).Error
}

// List 列表查询
func (r *DialectRepository) List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*model.Dialect, int64, error) {
	var dialects []*model.Dialect
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Dialect{})

	// 过滤条件
	for key, value := range filters {
		if value != nil && value != "" {
			switch key {
			case "title":
				query = query.Where("title LIKE ?", "%"+value.(string)+"%")
			case "province":
				query = query.Where("province = ?", value)
			case "city":
				query = query.Where("city = ?", value)
			case "district":
				query = query.Where("district = ?", value)
			case "collector_id":
				query = query.Where("collector_id = ?", value)
			case "org_id":
				query = query.Where("org_id = ?", value)
			default:
				query = query.Where(key+" = ?", value)
			}
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Collector").Preload("Org").Preload("Tags").
		Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&dialects).Error; err != nil {
		return nil, 0, err
	}

	return dialects, total, nil
}

// IncrementPlayCount 增加播放次数
func (r *DialectRepository) IncrementPlayCount(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.Dialect{}).Where("id = ?", id).UpdateColumn("play_count", gorm.Expr("play_count + 1")).Error
}

// AddPlayLog 添加播放记录
func (r *DialectRepository) AddPlayLog(ctx context.Context, log *model.DialectPlayLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// AddLike 添加点赞
func (r *DialectRepository) AddLike(ctx context.Context, like *model.DialectLike) error {
	return r.db.WithContext(ctx).Create(like).Error
}

// RemoveLike 取消点赞
func (r *DialectRepository) RemoveLike(ctx context.Context, dialectID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("dialect_id = ? AND user_id = ?", dialectID, userID).Delete(&model.DialectLike{}).Error
}

// IsLiked 检查是否已点赞
func (r *DialectRepository) IsLiked(ctx context.Context, dialectID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.DialectLike{}).Where("dialect_id = ? AND user_id = ?", dialectID, userID).Count(&count).Error
	return count > 0, err
}

// UpdateLikeCount 更新点赞数
func (r *DialectRepository) UpdateLikeCount(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.Dialect{}).Where("id = ?", id).UpdateColumn("like_count",
		r.db.Model(&model.DialectLike{}).Where("dialect_id = ?", id).Select("count(*)"),
	).Error
}

// AddComment 添加评论
func (r *DialectRepository) AddComment(ctx context.Context, comment *model.DialectComment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

// GetComments 获取评论列表
func (r *DialectRepository) GetComments(ctx context.Context, dialectID uuid.UUID, page, pageSize int) ([]*model.DialectComment, int64, error) {
	var comments []*model.DialectComment
	var total int64

	query := r.db.WithContext(ctx).Model(&model.DialectComment{}).Where("dialect_id = ?", dialectID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("User").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// GetNearbyDialects 获取附近的方言
func (r *DialectRepository) GetNearbyDialects(ctx context.Context, lat, lng float64, radius float64) ([]*model.Dialect, error) {
	var dialects []*model.Dialect
	
	err := r.db.WithContext(ctx).Where(
		"latitude BETWEEN ? AND ? AND longitude BETWEEN ? AND ?",
		lat-radius, lat+radius, lng-radius, lng+radius,
	).Where("status = ?", "active").
		Find(&dialects).Error
	
	return dialects, err
}
