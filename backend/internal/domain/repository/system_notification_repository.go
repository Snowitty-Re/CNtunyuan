package repository

import (
	"context"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
)

// SystemNotificationRepository 站内通知仓储
type SystemNotificationRepository interface {
	Create(ctx context.Context, notification *entity.SystemNotification) error
	CreateBatch(ctx context.Context, notifications []*entity.SystemNotification) error
	List(ctx context.Context, query *entity.SystemNotificationQuery) (*SystemNotificationPaginatedResult, error)
	MarkRead(ctx context.Context, id, userID, userRole string) (bool, error)
	MarkAllRead(ctx context.Context, userID, userRole string) (int64, error)
	CountUnread(ctx context.Context, userID, userRole string) (int64, error)
}

type SystemNotificationPaginatedResult struct {
	List     []entity.SystemNotification `json:"list"`
	Total    int64                       `json:"total"`
	Page     int                         `json:"page"`
	PageSize int                         `json:"page_size"`
}
