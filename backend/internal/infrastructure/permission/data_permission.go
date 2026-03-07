package permission

import (
	"context"
	"fmt"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/pkg/errors"
	"gorm.io/gorm"
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

// DataPermissionContext 数据权限上下文
type DataPermissionContext struct {
	UserID            string
	OrgID             string
	Role              string
	DataScope         DataScope
	IsSuperAdmin      bool
	AccessibleOrgIDs  []string
	CustomFilter      map[string]interface{}
}

// HasPermission 检查是否有权限访问特定组织的数据
func (ctx *DataPermissionContext) HasPermission(targetOrgID string) bool {
	if ctx.IsSuperAdmin {
		return true
	}
	
	if ctx.DataScope == DataScopeAll {
		return true
	}
	
	if ctx.DataScope == DataScopeSelf {
		return targetOrgID == ctx.OrgID
	}
	
	for _, orgID := range ctx.AccessibleOrgIDs {
		if orgID == targetOrgID {
			return true
		}
	}
	
	return false
}

// CanAccessAll 是否可以访问全部数据
func (ctx *DataPermissionContext) CanAccessAll() bool {
	return ctx.IsSuperAdmin || ctx.DataScope == DataScopeAll
}

// DataPermissionProvider 数据权限提供者接口
type DataPermissionProvider interface {
	// GetDataPermissionContext 获取数据权限上下文
	GetDataPermissionContext(ctx context.Context) (*DataPermissionContext, error)
	// GetAccessibleOrgs 获取可访问的组织列表
	GetAccessibleOrgs(ctx context.Context, orgID string, scope DataScope) ([]string, error)
	// GetSubOrgIDs 获取子组织ID列表
	GetSubOrgIDs(ctx context.Context, orgID string) ([]string, error)
}

// DefaultDataPermissionProvider 默认数据权限提供者
type DefaultDataPermissionProvider struct {
	db *gorm.DB
}

// NewDataPermissionProvider 创建数据权限提供者
func NewDataPermissionProvider(db *gorm.DB) DataPermissionProvider {
	return &DefaultDataPermissionProvider{db: db}
}

// GetDataPermissionContext 获取数据权限上下文
func (p *DefaultDataPermissionProvider) GetDataPermissionContext(ctx context.Context) (*DataPermissionContext, error) {
	// 从 context 中获取用户信息
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, errors.New(errors.CodeUnauthorized, "user not authenticated")
	}
	
	orgID, _ := ctx.Value("org_id").(string)
	role, _ := ctx.Value("user_role").(string)
	
	// 超级管理员有全部权限
	isSuperAdmin := role == string(entity.RoleSuperAdmin)
	
	// 确定数据范围
	var dataScope DataScope
	switch role {
	case string(entity.RoleSuperAdmin):
		dataScope = DataScopeAll
	case string(entity.RoleAdmin):
		dataScope = DataScopeOrgAndSub
	case string(entity.RoleManager):
		dataScope = DataScopeOrgAndSub
	default:
		dataScope = DataScopeOrgOnly
	}
	
	dpCtx := &DataPermissionContext{
		UserID:       userID,
		OrgID:        orgID,
		Role:         role,
		DataScope:    dataScope,
		IsSuperAdmin: isSuperAdmin,
	}
	
	// 获取可访问的组织列表
	if !isSuperAdmin && orgID != "" {
		accessibleOrgs, err := p.GetAccessibleOrgs(ctx, orgID, dataScope)
		if err != nil {
			return nil, err
		}
		dpCtx.AccessibleOrgIDs = accessibleOrgs
	}
	
	return dpCtx, nil
}

