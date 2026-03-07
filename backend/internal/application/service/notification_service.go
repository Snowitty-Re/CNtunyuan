package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/websocket"
	"github.com/Snowitty-Re/CNtunyuan/pkg/errors"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// NotificationAppService 通知应用服务
type NotificationAppService struct {
	notifRepo   repository.NotificationRepository
	settingRepo repository.NotificationSettingRepository
	templateRepo repository.MessageTemplateRepository
	wsManager   *websocket.Manager
}

// NewNotificationAppService 创建通知应用服务
func NewNotificationAppService(
	notifRepo repository.NotificationRepository,
	settingRepo repository.NotificationSettingRepository,
	templateRepo repository.MessageTemplateRepository,
	wsManager *websocket.Manager,
) *NotificationAppService {
	return &NotificationAppService{
		notifRepo:    notifRepo,
		settingRepo:  settingRepo,
		templateRepo: templateRepo,
		wsManager:    wsManager,
	}
}

// SendNotification 发送通知
func (s *NotificationAppService) SendNotification(ctx context.Context, req *dto.SendNotificationRequest) error {
	// 创建通知实体
	notification := &entity.Notification{
		BaseEntity: entity.BaseEntity{
			ID:        uuid.New().String(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Title:        req.Title,
		Content:      req.Content,
		Type:         req.Type,
		Channel:      req.Channel,
		Priority:     req.Priority,
		Status:       entity.NotificationStatusUnread,
		ToUserID:     req.ToUserID,
		FromUserID:   req.FromUserID,
		OrgID:        req.OrgID,
		BusinessType: req.BusinessType,
		BusinessID:   req.BusinessID,
		Data:         req.Data,
		ActionURL:    req.ActionURL,
		ActionText:   req.ActionText,
	}
	
	if req.ExpireAt != nil {
		notification.ExpireAt = req.ExpireAt
	}
	
	// 保存到数据库
	if err := s.notifRepo.Create(ctx, notification); err != nil {
		return errors.Wrap(err, errors.CodeInternal, "创建通知失败")
	}
	
	// 异步发送
	go s.dispatchNotification(notification)
	
	return nil
}

// SendBatchNotifications 批量发送通知
func (s *NotificationAppService) SendBatchNotifications(ctx context.Context, req *dto.BatchSendNotificationRequest) error {
	notifications := make([]*entity.Notification, 0, len(req.ToUserIDs))
	
	now := time.Now()
	for _, userID := range req.ToUserIDs {
		notification := &entity.Notification{
			BaseEntity: entity.BaseEntity{
				ID:        uuid.New().String(),
				CreatedAt: now,
				UpdatedAt: now,
			},
			Title:        req.Title,
			Content:      req.Content,
			Type:         req.Type,
			Channel:      req.Channel,
			Priority:     req.Priority,
			Status:       entity.NotificationStatusUnread,
			ToUserID:     userID,
			FromUserID:   req.FromUserID,
			OrgID:        req.OrgID,
			BusinessType: req.BusinessType,
			BusinessID:   req.BusinessID,
			Data:         req.Data,
			ActionURL:    req.ActionURL,
			ActionText:   req.ActionText,
		}
		notifications = append(notifications, notification)
	}
	
	// 批量保存
	if err := s.notifRepo.CreateBatch(ctx, notifications); err != nil {
		return errors.Wrap(err, errors.CodeInternal, "批量创建通知失败")
	}
	
	// 异步发送
	for _, notification := range notifications {
		go s.dispatchNotification(notification)
	}
	
	return nil
}

// SendNotificationWithTemplate 使用模板发送通知
func (s *NotificationAppService) SendNotificationWithTemplate(ctx context.Context, req *dto.SendTemplateNotificationRequest) error {
	// 查找模板
	template, err := s.templateRepo.FindByCode(ctx, req.TemplateCode, req.Channel)
	if err != nil {
		return errors.Wrap(err, errors.CodeNotFound, "模板不存在")
	}
	
	if !template.IsActive() {
		return errors.New(errors.CodeInvalidParam, "模板未激活")
	}
	
	// 渲染模板
	content, err := template.Render(req.Variables)
	if err != nil {
		return errors.Wrap(err, errors.CodeInternal, "模板渲染失败")
	}
	
	subject, _ := template.RenderSubject(req.Variables)
	
	// 使用模板数据创建通知
	sendReq := &dto.SendNotificationRequest{
		Title:        subject,
		Content:      content,
		Type:         template.Type,
		Channel:      template.Channel,
		Priority:     req.Priority,
		ToUserID:     req.ToUserID,
		FromUserID:   req.FromUserID,
		OrgID:        req.OrgID,
		BusinessType: req.BusinessType,
		BusinessID:   req.BusinessID,
		Data:         req.Variables,
		ActionURL:    req.ActionURL,
		ActionText:   req.ActionText,
		ExpireAt:     req.ExpireAt,
	}
	
	return s.SendNotification(ctx, sendReq)
}

// dispatchNotification 分发通知到各渠道
func (s *NotificationAppService) dispatchNotification(notification *entity.Notification) {
	ctx := context.Background()
	
	// 获取用户设置
	setting, err := s.settingRepo.GetByUserID(ctx, notification.ToUserID)
	if err != nil {
		logger.Error("Failed to get notification setting", zap.Error(err), zap.String("user_id", notification.ToUserID))
		// 使用默认设置继续
		setting = &entity.NotificationSetting{
			WebSocketEnabled: true,
			PushEnabled:      true,
			InAppEnabled:     true,
		}
	}
	
	// 检查通知类型是否启用
	if !setting.IsTypeEnabled(notification.Type) {
		logger.Debug("Notification type disabled", zap.String("user_id", notification.ToUserID), zap.String("type", string(notification.Type)))
		return
	}
	
	// 根据渠道发送
	switch notification.Channel {
	case entity.NotificationChannelWebSocket:
		if setting.IsChannelEnabled(entity.NotificationChannelWebSocket) {
			s.sendWebSocket(notification)
		}
	case entity.NotificationChannelInApp:
		if setting.IsChannelEnabled(entity.NotificationChannelInApp) {
			// 站内信只需保存到数据库即可
			s.markAsSent(ctx, notification.ID)
		}
	case entity.NotificationChannelPush:
		if setting.IsChannelEnabled(entity.NotificationChannelPush) {
			s.sendPush(notification)
		}
	case entity.NotificationChannelSMS:
		if setting.IsChannelEnabled(entity.NotificationChannelSMS) {
			s.sendSMS(notification)
		}
	case entity.NotificationChannelEmail:
		if setting.IsChannelEnabled(entity.NotificationChannelEmail) {
			s.sendEmail(notification)
		}
	default:
		// 默认使用WebSocket
		s.sendWebSocket(notification)
	}
}

// sendWebSocket 发送WebSocket通知
func (s *NotificationAppService) sendWebSocket(notification *entity.Notification) {
	msg := &entity.WebSocketMessage{
		ID:         notification.ID,
		Type:       string(notification.Type),
		Title:      notification.Title,
		Content:    notification.Content,
		Data:       notification.Data,
		Timestamp:  notification.CreatedAt,
		FromUserID: "",
	}
	
	if notification.FromUserID != nil {
		msg.FromUserID = *notification.FromUserID
	}
	
	// 添加业务数据
	if msg.Data == nil {
		msg.Data = make(map[string]interface{})
	}
	msg.Data["notification_id"] = notification.ID
	msg.Data["business_type"] = notification.BusinessType
	msg.Data["business_id"] = notification.BusinessID
	msg.Data["action_url"] = notification.ActionURL
	msg.Data["action_text"] = notification.ActionText
	
	if err := s.wsManager.SendToUser(notification.ToUserID, msg); err != nil {
		logger.Error("Failed to send WebSocket message", zap.Error(err), zap.String("user_id", notification.ToUserID))
	} else {
		s.markAsSent(context.Background(), notification.ID)
	}
}

// sendPush 发送App推送（预留接口）
func (s *NotificationAppService) sendPush(notification *entity.Notification) {
	// TODO: 集成极光推送/个推等第三方推送服务
	logger.Debug("Push notification", zap.String("title", notification.Title), zap.String("user_id", notification.ToUserID))
	s.markAsSent(context.Background(), notification.ID)
}

// sendSMS 发送短信（预留接口）
func (s *NotificationAppService) sendSMS(notification *entity.Notification) {
	// TODO: 集成阿里云短信/腾讯云短信等
	logger.Debug("SMS notification", zap.String("content", notification.Content), zap.String("user_id", notification.ToUserID))
	s.markAsSent(context.Background(), notification.ID)
}

// sendEmail 发送邮件（预留接口）
func (s *NotificationAppService) sendEmail(notification *entity.Notification) {
	// TODO: 集成SMTP邮件服务
	logger.Debug("Email notification", zap.String("subject", notification.Title), zap.String("user_id", notification.ToUserID))
	s.markAsSent(context.Background(), notification.ID)
}

// markAsSent 标记通知为已发送
func (s *NotificationAppService) markAsSent(ctx context.Context, notificationID string) {
	now := time.Now()
	err := s.notifRepo.Update(ctx, &entity.Notification{
		BaseEntity: entity.BaseEntity{
			ID:        notificationID,
			UpdatedAt: now,
		},
		SentAt:      &now,
		SentSuccess: true,
	})
	if err != nil {
		logger.Error("Failed to mark notification as sent", zap.Error(err), zap.String("notification_id", notificationID))
	}
}

// GetNotificationList 获取通知列表
func (s *NotificationAppService) GetNotificationList(ctx context.Context, req *dto.NotificationListRequest) (*dto.NotificationListResponse, error) {
	query := &entity.NotificationQuery{
		UserID:       req.UserID,
		Type:         req.Type,
		Channel:      req.Channel,
		Status:       req.Status,
		BusinessType: req.BusinessType,
		UnreadOnly:   req.UnreadOnly,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		Page:         req.Page,
		PageSize:     req.PageSize,
	}
	
	notifications, total, err := s.notifRepo.List(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeInternal, "查询通知失败")
	}
	
	items := make([]dto.NotificationResponse, len(notifications))
	for i, n := range notifications {
		items[i] = dto.ToNotificationResponse(n)
	}
	
	return &dto.NotificationListResponse{
		List:     items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetUnreadNotifications 获取未读通知
func (s *NotificationAppService) GetUnreadNotifications(ctx context.Context, userID string, limit int) ([]dto.NotificationResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	
	notifications, err := s.notifRepo.GetUnreadList(ctx, userID, limit)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeInternal, "获取未读通知失败")
	}
	
	items := make([]dto.NotificationResponse, len(notifications))
	for i, n := range notifications {
		items[i] = dto.ToNotificationResponse(n)
	}
	
	return items, nil
}

// GetUnreadCount 获取未读数量
func (s *NotificationAppService) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	return s.notifRepo.CountUnread(ctx, userID)
}

