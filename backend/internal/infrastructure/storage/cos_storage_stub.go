// Package storage COS存储存根（未启用COS构建标签时使用）
//go:build !cos
// +build !cos

package storage

import (
	"errors"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
)

// ErrCOSNotEnabled COS未启用
var ErrCOSNotEnabled = errors.New("COS storage not enabled. Build with -tags cos to enable")

// NewCOSStorage 创建COS存储（存根实现）
func NewCOSStorage(cfg *config.StorageConfig) (*LocalStorage, error) {
	return nil, ErrCOSNotEnabled
}
