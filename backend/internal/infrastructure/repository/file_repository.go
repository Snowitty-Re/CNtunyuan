package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"gorm.io/gorm"
)

// FileRepositoryImpl 文件仓储实现
type FileRepositoryImpl struct {
	*BaseRepository[entity.File]
}

// NewFileRepository 创建文件仓储
func NewFileRepository(db *gorm.DB) repository.FileRepository {
	return &FileRepositoryImpl{
		BaseRepository: NewBaseRepository[entity.File](db),
	}
}

// Create 创建文件记录（处理空 UUID 字段，避免写入空字符串导致 PG 报错）
func (r *FileRepositoryImpl) Create(ctx context.Context, file *entity.File) error {
	db := r.db.WithContext(ctx)
	if strings.TrimSpace(file.UploaderID) == "" {
		db = db.Omit("uploader_id")
	}
	if strings.TrimSpace(file.EntityID) == "" {
		db = db.Omit("entity_id")
	}
	return db.Create(file).Error
}

// FindByUploader 根据上传者查找
func (r *FileRepositoryImpl) FindByUploader(ctx context.Context, uploaderID string, pagination repository.Pagination) (*repository.PageResult[entity.File], error) {
	var files []entity.File
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.File{}).Where("uploader_id = ?", uploaderID)

	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.Paginate(db.Order("created_at DESC"), pagination).Find(&files).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(files, total, pagination.Page, pagination.PageSize), nil
}

// FindByType 根据类型查找
func (r *FileRepositoryImpl) FindByType(ctx context.Context, fileType entity.FileType, pagination repository.Pagination) (*repository.PageResult[entity.File], error) {
	var files []entity.File
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.File{}).Where("file_type = ?", fileType)

	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.Paginate(db.Order("created_at DESC"), pagination).Find(&files).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(files, total, pagination.Page, pagination.PageSize), nil
}

// FindByEntity 根据关联实体查找
func (r *FileRepositoryImpl) FindByEntity(ctx context.Context, entityType string, entityID string) ([]entity.File, error) {
	var files []entity.File
	err := r.db.WithContext(ctx).
		Model(&entity.File{}).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("created_at DESC").
		Find(&files).Error
	return files, err
}

// FindByURLOrPath 根据URL或存储路径查找
func (r *FileRepositoryImpl) FindByURLOrPath(ctx context.Context, fileURL string, filePath string) (*entity.File, error) {
	fileURL = strings.TrimSpace(fileURL)
	filePath = strings.TrimSpace(filePath)
	if fileURL == "" && filePath == "" {
		return nil, gorm.ErrRecordNotFound
	}

	var file entity.File
	db := r.db.WithContext(ctx).Model(&entity.File{})
	switch {
	case fileURL != "" && filePath != "":
		db = db.Where("url = ? OR path = ?", fileURL, filePath)
	case fileURL != "":
		db = db.Where("url = ?", fileURL)
	default:
		db = db.Where("path = ?", filePath)
	}

	err := db.First(&file).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &file, nil
}

// FindByStorageType 根据存储类型查找
func (r *FileRepositoryImpl) FindByStorageType(ctx context.Context, storageType entity.StorageType, pagination repository.Pagination) (*repository.PageResult[entity.File], error) {
	var files []entity.File
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.File{}).Where("storage_type = ?", storageType)

	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.Paginate(db.Order("created_at DESC"), pagination).Find(&files).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(files, total, pagination.Page, pagination.PageSize), nil
}

// Search 搜索文件名
func (r *FileRepositoryImpl) Search(ctx context.Context, keyword string, pagination repository.Pagination) (*repository.PageResult[entity.File], error) {
	var files []entity.File
	var total int64

	db := r.db.WithContext(ctx).
		Model(&entity.File{}).
		Where("(file_name LIKE ? OR original_name LIKE ?)",
			"%"+keyword+"%", "%"+keyword+"%")

	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.Paginate(db.Order("created_at DESC"), pagination).Find(&files).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(files, total, pagination.Page, pagination.PageSize), nil
}

// UpdateEntity 更新关联实体
func (r *FileRepositoryImpl) UpdateEntity(ctx context.Context, fileID string, entityType string, entityID string) error {
	updates := map[string]interface{}{
		"entity_type": entityType,
		"entity_id":   entityID,
	}
	if strings.TrimSpace(entityID) == "" {
		updates["entity_id"] = nil
		updates["entity_type"] = ""
	}

	return r.db.WithContext(ctx).
		Model(&entity.File{}).
		Where("id = ?", fileID).
		Updates(updates).Error
}

// GetStats 获取统计
func (r *FileRepositoryImpl) GetStats(ctx context.Context) (*entity.FileStats, error) {
	stats := &entity.FileStats{}

	// 总数
	if err := r.db.WithContext(ctx).Model(&entity.File{}).
		Count(&stats.TotalCount).Error; err != nil {
		return nil, err
	}

	// 总大小
	if err := r.db.WithContext(ctx).Model(&entity.File{}).
		Select("COALESCE(SUM(size), 0)").Scan(&stats.TotalSize).Error; err != nil {
		return nil, err
	}

	// 图片统计
	if err := r.db.WithContext(ctx).Model(&entity.File{}).
		Where("file_type = ?", entity.FileTypeImage).
		Count(&stats.ImageCount).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Model(&entity.File{}).
		Where("file_type = ?", entity.FileTypeImage).
		Select("COALESCE(SUM(size), 0)").Scan(&stats.ImageSize).Error; err != nil {
		return nil, err
	}

	// 音频统计
	if err := r.db.WithContext(ctx).Model(&entity.File{}).
		Where("file_type = ?", entity.FileTypeAudio).
		Count(&stats.AudioCount).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Model(&entity.File{}).
		Where("file_type = ?", entity.FileTypeAudio).
		Select("COALESCE(SUM(size), 0)").Scan(&stats.AudioSize).Error; err != nil {
		return nil, err
	}

	// 视频统计
	if err := r.db.WithContext(ctx).Model(&entity.File{}).
		Where("file_type = ?", entity.FileTypeVideo).
		Count(&stats.VideoCount).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Model(&entity.File{}).
		Where("file_type = ?", entity.FileTypeVideo).
		Select("COALESCE(SUM(size), 0)").Scan(&stats.VideoSize).Error; err != nil {
		return nil, err
	}

	// 文档统计
	if err := r.db.WithContext(ctx).Model(&entity.File{}).
		Where("file_type = ?", entity.FileTypeDocument).
		Count(&stats.DocCount).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Model(&entity.File{}).
		Where("file_type = ?", entity.FileTypeDocument).
		Select("COALESCE(SUM(size), 0)").Scan(&stats.DocSize).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// SoftDelete 软删除
func (r *FileRepositoryImpl) SoftDelete(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entity.File{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now).
		Error
}

// GetTotalSize 获取总文件大小
func (r *FileRepositoryImpl) GetTotalSize(ctx context.Context) (int64, error) {
	var size int64
	err := r.db.WithContext(ctx).Model(&entity.File{}).
		Select("COALESCE(SUM(size), 0)").Scan(&size).Error
	return size, err
}

// CountByType 按类型统计
func (r *FileRepositoryImpl) CountByType(ctx context.Context, fileType entity.FileType) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.File{}).
		Where("file_type = ?", fileType).
		Count(&count).Error
	return count, err
}
