package entity

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/valueobject"
	"github.com/Snowitty-Re/CNtunyuan/pkg/utils"
	"github.com/google/uuid"
)

// UserStatus 用户状态
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusBanned   UserStatus = "banned"
)

// Role 角色类型
type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleAdmin      Role = "admin"
	RoleManager    Role = "manager"
	RoleVolunteer  Role = "volunteer"
)

// RoleHierarchy 角色层级
type RoleHierarchy int

const (
	RoleLevelSuperAdmin RoleHierarchy = 4
	RoleLevelAdmin      RoleHierarchy = 3
	RoleLevelManager    RoleHierarchy = 2
	RoleLevelVolunteer  RoleHierarchy = 1
)

// GetRoleLevel 获取角色等级
func GetRoleLevel(role Role) RoleHierarchy {
	switch role {
	case RoleSuperAdmin:
		return RoleLevelSuperAdmin
	case RoleAdmin:
		return RoleLevelAdmin
	case RoleManager:
		return RoleLevelManager
	default:
		return RoleLevelVolunteer
	}
}

// HasRole 检查角色权限
func HasRole(userRole Role, requiredRole Role) bool {
	return GetRoleLevel(userRole) >= GetRoleLevel(requiredRole)
}

// User 用户领域实体
type User struct {
	BaseEntity
	Nickname     string           `gorm:"size:100;not null" json:"nickname"`
	Phone        string           `gorm:"size:20;uniqueIndex;not null" json:"phone"`
	Email        string           `gorm:"size:100;uniqueIndex" json:"email,omitempty"`
	Password     string           `gorm:"size:255;not null" json:"-"`
	Role         Role             `gorm:"size:20;not null;default:'volunteer'" json:"role"`
	Status       UserStatus       `gorm:"size:20;not null;default:'active'" json:"status"`
	OrgID        string           `gorm:"type:uuid;not null;index" json:"org_id"`
	Avatar       string           `gorm:"size:255" json:"avatar,omitempty"`
	LastLoginAt  *time.Time       `json:"last_login_at,omitempty"`
	LastLoginIP  string           `gorm:"size:50" json:"last_login_ip,omitempty"`
	RealName     string           `gorm:"size:50" json:"real_name,omitempty"`
	IDCard       string           `gorm:"size:18" json:"id_card,omitempty"`
	Gender       string           `gorm:"size:10" json:"gender,omitempty"`
	Address      string           `gorm:"size:255" json:"address,omitempty"`
	Emergency    string           `gorm:"size:50" json:"emergency,omitempty"`
	EmergencyTel string           `gorm:"size:20" json:"emergency_tel,omitempty"`
	Introduction string           `gorm:"type:text" json:"introduction,omitempty"`
	WxOpenID     string           `gorm:"column:wx_openid;size:100;uniqueIndex" json:"wx_openid,omitempty"`
	WxUnionID    string           `gorm:"column:wx_unionid;size:100" json:"wx_unionid,omitempty"`
	Org          *Organization    `gorm:"foreignKey:OrgID" json:"org,omitempty"`
	Permissions  []Permission     `gorm:"many2many:user_permissions;" json:"permissions,omitempty"`
}

// TableName 表名
func (User) TableName() string {
	return "ty_users"
}

// IsActive 是否活跃
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// IsAdmin 是否管理员
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin || u.Role == RoleSuperAdmin
}

// IsSuperAdmin 是否超级管理员
func (u *User) IsSuperAdmin() bool {
	return u.Role == RoleSuperAdmin
}

// HasPermission 检查权限
func (u *User) HasPermission(required Role) bool {
	return HasRole(u.Role, required)
}

// Validate 验证用户
func (u *User) Validate() error {
	if u.Nickname == "" {
		return errors.New("昵称不能为空")
	}
	if len(u.Nickname) > 100 {
		return errors.New("昵称不能超过100字符")
	}
	if err := u.ValidatePhone(); err != nil {
		return err
	}
	if u.Email != "" && !utils.IsValidEmail(u.Email) {
		return errors.New("邮箱格式不正确")
	}
	return nil
}

// ValidatePhone 验证手机号
func (u *User) ValidatePhone() error {
	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	if !phoneRegex.MatchString(u.Phone) {
		return errors.New("手机号格式不正确")
	}
	return nil
}

// SetPassword 设置密码
func (u *User) SetPassword(plainPassword string) error {
	if len(plainPassword) < 6 {
		return errors.New("密码至少需要6位")
	}
	hash, err := utils.HashPassword(plainPassword)
	if err != nil {
		return err
	}
	u.Password = hash
	return nil
}

