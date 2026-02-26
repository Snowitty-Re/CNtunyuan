package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 用户状态
const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
	UserStatusBanned   = "banned"
)

// 用户角色
const (
	RoleSuperAdmin = "super_admin" // 超级管理员
	RoleAdmin      = "admin"       // 管理员
	RoleManager    = "manager"     // 管理者
	RoleVolunteer  = "volunteer"   // 志愿者
)

// User 用户模型
type User struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnionID   string         `gorm:"size:100;uniqueIndex;comment:微信UnionID" json:"union_id"`
	OpenID    string         `gorm:"size:100;uniqueIndex;comment:微信OpenID" json:"open_id"`
	Nickname  string         `gorm:"size:100;comment:昵称" json:"nickname"`
	Avatar    string         `gorm:"size:500;comment:头像" json:"avatar"`
	Phone     string         `gorm:"size:20;comment:手机号" json:"phone"`
	Email     string         `gorm:"size:100;comment:邮箱" json:"email"`
	RealName  string         `gorm:"size:50;comment:真实姓名" json:"real_name"`
	IDCard    string         `gorm:"size:18;comment:身份证号" json:"id_card"`
	Role      string         `gorm:"size:20;default:volunteer;comment:角色" json:"role"`
	Status    string         `gorm:"size:20;default:active;comment:状态" json:"status"`
	OrgID     *uuid.UUID     `gorm:"type:uuid;index;comment:所属机构ID" json:"org_id"`
	Org       *Organization  `gorm:"foreignKey:OrgID" json:"org,omitempty"`
	LastLogin *time.Time     `gorm:"comment:最后登录时间" json:"last_login"`
	LoginIP   string         `gorm:"size:50;comment:登录IP" json:"login_ip"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// UserProfile 用户扩展信息
type UserProfile struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	Gender       string    `gorm:"size:10;comment:性别" json:"gender"`
	BirthDate    *time.Time `gorm:"comment:出生日期" json:"birth_date"`
	Address      string    `gorm:"size:200;comment:地址" json:"address"`
	EmergencyContact string `gorm:"size:50;comment:紧急联系人" json:"emergency_contact"`
	EmergencyPhone   string `gorm:"size:20;comment:紧急联系人电话" json:"emergency_phone"`
	Skills       string    `gorm:"size:500;comment:技能特长" json:"skills"`
	Experience   string    `gorm:"type:text;comment:志愿服务经历" json:"experience"`
	CertificateNo string   `gorm:"size:50;comment:志愿者证书编号" json:"certificate_no"`
	JoinDate     *time.Time `gorm:"comment:加入日期" json:"join_date"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

func (UserProfile) TableName() string {
	return "user_profiles"
}

// IsAdmin 检查是否为管理员
func (u *User) IsAdmin() bool {
	return u.Role == RoleSuperAdmin || u.Role == RoleAdmin
}

// IsManager 检查是否为管理者
func (u *User) IsManager() bool {
	return u.Role == RoleSuperAdmin || u.Role == RoleAdmin || u.Role == RoleManager
}
