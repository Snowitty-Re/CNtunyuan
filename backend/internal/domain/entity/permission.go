package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PermissionResource 权限资源类型
type PermissionResource string

// 系统资源常量定义
const (
	ResourceUser          PermissionResource = "user"
	ResourceOrganization  PermissionResource = "organization"
	ResourceTask          PermissionResource = "task"
	ResourceMissingPerson PermissionResource = "missing_person"
	ResourceDialect       PermissionResource = "dialect"
	ResourceFile          PermissionResource = "file"
	ResourceWorkflow      PermissionResource = "workflow"
	ResourceAuditLog      PermissionResource = "audit_log"
	ResourceDashboard     PermissionResource = "dashboard"
	ResourceSystem        PermissionResource = "system"
)

// PermissionAction 权限操作类型
type PermissionAction string

// 系统操作常量定义
const (
	ActionCreate PermissionAction = "create"
	ActionRead   PermissionAction = "read"
	ActionUpdate PermissionAction = "update"
	ActionDelete PermissionAction = "delete"
	ActionList   PermissionAction = "list"
	ActionExport PermissionAction = "export"
	ActionImport PermissionAction = "import"
	ActionApprove PermissionAction = "approve"
	ActionAssign  PermissionAction = "assign"
	ActionTransfer PermissionAction = "transfer"
	ActionAll    PermissionAction = "*"
)

// DataScope 数据范围类型
type DataScope string

const (
	// DataScopeAll 全部数据
	DataScopeAll DataScope = "all"
	// DataScopeOrgAndSub 本组织及子组织
	DataScopeOrgAndSub DataScope = "org_and_sub"
	// DataScopeOrgOnly 仅本组织
	DataScopeOrgOnly DataScope = "org_only"
	// DataScopeSelf 仅本人
	DataScopeSelf DataScope = "self"
	// DataScopeCustom 自定义
	DataScopeCustom DataScope = "custom"
)

// IsValid 检查数据范围是否有效
func (s DataScope) IsValid() bool {
	switch s {
	case DataScopeAll, DataScopeOrgAndSub, DataScopeOrgOnly, DataScopeSelf, DataScopeCustom:
		return true
	}
	return false
}

// String 返回字符串表示
func (s DataScope) String() string {
	return string(s)
}

// RoleType 角色类型
type RoleType string

const (
	RoleTypeSystem RoleType = "system" // 系统角色，不可删除
	RoleTypeCustom RoleType = "custom" // 自定义角色
)

// Permission 权限实体
type Permission struct {
	BaseEntity
	Name        string             `gorm:"size:100;not null" json:"name"`
	Code        string             `gorm:"size:100;not null;uniqueIndex" json:"code"`
	Description string             `gorm:"size:255" json:"description,omitempty"`
	Resource    PermissionResource `gorm:"size:50;not null;index" json:"resource"`
	Action      PermissionAction   `gorm:"size:50;not null;index" json:"action"`
	Category    string             `gorm:"size:50" json:"category,omitempty"`
	IsSystem    bool               `gorm:"default:false" json:"is_system"`
}

// TableName 表名
func (Permission) TableName() string {
	return "ty_permissions"
}

// GetCode 生成权限码
func GetCode(resource PermissionResource, action PermissionAction) string {
	return string(resource) + ":" + string(action)
}

