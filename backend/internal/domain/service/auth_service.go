package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/valueobject"
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

// SMSProvider 短信服务接口
type SMSProvider interface {
	SendVerifyCode(ctx context.Context, phone string) (string, error)
	VerifyCode(ctx context.Context, phone, code string) bool
}

// AuthService auth service
type AuthService struct {
	userRepo       repository.UserRepository
	tokenService   TokenService
	cache          Cache
	wechatClient   WechatClient
	smsService     SMSProvider
	securityConfig *config.SecurityConfig
}

// NewAuthService create auth service
func NewAuthService(
	userRepo repository.UserRepository,
	tokenService TokenService,
	cache Cache,
	wechatClient WechatClient,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		tokenService: tokenService,
		cache:        cache,
		wechatClient: wechatClient,
	}
}

func (s *AuthService) ensureSMSService() error {
	if s.smsService == nil {
		return errors.New(errors.CodeInternal, "sms service not configured")
	}
	return nil
}

// SetSecurityConfig 设置安全配置
func (s *AuthService) SetSecurityConfig(cfg *config.SecurityConfig) {
	s.securityConfig = cfg
}

// SetSMSService 设置短信服务（用于依赖注入）
func (s *AuthService) SetSMSService(smsService SMSProvider) {
	s.smsService = smsService
}

// Login login
func (s *AuthService) Login(ctx context.Context, creds valueobject.LoginCredentials, ip string) (*valueobject.LoginResult, *entity.User, error) {
	// Check login lockout
	if s.isLoginLocked(ctx, creds.Username) {
		return nil, nil, errors.New(errors.CodeAccountLocked, "登录尝试次数过多，账户已临时锁定")
	}

	// Find user
	user, err := s.userRepo.FindByPhoneOrNickname(ctx, creds.Username)
	if err != nil {
		logger.Warn("Login failed - user not found", logger.String("username", creds.Username))
		s.recordFailedLogin(ctx, creds.Username)
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
		s.recordFailedLogin(ctx, creds.Username)
		return nil, nil, errors.ErrInvalidPassword
	}

	// Clear failed login attempts on success
	s.clearLoginAttempts(ctx, creds.Username)

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
			BaseEntity: entity.BaseEntity{
				ID: uuid.New().String(),
			},
			Nickname: nickname,
			Avatar:   avatar,
			// 使用唯一占位符避免 phone 唯一约束冲突，绑定手机号后会被真实号码覆盖
			Phone:    "wx" + strings.ReplaceAll(uuid.New().String(), "-", "")[:10],
			Role:     entity.RoleVolunteer,
			Status:   entity.UserStatusActive,
			OrgID:    orgID,
			WxOpenID: session.OpenID,
		}
		// Set a random password (user will login via wechat)
		if pwdErr := tempUser.SetPassword(uuid.New().String()[:12]); pwdErr != nil {
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
	const defaultOrgID = "00000000-0000-0000-0000-000000000000"
	return defaultOrgID, nil
}

