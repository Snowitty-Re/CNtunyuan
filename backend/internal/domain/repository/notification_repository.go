package repository

import (
	"context"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
)

// NotificationRepository 通知仓储接口
type NotificationRepository interface {
	// Create 创建通知
	Create(ctx context.Context, notification *entity.Notification) error
	
	// CreateBatch 批量创建通知
	CreateBatch(ctx context.Context, notifications []*entity.Notification) error
	
	// FindByID 根据ID查找通知
	FindByID(ctx context.Context, id string) (*entity.Notification, error)
	
	// Update 更新通知
	Update(ctx context.Context, notification *entity.Notification) error
	
	// UpdateStatus 更新状态
	UpdateStatus(ctx context.Context, id string, status entity.NotificationStatus) error
	
	// MarkAsRead 标记为已读
	MarkAsRead(ctx context.Context, id string) error
	
	// MarkAllAsRead 标记所有为已读
	MarkAllAsRead(ctx context.Context, userID string) error
	
	// MarkAsArchived 归档通知
	MarkAsArchived(ctx context.Context, id string) error
	
	// Delete 删除通知（软删除）
	Delete(ctx context.Context, id string) error
	
	// List 查询通知列表
	List(ctx context.Context, query *entity.NotificationQuery) ([]*entity.Notification, int64, error)
	
	// GetUnreadList 获取未读通知列表
	GetUnreadList(ctx context.Context, userID string, limit int) ([]*entity.Notification, error)
	
	// CountUnread 统计未读数量
	CountUnread(ctx context.Context, userID string) (int64, error)
	
	// GetStats 获取通知统计
	GetStats(ctx context.Context, userID string) (*entity.NotificationStats, error)
	
	// GetRecentNotifications 获取最近通知
	GetRecentNotifications(ctx context.Context, userID string, days int) ([]*entity.Notification, error)
}

// NotificationSettingRepository 通知设置仓储接口
type NotificationSettingRepository interface {
	// GetByUserID 获取用户通知设置
	GetByUserID(ctx context.Context, userID string) (*entity.NotificationSetting, error)
	
	// CreateOrUpdate 创建或更新设置
	CreateOrUpdate(ctx context.Context, setting *entity.NotificationSetting) error
	
	// UpdateChannel 更新渠道设置
	UpdateChannel(ctx context.Context, userID string, channel entity.NotificationChannel, enabled bool) error
	
	// UpdateType 更新类型设置
	UpdateType(ctx context.Context, userID string, notifType entity.NotificationType, enabled bool) error
	
	// UpdateDoNotDisturb 更新免打扰设置
	UpdateDoNotDisturb(ctx context.Context, userID string, enabled bool, from, to *time.Time) error
}

// MessageTemplateRepository 消息模板仓储接口
type MessageTemplateRepository interface {
	// Create 创建模板
	Create(ctx context.Context, template *entity.MessageTemplate) error
	
	// Update 更新模板
	Update(ctx context.Context, template *entity.MessageTemplate) error
	
	// FindByID 根据ID查找
	FindByID(ctx context.Context, id string) (*entity.MessageTemplate, error)
	
	// FindByCode 根据编码查找
	FindByCode(ctx context.Context, code string, channel entity.NotificationChannel) (*entity.MessageTemplate, error)
	
	// List 查询模板列表
	List(ctx context.Context, channel entity.NotificationChannel, notifType entity.NotificationType, page, pageSize int) ([]*entity.MessageTemplate, int64, error)
	
	// Delete 删除模板
	Delete(ctx context.Context, id string) error
	
	// GetActiveByType 获取指定类型的活跃模板
	GetActiveByType(ctx context.Context, notifType entity.NotificationType, channel entity.NotificationChannel) ([]*entity.MessageTemplate, error)
}