// CheckPassword 检查密码
func (u *User) CheckPassword(plainPassword string) bool {
	return utils.CheckPassword(plainPassword, u.Password)
}

// RecordLogin 记录登录
func (u *User) RecordLogin(ip string) {
	now := time.Now()
	u.LastLoginAt = &now
	u.LastLoginIP = ip
}

// UpdateLoginInfo 更新登录信息
func (u *User) UpdateLoginInfo(ip string) {
	u.RecordLogin(ip)
}

// Enable 启用用户
func (u *User) Enable() {
	u.Status = UserStatusActive
}

// Disable 禁用用户
func (u *User) Disable() {
	u.Status = UserStatusInactive
}

// Ban 封禁用户
func (u *User) Ban() {
	u.Status = UserStatusBanned
}

// ChangeRole 修改角色
func (u *User) ChangeRole(role Role) error {
	if !isValidRole(role) {
		return fmt.Errorf("无效的角色: %s", role)
	}
	u.Role = role
	return nil
}

// CanModify 检查是否可以被修改
func (u *User) CanModify(operator *User) bool {
	if operator.IsSuperAdmin() {
		return true
	}
	if u.IsSuperAdmin() {
		return false
	}
	return GetRoleLevel(operator.Role) > GetRoleLevel(u.Role)
}

// GetPublicProfile 获取公开资料
func (u *User) GetPublicProfile() *valueobject.UserProfile {
	return &valueobject.UserProfile{
		ID:          u.ID,
		Nickname:    u.Nickname,
		Avatar:      u.Avatar,
		Role:        string(u.Role),
		OrgID:       u.OrgID,
		LastLoginAt: u.LastLoginAt,
	}
}

// GetFullProfile 获取完整资料
func (u *User) GetFullProfile() *valueobject.UserFullProfile {
	return &valueobject.UserFullProfile{
		UserProfile:  *u.GetPublicProfile(),
		Phone:        u.Phone,
		Email:        u.Email,
		RealName:     u.RealName,
		IDCard:       u.IDCard,
		Gender:       u.Gender,
		Address:      u.Address,
		Emergency:    u.Emergency,
		EmergencyTel: u.EmergencyTel,
		Introduction: u.Introduction,
		Status:       string(u.Status),
		CreatedAt:    u.CreatedAt,
	}
}

// isValidRole 检查角色是否有效
func isValidRole(role Role) bool {
	switch role {
	case RoleSuperAdmin, RoleAdmin, RoleManager, RoleVolunteer:
		return true
	default:
		return false
	}
}

// Permission 权限
type Permission struct {
	ID          string `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Code        string `gorm:"size:100;uniqueIndex;not null" json:"code"`
	Description string `gorm:"size:255" json:"description,omitempty"`
	Resource    string `gorm:"size:100;not null" json:"resource"`
	Action      string `gorm:"size:50;not null" json:"action"`
	BaseEntity  `json:"-"`
}

// TableName 表名
func (Permission) TableName() string {
	return "ty_permissions"
}

// UserPermission 用户权限关联
type UserPermission struct {
	UserID       string    `gorm:"type:uuid;primaryKey" json:"user_id"`
	PermissionID string    `gorm:"type:uuid;primaryKey" json:"permission_id"`
	GrantedAt    time.Time `json:"granted_at"`
	GrantedBy    string    `gorm:"type:uuid" json:"granted_by"`
}

// TableName 表名
func (UserPermission) TableName() string {
	return "ty_user_permissions"
}

// NewUser 创建新用户
func NewUser(nickname, phone, orgID string, role Role) (*User, error) {
	user := &User{
		BaseEntity: BaseEntity{
			ID: uuid.New().String(),
		},
		Nickname: nickname,
		Phone:    phone,
		OrgID:    orgID,
		Role:     role,
		Status:   UserStatusActive,
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	return user, nil
}

// NewSuperAdmin 创建超级管理员
func NewSuperAdmin(nickname, phone, password string) (*User, error) {
	user := &User{
		BaseEntity: BaseEntity{
			ID: uuid.MustParse("00000000-0000-0000-0000-000000000001").String(),
		},
		Nickname: nickname,
		Phone:    phone,
		OrgID:    uuid.MustParse("00000000-0000-0000-0000-000000000000").String(),
		Role:     RoleSuperAdmin,
		Status:   UserStatusActive,
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	if err := user.SetPassword(password); err != nil {
		return nil, err
	}

	return user, nil
}
