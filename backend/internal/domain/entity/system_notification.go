package entity

import (
	"time"

	"github.com/google/uuid"
)

// SystemNotificationCategory 站内通知分类
type SystemNotificationCategory string

const (
	SystemNotificationCategoryAuthz SystemNotificationCategory = "authz"
)

// SystemNotificationStatus 通知状态
type SystemNotificationStatus string

const (
	SystemNotificationStatusUnread SystemNotificationStatus = "unread"
	SystemNotificationStatusRead   SystemNotificationStatus = "read"
)

// SystemNotification 站内通知
type SystemNotification struct {
	BaseEntity
	Category      SystemNotificationCategory `gorm:"size:40;not null;index" json:"category"`
	Title         string                     `gorm:"size:200;not null" json:"title"`
	Content       string                     `gorm:"type:text" json:"content"`
	RecipientID   string                     `gorm:"type:uuid;index" json:"recipient_id,omitempty"`
	RecipientRole string                     `gorm:"size:20;index" json:"recipient_role,omitempty"`
	Status        SystemNotificationStatus   `gorm:"size:20;not null;default:unread;index" json:"status"`
	RelatedType   string                     `gorm:"size:60;index" json:"related_type,omitempty"`
	RelatedID     string                     `gorm:"size:80;index" json:"related_id,omitempty"`
	OperatorID    string                     `gorm:"type:uuid;index" json:"operator_id,omitempty"`
	ReadAt        *time.Time                 `json:"read_at,omitempty"`
}

func (SystemNotification) TableName() string {
	return "ty_system_notifications"
}

func NewSystemNotification() *SystemNotification {
	now := time.Now()
	return &SystemNotification{
		BaseEntity: BaseEntity{
			ID:        uuid.New().String(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		Status: SystemNotificationStatusUnread,
	}
}

// SystemNotificationQuery 通知查询
type SystemNotificationQuery struct {
	Page        int
	PageSize    int
	Status      string
	Category    string
	UserID      string
	UserRole    string
	RelatedType string
	RelatedID   string
}

func NewSystemNotificationQuery() *SystemNotificationQuery {
	return &SystemNotificationQuery{Page: 1, PageSize: 20}
}
