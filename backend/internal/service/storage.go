package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/google/uuid"
	"github.com/tencentyun/cos-go-sdk-v5"
)

// StorageService 文件存储服务
type StorageService struct {
	config *config.StorageConfig
}

// FileInfo 文件信息
type FileInfo struct {
	ID         string    `json:"id"`
	FileName   string    `json:"file_name"`
	FileType   string    `json:"file_type"`
	FileSize   int64     `json:"file_size"`
	URL        string    `json:"url"`
	UploadTime time.Time `json:"upload_time"`
}

// NewStorageService 创建存储服务
func NewStorageService(cfg *config.StorageConfig) *StorageService {
	// 确保本地存储目录存在
	if cfg.Type == "local" {
		os.MkdirAll(cfg.LocalPath, 0755)
	}
	return &StorageService{
		config: cfg,
	}
}

// UploadFile 上传文件
func (s *StorageService) UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, fileType string) (*FileInfo, error) {
	// 检查文件大小
	if s.config.MaxFileSize > 0 && fileHeader.Size > s.config.MaxFileSize {
		return nil, fmt.Errorf("文件大小超过限制，最大允许 %d MB", s.config.MaxFileSize/1024/1024)
	}

	// 检查文件类型
	if !s.isAllowedFileType(fileHeader.Filename) {
		return nil, fmt.Errorf("不支持的文件类型")
	}

	// 打开上传的文件
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 生成唯一文件名
	ext := filepath.Ext(fileHeader.Filename)
	if ext == "" {
		ext = ".bin"
	}
	fileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// 按日期组织目录
	now := time.Now()
	datePath := now.Format("2006/01/02")
	if fileType != "" {
		datePath = fmt.Sprintf("%s/%s", fileType, datePath)
	}

	var fileURL string

	switch s.config.Type {
	case "oss":
		fileURL, err = s.uploadToOSS(ctx, file, datePath, fileName)
	case "cos":
		fileURL, err = s.uploadToCOS(ctx, file, datePath, fileName)
	default:
		fileURL, err = s.uploadToLocal(ctx, file, datePath, fileName)
	}

	if err != nil {
		return nil, err
	}

	return &FileInfo{
		ID:         uuid.New().String(),
		FileName:   fileHeader.Filename,
		FileType:   fileHeader.Header.Get("Content-Type"),
		FileSize:   fileHeader.Size,
		URL:        fileURL,
		UploadTime: now,
	}, nil
}

// UploadMultipleFiles 上传多个文件
func (s *StorageService) UploadMultipleFiles(ctx context.Context, fileHeaders []*multipart.FileHeader, fileType string) ([]*FileInfo, error) {
	results := make([]*FileInfo, 0, len(fileHeaders))

	for _, fileHeader := range fileHeaders {
		fileInfo, err := s.UploadFile(ctx, fileHeader, fileType)
		if err != nil {
			return nil, err
		}
		results = append(results, fileInfo)
	}

	return results, nil
}

// uploadToLocal 上传到本地
func (s *StorageService) uploadToLocal(ctx context.Context, file io.Reader, datePath, fileName string) (string, error) {
	// 创建目录
	dir := filepath.Join(s.config.LocalPath, datePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	// 保存文件
	filePath := filepath.Join(dir, fileName)
	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		return "", fmt.Errorf("保存文件失败: %w", err)
	}

	// 返回URL
	relativePath := filepath.Join(datePath, fileName)
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.config.BaseURL, "/"), strings.ReplaceAll(relativePath, "\\", "/")), nil
}

// uploadToOSS 上传到阿里云OSS
func (s *StorageService) uploadToOSS(ctx context.Context, file io.Reader, datePath, fileName string) (string, error) {
	client, err := oss.New(s.config.OSSEndpoint, s.config.OSSAccessKeyID, s.config.OSSAccessKeySecret)
	if err != nil {
		return "", fmt.Errorf("创建OSS客户端失败: %w", err)
	}

	bucket, err := client.Bucket(s.config.OSSBucket)
	if err != nil {
		return "", fmt.Errorf("获取OSS Bucket失败: %w", err)
	}

	objectKey := fmt.Sprintf("%s/%s", datePath, fileName)
	err = bucket.PutObject(objectKey, file)
	if err != nil {
		return "", fmt.Errorf("上传OSS失败: %w", err)
	}

	return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.config.BaseURL, "/"), objectKey), nil
}

