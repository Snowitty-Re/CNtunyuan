package repository

import (
	"context"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
)

// FileRepository 文件仓储接口
type FileRepository interface {
	Repository[entity.File]

	// FindByUploader 根据上传者查找
	FindByUploader(ctx context.Context, uploaderID string, pagination Pagination) (*PageResult[entity.File], error)

	// FindByType 根据类型查找
	FindByType(ctx context.Context, fileType entity.FileType, pagination Pagination) (*PageResult[entity.File], error)

	// FindByEntity 根据关联实体查找
	FindByEntity(ctx context.Context, entityType string, entityID string) ([]entity.File, error)

	// FindByStorageType 根据存储类型查找
	FindByStorageType(ctx context.Context, storageType entity.StorageType, pagination Pagination) (*PageResult[entity.File], error)

	// Search 搜索文件名
	Search(ctx context.Context, keyword string, pagination Pagination) (*PageResult[entity.File], error)

	// UpdateEntity 更新关联实体
	UpdateEntity(ctx context.Context, fileID string, entityType string, entityID string) error

	// GetStats 获取统计
	GetStats(ctx context.Context) (*entity.FileStats, error)

	// SoftDelete 软删除
	SoftDelete(ctx context.Context, id string) error

	// GetTotalSize 获取总文件大小
	GetTotalSize(ctx context.Context) (int64, error)

	// CountByType 按类型统计
	CountByType(ctx context.Context, fileType entity.FileType) (int64, error)
}

// FileQuery 文件查询参数
type FileQuery struct {
	Pagination
	Keyword      string             `json:"keyword"`
	FileType     entity.FileType    `json:"file_type"`
	UploaderID   string             `json:"uploader_id"`
	EntityType   string             `json:"entity_type"`
	EntityID     string             `json:"entity_id"`
	StorageType  entity.StorageType `json:"storage_type"`
	StartDate    string             `json:"start_date"`
	EndDate      string             `json:"end_date"`
}

// NewFileQuery 创建默认文件查询
func NewFileQuery() *FileQuery {
	return &FileQuery{
		Pagination: Pagination{
			Page:     1,
			PageSize: 10,
		},
	}
}
