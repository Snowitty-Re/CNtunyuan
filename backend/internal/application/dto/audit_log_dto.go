package dto

import (
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
)

// AuditLogListRequest 审计日志列表请求
type AuditLogListRequest struct {
	UserID     string `json:"user_id,omitempty" form:"user_id"`
	OrgID      string `json:"org_id,omitempty" form:"org_id"`
	Action     string `json:"action,omitempty" form:"action"`
	Resource   string `json:"resource,omitempty" form:"resource"`
	ResourceID string `json:"resource_id,omitempty" form:"resource_id"`
	IPAddress  string `json:"ip_address,omitempty" form:"ip_address"`
	TraceID    string `json:"trace_id,omitempty" form:"trace_id"`
	StartTime  string `json:"start_time,omitempty" form:"start_time"`
	EndTime    string `json:"end_time,omitempty" form:"end_time"`
	Page       int    `json:"page,omitempty" form:"page"`
	PageSize   int    `json:"page_size,omitempty" form:"page_size"`
}

// AuditLogResponse 审计日志响应
type AuditLogResponse struct {
	ID            string                 `json:"id"`
	CreatedAt     time.Time              `json:"created_at"`
	UserID        string                 `json:"user_id"`
	Username      string                 `json:"username,omitempty"`
	OrgID         string                 `json:"org_id"`
	Action        string                 `json:"action"`
	Resource      string                 `json:"resource"`
	ResourceID    string                 `json:"resource_id,omitempty"`
	ResourceName  string                 `json:"resource_name,omitempty"`
	Description   string                 `json:"description,omitempty"`
	OldValues     map[string]interface{} `json:"old_values,omitempty"`
	NewValues     map[string]interface{} `json:"new_values,omitempty"`
	Delta         map[string]interface{} `json:"delta,omitempty"`
	IPAddress     string                 `json:"ip_address,omitempty"`
	UserAgent     string                 `json:"user_agent,omitempty"`
	RequestURL    string                 `json:"request_url,omitempty"`
	RequestMethod string                 `json:"request_method,omitempty"`
	TraceID       string                 `json:"trace_id,omitempty"`
	Status        int                    `json:"status"`
	Duration      int64                  `json:"duration,omitempty"`
	Error         string                 `json:"error,omitempty"`
	Extra         map[string]interface{} `json:"extra,omitempty"`
}

// ToAuditLogResponse 转换为审计日志响应
func ToAuditLogResponse(log *entity.AuditLog) AuditLogResponse {
	return AuditLogResponse{
		ID:            log.ID,
		CreatedAt:     log.CreatedAt,
		UserID:        log.UserID,
		Username:      log.Username,
		OrgID:         log.OrgID,
		Action:        string(log.Action),
		Resource:      log.Resource,
		ResourceID:    log.ResourceID,
		ResourceName:  log.ResourceName,
		Description:   log.Description,
		OldValues:     log.OldValues,
		NewValues:     log.NewValues,
		Delta:         log.Delta,
		IPAddress:     log.IPAddress,
		UserAgent:     log.UserAgent,
		RequestURL:    log.RequestURL,
		RequestMethod: log.RequestMethod,
		TraceID:       log.TraceID,
		Status:        log.Status,
		Duration:      log.Duration,
		Error:         log.Error,
		Extra:         log.Extra,
	}
}

// AuditLogListResponse 审计日志列表响应
type AuditLogListResponse struct {
	List     []AuditLogResponse `json:"list"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

// AuditLogStatsResponse 审计日志统计响应
type AuditLogStatsResponse struct {
	TotalCount    int64            `json:"total_count"`
	TodayCount    int64            `json:"today_count"`
	ActionStats   map[string]int64 `json:"action_stats"`
	ResourceStats map[string]int64 `json:"resource_stats"`
	PeriodDays    int              `json:"period_days"`
}

// AuditLogUserActivityResponse 用户活动响应
type AuditLogUserActivityResponse struct {
	TotalCount    int64                  `json:"total_count"`
	ActionStats   map[string]int64       `json:"action_stats"`
	ResourceStats map[string]int64       `json:"resource_stats"`
	DailyStats    map[string]int64       `json:"daily_stats"`
	PeriodDays    int                    `json:"period_days"`
}

// AuditLogRetentionStatsResponse 保留统计响应
type AuditLogRetentionStatsResponse struct {
	TotalCount     int64              `json:"total_count"`
	RetentionDays  int                `json:"retention_days"`
	LogsToCleanup  int64              `json:"logs_to_cleanup"`
	OldestLogTime  *time.Time         `json:"oldest_log_time,omitempty"`
	MonthlyStats   map[string]int64   `json:"monthly_stats"`
}
