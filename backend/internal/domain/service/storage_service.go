package service

import (
	"context"
	"io"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
)

// StorageService 存储服务接口
type StorageService interface {
	// Upload 上传文件
	Upload(ctx context.Context, reader io.Reader, filename string, size int64, contentType string) (*entity.File, error)

	// Download 下载文件
	Download(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete 删除文件
	Delete(ctx context.Context, path string) error

	// GetURL 获取文件访问URL
	GetURL(ctx context.Context, path string) string

	// Exists 检查文件是否存在
	Exists(ctx context.Context, path string) (bool, error)

	// GetType 获取存储类型
	GetType() entity.StorageType
}

// UploadResult 上传结果
type UploadResult struct {
	Path        string
	URL         string
	Size        int64
	ContentType string
	StorageType entity.StorageType
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type         string
	LocalPath    string
	BaseURL      string
	MaxSize      int64
	AllowedTypes []string
}
