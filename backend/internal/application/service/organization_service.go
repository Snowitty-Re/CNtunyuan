package service

import (
	"context"
	"errors"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
)

var (
	ErrOrganizationNotFound = errors.New("organization not found")
	ErrOrganizationExists   = errors.New("organization already exists")
	ErrInvalidOrgType       = errors.New("invalid organization type")
	ErrCannotDeleteOrg      = errors.New("cannot delete organization with children")
)

// OrganizationAppService 组织应用服务
type OrganizationAppService struct {
	orgRepo repository.OrganizationRepository
}

// NewOrganizationAppService 创建组织应用服务
func NewOrganizationAppService(orgRepo repository.OrganizationRepository) *OrganizationAppService {
	return &OrganizationAppService{orgRepo: orgRepo}
}

// Create 创建组织
func (s *OrganizationAppService) Create(ctx context.Context, req *dto.CreateOrganizationRequest) (*dto.OrganizationResponse, error) {
	// 检查编码是否已存在
	exists, err := s.orgRepo.ExistsCode(ctx, req.Code)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrOrganizationExists
	}

	// 验证组织类型
	orgType := entity.OrgType(req.Type)
	if !isValidOrgType(orgType) {
		return nil, ErrInvalidOrgType
	}

	// 创建组织实体
	var parentID *string
	if req.ParentID != "" {
		parentID = &req.ParentID
	}

	org, err := entity.NewOrganization(req.Name, req.Code, orgType, parentID)
	if err != nil {
		return nil, err
	}

	org.Description = req.Description
	org.Address = req.Address
	org.ContactName = req.ContactName
	org.ContactPhone = req.ContactPhone
	org.SortOrder = req.SortOrder

	// 保存
	if err := s.orgRepo.Create(ctx, org); err != nil {
		logger.Error("Failed to create organization", logger.Err(err))
		return nil, err
	}

	logger.Info("Organization created", logger.String("org_id", org.ID), logger.String("code", org.Code))

	resp := dto.ToOrganizationResponse(org)
	return &resp, nil
}

// GetByID 根据ID获取组织
func (s *OrganizationAppService) GetByID(ctx context.Context, id string) (*dto.OrganizationResponse, error) {
	org, err := s.orgRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrOrganizationNotFound
	}

	resp := dto.ToOrganizationResponse(org)
	return &resp, nil
}

// GetByCode 根据编码获取组织
func (s *OrganizationAppService) GetByCode(ctx context.Context, code string) (*dto.OrganizationResponse, error) {
	org, err := s.orgRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, ErrOrganizationNotFound
	}

	resp := dto.ToOrganizationResponse(org)
	return &resp, nil
}

// List 组织列表
func (s *OrganizationAppService) List(ctx context.Context, req *dto.OrganizationListRequest) (*dto.OrganizationListResponse, error) {
	query := repository.NewOrgQuery()
	query.Page = req.Page
	query.PageSize = req.PageSize
	query.Keyword = req.Keyword
	query.Type = entity.OrgType(req.Type)
	query.Status = entity.OrgStatus(req.Status)
	query.ParentID = req.ParentID

	result, err := s.orgRepo.List(ctx, query)
	if err != nil {
		return nil, err
	}

	list := make([]dto.OrganizationResponse, len(result.List))
	for i, org := range result.List {
		list[i] = dto.ToOrganizationResponse(&org)
	}

	resp := dto.NewOrganizationListResponse(list, result.Total, result.Page, result.PageSize)
	return &resp, nil
}

