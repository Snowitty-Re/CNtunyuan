package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Tag 标签模型
type Tag struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"size:50;uniqueIndex;not null;comment:标签名" json:"name"`
	Color       string         `gorm:"size:20;default:#1890ff;comment:颜色" json:"color"`
	Category    string         `gorm:"size:30;comment:分类" json:"category"`
	Description string         `gorm:"size:200;comment:描述" json:"description"`
	UsageCount  int            `gorm:"default:0;comment:使用次数" json:"usage_count"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Notification 通知消息
type Notification struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title      string         `gorm:"size:200;not null;comment:标题" json:"title"`
	Content    string         `gorm:"type:text;comment:内容" json:"content"`
	Type       string         `gorm:"size:30;comment:类型" json:"type"`
	Priority   string         `gorm:"size:20;default:normal;comment:优先级" json:"priority"`
	
	// 关联
	SenderID   uuid.UUID      `gorm:"type:uuid;index;comment:发送人ID" json:"sender_id"`
	ReceiverID uuid.UUID      `gorm:"type:uuid;index;comment:接收人ID" json:"receiver_id"`
	
	// 业务关联
	BusinessID   *uuid.UUID    `gorm:"type:uuid;index;comment:业务ID" json:"business_id"`
	BusinessType string        `gorm:"size:30;comment:业务类型" json:"business_type"`
	
	// 状态
	IsRead     bool           `gorm:"default:false;comment:是否已读" json:"is_read"`
	ReadTime   *time.Time     `gorm:"comment:阅读时间" json:"read_time"`
	IsDeleted  bool           `gorm:"default:false;comment:是否删除" json:"is_deleted"`
	
	// 额外数据
	ExtraData  string         `gorm:"type:jsonb;comment:额外数据" json:"extra_data"`
	
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// OperationLog 操作日志
type OperationLog struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;index;comment:用户ID" json:"user_id"`
	Username     string    `gorm:"size:50;comment:用户名" json:"username"`
	Module       string    `gorm:"size:50;comment:模块" json:"module"`
	Action       string    `gorm:"size:50;comment:操作" json:"action"`
	Method       string    `gorm:"size:10;comment:请求方法" json:"method"`
	Path         string    `gorm:"size:500;comment:请求路径" json:"path"`
	IP           string    `gorm:"size:50;comment:IP地址" json:"ip"`
	UserAgent    string    `gorm:"size:500;comment:用户代理" json:"user_agent"`
	RequestBody  string    `gorm:"type:text;comment:请求参数" json:"request_body"`
	ResponseBody string    `gorm:"type:text;comment:响应内容" json:"response_body"`
	StatusCode   int       `gorm:"comment:状态码" json:"status_code"`
	Duration     int       `gorm:"comment:耗时(ms)" json:"duration"`
	ErrorMsg     string    `gorm:"type:text;comment:错误信息" json:"error_msg"`
	CreatedAt    time.Time `json:"created_at"`
}

// Config 系统配置
type Config struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Key         string         `gorm:"size:100;uniqueIndex;not null;comment:配置键" json:"key"`
	Value       string         `gorm:"type:text;comment:配置值" json:"value"`
	Type        string         `gorm:"size:20;comment:类型" json:"type"`
	Group       string         `gorm:"size:50;comment:分组" json:"group"`
	Description string         `gorm:"size:200;comment:描述" json:"description"`
	IsEditable  bool           `gorm:"default:true;comment:是否可编辑" json:"is_editable"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// DashboardStats 仪表盘统计
type DashboardStats struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Date               time.Time `gorm:"uniqueIndex;not null;comment:日期" json:"date"`
	
	// 志愿者统计
	TotalVolunteers    int       `gorm:"default:0;comment:志愿者总数" json:"total_volunteers"`
	NewVolunteers      int       `gorm:"default:0;comment:新增志愿者" json:"new_volunteers"`
	ActiveVolunteers   int       `gorm:"default:0;comment:活跃志愿者" json:"active_volunteers"`
	
	// 案件统计
	TotalCases         int       `gorm:"default:0;comment:案件总数" json:"total_cases"`
	NewCases           int       `gorm:"default:0;comment:新增案件" json:"new_cases"`
	ResolvedCases      int       `gorm:"default:0;comment:已解决案件" json:"resolved_cases"`
	PendingCases       int       `gorm:"default:0;comment:待处理案件" json:"pending_cases"`
	
	// 任务统计
	TotalTasks         int       `gorm:"default:0;comment:任务总数" json:"total_tasks"`
	NewTasks           int       `gorm:"default:0;comment:新增任务" json:"new_tasks"`
	CompletedTasks     int       `gorm:"default:0;comment:已完成任务" json:"completed_tasks"`
	OverdueTasks       int       `gorm:"default:0;comment:逾期任务" json:"overdue_tasks"`
	
	// 方言统计
	TotalDialects      int       `gorm:"default:0;comment:方言总数" json:"total_dialects"`
	NewDialects        int       `gorm:"default:0;comment:新增方言" json:"new_dialects"`
	DialectPlays       int       `gorm:"default:0;comment:方言播放次数" json:"dialect_plays"`
	
	// 其他统计
	LoginCount         int       `gorm:"default:0;comment:登录次数" json:"login_count"`
	NotificationCount  int       `gorm:"default:0;comment:通知数" json:"notification_count"`
	
	UpdatedAt          time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Tag) TableName() string {
	return "tags"
}

func (Notification) TableName() string {
	return "notifications"
}

func (OperationLog) TableName() string {
	return "operation_logs"
}

func (Config) TableName() string {
	return "configs"
}

func (DashboardStats) TableName() string {
	return "dashboard_stats"
}
