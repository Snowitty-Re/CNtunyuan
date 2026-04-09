// Package entity 审计日志实体
package entity

import (
	"time"

	"github.com/google/uuid"
)

// AuditLogType 审计日志类型
type AuditLogType string

const (
	AuditLogTypeLogin    AuditLogType = "login"    // 登录
	AuditLogTypeLogout   AuditLogType = "logout"   // 登出
	AuditLogTypeCreate   AuditLogType = "create"   // 创建
	AuditLogTypeUpdate   AuditLogType = "update"   // 更新
	AuditLogTypeDelete   AuditLogType = "delete"   // 删除
	AuditLogTypeQuery    AuditLogType = "query"    // 查询
	AuditLogTypeExport   AuditLogType = "export"   // 导出
	AuditLogTypeUpload   AuditLogType = "upload"   // 上传
	AuditLogTypeDownload AuditLogType = "download" // 下载
	AuditLogTypeOther    AuditLogType = "other"    // 其他
)

// AuditLogStatus 审计日志状态
type AuditLogStatus string

const (
	AuditLogStatusSuccess AuditLogStatus = "success" // 成功
	AuditLogStatusFailure AuditLogStatus = "failure" // 失败
)

// AuditLog 审计日志实体
type AuditLog struct {
	BaseEntity
	UserID        string         `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Username      string         `gorm:"size:100" json:"username,omitempty"`
	UserRole      string         `gorm:"size:20" json:"user_role,omitempty"`
	Module        string         `gorm:"size:50;not null" json:"module"`                  // 模块：用户管理、案件管理、任务管理等
	Action        string         `gorm:"size:50;not null" json:"action"`                  // 操作：创建、更新、删除、登录等
	Type          AuditLogType   `gorm:"size:20;not null" json:"type"`                    // 类型
	Description   string         `gorm:"type:text" json:"description,omitempty"`          // 操作描述
	RequestMethod string         `gorm:"size:10" json:"request_method,omitempty"`         // HTTP方法
	RequestURL    string         `gorm:"size:500" json:"request_url,omitempty"`           // 请求URL
	RequestIP     string         `gorm:"size:50" json:"request_ip,omitempty"`             // 请求IP
	RequestBody   string         `gorm:"type:text" json:"request_body,omitempty"`         // 请求体（敏感信息脱敏）
	ResponseCode  int            `json:"response_code,omitempty"`                         // 响应状态码
	ResponseBody  string         `gorm:"type:text" json:"response_body,omitempty"`        // 响应体（简化）
	UserAgent     string         `gorm:"size:500" json:"user_agent,omitempty"`            // 用户代理
	Duration      int64          `gorm:"column:duration_ms" json:"duration_ms,omitempty"` // 执行时长（毫秒）
	Status        AuditLogStatus `gorm:"size:20;not null" json:"status"`                  // 状态
	ErrorMessage  string         `gorm:"type:text" json:"error_message,omitempty"`        // 错误信息
	TraceID       string         `gorm:"size:100;index" json:"trace_id,omitempty"`        // 追踪ID
}

// TableName 表名
func (AuditLog) TableName() string {
	return "ty_audit_logs"
}

// IsSensitiveAction 是否是敏感操作
func (a *AuditLog) IsSensitiveAction() bool {
	switch a.Type {
	case AuditLogTypeLogin, AuditLogTypeLogout, AuditLogTypeDelete:
		return true
	default:
		return false
	}
}

// MarkSuccess 标记为成功
func (a *AuditLog) MarkSuccess() {
	a.Status = AuditLogStatusSuccess
}

// MarkFailure 标记为失败
func (a *AuditLog) MarkFailure(errMsg string) {
	a.Status = AuditLogStatusFailure
	a.ErrorMessage = errMsg
}

// NewAuditLog 创建新的审计日志
func NewAuditLog(module, action string, logType AuditLogType) *AuditLog {
	return &AuditLog{
		BaseEntity: BaseEntity{
			ID:        uuid.New().String(),
			CreatedAt: time.Now(),
		},
		Module: module,
		Action: action,
		Type:   logType,
		Status: AuditLogStatusSuccess,
	}
}

// AuditLogQuery 审计日志查询条件
type AuditLogQuery struct {
	Page         int
	PageSize     int
	UserID       string
	Username     string
	Module       string
	Action       string
	Type         AuditLogType
	Status       AuditLogStatus
	StartTime    *time.Time
	EndTime      *time.Time
	Keyword      string
	RequestIP    string
	RequestPath  string
	ResponseCode *int
}

// NewAuditLogQuery 创建新的查询条件
func NewAuditLogQuery() *AuditLogQuery {
	return &AuditLogQuery{
		Page:     1,
		PageSize: 20,
	}
}

// AuditLogStats 审计日志统计
type AuditLogStats struct {
	TotalCount     int64 `json:"total_count"`
	TodayCount     int64 `json:"today_count"`
	LoginCount     int64 `json:"login_count"`
	OperationCount int64 `json:"operation_count"`
	ErrorCount     int64 `json:"error_count"`
}
