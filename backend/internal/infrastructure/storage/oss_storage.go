// Package storage 阿里云OSS存储实现
// 注意：使用此文件需要安装依赖：go get github.com/aliyun/aliyun-oss-go-sdk/oss
//go:build oss
// +build oss

package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// OSSStorage 阿里云OSS存储
type OSSStorage struct {
	client     *oss.Client
	bucket     *oss.Bucket
	bucketName string
	endpoint   string
	baseURL    string
}

// NewOSSStorage 创建OSS存储
func NewOSSStorage(cfg *config.StorageConfig) (*OSSStorage, error) {
	// 创建OSS客户端
	client, err := oss.New(
		cfg.OSSEndpoint,
		cfg.OSSAccessKeyID,
		cfg.OSSAccessKeySecret,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OSS client: %w", err)
	}

	// 获取存储空间
	bucket, err := client.Bucket(cfg.OSSBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to get OSS bucket: %w", err)
	}

	return &OSSStorage{
		client:     client,
		bucket:     bucket,
		bucketName: cfg.OSSBucket,
		endpoint:   cfg.OSSEndpoint,
		baseURL:    cfg.BaseURL,
	}, nil
}

// Upload 上传文件
func (s *OSSStorage) Upload(ctx context.Context, reader io.Reader, filename string, size int64, contentType string) (*entity.File, error) {
	objectKey := generateObjectKey(filename)

	options := []oss.Option{
		oss.ContentType(contentType),
	}

	err := s.bucket.PutObject(objectKey, reader, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to OSS: %w", err)
	}

	fileURL := s.generateURL(objectKey)

	file := &entity.File{
		FileName:     filename,
		OriginalName: filename,
		FileType:     getFileType(filename),
		MimeType:     contentType,
		Size:         size,
		Path:         objectKey,
		URL:          fileURL,
		StorageType:  entity.StorageTypeOSS,
	}

	return file, nil
}

// Download 下载文件
func (s *OSSStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	body, err := s.bucket.GetObject(path)
	if err != nil {
		return nil, fmt.Errorf("failed to download from OSS: %w", err)
	}
	return body, nil
}

// Delete 删除文件
func (s *OSSStorage) Delete(ctx context.Context, path string) error {
	return s.bucket.DeleteObject(path)
}

// Exists 检查文件是否存在
func (s *OSSStorage) Exists(ctx context.Context, path string) (bool, error) {
	return s.bucket.IsObjectExist(path)
}

// GetURL 获取文件URL
func (s *OSSStorage) GetURL(ctx context.Context, path string, expire time.Duration) (string, error) {
	if s.baseURL != "" {
		return s.baseURL + "/" + path, nil
	}
	return s.generateURL(path), nil
}

func (s *OSSStorage) generateURL(objectKey string) string {
	return fmt.Sprintf("https://%s.%s/%s", s.bucketName, s.endpoint, objectKey)
}
