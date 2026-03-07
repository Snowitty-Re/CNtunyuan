package dto

import (
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
)

// SendNotificationRequest 发送通知请求
type SendNotificationRequest struct {
	Title        string                 `json:"title" binding:"required"`
	Content      string                 `json:"content" binding:"required"`
	Type         entity.NotificationType `json:"type" binding:"required"`
	Channel      entity.NotificationChannel `json:"channel" binding:"required"`
	Priority     entity.Priority        `json:"priority"`
	ToUserID     string                 `json:"to_user_id" binding:"required"`
	FromUserID   *string                `json:"from_user_id,omitempty"`
	OrgID        string                 `json:"org_id" binding:"required"`
	BusinessType string                 `json:"business_type,omitempty"`
	BusinessID   *string                `json:"business_id,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
	ActionURL    string                 `json:"action_url,omitempty"`
	ActionText   string                 `json:"action_text,omitempty"`
	ExpireAt     *time.Time             `json:"expire_at,omitempty"`
}

// BatchSendNotificationRequest 批量发送通知请求
type BatchSendNotificationRequest struct {
	Title        string                 `json:"title" binding:"required"`
	Content      string                 `json:"content" binding:"required"`
	Type         entity.NotificationType `json:"type" binding:"required"`
	Channel      entity.NotificationChannel `json:"channel" binding:"required"`
	Priority     entity.Priority        `json:"priority"`
	ToUserIDs    []string               `json:"to_user_ids" binding:"required,min=1,max=1000"`
	FromUserID   *string                `json:"from_user_id,omitempty"`
	OrgID        string                 `json:"org_id" binding:"required"`
	BusinessType string                 `json:"business_type,omitempty"`
	BusinessID   *string                `json:"business_id,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
	ActionURL    string                 `json:"action_url,omitempty"`
	ActionText   string                 `json:"action_text,omitempty"`
	ExpireAt     *time.Time             `json:"expire_at,omitempty"`
}

// SendTemplateNotificationRequest 使用模板发送通知请求
type SendTemplateNotificationRequest struct {
	TemplateCode string                 `json:"template_code" binding:"required"`
	Channel      entity.NotificationChannel `json:"channel" binding:"required"`
	Priority     entity.Priority        `json:"priority"`
	ToUserID     string                 `json:"to_user_id" binding:"required"`
	FromUserID   *string                `json:"from_user_id,omitempty"`
	OrgID        string                 `json:"org_id" binding:"required"`
	BusinessType string                 `json:"business_type,omitempty"`
	BusinessID   *string                `json:"business_id,omitempty"`
	Variables    map[string]interface{} `json:"variables,omitempty"`
	ActionURL    string                 `json:"action_url,omitempty"`
	ActionText   string                 `json:"action_text,omitempty"`
	ExpireAt     *time.Time             `json:"expire_at,omitempty"`
}

// NotificationResponse 通知响应
type NotificationResponse struct {
	ID           string                 `json:"id"`
	Title        string                 `json:"title"`
	Content      string                 `json:"content"`
	Type         string                 `json:"type"`
	Channel      string                 `json:"channel"`
	Priority     string                 `json:"priority"`
	Status       string                 `json:"status"`
	ToUserID     string                 `json:"to_user_id"`
	FromUserID   *string                `json:"from_user_id,omitempty"`
	OrgID        string                 `json:"org_id"`
	BusinessType string                 `json:"business_type,omitempty"`
	BusinessID   *string                `json:"business_id,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
	ActionURL    string                 `json:"action_url,omitempty"`
	ActionText   string                 `json:"action_text,omitempty"`
	ReadAt       *time.Time             `json:"read_at,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	ExpireAt     *time.Time             `json:"expire_at,omitempty"`
}

// ToNotificationResponse 转换为通知响应
func ToNotificationResponse(n *entity.Notification) NotificationResponse {
	return NotificationResponse{
		ID:           n.ID,
		Title:        n.Title,
		Content:      n.Content,
		Type:         string(n.Type),
		Channel:      string(n.Channel),
		Priority:     string(n.Priority),
		Status:       string(n.Status),
		ToUserID:     n.ToUserID,
		FromUserID:   n.FromUserID,
		OrgID:        n.OrgID,
		BusinessType: n.BusinessType,
		BusinessID:   n.BusinessID,
		Data:         n.Data,
		ActionURL:    n.ActionURL,
		ActionText:   n.ActionText,
		ReadAt:       n.ReadAt,
		CreatedAt:    n.CreatedAt,
		ExpireAt:     n.ExpireAt,
	}
}

