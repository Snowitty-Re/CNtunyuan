package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Snowitty-Re/CNtunyuan/internal/model"
	"github.com/Snowitty-Re/CNtunyuan/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	userRepo *repository.UserRepository
	orgRepo  *repository.OrganizationRepository
}

// NewUserService 创建用户服务
func NewUserService(userRepo *repository.UserRepository, orgRepo *repository.OrganizationRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
		orgRepo:  orgRepo,
	}
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	OpenID   string     `json:"open_id"`
	UnionID  string     `json:"union_id"`
	Nickname string     `json:"nickname"`
	Avatar   string     `json:"avatar"`
	Phone    string     `json:"phone"`
	RealName string     `json:"real_name"`
	IDCard   string     `json:"id_card"`
	Role     string     `json:"role"`
	OrgID    *uuid.UUID `json:"org_id"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Nickname string     `json:"nickname"`
	Avatar   string     `json:"avatar"`
	Phone    string     `json:"phone"`
	RealName string     `json:"real_name"`
	IDCard   string     `json:"id_card"`
	Role     string     `json:"role"`
	Status   string     `json:"status"`
	OrgID    *uuid.UUID `json:"org_id"`
}

// GetOrCreateByWeChat 根据微信信息获取或创建用户
func (s *UserService) GetOrCreateByWeChat(ctx context.Context, openID, unionID, nickname, avatar string) (*model.User, bool, error) {
	// 先尝试通过OpenID查找
	user, err := s.userRepo.GetByOpenID(ctx, openID)
	if err == nil {
		// 更新用户信息
		if nickname != "" && user.Nickname != nickname {
			user.Nickname = nickname
		}
		if avatar != "" && user.Avatar != avatar {
			user.Avatar = avatar
		}
		if err := s.userRepo.Update(ctx, user); err != nil {
			return nil, false, err
		}
		return user, false, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}

	// 尝试通过UnionID查找
	if unionID != "" {
		user, err = s.userRepo.GetByUnionID(ctx, unionID)
		if err == nil {
			// 更新OpenID
			user.OpenID = openID
			if err := s.userRepo.Update(ctx, user); err != nil {
				return nil, false, err
			}
			return user, false, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, err
		}
	}

	// 创建新用户
	newUser := &model.User{
		OpenID:   openID,
		UnionID:  unionID,
		Nickname: nickname,
		Avatar:   avatar,
		Role:     model.RoleVolunteer,
		Status:   model.UserStatusActive,
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, false, err
	}

	return newUser, true, nil
}

// GetByID 根据ID获取用户
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// Update 更新用户
func (s *UserService) Update(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.RealName != "" {
		user.RealName = req.RealName
	}
	if req.IDCard != "" {
		user.IDCard = req.IDCard
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.Status != "" {
		user.Status = req.Status
	}
	if req.OrgID != nil {
		user.OrgID = req.OrgID
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	// 更新组织志愿者数量
	if req.OrgID != nil {
		s.orgRepo.UpdateVolunteerCount(ctx, *req.OrgID)
	}

	return user, nil
}

// Delete 删除用户
func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return err
	}

	// 更新组织志愿者数量
	if user.OrgID != nil {
		s.orgRepo.UpdateVolunteerCount(ctx, *user.OrgID)
	}

	return nil
}

// List 获取用户列表
func (s *UserService) List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*model.User, int64, error) {
	return s.userRepo.List(ctx, page, pageSize, filters)
}

// UpdateLastLogin 更新最后登录时间
func (s *UserService) UpdateLastLogin(ctx context.Context, id uuid.UUID, ip string) error {
	return s.userRepo.UpdateLastLogin(ctx, id, ip)
}

// GetStatistics 获取用户统计
func (s *UserService) GetStatistics(ctx context.Context) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 总用户数
	total, err := s.userRepo.Count(ctx, nil)
	if err != nil {
		return nil, err
	}
	result["total"] = total

	// 各角色数量
	roles := []string{model.RoleSuperAdmin, model.RoleAdmin, model.RoleManager, model.RoleVolunteer}
	for _, role := range roles {
		count, err := s.userRepo.Count(ctx, map[string]interface{}{"role": role})
		if err != nil {
			return nil, err
		}
		result[role] = count
	}

	// 各状态数量
	statuses := []string{model.UserStatusActive, model.UserStatusInactive, model.UserStatusBanned}
	for _, status := range statuses {
		count, err := s.userRepo.Count(ctx, map[string]interface{}{"status": status})
		if err != nil {
			return nil, err
		}
		result[status] = count
	}

	return result, nil
}

// AssignToOrg 分配用户到组织
func (s *UserService) AssignToOrg(ctx context.Context, userID, orgID uuid.UUID) error {
	// 检查组织是否存在
	_, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return fmt.Errorf("组织不存在: %w", err)
	}

	// 更新用户组织
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	oldOrgID := user.OrgID
	user.OrgID = &orgID

	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// 更新组织志愿者数量
	s.orgRepo.UpdateVolunteerCount(ctx, orgID)
	if oldOrgID != nil {
		s.orgRepo.UpdateVolunteerCount(ctx, *oldOrgID)
	}

	return nil
}
