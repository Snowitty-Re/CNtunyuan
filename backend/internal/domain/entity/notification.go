package entity

import (
	"time"
)

// NotificationType 通知类型
type NotificationType string

const (
	NotificationTypeSystem   NotificationType = "system"   // 系统通知
	NotificationTypeMessage  NotificationType = "message"  // 站内消息
	NotificationTypeTask     NotificationType = "task"     // 任务通知
	NotificationTypeWorkflow NotificationType = "workflow" // 工作流通知
	NotificationTypeAlert    NotificationType = "alert"    // 告警通知
)

// NotificationChannel 通知渠道
type NotificationChannel string

const (
	NotificationChannelWebSocket NotificationChannel = "websocket" // WebSocket实时推送
	NotificationChannelPush      NotificationChannel = "push"      // App推送
	NotificationChannelSMS       NotificationChannel = "sms"       // 短信
	NotificationChannelEmail     NotificationChannel = "email"     // 邮件
	NotificationChannelInApp     NotificationChannel = "inapp"     // 站内信
)

// NotificationStatus 通知状态
type NotificationStatus string

const (
	NotificationStatusUnread   NotificationStatus = "unread"   // 未读
	NotificationStatusRead     NotificationStatus = "read"     // 已读
	NotificationStatusArchived NotificationStatus = "archived" // 已归档
	NotificationStatusDeleted  NotificationStatus = "deleted"  // 已删除
)

// Priority 优先级
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityNormal Priority = "normal"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

// Notification 通知实体
type Notification struct {
	BaseEntity
	Title       string             `gorm:"size:200;not null" json:"title"`
	Content     string             `gorm:"type:text" json:"content"`
	Type        NotificationType   `gorm:"size:20;not null;index" json:"type"`
	Channel     NotificationChannel `gorm:"size:20;not null" json:"channel"`
	Priority    Priority           `gorm:"size:10;default:'normal'" json:"priority"`
	Status      NotificationStatus `gorm:"size:20;default:'unread'" json:"status"`
	ToUserID    string             `gorm:"type:uuid;not null;index" json:"to_user_id"`
	FromUserID  *string            `gorm:"type:uuid" json:"from_user_id,omitempty"`
	OrgID       string             `gorm:"type:uuid;not null;index" json:"org_id"`
	
	// 业务关联
	BusinessType string  `gorm:"size:50" json:"business_type,omitempty"` // task/workflow/system
	BusinessID   *string `gorm:"type:uuid" json:"business_id,omitempty"`
	
	// 扩展数据
	Data        map[string]interface{} `gorm:"serializer:json" json:"data,omitempty"`
	ActionURL   string                 `gorm:"size:500" json:"action_url,omitempty"`
	ActionText  string                 `gorm:"size:50" json:"action_text,omitempty"`
	
	// 时间追踪
	ReadAt      *time.Time `json:"read_at,omitempty"`
	ArchivedAt  *time.Time `json:"archived_at,omitempty"`
	ExpireAt    *time.Time `json:"expire_at,omitempty"`
	
	// 发送追踪
	SentAt      *time.Time `json:"sent_at,omitempty"`
	SentSuccess bool       `gorm:"default:false" json:"sent_success"`
	ErrorMsg    string     `gorm:"size:500" json:"error_msg,omitempty"`
	RetryCount  int        `gorm:"default:0" json:"retry_count"`
}

// TableName 表名
func (Notification) TableName() string {
	return "ty_notifications"
}

// IsRead 是否已读
func (n *Notification) IsRead() bool {
	return n.Status == NotificationStatusRead
}

// IsExpired 是否已过期
func (n *Notification) IsExpired() bool {
	if n.ExpireAt == nil {
		return false
	}
	return time.Now().After(*n.ExpireAt)
}

// MarkAsRead 标记为已读
func (n *Notification) MarkAsRead() {
	if n.Status == NotificationStatusUnread {
		n.Status = NotificationStatusRead
		now := time.Now()
		n.ReadAt = &now
	}
}

