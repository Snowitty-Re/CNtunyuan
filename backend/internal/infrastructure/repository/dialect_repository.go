package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

// Create 创建方言（兼容 PostgreSQL JSON/UUID 字段）
func (r *DialectRepositoryImpl) Create(ctx context.Context, dialect *entity.Dialect) error {
	sanitizeDialectOptionalFields(dialect)

	db := r.db.WithContext(ctx)
	if strings.TrimSpace(dialect.Tags) == "" {
		db = db.Omit("Tags")
	}
	if err := db.Create(dialect).Error; err != nil {
		if shouldRetryDialectWithoutLegacyColumns(err) {
			return db.Omit("MissingPersonID", "CollectAddress", "CollectLatitude", "CollectLongitude").Create(dialect).Error
		}
		return err
	}
	return nil
}

// Update 更新方言（兼容 PostgreSQL JSON/UUID 字段）
func (r *DialectRepositoryImpl) Update(ctx context.Context, dialect *entity.Dialect) error {
	sanitizeDialectOptionalFields(dialect)

	db := r.db.WithContext(ctx)
	if strings.TrimSpace(dialect.Tags) == "" {
		db = db.Omit("Tags")
	}
	if err := db.Save(dialect).Error; err != nil {
		if shouldRetryDialectWithoutLegacyColumns(err) {
			return db.Omit("MissingPersonID", "CollectAddress", "CollectLatitude", "CollectLongitude").Save(dialect).Error
		}
		return err
	}
	return nil
}

func sanitizeDialectOptionalFields(dialect *entity.Dialect) {
	if dialect.MissingPersonID != nil && strings.TrimSpace(*dialect.MissingPersonID) == "" {
		dialect.MissingPersonID = nil
	}
	if dialect.BatchID != nil && strings.TrimSpace(*dialect.BatchID) == "" {
		dialect.BatchID = nil
	}
	if dialect.CardGroupID != nil && strings.TrimSpace(*dialect.CardGroupID) == "" {
		dialect.CardGroupID = nil
	}
	if dialect.CardID != nil && strings.TrimSpace(*dialect.CardID) == "" {
		dialect.CardID = nil
	}
}

func shouldRetryDialectWithoutLegacyColumns(err error) bool {
	errMsg := strings.ToLower(err.Error())
	if !strings.Contains(errMsg, "does not exist") && !strings.Contains(errMsg, "unknown column") {
		return false
	}
	return strings.Contains(errMsg, "missing_person_id") ||
		strings.Contains(errMsg, "collect_address") ||
		strings.Contains(errMsg, "collect_latitude") ||
		strings.Contains(errMsg, "collect_longitude")
}

