package repository

import (
	"context"

	"github.com/Snowitty-Re/CNtunyuan/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OrganizationRepository 组织仓库
type OrganizationRepository struct {
	db *gorm.DB
}

// NewOrganizationRepository 创建组织仓库
func NewOrganizationRepository(db *gorm.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

// Create 创建组织
func (r *OrganizationRepository) Create(ctx context.Context, org *model.Organization) error {
	return r.db.WithContext(ctx).Create(org).Error
}

// GetByID 根据ID获取组织
func (r *OrganizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	var org model.Organization
	err := r.db.WithContext(ctx).Preload("Parent").Preload("Leader").First(&org, id).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// GetByCode 根据编码获取组织
func (r *OrganizationRepository) GetByCode(ctx context.Context, code string) (*model.Organization, error) {
	var org model.Organization
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// Update 更新组织
func (r *OrganizationRepository) Update(ctx context.Context, org *model.Organization) error {
	return r.db.WithContext(ctx).Save(org).Error
}

// Delete 删除组织
func (r *OrganizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Organization{}, id).Error
}

// GetTree 获取组织架构树
func (r *OrganizationRepository) GetTree(ctx context.Context, parentID *uuid.UUID) ([]*model.Organization, error) {
	var orgs []*model.Organization
	query := r.db.WithContext(ctx).Preload("Children").Preload("Leader")
	
	if parentID != nil {
		query = query.Where("parent_id = ?", parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}
	
	err := query.Order("sort ASC, created_at ASC").Find(&orgs).Error
	return orgs, err
}

// GetChildren 获取子组织
func (r *OrganizationRepository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*model.Organization, error) {
	var orgs []*model.Organization
	err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Order("sort ASC").Find(&orgs).Error
	return orgs, err
}

// GetAllChildren 获取所有子组织(递归)
func (r *OrganizationRepository) GetAllChildren(ctx context.Context, parentID uuid.UUID) ([]*model.Organization, error) {
	var allOrgs []*model.Organization
	
	var fetchChildren func(uuid.UUID) error
	fetchChildren = func(pid uuid.UUID) error {
		children, err := r.GetChildren(ctx, pid)
		if err != nil {
			return err
		}
		for _, child := range children {
			allOrgs = append(allOrgs, child)
			if err := fetchChildren(child.ID); err != nil {
				return err
			}
		}
		return nil
	}
	
	if err := fetchChildren(parentID); err != nil {
		return nil, err
	}
	
	return allOrgs, nil
}

// List 获取组织列表
func (r *OrganizationRepository) List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*model.Organization, int64, error) {
	var orgs []*model.Organization
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Organization{})

	for key, value := range filters {
		if value != nil && value != "" {
			query = query.Where(key+" = ?", value)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Parent").Preload("Leader").Offset(offset).Limit(pageSize).Order("sort ASC").Find(&orgs).Error; err != nil {
		return nil, 0, err
	}

	return orgs, total, nil
}

// UpdateVolunteerCount 更新志愿者数量
func (r *OrganizationRepository) UpdateVolunteerCount(ctx context.Context, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.Organization{}).Where("id = ?", orgID).UpdateColumn("volunteer_count", 
		r.db.Model(&model.User{}).Where("org_id = ?", orgID).Select("count(*)"),
	).Error
}
