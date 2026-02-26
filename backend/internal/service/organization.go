package service

import (
	"context"
	"fmt"

	"github.com/Snowitty-Re/CNtunyuan/internal/model"
	"github.com/Snowitty-Re/CNtunyuan/internal/repository"
	"github.com/google/uuid"
)

// OrganizationService 组织服务
type OrganizationService struct {
	orgRepo *repository.OrganizationRepository
}

// NewOrganizationService 创建组织服务
func NewOrganizationService(orgRepo *repository.OrganizationRepository) *OrganizationService {
	return &OrganizationService{orgRepo: orgRepo}
}

// CreateOrgRequest 创建组织请求
type CreateOrgRequest struct {
	Name        string     `json:"name" binding:"required"`
	Code        string     `json:"code" binding:"required"`
	Type        string     `json:"type" binding:"required"`
	ParentID    *uuid.UUID `json:"parent_id"`
	LeaderID    *uuid.UUID `json:"leader_id"`
	Province    string     `json:"province"`
	City        string     `json:"city"`
	District    string     `json:"district"`
	Street      string     `json:"street"`
	Address     string     `json:"address"`
	Contact     string     `json:"contact"`
	Phone       string     `json:"phone"`
	Email       string     `json:"email"`
	Description string     `json:"description"`
}

// UpdateOrgRequest 更新组织请求
type UpdateOrgRequest struct {
	Name        string     `json:"name"`
	LeaderID    *uuid.UUID `json:"leader_id"`
	Province    string     `json:"province"`
	City        string     `json:"city"`
	District    string     `json:"district"`
	Street      string     `json:"street"`
	Address     string     `json:"address"`
	Contact     string     `json:"contact"`
	Phone       string     `json:"phone"`
	Email       string     `json:"email"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Sort        int        `json:"sort"`
}

// Create 创建组织
func (s *OrganizationService) Create(ctx context.Context, req *CreateOrgRequest) (*model.Organization, error) {
	// 检查编码是否已存在
	if _, err := s.orgRepo.GetByCode(ctx, req.Code); err == nil {
		return nil, fmt.Errorf("组织编码已存在")
	}

	// 确定层级
	level := 1
	if req.ParentID != nil {
		parent, err := s.orgRepo.GetByID(ctx, *req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("父组织不存在")
		}
		level = parent.Level + 1
	}

	org := &model.Organization{
		Name:     req.Name,
		Code:     req.Code,
		Type:     req.Type,
		Level:    level,
		ParentID: req.ParentID,
		LeaderID: req.LeaderID,
		Province: req.Province,
		City:     req.City,
		District: req.District,
		Street:   req.Street,
		Address:  req.Address,
		Contact:  req.Contact,
		Phone:    req.Phone,
		Email:    req.Email,
		Description: req.Description,
		Status:   model.UserStatusActive,
	}

	if err := s.orgRepo.Create(ctx, org); err != nil {
		return nil, err
	}

	return org, nil
}

// GetByID 根据ID获取组织
func (s *OrganizationService) GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	return s.orgRepo.GetByID(ctx, id)
}

// Update 更新组织
func (s *OrganizationService) Update(ctx context.Context, id uuid.UUID, req *UpdateOrgRequest) (*model.Organization, error) {
	org, err := s.orgRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		org.Name = req.Name
	}
	if req.LeaderID != nil {
		org.LeaderID = req.LeaderID
	}
	if req.Province != "" {
		org.Province = req.Province
	}
	if req.City != "" {
		org.City = req.City
	}
	if req.District != "" {
		org.District = req.District
	}
	if req.Street != "" {
		org.Street = req.Street
	}
	if req.Address != "" {
		org.Address = req.Address
	}
	if req.Contact != "" {
		org.Contact = req.Contact
	}
	if req.Phone != "" {
		org.Phone = req.Phone
	}
	if req.Email != "" {
		org.Email = req.Email
	}
	if req.Description != "" {
		org.Description = req.Description
	}
	if req.Status != "" {
		org.Status = req.Status
	}
	if req.Sort != 0 {
		org.Sort = req.Sort
	}

	if err := s.orgRepo.Update(ctx, org); err != nil {
		return nil, err
	}

	return org, nil
}

// Delete 删除组织
func (s *OrganizationService) Delete(ctx context.Context, id uuid.UUID) error {
	// 检查是否有子组织
	children, err := s.orgRepo.GetChildren(ctx, id)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		return fmt.Errorf("该组织下存在子组织，无法删除")
	}

	return s.orgRepo.Delete(ctx, id)
}

// GetTree 获取组织架构树
func (s *OrganizationService) GetTree(ctx context.Context, parentID *uuid.UUID) ([]*model.Organization, error) {
	return s.orgRepo.GetTree(ctx, parentID)
}

// List 获取组织列表
func (s *OrganizationService) List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*model.Organization, int64, error) {
	return s.orgRepo.List(ctx, page, pageSize, filters)
}

// GetOrgPath 获取组织路径
func (s *OrganizationService) GetOrgPath(ctx context.Context, id uuid.UUID) (string, error) {
	org, err := s.orgRepo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}
	return org.GetFullPath(model.DB)
}
