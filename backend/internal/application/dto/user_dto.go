package dto

import (
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
)

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Nickname string        `json:"nickname" binding:"required"`
	Phone    string        `json:"phone" binding:"required"`
	Email    string        `json:"email"`
	Password string        `json:"password" binding:"required,min=6"`
	Role     entity.Role   `json:"role" binding:"required"`
	OrgID    string        `json:"org_id" binding:"required"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Nickname string        `json:"nickname"`
	Email    string        `json:"email"`
	Role     entity.Role   `json:"role"`
	OrgID    string        `json:"org_id"`
	Status   entity.UserStatus `json:"status"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID          string     `json:"id"`
	Nickname    string     `json:"nickname"`
	Phone       string     `json:"phone"`
	Email       string     `json:"email"`
	Role        string     `json:"role"`
	Status      string     `json:"status"`
	OrgID       string     `json:"org_id"`
	OrgName     string     `json:"org_name,omitempty"`
	Avatar      string     `json:"avatar"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// UserListRequest 用户列表请求
type UserListRequest struct {
	Page     int    `form:"page,default=1" binding:"min=1"`
	PageSize int    `form:"page_size,default=10" binding:"min=1,max=100"`
	Keyword  string `form:"keyword"`
	Role     string `form:"role"`
	Status   string `form:"status"`
	OrgID    string `form:"org_id"`
}

// PageResult 分页结果
type PageResult[T any] struct {
	List       []T   `json:"list"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

// UserListResponse 用户列表响应
type UserListResponse = PageResult[UserResponse]

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int          `json:"expires_in"`
	TokenType    string       `json:"token_type"`
	User         UserResponse `json:"user"`
}

// RefreshTokenRequest 刷新token请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	Nickname     string `json:"nickname"`
	Email        string `json:"email"`
	RealName     string `json:"real_name"`
	IDCard       string `json:"id_card"`
	Gender       string `json:"gender"`
	Address      string `json:"address"`
	Emergency    string `json:"emergency"`
	EmergencyTel string `json:"emergency_tel"`
	Introduction string `json:"introduction"`
}

// UserProfileResponse 用户资料响应
type UserProfileResponse struct {
	UserResponse
	RealName     string `json:"real_name"`
	IDCard       string `json:"id_card"`
	Gender       string `json:"gender"`
	Address      string `json:"address"`
	Emergency    string `json:"emergency"`
	EmergencyTel string `json:"emergency_tel"`
	Introduction string `json:"introduction"`
}

// UserStatsResponse 用户统计响应
type UserStatsResponse struct {
	TotalCases     int64 `json:"total_cases"`
	ActiveCases    int64 `json:"active_cases"`
	CompletedCases int64 `json:"completed_cases"`
	TotalTasks     int64 `json:"total_tasks"`
	PendingTasks   int64 `json:"pending_tasks"`
}

// ToUserResponse 转换为用户响应
func ToUserResponse(user *entity.User) UserResponse {
	orgName := ""
	if user.Org != nil {
		orgName = user.Org.Name
	}

	return UserResponse{
		ID:          user.ID,
		Nickname:    user.Nickname,
		Phone:       user.Phone,
		Email:       user.Email,
		Role:        string(user.Role),
		Status:      string(user.Status),
		OrgID:       user.OrgID,
		OrgName:     orgName,
		Avatar:      user.Avatar,
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
	}
}

// NewUserListResponse 创建用户列表响应
func NewUserListResponse(list []UserResponse, total int64, page, pageSize int) UserListResponse {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return UserListResponse{
		List:       list,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

// ToUserProfileResponse 转换为用户资料响应
func ToUserProfileResponse(user *entity.User) UserProfileResponse {
	return UserProfileResponse{
		UserResponse: ToUserResponse(user),
		RealName:     user.RealName,
		IDCard:       user.IDCard,
		Gender:       user.Gender,
		Address:      user.Address,
		Emergency:    user.Emergency,
		EmergencyTel: user.EmergencyTel,
		Introduction: user.Introduction,
	}
}
