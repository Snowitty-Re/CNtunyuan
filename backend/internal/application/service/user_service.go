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
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrPhoneExists       = errors.New("phone already exists")
	ErrEmailExists       = errors.New("email already exists")
	ErrInvalidRole       = errors.New("invalid role")
	ErrCannotModify      = errors.New("cannot modify this user")
)

// UserAppService user application service
type UserAppService struct {
	userRepo repository.UserRepository
}

// NewUserAppService create user application service
func NewUserAppService(userRepo repository.UserRepository) *UserAppService {
	return &UserAppService{userRepo: userRepo}
}

// Create create user
func (s *UserAppService) Create(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	exists, err := s.userRepo.ExistsPhone(ctx, req.Phone)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrPhoneExists
	}

	if req.Email != "" {
		exists, err = s.userRepo.ExistsEmail(ctx, req.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrEmailExists
		}
	}

	if !isValidRole(req.Role) {
		return nil, ErrInvalidRole
	}

	user, err := entity.NewUser(req.Nickname, req.Phone, req.OrgID, req.Role)
	if err != nil {
		return nil, err
	}

	user.Email = req.Email

	if err := user.SetPassword(req.Password); err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		logger.Error("Failed to create user", logger.Err(err))
		return nil, err
	}

	logger.Info("User created", logger.String("user_id", user.ID), logger.String("phone", user.Phone))

	resp := dto.ToUserResponse(user)
	return &resp, nil
}

// Update update user
func (s *UserAppService) Update(ctx context.Context, id string, req *dto.UpdateUserRequest, operator *entity.User) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if !s.canModify(operator, user) {
		return nil, ErrCannotModify
	}

	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Role != "" {
		if !isValidRole(req.Role) {
			return nil, ErrInvalidRole
		}
		if !operator.HasPermission(req.Role) {
			return nil, ErrCannotModify
		}
		user.Role = req.Role
	}
	if req.OrgID != "" {
		user.OrgID = req.OrgID
	}
	if req.Status != "" {
		user.Status = req.Status
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		logger.Error("Failed to update user", logger.Err(err))
		return nil, err
	}

	logger.Info("User updated", logger.String("user_id", user.ID))

	resp := dto.ToUserResponse(user)
	return &resp, nil
}

// Delete delete user
func (s *UserAppService) Delete(ctx context.Context, id string, operator *entity.User) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	if !s.canModify(operator, user) {
		return ErrCannotModify
	}

	if err := s.userRepo.SoftDelete(ctx, id); err != nil {
		logger.Error("Failed to delete user", logger.Err(err))
		return err
	}

	logger.Info("User deleted", logger.String("user_id", id))
	return nil
}

// GetByID get user by ID
func (s *UserAppService) GetByID(ctx context.Context, id string) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	resp := dto.ToUserResponse(user)
	return &resp, nil
}

// GetByPhone get user by phone
func (s *UserAppService) GetByPhone(ctx context.Context, phone string) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByPhone(ctx, phone)
	if err != nil {
		return nil, ErrUserNotFound
	}

	resp := dto.ToUserResponse(user)
	return &resp, nil
}

// List user list
func (s *UserAppService) List(ctx context.Context, req *dto.UserListRequest) (*dto.UserListResponse, error) {
	query := repository.NewUserQuery()
	query.Page = req.Page
	query.PageSize = req.PageSize
	query.Keyword = req.Keyword
	query.Role = req.Role
	query.Status = entity.UserStatus(req.Status)
	query.OrgID = req.OrgID

	result, err := s.userRepo.List(ctx, query)
	if err != nil {
		return nil, err
	}

	list := make([]dto.UserResponse, len(result.List))
	for i, user := range result.List {
		list[i] = dto.ToUserResponse(&user)
	}

	resp := dto.NewUserListResponse(list, result.Total, result.Page, result.PageSize)
	return &resp, nil
}

// UpdateStatus update user status
func (s *UserAppService) UpdateStatus(ctx context.Context, id string, status entity.UserStatus, operator *entity.User) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	if !s.canModify(operator, user) {
		return ErrCannotModify
	}

	if err := s.userRepo.UpdateStatus(ctx, id, status); err != nil {
		return err
	}

	logger.Info("User status updated",
		logger.String("user_id", id),
		logger.String("status", string(status)),
	)
	return nil
}

// UpdateRole update user role
func (s *UserAppService) UpdateRole(ctx context.Context, id string, role string, operator *entity.User) error {
	if !isValidRole(role) {
		return ErrInvalidRole
	}

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	if !s.canModify(operator, user) {
		return ErrCannotModify
	}

	if !operator.HasPermission(role) {
		return ErrCannotModify
	}

	if err := s.userRepo.UpdateRole(ctx, id, role); err != nil {
		return err
	}

	logger.Info("User role updated",
		logger.String("user_id", id),
		logger.String("role", string(role)),
	)
	return nil
}

// UpdateProfile update profile
func (s *UserAppService) UpdateProfile(ctx context.Context, id string, req *dto.UpdateProfileRequest) (*dto.UserProfileResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.RealName != "" {
		user.RealName = req.RealName
	}
	if req.IDCard != "" {
		user.IDCard = req.IDCard
	}
	if req.Gender != "" {
		user.Gender = req.Gender
	}
	if req.Address != "" {
		user.Address = req.Address
	}
	if req.Emergency != "" {
		user.Emergency = req.Emergency
	}
	if req.EmergencyTel != "" {
		user.EmergencyTel = req.EmergencyTel
	}
	if req.Introduction != "" {
		user.Introduction = req.Introduction
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	resp := dto.ToUserProfileResponse(user)
	return &resp, nil
}

// ChangePassword change password
func (s *UserAppService) ChangePassword(ctx context.Context, id string, req *dto.ChangePasswordRequest) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	if !user.CheckPassword(req.OldPassword) {
		return errors.New("old password is wrong")
	}

	if err := user.SetPassword(req.NewPassword); err != nil {
		return err
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	return nil
}

// GetProfile get profile
func (s *UserAppService) GetProfile(ctx context.Context, id string) (*dto.UserProfileResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	resp := dto.ToUserProfileResponse(user)
	return &resp, nil
}

// GetStats get stats
func (s *UserAppService) GetStats(ctx context.Context, id string) (*dto.UserStatsResponse, error) {
	return &dto.UserStatsResponse{}, nil
}

// canModify check if can modify user
func (s *UserAppService) canModify(operator, target *entity.User) bool {
	if operator.IsSuperAdmin() {
		return true
	}
	if target.IsSuperAdmin() {
		return false
	}
	if operator.IsAdmin() {
		return true
	}
	if operator.Role == string(entity.RoleManager) {
		return target.Role == string(entity.RoleVolunteer)
	}
	return operator.ID == target.ID
}

// isValidRole check if role is valid
func isValidRole(role string) bool {
	switch role {
	case string(entity.RoleSuperAdmin), string(entity.RoleAdmin), string(entity.RoleManager), string(entity.RoleVolunteer):
		return true
	default:
		return false
	}
}
