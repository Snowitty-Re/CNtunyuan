package entity

import (
	"time"

	"github.com/google/uuid"
)

// AuthzDecisionLog 权限决策审计日志
type AuthzDecisionLog struct {
	BaseEntity
	OperatorID    string `gorm:"type:uuid;index" json:"operator_id,omitempty"`
	OperatorRole  string `gorm:"size:20;index" json:"operator_role,omitempty"`
	OperatorOrgID string `gorm:"type:uuid;index" json:"operator_org_id,omitempty"`
	Action        string `gorm:"size:100;not null;index" json:"action"`
	ResourceType  string `gorm:"size:50;not null;index" json:"resource_type"`
	ResourceID    string `gorm:"size:100;index" json:"resource_id,omitempty"`
	Allowed       bool   `gorm:"not null;index" json:"allowed"`
	Reason        string `gorm:"size:50;not null;index" json:"reason"`
	TraceID       string `gorm:"size:100;index" json:"trace_id,omitempty"`
}

// TableName 表名
func (AuthzDecisionLog) TableName() string {
	return "ty_authz_decisions"
}

// NewAuthzDecisionLog 创建权限决策日志
func NewAuthzDecisionLog() *AuthzDecisionLog {
	return &AuthzDecisionLog{
		BaseEntity: BaseEntity{
			ID:        uuid.New().String(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

// AuthzDecisionQuery 授权决策查询条件
type AuthzDecisionQuery struct {
	Page         int
	PageSize     int
	Action       string
	Allowed      *bool
	OperatorID   string
	ResourceType string
	Reason       string
	StartTime    *time.Time
	EndTime      *time.Time
}

// NewAuthzDecisionQuery 默认查询
func NewAuthzDecisionQuery() *AuthzDecisionQuery {
	return &AuthzDecisionQuery{
		Page:     1,
		PageSize: 20,
	}
}