// GetAccessibleOrgs 获取可访问的组织列表
func (p *DefaultDataPermissionProvider) GetAccessibleOrgs(ctx context.Context, orgID string, scope DataScope) ([]string, error) {
	switch scope {
	case DataScopeAll:
		return nil, nil // nil 表示全部
	case DataScopeOrgOnly:
		return []string{orgID}, nil
	case DataScopeOrgAndSub:
		subOrgs, err := p.GetSubOrgIDs(ctx, orgID)
		if err != nil {
			return nil, err
		}
		return append([]string{orgID}, subOrgs...), nil
	case DataScopeSelf:
		return []string{orgID}, nil
	default:
		return []string{orgID}, nil
	}
}

// GetSubOrgIDs 获取子组织ID列表（递归）
func (p *DefaultDataPermissionProvider) GetSubOrgIDs(ctx context.Context, orgID string) ([]string, error) {
	var result []string
	var queue []string
	queue = append(queue, orgID)
	
	visited := make(map[string]bool)
	visited[orgID] = true
	
	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]
		
		var childIDs []string
		err := p.db.WithContext(ctx).
			Model(&entity.Organization{}).
			Where("parent_id = ?", currentID).
			Pluck("id", &childIDs).Error
		
		if err != nil {
			return nil, err
		}
		
		for _, childID := range childIDs {
			if !visited[childID] {
				visited[childID] = true
				result = append(result, childID)
				queue = append(queue, childID)
			}
		}
	}
	
	return result, nil
}

// DataPermissionFilter 数据权限过滤器
type DataPermissionFilter struct {
	provider DataPermissionProvider
}

// NewDataPermissionFilter 创建数据权限过滤器
func NewDataPermissionFilter(provider DataPermissionProvider) *DataPermissionFilter {
	return &DataPermissionFilter{provider: provider}
}

// Apply 应用数据权限过滤
func (f *DataPermissionFilter) Apply(ctx context.Context, db *gorm.DB, resourceOrgField string) (*gorm.DB, error) {
	dpCtx, err := f.provider.GetDataPermissionContext(ctx)
	if err != nil {
		return nil, err
	}
	
	// 超级管理员不过滤
	if dpCtx.CanAccessAll() {
		return db, nil
	}
	
	// 应用组织过滤
	if len(dpCtx.AccessibleOrgIDs) > 0 {
		return db.Where(fmt.Sprintf("%s IN ?", resourceOrgField), dpCtx.AccessibleOrgIDs), nil
	}
	
	// 无权限
	return db.Where("1=0"), nil
}

// ApplyWithCreator 应用包含创建者权限的过滤
func (f *DataPermissionFilter) ApplyWithCreator(ctx context.Context, db *gorm.DB, resourceOrgField, creatorField string) (*gorm.DB, error) {
	dpCtx, err := f.provider.GetDataPermissionContext(ctx)
	if err != nil {
		return nil, err
	}
	
	// 超级管理员不过滤
	if dpCtx.CanAccessAll() {
		return db, nil
	}
	
	// 数据范围过滤 + 本人创建的数据
	if len(dpCtx.AccessibleOrgIDs) > 0 {
		return db.Where(
			fmt.Sprintf("%s IN ? OR %s = ?", resourceOrgField, creatorField),
			dpCtx.AccessibleOrgIDs,
			dpCtx.UserID,
		), nil
	}
	
	// 只能看自己的
	return db.Where(fmt.Sprintf("%s = ?", creatorField), dpCtx.UserID), nil
}

// MustGetDataPermissionContext 必须获取数据权限上下文，否则panic
func MustGetDataPermissionContext(ctx context.Context, provider DataPermissionProvider) *DataPermissionContext {
	dpCtx, err := provider.GetDataPermissionContext(ctx)
	if err != nil {
		panic(err)
	}
	return dpCtx
}

// CheckDataPermission 检查数据权限
func CheckDataPermission(ctx context.Context, provider DataPermissionProvider, targetOrgID string) error {
	dpCtx, err := provider.GetDataPermissionContext(ctx)
	if err != nil {
		return err
	}
	
	if !dpCtx.HasPermission(targetOrgID) {
		return errors.ErrDataPermissionDenied
	}
	
	return nil
}
