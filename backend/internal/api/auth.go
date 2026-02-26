package api

import (
	"github.com/Snowitty-Re/CNtunyuan/internal/middleware"
	"github.com/Snowitty-Re/CNtunyuan/internal/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/utils"
	"github.com/Snowitty-Re/CNtunyuan/pkg/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	userService   *service.UserService
	wechatService *service.WeChatService
	jwtAuth       *auth.JWTAuth
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(userService *service.UserService, wechatService *service.WeChatService, jwtAuth *auth.JWTAuth) *AuthHandler {
	return &AuthHandler{
		userService:   userService,
		wechatService: wechatService,
		jwtAuth:       jwtAuth,
	}
}

// WeChatLoginRequest 微信登录请求
type WeChatLoginRequest struct {
	Code          string `json:"code" binding:"required"`
	EncryptedData string `json:"encrypted_data"`
	IV            string `json:"iv"`
}

// WeChatLoginResponse 微信登录响应
type WeChatLoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	IsNewUser    bool   `json:"is_new_user"`
}

// WeChatLogin 微信登录
// @Summary 微信小程序登录
// @Description 使用微信code换取登录凭证
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body WeChatLoginRequest true "登录参数"
// @Success 200 {object} utils.Response{data=WeChatLoginResponse}
// @Failure 400 {object} utils.Response
// @Router /auth/wechat-login [post]
func (h *AuthHandler) WeChatLogin(c *gin.Context) {
	var req WeChatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 调用微信接口换取OpenID和SessionKey
	wxResp, err := h.wechatService.Code2Session(req.Code)
	if err != nil {
		utils.Error(c, utils.CodeBadRequest, "微信登录失败: "+err.Error())
		return
	}

	// 解密用户信息(如果有)
	var nickname, avatar string
	if req.EncryptedData != "" && req.IV != "" {
		userInfo, err := h.wechatService.DecryptUserInfo(req.EncryptedData, wxResp.SessionKey, req.IV)
		if err == nil {
			if n, ok := userInfo["nickName"].(string); ok {
				nickname = n
			}
			if a, ok := userInfo["avatarUrl"].(string); ok {
				avatar = a
			}
		}
	}

	// 获取或创建用户
	user, isNew, err := h.userService.GetOrCreateByWeChat(c.Request.Context(), wxResp.OpenID, wxResp.UnionID, nickname, avatar)
	if err != nil {
		utils.ServerError(c, "用户处理失败: "+err.Error())
		return
	}

	// 更新最后登录时间
	go h.userService.UpdateLastLogin(c.Request.Context(), user.ID, c.ClientIP())

	// 生成Token
	orgID := ""
	if user.OrgID != nil {
		orgID = user.OrgID.String()
	}

	token, err := h.jwtAuth.GenerateToken(user.ID.String(), user.OpenID, user.UnionID, user.Role, orgID)
	if err != nil {
		utils.ServerError(c, "生成Token失败")
		return
	}

	refreshToken, err := h.jwtAuth.GenerateRefreshToken(user.ID.String(), user.OpenID, user.UnionID, user.Role, orgID)
	if err != nil {
		utils.ServerError(c, "生成RefreshToken失败")
		return
	}

	utils.Success(c, WeChatLoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    604800, // 7天
		IsNewUser:    isNew,
	})
}

// RefreshTokenRequest 刷新Token请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshToken 刷新Token
// @Summary 刷新访问令牌
// @Description 使用RefreshToken换取新的AccessToken
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body RefreshTokenRequest true "刷新参数"
// @Success 200 {object} utils.Response{data=WeChatLoginResponse}
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	accessToken, refreshToken, err := h.jwtAuth.RefreshToken(req.RefreshToken)
	if err != nil {
		utils.Unauthorized(c, err.Error())
		return
	}

	utils.Success(c, WeChatLoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    604800,
		IsNewUser:    false,
	})
}

// GetCurrentUser 获取当前用户
// @Summary 获取当前登录用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 认证
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response{data=model.User}
// @Router /auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		utils.Unauthorized(c, "未登录")
		return
	}

	userIDUUID, err := uuid.Parse(userID)
	if err != nil {
		utils.BadRequest(c, "用户ID格式错误")
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), userIDUUID)
	if err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}

	utils.Success(c, user)
}

// Logout 登出
// @Summary 用户登出
// @Description 用户登出
// @Tags 认证
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// 这里可以实现Token黑名单等逻辑
	utils.Success(c, nil)
}