// MarkAsRead 标记为已读
func (s *NotificationAppService) MarkAsRead(ctx context.Context, notificationID string) error {
	return s.notifRepo.MarkAsRead(ctx, notificationID)
}

// MarkAllAsRead 标记所有为已读
func (s *NotificationAppService) MarkAllAsRead(ctx context.Context, userID string) error {
	return s.notifRepo.MarkAllAsRead(ctx, userID)
}

// MarkAsArchived 归档通知
func (s *NotificationAppService) MarkAsArchived(ctx context.Context, notificationID string) error {
	return s.notifRepo.MarkAsArchived(ctx, notificationID)
}

// DeleteNotification 删除通知
func (s *NotificationAppService) DeleteNotification(ctx context.Context, notificationID string) error {
	return s.notifRepo.Delete(ctx, notificationID)
}

// GetNotificationStats 获取通知统计
func (s *NotificationAppService) GetNotificationStats(ctx context.Context, userID string) (*entity.NotificationStats, error) {
	return s.notifRepo.GetStats(ctx, userID)
}

// GetUserSettings 获取用户通知设置
func (s *NotificationAppService) GetUserSettings(ctx context.Context, userID string) (*entity.NotificationSetting, error) {
	return s.settingRepo.GetByUserID(ctx, userID)
}

