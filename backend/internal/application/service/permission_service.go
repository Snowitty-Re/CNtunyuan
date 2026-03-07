package service

import (
	"context"
	"errors"
	"reflect"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/google/uuid"
)

var (
	ErrPermissionNotFound = errors.New("permission not found")
	ErrRoleNotFound       = errors.New("role not found")
	ErrRoleExists         = errors.New("role already exists")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrFieldNotReadable   = errors.New("field not readable")
	ErrFieldNotWritable   = errors.New("field not writable")
)

// PermissionService 权限服务接口
type PermissionService interface {
	// Permission operations
	CreatePermission(ctx context.Context, req *dto.CreatePermissionRequest) (*dto.PermissionResponse, error)
	UpdatePermission(ctx context.Context, id string, req *dto.UpdatePermissionRequest) (*dto.PermissionResponse, error)
	DeletePermission(ctx context.Context, id string) error
	GetPermission(ctx context.Context, id string) (*dto.PermissionResponse, error)
	ListPermissions(ctx context.Context, req *dto.ListPermissionsRequest) (*dto.ListPermissionsResponse, error)
	
	// Role operations
	CreateRole(ctx context.Context, req *dto.CreateRoleRequest) (*dto.RoleResponse, error)
	UpdateRole(ctx context.Context, id string, req *dto.UpdateRoleRequest) (*dto.RoleResponse, error)
	DeleteRole(ctx context.Context, id string) error
	GetRole(ctx context.Context, id string) (*dto.RoleResponse, error)
	ListRoles(ctx context.Context, req *dto.ListRolesRequest) (*dto.ListRolesResponse, error)
	
	// Role permission operations
	GrantPermissions(ctx context.Context, roleID string, req *dto.GrantPermissionsRequest) error
	RevokePermissions(ctx context.Context, roleID string, req *dto.RevokePermissionsRequest) error
	GetRolePermissions(ctx context.Context, roleID string) (*dto.RolePermissionsResponse, error)
	
	// User role operations
	AssignRole(ctx context.Context, userID string, req *dto.AssignRoleRequest) error
	RemoveRole(ctx context.Context, userID string, roleID string) error
	GetUserRoles(ctx context.Context, userID string) (*dto.UserRolesResponse, error)
	
	// Field permission operations
	SetFieldPermission(ctx context.Context, req *dto.SetFieldPermissionRequest) error
	GetFieldPermissions(ctx context.Context, resource, roleID string) (*dto.FieldPermissionsResponse, error)
	
	// Check operations
	CheckPermission(ctx context.Context, userID string, resource entity.PermissionResource, action entity.PermissionAction) (*entity.PermissionCheckResult, error)
	CheckDataPermission(ctx context.Context, userID string, orgID string) error
	
	// Filter operations
	FilterFields(ctx context.Context, userID string, resource string, data interface{}) (interface{}, error)
	
	// Init operations
	InitSystemPermissions(ctx context.Context) error
	InitSystemRoles(ctx context.Context, orgID string) error
}

// PermissionAppService 权限应用服务
type PermissionAppService struct {
	permRepo repository.PermissionRepository
}

// NewPermissionAppService 创建权限服务
func NewPermissionAppService(permRepo repository.PermissionRepository) *PermissionAppService {
	return &PermissionAppService{
		permRepo: permRepo,
	}
}

// CreatePermission 创建权限
func (s *PermissionAppService) CreatePermission(ctx context.Context, req *dto.CreatePermissionRequest) (*dto.PermissionResponse, error) {
	perm := entity.NewPermission(req.Name, req.Resource, req.Action)
	perm.Description = req.Description
	perm.Category = req.Category
	perm.IsSystem = false
	
	if err := s.permRepo.CreatePermission(ctx, perm); err != nil {
		logger.Error("Failed to create permission", logger.Err(err))
		return nil, err
	}
	
	resp := dto.ToPermissionResponse(perm)
	return &resp, nil
}

// UpdatePermission 更新权限
func (s *PermissionAppService) UpdatePermission(ctx context.Context, id string, req *dto.UpdatePermissionRequest) (*dto.PermissionResponse, error) {
	perm, err := s.permRepo.FindPermissionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if perm == nil {
		return nil, ErrPermissionNotFound
	}
	
	if perm.IsSystem {
		return nil, errors.New("cannot modify system permission")
	}
	
	if req.Name != "" {
		perm.Name = req.Name
	}
	if req.Description != "" {
		perm.Description = req.Description
	}
	if req.Category != "" {
		perm.Category = req.Category
	}
	
	if err := s.permRepo.UpdatePermission(ctx, perm); err != nil {
		logger.Error("Failed to update permission", logger.Err(err))
		return nil, err
	}
	
	resp := dto.ToPermissionResponse(perm)
	return &resp, nil
}

