// Package sms 提供短信服务接口
package sms

import (
	"context"
	cryptoRand "crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/cache"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
)

// Service 短信服务接口
type Service interface {
	// SendVerifyCode 发送验证码
	SendVerifyCode(ctx context.Context, phone string) (string, error)
	// VerifyCode 验证验证码
	VerifyCode(ctx context.Context, phone, code string) bool
	// SendSMS 发送普通短信
	SendSMS(ctx context.Context, phone, templateCode string, params map[string]string) error
}

// smsService 短信服务实现
type smsService struct {
	config      *config.SMSConfig
	cache       cache.Cache
	provider    Provider
	devMode     bool
}

// Provider 短信提供商接口
type Provider interface {
	SendSMS(ctx context.Context, phone, signName, templateCode string, params map[string]string) error
}

// NewService 创建短信服务
func NewService(cfg *config.SMSConfig, cache cache.Cache) Service {
	s := &smsService{
		config:  cfg,
		cache:   cache,
		devMode: cfg.DevMode,
	}

	// 根据配置选择提供商
	if cfg.Provider == "aliyun" && cfg.AliyunAccessKeyID != "" {
		s.provider = NewAliyunProvider(cfg)
	} else if cfg.Provider == "tencent" && cfg.TencentSecretID != "" {
		s.provider = NewTencentProvider(cfg)
	}

	if s.devMode {
		logger.Warn("SMS service running in DEV mode - no real SMS will be sent")
	}

	return s
}

// SendVerifyCode 发送验证码
func (s *smsService) SendVerifyCode(ctx context.Context, phone string) (string, error) {
	// 生成6位验证码
	code := generateVerifyCode()

	// 计算验证码有效期
	expiry := 5 * time.Minute
	if s.config.CodeExpiry > 0 {
		expiry = time.Duration(s.config.CodeExpiry) * time.Second
	}

	// 缓存验证码
	if s.cache != nil {
		key := fmt.Sprintf("verify_code:%s", phone)
		if err := s.cache.Set(ctx, key, code, expiry); err != nil {
			logger.Error("Failed to cache verify code", logger.Err(err))
		}
	}

	// 开发模式：直接返回验证码，不实际发送
	if s.devMode {
		logger.Info("[DEV MODE] Verify code generated",
			logger.String("phone", phone),
			logger.String("code", code),
		)
		return code, nil
	}

	// 生产模式：必须有 provider
	if s.provider == nil {
		return "", fmt.Errorf("SMS provider not configured and dev mode is disabled")
	}

	// 生产模式：调用短信服务发送
	params := map[string]string{
		"code": code,
	}
	if err := s.provider.SendSMS(ctx, phone, s.config.SignName, "verify_code", params); err != nil {
		logger.Error("Failed to send SMS", logger.Err(err), logger.String("phone", phone))
		return "", err
	}

	logger.Info("Verify code sent", logger.String("phone", phone))
	return code, nil
}

// VerifyCode 验证验证码
func (s *smsService) VerifyCode(ctx context.Context, phone, code string) bool {
	if code == "" {
		return false
	}

	if s.cache == nil {
		// 没有缓存，无法验证
		logger.Warn("Cache not available, cannot verify code")
		return false
	}

	key := fmt.Sprintf("verify_code:%s", phone)
	var cachedCode string
	if err := s.cache.Get(ctx, key, &cachedCode); err != nil {
		logger.Warn("Failed to get verify code from cache", logger.Err(err))
		return false
	}

	if cachedCode != code {
		return false
	}

	// 验证成功后删除验证码
	s.cache.Delete(ctx, key)
	return true
}

// SendSMS 发送普通短信
func (s *smsService) SendSMS(ctx context.Context, phone, templateCode string, params map[string]string) error {
	if s.devMode {
		logger.Info("[DEV MODE] SMS would be sent",
			logger.String("phone", phone),
			logger.String("template", templateCode),
		)
		return nil
	}

	if s.provider == nil {
		return fmt.Errorf("SMS provider not configured")
	}

	return s.provider.SendSMS(ctx, phone, s.config.SignName, templateCode, params)
}

// generateVerifyCode 生成6位数字验证码（使用 crypto/rand）
func generateVerifyCode() string {
	n, err := cryptoRand.Int(cryptoRand.Reader, big.NewInt(1000000))
	if err != nil {
		// 降级方案：使用时间戳生成（极端情况）
		return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}
	return fmt.Sprintf("%06d", n.Int64())
}
