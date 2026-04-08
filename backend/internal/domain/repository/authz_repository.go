package repository

import (
	"context"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
)

type PolicyRule struct {
	ID             string
	PermissionCode string
	OrgID          string
	ResourceType   string
	ScopeRule      string
	Effect         string
	Priority       int
	Enabled        bool
}

type RolePermission struct {
	ID             string
	Role           entity.Role
	PermissionCode string
	OrgID          string
	Effect         string
	Enabled        bool
	Priority       int
}

// AuthzPolicyRepository 权限策略仓储
type AuthzPolicyRepository interface {
	// ListRolePermissionCodes 返回角色允许的 permission code 列表
	ListRolePermissionCodes(ctx context.Context, role entity.Role) ([]string, error)
	// ListRolePermissionCodesForOrg 返回角色在指定组织上下文的有效 permission code 列表（含上级/全局继承）
	ListRolePermissionCodesForOrg(ctx context.Context, role entity.Role, orgID string) ([]string, error)
	// ListAllRolePermissions 返回全部角色权限映射
	ListAllRolePermissions(ctx context.Context) ([]RolePermission, error)
	// ReplaceRolePermissions 全量替换角色的 permission code（enabled=1）
	ReplaceRolePermissions(ctx context.Context, role entity.Role, permissionCodes []string) error
	// ReplaceRolePermissionsForOrg 全量替换指定组织作用域角色权限（orgID为空表示全局）
	ReplaceRolePermissionsForOrg(ctx context.Context, role entity.Role, orgID string, permissionCodes []string) error
	// ListPolicyRules 返回指定 permission 的策略规则（按优先级排序）
	ListPolicyRules(ctx context.Context, permissionCode string) ([]PolicyRule, error)
	// ListPolicyRulesForOrg 返回指定 permission 在组织上下文的有效策略规则
	ListPolicyRulesForOrg(ctx context.Context, permissionCode, orgID string) ([]PolicyRule, error)
	// ListAllPolicyRules 返回全部策略规则
	ListAllPolicyRules(ctx context.Context) ([]PolicyRule, error)
	// ReplacePolicyRules 全量替换指定 permission 的规则
	ReplacePolicyRules(ctx context.Context, permissionCode string, rules []PolicyRule) error
	// ReplacePolicyRulesForOrg 全量替换指定组织作用域策略规则（orgID为空表示全局）
	ReplacePolicyRulesForOrg(ctx context.Context, permissionCode, orgID string, rules []PolicyRule) error
	// CreateDecision 创建授权决策审计日志
	CreateDecision(ctx context.Context, decision *entity.AuthzDecisionLog) error
	// ListDecisions 查询授权决策审计日志
	ListDecisions(ctx context.Context, query *entity.AuthzDecisionQuery) (*AuthzDecisionPaginatedResult, error)
	// CreatePolicyChange 创建权限策略变更日志
	CreatePolicyChange(ctx context.Context, log *entity.AuthzPolicyChangeLog) error
	// ListPolicyChanges 查询权限策略变更日志
	ListPolicyChanges(ctx context.Context, query *entity.AuthzPolicyChangeQuery) (*AuthzPolicyChangePaginatedResult, error)
	// FindPolicyChangeByID 根据ID查询策略变更日志
	FindPolicyChangeByID(ctx context.Context, id string) (*entity.AuthzPolicyChangeLog, error)
	// ExistsRollbackForChange 是否已存在针对目标变更的回滚记录
	ExistsRollbackForChange(ctx context.Context, changeID string) (bool, error)
	// CreatePolicyChangeRequest 创建策略变更申请
	CreatePolicyChangeRequest(ctx context.Context, req *entity.AuthzPolicyChangeRequest) error
	// ListPolicyChangeRequests 查询策略变更申请
	ListPolicyChangeRequests(ctx context.Context, query *entity.AuthzPolicyChangeRequestQuery) (*AuthzPolicyChangeRequestPaginatedResult, error)
	// FindPolicyChangeRequestByID 根据ID查询策略变更申请
	FindPolicyChangeRequestByID(ctx context.Context, id string) (*entity.AuthzPolicyChangeRequest, error)
	// UpdatePolicyChangeRequest 更新策略变更申请
	UpdatePolicyChangeRequest(ctx context.Context, req *entity.AuthzPolicyChangeRequest) error
	// RejectExpiredPolicyChangeRequests 自动拒绝超时未审批申请
	RejectExpiredPolicyChangeRequests(ctx context.Context, before time.Time, reviewedBy, reviewNote string) (int64, error)
}

// AuthzDecisionPaginatedResult 授权决策分页结果
type AuthzDecisionPaginatedResult struct {
	List     []entity.AuthzDecisionLog `json:"list"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	PageSize int                       `json:"page_size"`
}

// AuthzPolicyChangePaginatedResult 权限策略变更分页结果
type AuthzPolicyChangePaginatedResult struct {
	List     []entity.AuthzPolicyChangeLog `json:"list"`
	Total    int64                         `json:"total"`
	Page     int                           `json:"page"`
	PageSize int                           `json:"page_size"`
}

// AuthzPolicyChangeRequestPaginatedResult 策略变更申请分页结果
type AuthzPolicyChangeRequestPaginatedResult struct {
	List     []entity.AuthzPolicyChangeRequest `json:"list"`
	Total    int64                             `json:"total"`
	Page     int                               `json:"page"`
	PageSize int                               `json:"page_size"`
}
