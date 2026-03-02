package repository

import (
	"context"
	"errors"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"gorm.io/gorm"
)

// DialectRepositoryImpl 方言仓储实现
type DialectRepositoryImpl struct {
	*BaseRepository[entity.Dialect]
}

// NewDialectRepository 创建方言仓储
func NewDialectRepository(db *gorm.DB) repository.DialectRepository {
	return &DialectRepositoryImpl{
		BaseRepository: NewBaseRepository[entity.Dialect](db),
	}
}

// List 分页查询
func (r *DialectRepositoryImpl) List(ctx context.Context, query *repository.DialectQuery) (*repository.PageResult[entity.Dialect], error) {
	var dialects []entity.Dialect
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Dialect{})

	// 关键词搜索
	if query.Keyword != "" {
		db = db.Where("title LIKE ? OR content LIKE ? OR region LIKE ?",
			"%"+query.Keyword+"%", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}

	// 地区筛选
	if query.Region != "" {
		db = db.Where("region = ?", query.Region)
	}
	if query.Province != "" {
		db = db.Where("province = ?", query.Province)
	}
	if query.City != "" {
		db = db.Where("city = ?", query.City)
	}

	// 类型筛选
	if query.Type != "" {
		db = db.Where("dialect_type = ?", query.Type)
	}

	// 状态筛选
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	// 上传者筛选
	if query.UploaderID != "" {
		db = db.Where("uploader_id = ?", query.UploaderID)
	}

	// 精选筛选
	if query.IsFeatured != nil {
		db = db.Where("is_featured = ?", *query.IsFeatured)
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	// 排序
	sortBy := query.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortOrder := query.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}
	order := sortBy + " " + sortOrder

	// 分页查询
	if err := db.Order(order).
		Preload("Uploader").
		Preload("Org").
		Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).
		Find(&dialects).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(dialects, total, query.Page, query.PageSize), nil
}

// FindByRegion 根据地区查找
func (r *DialectRepositoryImpl) FindByRegion(ctx context.Context, province, city string, pagination repository.Pagination) (*repository.PageResult[entity.Dialect], error) {
	var dialects []entity.Dialect
	var total int64

	db := r.db.WithContext(ctx)

	if province != "" {
		db = db.Where("province = ?", province)
	}
	if city != "" {
		db = db.Where("city = ?", city)
	}

	if err := db.Model(&entity.Dialect{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.Paginate(db, pagination).Find(&dialects).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(dialects, total, pagination.Page, pagination.PageSize), nil
}

// FindByUploader 根据上传者查找
func (r *DialectRepositoryImpl) FindByUploader(ctx context.Context, uploaderID string, pagination repository.Pagination) (*repository.PageResult[entity.Dialect], error) {
	var dialects []entity.Dialect
	var total int64

	db := r.db.WithContext(ctx).Where("uploader_id = ?", uploaderID)

	if err := db.Model(&entity.Dialect{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.Paginate(db, pagination).Find(&dialects).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(dialects, total, pagination.Page, pagination.PageSize), nil
}

// FindFeatured 查找精选
func (r *DialectRepositoryImpl) FindFeatured(ctx context.Context, pagination repository.Pagination) (*repository.PageResult[entity.Dialect], error) {
	var dialects []entity.Dialect
	var total int64

	db := r.db.WithContext(ctx).Where("is_featured = ?", true)

	if err := db.Model(&entity.Dialect{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.Paginate(db, pagination).Find(&dialects).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(dialects, total, pagination.Page, pagination.PageSize), nil
}

// FindByType 根据类型查找
func (r *DialectRepositoryImpl) FindByType(ctx context.Context, dialectType entity.DialectType, pagination repository.Pagination) (*repository.PageResult[entity.Dialect], error) {
	var dialects []entity.Dialect
	var total int64

	db := r.db.WithContext(ctx).Where("dialect_type = ?", dialectType)

	if err := db.Model(&entity.Dialect{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.Paginate(db, pagination).Find(&dialects).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(dialects, total, pagination.Page, pagination.PageSize), nil
}

// Search 搜索
func (r *DialectRepositoryImpl) Search(ctx context.Context, keyword string, pagination repository.Pagination) (*repository.PageResult[entity.Dialect], error) {
	var dialects []entity.Dialect
	var total int64

	db := r.db.WithContext(ctx).
		Where("title LIKE ? OR content LIKE ? OR region LIKE ? OR description LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")

	if err := db.Model(&entity.Dialect{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.Paginate(db, pagination).Find(&dialects).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(dialects, total, pagination.Page, pagination.PageSize), nil
}

// IncrementPlayCount 增加播放次数
func (r *DialectRepositoryImpl) IncrementPlayCount(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&entity.Dialect{}).
		Where("id = ?", id).
		UpdateColumn("play_count", gorm.Expr("play_count + 1")).
		Error
}

// IncrementLikeCount 增加点赞数
func (r *DialectRepositoryImpl) IncrementLikeCount(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&entity.Dialect{}).
		Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).
		Error
}

// DecrementLikeCount 减少点赞数
func (r *DialectRepositoryImpl) DecrementLikeCount(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&entity.Dialect{}).
		Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("CASE WHEN like_count > 0 THEN like_count - 1 ELSE 0 END")).
		Error
}

// AddComment 添加评论
func (r *DialectRepositoryImpl) AddComment(ctx context.Context, comment *entity.DialectComment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
			return err
		}
		// 更新评论数
		return tx.Model(&entity.Dialect{}).
			Where("id = ?", comment.DialectID).
			UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).
			Error
	})
}

// GetComments 获取评论
func (r *DialectRepositoryImpl) GetComments(ctx context.Context, dialectID string, pagination repository.Pagination) (*repository.PageResult[entity.DialectComment], error) {
	var comments []entity.DialectComment
	var total int64

	db := r.db.WithContext(ctx).Where("dialect_id = ?", dialectID)

	if err := db.Model(&entity.DialectComment{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.Paginate(db.Order("created_at DESC"), pagination).
		Preload("User").
		Find(&comments).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(comments, total, pagination.Page, pagination.PageSize), nil
}

// AddLike 添加点赞
func (r *DialectRepositoryImpl) AddLike(ctx context.Context, like *entity.DialectLike) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 检查是否已点赞
		var count int64
		if err := tx.Model(&entity.DialectLike{}).
			Where("dialect_id = ? AND user_id = ?", like.DialectID, like.UserID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("already liked")
		}

		if err := tx.Create(like).Error; err != nil {
			return err
		}

		// 更新点赞数
		return tx.Model(&entity.Dialect{}).
			Where("id = ?", like.DialectID).
			UpdateColumn("like_count", gorm.Expr("like_count + 1")).
			Error
	})
}

