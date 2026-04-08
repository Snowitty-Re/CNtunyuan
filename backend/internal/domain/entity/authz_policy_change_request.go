package entity

import (
	"time"

	"github.com/google/uuid"
)

// AuthzPolicyRequestType 策略变更申请类型
type AuthzPolicyRequestType string

const (
	AuthzPolicyRequestRolePermissions AuthzPolicyRequestType = "role_permissions"
	AuthzPolicyRequestPolicyRules     AuthzPolicyRequestType = "policy_rules"
	AuthzPolicyRequestRollback        AuthzPolicyRequestType = "rollback"
)

// AuthzPolicyRequestStatus 策略变更申请状态
type AuthzPolicyRequestStatus string

const (
	AuthzPolicyRequestPending  AuthzPolicyRequestStatus = "pending"
	AuthzPolicyRequestApproved AuthzPolicyRequestStatus = "approved"
	AuthzPolicyRequestRejected AuthzPolicyRequestStatus = "rejected"
)

// AuthzPolicyScopeType 策略申请作用域
type AuthzPolicyScopeType string

const (
	AuthzPolicyScopeGlobal AuthzPolicyScopeType = "global"
	AuthzPolicyScopeOrg    AuthzPolicyScopeType = "org"
)

// AuthzPolicyChangeRequest 策略变更申请
type AuthzPolicyChangeRequest struct {
	BaseEntity
	RequestType     AuthzPolicyRequestType   `gorm:"size:30;not null;index" json:"request_type"`
	Status          AuthzPolicyRequestStatus `gorm:"size:20;not null;default:pending;index" json:"status"`
	ScopeType       AuthzPolicyScopeType     `gorm:"size:20;not null;default:global;index" json:"scope_type"`
	TargetOrgID     string                   `gorm:"type:uuid;index" json:"target_org_id,omitempty"`
	TargetKey       string                   `gorm:"size:120;not null;index" json:"target_key"`
	PayloadJSON     string                   `gorm:"type:text" json:"payload_json,omitempty"`
	PreviewJSON     string                   `gorm:"type:text" json:"preview_json,omitempty"`
	RequestNote     string                   `gorm:"type:text" json:"request_note,omitempty"`
	RequestedBy     string                   `gorm:"type:uuid;index" json:"requested_by,omitempty"`
	RequestedByRole string                   `gorm:"size:20;index" json:"requested_by_role,omitempty"`
	ReviewNote      string                   `gorm:"type:text" json:"review_note,omitempty"`
	ReviewedBy      string                   `gorm:"type:uuid;index" json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time               `json:"reviewed_at,omitempty"`
	Executed        bool                     `gorm:"not null;default:false;index" json:"executed"`
	ExecutedAt      *time.Time               `json:"executed_at,omitempty"`
	ExecutedLogID   string                   `gorm:"type:uuid;index" json:"executed_log_id,omitempty"`
	TraceID         string                   `gorm:"size:100;index" json:"trace_id,omitempty"`
}

// TableName 表名
func (AuthzPolicyChangeRequest) TableName() string {
	return "ty_authz_policy_change_requests"
}

// NewAuthzPolicyChangeRequest 创建申请
func NewAuthzPolicyChangeRequest() *AuthzPolicyChangeRequest {
	now := time.Now()
	return &AuthzPolicyChangeRequest{
		BaseEntity: BaseEntity{
			ID:        uuid.New().String(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		Status:    AuthzPolicyRequestPending,
		ScopeType: AuthzPolicyScopeGlobal,
	}
}

// AuthzPolicyChangeRequestQuery 申请查询条件
type AuthzPolicyChangeRequestQuery struct {
	Page        int
	PageSize    int
	RequestType string
	Status      string
	ScopeType   string
	RequestedBy string
	TargetOrgID string
	TargetKey   string
	StartTime   *time.Time
	EndTime     *time.Time
}

// NewAuthzPolicyChangeRequestQuery 默认查询
func NewAuthzPolicyChangeRequestQuery() *AuthzPolicyChangeRequestQuery {
	return &AuthzPolicyChangeRequestQuery{
		Page:     1,
		PageSize: 20,
	}
}
