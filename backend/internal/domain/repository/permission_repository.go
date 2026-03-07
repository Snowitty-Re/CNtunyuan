package repository

import (
	"context"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
)

// PermissionRepository 权限仓储接口
type PermissionRepository interface {
	// Permission operations
	CreatePermission(ctx context.Context, perm *entity.Permission) error
	UpdatePermission(ctx context.Context, perm *entity.Permission) error
	DeletePermission(ctx context.Context, id string) error
	FindPermissionByID(ctx context.Context, id string) (*entity.Permission, error)
	FindPermissionByCode(ctx context.Context, code string) (*entity.Permission, error)
	ListPermissions(ctx context.Context, resource, category string, page, pageSize int) ([]*entity.Permission, int64, error)
	ListAllPermissions(ctx context.Context) ([]*entity.Permission, error)
	
	// Role operations
	CreateRole(ctx context.Context, role *entity.RoleEntity) error
	UpdateRole(ctx context.Context, role *entity.RoleEntity) error
	DeleteRole(ctx context.Context, id string) error
	FindRoleByID(ctx context.Context, id string) (*entity.RoleEntity, error)
	FindRoleByCode(ctx context.Context, code string, orgID string) (*entity.RoleEntity, error)
	ListRoles(ctx context.Context, query *entity.RoleQuery) ([]*entity.RoleEntity, int64, error)
	ListRolesByUserID(ctx context.Context, userID string) ([]*entity.RoleEntity, error)
	
	// Role permission operations
	GrantPermissionToRole(ctx context.Context, roleID, permissionID string, grantedBy string) error
	RevokePermissionFromRole(ctx context.Context, roleID, permissionID string) error
	RevokeAllPermissionsFromRole(ctx context.Context, roleID string) error
	ListPermissionsByRoleID(ctx context.Context, roleID string) ([]*entity.Permission, error)
	HasPermission(ctx context.Context, roleID, permissionCode string) (bool, error)
	
	// User role operations
	AssignRoleToUser(ctx context.Context, userID, roleID, orgID, assignedBy string) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID string) error
	RemoveAllRolesFromUser(ctx context.Context, userID string) error
	ListUserRoles(ctx context.Context, userID string) ([]*entity.UserRole, error)
	ListUsersByRoleID(ctx context.Context, roleID string) ([]string, error)
	
	// Field permission operations
	CreateFieldPermission(ctx context.Context, fp *entity.FieldPermission) error
	UpdateFieldPermission(ctx context.Context, fp *entity.FieldPermission) error
	DeleteFieldPermission(ctx context.Context, id string) error
	FindFieldPermission(ctx context.Context, resource, field, roleID string) (*entity.FieldPermission, error)
	ListFieldPermissions(ctx context.Context, resource, roleID string) ([]*entity.FieldPermission, error)
	GetFieldPermissionsForRole(ctx context.Context, roleID string) (map[string]map[string]string, error)
	
	// Permission policy operations
	CreatePolicy(ctx context.Context, policy *entity.PermissionPolicy) error
	UpdatePolicy(ctx context.Context, policy *entity.PermissionPolicy) error
	DeletePolicy(ctx context.Context, id string) error
	FindPolicyByID(ctx context.Context, id string) (*entity.PermissionPolicy, error)
	ListPolicies(ctx context.Context, orgID, resource string) ([]*entity.PermissionPolicy, error)
	
	// Batch operations
	BatchGrantPermissions(ctx context.Context, roleID string, permissionIDs []string, grantedBy string) error
	BatchRevokePermissions(ctx context.Context, roleID string, permissionIDs []string) error
	
	// Check operations
	CheckUserPermission(ctx context.Context, userID string, resource entity.PermissionResource, action entity.PermissionAction) (bool, error)
	GetUserDataScope(ctx context.Context, userID string) (entity.DataScope, error)
	GetUserPermissions(ctx context.Context, userID string) ([]*entity.Permission, error)
}
