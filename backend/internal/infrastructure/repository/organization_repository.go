package repository

import (
	"context"
	"errors"
	"strings"

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

// Delete 硬删除组织（组织删除应真正释放唯一编码）
func (r *OrganizationRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.Organization{}, "id = ?", id).Error
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

	// 一次查出全部组织，内存建树，避免 N+1
	var all []entity.Organization
	if err := r.db.WithContext(ctx).Order("sort_order ASC, created_at ASC").Find(&all).Error; err != nil {
		return nil, err
	}

	childrenMap := make(map[string][]entity.Organization)
	for _, org := range all {
		if org.ParentID == nil || *org.ParentID == "" {
			continue
		}
		pid := *org.ParentID
		childrenMap[pid] = append(childrenMap[pid], org)
	}

	node := &entity.OrgTreeNode{Organization: root}
	r.buildTreeFromMap(node, childrenMap)
	return node, nil
}

// buildTreeFromMap 基于预加载 map 构建树
func (r *OrganizationRepositoryImpl) buildTreeFromMap(node *entity.OrgTreeNode, childrenMap map[string][]entity.Organization) {
	for _, child := range childrenMap[node.ID] {
		childNode := &entity.OrgTreeNode{Organization: child}
		r.buildTreeFromMap(childNode, childrenMap)
		node.Children = append(node.Children, childNode)
	}
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

// FindChildren 获取所有子组织（递归，单次查询后内存 BFS）
func (r *OrganizationRepositoryImpl) FindChildren(ctx context.Context, parentID string) ([]entity.Organization, error) {
	var all []entity.Organization
	if err := r.db.WithContext(ctx).Find(&all).Error; err != nil {
		return nil, err
	}

	childrenMap := make(map[string][]entity.Organization)
	for _, org := range all {
		if org.ParentID == nil || *org.ParentID == "" {
			continue
		}
		pid := *org.ParentID
		childrenMap[pid] = append(childrenMap[pid], org)
	}

	var allChildren []entity.Organization
	queue := []string{parentID}
	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]
		for _, child := range childrenMap[currentID] {
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
	db := r.db.WithContext(ctx).
		Model(&entity.Organization{}).
		Where("id = ?", orgID)

	if strings.TrimSpace(newParentID) == "" {
		return db.Update("parent_id", nil).Error
	}

	return db.Update("parent_id", newParentID).Error
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

// PurgeSoftDeletedByCode 清理指定编码的软删除残留记录
func (r *OrganizationRepositoryImpl) PurgeSoftDeletedByCode(ctx context.Context, code string) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Where("code = ? AND deleted_at IS NOT NULL", code).
		Delete(&entity.Organization{}).Error
}
