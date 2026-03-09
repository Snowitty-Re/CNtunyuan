package handler

import (
	"strings"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/valueobject"
	"github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/middleware"
	"github.com/Snowitty-Re/CNtunyuan/pkg/errors"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"
	"github.com/Snowitty-Re/CNtunyuan/pkg/validator"
	"github.com/gin-gonic/gin"
)

// AuthHandler auth handler
type AuthHandler struct {
	authService    *service.AuthService
	authMiddleware *middleware.AuthMiddleware
}

// NewAuthHandler create auth handler
func NewAuthHandler(authService *service.AuthService, authMiddleware *middleware.AuthMiddleware) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		authMiddleware: authMiddleware,
	}
}

// WechatLoginRequest WeChat mini-program login request
// @Description 微信小程序登录请求参数
type WechatLoginRequest struct {
	Code     string `json:"code" binding:"required" example:"wx_login_code_123"`         // 微信登录临时凭证
	Nickname string `json:"nickname" example:"张三"`                                    // 用户昵称（可选）
	Avatar   string `json:"avatar" example:"https://example.com/avatar.jpg"`            // 用户头像（可选）
}

// LoginRequest 登录请求
// @Description 用户登录请求参数
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"13800138000"`          // 用户名/手机号
	Password string `json:"password" binding:"required" example:"admin123"`            // 密码
}

// BindPhoneRequest 绑定手机号请求
// @Description 绑定手机号请求参数
type BindPhoneRequest struct {
	Phone string `json:"phone" binding:"required" example:"13800138000"`                // 手机号
	Code  string `json:"code" example:"123456"`                                          // 验证码（测试阶段可选）
}

// SendCodeRequest 发送验证码请求
// @Description 发送短信验证码请求参数
type SendCodeRequest struct {
	Phone string `json:"phone" binding:"required" example:"13800138000"`                // 手机号
}

// WechatLoginResponse 微信登录响应（需要绑定手机号）
// @Description 微信登录响应，当need_bind_phone为true时需要绑定手机号
type WechatLoginResponse struct {
	NeedBindPhone bool              `json:"need_bind_phone" example:"true"`                // 是否需要绑定手机号
	AccessToken   string            `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."` // 访问令牌（临时）
	RefreshToken  string            `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIs..."` // 刷新令牌
	ExpiresIn     int               `json:"expires_in" example:"604800"`                    // 过期时间（秒）
	TokenType     string            `json:"token_type" example:"Bearer"`                    // 令牌类型
	User          WechatLoginUser   `json:"user"`                                           // 用户信息
}

