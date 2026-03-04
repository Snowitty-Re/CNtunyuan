package handler

import (
	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/valueobject"
	"github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/middleware"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"
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
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, user, err := h.authService.Login(c.Request.Context(), valueobject.LoginCredentials{
		Username: req.Username,
		Password: req.Password,
	}, c.ClientIP())

	if err != nil {
		logger.Warn("Login failed", logger.String("username", req.Username), logger.Err(err))
		switch err {
		case service.ErrInvalidCredentials:
			response.Unauthorized(c, "invalid username or password")
		case service.ErrUserDisabled:
			response.Unauthorized(c, "user is disabled")
		case service.ErrUserBanned:
			response.Unauthorized(c, "user is banned")
		default:
			response.InternalServerError(c, "login failed")
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
		response.BadRequest(c, err.Error())
		return
	}

	result, user, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		switch err {
		case service.ErrTokenExpired:
			response.Unauthorized(c, "refresh token expired")
		case service.ErrTokenInvalid:
			response.Unauthorized(c, "invalid token")
		default:
			response.InternalServerError(c, "refresh token failed")
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
		response.Unauthorized(c, "please login first")
		return
	}

	user, err := h.authService.GetCurrentUser(c.Request.Context(), userID)
	if err != nil {
		logger.Warn("Get current user failed", logger.String("user_id", userID), logger.Err(err))
		response.NotFound(c, "user not found")
		return
	}

	response.Success(c, dto.ToUserResponse(user))
}

// WechatLogin WeChat mini-program login
func (h *AuthHandler) WechatLogin(c *gin.Context) {
	var req WechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, user, needBind, err := h.authService.WechatLogin(c.Request.Context(), req.Code, c.ClientIP())
	if err != nil {
		logger.Error("Wechat login failed", logger.Err(err))
		response.Unauthorized(c, "wechat login failed")
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
