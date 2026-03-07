package repository

import (
	"context"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"gorm.io/gorm"
)

// NotificationRepositoryImpl 通知仓储实现
type NotificationRepositoryImpl struct {
	db *gorm.DB
}

// NewNotificationRepository 创建通知仓储
func NewNotificationRepository(db *gorm.DB) repository.NotificationRepository {
	return &NotificationRepositoryImpl{db: db}
}

// Create 创建通知
func (r *NotificationRepositoryImpl) Create(ctx context.Context, notification *entity.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

// CreateBatch 批量创建通知
func (r *NotificationRepositoryImpl) CreateBatch(ctx context.Context, notifications []*entity.Notification) error {
	return r.db.WithContext(ctx).CreateInBatches(notifications, 100).Error
}

// FindByID 根据ID查找通知
func (r *NotificationRepositoryImpl) FindByID(ctx context.Context, id string) (*entity.Notification, error) {
	var notification entity.Notification
	err := r.db.WithContext(ctx).First(&notification, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// Update 更新通知
func (r *NotificationRepositoryImpl) Update(ctx context.Context, notification *entity.Notification) error {
	return r.db.WithContext(ctx).Save(notification).Error
}

// UpdateStatus 更新状态
func (r *NotificationRepositoryImpl) UpdateStatus(ctx context.Context, id string, status entity.NotificationStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}
	
	if status == entity.NotificationStatusRead {
		now := time.Now()
		updates["read_at"] = &now
	}
	
	return r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// MarkAsRead 标记为已读
func (r *NotificationRepositoryImpl) MarkAsRead(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":  entity.NotificationStatusRead,
			"read_at": &now,
		}).Error
}

// MarkAllAsRead 标记所有为已读
func (r *NotificationRepositoryImpl) MarkAllAsRead(ctx context.Context, userID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("to_user_id = ? AND status = ?", userID, entity.NotificationStatusUnread).
		Updates(map[string]interface{}{
			"status":  entity.NotificationStatusRead,
			"read_at": &now,
		}).Error
}

// MarkAsArchived 归档通知
func (r *NotificationRepositoryImpl) MarkAsArchived(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      entity.NotificationStatusArchived,
			"archived_at": &now,
		}).Error
}

// Delete 删除通知（软删除）
func (r *NotificationRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&entity.Notification{}).Error
}

// List 查询通知列表
func (r *NotificationRepositoryImpl) List(ctx context.Context, query *entity.NotificationQuery) ([]*entity.Notification, int64, error) {
	db := r.db.WithContext(ctx).Model(&entity.Notification{})
	
	// 应用查询条件
	if query.UserID != "" {
		db = db.Where("to_user_id = ?", query.UserID)
	}
	if query.OrgID != "" {
		db = db.Where("org_id = ?", query.OrgID)
	}
	if query.Type != "" {
		db = db.Where("type = ?", query.Type)
	}
	if query.Channel != "" {
		db = db.Where("channel = ?", query.Channel)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.BusinessType != "" {
		db = db.Where("business_type = ?", query.BusinessType)
	}
	if query.UnreadOnly {
		db = db.Where("status = ?", entity.NotificationStatusUnread)
	}
	if query.StartTime != nil {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("created_at <= ?", query.EndTime)
	}
	
	// 排除已删除的
	db = db.Where("deleted_at IS NULL")
	
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 分页
	page := query.Page
	if page < 1 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	
	var notifications []*entity.Notification
	err := db.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&notifications).Error
	
	return notifications, total, err
}

// GetUnreadList 获取未读通知列表
func (r *NotificationRepositoryImpl) GetUnreadList(ctx context.Context, userID string, limit int) ([]*entity.Notification, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	
	var notifications []*entity.Notification
	err := r.db.WithContext(ctx).
		Where("to_user_id = ? AND status = ?", userID, entity.NotificationStatusUnread).
		Order("priority DESC, created_at DESC").
		Limit(limit).
		Find(&notifications).Error
	
	return notifications, err
}

// CountUnread 统计未读数量
func (r *NotificationRepositoryImpl) CountUnread(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("to_user_id = ? AND status = ?", userID, entity.NotificationStatusUnread).
		Count(&count).Error
	return count, err
}

