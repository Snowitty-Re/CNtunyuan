package dto

import (
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
)

// CreatePermissionRequest 创建权限请求
type CreatePermissionRequest struct {
	Name        string                   `json:"name" binding:"required"`
	Resource    entity.PermissionResource `json:"resource" binding:"required"`
	Action      entity.PermissionAction   `json:"action" binding:"required"`
	Description string                   `json:"description,omitempty"`
	Category    string                   `json:"category,omitempty"`
}

// UpdatePermissionRequest 更新权限请求
type UpdatePermissionRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
}

// PermissionResponse 权限响应
type PermissionResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Category    string `json:"category,omitempty"`
	IsSystem    bool   `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
}

// ListPermissionsRequest 列表查询请求
type ListPermissionsRequest struct {
	Resource string `json:"resource,omitempty" form:"resource"`
	Category string `json:"category,omitempty" form:"category"`
	Page     int    `json:"page,omitempty" form:"page"`
	PageSize int    `json:"page_size,omitempty" form:"page_size"`
}

// ListPermissionsResponse 列表响应
type ListPermissionsResponse struct {
	List     []PermissionResponse `json:"list"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name          string   `json:"name" binding:"required"`
	Code          string   `json:"code" binding:"required"`
	Description   string   `json:"description,omitempty"`
	OrgID         string   `json:"org_id" binding:"required"`
	DataScope     string   `json:"data_scope,omitempty"`
	PermissionIDs []string `json:"permission_ids,omitempty"`
	SortOrder     int      `json:"sort_order,omitempty"`
	IsSystem      bool     `json:"is_system,omitempty"`
	CreatedBy     string   `json:"created_by,omitempty"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	DataScope   string `json:"data_scope,omitempty"`
	SortOrder   int    `json:"sort_order,omitempty"`
	IsSystem    bool   `json:"is_system,omitempty"`
}

// RoleResponse 角色响应
type RoleResponse struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Code        string               `json:"code"`
	Description string               `json:"description,omitempty"`
	Type        string               `json:"type"`
	OrgID       string               `json:"org_id"`
	DataScope   string               `json:"data_scope"`
	IsSystem    bool                 `json:"is_system"`
	Status      string               `json:"status"`
	SortOrder   int                  `json:"sort_order"`
	Permissions []PermissionResponse `json:"permissions,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
}

// ListRolesRequest 列表查询请求
type ListRolesRequest struct {
	OrgID    string `json:"org_id,omitempty" form:"org_id"`
	Type     string `json:"type,omitempty" form:"type"`
	Status   string `json:"status,omitempty" form:"status"`
	Page     int    `json:"page,omitempty" form:"page"`
	PageSize int    `json:"page_size,omitempty" form:"page_size"`
}

// ListRolesResponse 列表响应
type ListRolesResponse struct {
	List     []RoleResponse `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// GrantPermissionsRequest 授予权限请求
type GrantPermissionsRequest struct {
	PermissionIDs []string `json:"permission_ids" binding:"required"`
	GrantedBy     string   `json:"granted_by,omitempty"`
}

// RevokePermissionsRequest 撤销权限请求
type RevokePermissionsRequest struct {
	PermissionIDs []string `json:"permission_ids" binding:"required"`
}

// RolePermissionsResponse 角色权限响应
type RolePermissionsResponse struct {
	Permissions []PermissionResponse `json:"permissions"`
}

// AssignRoleRequest 分配角色请求
type AssignRoleRequest struct {
	RoleID     string `json:"role_id" binding:"required"`
	OrgID      string `json:"org_id" binding:"required"`
	AssignedBy string `json:"assigned_by,omitempty"`
}

// UserRolesResponse 用户角色响应
type UserRolesResponse struct {
	Roles []RoleResponse `json:"roles"`
}

// SetFieldPermissionRequest 设置字段权限请求
type SetFieldPermissionRequest struct {
	Resource   string `json:"resource" binding:"required"`
	Field      string `json:"field" binding:"required"`
	RoleID     string `json:"role_id" binding:"required"`
	Permission string `json:"permission" binding:"required,oneof=read write none"`
	Condition  string `json:"condition,omitempty"`
}

// FieldPermissionResponse 字段权限响应
type FieldPermissionResponse struct {
	ID         string `json:"id"`
	Resource   string `json:"resource"`
	Field      string `json:"field"`
	RoleID     string `json:"role_id"`
	Permission string `json:"permission"`
	Condition  string `json:"condition,omitempty"`
}

// FieldPermissionsResponse 字段权限列表响应
type FieldPermissionsResponse struct {
	Permissions []FieldPermissionResponse `json:"permissions"`
}

// PermissionCheckResponse 权限检查响应
type PermissionCheckResponse struct {
	Allowed        bool     `json:"allowed"`
	Reason         string   `json:"reason,omitempty"`
	MissingPerm    string   `json:"missing_permission,omitempty"`
	DataScope      string   `json:"data_scope,omitempty"`
}

// ToPermissionResponse 转换为响应
func ToPermissionResponse(perm *entity.Permission) PermissionResponse {
	return PermissionResponse{
		ID:          perm.ID,
		Name:        perm.Name,
		Code:        perm.Code,
		Description: perm.Description,
		Resource:    string(perm.Resource),
		Action:      string(perm.Action),
		Category:    perm.Category,
		IsSystem:    perm.IsSystem,
		CreatedAt:   perm.CreatedAt,
	}
}

// ToRoleResponse 转换为响应
func ToRoleResponse(role *entity.RoleEntity) RoleResponse {
	resp := RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Description,
		Type:        string(role.Type),
		OrgID:       role.OrgID,
		DataScope:   string(role.DataScope),
		IsSystem:    role.IsSystem,
		Status:      role.Status,
		SortOrder:   role.SortOrder,
		CreatedAt:   role.CreatedAt,
	}
	
	if len(role.Permissions) > 0 {
		resp.Permissions = make([]PermissionResponse, len(role.Permissions))
		for i, perm := range role.Permissions {
			resp.Permissions[i] = ToPermissionResponse(&perm)
		}
	}
	
	return resp
}