// NotificationListRequest 通知列表请求
type NotificationListRequest struct {
	UserID       string    `json:"user_id"`
	Type         entity.NotificationType `json:"type,omitempty"`
	Channel      entity.NotificationChannel `json:"channel,omitempty"`
	Status       entity.NotificationStatus `json:"status,omitempty"`
	BusinessType string    `json:"business_type,omitempty"`
	UnreadOnly   bool      `json:"unread_only,omitempty"`
	StartTime    *time.Time `json:"start_time,omitempty"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	Page         int       `json:"page"`
	PageSize     int       `json:"page_size"`
}

// NotificationListResponse 通知列表响应
type NotificationListResponse struct {
	List     []NotificationResponse `json:"list"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

// NotificationSettingResponse 通知设置响应
type NotificationSettingResponse struct {
	UserID              string     `json:"user_id"`
	WebSocketEnabled    bool       `json:"websocket_enabled"`
	PushEnabled         bool       `json:"push_enabled"`
	SMSEnabled          bool       `json:"sms_enabled"`
	EmailEnabled        bool       `json:"email_enabled"`
	InAppEnabled        bool       `json:"inapp_enabled"`
	SystemEnabled       bool       `json:"system_enabled"`
	TaskEnabled         bool       `json:"task_enabled"`
	WorkflowEnabled     bool       `json:"workflow_enabled"`
	AlertEnabled        bool       `json:"alert_enabled"`
	DoNotDisturb        bool       `json:"do_not_disturb"`
	DoNotDisturbFrom    *time.Time `json:"do_not_disturb_from,omitempty"`
	DoNotDisturbTo      *time.Time `json:"do_not_disturb_to,omitempty"`
	DailyDigestEnabled  bool       `json:"daily_digest_enabled"`
	WeeklyDigestEnabled bool       `json:"weekly_digest_enabled"`
}

// ToNotificationSettingResponse 转换为通知设置响应
func ToNotificationSettingResponse(s *entity.NotificationSetting) NotificationSettingResponse {
	return NotificationSettingResponse{
		UserID:              s.UserID,
		WebSocketEnabled:    s.WebSocketEnabled,
		PushEnabled:         s.PushEnabled,
		SMSEnabled:          s.SMSEnabled,
		EmailEnabled:        s.EmailEnabled,
		InAppEnabled:        s.InAppEnabled,
		SystemEnabled:       s.SystemEnabled,
		TaskEnabled:         s.TaskEnabled,
		WorkflowEnabled:     s.WorkflowEnabled,
		AlertEnabled:        s.AlertEnabled,
		DoNotDisturb:        s.DoNotDisturb,
		DoNotDisturbFrom:    s.DoNotDisturbFrom,
		DoNotDisturbTo:      s.DoNotDisturbTo,
		DailyDigestEnabled:  s.DailyDigestEnabled,
		WeeklyDigestEnabled: s.WeeklyDigestEnabled,
	}
}

// UpdateNotificationSettingRequest 更新通知设置请求
type UpdateNotificationSettingRequest struct {
	WebSocketEnabled    bool       `json:"websocket_enabled"`
	PushEnabled         bool       `json:"push_enabled"`
	SMSEnabled          bool       `json:"sms_enabled"`
	EmailEnabled        bool       `json:"email_enabled"`
	InAppEnabled        bool       `json:"inapp_enabled"`
	SystemEnabled       bool       `json:"system_enabled"`
	TaskEnabled         bool       `json:"task_enabled"`
	WorkflowEnabled     bool       `json:"workflow_enabled"`
	AlertEnabled        bool       `json:"alert_enabled"`
	DoNotDisturb        bool       `json:"do_not_disturb"`
	DoNotDisturbFrom    *time.Time `json:"do_not_disturb_from,omitempty"`
	DoNotDisturbTo      *time.Time `json:"do_not_disturb_to,omitempty"`
	DailyDigestEnabled  bool       `json:"daily_digest_enabled"`
	WeeklyDigestEnabled bool       `json:"weekly_digest_enabled"`
}

// CreateMessageTemplateRequest 创建消息模板请求
type CreateMessageTemplateRequest struct {
	Code          string                 `json:"code" binding:"required,max=50"`
	Name          string                 `json:"name" binding:"required,max=100"`
	Channel       entity.NotificationChannel `json:"channel" binding:"required"`
	Type          entity.NotificationType `json:"type" binding:"required"`
	Subject       string                 `json:"subject"`
	Content       string                 `json:"content" binding:"required"`
	ContentSMS    string                 `json:"content_sms,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	VariablesDesc map[string]string      `json:"variables_desc,omitempty"`
	ExampleData   map[string]interface{} `json:"example_data,omitempty"`
}

// UpdateMessageTemplateRequest 更新消息模板请求
type UpdateMessageTemplateRequest struct {
	Name          string                 `json:"name,omitempty"`
	Subject       string                 `json:"subject,omitempty"`
	Content       string                 `json:"content,omitempty"`
	ContentSMS    string                 `json:"content_sms,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	VariablesDesc map[string]string      `json:"variables_desc,omitempty"`
	Status        string                 `json:"status,omitempty"`
	ExampleData   map[string]interface{} `json:"example_data,omitempty"`
}

// MessageTemplateResponse 消息模板响应
type MessageTemplateResponse struct {
	ID            string                 `json:"id"`
	Code          string                 `json:"code"`
	Name          string                 `json:"name"`
	Channel       string                 `json:"channel"`
	Type          string                 `json:"type"`
	Subject       string                 `json:"subject,omitempty"`
	Content       string                 `json:"content"`
	ContentSMS    string                 `json:"content_sms,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	VariablesDesc map[string]string      `json:"variables_desc,omitempty"`
	Status        string                 `json:"status"`
	Version       int                    `json:"version"`
	IsSystem      bool                   `json:"is_system"`
	ExampleData   map[string]interface{} `json:"example_data,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// ToMessageTemplateResponse 转换为消息模板响应
func ToMessageTemplateResponse(t *entity.MessageTemplate) MessageTemplateResponse {
	return MessageTemplateResponse{
		ID:            t.ID,
		Code:          t.Code,
		Name:          t.Name,
		Channel:       string(t.Channel),
		Type:          string(t.Type),
		Subject:       t.Subject,
		Content:       t.Content,
		ContentSMS:    t.ContentSMS,
		Variables:     t.Variables,
		VariablesDesc: t.VariablesDesc,
		Status:        t.Status,
		Version:       t.Version,
		IsSystem:      t.IsSystem,
		ExampleData:   t.ExampleData,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
	}
}

// MessageTemplateListRequest 消息模板列表请求
type MessageTemplateListRequest struct {
	Channel  entity.NotificationChannel `json:"channel,omitempty"`
	Type     entity.NotificationType `json:"type,omitempty"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"page_size"`
}

// MessageTemplateListResponse 消息模板列表响应
type MessageTemplateListResponse struct {
	List     []MessageTemplateResponse `json:"list"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	PageSize int                       `json:"page_size"`
}

// WebSocketStatsResponse WebSocket统计响应
type WebSocketStatsResponse struct {
	TotalConnections  int64    `json:"total_connections"`
	ActiveConnections int64    `json:"active_connections"`
	MessagesSent      int64    `json:"messages_sent"`
	MessagesReceived  int64    `json:"messages_received"`
	BroadcastsSent    int64    `json:"broadcasts_sent"`
	OnlineUsers       []string `json:"online_users,omitempty"`
	OnlineCount       int      `json:"online_count"`
}

// RenderTemplateRequest 渲染模板请求
type RenderTemplateRequest struct {
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// RenderTemplateResponse 渲染模板响应
type RenderTemplateResponse struct {
	Result string `json:"result"`
}

// UnreadCountResponse 未读数量响应
type UnreadCountResponse struct {
	Count int64 `json:"count"`
}

// NotificationStatsResponse 通知统计响应
type NotificationStatsResponse struct {
	TotalCount    int64 `json:"total_count"`
	UnreadCount   int64 `json:"unread_count"`
	ReadCount     int64 `json:"read_count"`
	SystemCount   int64 `json:"system_count"`
	TaskCount     int64 `json:"task_count"`
	WorkflowCount int64 `json:"workflow_count"`
	AlertCount    int64 `json:"alert_count"`
}

// BroadcastRequest 广播请求
type BroadcastRequest struct {
	Title   string                 `json:"title" binding:"required"`
	Content string                 `json:"content" binding:"required"`
	OrgID   string                 `json:"org_id,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// MarkReadRequest 标记已读请求
type MarkReadRequest struct {
	IDs []string `json:"ids,omitempty"`
}

// MarkReadResponse 标记已读响应
type MarkReadResponse struct {
	MarkedCount int `json:"marked_count"`
}
