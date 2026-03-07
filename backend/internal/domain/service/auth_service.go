package service

import (
	"context"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/valueobject"
	"github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/cache"
	"github.com/Snowitty-Re/CNtunyuan/pkg/errors"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/google/uuid"
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
func (s *AuthService) WechatLogin(ctx context.Context, code string, ip string, userInfo *valueobject.WechatUserInfo) (*valueobject.LoginResult, *entity.User, bool, error) {
	// Get session from WeChat
	session, err := s.wechatClient.Code2Session(code)
	if err != nil {
		logger.Error("Wechat code2session failed", logger.Err(err))
		return nil, nil, false, errors.Wrap(err, errors.CodeInternal, "wechat code2session failed")
	}

	logger.Info("Wechat login code2session success", logger.String("openid", session.OpenID))

	// 使用微信提供的用户信息
	nickname := "微信用户"
	avatar := ""
	if userInfo != nil {
		if userInfo.Nickname != "" {
			nickname = userInfo.Nickname
		}
		if userInfo.Avatar != "" {
			avatar = userInfo.Avatar
		}
	}

	// Find user by openid
	user, err := s.userRepo.FindByOpenID(ctx, session.OpenID)
	if err != nil {
		// User not found, create a temporary user with openid
		// This allows binding phone later while preserving the openid

		// Get or create default org
		orgID, orgErr := s.getDefaultOrgID(ctx)
		if orgErr != nil {
			logger.Error("Failed to get default org", logger.Err(orgErr))
			return nil, nil, false, errors.Wrap(orgErr, errors.CodeInternal, "get default org failed")
		}

		tempUser := &entity.User{
			Nickname: nickname,
			Avatar:   avatar,
			Phone:    "", // Will be filled when binding phone
			Role:     entity.RoleVolunteer,
			Status:   entity.UserStatusActive,
			OrgID:    orgID,
			WxOpenID: session.OpenID,
		}
		// Set a random password (user will login via wechat)
		if pwdErr := tempUser.SetPassword(uuid.New().String()[:8]); pwdErr != nil {
			logger.Error("Failed to set temp user password", logger.Err(pwdErr))
			return nil, nil, false, errors.Wrap(pwdErr, errors.CodeInternal, "set password failed")
		}

		if createErr := s.userRepo.Create(ctx, tempUser); createErr != nil {
			logger.Error("Failed to create temp user", logger.Err(createErr), logger.String("openid", session.OpenID))
			return nil, nil, false, errors.Wrap(createErr, errors.CodeInternal, "create temp user failed")
		}

		logger.Info("Created temp user for wechat login",
			logger.String("user_id", tempUser.ID),
			logger.String("openid", session.OpenID),
		)

		// Return the temp user, frontend still needs to bind phone
		return nil, tempUser, true, nil
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

// GenerateTokenPair 生成token对（供handler使用）
func (s *AuthService) GenerateTokenPair(user *entity.User) (*TokenPair, error) {
	return s.tokenService.GenerateTokenPair(context.Background(), user)
}

// getDefaultOrgID 获取默认组织ID
func (s *AuthService) getDefaultOrgID(ctx context.Context) (string, error) {
	// 使用根组织ID作为默认值
	// 实际应用中应该检查组织是否存在
	return "00000000-0000-0000-0000-000000000000", nil
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