// DeletePermission 删除权限
func (s *PermissionAppService) DeletePermission(ctx context.Context, id string) error {
	perm, err := s.permRepo.FindPermissionByID(ctx, id)
	if err != nil {
		return err
	}
	if perm == nil {
		return ErrPermissionNotFound
	}
	
	if perm.IsSystem {
		return errors.New("cannot delete system permission")
	}
	
	return s.permRepo.DeletePermission(ctx, id)
}

// GetPermission 获取权限
func (s *PermissionAppService) GetPermission(ctx context.Context, id string) (*dto.PermissionResponse, error) {
	perm, err := s.permRepo.FindPermissionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if perm == nil {
		return nil, ErrPermissionNotFound
	}
	
	resp := dto.ToPermissionResponse(perm)
	return &resp, nil
}

// ListPermissions 列表查询权限
func (s *PermissionAppService) ListPermissions(ctx context.Context, req *dto.ListPermissionsRequest) (*dto.ListPermissionsResponse, error) {
	perms, total, err := s.permRepo.ListPermissions(ctx, req.Resource, req.Category, req.Page, req.PageSize)
	if err != nil {
		logger.Error("Failed to list permissions", logger.Err(err))
		return nil, err
	}
	
	items := make([]dto.PermissionResponse, len(perms))
	for i, perm := range perms {
		items[i] = dto.ToPermissionResponse(perm)
	}
	
	return &dto.ListPermissionsResponse{
		List:     items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// CreateRole 创建角色
func (s *PermissionAppService) CreateRole(ctx context.Context, req *dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	// 检查code是否已存在
	existing, err := s.permRepo.FindRoleByCode(ctx, req.Code, req.OrgID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrRoleExists
	}
	
	roleType := entity.RoleTypeCustom
	if req.IsSystem {
		roleType = entity.RoleTypeSystem
	}
	
	role := entity.NewRole(req.Name, req.Code, req.OrgID, roleType)
	role.Description = req.Description
	role.DataScope = entity.DataScope(req.DataScope)
	role.SortOrder = req.SortOrder
	
	if err := s.permRepo.CreateRole(ctx, role); err != nil {
		logger.Error("Failed to create role", logger.Err(err))
		return nil, err
	}
	
	// 分配权限
	if len(req.PermissionIDs) > 0 {
		if err := s.permRepo.BatchGrantPermissions(ctx, role.ID, req.PermissionIDs, req.CreatedBy); err != nil {
			logger.Error("Failed to grant permissions", logger.Err(err))
		}
	}
	
	resp := dto.ToRoleResponse(role)
	return &resp, nil
}

// UpdateRole 更新角色
func (s *PermissionAppService) UpdateRole(ctx context.Context, id string, req *dto.UpdateRoleRequest) (*dto.RoleResponse, error) {
	role, err := s.permRepo.FindRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, ErrRoleNotFound
	}
	
	if role.IsSystemRole() && !req.IsSystem {
		return nil, errors.New("cannot modify system role")
	}
	
	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	if req.DataScope != "" {
		role.DataScope = entity.DataScope(req.DataScope)
	}
	if req.SortOrder > 0 {
		role.SortOrder = req.SortOrder
	}
	
	if err := s.permRepo.UpdateRole(ctx, role); err != nil {
		logger.Error("Failed to update role", logger.Err(err))
		return nil, err
	}
	
	resp := dto.ToRoleResponse(role)
	return &resp, nil
}

// DeleteRole 删除角色
func (s *PermissionAppService) DeleteRole(ctx context.Context, id string) error {
	role, err := s.permRepo.FindRoleByID(ctx, id)
	if err != nil {
		return err
	}
	if role == nil {
		return ErrRoleNotFound
	}
	
	if role.IsSystemRole() {
		return errors.New("cannot delete system role")
	}
	
	return s.permRepo.DeleteRole(ctx, id)
}

// GetRole 获取角色
func (s *PermissionAppService) GetRole(ctx context.Context, id string) (*dto.RoleResponse, error) {
	role, err := s.permRepo.FindRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, ErrRoleNotFound
	}
	
	resp := dto.ToRoleResponse(role)
	return &resp, nil
}

