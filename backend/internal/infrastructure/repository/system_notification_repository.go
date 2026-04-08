package repository

import (
	"context"
	"strings"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	domainrepo "github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"gorm.io/gorm"
)

type SystemNotificationRepositoryImpl struct {
	db *gorm.DB
}

func NewSystemNotificationRepository(db *gorm.DB) domainrepo.SystemNotificationRepository {
	return &SystemNotificationRepositoryImpl{db: db}
}

type systemNotificationRecord struct {
	ID            string         `gorm:"column:id"`
	CreatedAt     time.Time      `gorm:"column:created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at"`
	Category      string         `gorm:"column:category"`
	Title         string         `gorm:"column:title"`
	Content       string         `gorm:"column:content"`
	RecipientID   *string        `gorm:"column:recipient_id"`
	RecipientRole string         `gorm:"column:recipient_role"`
	Status        string         `gorm:"column:status"`
	RelatedType   string         `gorm:"column:related_type"`
	RelatedID     string         `gorm:"column:related_id"`
	OperatorID    *string        `gorm:"column:operator_id"`
	ReadAt        *time.Time     `gorm:"column:read_at"`
}

func (systemNotificationRecord) TableName() string {
	return "ty_system_notifications"
}

func (r *SystemNotificationRepositoryImpl) Create(ctx context.Context, notification *entity.SystemNotification) error {
	if notification == nil {
		return nil
	}
	record := toSystemNotificationRecord(notification)
	return r.db.WithContext(ctx).Create(&record).Error
}

func (r *SystemNotificationRepositoryImpl) CreateBatch(ctx context.Context, notifications []*entity.SystemNotification) error {
	if len(notifications) == 0 {
		return nil
	}
	records := make([]systemNotificationRecord, 0, len(notifications))
	for _, item := range notifications {
		if item == nil {
			continue
		}
		records = append(records, toSystemNotificationRecord(item))
	}
	if len(records) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&records).Error
}

func (r *SystemNotificationRepositoryImpl) List(ctx context.Context, query *entity.SystemNotificationQuery) (*domainrepo.SystemNotificationPaginatedResult, error) {
	if query == nil {
		query = entity.NewSystemNotificationQuery()
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}

	db := r.db.WithContext(ctx).Model(&systemNotificationRecord{})
	db = applyNotificationRecipientScope(db, strings.TrimSpace(query.UserID), strings.TrimSpace(query.UserRole))

	if status := strings.TrimSpace(query.Status); status != "" {
		db = db.Where("status = ?", status)
	}
	if category := strings.TrimSpace(query.Category); category != "" {
		db = db.Where("category = ?", category)
	}
	if relatedType := strings.TrimSpace(query.RelatedType); relatedType != "" {
		db = db.Where("related_type = ?", relatedType)
	}
	if relatedID := strings.TrimSpace(query.RelatedID); relatedID != "" {
		db = db.Where("related_id = ?", relatedID)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (query.Page - 1) * query.PageSize
	var records []systemNotificationRecord
	if err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&records).Error; err != nil {
		return nil, err
	}

	list := make([]entity.SystemNotification, 0, len(records))
	for _, rec := range records {
		list = append(list, fromSystemNotificationRecord(rec))
	}

	return &domainrepo.SystemNotificationPaginatedResult{
		List:     list,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

func (r *SystemNotificationRepositoryImpl) MarkRead(ctx context.Context, id, userID, userRole string) (bool, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return false, nil
	}
	updates := map[string]interface{}{
		"status":  string(entity.SystemNotificationStatusRead),
		"read_at": time.Now(),
	}
	tx := applyNotificationRecipientScope(
		r.db.WithContext(ctx).
			Model(&systemNotificationRecord{}).
			Where("id = ?", id).
			Where("status = ?", string(entity.SystemNotificationStatusUnread)),
		strings.TrimSpace(userID),
		strings.TrimSpace(userRole),
	).
		Updates(updates)
	if tx.Error != nil {
		return false, tx.Error
	}
	return tx.RowsAffected > 0, nil
}

func (r *SystemNotificationRepositoryImpl) MarkAllRead(ctx context.Context, userID, userRole string) (int64, error) {
	updates := map[string]interface{}{
		"status":  string(entity.SystemNotificationStatusRead),
		"read_at": time.Now(),
	}
	tx := applyNotificationRecipientScope(
		r.db.WithContext(ctx).
			Model(&systemNotificationRecord{}).
			Where("status = ?", string(entity.SystemNotificationStatusUnread)),
		strings.TrimSpace(userID),
		strings.TrimSpace(userRole),
	).
		Updates(updates)
	if tx.Error != nil {
		return 0, tx.Error
	}
	return tx.RowsAffected, nil
}

func (r *SystemNotificationRepositoryImpl) CountUnread(ctx context.Context, userID, userRole string) (int64, error) {
	var total int64
	err := applyNotificationRecipientScope(
		r.db.WithContext(ctx).
			Model(&systemNotificationRecord{}).
			Where("status = ?", string(entity.SystemNotificationStatusUnread)),
		strings.TrimSpace(userID),
		strings.TrimSpace(userRole),
	).
		Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

func toSystemNotificationRecord(item *entity.SystemNotification) systemNotificationRecord {
	return systemNotificationRecord{
		ID:            item.ID,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
		DeletedAt:     item.DeletedAt,
		Category:      string(item.Category),
		Title:         item.Title,
		Content:       item.Content,
		RecipientID:   nullableRecipientUUID(item.RecipientID),
		RecipientRole: item.RecipientRole,
		Status:        string(item.Status),
		RelatedType:   item.RelatedType,
		RelatedID:     item.RelatedID,
		OperatorID:    nullableRecipientUUID(item.OperatorID),
		ReadAt:        item.ReadAt,
	}
}

func fromSystemNotificationRecord(rec systemNotificationRecord) entity.SystemNotification {
	return entity.SystemNotification{
		BaseEntity: entity.BaseEntity{
			ID:        rec.ID,
			CreatedAt: rec.CreatedAt,
			UpdatedAt: rec.UpdatedAt,
			DeletedAt: rec.DeletedAt,
		},
		Category:      entity.SystemNotificationCategory(rec.Category),
		Title:         rec.Title,
		Content:       rec.Content,
		RecipientID:   derefRecipientUUID(rec.RecipientID),
		RecipientRole: rec.RecipientRole,
		Status:        entity.SystemNotificationStatus(rec.Status),
		RelatedType:   rec.RelatedType,
		RelatedID:     rec.RelatedID,
		OperatorID:    derefRecipientUUID(rec.OperatorID),
		ReadAt:        rec.ReadAt,
	}
}

func applyNotificationRecipientScope(db *gorm.DB, userID, userRole string) *gorm.DB {
	scope := db.Where("recipient_id IS NULL AND (recipient_role IS NULL OR recipient_role = '')")
	if userID != "" {
		scope = scope.Or("recipient_id = ?", userID)
	}
	if userRole != "" {
		scope = scope.Or("recipient_role = ?", userRole)
	}
	return scope
}

func nullableRecipientUUID(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func derefRecipientUUID(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
