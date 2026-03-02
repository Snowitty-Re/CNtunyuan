package repository

import (
	"context"
	"errors"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"gorm.io/gorm"
)

// BaseRepository 基础仓储实现
type BaseRepository[T any] struct {
	db *gorm.DB
}

// NewBaseRepository 创建基础仓储
func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: db}
}

// DB 获取数据库连接
func (r *BaseRepository[T]) DB() *gorm.DB {
	return r.db
}

// SetDB 设置数据库连接
func (r *BaseRepository[T]) SetDB(db *gorm.DB) {
	r.db = db
}

// WithContext 使用上下文
func (r *BaseRepository[T]) WithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

// Create 创建
func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// Update 更新
func (r *BaseRepository[T]) Update(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// Delete 硬删除
func (r *BaseRepository[T]) Delete(ctx context.Context, id string) error {
	var entity T
	return r.db.WithContext(ctx).Delete(&entity, "id = ?", id).Error
}

// SoftDelete 软删除
func (r *BaseRepository[T]) SoftDelete(ctx context.Context, id string) error {
	var entity T
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity).Error
}

// FindByID 根据ID查找
func (r *BaseRepository[T]) FindByID(ctx context.Context, id string) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("记录不存在")
		}
		return nil, err
	}
	return &entity, nil
}

// FindAll 查找所有
func (r *BaseRepository[T]) FindAll(ctx context.Context) ([]T, error) {
	var entities []T
	err := r.db.WithContext(ctx).Find(&entities).Error
	return entities, err
}

// Count 统计
func (r *BaseRepository[T]) Count(ctx context.Context) (int64, error) {
	var count int64
	var entity T
	err := r.db.WithContext(ctx).Model(&entity).Count(&count).Error
	return count, err
}

// Exists 检查是否存在
func (r *BaseRepository[T]) Exists(ctx context.Context, id string) (bool, error) {
	var count int64
	var entity T
	err := r.db.WithContext(ctx).Model(&entity).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// Paginate 分页
func (r *BaseRepository[T]) Paginate(db *gorm.DB, pagination repository.Pagination) *gorm.DB {
	offset := (pagination.Page - 1) * pagination.PageSize
	return db.Offset(offset).Limit(pagination.PageSize)
}

// ApplyQueryOptions 应用查询选项
func (r *BaseRepository[T]) ApplyQueryOptions(db *gorm.DB, options *repository.QueryOptions) *gorm.DB {
	if options == nil {
		return db
	}

	if len(options.Select) > 0 {
		db = db.Select(options.Select)
	}

	if len(options.Omit) > 0 {
		db = db.Omit(options.Omit...)
	}

	if options.Order != "" {
		db = db.Order(options.Order)
	}

	for _, preload := range options.Preload {
		db = db.Preload(preload)
	}

	return db
}

// UnitOfWorkImpl 工作单元实现
type UnitOfWorkImpl struct {
	db     *gorm.DB
	tx     *gorm.DB
	isTx   bool
}

// NewUnitOfWork 创建工作单元
func NewUnitOfWork(db *gorm.DB) *UnitOfWorkImpl {
	return &UnitOfWorkImpl{db: db}
}

// Begin 开始事务
func (u *UnitOfWorkImpl) Begin() error {
	u.tx = u.db.Begin()
	u.isTx = true
	return u.tx.Error
}

// Commit 提交事务
func (u *UnitOfWorkImpl) Commit() error {
	if !u.isTx {
		return errors.New("没有活动的事务")
	}
	return u.tx.Commit().Error
}

// Rollback 回滚事务
func (u *UnitOfWorkImpl) Rollback() error {
	if !u.isTx {
		return errors.New("没有活动的事务")
	}
	return u.tx.Rollback().Error
}

// GetDB 获取数据库连接
func (u *UnitOfWorkImpl) GetDB() *gorm.DB {
	if u.isTx {
		return u.tx
	}
	return u.db
}