// UpdateUserSettings 更新用户通知设置
func (s *NotificationAppService) UpdateUserSettings(ctx context.Context, userID string, req *dto.UpdateNotificationSettingRequest) error {
	setting := &entity.NotificationSetting{
		UserID:              userID,
		WebSocketEnabled:    req.WebSocketEnabled,
		PushEnabled:         req.PushEnabled,
		SMSEnabled:          req.SMSEnabled,
		EmailEnabled:        req.EmailEnabled,
		InAppEnabled:        req.InAppEnabled,
		SystemEnabled:       req.SystemEnabled,
		TaskEnabled:         req.TaskEnabled,
		WorkflowEnabled:     req.WorkflowEnabled,
		AlertEnabled:        req.AlertEnabled,
		DoNotDisturb:        req.DoNotDisturb,
		DoNotDisturbFrom:    req.DoNotDisturbFrom,
		DoNotDisturbTo:      req.DoNotDisturbTo,
		DailyDigestEnabled:  req.DailyDigestEnabled,
		WeeklyDigestEnabled: req.WeeklyDigestEnabled,
	}
	
	return s.settingRepo.CreateOrUpdate(ctx, setting)
}

// CreateMessageTemplate 创建消息模板
func (s *NotificationAppService) CreateMessageTemplate(ctx context.Context, req *dto.CreateMessageTemplateRequest) (*entity.MessageTemplate, error) {
	template := &entity.MessageTemplate{
		BaseEntity: entity.BaseEntity{
			ID:        uuid.New().String(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Code:          req.Code,
		Name:          req.Name,
		Channel:       req.Channel,
		Type:          req.Type,
		Subject:       req.Subject,
		Content:       req.Content,
		ContentSMS:    req.ContentSMS,
		Variables:     req.Variables,
		VariablesDesc: req.VariablesDesc,
		Status:        "active",
		Version:       1,
		IsSystem:      false,
		ExampleData:   req.ExampleData,
	}
	
	if err := s.templateRepo.Create(ctx, template); err != nil {
		return nil, errors.Wrap(err, errors.CodeInternal, "创建模板失败")
	}
	
	return template, nil
}

// UpdateMessageTemplate 更新消息模板
func (s *NotificationAppService) UpdateMessageTemplate(ctx context.Context, templateID string, req *dto.UpdateMessageTemplateRequest) (*entity.MessageTemplate, error) {
	template, err := s.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeNotFound, "模板不存在")
	}
	
	if req.Name != "" {
		template.Name = req.Name
	}
	if req.Subject != "" {
		template.Subject = req.Subject
	}
	if req.Content != "" {
		template.Content = req.Content
	}
	if req.ContentSMS != "" {
		template.ContentSMS = req.ContentSMS
	}
	if req.Variables != nil {
		template.Variables = req.Variables
	}
	if req.VariablesDesc != nil {
		template.VariablesDesc = req.VariablesDesc
	}
	if req.Status != "" {
		template.Status = req.Status
	}
	if req.ExampleData != nil {
		template.ExampleData = req.ExampleData
	}
	
	template.UpdatedAt = time.Now()
	template.Version++
	
	if err := s.templateRepo.Update(ctx, template); err != nil {
		return nil, errors.Wrap(err, errors.CodeInternal, "更新模板失败")
	}
	
	return template, nil
}