// BindPhone 绑定手机号
// userID 可为空，表示新用户注册
func (s *AuthService) BindPhone(ctx context.Context, userID string, phone string, code string) (*valueobject.LoginResult, error) {
	logger.Info("BindPhone called",
		logger.String("user_id", userID),
		logger.String("phone", phone),
	)

	// 在 Service 层校验手机号格式（防御纵深，不依赖 Handler 层校验）
	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	if !phoneRegex.MatchString(phone) {
		return nil, errors.New(errors.CodeInvalidParam, "手机号格式不正确")
	}

	// 验证验证码
	if err := s.ensureSMSService(); err != nil {
		return nil, err
	}

	if !s.smsService.VerifyCode(ctx, phone, code) {
		return nil, errors.New(errors.CodeInvalidCaptcha, "验证码错误或已过期")
	}

	// 检查手机号是否已被绑定
	existingUser, err := s.userRepo.FindByPhone(ctx, phone)
	if err == nil && existingUser != nil {
		logger.Info("Phone already bound to user", logger.String("existing_user_id", existingUser.ID))
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
		logger.Info("Updating existing user", logger.String("user_id", userID))
		// 更新现有用户
		user, err = s.userRepo.FindByID(ctx, userID)
		if err != nil {
			logger.Error("Find user by ID failed", logger.String("user_id", userID), logger.Err(err))
			return nil, errors.ErrUserNotFound
		}
		user.Phone = phone
		if err := s.userRepo.Update(ctx, user); err != nil {
			logger.Error("Update user failed", logger.String("user_id", userID), logger.Err(err))
			return nil, errors.Wrap(err, errors.CodeInternal, "update user failed")
		}
	} else {
		logger.Info("Creating new user for phone", logger.String("phone", phone))
		// 创建新用户（志愿者角色）
		// 安全地获取手机号后4位
		suffix := phone
		if len(phone) >= 4 {
			suffix = phone[len(phone)-4:]
		}

		user = &entity.User{
			Nickname: "志愿者" + suffix,
			Phone:    phone,
			Role:     entity.RoleVolunteer,
			Status:   entity.UserStatusActive,
			OrgID:    "00000000-0000-0000-0000-000000000000", // 默认组织
		}
		// 设置随机密码（用户通过微信或短信验证码登录，不需要记住密码）
		if err := user.SetPassword(uuid.New().String()[:12]); err != nil {
			logger.Error("Set password failed", logger.Err(err))
			return nil, errors.Wrap(err, errors.CodeInternal, "set password failed")
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			logger.Error("Create user failed", logger.String("phone", phone), logger.Err(err))
			return nil, errors.Wrap(err, errors.CodeInternal, "create user failed")
		}
		logger.Info("User created successfully", logger.String("user_id", user.ID))
	}

	// 生成 token
	tokens, err := s.tokenService.GenerateTokenPair(ctx, user)
	if err != nil {
		logger.Error("Generate token pair failed", logger.Err(err))
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
	if err := s.ensureSMSService(); err != nil {
		return err
	}

	_, err := s.smsService.SendVerifyCode(ctx, phone)
	return err
}

// ResetPassword 重置密码（通过短信验证码）
func (s *AuthService) ResetPassword(ctx context.Context, phone, code, newPassword string) error {
	// 验证验证码
	if err := s.ensureSMSService(); err != nil {
		return err
	}

	if !s.smsService.VerifyCode(ctx, phone, code) {
		return errors.New(errors.CodeInvalidCaptcha, "验证码错误或已过期")
	}

	// 查找用户
	user, err := s.userRepo.FindByPhone(ctx, phone)
	if err != nil {
		return errors.ErrUserNotFound
	}

	// 设置新密码
	if err := user.SetPassword(newPassword); err != nil {
		return errors.Wrap(err, errors.CodeInvalidParam, "密码格式不正确")
	}

	// 更新用户
	if err := s.userRepo.Update(ctx, user); err != nil {
		return errors.Wrap(err, errors.CodeInternal, "更新密码失败")
	}

	logger.Info("Password reset success", logger.String("user_id", user.ID))
	return nil
}

// isLoginLocked 检查登录是否被锁定
func (s *AuthService) isLoginLocked(ctx context.Context, username string) bool {
	if s.cache == nil {
		return false
	}
	key := fmt.Sprintf("login_lock:%s", username)
	exists, err := s.cache.Exists(ctx, key)
	if err != nil {
		return false
	}
	return exists
}

// recordFailedLogin 记录登录失败
func (s *AuthService) recordFailedLogin(ctx context.Context, username string) {
	if s.cache == nil {
		return
	}

	maxAttempts := 5
	lockoutDuration := 1800
	if s.securityConfig != nil {
		if s.securityConfig.MaxLoginAttempts > 0 {
			maxAttempts = s.securityConfig.MaxLoginAttempts
		}
		if s.securityConfig.LockoutDuration > 0 {
			lockoutDuration = s.securityConfig.LockoutDuration
		}
	}

	attemptsKey := fmt.Sprintf("login_attempts:%s", username)
	var attempts int
	if err := s.cache.Get(ctx, attemptsKey, &attempts); err != nil {
		attempts = 0
	}
	attempts++

	// Store attempts with a window equal to lockout duration
	window := time.Duration(lockoutDuration) * time.Second
	if err := s.cache.Set(ctx, attemptsKey, attempts, window); err != nil {
		logger.Error("Failed to record login attempt", logger.Err(err))
		return
	}

	if attempts >= maxAttempts {
		lockKey := fmt.Sprintf("login_lock:%s", username)
		if err := s.cache.Set(ctx, lockKey, true, window); err != nil {
			logger.Error("Failed to set login lock", logger.Err(err))
		}
		logger.Warn("Account locked due to too many failed login attempts",
			logger.String("username", username),
			logger.Int("attempts", attempts),
		)
	}
}

// clearLoginAttempts 清除登录失败记录
func (s *AuthService) clearLoginAttempts(ctx context.Context, username string) {
	if s.cache == nil {
		return
	}
	attemptsKey := fmt.Sprintf("login_attempts:%s", username)
	lockKey := fmt.Sprintf("login_lock:%s", username)
	s.cache.Delete(ctx, attemptsKey, lockKey)
}
