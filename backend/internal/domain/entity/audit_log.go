package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AuditAction 审计操作类型
type AuditAction string

const (
	AuditActionCreate  AuditAction = "CREATE"
	AuditActionUpdate  AuditAction = "UPDATE"
	AuditActionDelete  AuditAction = "DELETE"
	AuditActionLogin   AuditAction = "LOGIN"
	AuditActionLogout  AuditAction = "LOGOUT"
	AuditActionQuery   AuditAction = "QUERY"
	AuditActionExport  AuditAction = "EXPORT"
	AuditActionImport  AuditAction = "IMPORT"
	AuditActionApprove AuditAction = "APPROVE"
	AuditActionReject  AuditAction = "REJECT"
	AuditActionOther   AuditAction = "OTHER"
)

// IsValid 检查操作类型是否有效
func (a AuditAction) IsValid() bool {
	switch a {
	case AuditActionCreate, AuditActionUpdate, AuditActionDelete,
		AuditActionLogin, AuditActionLogout, AuditActionQuery,
		AuditActionExport, AuditActionImport, AuditActionApprove,
		AuditActionReject, AuditActionOther:
		return true
	}
	return false
}

// String 返回字符串表示
func (a AuditAction) String() string {
	return string(a)
}

// AuditLog 审计日志实体
type AuditLog struct {
	ID           string          `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt    time.Time       `gorm:"not null;index" json:"created_at"`
	UserID       string          `gorm:"type:uuid;not null;index" json:"user_id"`
	Username     string          `gorm:"size:100" json:"username,omitempty"`
	OrgID        string          `gorm:"type:uuid;not null;index" json:"org_id"`
	Action       AuditAction     `gorm:"size:20;not null;index" json:"action"`
	Resource     string          `gorm:"size:50;not null;index" json:"resource"`
	ResourceID   string          `gorm:"type:uuid;index" json:"resource_id,omitempty"`
	ResourceName string          `gorm:"size:200" json:"resource_name,omitempty"`
	Description  string          `gorm:"type:text" json:"description,omitempty"`
	OldValues    JSONMap         `gorm:"type:jsonb" json:"old_values,omitempty"`
	NewValues    JSONMap         `gorm:"type:jsonb" json:"new_values,omitempty"`
	Delta        JSONMap         `gorm:"type:jsonb" json:"delta,omitempty"` // 变更字段
	IPAddress    string          `gorm:"size:50" json:"ip_address,omitempty"`
	UserAgent    string          `gorm:"type:text" json:"user_agent,omitempty"`
	RequestURL   string          `gorm:"type:text" json:"request_url,omitempty"`
	RequestMethod string         `gorm:"size:10" json:"request_method,omitempty"`
	TraceID      string          `gorm:"size:50;index" json:"trace_id,omitempty"`
	Status       int             `json:"status"` // HTTP status code
	Duration     int64           `json:"duration,omitempty"` // 执行时间(ms)
	Error        string          `gorm:"type:text" json:"error,omitempty"`
	Extra        JSONMap         `gorm:"type:jsonb" json:"extra,omitempty"`
}

// TableName 表名
func (AuditLog) TableName() string {
	return "ty_audit_logs"
}

// JSONMap 用于存储JSON数据的map类型
type JSONMap map[string]interface{}

// Value 实现 driver.Valuer 接口
func (m JSONMap) Value() (interface{}, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan 实现 sql.Scanner 接口
func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return nil
	}
	
	return json.Unmarshal(bytes, m)
}

// AuditLogQuery 审计日志查询条件
type AuditLogQuery struct {
	UserID       string
	OrgID        string
	Action       AuditAction
	Resource     string
	ResourceID   string
	StartTime    *time.Time
	EndTime      *time.Time
	IPAddress    string
	TraceID      string
	Page         int
	PageSize     int
}

// NewAuditLog 创建审计日志
func NewAuditLog(userID, orgID string, action AuditAction, resource string) *AuditLog {
	return &AuditLog{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		UserID:    userID,
		OrgID:     orgID,
		Action:    action,
		Resource:  resource,
		OldValues: make(JSONMap),
		NewValues: make(JSONMap),
		Delta:     make(JSONMap),
		Extra:     make(JSONMap),
	}
}

// SetResourceID 设置资源ID
func (a *AuditLog) SetResourceID(id string) *AuditLog {
	a.ResourceID = id
	return a
}

// SetResourceName 设置资源名称
func (a *AuditLog) SetResourceName(name string) *AuditLog {
	a.ResourceName = name
	return a
}

// SetDescription 设置描述
func (a *AuditLog) SetDescription(desc string) *AuditLog {
	a.Description = desc
	return a
}

// SetOldValues 设置旧值
func (a *AuditLog) SetOldValues(values map[string]interface{}) *AuditLog {
	a.OldValues = values
	return a
}

// SetNewValues 设置新值
func (a *AuditLog) SetNewValues(values map[string]interface{}) *AuditLog {
	a.NewValues = values
	return a
}

// CalculateDelta 计算变更字段
func (a *AuditLog) CalculateDelta() *AuditLog {
	if a.OldValues == nil || a.NewValues == nil {
		return a
	}
	
	delta := make(JSONMap)
	for key, newVal := range a.NewValues {
		if oldVal, exists := a.OldValues[key]; !exists || oldVal != newVal {
			delta[key] = map[string]interface{}{
				"old": oldVal,
				"new": newVal,
			}
		}
	}
	// 检查删除的字段
	for key, oldVal := range a.OldValues {
		if _, exists := a.NewValues[key]; !exists {
			delta[key] = map[string]interface{}{
				"old": oldVal,
				"new": nil,
			}
		}
	}
	a.Delta = delta
	return a
}

// SetRequestInfo 设置请求信息
func (a *AuditLog) SetRequestInfo(method, url, ip, userAgent string) *AuditLog {
	a.RequestMethod = method
	a.RequestURL = url
	a.IPAddress = ip
	a.UserAgent = userAgent
	return a
}

// SetTraceID 设置追踪ID
func (a *AuditLog) SetTraceID(traceID string) *AuditLog {
	a.TraceID = traceID
	return a
}

// SetStatus 设置状态
func (a *AuditLog) SetStatus(status int) *AuditLog {
	a.Status = status
	return a
}

// SetDuration 设置执行时间
func (a *AuditLog) SetDuration(duration int64) *AuditLog {
	a.Duration = duration
	return a
}

// SetError 设置错误信息
func (a *AuditLog) SetError(err string) *AuditLog {
	a.Error = err
	return a
}

// SetUsername 设置用户名
func (a *AuditLog) SetUsername(username string) *AuditLog {
	a.Username = username
	return a
}

// AddExtra 添加额外信息
func (a *AuditLog) AddExtra(key string, value interface{}) *AuditLog {
	if a.Extra == nil {
		a.Extra = make(JSONMap)
	}
	a.Extra[key] = value
	return a
}

// IsSensitiveField 检查是否是敏感字段
func IsSensitiveField(field string) bool {
	sensitiveFields := []string{
		"password", "passwd", "pwd",
		"id_card", "idcard", "identity_card",
		"phone", "mobile", "tel",
		"email",
		"bank_card", "bankcard",
		"credit_card", "creditcard",
	}
	
	lowerField := string([]byte(field))
	for _, sf := range sensitiveFields {
		if lowerField == sf {
			return true
		}
	}
	return false
}

// MaskSensitiveData 脱敏敏感数据
func MaskSensitiveData(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return nil
	}
	
	result := make(map[string]interface{})
	for key, value := range data {
		if IsSensitiveField(key) {
			result[key] = "***"
		} else {
			result[key] = value
		}
	}
	return result
}

// AuditLogStats 审计日志统计
type AuditLogStats struct {
	TotalCount   int64            `json:"total_count"`
	TodayCount   int64            `json:"today_count"`
	ActionStats  map[string]int64 `json:"action_stats"`
	ResourceStats map[string]int64 `json:"resource_stats"`
}

// AuditLogConfig 审计日志配置
type AuditLogConfig struct {
	Enabled         bool     `json:"enabled"`
	LogLevel        string   `json:"log_level"`         // info/warn/error
	RetentionDays   int      `json:"retention_days"`    // 保留天数
	SensitiveFields []string `json:"sensitive_fields"`  // 敏感字段列表
	ExcludePaths    []string `json:"exclude_paths"`     // 排除路径
	MaxBodySize     int      `json:"max_body_size"`     // 最大请求体大小
}

// DefaultAuditLogConfig 默认配置
func DefaultAuditLogConfig() *AuditLogConfig {
	return &AuditLogConfig{
		Enabled:       true,
		LogLevel:      "info",
		RetentionDays: 365,
		SensitiveFields: []string{
			"password", "passwd", "pwd",
			"id_card", "idcard",
			"bank_card", "credit_card",
		},
		ExcludePaths: []string{
			"/api/v1/health",
			"/api/v1/metrics",
			"/uploads/",
		},
		MaxBodySize: 1024 * 1024, // 1MB
	}
}

// ShouldAudit 检查是否应该审计
func (c *AuditLogConfig) ShouldAudit(path string) bool {
	if !c.Enabled {
		return false
	}
	
	for _, exclude := range c.ExcludePaths {
		if path == exclude || (len(exclude) > 0 && exclude[len(exclude)-1] == '/' && len(path) > len(exclude) && path[:len(exclude)] == exclude) {
			return false
		}
	}
	return true
}

// IsSensitiveField 检查字段是否敏感
func (c *AuditLogConfig) IsSensitiveField(field string) bool {
	for _, sf := range c.SensitiveFields {
		if field == sf {
			return true
		}
	}
	return false
}
