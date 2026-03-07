package repository

import (
	"context"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PermissionRepositoryImpl 权限仓储实现
type PermissionRepositoryImpl struct {
	db *gorm.DB
}

// NewPermissionRepository 创建权限仓储
func NewPermissionRepository(db *gorm.DB) repository.PermissionRepository {
	return &PermissionRepositoryImpl{db: db}
}

// CreatePermission 创建权限
func (r *PermissionRepositoryImpl) CreatePermission(ctx context.Context, perm *entity.Permission) error {
	return r.db.WithContext(ctx).Create(perm).Error
}

// UpdatePermission 更新权限
func (r *PermissionRepositoryImpl) UpdatePermission(ctx context.Context, perm *entity.Permission) error {
	return r.db.WithContext(ctx).Save(perm).Error
}

// DeletePermission 删除权限
func (r *PermissionRepositoryImpl) DeletePermission(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Permission{}, "id = ?", id).Error
}

// FindPermissionByID 根据ID查找权限
func (r *PermissionRepositoryImpl) FindPermissionByID(ctx context.Context, id string) (*entity.Permission, error) {
	var perm entity.Permission
	err := r.db.WithContext(ctx).First(&perm, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &perm, err
}

// FindPermissionByCode 根据Code查找权限
func (r *PermissionRepositoryImpl) FindPermissionByCode(ctx context.Context, code string) (*entity.Permission, error) {
	var perm entity.Permission
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&perm).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &perm, err
}

// ListPermissions 列表查询权限
func (r *PermissionRepositoryImpl) ListPermissions(ctx context.Context, resource, category string, page, pageSize int) ([]*entity.Permission, int64, error) {
	db := r.db.WithContext(ctx).Model(&entity.Permission{})
	
	if resource != "" {
		db = db.Where("resource = ?", resource)
	}
	if category != "" {
		db = db.Where("category = ?", category)
	}
	
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	
	var perms []*entity.Permission
	err := db.Order("category ASC, action ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&perms).Error
	
	return perms, total, err
}

// ListAllPermissions 查询所有权限
func (r *PermissionRepositoryImpl) ListAllPermissions(ctx context.Context) ([]*entity.Permission, error) {
	var perms []*entity.Permission
	err := r.db.WithContext(ctx).
		Order("category ASC, action ASC").
		Find(&perms).Error
	return perms, err
}

// CreateRole 创建角色
func (r *PermissionRepositoryImpl) CreateRole(ctx context.Context, role *entity.RoleEntity) error {
	return r.db.WithContext(ctx).Create(role).Error
}

// UpdateRole 更新角色
func (r *PermissionRepositoryImpl) UpdateRole(ctx context.Context, role *entity.RoleEntity) error {
	return r.db.WithContext(ctx).Save(role).Error
}

// DeleteRole 删除角色
func (r *PermissionRepositoryImpl) DeleteRole(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除角色权限关联
		if err := tx.Where("role_id = ?", id).Delete(&entity.RolePermission{}).Error; err != nil {
			return err
		}
		// 删除用户角色关联
		if err := tx.Where("role_id = ?", id).Delete(&entity.UserRole{}).Error; err != nil {
			return err
		}
		// 删除角色
		return tx.Delete(&entity.RoleEntity{}, "id = ?", id).Error
	})
}