// MarkAsArchived 标记为已归档
func (n *Notification) MarkAsArchived() {
	n.Status = NotificationStatusArchived
	now := time.Now()
	n.ArchivedAt = &now
}

// CanRetry 是否可以重试
func (n *Notification) CanRetry(maxRetry int) bool {
	return !n.SentSuccess && n.RetryCount < maxRetry
}

// NotificationSetting 用户通知设置
type NotificationSetting struct {
	BaseEntity
	UserID        string `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	OrgID         string `gorm:"type:uuid;not null" json:"org_id"`
	
	// 各渠道开关
	WebSocketEnabled bool `gorm:"default:true" json:"websocket_enabled"`
	PushEnabled      bool `gorm:"default:true" json:"push_enabled"`
	SMSEnabled       bool `gorm:"default:false" json:"sms_enabled"`
	EmailEnabled     bool `gorm:"default:false" json:"email_enabled"`
	InAppEnabled     bool `gorm:"default:true" json:"inapp_enabled"`
	
	// 各类型通知设置
	SystemEnabled   bool `gorm:"default:true" json:"system_enabled"`
	TaskEnabled     bool `gorm:"default:true" json:"task_enabled"`
	WorkflowEnabled bool `gorm:"default:true" json:"workflow_enabled"`
	AlertEnabled    bool `gorm:"default:true" json:"alert_enabled"`
	
	// 免打扰设置
	DoNotDisturb     bool       `gorm:"default:false" json:"do_not_disturb"`
	DoNotDisturbFrom *time.Time `json:"do_not_disturb_from,omitempty"`
	DoNotDisturbTo   *time.Time `json:"do_not_disturb_to,omitempty"`
	
	// 摘要设置
	DailyDigestEnabled  bool `gorm:"default:false" json:"daily_digest_enabled"`
	WeeklyDigestEnabled bool `gorm:"default:false" json:"weekly_digest_enabled"`
}

// TableName 表名
func (NotificationSetting) TableName() string {
	return "ty_notification_settings"
}

// IsChannelEnabled 检查渠道是否启用
func (s *NotificationSetting) IsChannelEnabled(channel NotificationChannel) bool {
	if s.DoNotDisturb && s.isInDoNotDisturbPeriod() {
		return false
	}
	
	switch channel {
	case NotificationChannelWebSocket:
		return s.WebSocketEnabled
	case NotificationChannelPush:
		return s.PushEnabled
	case NotificationChannelSMS:
		return s.SMSEnabled
	case NotificationChannelEmail:
		return s.EmailEnabled
	case NotificationChannelInApp:
		return s.InAppEnabled
	default:
		return true
	}
}

// IsTypeEnabled 检查通知类型是否启用
func (s *NotificationSetting) IsTypeEnabled(notifType NotificationType) bool {
	switch notifType {
	case NotificationTypeSystem:
		return s.SystemEnabled
	case NotificationTypeTask:
		return s.TaskEnabled
	case NotificationTypeWorkflow:
		return s.WorkflowEnabled
	case NotificationTypeAlert:
		return s.AlertEnabled
	default:
		return true
	}
}

// isInDoNotDisturbPeriod 是否在免打扰时间段内
func (s *NotificationSetting) isInDoNotDisturbPeriod() bool {
	if !s.DoNotDisturb || s.DoNotDisturbFrom == nil || s.DoNotDisturbTo == nil {
		return false
	}
	
	now := time.Now()
	from := s.DoNotDisturbFrom
	to := s.DoNotDisturbTo
	
	// 处理跨天的情况
	if to.Before(*from) {
		// 跨天，例如 22:00 - 08:00
		return now.After(*from) || now.Before(*to)
	}
	return now.After(*from) && now.Before(*to)
}

// MessageTemplate 消息模板
type MessageTemplate struct {
	BaseEntity
	Code        string             `gorm:"size:50;uniqueIndex;not null" json:"code"`
	Name        string             `gorm:"size:100;not null" json:"name"`
	Channel     NotificationChannel `gorm:"size:20;not null" json:"channel"`
	Type        NotificationType   `gorm:"size:20;not null" json:"type"`
	
	// 模板内容
	Subject    string `gorm:"size:200" json:"subject,omitempty"` // 邮件/推送标题
	Content    string `gorm:"type:text;not null" json:"content"`  // 模板内容
	ContentSMS string `gorm:"size:500" json:"content_sms,omitempty"` // 短信精简版
	
	// 变量定义
	Variables  map[string]interface{} `gorm:"serializer:json" json:"variables,omitempty"`
	VariablesDesc map[string]string  `gorm:"serializer:json" json:"variables_desc,omitempty"`
	
	// 模板状态
	Status     string `gorm:"size:20;default:'active'" json:"status"`
	Version    int    `gorm:"default:1" json:"version"`
	IsSystem   bool   `gorm:"default:false" json:"is_system"`
	
	// 示例
	ExampleData map[string]interface{} `gorm:"serializer:json" json:"example_data,omitempty"`
}

// TableName 表名
func (MessageTemplate) TableName() string {
	return "ty_message_templates"
}

// IsActive 是否激活
func (t *MessageTemplate) IsActive() bool {
	return t.Status == "active"
}

// Render 渲染模板
func (t *MessageTemplate) Render(variables map[string]interface{}) (string, error) {
	// 简单变量替换实现
	result := t.Content
	for key, value := range variables {
		placeholder := "{{" + key + "}}"
		result = replaceAll(result, placeholder, toString(value))
	}
	return result, nil
}

// RenderSubject 渲染标题
func (t *MessageTemplate) RenderSubject(variables map[string]interface{}) (string, error) {
	result := t.Subject
	for key, value := range variables {
		placeholder := "{{" + key + "}}"
		result = replaceAll(result, placeholder, toString(value))
	}
	return result, nil
}

// replaceAll 替换所有匹配项
func replaceAll(s, old, new string) string {
	// 简单实现，生产环境可以使用 template 包
	for {
		idx := 0
		for i := 0; i <= len(s)-len(old); i++ {
			if s[i:i+len(old)] == old {
				idx = i
				s = s[:idx] + new + s[idx+len(old):]
				break
			}
		}
		if idx == 0 && (len(s) < len(old) || s[:len(old)] != old) {
			break
		}
	}
	return s
}

// toString 转换为字符串
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int, int8, int16, int32, int64:
		return toString(val)
	case uint, uint8, uint16, uint32, uint64:
		return toString(val)
	case float32, float64:
		return toString(val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case time.Time:
		return val.Format("2006-01-02 15:04:05")
	case *time.Time:
		if val != nil {
			return val.Format("2006-01-02 15:04:05")
		}
		return ""
	default:
		return ""
	}
}

// NotificationQuery 通知查询条件
type NotificationQuery struct {
	UserID       string
	OrgID        string
	Type         NotificationType
	Channel      NotificationChannel
	Status       NotificationStatus
	BusinessType string
	UnreadOnly   bool
	StartTime    *time.Time
	EndTime      *time.Time
	Page         int
	PageSize     int
}

// NotificationStats 通知统计
type NotificationStats struct {
	TotalCount   int64 `json:"total_count"`
	UnreadCount  int64 `json:"unread_count"`
	ReadCount    int64 `json:"read_count"`
	SystemCount  int64 `json:"system_count"`
	TaskCount    int64 `json:"task_count"`
	WorkflowCount int64 `json:"workflow_count"`
	AlertCount   int64 `json:"alert_count"`
}

// WebSocketMessage WebSocket消息结构
type WebSocketMessage struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // notification/message/ping/pong
	Title       string                 `json:"title,omitempty"`
	Content     string                 `json:"content,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	FromUserID  string                 `json:"from_user_id,omitempty"`
}

// NewWebSocketMessage 创建WebSocket消息
func NewWebSocketMessage(msgType string, title, content string) *WebSocketMessage {
	return &WebSocketMessage{
		ID:        generateID(),
		Type:      msgType,
		Title:     title,
		Content:   content,
		Timestamp: time.Now(),
	}
}

// generateID 生成唯一ID
func generateID() string {
	return time.Now().Format("20060102150405") + randomString(6)
}

// randomString 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}
