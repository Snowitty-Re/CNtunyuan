package service

import (
	"context"
	"errors"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/valueobject"
	"github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/cache"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
)

var (
	ErrInvalidCredentials = errors.New("username or password error")
	ErrUserDisabled       = errors.New("user is disabled")
	ErrUserBanned         = errors.New("user is banned")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenInvalid       = errors.New("token invalid")
)

// WechatSession 微信会话信息
type WechatSession struct {
	OpenID     string
	SessionKey string
	UnionID    string
}

// WechatClient 微信客户端接口
type WechatClient interface {
	Code2Session(code string) (*WechatSession, error)
}

// AuthService auth service
type AuthService struct {
	userRepo     repository.UserRepository
	tokenService TokenService
	cache        cache.Cache
	wechatClient WechatClient
}

// NewAuthService create auth service
func NewAuthService(
	userRepo repository.UserRepository,
	tokenService TokenService,
	cache cache.Cache,
	wechatClient WechatClient,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		tokenService: tokenService,
		cache:        cache,
		wechatClient: wechatClient,
	}
}

// Login login
func (s *AuthService) Login(ctx context.Context, creds valueobject.LoginCredentials, ip string) (*valueobject.LoginResult, *entity.User, error) {
	// Find user
	user, err := s.userRepo.FindByPhoneOrNickname(ctx, creds.Username)
	if err != nil {
		logger.Warn("Login failed - user not found", logger.String("username", creds.Username))
		return nil, nil, ErrInvalidCredentials
	}

	// Check user status
	switch user.Status {
	case entity.UserStatusInactive:
		return nil, nil, ErrUserDisabled
	case entity.UserStatusBanned:
		return nil, nil, ErrUserBanned
	}

	// Verify password
	if !user.CheckPassword(creds.Password) {
		logger.Warn("Login failed - wrong password", logger.String("username", creds.Username))
		return nil, nil, ErrInvalidCredentials
	}

	// Record login
	user.RecordLogin(ip)
	if err := s.userRepo.Update(ctx, user); err != nil {
		logger.Error("Failed to record login", logger.Err(err))
	}

	// Generate token
	tokens, err := s.tokenService.GenerateTokenPair(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	logger.Info("User login success",
		logger.String("user_id", user.ID),
		logger.String("role", string(user.Role)),
		logger.String("ip", ip),
	)

	return &valueobject.LoginResult{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		TokenType:    "Bearer",
	}, user, nil
}

// Logout logout
func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.tokenService.RevokeToken(ctx, token)
}

// RefreshToken refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*valueobject.LoginResult, *entity.User, error) {
	claims, err := s.tokenService.ValidateToken(ctx, refreshToken)
	if err != nil {
		return nil, nil, err
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, nil, ErrTokenInvalid
	}

	if !user.IsActive() {
		return nil, nil, ErrUserDisabled
	}

	tokens, err := s.tokenService.GenerateTokenPair(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	return &valueobject.LoginResult{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		TokenType:    "Bearer",
	}, user, nil
}

// ValidateToken validate token
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*TokenClaims, error) {
	return s.tokenService.ValidateToken(ctx, token)
}

// GetCurrentUser get current user
func (s *AuthService) GetCurrentUser(ctx context.Context, userID string) (*entity.User, error) {
	if s.cache != nil {
		var user entity.User
		if err := s.cache.Get(ctx, cacheKeyUser(userID), &user); err == nil {
			return &user, nil
		}
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		if err := s.cache.Set(ctx, cacheKeyUser(userID), user, 5*time.Minute); err != nil {
			logger.Warn("Failed to cache user", logger.Err(err))
		}
	}

	return user, nil
}

// ChangePassword change password
func (s *AuthService) ChangePassword(ctx context.Context, userID string, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if !user.CheckPassword(oldPassword) {
		return errors.New("old password is wrong")
	}

	if err := user.SetPassword(newPassword); err != nil {
		return err
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	if s.cache != nil {
		s.cache.Delete(ctx, cacheKeyUser(userID))
	}

	return nil
}

// TokenService token service interface
type TokenService interface {
	GenerateTokenPair(ctx context.Context, user *entity.User) (*TokenPair, error)
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
	RevokeToken(ctx context.Context, token string) error
}

// TokenPair token pair
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

// TokenClaims token claims
type TokenClaims struct {
	UserID    string      `json:"user_id"`
	Nickname  string      `json:"nickname"`
	Role      entity.Role `json:"role"`
	OrgID     string      `json:"org_id"`
	IssuedAt  time.Time   `json:"iat"`
	ExpiresAt time.Time   `json:"exp"`
}

// WechatLogin WeChat mini-program login
func (s *AuthService) WechatLogin(ctx context.Context, code string, ip string) (*valueobject.LoginResult, *entity.User, bool, error) {
	// 检查数据库连接
	if s.userRepo == nil {
		return nil, nil, false, errors.New("database not initialized, please complete system setup first")
	}

	// 检查微信客户端是否配置
	if s.wechatClient == nil {
		return nil, nil, false, errors.New("wechat login not configured")
	}

	// 调用微信API获取openid
	session, err := s.wechatClient.Code2Session(code)
	if err != nil {
		logger.Error("Wechat code2session failed", logger.Err(err))
		return nil, nil, false, err
	}

	openid := session.OpenID
	if openid == "" {
		return nil, nil, false, errors.New("failed to get openid from wechat")
	}

	logger.Info("Wechat login", logger.String("openid", openid))

	// 根据openid查找用户
	// TODO: 应该使用专门的 wechat_openid 字段
	user, err := s.userRepo.FindByPhone(ctx, openid)
	if err != nil {
		// 用户不存在，需要绑定手机号创建账号
		logger.Info("Wechat user not found, need bind phone", logger.String("openid", openid))
		return nil, nil, true, nil
	}

	// 检查用户状态
	if !user.IsActive() {
		return nil, nil, false, ErrUserDisabled
	}

	// 记录登录
	user.RecordLogin(ip)
	if err := s.userRepo.Update(ctx, user); err != nil {
		logger.Error("Failed to record wechat login", logger.Err(err))
	}

	// 生成token
	tokens, err := s.tokenService.GenerateTokenPair(ctx, user)
	if err != nil {
		return nil, nil, false, err
	}

	logger.Info("Wechat login success",
		logger.String("user_id", user.ID),
		logger.String("role", string(user.Role)),
		logger.String("ip", ip),
	)

	return &valueobject.LoginResult{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		TokenType:    "Bearer",
	}, user, false, nil
}

// cacheKeyUser user cache key
func cacheKeyUser(userID string) string {
	return cache.CacheKey("user", userID)
}