// GetStats 获取通知统计
func (r *NotificationRepositoryImpl) GetStats(ctx context.Context, userID string) (*entity.NotificationStats, error) {
	stats := &entity.NotificationStats{}
	
	// 总数量
	if err := r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("to_user_id = ? AND deleted_at IS NULL", userID).
		Count(&stats.TotalCount).Error; err != nil {
		return nil, err
	}
	
	// 未读数量
	if err := r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("to_user_id = ? AND status = ?", userID, entity.NotificationStatusUnread).
		Count(&stats.UnreadCount).Error; err != nil {
		return nil, err
	}
	
	// 已读数量
	if err := r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("to_user_id = ? AND status = ?", userID, entity.NotificationStatusRead).
		Count(&stats.ReadCount).Error; err != nil {
		return nil, err
	}
	
	// 各类型统计
	typeCounts := make(map[string]int64)
	rows, err := r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Select("type, COUNT(*) as count").
		Where("to_user_id = ? AND status = ?", userID, entity.NotificationStatusUnread).
		Group("type").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var t string
		var c int64
		if err := rows.Scan(&t, &c); err == nil {
			typeCounts[t] = c
		}
	}
	
	stats.SystemCount = typeCounts[string(entity.NotificationTypeSystem)]
	stats.TaskCount = typeCounts[string(entity.NotificationTypeTask)]
	stats.WorkflowCount = typeCounts[string(entity.NotificationTypeWorkflow)]
	stats.AlertCount = typeCounts[string(entity.NotificationTypeAlert)]
	
	return stats, nil
}

// GetRecentNotifications 获取最近通知
func (r *NotificationRepositoryImpl) GetRecentNotifications(ctx context.Context, userID string, days int) ([]*entity.Notification, error) {
	if days <= 0 {
		days = 7
	}
	
	startTime := time.Now().AddDate(0, 0, -days)
	
	var notifications []*entity.Notification
	err := r.db.WithContext(ctx).
		Where("to_user_id = ? AND created_at >= ?", userID, startTime).
		Order("created_at DESC").
		Find(&notifications).Error
	
	return notifications, err
}

// NotificationSettingRepositoryImpl 通知设置仓储实现
type NotificationSettingRepositoryImpl struct {
	db *gorm.DB
}

// NewNotificationSettingRepository 创建通知设置仓储
func NewNotificationSettingRepository(db *gorm.DB) repository.NotificationSettingRepository {
	return &NotificationSettingRepositoryImpl{db: db}
}

