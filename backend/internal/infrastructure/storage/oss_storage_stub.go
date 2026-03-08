// Package storage OSS存储存根（未启用OSS构建标签时使用）
//go:build !oss
// +build !oss

package storage

import (
	"errors"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
)

// ErrOSSNotEnabled OSS未启用
var ErrOSSNotEnabled = errors.New("OSS storage not enabled. Build with -tags oss to enable")

// NewOSSStorage 创建OSS存储（存根实现）
func NewOSSStorage(cfg *config.StorageConfig) (*LocalStorage, error) {
	return nil, ErrOSSNotEnabled
}