// List 分页查询
func (r *DialectRepositoryImpl) List(ctx context.Context, query *repository.DialectQuery) (*repository.PageResult[entity.Dialect], error) {
	var dialects []entity.Dialect
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Dialect{})

	// 关键词搜索
	if query.Keyword != "" {
		db = db.Where("title LIKE ? OR content LIKE ? OR description LIKE ?",
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
	if query.OrgID != "" {
		db = db.Where("org_id = ?", query.OrgID)
	}
	if len(query.OrgIDs) > 0 {
		db = db.Where("org_id IN ?", query.OrgIDs)
	}

	// 精选筛选
	if query.IsFeatured != nil {
		db = db.Where("is_featured = ?", *query.IsFeatured)
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	// 排序（白名单校验防止 SQL 注入）
	order := "created_at DESC"
	if query.SortBy != "" {
		allowedSortFields := map[string]bool{
			"created_at": true, "updated_at": true, "play_count": true,
			"like_count": true, "title": true, "region": true,
		}
		if allowedSortFields[query.SortBy] {
			sortOrder := "DESC"
			if query.SortOrder == "asc" || query.SortOrder == "ASC" {
				sortOrder = "ASC"
			}
			order = query.SortBy + " " + sortOrder
		}
	}

	// 分页查询
	if err := db.Order(order).
		Preload("Uploader").
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

	if err := r.Paginate(db, pagination).Order("created_at DESC").Find(&dialects).Error; err != nil {
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

	if err := r.Paginate(db, pagination).Order("created_at DESC").Find(&dialects).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(dialects, total, pagination.Page, pagination.PageSize), nil
}

// FindFeatured 查找精选
func (r *DialectRepositoryImpl) FindFeatured(ctx context.Context, pagination repository.Pagination) (*repository.PageResult[entity.Dialect], error) {
	var dialects []entity.Dialect
	var total int64

	db := r.db.WithContext(ctx).Where("is_featured = ? AND status = ?", true, entity.DialectStatusActive)

	if err := db.Model(&entity.Dialect{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.Paginate(db, pagination).Order("created_at DESC").Find(&dialects).Error; err != nil {
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

	if err := r.Paginate(db, pagination).Order("created_at DESC").Find(&dialects).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(dialects, total, pagination.Page, pagination.PageSize), nil
}

// Search 搜索
func (r *DialectRepositoryImpl) Search(ctx context.Context, keyword string, pagination repository.Pagination) (*repository.PageResult[entity.Dialect], error) {
	var dialects []entity.Dialect
	var total int64

	db := r.db.WithContext(ctx).Where(
		"title LIKE ? OR content LIKE ? OR region LIKE ? OR description LIKE ?",
		"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%",
	)

	if err := db.Model(&entity.Dialect{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.Paginate(db, pagination).Order("created_at DESC").Find(&dialects).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(dialects, total, pagination.Page, pagination.PageSize), nil
}

// IncrementPlayCount 增加播放次数
func (r *DialectRepositoryImpl) IncrementPlayCount(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&entity.Dialect{}).Where("id = ?", id).UpdateColumn("play_count", gorm.Expr("play_count + 1")).Error
}

// IncrementLikeCount 增加点赞数
func (r *DialectRepositoryImpl) IncrementLikeCount(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&entity.Dialect{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

// DecrementLikeCount 减少点赞数
func (r *DialectRepositoryImpl) DecrementLikeCount(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&entity.Dialect{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("CASE WHEN like_count > 0 THEN like_count - 1 ELSE 0 END")).Error
}

// AddComment 添加评论
func (r *DialectRepositoryImpl) AddComment(ctx context.Context, comment *entity.DialectComment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
			return err
		}
		// 更新评论数
		return tx.Model(&entity.Dialect{}).Where("id = ?", comment.DialectID).UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error
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

	if err := db.Order("created_at DESC").Preload("User").
		Offset((pagination.Page - 1) * pagination.PageSize).
		Limit(pagination.PageSize).
		Find(&comments).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(comments, total, pagination.Page, pagination.PageSize), nil
}

// AddLike 添加点赞
func (r *DialectRepositoryImpl) AddLike(ctx context.Context, like *entity.DialectLike) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(like).Error; err != nil {
			return err
		}
		// 更新点赞数
		return tx.Model(&entity.Dialect{}).Where("id = ?", like.DialectID).UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
	})
}

// RemoveLike 取消点赞
func (r *DialectRepositoryImpl) RemoveLike(ctx context.Context, dialectID, userID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("dialect_id = ? AND user_id = ?", dialectID, userID).Delete(&entity.DialectLike{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected > 0 {
			// 更新点赞数
			return tx.Model(&entity.Dialect{}).Where("id = ?", dialectID).UpdateColumn("like_count", gorm.Expr("CASE WHEN like_count > 0 THEN like_count - 1 ELSE 0 END")).Error
		}
		return nil
	})
}

// HasLiked 是否已点赞
func (r *DialectRepositoryImpl) HasLiked(ctx context.Context, dialectID, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.DialectLike{}).Where("dialect_id = ? AND user_id = ?", dialectID, userID).Count(&count).Error
	return count > 0, err
}

// AddPlayLog 添加播放记录
func (r *DialectRepositoryImpl) AddPlayLog(ctx context.Context, log *entity.DialectPlayLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetStats 获取统计
func (r *DialectRepositoryImpl) GetStats(ctx context.Context) (*entity.DialectStats, error) {
	stats := &entity.DialectStats{}
	db := r.db.WithContext(ctx).Model(&entity.Dialect{})

	// 总数
	if err := db.Count(&stats.Total).Error; err != nil {
		return nil, err
	}

	// 活跃数
	if err := db.Where("status = ?", entity.DialectStatusActive).Count(&stats.Active).Error; err != nil {
		return nil, err
	}

	// 待审核数
	if err := db.Where("status = ?", entity.DialectStatusPending).Count(&stats.Pending).Error; err != nil {
		return nil, err
	}

	// 精选数
	if err := db.Where("is_featured = ?", true).Count(&stats.Featured).Error; err != nil {
		return nil, err
	}

	// 总播放数
	if err := db.Select("COALESCE(SUM(play_count), 0)").Scan(&stats.TotalPlays).Error; err != nil {
		return nil, err
	}

	// 总点赞数
	if err := db.Select("COALESCE(SUM(like_count), 0)").Scan(&stats.TotalLikes).Error; err != nil {
		return nil, err
	}

	// 总评论数
	if err := db.Select("COALESCE(SUM(comment_count), 0)").Scan(&stats.TotalComments).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// FindByID 根据ID查找
func (r *DialectRepositoryImpl) FindByID(ctx context.Context, id string) (*entity.Dialect, error) {
	var dialect entity.Dialect
	err := r.db.WithContext(ctx).
		Preload("Uploader").
		Preload("Card").
		Preload("Card.Group").
		First(&dialect, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("dialect not found")
		}
		return nil, err
	}
	return &dialect, nil
}

// ListCardGroups 获取卡片分组列表
func (r *DialectRepositoryImpl) ListCardGroups(ctx context.Context, includeInactive bool) ([]entity.DialectCardGroup, error) {
	var groups []entity.DialectCardGroup
	db := r.db.WithContext(ctx).Model(&entity.DialectCardGroup{})
	if !includeInactive {
		db = db.Where("status = ?", entity.DialectCardGroupStatusActive)
	}
	err := db.Preload("Cards", func(tx *gorm.DB) *gorm.DB {
		if includeInactive {
			return tx.Order("sort_order ASC, created_at ASC")
		}
		return tx.Where("status = ?", entity.DialectCardStatusActive).Order("sort_order ASC, created_at ASC")
	}).Order("sort_order ASC, created_at ASC").Find(&groups).Error
	return groups, err
}

// CreateCardGroup 创建卡片分组
func (r *DialectRepositoryImpl) CreateCardGroup(ctx context.Context, group *entity.DialectCardGroup) error {
	return r.db.WithContext(ctx).Create(group).Error
}

// UpdateCardGroup 更新卡片分组
func (r *DialectRepositoryImpl) UpdateCardGroup(ctx context.Context, group *entity.DialectCardGroup) error {
	return r.db.WithContext(ctx).Save(group).Error
}

// DeleteCardGroup 删除卡片分组
func (r *DialectRepositoryImpl) DeleteCardGroup(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.DialectCardGroup{}, "id = ?", id).Error
}

// FindCardGroupByID 根据ID获取卡片分组
func (r *DialectRepositoryImpl) FindCardGroupByID(ctx context.Context, id string) (*entity.DialectCardGroup, error) {
	var group entity.DialectCardGroup
	err := r.db.WithContext(ctx).Preload("Cards", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("sort_order ASC, created_at ASC")
	}).First(&group, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// ListCards 获取卡片列表
func (r *DialectRepositoryImpl) ListCards(ctx context.Context, query *repository.DialectCardQuery) ([]entity.DialectCard, error) {
	var cards []entity.DialectCard
	db := r.db.WithContext(ctx).Model(&entity.DialectCard{})
	if query != nil {
		if strings.TrimSpace(query.GroupID) != "" {
			db = db.Where("group_id = ?", strings.TrimSpace(query.GroupID))
		}
		if !query.IncludeInactive {
			db = db.Where("status = ?", entity.DialectCardStatusActive)
		}
		if query.IncludeGroupInfo {
			db = db.Preload("Group")
		}
	}
	err := db.Order("sort_order ASC, created_at ASC").Find(&cards).Error
	return cards, err
}

// CreateCard 创建卡片
func (r *DialectRepositoryImpl) CreateCard(ctx context.Context, card *entity.DialectCard) error {
	return r.db.WithContext(ctx).Create(card).Error
}

// UpdateCard 更新卡片
func (r *DialectRepositoryImpl) UpdateCard(ctx context.Context, card *entity.DialectCard) error {
	return r.db.WithContext(ctx).Save(card).Error
}

// DeleteCard 删除卡片
func (r *DialectRepositoryImpl) DeleteCard(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.DialectCard{}, "id = ?", id).Error
}

// FindCardByID 根据ID获取卡片
func (r *DialectRepositoryImpl) FindCardByID(ctx context.Context, id string) (*entity.DialectCard, error) {
	var card entity.DialectCard
	err := r.db.WithContext(ctx).Preload("Group").First(&card, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &card, nil
}
