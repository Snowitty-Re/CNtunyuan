// Package config 配置验证
package config

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error 实现 error 接口
func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ConfigValidator 配置验证器
type ConfigValidator struct {
	errors []ValidationError
}

// NewConfigValidator 创建配置验证器
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{
		errors: make([]ValidationError, 0),
	}
}

// Validate 验证配置
func (v *ConfigValidator) Validate(cfg *Config) []ValidationError {
	v.errors = make([]ValidationError, 0)

	// 验证服务器配置
	v.validateServer(cfg)

	// 验证数据库配置
	v.validateDatabase(cfg)

	// 验证JWT配置
	v.validateJWT(cfg)

	// 验证存储配置
	v.validateStorage(cfg)

	// 验证微信配置（如果启用）
	v.validateWeChat(cfg)

	return v.errors
}

// HasErrors 是否有错误
func (v *ConfigValidator) HasErrors() bool {
	return len(v.errors) > 0
}

// addError 添加错误
func (v *ConfigValidator) addError(field, message string) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// validateServer 验证服务器配置
func (v *ConfigValidator) validateServer(cfg *Config) {
	if cfg.Server.Port == "" {
		v.addError("server.port", "端口不能为空")
	}

	if cfg.Server.Mode != "debug" && cfg.Server.Mode != "release" {
		v.addError("server.mode", "模式必须是 debug 或 release")
	}
}

// validateDatabase 验证数据库配置
func (v *ConfigValidator) validateDatabase(cfg *Config) {
	if cfg.Database.Type != "postgres" && cfg.Database.Type != "mysql" {
		v.addError("database.type", "数据库类型必须是 postgres 或 mysql")
	}

	if cfg.Database.Host == "" {
		v.addError("database.host", "数据库主机不能为空")
	}

	if cfg.Database.Port <= 0 || cfg.Database.Port > 65535 {
		v.addError("database.port", "数据库端口无效")
	}

	if cfg.Database.User == "" {
		v.addError("database.user", "数据库用户不能为空")
	}

	if cfg.Database.Database == "" {
		v.addError("database.database", "数据库名称不能为空")
	}
}

// validateJWT 验证JWT配置
func (v *ConfigValidator) validateJWT(cfg *Config) {
	if cfg.JWT.Secret == "" {
		v.addError("jwt.secret", "JWT密钥不能为空")
	}

	if len(cfg.JWT.Secret) < 32 {
		v.addError("jwt.secret", "JWT密钥长度至少32位")
	}

	// 检查是否使用默认密钥（生产环境危险）
	weakSecrets := []string{
		"your-secret-key",
		"your-secret-key-here-change-in-production",
		"secret",
		"123456",
		"password",
	}

	for _, weak := range weakSecrets {
		if strings.Contains(strings.ToLower(cfg.JWT.Secret), weak) {
			v.addError("jwt.secret", "JWT密钥过于简单，请使用随机生成的强密钥")
			break
		}
	}

	if cfg.JWT.ExpireTime <= 0 {
		v.addError("jwt.expire_time", "JWT过期时间必须大于0")
	}
}

// validateStorage 验证存储配置
func (v *ConfigValidator) validateStorage(cfg *Config) {
	if cfg.Storage.Type != "local" && cfg.Storage.Type != "oss" && cfg.Storage.Type != "cos" {
		v.addError("storage.type", "存储类型必须是 local, oss 或 cos")
	}

	if cfg.Storage.MaxFileSize <= 0 {
		v.addError("storage.max_file_size", "最大文件大小必须大于0")
	}

	if cfg.Storage.MaxFileSize > 1024*1024*1024 {
		v.addError("storage.max_file_size", "最大文件大小不能超过1GB")
	}

	if cfg.Storage.Type == "local" {
		if strings.TrimSpace(cfg.Storage.LocalPath) == "" {
			v.addError("storage.local_path", "使用本地存储时 local_path 不能为空")
		} else if !filepath.IsAbs(cfg.Storage.LocalPath) {
			v.addError("storage.local_path", "使用本地存储时 local_path 必须是绝对路径")
		}

		if strings.TrimSpace(cfg.Storage.BaseURL) == "" {
			v.addError("storage.base_url", "使用本地存储时 base_url 不能为空")
		} else {
			u, err := url.Parse(cfg.Storage.BaseURL)
			if err != nil || u.Scheme == "" || u.Host == "" {
				v.addError("storage.base_url", "base_url 必须是合法的完整 URL，例如 http://localhost:8080/uploads")
			}
		}
	}

	// 验证OSS配置
	if cfg.Storage.Type == "oss" {
		if cfg.Storage.OSSAccessKeyID == "" {
			v.addError("storage.oss_access_key_id", "使用OSS时AccessKeyID不能为空")
		}
		if cfg.Storage.OSSAccessKeySecret == "" {
			v.addError("storage.oss_access_key_secret", "使用OSS时AccessKeySecret不能为空")
		}
		if cfg.Storage.OSSEndpoint == "" {
			v.addError("storage.oss_endpoint", "使用OSS时Endpoint不能为空")
		}
		if cfg.Storage.OSSBucket == "" {
			v.addError("storage.oss_bucket", "使用OSS时Bucket不能为空")
		}
	}

	// 验证COS配置
	if cfg.Storage.Type == "cos" {
		if cfg.Storage.COSSecretID == "" {
			v.addError("storage.cos_secret_id", "使用COS时SecretID不能为空")
		}
		if cfg.Storage.COSSecretKey == "" {
			v.addError("storage.cos_secret_key", "使用COS时SecretKey不能为空")
		}
		if cfg.Storage.COSBucket == "" {
			v.addError("storage.cos_bucket", "使用COS时Bucket不能为空")
		}
		if cfg.Storage.COSRegion == "" {
			v.addError("storage.cos_region", "使用COS时Region不能为空")
		}
	}
}

// validateWeChat 验证微信配置
func (v *ConfigValidator) validateWeChat(cfg *Config) {
	// 如果启用了微信登录，需要验证配置
	if cfg.System.EnableWechatLogin {
		if cfg.WeChat.AppID == "" || cfg.WeChat.AppID == "your-app-id" {
			v.addError("wechat.app_id", "启用微信登录时AppID不能为空")
		}
		if cfg.WeChat.AppSecret == "" || cfg.WeChat.AppSecret == "your-app-secret" {
			v.addError("wechat.app_secret", "启用微信登录时AppSecret不能为空")
		}
	}
}

// ValidateAndPrint 验证配置并打印结果
func ValidateAndPrint(cfg *Config) bool {
	validator := NewConfigValidator()
	errors := validator.Validate(cfg)

	if len(errors) == 0 {
		fmt.Println("✓ Configuration validation passed")
		return true
	}

	fmt.Println("✗ Configuration validation failed:")
	for _, err := range errors {
		fmt.Printf("  - %s: %s\n", err.Field, err.Message)
	}
	return false
}