// FindRoleByID 根据ID查找角色
func (r *PermissionRepositoryImpl) FindRoleByID(ctx context.Context, id string) (*entity.RoleEntity, error) {
	var role entity.RoleEntity
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		First(&role, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &role, err
}

// FindRoleByCode 根据Code查找角色
func (r *PermissionRepositoryImpl) FindRoleByCode(ctx context.Context, code string, orgID string) (*entity.RoleEntity, error) {
	var role entity.RoleEntity
	db := r.db.WithContext(ctx).Preload("Permissions")
	
	if orgID != "" {
		db = db.Where("code = ? AND org_id = ?", code, orgID)
	} else {
		db = db.Where("code = ?", code)
	}
	
	err := db.First(&role).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &role, err
}

// ListRoles 列表查询角色
func (r *PermissionRepositoryImpl) ListRoles(ctx context.Context, query *entity.RoleQuery) ([]*entity.RoleEntity, int64, error) {
	db := r.db.WithContext(ctx).Model(&entity.RoleEntity{})
	
	if query.OrgID != "" {
		db = db.Where("org_id = ?", query.OrgID)
	}
	if query.Type != "" {
		db = db.Where("type = ?", query.Type)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	
	var roles []*entity.RoleEntity
	err := db.Order("sort_order ASC, created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&roles).Error
	
	return roles, total, err
}

// ListRolesByUserID 查询用户角色
func (r *PermissionRepositoryImpl) ListRolesByUserID(ctx context.Context, userID string) ([]*entity.RoleEntity, error) {
	var roles []*entity.RoleEntity
	err := r.db.WithContext(ctx).
		Joins("JOIN ty_user_roles ON ty_user_roles.role_id = ty_roles.id").
		Where("ty_user_roles.user_id = ?", userID).
		Where("ty_roles.status = ?", "active").
		Find(&roles).Error
	return roles, err
}

// GrantPermissionToRole 授予权限
func (r *PermissionRepositoryImpl) GrantPermissionToRole(ctx context.Context, roleID, permissionID string, grantedBy string) error {
	rp := &entity.RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
		GrantedAt:    time.Now(),
		GrantedBy:    grantedBy,
	}
	return r.db.WithContext(ctx).Create(rp).Error
}

// RevokePermissionFromRole 撤销权限
func (r *PermissionRepositoryImpl) RevokePermissionFromRole(ctx context.Context, roleID, permissionID string) error {
	return r.db.WithContext(ctx).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&entity.RolePermission{}).Error
}

// RevokeAllPermissionsFromRole 撤销所有权限
func (r *PermissionRepositoryImpl) RevokeAllPermissionsFromRole(ctx context.Context, roleID string) error {
	return r.db.WithContext(ctx).
		Where("role_id = ?", roleID).
		Delete(&entity.RolePermission{}).Error
}

// ListPermissionsByRoleID 查询角色权限
func (r *PermissionRepositoryImpl) ListPermissionsByRoleID(ctx context.Context, roleID string) ([]*entity.Permission, error) {
	var perms []*entity.Permission
	err := r.db.WithContext(ctx).
		Joins("JOIN ty_role_permissions ON ty_role_permissions.permission_id = ty_permissions.id").
		Where("ty_role_permissions.role_id = ?", roleID).
		Find(&perms).Error
	return perms, err
}

// HasPermission 检查角色是否有权限
func (r *PermissionRepositoryImpl) HasPermission(ctx context.Context, roleID, permissionCode string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.RolePermission{}).
		Joins("JOIN ty_permissions ON ty_permissions.id = ty_role_permissions.permission_id").
		Where("ty_role_permissions.role_id = ?", roleID).
		Where("ty_permissions.code = ?", permissionCode).
		Count(&count).Error
	return count > 0, err
}

// AssignRoleToUser 分配角色给用户
func (r *PermissionRepositoryImpl) AssignRoleToUser(ctx context.Context, userID, roleID, orgID, assignedBy string) error {
	// 检查是否已存在
	var count int64
	r.db.WithContext(ctx).Model(&entity.UserRole{}).
		Where("user_id = ? AND role_id = ? AND org_id = ?", userID, roleID, orgID).
		Count(&count)
	
	if count > 0 {
		// 已存在，更新
		return r.db.WithContext(ctx).
			Model(&entity.UserRole{}).
			Where("user_id = ? AND role_id = ? AND org_id = ?", userID, roleID, orgID).
			Updates(map[string]interface{}{
				"assigned_by": assignedBy,
				"assigned_at": time.Now(),
			}).Error
	}
	
	ur := &entity.UserRole{
		BaseEntity: entity.BaseEntity{
			ID: uuid.New().String(),
		},
		UserID:     userID,
		RoleID:     roleID,
		OrgID:      orgID,
		AssignedBy: assignedBy,
		AssignedAt: time.Now(),
	}
	return r.db.WithContext(ctx).Create(ur).Error
}

