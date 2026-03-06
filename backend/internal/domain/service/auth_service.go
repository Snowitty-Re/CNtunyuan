package service

import (
	"context"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/valueobject"
	"github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/cache"
	"github.com/Snowitty-Re/CNtunyuan/pkg/errors"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
)

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
	UserID   string
	Nickname string
	Role     string
	OrgID    string
}

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
		return nil, nil, errors.ErrInvalidPassword
	}

	// Check user status
	switch user.Status {
	case entity.UserStatusInactive:
		return nil, nil, errors.ErrAccountDisabled
	case entity.UserStatusBanned:
		return nil, nil, errors.ErrAccountLocked
	}

	// Verify password
	if !user.CheckPassword(creds.Password) {
		logger.Warn("Login failed - wrong password", logger.String("username", creds.Username))
		return nil, nil, errors.ErrInvalidPassword
	}

	// Record login
	user.RecordLogin(ip)
	if err := s.userRepo.Update(ctx, user); err != nil {
		logger.Error("Failed to record login", logger.Err(err))
	}

	// Generate token
	tokens, err := s.tokenService.GenerateTokenPair(ctx, user)
	if err != nil {
		return nil, nil, errors.Wrap(err, errors.CodeInternal, "token generation failed")
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
		if errors.IsCode(err, errors.CodeTokenExpired) {
			return nil, nil, errors.ErrTokenExpired
		}
		return nil, nil, errors.ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, nil, errors.ErrInvalidToken
	}

	if !user.IsActive() {
		return nil, nil, errors.ErrAccountDisabled
	}

	tokens, err := s.tokenService.GenerateTokenPair(ctx, user)
	if err != nil {
		return nil, nil, errors.Wrap(err, errors.CodeInternal, "token generation failed")
	}

	return &valueobject.LoginResult{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		TokenType:    "Bearer",
	}, user, nil
}


// GetCurrentUser get current user info
func (s *AuthService) GetCurrentUser(ctx context.Context, userID string) (*entity.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

// WechatLogin WeChat mini-program login
func (s *AuthService) WechatLogin(ctx context.Context, code string, ip string) (*valueobject.LoginResult, *entity.User, bool, error) {
	// Get session from WeChat
	session, err := s.wechatClient.Code2Session(code)
	if err != nil {
		return nil, nil, false, errors.Wrap(err, errors.CodeInternal, "wechat code2session failed")
	}

	// Find user by openid
	user, err := s.userRepo.FindByOpenID(ctx, session.OpenID)
	if err != nil {
		// User not found, need to bind phone
		return nil, nil, true, nil
	}

	// Check user status
	if !user.IsActive() {
		return nil, nil, false, errors.ErrAccountDisabled
	}

	// Record login
	user.RecordLogin(ip)
	if err := s.userRepo.Update(ctx, user); err != nil {
		logger.Error("Failed to record login", logger.Err(err))
	}

	// Generate token
	tokens, err := s.tokenService.GenerateTokenPair(ctx, user)
	if err != nil {
		return nil, nil, false, errors.Wrap(err, errors.CodeInternal, "token generation failed")
	}

	logger.Info("Wechat login success",
		logger.String("user_id", user.ID),
		logger.String("openid", session.OpenID),
		logger.String("ip", ip),
	)

	return &valueobject.LoginResult{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		TokenType:    "Bearer",
	}, user, false, nil
}

// ValidateToken 验证token
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*TokenClaims, error) {
	return s.tokenService.ValidateToken(ctx, token)
}
