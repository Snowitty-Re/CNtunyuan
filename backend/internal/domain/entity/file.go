package entity

import (
	"errors"
	"path/filepath"
	"strings"
)

// FileType 文件类型
type FileType string

const (
	FileTypeImage    FileType = "image"
	FileTypeAudio    FileType = "audio"
	FileTypeVideo    FileType = "video"
	FileTypeDocument FileType = "document"
)

// StorageType 存储类型
type StorageType string

const (
	StorageTypeLocal StorageType = "local"
	StorageTypeOSS   StorageType = "oss"
	StorageTypeCOS   StorageType = "cos"
)

// File 文件领域实体
type File struct {
	BaseEntity
	FileName     string      `gorm:"size:255;not null" json:"file_name"`
	OriginalName string      `gorm:"size:255;not null" json:"original_name"`
	FileType     FileType    `gorm:"size:20;not null" json:"file_type"`
	MimeType     string      `gorm:"size:100" json:"mime_type"`
	Size         int64       `json:"size"`
	Path         string      `gorm:"size:500;not null" json:"path"`
	URL          string      `gorm:"size:500" json:"url"`
	StorageType  StorageType `gorm:"size:20;not null" json:"storage_type"`
	UploaderID   string      `gorm:"type:uuid;index" json:"uploader_id"`
	EntityType   string      `gorm:"size:50;index" json:"entity_type"` // 关联实体类型
	EntityID     string      `gorm:"type:uuid;index" json:"entity_id"` // 关联实体ID
	Description  string      `gorm:"type:text" json:"description"`
	IsDeleted    bool        `gorm:"default:false" json:"is_deleted"`
}

// TableName 表名
func (File) TableName() string {
	return "ty_files"
}

// Validate 验证文件
func (f *File) Validate() error {
	if f.FileName == "" {
		return errors.New("file name is required")
	}
	if f.Path == "" {
		return errors.New("file path is required")
	}
	if f.Size <= 0 {
		return errors.New("file size must be greater than 0")
	}
	return nil
}

// GetExtension 获取文件扩展名
func (f *File) GetExtension() string {
	return strings.ToLower(filepath.Ext(f.FileName))
}

// IsImage 是否为图片
func (f *File) IsImage() bool {
	return f.FileType == FileTypeImage
}

// IsAudio 是否为音频
func (f *File) IsAudio() bool {
	return f.FileType == FileTypeAudio
}

// IsVideo 是否为视频
func (f *File) IsVideo() bool {
	return f.FileType == FileTypeVideo
}

// IsDocument 是否为文档
func (f *File) IsDocument() bool {
	return f.FileType == FileTypeDocument
}

// MarkAsDeleted 标记为已删除
func (f *File) MarkAsDeleted() {
	f.IsDeleted = true
}

// FileStats 文件统计
type FileStats struct {
	TotalCount int64 `json:"total_count"`
	TotalSize  int64 `json:"total_size"`
	ImageCount int64 `json:"image_count"`
	ImageSize  int64 `json:"image_size"`
	AudioCount int64 `json:"audio_count"`
	AudioSize  int64 `json:"audio_size"`
	VideoCount int64 `json:"video_count"`
	VideoSize  int64 `json:"video_size"`
	DocCount   int64 `json:"doc_count"`
	DocSize    int64 `json:"doc_size"`
}

// DetectFileType 检测文件类型
func DetectFileType(mimeType string) FileType {
	mimeType = strings.ToLower(mimeType)
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return FileTypeImage
	case strings.HasPrefix(mimeType, "audio/"):
		return FileTypeAudio
	case strings.HasPrefix(mimeType, "video/"):
		return FileTypeVideo
	default:
		return FileTypeDocument
	}
}