// GetByUserID 获取用户通知设置
func (r *NotificationSettingRepositoryImpl) GetByUserID(ctx context.Context, userID string) (*entity.NotificationSetting, error) {
	var setting entity.NotificationSetting
	err := r.db.WithContext(ctx).First(&setting, "user_id = ?", userID).Error
	if err == gorm.ErrRecordNotFound {
		// 返回默认设置
		return &entity.NotificationSetting{
			UserID:           userID,
			WebSocketEnabled: true,
			PushEnabled:      true,
			SMSEnabled:       false,
			EmailEnabled:     false,
			InAppEnabled:     true,
			SystemEnabled:    true,
			TaskEnabled:      true,
			WorkflowEnabled:  true,
			AlertEnabled:     true,
			DoNotDisturb:     false,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

// CreateOrUpdate 创建或更新设置
func (r *NotificationSettingRepositoryImpl) CreateOrUpdate(ctx context.Context, setting *entity.NotificationSetting) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", setting.UserID).
		Assign(setting).
		FirstOrCreate(setting).Error
}

// UpdateChannel 更新渠道设置
func (r *NotificationSettingRepositoryImpl) UpdateChannel(ctx context.Context, userID string, channel entity.NotificationChannel, enabled bool) error {
	field := ""
	switch channel {
	case entity.NotificationChannelWebSocket:
		field = "web_socket_enabled"
	case entity.NotificationChannelPush:
		field = "push_enabled"
	case entity.NotificationChannelSMS:
		field = "sms_enabled"
	case entity.NotificationChannelEmail:
		field = "email_enabled"
	case entity.NotificationChannelInApp:
		field = "in_app_enabled"
	}
	
	if field == "" {
		return nil
	}
	
	return r.db.WithContext(ctx).
		Model(&entity.NotificationSetting{}).
		Where("user_id = ?", userID).
		Update(field, enabled).Error
}

// UpdateType 更新类型设置
func (r *NotificationSettingRepositoryImpl) UpdateType(ctx context.Context, userID string, notifType entity.NotificationType, enabled bool) error {
	field := ""
	switch notifType {
	case entity.NotificationTypeSystem:
		field = "system_enabled"
	case entity.NotificationTypeTask:
		field = "task_enabled"
	case entity.NotificationTypeWorkflow:
		field = "workflow_enabled"
	case entity.NotificationTypeAlert:
		field = "alert_enabled"
	}
	
	if field == "" {
		return nil
	}
	
	return r.db.WithContext(ctx).
		Model(&entity.NotificationSetting{}).
		Where("user_id = ?", userID).
		Update(field, enabled).Error
}

// UpdateDoNotDisturb 更新免打扰设置
func (r *NotificationSettingRepositoryImpl) UpdateDoNotDisturb(ctx context.Context, userID string, enabled bool, from, to *time.Time) error {
	return r.db.WithContext(ctx).
		Model(&entity.NotificationSetting{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"do_not_disturb":       enabled,
			"do_not_disturb_from":  from,
			"do_not_disturb_to":    to,
		}).Error
}

// MessageTemplateRepositoryImpl 消息模板仓储实现
type MessageTemplateRepositoryImpl struct {
	db *gorm.DB
}

// NewMessageTemplateRepository 创建消息模板仓储
func NewMessageTemplateRepository(db *gorm.DB) repository.MessageTemplateRepository {
	return &MessageTemplateRepositoryImpl{db: db}
}

// Create 创建模板
func (r *MessageTemplateRepositoryImpl) Create(ctx context.Context, template *entity.MessageTemplate) error {
	return r.db.WithContext(ctx).Create(template).Error
}

// Update 更新模板
func (r *MessageTemplateRepositoryImpl) Update(ctx context.Context, template *entity.MessageTemplate) error {
	return r.db.WithContext(ctx).Save(template).Error
}

// FindByID 根据ID查找
func (r *MessageTemplateRepositoryImpl) FindByID(ctx context.Context, id string) (*entity.MessageTemplate, error) {
	var template entity.MessageTemplate
	err := r.db.WithContext(ctx).First(&template, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// FindByCode 根据编码查找
func (r *MessageTemplateRepositoryImpl) FindByCode(ctx context.Context, code string, channel entity.NotificationChannel) (*entity.MessageTemplate, error) {
	var template entity.MessageTemplate
	query := r.db.WithContext(ctx).Where("code = ? AND status = ?", code, "active")
	if channel != "" {
		query = query.Where("channel = ?", channel)
	}
	err := query.First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// List 查询模板列表
func (r *MessageTemplateRepositoryImpl) List(ctx context.Context, channel entity.NotificationChannel, notifType entity.NotificationType, page, pageSize int) ([]*entity.MessageTemplate, int64, error) {
	db := r.db.WithContext(ctx).Model(&entity.MessageTemplate{})
	
	if channel != "" {
		db = db.Where("channel = ?", channel)
	}
	if notifType != "" {
		db = db.Where("type = ?", notifType)
	}
	
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	
	var templates []*entity.MessageTemplate
	err := db.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&templates).Error
	
	return templates, total, err
}

// Delete 删除模板
func (r *MessageTemplateRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.MessageTemplate{}, "id = ?", id).Error
}

// GetActiveByType 获取指定类型的活跃模板
func (r *MessageTemplateRepositoryImpl) GetActiveByType(ctx context.Context, notifType entity.NotificationType, channel entity.NotificationChannel) ([]*entity.MessageTemplate, error) {
	var templates []*entity.MessageTemplate
	query := r.db.WithContext(ctx).
		Where("type = ? AND status = ?", notifType, "active")
	
	if channel != "" {
		query = query.Where("channel = ?", channel)
	}
	
	err := query.Find(&templates).Error
	return templates, err
}