// RemoveRoleFromUser 移除用户角色
func (r *PermissionRepositoryImpl) RemoveRoleFromUser(ctx context.Context, userID, roleID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&entity.UserRole{}).Error
}

// RemoveAllRolesFromUser 移除用户所有角色
func (r *PermissionRepositoryImpl) RemoveAllRolesFromUser(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&entity.UserRole{}).Error
}

// ListUserRoles 查询用户角色关联
func (r *PermissionRepositoryImpl) ListUserRoles(ctx context.Context, userID string) ([]*entity.UserRole, error) {
	var urs []*entity.UserRole
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&urs).Error
	return urs, err
}

// ListUsersByRoleID 查询角色下的用户
func (r *PermissionRepositoryImpl) ListUsersByRoleID(ctx context.Context, roleID string) ([]string, error) {
	var userIDs []string
	err := r.db.WithContext(ctx).
		Model(&entity.UserRole{}).
		Where("role_id = ?", roleID).
		Pluck("user_id", &userIDs).Error
	return userIDs, err
}

// CreateFieldPermission 创建字段权限
func (r *PermissionRepositoryImpl) CreateFieldPermission(ctx context.Context, fp *entity.FieldPermission) error {
	return r.db.WithContext(ctx).Create(fp).Error
}

// UpdateFieldPermission 更新字段权限
func (r *PermissionRepositoryImpl) UpdateFieldPermission(ctx context.Context, fp *entity.FieldPermission) error {
	return r.db.WithContext(ctx).Save(fp).Error
}

// DeleteFieldPermission 删除字段权限
func (r *PermissionRepositoryImpl) DeleteFieldPermission(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.FieldPermission{}, "id = ?", id).Error
}

// FindFieldPermission 查找字段权限
func (r *PermissionRepositoryImpl) FindFieldPermission(ctx context.Context, resource, field, roleID string) (*entity.FieldPermission, error) {
	var fp entity.FieldPermission
	err := r.db.WithContext(ctx).
		Where("resource = ? AND field = ? AND role_id = ?", resource, field, roleID).
		First(&fp).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &fp, err
}

// ListFieldPermissions 列表查询字段权限
func (r *PermissionRepositoryImpl) ListFieldPermissions(ctx context.Context, resource, roleID string) ([]*entity.FieldPermission, error) {
	db := r.db.WithContext(ctx).Model(&entity.FieldPermission{})
	
	if resource != "" {
		db = db.Where("resource = ?", resource)
	}
	if roleID != "" {
		db = db.Where("role_id = ?", roleID)
	}
	
	var fps []*entity.FieldPermission
	err := db.Find(&fps).Error
	return fps, err
}

// GetFieldPermissionsForRole 获取角色的字段权限
func (r *PermissionRepositoryImpl) GetFieldPermissionsForRole(ctx context.Context, roleID string) (map[string]map[string]string, error) {
	var fps []*entity.FieldPermission
	err := r.db.WithContext(ctx).
		Where("role_id = ?", roleID).
		Find(&fps).Error
	if err != nil {
		return nil, err
	}
	
	result := make(map[string]map[string]string)
	for _, fp := range fps {
		if result[fp.Resource] == nil {
			result[fp.Resource] = make(map[string]string)
		}
		result[fp.Resource][fp.Field] = fp.Permission
	}
	
	return result, nil
}

// CreatePolicy 创建策略
func (r *PermissionRepositoryImpl) CreatePolicy(ctx context.Context, policy *entity.PermissionPolicy) error {
	return r.db.WithContext(ctx).Create(policy).Error
}

// UpdatePolicy 更新策略
func (r *PermissionRepositoryImpl) UpdatePolicy(ctx context.Context, policy *entity.PermissionPolicy) error {
	return r.db.WithContext(ctx).Save(policy).Error
}

// DeletePolicy 删除策略
func (r *PermissionRepositoryImpl) DeletePolicy(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.PermissionPolicy{}, "id = ?", id).Error
}