// uploadToCOS 上传到腾讯云COS
func (s *StorageService) uploadToCOS(ctx context.Context, file io.Reader, datePath, fileName string) (string, error) {
	u, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", s.config.COSBucket, s.config.COSRegion))
	if err != nil {
		return "", fmt.Errorf("解析COS地址失败: %w", err)
	}

	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  s.config.COSSecretID,
			SecretKey: s.config.COSSecretKey,
		},
	})

	objectKey := fmt.Sprintf("%s/%s", datePath, fileName)
	_, err = client.Object.Put(ctx, objectKey, file, nil)
	if err != nil {
		return "", fmt.Errorf("上传COS失败: %w", err)
	}

	return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.config.BaseURL, "/"), objectKey), nil
}

// DeleteFile 删除文件
func (s *StorageService) DeleteFile(ctx context.Context, fileURL string) error {
	switch s.config.Type {
	case "oss":
		return s.deleteFromOSS(ctx, fileURL)
	case "cos":
		return s.deleteFromCOS(ctx, fileURL)
	default:
		return s.deleteFromLocal(ctx, fileURL)
	}
}

// deleteFromLocal 从本地删除
func (s *StorageService) deleteFromLocal(ctx context.Context, fileURL string) error {
	// 从URL中提取相对路径
	prefix := strings.TrimSuffix(s.config.BaseURL, "/") + "/"
	relativePath := strings.TrimPrefix(fileURL, prefix)

	filePath := filepath.Join(s.config.LocalPath, relativePath)
	err := os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除文件失败: %w", err)
	}
	return nil
}

// deleteFromOSS 从OSS删除
func (s *StorageService) deleteFromOSS(ctx context.Context, fileURL string) error {
	client, err := oss.New(s.config.OSSEndpoint, s.config.OSSAccessKeyID, s.config.OSSAccessKeySecret)
	if err != nil {
		return fmt.Errorf("创建OSS客户端失败: %w", err)
	}

	bucket, err := client.Bucket(s.config.OSSBucket)
	if err != nil {
		return fmt.Errorf("获取OSS Bucket失败: %w", err)
	}

	// 从URL中提取ObjectKey
	prefix := strings.TrimSuffix(s.config.BaseURL, "/") + "/"
	objectKey := strings.TrimPrefix(fileURL, prefix)

	err = bucket.DeleteObject(objectKey)
	if err != nil {
		return fmt.Errorf("删除OSS文件失败: %w", err)
	}
	return nil
}

// deleteFromCOS 从COS删除
func (s *StorageService) deleteFromCOS(ctx context.Context, fileURL string) error {
	u, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", s.config.COSBucket, s.config.COSRegion))
	if err != nil {
		return fmt.Errorf("解析COS地址失败: %w", err)
	}

	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  s.config.COSSecretID,
			SecretKey: s.config.COSSecretKey,
		},
	})

	// 从URL中提取ObjectKey
	prefix := strings.TrimSuffix(s.config.BaseURL, "/") + "/"
	objectKey := strings.TrimPrefix(fileURL, prefix)

	_, err = client.Object.Delete(ctx, objectKey)
	if err != nil {
		return fmt.Errorf("删除COS文件失败: %w", err)
	}
	return nil
}

// GetFilePath 获取本地文件路径
func (s *StorageService) GetFilePath(fileURL string) string {
	if s.config.Type != "local" {
		return ""
	}
	prefix := strings.TrimSuffix(s.config.BaseURL, "/") + "/"
	relativePath := strings.TrimPrefix(fileURL, prefix)
	return filepath.Join(s.config.LocalPath, relativePath)
}

// isAllowedFileType 检查文件类型是否允许
func (s *StorageService) isAllowedFileType(filename string) bool {
	if s.config.AllowedTypes == "" {
		return true
	}

	ext := strings.ToLower(filepath.Ext(filename))
	ext = strings.TrimPrefix(ext, ".")

	allowedTypes := strings.Split(s.config.AllowedTypes, ",")
	for _, t := range allowedTypes {
		if strings.TrimSpace(t) == ext {
			return true
		}
	}

	return false
}

// GetConfig 获取存储配置
func (s *StorageService) GetConfig() *config.StorageConfig {
	return s.config
}