// ListRoles 列表查询角色
func (s *PermissionAppService) ListRoles(ctx context.Context, req *dto.ListRolesRequest) (*dto.ListRolesResponse, error) {
	query := &entity.RoleQuery{
		OrgID:    req.OrgID,
		Type:     entity.RoleType(req.Type),
		Status:   req.Status,
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	
	roles, total, err := s.permRepo.ListRoles(ctx, query)
	if err != nil {
		logger.Error("Failed to list roles", logger.Err(err))
		return nil, err
	}
	
	items := make([]dto.RoleResponse, len(roles))
	for i, role := range roles {
		items[i] = dto.ToRoleResponse(role)
	}
	
	return &dto.ListRolesResponse{
		List:     items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GrantPermissions 授予权限
func (s *PermissionAppService) GrantPermissions(ctx context.Context, roleID string, req *dto.GrantPermissionsRequest) error {
	role, err := s.permRepo.FindRoleByID(ctx, roleID)
	if err != nil {
		return err
	}
	if role == nil {
		return ErrRoleNotFound
	}
	
	return s.permRepo.BatchGrantPermissions(ctx, roleID, req.PermissionIDs, req.GrantedBy)
}

// RevokePermissions 撤销权限
func (s *PermissionAppService) RevokePermissions(ctx context.Context, roleID string, req *dto.RevokePermissionsRequest) error {
	return s.permRepo.BatchRevokePermissions(ctx, roleID, req.PermissionIDs)
}

// GetRolePermissions 获取角色权限
func (s *PermissionAppService) GetRolePermissions(ctx context.Context, roleID string) (*dto.RolePermissionsResponse, error) {
	perms, err := s.permRepo.ListPermissionsByRoleID(ctx, roleID)
	if err != nil {
		return nil, err
	}
	
	items := make([]dto.PermissionResponse, len(perms))
	for i, perm := range perms {
		items[i] = dto.ToPermissionResponse(perm)
	}
	
	return &dto.RolePermissionsResponse{
		Permissions: items,
	}, nil
}

// AssignRole 分配角色给用户
func (s *PermissionAppService) AssignRole(ctx context.Context, userID string, req *dto.AssignRoleRequest) error {
	return s.permRepo.AssignRoleToUser(ctx, userID, req.RoleID, req.OrgID, req.AssignedBy)
}

// RemoveRole 移除用户角色
func (s *PermissionAppService) RemoveRole(ctx context.Context, userID string, roleID string) error {
	return s.permRepo.RemoveRoleFromUser(ctx, userID, roleID)
}

// GetUserRoles 获取用户角色
func (s *PermissionAppService) GetUserRoles(ctx context.Context, userID string) (*dto.UserRolesResponse, error) {
	roles, err := s.permRepo.ListRolesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	items := make([]dto.RoleResponse, len(roles))
	for i, role := range roles {
		items[i] = dto.ToRoleResponse(role)
	}
	
	return &dto.UserRolesResponse{
		Roles: items,
	}, nil
}

// SetFieldPermission 设置字段权限
func (s *PermissionAppService) SetFieldPermission(ctx context.Context, req *dto.SetFieldPermissionRequest) error {
	fp := &entity.FieldPermission{
		BaseEntity: entity.BaseEntity{
			ID: uuid.New().String(),
		},
		Resource:   req.Resource,
		Field:      req.Field,
		RoleID:     req.RoleID,
		Permission: req.Permission,
		Condition:  req.Condition,
	}
	
	// 检查是否已存在
	existing, err := s.permRepo.FindFieldPermission(ctx, req.Resource, req.Field, req.RoleID)
	if err != nil {
		return err
	}
	
	if existing != nil {
		existing.Permission = req.Permission
		existing.Condition = req.Condition
		return s.permRepo.UpdateFieldPermission(ctx, existing)
	}
	
	return s.permRepo.CreateFieldPermission(ctx, fp)
}

// GetFieldPermissions 获取字段权限
func (s *PermissionAppService) GetFieldPermissions(ctx context.Context, resource, roleID string) (*dto.FieldPermissionsResponse, error) {
	fps, err := s.permRepo.ListFieldPermissions(ctx, resource, roleID)
	if err != nil {
		return nil, err
	}
	
	items := make([]dto.FieldPermissionResponse, len(fps))
	for i, fp := range fps {
		items[i] = dto.FieldPermissionResponse{
			ID:         fp.ID,
			Resource:   fp.Resource,
			Field:      fp.Field,
			RoleID:     fp.RoleID,
			Permission: fp.Permission,
			Condition:  fp.Condition,
		}
	}
	
	return &dto.FieldPermissionsResponse{
		Permissions: items,
	}, nil
}

// CheckPermission 检查用户权限
func (s *PermissionAppService) CheckPermission(ctx context.Context, userID string, resource entity.PermissionResource, action entity.PermissionAction) (*entity.PermissionCheckResult, error) {
	allowed, err := s.permRepo.CheckUserPermission(ctx, userID, resource, action)
	if err != nil {
		return nil, err
	}
	
	result := &entity.PermissionCheckResult{
		Allowed: allowed,
	}
	
	if !allowed {
		result.Reason = "user does not have required permission"
		result.MissingPerm = entity.GetCode(resource, action)
	} else {
		// 获取数据范围
		dataScope, err := s.permRepo.GetUserDataScope(ctx, userID)
		if err != nil {
			logger.Warn("Failed to get user data scope", logger.Err(err))
		}
		result.DataScope = dataScope
	}
	
	return result, nil
}

// CheckDataPermission 检查数据权限
func (s *PermissionAppService) CheckDataPermission(ctx context.Context, userID string, targetOrgID string) error {
	// 获取用户数据范围
	dataScope, err := s.permRepo.GetUserDataScope(ctx, userID)
	if err != nil {
		return err
	}
	
	if dataScope == entity.DataScopeAll {
		return nil
	}
	
	// TODO: 根据数据范围检查是否有权限访问目标组织数据
	// 这里简化处理，实际应该查询用户的组织层级关系
	
	return nil
}

// FilterFields 过滤字段
func (s *PermissionAppService) FilterFields(ctx context.Context, userID string, resource string, data interface{}) (interface{}, error) {
	// 获取用户角色
	roles, err := s.permRepo.ListRolesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	// 获取所有字段权限
	fieldPerms := make(map[string]string) // field -> permission
	for _, role := range roles {
		fps, err := s.permRepo.GetFieldPermissionsForRole(ctx, role.ID)
		if err != nil {
			continue
		}
		if resPerms, ok := fps[resource]; ok {
			for field, perm := range resPerms {
				// 使用最宽松的权限
				if existing, exists := fieldPerms[field]; !exists || perm == "write" || (perm == "read" && existing == "none") {
					fieldPerms[field] = perm
				}
			}
		}
	}
	
	// 过滤字段
	return s.filterStructFields(data, fieldPerms), nil
}

// filterStructFields 过滤结构体字段
func (s *PermissionAppService) filterStructFields(data interface{}, fieldPerms map[string]string) interface{} {
	if data == nil {
		return nil
	}
	
	// 处理 map
	if m, ok := data.(map[string]interface{}); ok {
		filtered := make(map[string]interface{})
		for k, v := range m {
			if perm, exists := fieldPerms[k]; !exists || perm != "none" {
				filtered[k] = v
			}
		}
		return filtered
	}
	
	// 处理结构体
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	if val.Kind() != reflect.Struct {
		return data
	}
	
	// 转换为 map 过滤
	m := make(map[string]interface{})
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)
		
		// 获取 json tag
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = field.Name
		}
		if jsonTag == "-" {
			continue
		}
		
		// 检查权限
		if perm, exists := fieldPerms[jsonTag]; !exists || perm != "none" {
			m[jsonTag] = fieldValue.Interface()
		}
	}
	
	return m
}

// InitSystemPermissions 初始化系统权限
func (s *PermissionAppService) InitSystemPermissions(ctx context.Context) error {
	perms := entity.GenerateSystemPermissions()
	
	for _, perm := range perms {
		existing, err := s.permRepo.FindPermissionByCode(ctx, perm.Code)
		if err != nil {
			return err
		}
		if existing == nil {
			if err := s.permRepo.CreatePermission(ctx, perm); err != nil {
				logger.Error("Failed to create system permission", logger.String("code", perm.Code), logger.Err(err))
			}
		}
	}
	
	logger.Info("System permissions initialized")
	return nil
}

// InitSystemRoles 初始化系统角色
func (s *PermissionAppService) InitSystemRoles(ctx context.Context, orgID string) error {
	roles := entity.GenerateSystemRoles(orgID)
	
	for _, role := range roles {
		existing, err := s.permRepo.FindRoleByCode(ctx, role.Code, orgID)
		if err != nil {
			return err
		}
		if existing == nil {
			if err := s.permRepo.CreateRole(ctx, role); err != nil {
				logger.Error("Failed to create system role", logger.String("code", role.Code), logger.Err(err))
			}
		}
	}
	
	logger.Info("System roles initialized", logger.String("org_id", orgID))
	return nil
}