// GetMessageTemplate 获取消息模板
func (s *NotificationAppService) GetMessageTemplate(ctx context.Context, templateID string) (*entity.MessageTemplate, error) {
	return s.templateRepo.FindByID(ctx, templateID)
}

// GetMessageTemplateByCode 根据编码获取模板
func (s *NotificationAppService) GetMessageTemplateByCode(ctx context.Context, code string, channel entity.NotificationChannel) (*entity.MessageTemplate, error) {
	return s.templateRepo.FindByCode(ctx, code, channel)
}

// ListMessageTemplates 查询模板列表
func (s *NotificationAppService) ListMessageTemplates(ctx context.Context, req *dto.MessageTemplateListRequest) (*dto.MessageTemplateListResponse, error) {
	templates, total, err := s.templateRepo.List(ctx, req.Channel, req.Type, req.Page, req.PageSize)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeInternal, "查询模板失败")
	}
	
	items := make([]dto.MessageTemplateResponse, len(templates))
	for i, t := range templates {
		items[i] = dto.ToMessageTemplateResponse(t)
	}
	
	return &dto.MessageTemplateListResponse{
		List:     items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// DeleteMessageTemplate 删除消息模板
func (s *NotificationAppService) DeleteMessageTemplate(ctx context.Context, templateID string) error {
	return s.templateRepo.Delete(ctx, templateID)
}

// RenderTemplate 渲染模板（测试用）
func (s *NotificationAppService) RenderTemplate(ctx context.Context, templateID string, variables map[string]interface{}) (string, error) {
	template, err := s.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return "", errors.Wrap(err, errors.CodeNotFound, "模板不存在")
	}
	
	return template.Render(variables)
}

// GetWebSocketStats 获取WebSocket统计
func (s *NotificationAppService) GetWebSocketStats() websocket.Stats {
	return s.wsManager.GetStats()
}

// GetOnlineUsers 获取在线用户列表
func (s *NotificationAppService) GetOnlineUsers() []string {
	return s.wsManager.GetOnlineUsers()
}

// IsUserOnline 检查用户是否在线
func (s *NotificationAppService) IsUserOnline(userID string) bool {
	return s.wsManager.IsUserOnline(userID)
}

// Broadcast 广播消息
func (s *NotificationAppService) Broadcast(ctx context.Context, title, content string, data map[string]interface{}) error {
	msg := entity.NewWebSocketMessage("system", title, content)
	if data != nil {
		msg.Data = data
	}
	return s.wsManager.Broadcast(msg)
}

// BroadcastToOrg 向组织广播
func (s *NotificationAppService) BroadcastToOrg(ctx context.Context, orgID, title, content string, data map[string]interface{}) error {
	msg := entity.NewWebSocketMessage("system", title, content)
	if data != nil {
		msg.Data = data
	}
	return s.wsManager.BroadcastToOrg(orgID, msg)
}

// NotificationPayload 通知负载（用于序列化）
type NotificationPayload struct {
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Type        string                 `json:"type"`
	Data        map[string]interface{} `json:"data,omitempty"`
	ActionURL   string                 `json:"action_url,omitempty"`
	ActionText  string                 `json:"action_text,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// String 实现Stringer接口
func (p NotificationPayload) String() string {
	data, _ := json.Marshal(p)
	return string(data)
}

// Validate 验证
func (p NotificationPayload) Validate() error {
	if p.Title == "" {
		return fmt.Errorf("标题不能为空")
	}
	if p.Content == "" {
		return fmt.Errorf("内容不能为空")
	}
	return nil
}