// FindPolicyByID 根据ID查找策略
func (r *PermissionRepositoryImpl) FindPolicyByID(ctx context.Context, id string) (*entity.PermissionPolicy, error) {
	var policy entity.PermissionPolicy
	err := r.db.WithContext(ctx).First(&policy, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &policy, err
}

// ListPolicies 列表查询策略
func (r *PermissionRepositoryImpl) ListPolicies(ctx context.Context, orgID, resource string) ([]*entity.PermissionPolicy, error) {
	db := r.db.WithContext(ctx).Model(&entity.PermissionPolicy{})
	
	if orgID != "" {
		db = db.Where("org_id = ?", orgID)
	}
	if resource != "" {
		db = db.Where("resource = ?", resource)
	}
	
	var policies []*entity.PermissionPolicy
	err := db.Order("priority DESC").Find(&policies).Error
	return policies, err
}

// BatchGrantPermissions 批量授予权限
func (r *PermissionRepositoryImpl) BatchGrantPermissions(ctx context.Context, roleID string, permissionIDs []string, grantedBy string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, permID := range permissionIDs {
			rp := &entity.RolePermission{
				RoleID:       roleID,
				PermissionID: permID,
				GrantedAt:    time.Now(),
				GrantedBy:    grantedBy,
			}
			if err := tx.Create(rp).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchRevokePermissions 批量撤销权限
func (r *PermissionRepositoryImpl) BatchRevokePermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	return r.db.WithContext(ctx).
		Where("role_id = ? AND permission_id IN ?", roleID, permissionIDs).
		Delete(&entity.RolePermission{}).Error
}

// CheckUserPermission 检查用户权限
func (r *PermissionRepositoryImpl) CheckUserPermission(ctx context.Context, userID string, resource entity.PermissionResource, action entity.PermissionAction) (bool, error) {
	code := entity.GetCode(resource, action)
	
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.RolePermission{}).
		Joins("JOIN ty_user_roles ON ty_user_roles.role_id = ty_role_permissions.role_id").
		Joins("JOIN ty_permissions ON ty_permissions.id = ty_role_permissions.permission_id").
		Where("ty_user_roles.user_id = ?", userID).
		Where("(ty_permissions.code = ? OR ty_permissions.code = ?)", code, entity.GetCode(resource, entity.ActionAll)).
		Count(&count).Error
	
	return count > 0, err
}

// GetUserDataScope 获取用户数据范围
func (r *PermissionRepositoryImpl) GetUserDataScope(ctx context.Context, userID string) (entity.DataScope, error) {
	var roleIDs []string
	err := r.db.WithContext(ctx).Model(&entity.UserRole{}).
		Where("user_id = ?", userID).
		Pluck("role_id", &roleIDs).Error
	if err != nil {
		return entity.DataScopeOrgOnly, err
	}
	
	if len(roleIDs) == 0 {
		return entity.DataScopeSelf, nil
	}
	
	// 查询最高权限的数据范围
	var dataScope string
	err = r.db.WithContext(ctx).Model(&entity.RoleEntity{}).
		Select("data_scope").
		Where("id IN ?", roleIDs).
		Order("CASE data_scope WHEN 'all' THEN 4 WHEN 'org_and_sub' THEN 3 WHEN 'org_only' THEN 2 ELSE 1 END DESC").
		Limit(1).
		Scan(&dataScope).Error
	
	if err != nil {
		return entity.DataScopeOrgOnly, err
	}
	
	return entity.DataScope(dataScope), nil
}

// GetUserPermissions 获取用户权限列表
func (r *PermissionRepositoryImpl) GetUserPermissions(ctx context.Context, userID string) ([]*entity.Permission, error) {
	var perms []*entity.Permission
	err := r.db.WithContext(ctx).
		Joins("JOIN ty_role_permissions ON ty_role_permissions.permission_id = ty_permissions.id").
		Joins("JOIN ty_user_roles ON ty_user_roles.role_id = ty_role_permissions.role_id").
		Where("ty_user_roles.user_id = ?", userID).
		Distinct().
		Find(&perms).Error
	return perms, err
}
