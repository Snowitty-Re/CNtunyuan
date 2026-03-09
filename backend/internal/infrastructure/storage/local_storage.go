package storage

import (
	"context"
	cryptoRand "crypto/rand"
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
	logger.Info("LocalStorage Upload called",
		logger.String("filename", filename),
		logger.Int64("size", size),
		logger.String("content_type", contentType),
		logger.String("base_path", s.basePath),
		logger.String("base_url", s.baseURL),
	)

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

	logger.Info("Generated paths",
		logger.String("storage_path", storagePath),
		logger.String("full_path", fullPath),
	)

	// 创建目录
	dir := filepath.Dir(fullPath)
	logger.Info("Creating directory", logger.String("dir", dir))
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Error("Failed to create directory", logger.String("dir", dir), logger.Err(err))
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// 创建文件
	logger.Info("Creating file", logger.String("path", fullPath))
	file, err := os.Create(fullPath)
	if err != nil {
		logger.Error("Failed to create file", logger.String("path", fullPath), logger.Err(err))
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// 写入文件
	written, err := io.Copy(file, reader)
	if err != nil {
		os.Remove(fullPath)
		logger.Error("Failed to write file", logger.String("path", fullPath), logger.Err(err))
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

	logger.Info("File uploaded to local storage successfully",
		logger.String("path", storagePath),
		logger.Int64("size", written),
		logger.String("url", f.URL),
	)

	return f, nil
}

// safePath 校验路径安全性，防止路径遍历攻击
func (s *LocalStorage) safePath(path string) (string, error) {
	// 清理路径中的 .. 等
	cleanPath := filepath.Clean(path)
	fullPath := filepath.Join(s.basePath, cleanPath)
	// 确保最终路径在 basePath 目录内
	absBase, err := filepath.Abs(s.basePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base path: %w", err)
	}
	absFull, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve file path: %w", err)
	}
	if !strings.HasPrefix(absFull, absBase+string(filepath.Separator)) && absFull != absBase {
		return "", fmt.Errorf("path traversal detected: %s", path)
	}
	return absFull, nil
}

// Download 下载文件
func (s *LocalStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath, err := s.safePath(path)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

// Delete 删除文件
func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath, err := s.safePath(path)
	if err != nil {
		return err
	}
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
	fullPath, err := s.safePath(path)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(fullPath)
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

// getFileType 根据扩展名获取文件类型
func getFileType(filename string) entity.FileType {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg":
		return entity.FileTypeImage
	case ".mp3", ".wav", ".ogg", ".m4a", ".flac":
		return entity.FileTypeAudio
	case ".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv":
		return entity.FileTypeVideo
	case ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt":
		return entity.FileTypeDocument
	default:
		return entity.FileTypeDocument
	}
}

// generateObjectKey 生成对象存储键
func generateObjectKey(filename string) string {
	now := time.Now()
	ext := filepath.Ext(filename)
	return fmt.Sprintf("uploads/%04d/%02d/%02d/%s%s",
		now.Year(), now.Month(), now.Day(),
		generateRandomString(16),
		ext,
	)
}

// generateRandomString 生成加密安全的随机字符串
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	randomBytes := make([]byte, length)
	if _, err := cryptoRand.Read(randomBytes); err != nil {
		// 降级使用 UUID
		return uuid.New().String()[:length]
	}
	for i := range result {
		result[i] = charset[int(randomBytes[i])%len(charset)]
	}
	return string(result)
}