// RoleEntity 角色实体
type RoleEntity struct {
	BaseEntity
	Name        string         `gorm:"size:100;not null" json:"name"`
	Code        string         `gorm:"size:50;not null;uniqueIndex" json:"code"`
	Description string         `gorm:"size:255" json:"description,omitempty"`
	Type        RoleType       `gorm:"size:20;not null;default:'custom'" json:"type"`
	OrgID       string         `gorm:"type:uuid;not null;index" json:"org_id"`
	DataScope   DataScope      `gorm:"size:20;default:'org_only'" json:"data_scope"`
	IsSystem    bool           `gorm:"default:false" json:"is_system"`
	Status      string         `gorm:"size:20;default:'active'" json:"status"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	
	Permissions []Permission `gorm:"many2many:ty_role_permissions;" json:"permissions,omitempty"`
}

// TableName 表名
func (RoleEntity) TableName() string {
	return "ty_roles"
}

// IsActive 是否激活
func (r *RoleEntity) IsActive() bool {
	return r.Status == "active"
}

// IsSystemRole 是否是系统角色
func (r *RoleEntity) IsSystemRole() bool {
	return r.Type == RoleTypeSystem || r.IsSystem
}

// HasPermission 检查角色是否有权限
func (r *RoleEntity) HasPermission(resource PermissionResource, action PermissionAction) bool {
	code := GetCode(resource, action)
	for _, p := range r.Permissions {
		if p.Code == code || p.Code == GetCode(resource, ActionAll) {
			return true
		}
	}
	return false
}

// RolePermission 角色权限关联
type RolePermission struct {
	RoleID       string    `gorm:"type:uuid;primaryKey" json:"role_id"`
	PermissionID string    `gorm:"type:uuid;primaryKey" json:"permission_id"`
	GrantedAt    time.Time `gorm:"not null" json:"granted_at"`
	GrantedBy    string    `gorm:"type:uuid" json:"granted_by,omitempty"`
}

// TableName 表名
func (RolePermission) TableName() string {
	return "ty_role_permissions"
}

// UserRole 用户角色关联
type UserRole struct {
	BaseEntity
	UserID    string    `gorm:"type:uuid;not null;index" json:"user_id"`
	RoleID    string    `gorm:"type:uuid;not null;index" json:"role_id"`
	OrgID     string    `gorm:"type:uuid;not null;index" json:"org_id"`
	AssignedBy string   `gorm:"type:uuid" json:"assigned_by,omitempty"`
	AssignedAt time.Time `json:"assigned_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// TableName 表名
func (UserRole) TableName() string {
	return "ty_user_roles"
}

// IsValid 检查是否有效
func (ur *UserRole) IsValid() bool {
	if ur.ExpiresAt == nil {
		return true
	}
	return time.Now().Before(*ur.ExpiresAt)
}

// FieldPermission 字段权限
type FieldPermission struct {
	BaseEntity
	Resource   string `gorm:"size:50;not null;index" json:"resource"`
	Field      string `gorm:"size:50;not null;index" json:"field"`
	RoleID     string `gorm:"type:uuid;not null;index" json:"role_id"`
	Permission string `gorm:"size:20;not null" json:"permission"` // read/write/none
	Condition  string `gorm:"type:text" json:"condition,omitempty"` // 条件表达式
}

// TableName 表名
func (FieldPermission) TableName() string {
	return "ty_field_permissions"
}

// CanRead 是否可以读取
func (fp *FieldPermission) CanRead() bool {
	return fp.Permission == "read" || fp.Permission == "write"
}

// CanWrite 是否可以写入
func (fp *FieldPermission) CanWrite() bool {
	return fp.Permission == "write"
}

// PermissionPolicy 权限策略
type PermissionPolicy struct {
	BaseEntity
	Name        string         `gorm:"size:100;not null" json:"name"`
	Description string         `gorm:"size:255" json:"description,omitempty"`
	OrgID       string         `gorm:"type:uuid;not null;index" json:"org_id"`
	Resource    string         `gorm:"size:50;not null" json:"resource"`
	Rules       PermissionRules `gorm:"type:jsonb" json:"rules"`
	Priority    int            `gorm:"default:0" json:"priority"`
	Status      string         `gorm:"size:20;default:'active'" json:"status"`
}

// TableName 表名
func (PermissionPolicy) TableName() string {
	return "ty_permission_policies"
}

// PermissionRules 权限规则集合
type PermissionRules struct {
	Allow []PermissionRule `json:"allow,omitempty"`
	Deny  []PermissionRule `json:"deny,omitempty"`
}

// Value 实现 driver.Valuer 接口
func (pr PermissionRules) Value() (interface{}, error) {
	return json.Marshal(pr)
}

// Scan 实现 sql.Scanner 接口
func (pr *PermissionRules) Scan(value interface{}) error {
	if value == nil {
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
	
	return json.Unmarshal(bytes, pr)
}

// PermissionRule 权限规则
type PermissionRule struct {
	Subject    string            `json:"subject"`    // user/role
	SubjectID  string            `json:"subject_id"` // user_id/role_id
	Action     string            `json:"action"`     // read/write/delete
	Conditions map[string]interface{} `json:"conditions,omitempty"`
}

// PermissionCheckResult 权限检查结果
type PermissionCheckResult struct {
	Allowed     bool              `json:"allowed"`
	Reason      string            `json:"reason,omitempty"`
	MissingPerm string            `json:"missing_permission,omitempty"`
	DataScope   DataScope         `json:"data_scope,omitempty"`
	FieldMask   []string          `json:"field_mask,omitempty"` // 被屏蔽的字段
}

// PermissionQuery 权限查询
type PermissionQuery struct {
	UserID     string
	OrgID      string
	Resource   PermissionResource
	Action     PermissionAction
	ResourceID string
}

// RoleQuery 角色查询
type RoleQuery struct {
	OrgID    string
	Type     RoleType
	Status   string
	UserID   string
	Page     int
	PageSize int
}

// NewPermission 创建权限
func NewPermission(name string, resource PermissionResource, action PermissionAction) *Permission {
	return &Permission{
		BaseEntity: BaseEntity{
			ID: uuid.New().String(),
		},
		Name:     name,
		Code:     GetCode(resource, action),
		Resource: resource,
		Action:   action,
	}
}

// NewRole 创建角色
func NewRole(name, code, orgID string, roleType RoleType) *RoleEntity {
	return &RoleEntity{
		BaseEntity: BaseEntity{
			ID: uuid.New().String(),
		},
		Name:      name,
		Code:      code,
		OrgID:     orgID,
		Type:      roleType,
		DataScope: DataScopeOrgOnly,
		Status:    "active",
	}
}

// GenerateSystemPermissions 生成系统默认权限
func GenerateSystemPermissions() []*Permission {
	resources := []PermissionResource{
		ResourceUser,
		ResourceOrganization,
		ResourceTask,
		ResourceMissingPerson,
		ResourceDialect,
		ResourceFile,
		ResourceWorkflow,
		ResourceAuditLog,
		ResourceDashboard,
		ResourceSystem,
	}
	
	actions := []PermissionAction{
		ActionCreate,
		ActionRead,
		ActionUpdate,
		ActionDelete,
		ActionList,
		ActionExport,
	}
	
	var permissions []*Permission
	for _, resource := range resources {
		for _, action := range actions {
			perm := NewPermission(
				string(resource)+"_"+string(action),
				resource,
				action,
			)
			perm.IsSystem = true
			perm.Category = string(resource)
			permissions = append(permissions, perm)
		}
	}
	
	return permissions
}

// GenerateSystemRoles 生成系统默认角色
func GenerateSystemRoles(orgID string) []*RoleEntity {
	return []*RoleEntity{
		{
			BaseEntity: BaseEntity{ID: uuid.New().String()},
			Name:       "超级管理员",
			Code:       "super_admin",
			OrgID:      orgID,
			Type:       RoleTypeSystem,
			DataScope:  DataScopeAll,
			IsSystem:   true,
			Status:     "active",
		},
		{
			BaseEntity: BaseEntity{ID: uuid.New().String()},
			Name:       "管理员",
			Code:       "admin",
			OrgID:      orgID,
			Type:       RoleTypeSystem,
			DataScope:  DataScopeOrgAndSub,
			IsSystem:   true,
			Status:     "active",
		},
		{
			BaseEntity: BaseEntity{ID: uuid.New().String()},
			Name:       "管理者",
			Code:       "manager",
			OrgID:      orgID,
			Type:       RoleTypeSystem,
			DataScope:  DataScopeOrgAndSub,
			IsSystem:   true,
			Status:     "active",
		},
		{
			BaseEntity: BaseEntity{ID: uuid.New().String()},
			Name:       "志愿者",
			Code:       "volunteer",
			OrgID:      orgID,
			Type:       RoleTypeSystem,
			DataScope:  DataScopeOrgOnly,
			IsSystem:   true,
			Status:     "active",
		},
	}
}

// SensitiveFields 敏感字段配置
var SensitiveFields = map[string][]string{
	"user":          {"password", "id_card", "phone", "email"},
	"missing_person": {"id_card", "contact_phone"},
}

// IsResourceSensitiveField 检查是否是敏感字段（按资源）
func IsResourceSensitiveField(resource, field string) bool {
	if fields, ok := SensitiveFields[resource]; ok {
		for _, f := range fields {
			if f == field {
				return true
			}
		}
	}
	return false
}
