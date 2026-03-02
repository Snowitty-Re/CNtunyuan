package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/service"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/google/uuid"
)

// LocalStorage 本地存储
type LocalStorage struct {
	basePath     string
	baseURL      string
	maxSize      int64
	allowedTypes []string
}

// NewLocalStorage 创建本地存储
func NewLocalStorage(cfg *config.StorageConfig) service.StorageService {
	return &LocalStorage{
		basePath:     cfg.LocalPath,
		baseURL:      cfg.BaseURL,
		maxSize:      cfg.MaxFileSize,
		allowedTypes: strings.Split(cfg.AllowedTypes, ","),
	}
}

// Upload 上传文件
func (s *LocalStorage) Upload(ctx context.Context, reader io.Reader, filename string, size int64, contentType string) (*entity.File, error) {
	// 检查文件大小
	if s.maxSize > 0 && size > s.maxSize {
		return nil, fmt.Errorf("file size exceeds limit: %d", s.maxSize)
	}

	// 生成存储路径
	now := time.Now()
	datePath := now.Format("2006/01/02")
	fileID := uuid.New().String()
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".bin"
	}

	storagePath := filepath.Join(datePath, fileID+ext)
	fullPath := filepath.Join(s.basePath, storagePath)

	// 创建目录
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// 创建文件
	file, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// 写入文件
	written, err := io.Copy(file, reader)
	if err != nil {
		os.Remove(fullPath)
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// 构建文件实体
	f := &entity.File{
		FileName:     fileID + ext,
		OriginalName: filename,
		FileType:     entity.DetectFileType(contentType),
		MimeType:     contentType,
		Size:         written,
		Path:         storagePath,
		URL:          s.baseURL + "/" + strings.ReplaceAll(storagePath, "\\", "/"),
		StorageType:  entity.StorageTypeLocal,
	}

	logger.Info("File uploaded to local storage",
		logger.String("path", storagePath),
		logger.Int64("size", written),
	)

	return f, nil
}

// Download 下载文件
func (s *LocalStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, path)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

// Delete 删除文件
func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(s.basePath, path)
	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在，视为删除成功
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	logger.Info("File deleted from local storage", logger.String("path", path))
	return nil
}

// GetURL 获取文件访问URL
func (s *LocalStorage) GetURL(ctx context.Context, path string) string {
	return s.baseURL + "/" + strings.ReplaceAll(path, "\\", "/")
}

// Exists 检查文件是否存在
func (s *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	fullPath := filepath.Join(s.basePath, path)
	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// GetType 获取存储类型
func (s *LocalStorage) GetType() entity.StorageType {
	return entity.StorageTypeLocal
}