// RemoveLike 取消点赞
func (r *DialectRepositoryImpl) RemoveLike(ctx context.Context, dialectID, userID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("dialect_id = ? AND user_id = ?", dialectID, userID).Delete(&entity.DialectLike{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("like not found")
		}

		// 更新点赞数
		return tx.Model(&entity.Dialect{}).
			Where("id = ?", dialectID).
			UpdateColumn("like_count", gorm.Expr("CASE WHEN like_count > 0 THEN like_count - 1 ELSE 0 END")).
			Error
	})
}

// HasLiked 是否已点赞
func (r *DialectRepositoryImpl) HasLiked(ctx context.Context, dialectID, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.DialectLike{}).
		Where("dialect_id = ? AND user_id = ?", dialectID, userID).
		Count(&count).Error
	return count > 0, err
}

// AddPlayLog 添加播放记录
func (r *DialectRepositoryImpl) AddPlayLog(ctx context.Context, log *entity.DialectPlayLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetStats 获取统计
func (r *DialectRepositoryImpl) GetStats(ctx context.Context) (*entity.DialectStats, error) {
	stats := &entity.DialectStats{}

	// 总数
	if err := r.db.WithContext(ctx).Model(&entity.Dialect{}).Count(&stats.Total).Error; err != nil {
		return nil, err
	}

	// 活跃数
	if err := r.db.WithContext(ctx).Model(&entity.Dialect{}).
		Where("status = ?", entity.DialectStatusActive).Count(&stats.Active).Error; err != nil {
		return nil, err
	}

	// 待审核数
	if err := r.db.WithContext(ctx).Model(&entity.Dialect{}).
		Where("status = ?", entity.DialectStatusPending).Count(&stats.Pending).Error; err != nil {
		return nil, err
	}

	// 精选数
	if err := r.db.WithContext(ctx).Model(&entity.Dialect{}).
		Where("is_featured = ?", true).Count(&stats.Featured).Error; err != nil {
		return nil, err
	}

	// 总播放数
	if err := r.db.WithContext(ctx).Model(&entity.Dialect{}).
		Select("COALESCE(SUM(play_count), 0)").Scan(&stats.TotalPlays).Error; err != nil {
		return nil, err
	}

	// 总点赞数
	if err := r.db.WithContext(ctx).Model(&entity.Dialect{}).
		Select("COALESCE(SUM(like_count), 0)").Scan(&stats.TotalLikes).Error; err != nil {
		return nil, err
	}

	return stats, nil
}
