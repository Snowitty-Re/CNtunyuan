package valueobject

import (
	"time"
)

// UserProfile 用户公开资料
type UserProfile struct {
	ID          string     `json:"id"`
	Nickname    string     `json:"nickname"`
	Avatar      string     `json:"avatar"`
	Role        string     `json:"role"`
	OrgID       string     `json:"org_id"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

// UserFullProfile 用户完整资料
type UserFullProfile struct {
	UserProfile
	Phone        string    `json:"phone"`
	Email        string    `json:"email"`
	RealName     string    `json:"real_name"`
	IDCard       string    `json:"id_card"`
	Gender       string    `json:"gender"`
	Address      string    `json:"address"`
	Emergency    string    `json:"emergency"`
	EmergencyTel string    `json:"emergency_tel"`
	Introduction string    `json:"introduction"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserStats 用户统计
type UserStats struct {
	TotalCases    int64 `json:"total_cases"`
	ActiveCases   int64 `json:"active_cases"`
	CompletedCases int64 `json:"completed_cases"`
	TotalTasks    int64 `json:"total_tasks"`
	PendingTasks  int64 `json:"pending_tasks"`
}

// LoginCredentials 登录凭证
type LoginCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResult 登录结果
type LoginResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	Nickname     string `json:"nickname,omitempty"`
	Email        string `json:"email,omitempty"`
	RealName     string `json:"real_name,omitempty"`
	IDCard       string `json:"id_card,omitempty"`
	Gender       string `json:"gender,omitempty"`
	Address      string `json:"address,omitempty"`
	Emergency    string `json:"emergency,omitempty"`
	EmergencyTel string `json:"emergency_tel,omitempty"`
	Introduction string `json:"introduction,omitempty"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// WechatUserInfo 微信用户信息
type WechatUserInfo struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}
