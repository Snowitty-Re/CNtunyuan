package repository

import (
	"context"

	"gorm.io/gorm"
)

// Repository 基础仓储接口
type Repository[T any] interface {
	Create(ctx context.Context, entity *T) error
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id string) error
	SoftDelete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*T, error)
	FindAll(ctx context.Context) ([]T, error)
	Count(ctx context.Context) (int64, error)
	Exists(ctx context.Context, id string) (bool, error)
}

// Pagination 分页参数
type Pagination struct {
	Page     int
	PageSize int
}

// PageResult 分页结果
type PageResult[T any] struct {
	List       []T   `json:"list"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

// NewPageResult 创建分页结果
func NewPageResult[T any](list []T, total int64, page, pageSize int) *PageResult[T] {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	return &PageResult[T]{
		List:       list,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

// QueryOptions 查询选项
type QueryOptions struct {
	Select  []string
	Omit    []string
	Order   string
	Preload []string
}

// Specification 规格模式接口
type Specification interface {
	ToQuery(db *gorm.DB) *gorm.DB
}

// SpecificationFunc 规格函数
type SpecificationFunc func(*gorm.DB) *gorm.DB

// ToQuery 实现 Specification 接口
func (f SpecificationFunc) ToQuery(db *gorm.DB) *gorm.DB {
	return f(db)
}

// UnitOfWork 工作单元
type UnitOfWork interface {
	Begin() error
	Commit() error
	Rollback() error
	GetDB() *gorm.DB
}
