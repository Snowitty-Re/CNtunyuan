package repository

import (
	"context"

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

// FindByUploader 根据上传者查找
func (r *FileRepositoryImpl) FindByUploader(ctx context.Context, uploaderID string, pagination repository.Pagination) (*repository.PageResult[entity.File], error) {
	var files []entity.File
	var total int64

	db := r.db.WithContext(ctx).Where("uploader_id = ? AND is_deleted = ?", uploaderID, false)

	if err := db.Model(&entity.File{}).Count(&total).Error; err != nil {
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

	db := r.db.WithContext(ctx).Where("file_type = ? AND is_deleted = ?", fileType, false)

	if err := db.Model(&entity.File{}).Count(&total).Error; err != nil {
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
		Where("entity_type = ? AND entity_id = ? AND is_deleted = ?", entityType, entityID, false).
		Order("created_at DESC").
		Find(&files).Error
	return files, err
}

// FindByStorageType 根据存储类型查找
func (r *FileRepositoryImpl) FindByStorageType(ctx context.Context, storageType entity.StorageType, pagination repository.Pagination) (*repository.PageResult[entity.File], error) {
	var files []entity.File
	var total int64

	db := r.db.WithContext(ctx).Where("storage_type = ? AND is_deleted = ?", storageType, false)

	if err := db.Model(&entity.File{}).Count(&total).Error; err != nil {
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
		Where("(file_name LIKE ? OR original_name LIKE ?) AND is_deleted = ?",
			"%"+keyword+"%", "%"+keyword+"%", false)

	if err := db.Model(&entity.File{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.Paginate(db.Order("created_at DESC"), pagination).Find(&files).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(files, total, pagination.Page, pagination.PageSize), nil
}

// UpdateEntity 更新关联实体
func (r *FileRepositoryImpl) UpdateEntity(ctx context.Context, fileID string, entityType string, entityID string) error {
	return r.db.WithContext(ctx).
		Model(&entity.File{}).
		Where("id = ?", fileID).
		Updates(map[string]interface{}{
			"entity_type": entityType,
			"entity_id":   entityID,
		}).Error
}

// GetStats 获取统计
func (r *FileRepositoryImpl) GetStats(ctx context.Context) (*entity.FileStats, error) {
	stats := &entity.FileStats{}

	// 总数
	if err := r.db.WithContext(ctx).Model(&entity.File{}).
		Where("is_deleted = ?", false).
		Count(&stats.TotalCount).Error; err != nil {
		return nil, err
	}

	// 总大小
	if err := r.db.WithContext(ctx).Model(&entity.File{}).
		Where("is_deleted = ?", false).
		Select("COALESCE(SUM(size), 0)").Scan(&stats.TotalSize).Error; err != nil {
		return nil, err
	}

	// 图片统计
	if err := r.db.WithContext(ctx).Model(&entity.File{}).
		Where("file_type = ? AND is_deleted = ?", entity.FileTypeImage, false).
		Count(&stats.ImageCount).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Model(&entity.File{}).
		Where("file_type = ? AND is_deleted = ?", entity.FileTypeImage, false).
		Select("COALESCE(SUM(size), 0)").Scan(&stats.ImageSize).Error; err != nil {
		return nil, err
	}

	// 音频统计
	if err := r.db.WithContext(ctx).Model(&entity.File{}).
		Where("file_type = ? AND is_deleted = ?", entity.FileTypeAudio, false).
		Count(&stats.AudioCount).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Model(&entity.File{}).
		Where("file_type = ? AND is_deleted = ?", entity.FileTypeAudio, false).
		Select("COALESCE(SUM(size), 0)").Scan(&stats.AudioSize).Error; err != nil {
		return nil, err
	}

	// 视频统计
	if err := r.db.WithContext(ctx).Model(&entity.File{}).
		Where("file_type = ? AND is_deleted = ?", entity.FileTypeVideo, false).
		Count(&stats.VideoCount).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Model(&entity.File{}).
		Where("file_type = ? AND is_deleted = ?", entity.FileTypeVideo, false).
		Select("COALESCE(SUM(size), 0)").Scan(&stats.VideoSize).Error; err != nil {
		return nil, err
	}

	// 文档统计
	if err := r.db.WithContext(ctx).Model(&entity.File{}).
		Where("file_type = ? AND is_deleted = ?", entity.FileTypeDocument, false).
		Count(&stats.DocCount).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Model(&entity.File{}).
		Where("file_type = ? AND is_deleted = ?", entity.FileTypeDocument, false).
		Select("COALESCE(SUM(size), 0)").Scan(&stats.DocSize).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// SoftDelete 软删除
func (r *FileRepositoryImpl) SoftDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&entity.File{}).
		Where("id = ?", id).
		Update("is_deleted", true).
		Error
}

// GetTotalSize 获取总文件大小
func (r *FileRepositoryImpl) GetTotalSize(ctx context.Context) (int64, error) {
	var size int64
	err := r.db.WithContext(ctx).Model(&entity.File{}).
		Where("is_deleted = ?", false).
		Select("COALESCE(SUM(size), 0)").Scan(&size).Error
	return size, err
}

// CountByType 按类型统计
func (r *FileRepositoryImpl) CountByType(ctx context.Context, fileType entity.FileType) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.File{}).
		Where("file_type = ? AND is_deleted = ?", fileType, false).
		Count(&count).Error
	return count, err
}
