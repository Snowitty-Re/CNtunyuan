package repository

import (
	"context"
	"errors"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"gorm.io/gorm"
)

// OrganizationRepositoryImpl 组织仓储实现
type OrganizationRepositoryImpl struct {
	*BaseRepository[entity.Organization]
}

// NewOrganizationRepository 创建组织仓储
func NewOrganizationRepository(db *gorm.DB) repository.OrganizationRepository {
	return &OrganizationRepositoryImpl{
		BaseRepository: NewBaseRepository[entity.Organization](db),
	}
}

// FindByCode 根据编码查找
func (r *OrganizationRepositoryImpl) FindByCode(ctx context.Context, code string) (*entity.Organization, error) {
	var org entity.Organization
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&org).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("组织不存在")
		}
		return nil, err
	}
	return &org, nil
}

// FindByParentID 根据父ID查找子组织
func (r *OrganizationRepositoryImpl) FindByParentID(ctx context.Context, parentID string) ([]entity.Organization, error) {
	var orgs []entity.Organization
	err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Order("sort_order ASC, created_at ASC").
		Find(&orgs).Error
	return orgs, err
}

// FindRoot 查找根组织
func (r *OrganizationRepositoryImpl) FindRoot(ctx context.Context) (*entity.Organization, error) {
	var org entity.Organization
	err := r.db.WithContext(ctx).
		Where("parent_id IS NULL OR type = ?", entity.OrgTypeRoot).
		First(&org).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("根组织不存在")
		}
		return nil, err
	}
	return &org, nil
}

// FindTree 获取组织树
func (r *OrganizationRepositoryImpl) FindTree(ctx context.Context, rootID string) (*entity.OrgTreeNode, error) {
	// 获取根节点
	var root entity.Organization
	if err := r.db.WithContext(ctx).First(&root, "id = ?", rootID).Error; err != nil {
		return nil, err
	}

	node := &entity.OrgTreeNode{
		Organization: root,
	}

	// 递归获取子节点
	if err := r.buildTree(ctx, node); err != nil {
		return nil, err
	}

	return node, nil
}

// buildTree 递归构建树
func (r *OrganizationRepositoryImpl) buildTree(ctx context.Context, node *entity.OrgTreeNode) error {
	children, err := r.FindByParentID(ctx, node.ID)
	if err != nil {
		return err
	}

	for _, child := range children {
		childNode := &entity.OrgTreeNode{
			Organization: child,
		}
		if err := r.buildTree(ctx, childNode); err != nil {
			return err
		}
		node.Children = append(node.Children, childNode)
	}

	return nil
}

// FindPath 查找组织路径
func (r *OrganizationRepositoryImpl) FindPath(ctx context.Context, orgID string) ([]entity.Organization, error) {
	var path []entity.Organization
	currentID := orgID

	for currentID != "" {
		var org entity.Organization
		if err := r.db.WithContext(ctx).First(&org, "id = ?", currentID).Error; err != nil {
			break
		}
		path = append([]entity.Organization{org}, path...)

		if org.ParentID == nil {
			break
		}
		currentID = *org.ParentID
	}

	return path, nil
}

// List 分页查询
func (r *OrganizationRepositoryImpl) List(ctx context.Context, query *repository.OrgQuery) (*repository.PageResult[entity.Organization], error) {
	var orgs []entity.Organization
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Organization{})

	// 关键词搜索
	if query.Keyword != "" {
		db = db.Where("name LIKE ? OR code LIKE ?", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}

	// 类型筛选
	if query.Type != "" {
		db = db.Where("type = ?", query.Type)
	}

	// 状态筛选
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	// 父组织筛选
	if query.ParentID != "" {
		db = db.Where("parent_id = ?", query.ParentID)
	}

	// 层级筛选
	if query.Level > 0 {
		db = db.Where("level = ?", query.Level)
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	if err := db.Order("sort_order ASC, created_at ASC").
		Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).
		Find(&orgs).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(orgs, total, query.Page, query.PageSize), nil
}

// FindChildren 获取所有子组织（递归）
func (r *OrganizationRepositoryImpl) FindChildren(ctx context.Context, parentID string) ([]entity.Organization, error) {
	var allChildren []entity.Organization
	queue := []string{parentID}

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		children, err := r.FindByParentID(ctx, currentID)
		if err != nil {
			return nil, err
		}

		for _, child := range children {
			allChildren = append(allChildren, child)
			queue = append(queue, child.ID)
		}
	}

	return allChildren, nil
}

// UpdateStats 更新统计信息
func (r *OrganizationRepositoryImpl) UpdateStats(ctx context.Context, orgID string, stats *entity.OrgStats) error {
	return r.db.WithContext(ctx).
		Model(&entity.OrgStats{}).
		Where("org_id = ?", orgID).
		Save(stats).Error
}

// Move 移动组织
func (r *OrganizationRepositoryImpl) Move(ctx context.Context, orgID, newParentID string) error {
	return r.db.WithContext(ctx).
		Model(&entity.Organization{}).
		Where("id = ?", orgID).
		Update("parent_id", newParentID).
		Error
}

// ExistsCode 检查编码是否存在
func (r *OrganizationRepositoryImpl) ExistsCode(ctx context.Context, code string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Organization{}).
		Where("code = ?", code).
		Count(&count).Error
	return count > 0, err
}