// WechatLoginUser 微信登录用户信息
type WechatLoginUser struct {
	ID       string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`            // 用户ID
	Nickname string `json:"nickname" example:"微信用户"`                                   // 昵称
	Phone    string `json:"phone" example:""`                                            // 手机号（可能为空）
	Role     string `json:"role" example:"volunteer"`                                    // 角色
	Status   string `json:"status" example:"active"`                                     // 状态
}

// SendCodeResponse 发送验证码响应
// @Description 发送验证码成功响应
type SendCodeResponse struct {
	Message string `json:"message" example:"验证码已发送"`                                  // 响应消息
	Expire  int    `json:"expire" example:"300"`                                         // 验证码有效期（秒）
}

// RegisterRoutes register routes
func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/admin-login", h.AdminLogin)
		auth.POST("/refresh", h.RefreshToken)
		auth.POST("/logout", h.Logout)
		auth.POST("/wechat-login", h.WechatLogin)
		auth.POST("/bind-phone", h.authMiddleware.Required(), h.BindPhone)
		auth.POST("/send-code", h.SendVerifyCode)

		// Protected routes
		auth.GET("/me", h.authMiddleware.Required(), h.GetCurrentUser)
	}
}

// Login 用户登录
// @Summary      用户登录
// @Description  使用用户名和密码登录系统，返回访问令牌和用户信息
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      LoginRequest  true  "登录请求参数"
// @Success      200      {object}  response.Response{data=dto.LoginResponse}  "登录成功"
// @Failure      400      {object}  response.Response  "参数错误"
// @Failure      401      {object}  response.Response  "用户名或密码错误"
// @Failure      403      {object}  response.Response  "账号被禁用或锁定"
// @Failure      500      {object}  response.Response  "服务器内部错误"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, validator.ValidateStruct(&req))
		return
	}

	// 去除用户名和密码前后的空格
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)

	result, user, err := h.authService.Login(c.Request.Context(), valueobject.LoginCredentials{
		Username: req.Username,
		Password: req.Password,
	}, c.ClientIP())

	if err != nil {
		logger.Warn("Login failed",
			logger.String("username", req.Username),
			logger.Err(err))

		// 使用新的错误体系
		switch {
		case errors.IsCode(err, errors.CodeInvalidPassword):
			response.Error(c, errors.ErrInvalidPassword.WithDetail("用户名或密码错误"))
		case errors.IsCode(err, errors.CodeAccountDisabled):
			response.Error(c, errors.ErrAccountDisabled)
		case errors.IsCode(err, errors.CodeAccountLocked):
			response.Error(c, errors.ErrAccountLocked)
		default:
			response.Error(c, err)
		}
		return
	}

	response.Success(c, dto.LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		TokenType:    result.TokenType,
		User:         dto.ToUserResponse(user),
	})
}

// AdminLogin 管理员登录
// @Summary      管理员登录
// @Description  管理员登录接口，与普通登录使用相同的认证逻辑，但通常用于管理后台
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      LoginRequest  true  "登录请求参数"
// @Success      200      {object}  response.Response{data=dto.LoginResponse}  "登录成功"
// @Failure      400      {object}  response.Response  "参数错误"
// @Failure      401      {object}  response.Response  "用户名或密码错误"
// @Failure      403      {object}  response.Response  "账号被禁用、锁定或无管理员权限"
// @Failure      500      {object}  response.Response  "服务器内部错误"
// @Router       /auth/admin-login [post]
func (h *AuthHandler) AdminLogin(c *gin.Context) {
	h.Login(c)
}

// RefreshToken 刷新访问令牌
// @Summary      刷新访问令牌
// @Description  使用刷新令牌获取新的访问令牌，避免用户重新登录
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      dto.RefreshTokenRequest  true  "刷新令牌请求参数"
// @Success      200      {object}  response.Response{data=dto.LoginResponse}  "刷新成功"
// @Failure      400      {object}  response.Response  "参数错误"
// @Failure      401      {object}  response.Response  "刷新令牌过期或无效"
// @Failure      403      {object}  response.Response  "账号被禁用"
// @Failure      500      {object}  response.Response  "服务器内部错误"
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, validator.ValidateStruct(&req))
		return
	}

	result, user, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.IsCode(err, errors.CodeTokenExpired):
			response.Error(c, errors.ErrTokenExpired)
		case errors.IsCode(err, errors.CodeInvalidToken):
			response.Error(c, errors.ErrInvalidToken)
		case errors.IsCode(err, errors.CodeAccountDisabled):
			response.Error(c, errors.ErrAccountDisabled)
		default:
			response.Error(c, err)
		}
		return
	}

	response.Success(c, dto.LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		TokenType:    result.TokenType,
		User:         dto.ToUserResponse(user),
	})
}

// Logout 用户登出
// @Summary      用户登出
// @Description  用户退出登录，使当前访问令牌失效
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer 访问令牌"  default(Bearer <token>)
// @Success      200            {object}  response.Response  "登出成功"
// @Failure      500            {object}  response.Response  "服务器内部错误"
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token != "" {
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		if err := h.authService.Logout(c.Request.Context(), token); err != nil {
			logger.Warn("Logout failed", logger.Err(err))
		}
	}

	response.Success(c, nil)
}

// GetCurrentUser 获取当前登录用户信息
// @Summary      获取当前用户信息
// @Description  获取当前登录用户的详细信息
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer 访问令牌"  default(Bearer <token>)
// @Success      200            {object}  response.Response{data=dto.UserResponse}  "获取成功"
// @Failure      401            {object}  response.Response  "未授权或令牌无效"
// @Failure      404            {object}  response.Response  "用户不存在"
// @Failure      500            {object}  response.Response  "服务器内部错误"
// @Router       /auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	user, err := h.authService.GetCurrentUser(c.Request.Context(), userID)
	if err != nil {
		logger.Warn("Get current user failed",
			logger.String("user_id", userID),
			logger.Err(err))
		response.Error(c, errors.ErrUserNotFound)
		return
	}

	response.Success(c, dto.ToUserResponse(user))
}

// WechatLogin 微信小程序登录
// @Summary      微信小程序登录
// @Description  微信小程序登录接口，使用微信临时登录凭证换取系统访问令牌。新用户需要绑定手机号
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      WechatLoginRequest  true  "微信登录请求参数"
// @Success      200      {object}  response.Response{data=dto.LoginResponse}     "登录成功（已绑定手机号的用户）"
// @Success      200      {object}  response.Response{data=WechatLoginResponse}   "需要绑定手机号（新用户）"
// @Failure      400      {object}  response.Response  "参数错误或微信登录凭证无效"
// @Failure      500      {object}  response.Response  "服务器内部错误或微信服务异常"
// @Router       /auth/wechat-login [post]
func (h *AuthHandler) WechatLogin(c *gin.Context) {
	var req WechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, validator.ValidateStruct(&req))
		return
	}

	// 构建用户信息
	userInfo := &valueobject.WechatUserInfo{
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
	}

	result, user, needBind, err := h.authService.WechatLogin(c.Request.Context(), req.Code, c.ClientIP(), userInfo)
	if err != nil {
		logger.Error("Wechat login failed", logger.Err(err))
		response.Error(c, err)
		return
	}

	// Need to bind phone
	if needBind {
		// Generate temp token for the new user
		// This allows the frontend to make authenticated requests (like bind-phone)
		tokens, err := h.authService.GenerateTokenPair(user)
		if err != nil {
			logger.Error("Failed to generate temp token", logger.Err(err))
			response.Error(c, errors.New(errors.CodeInternal, "token generation failed"))
			return
		}

		response.Success(c, gin.H{
			"need_bind_phone": true,
			"access_token":    tokens.AccessToken,
			"refresh_token":   tokens.RefreshToken,
			"expires_in":      tokens.ExpiresIn,
			"token_type":      "Bearer",
			"user": gin.H{
				"id":       user.ID,
				"nickname": user.Nickname,
				"phone":    user.Phone,
				"role":     user.Role,
				"status":   user.Status,
			},
		})
		return
	}

	response.Success(c, dto.LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		TokenType:    result.TokenType,
		User:         dto.ToUserResponse(user),
	})
}

// BindPhone 绑定手机号
// @Summary      绑定手机号
// @Description  为微信登录用户绑定手机号，新用户首次登录时需要完成此步骤
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string             true  "Bearer 访问令牌（微信登录获取的临时令牌）"  default(Bearer <token>)
// @Param        request        body      BindPhoneRequest   true  "绑定手机号请求参数"
// @Success      200            {object}  response.Response{data=dto.LoginResponse}  "绑定成功"
// @Failure      400            {object}  response.Response  "参数错误或手机号格式不正确"
// @Failure      401            {object}  response.Response  "未授权"
// @Failure      409            {object}  response.Response  "手机号已被其他账号绑定"
// @Failure      500            {object}  response.Response  "服务器内部错误"
// @Router       /auth/bind-phone [post]
func (h *AuthHandler) BindPhone(c *gin.Context) {
	var req BindPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, validator.ValidateStruct(&req))
		return
	}

	// 验证手机号格式
	if !validator.IsValidPhone(req.Phone) {
		response.Error(c, errors.New(errors.CodeInvalidParam, "手机号格式不正确"))
		return
	}

	// 获取当前用户ID（如果已登录）
	userID := middleware.GetUserID(c)

	result, err := h.authService.BindPhone(c.Request.Context(), userID, req.Phone, req.Code)
	if err != nil {
		logger.Error("Bind phone failed", logger.Err(err))
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// SendVerifyCode 发送短信验证码
// @Summary      发送短信验证码
// @Description  向指定手机号发送短信验证码，用于手机号绑定、密码重置等场景
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      SendCodeRequest  true  "发送验证码请求参数"
// @Success      200      {object}  response.Response{data=SendCodeResponse}  "发送成功"
// @Failure      400      {object}  response.Response  "参数错误或手机号格式不正确"
// @Failure      429      {object}  response.Response  "发送过于频繁，请稍后再试"
// @Failure      500      {object}  response.Response  "服务器内部错误或短信服务异常"
// @Router       /auth/send-code [post]
func (h *AuthHandler) SendVerifyCode(c *gin.Context) {
	var req SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, validator.ValidateStruct(&req))
		return
	}

	// 验证手机号格式
	if !validator.IsValidPhone(req.Phone) {
		response.Error(c, errors.New(errors.CodeInvalidParam, "手机号格式不正确"))
		return
	}

	if err := h.authService.SendVerifyCode(c.Request.Context(), req.Phone); err != nil {
		logger.Error("Send verify code failed", logger.Err(err))
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "验证码已发送",
		"expire":  300, // 5分钟有效期
	})
}