// Update 更新组织
func (s *OrganizationAppService) Update(ctx context.Context, id string, req *dto.UpdateOrganizationRequest) (*dto.OrganizationResponse, error) {
	org, err := s.orgRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrOrganizationNotFound
	}

	// 如果修改了编码，检查是否冲突
	if req.Code != "" && req.Code != org.Code {
		exists, err := s.orgRepo.ExistsCode(ctx, req.Code)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrOrganizationExists
		}
		org.Code = req.Code
	}

	if req.Name != "" {
		org.Name = req.Name
	}
	if req.Description != "" {
		org.Description = req.Description
	}
	if req.Address != "" {
		org.Address = req.Address
	}
	if req.ContactName != "" {
		org.ContactName = req.ContactName
	}
	if req.ContactPhone != "" {
		org.ContactPhone = req.ContactPhone
	}
	if req.Status != "" {
		org.Status = entity.OrgStatus(req.Status)
	}
	if req.SortOrder != 0 {
		org.SortOrder = req.SortOrder
	}

	if err := s.orgRepo.Update(ctx, org); err != nil {
		logger.Error("Failed to update organization", logger.Err(err))
		return nil, err
	}

	logger.Info("Organization updated", logger.String("org_id", org.ID))

	resp := dto.ToOrganizationResponse(org)
	return &resp, nil
}

// Delete 删除组织
func (s *OrganizationAppService) Delete(ctx context.Context, id string) error {
	// 检查是否存在子组织
	children, err := s.orgRepo.FindByParentID(ctx, id)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		return ErrCannotDeleteOrg
	}

	if err := s.orgRepo.SoftDelete(ctx, id); err != nil {
		logger.Error("Failed to delete organization", logger.Err(err))
		return err
	}

	logger.Info("Organization deleted", logger.String("org_id", id))
	return nil
}

// GetTree 获取组织树
func (s *OrganizationAppService) GetTree(ctx context.Context, rootID string) (*dto.OrganizationTreeResponse, error) {
	if rootID == "" {
		// 获取根组织
		root, err := s.orgRepo.FindRoot(ctx)
		if err != nil {
			return nil, err
		}
		rootID = root.ID
	}

	tree, err := s.orgRepo.FindTree(ctx, rootID)
	if err != nil {
		return nil, err
	}

	return dto.ToOrganizationTreeResponse(tree), nil
}

// GetChildren 获取子组织
func (s *OrganizationAppService) GetChildren(ctx context.Context, parentID string) ([]dto.OrganizationResponse, error) {
	orgs, err := s.orgRepo.FindByParentID(ctx, parentID)
	if err != nil {
		return nil, err
	}

	list := make([]dto.OrganizationResponse, len(orgs))
	for i, org := range orgs {
		list[i] = dto.ToOrganizationResponse(&org)
	}

	return list, nil
}

// Move 移动组织
func (s *OrganizationAppService) Move(ctx context.Context, id string, newParentID string) error {
	// 检查组织是否存在
	if _, err := s.orgRepo.FindByID(ctx, id); err != nil {
		return ErrOrganizationNotFound
	}

	// 检查新父组织是否存在
	if newParentID != "" {
		if _, err := s.orgRepo.FindByID(ctx, newParentID); err != nil {
			return ErrOrganizationNotFound
		}
	}

	if err := s.orgRepo.Move(ctx, id, newParentID); err != nil {
		return err
	}

	logger.Info("Organization moved", logger.String("org_id", id), logger.String("new_parent_id", newParentID))
	return nil
}

// GetPath 获取组织路径
func (s *OrganizationAppService) GetPath(ctx context.Context, id string) ([]dto.OrganizationResponse, error) {
	orgs, err := s.orgRepo.FindPath(ctx, id)
	if err != nil {
		return nil, err
	}

	list := make([]dto.OrganizationResponse, len(orgs))
	for i, org := range orgs {
		list[i] = dto.ToOrganizationResponse(&org)
	}

	return list, nil
}

// isValidOrgType 验证组织类型
func isValidOrgType(t entity.OrgType) bool {
	switch t {
	case entity.OrgTypeRoot, entity.OrgTypeProvince, entity.OrgTypeCity, entity.OrgTypeDistrict, entity.OrgTypeStreet, entity.OrgTypeCommunity, entity.OrgTypeTeam:
		return true
	default:
		return false
	}
}
