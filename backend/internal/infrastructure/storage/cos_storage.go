// Package storage 腾讯云COS存储实现
// 注意：使用此文件需要安装依赖：go get github.com/tencentyun/cos-go-sdk-v5
//go:build cos
// +build cos

package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/tencentyun/cos-go-sdk-v5"
)

// COSStorage 腾讯云COS存储
type COSStorage struct {
	client     *cos.Client
	bucketName string
	region     string
	baseURL    string
}

// NewCOSStorage 创建COS存储
func NewCOSStorage(cfg *config.StorageConfig) (*COSStorage, error) {
	bucketURL, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cfg.COSBucket, cfg.COSRegion))
	if err != nil {
		return nil, fmt.Errorf("failed to parse COS bucket URL: %w", err)
	}

	baseURL := &cos.BaseURL{BucketURL: bucketURL}
	client := cos.NewClient(baseURL, &http.Client{
		Timeout: 100 * time.Second,
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.COSSecretID,
			SecretKey: cfg.COSSecretKey,
		},
	})

	return &COSStorage{
		client:     client,
		bucketName: cfg.COSBucket,
		region:     cfg.COSRegion,
		baseURL:    cfg.BaseURL,
	}, nil
}

// Upload 上传文件
func (s *COSStorage) Upload(ctx context.Context, reader io.Reader, filename string, size int64, contentType string) (*entity.File, error) {
	objectKey := generateObjectKey(filename)

	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	}

	_, err := s.client.Object.Put(ctx, objectKey, reader, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to COS: %w", err)
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
		StorageType:  entity.StorageTypeCOS,
	}

	return file, nil
}

// Download 下载文件
func (s *COSStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	resp, err := s.client.Object.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to download from COS: %w", err)
	}
	return resp.Body, nil
}

// Delete 删除文件
func (s *COSStorage) Delete(ctx context.Context, path string) error {
	_, err := s.client.Object.Delete(ctx, path)
	return err
}

// Exists 检查文件是否存在
func (s *COSStorage) Exists(ctx context.Context, path string) (bool, error) {
	return s.client.Object.IsExist(ctx, path)
}

// GetURL 获取文件URL
func (s *COSStorage) GetURL(ctx context.Context, path string, expire time.Duration) (string, error) {
	if s.baseURL != "" {
		return s.baseURL + "/" + path, nil
	}
	return s.generateURL(path), nil
}

func (s *COSStorage) generateURL(objectKey string) string {
	return fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", s.bucketName, s.region, objectKey)
}
