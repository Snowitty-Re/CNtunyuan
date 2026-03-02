package repository

import (
	"context"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
)

// OrganizationRepository 组织仓储接口
type OrganizationRepository interface {
	Repository[entity.Organization]

	// FindByCode 根据编码查找
	FindByCode(ctx context.Context, code string) (*entity.Organization, error)

	// FindByParentID 根据父ID查找子组织
	FindByParentID(ctx context.Context, parentID string) ([]entity.Organization, error)

	// FindRoot 查找根组织
	FindRoot(ctx context.Context) (*entity.Organization, error)

	// FindTree 获取组织树
	FindTree(ctx context.Context, rootID string) (*entity.OrgTreeNode, error)

	// FindPath 查找组织路径
	FindPath(ctx context.Context, orgID string) ([]entity.Organization, error)

	// List 分页查询
	List(ctx context.Context, query *OrgQuery) (*PageResult[entity.Organization], error)

	// FindChildren 获取所有子组织（递归）
	FindChildren(ctx context.Context, parentID string) ([]entity.Organization, error)

	// UpdateStats 更新统计信息
	UpdateStats(ctx context.Context, orgID string, stats *entity.OrgStats) error

	// Move 移动组织
	Move(ctx context.Context, orgID, newParentID string) error

	// ExistsCode 检查编码是否存在
	ExistsCode(ctx context.Context, code string) (bool, error)
}

// OrgQuery 组织查询参数
type OrgQuery struct {
	Pagination
	Keyword   string              `json:"keyword"`
	Type      entity.OrgType      `json:"type"`
	Status    entity.OrgStatus    `json:"status"`
	ParentID  string              `json:"parent_id"`
	Level     int                 `json:"level"`
}

// NewOrgQuery 创建默认组织查询
func NewOrgQuery() *OrgQuery {
	return &OrgQuery{
		Pagination: Pagination{
			Page:     1,
			PageSize: 10,
		},
	}
}
