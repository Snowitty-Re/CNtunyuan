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

// BindPhone 绑定手机号
// userID 可为空，表示新用户注册
func (s *AuthService) BindPhone(ctx context.Context, userID string, phone string, code string) (*valueobject.LoginResult, error) {
	// TODO: 验证验证码
	// 这里应该调用短信服务验证验证码
	// 为了简化，暂时跳过验证码验证
	
	// 检查手机号是否已被绑定
	existingUser, err := s.userRepo.FindByPhone(ctx, phone)
	if err == nil && existingUser != nil {
		// 如果提供了 userID，且是同一用户，则更新
		if userID != "" && existingUser.ID == userID {
			// 同一用户，无需操作
			tokens, err := s.tokenService.GenerateTokenPair(ctx, existingUser)
			if err != nil {
				return nil, errors.Wrap(err, errors.CodeInternal, "token generation failed")
			}
			return &valueobject.LoginResult{
				AccessToken:  tokens.AccessToken,
				RefreshToken: tokens.RefreshToken,
				ExpiresIn:    tokens.ExpiresIn,
				TokenType:    "Bearer",
			}, nil
		}
		return nil, errors.ErrUserExists
	}
	
	// 创建新用户或更新现有用户
	var user *entity.User
	
	if userID != "" {
		// 更新现有用户
		user, err = s.userRepo.FindByID(ctx, userID)
		if err != nil {
			return nil, errors.ErrUserNotFound
		}
		user.Phone = phone
		if err := s.userRepo.Update(ctx, user); err != nil {
			return nil, errors.Wrap(err, errors.CodeInternal, "update user failed")
		}
	} else {
		// 创建新用户（志愿者角色）
		user = &entity.User{
			Nickname: "志愿者" + phone[len(phone)-4:],
			Phone:    phone,
			Role:     entity.RoleVolunteer,
			Status:   entity.UserStatusActive,
			OrgID:    "00000000-0000-0000-0000-000000000000", // 默认组织
		}
		// 设置默认密码
		user.SetPassword("123456") // 实际应该发送随机密码到手机
		
		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, errors.Wrap(err, errors.CodeInternal, "create user failed")
		}
	}
	
	// 生成 token
	tokens, err := s.tokenService.GenerateTokenPair(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeInternal, "token generation failed")
	}
	
	logger.Info("Bind phone success",
		logger.String("user_id", user.ID),
		logger.String("phone", phone),
	)
	
	return &valueobject.LoginResult{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		TokenType:    "Bearer",
	}, nil
}

// SendVerifyCode 发送验证码
func (s *AuthService) SendVerifyCode(ctx context.Context, phone string) error {
	// TODO: 集成短信服务（如阿里云短信、腾讯云短信）
	// 这里模拟发送验证码
	
	// 生成6位验证码
	// code := fmt.Sprintf("%06d", rand.Intn(1000000))
	
	// 将验证码存入缓存，5分钟有效期
	// if s.cache != nil {
	//     key := fmt.Sprintf("verify_code:%s", phone)
	//     s.cache.Set(ctx, key, code, 5*time.Minute)
	// }
	
	// 模拟发送成功
	logger.Info("Send verify code", logger.String("phone", phone))
	
	// 实际项目中应该调用短信API
	// 如: smsClient.Send(phone, fmt.Sprintf("您的验证码是：%s，5分钟内有效。", code))
	
	return nil
}
