package handler

import (
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
type WechatLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
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

		// Protected routes
		auth.GET("/me", h.authMiddleware.Required(), h.GetCurrentUser)
	}
}

// Login login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, validator.ValidateStruct(&req))
		return
	}

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

// AdminLogin admin login
func (h *AuthHandler) AdminLogin(c *gin.Context) {
	h.Login(c)
}

// RefreshToken refresh token
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

// Logout logout
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

// GetCurrentUser get current user info
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

// WechatLogin WeChat mini-program login
func (h *AuthHandler) WechatLogin(c *gin.Context) {
	var req WechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, validator.ValidateStruct(&req))
		return
	}

	result, user, needBind, err := h.authService.WechatLogin(c.Request.Context(), req.Code, c.ClientIP())
	if err != nil {
		logger.Error("Wechat login failed", logger.Err(err))
		response.Error(c, err)
		return
	}

	// Need to bind phone
	if needBind {
		response.Success(c, gin.H{
			"need_bind_phone": true,
			"user": gin.H{
				"id":       "",
				"nickname": "微信用户",
				"phone":    "",
				"role":     "volunteer",
				"status":   "active",
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
