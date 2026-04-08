package entity

import (
	"time"

	"github.com/google/uuid"
)

// AuthzPolicyChangeType 权限策略变更类型
type AuthzPolicyChangeType string

const (
	AuthzPolicyChangeRolePermissions AuthzPolicyChangeType = "role_permissions"
	AuthzPolicyChangePolicyRules     AuthzPolicyChangeType = "policy_rules"
)

// AuthzPolicyOperationType 策略变更操作类型
type AuthzPolicyOperationType string

const (
	AuthzPolicyOpApply    AuthzPolicyOperationType = "apply"
	AuthzPolicyOpRollback AuthzPolicyOperationType = "rollback"
)

// AuthzPolicyChangeLog 权限策略变更审计日志
type AuthzPolicyChangeLog struct {
	BaseEntity
	OperatorID   string                   `gorm:"type:uuid;index" json:"operator_id,omitempty"`
	OperatorRole string                   `gorm:"size:20;index" json:"operator_role,omitempty"`
	Operation    AuthzPolicyOperationType `gorm:"size:20;not null;default:apply;index" json:"operation"`
	ChangeType   AuthzPolicyChangeType    `gorm:"size:30;not null;index" json:"change_type"`
	TargetKey    string                   `gorm:"size:120;not null;index" json:"target_key"`
	RollbackOfID string                   `gorm:"type:uuid;index" json:"rollback_of_id,omitempty"`
	BeforeJSON   string                   `gorm:"type:text" json:"before_json,omitempty"`
	AfterJSON    string                   `gorm:"type:text" json:"after_json,omitempty"`
	TraceID      string                   `gorm:"size:100;index" json:"trace_id,omitempty"`
}

// TableName 表名
func (AuthzPolicyChangeLog) TableName() string {
	return "ty_authz_policy_changes"
}

// NewAuthzPolicyChangeLog 创建策略变更日志
func NewAuthzPolicyChangeLog() *AuthzPolicyChangeLog {
	now := time.Now()
	return &AuthzPolicyChangeLog{
		BaseEntity: BaseEntity{
			ID:        uuid.New().String(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		Operation: AuthzPolicyOpApply,
	}
}

// AuthzPolicyChangeQuery 策略变更查询条件
type AuthzPolicyChangeQuery struct {
	Page         int
	PageSize     int
	Operation    string
	ChangeType   string
	TargetKey    string
	RollbackOfID string
	OperatorID   string
	StartTime    *time.Time
	EndTime      *time.Time
}

// NewAuthzPolicyChangeQuery 默认查询
func NewAuthzPolicyChangeQuery() *AuthzPolicyChangeQuery {
	return &AuthzPolicyChangeQuery{
		Page:     1,
		PageSize: 20,
	}
}
