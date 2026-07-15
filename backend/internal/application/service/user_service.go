package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/Snowitty-Re/CNtunyuan/pkg/validator"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrPhoneExists       = errors.New("phone already exists")
	ErrEmailExists       = errors.New("email already exists")
	ErrInvalidRole       = errors.New("invalid role")
	ErrInvalidOrgID      = errors.New("invalid organization id")
	ErrCannotModify      = errors.New("cannot modify this user")
	ErrOldPasswordWrong  = errors.New("old password is wrong")
)

// UserAppService user application service
type UserAppService struct {
	userRepo repository.UserRepository
	taskRepo repository.TaskRepository
	mpRepo   repository.MissingPersonRepository
	orgRepo  repository.OrganizationRepository
	authz    *AuthorizationService
}

// NewUserAppService create user application service
func NewUserAppService(
	userRepo repository.UserRepository,
	taskRepo repository.TaskRepository,
	mpRepo repository.MissingPersonRepository,
	orgRepo repository.OrganizationRepository,
	authz ...*AuthorizationService,
) *UserAppService {
	var authzSvc *AuthorizationService
	if len(authz) > 0 && authz[0] != nil {
		authzSvc = authz[0]
	} else {
		authzSvc = NewAuthorizationService(orgRepo)
	}
	return &UserAppService{
		userRepo: userRepo,
		taskRepo: taskRepo,
		mpRepo:   mpRepo,
		orgRepo:  orgRepo,
		authz:    authzSvc,
	}
}

// canAssignRole enforces strict role ceiling: operator level must be strictly higher than target role.
// Only super_admin may assign super_admin.
func canAssignRole(operator *entity.User, targetRole entity.Role) bool {
	if operator == nil {
		return false
	}
	if targetRole == entity.RoleSuperAdmin {
		return operator.IsSuperAdmin()
	}
	return entity.GetRoleLevel(operator.Role) > entity.GetRoleLevel(targetRole)
}

// Create create user
func (s *UserAppService) Create(ctx context.Context, req *dto.CreateUserRequest, operators ...*entity.User) (*dto.UserResponse, error) {
	req.Email = strings.TrimSpace(req.Email)
	req.OrgID = strings.TrimSpace(req.OrgID)

	if req.OrgID == "" {
		return nil, ErrInvalidOrgID
	}
	var operator *entity.User
	if len(operators) > 0 {
		operator = operators[0]
	}
	if operator == nil {
		return nil, ErrCannotModify
	}
	if !isValidRole(req.Role) {
		return nil, ErrInvalidRole
	}
	if !canAssignRole(operator, req.Role) {
		return nil, ErrCannotModify
	}
	if !operator.IsSuperAdmin() {
		decision, err := s.authz.CanCreateUserInOrg(ctx, operator, req.OrgID)
		if err != nil {
			return nil, err
		}
		if !decision.Allowed {
			return nil, ErrCannotModify
		}
	}

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

	user, err := entity.NewUser(req.Nickname, req.Phone, req.OrgID, req.Role)
	if err != nil {
		return nil, err
	}

	user.Email = req.Email

	if err := user.SetPassword(req.Password); err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "duplicate key") || strings.Contains(errMsg, "unique constraint") ||
			strings.Contains(errMsg, "unique") || strings.Contains(errMsg, "duplicate entry") {
			if strings.Contains(errMsg, "phone") {
				return nil, ErrPhoneExists
			}
			if strings.Contains(errMsg, "email") {
				return nil, ErrEmailExists
			}
			return nil, ErrUserAlreadyExists
		}
		if strings.Contains(errMsg, "foreign key") ||
			strings.Contains(errMsg, "fk_user_org") ||
			(strings.Contains(errMsg, "org_id") && strings.Contains(errMsg, "constraint")) {
			return nil, ErrInvalidOrgID
		}
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

	allowed, err := s.canModify(ctx, operator, user)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrCannotModify
	}

	req.Email = strings.TrimSpace(req.Email)
	req.OrgID = strings.TrimSpace(req.OrgID)

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
		if !canAssignRole(operator, req.Role) {
			return nil, ErrCannotModify
		}
		user.Role = req.Role
	}
	if req.OrgID != "" && req.OrgID != user.OrgID {
		// Destination org must be within operator scope
		if !operator.IsSuperAdmin() {
			decision, scopeErr := s.authz.CanCreateUserInOrg(ctx, operator, req.OrgID)
			if scopeErr != nil {
				return nil, scopeErr
			}
			if !decision.Allowed {
				return nil, ErrCannotModify
			}
		}
		user.OrgID = req.OrgID
	}
	if req.Status != "" {
		user.Status = req.Status
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "duplicate key") || strings.Contains(errMsg, "unique constraint") ||
			strings.Contains(errMsg, "unique") || strings.Contains(errMsg, "duplicate entry") {
			if strings.Contains(errMsg, "phone") {
				return nil, ErrPhoneExists
			}
			if strings.Contains(errMsg, "email") {
				return nil, ErrEmailExists
			}
			return nil, ErrUserAlreadyExists
		}
		if strings.Contains(errMsg, "foreign key") ||
			strings.Contains(errMsg, "fk_user_org") ||
			(strings.Contains(errMsg, "org_id") && strings.Contains(errMsg, "constraint")) {
			return nil, ErrInvalidOrgID
		}
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

	allowed, err := s.canModify(ctx, operator, user)
	if err != nil {
		return err
	}
	if !allowed {
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
func (s *UserAppService) GetByID(ctx context.Context, id string, operators ...*entity.User) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	var operator *entity.User
	if len(operators) > 0 {
		operator = operators[0]
	}
	if operator != nil && !operator.IsSuperAdmin() {
		decision, scopeErr := s.authz.CanViewUser(ctx, operator, user)
		if scopeErr != nil {
			return nil, scopeErr
		}
		if !decision.Allowed {
			return nil, ErrCannotModify
		}
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
func (s *UserAppService) List(ctx context.Context, req *dto.UserListRequest, operator *entity.User) (*dto.UserListResponse, error) {
	req.Page, req.PageSize = validator.SanitizePagination(req.Page, req.PageSize)

	query := repository.NewUserQuery()
	query.Page = req.Page
	query.PageSize = req.PageSize
	query.Keyword = req.Keyword
	query.Role = entity.Role(req.Role)
	query.Status = entity.UserStatus(req.Status)
	query.OrgID = req.OrgID
	if operator != nil && !operator.IsSuperAdmin() {
		orgIDs, err := collectManageableOrgIDs(ctx, s.orgRepo, operator)
		if err != nil {
			return nil, err
		}
		query.OrgIDs = orgIDs
		if query.OrgID != "" {
			decision, err := s.authz.CanManageOrg(ctx, operator, query.OrgID)
			if err != nil {
				return nil, err
			}
			if !decision.Allowed {
				empty := dto.NewUserListResponse([]dto.UserResponse{}, 0, req.Page, req.PageSize)
				return &empty, nil
			}
		}
	}

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
	if !entity.IsValidUserStatus(status) {
		return errors.New("invalid status: must be one of active/inactive/banned")
	}

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	allowed, err := s.canModify(ctx, operator, user)
	if err != nil {
		return err
	}
	if !allowed {
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
func (s *UserAppService) UpdateRole(ctx context.Context, id string, role entity.Role, operator *entity.User) error {
	if !isValidRole(role) {
		return ErrInvalidRole
	}

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	allowed, err := s.canModify(ctx, operator, user)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrCannotModify
	}

	if !canAssignRole(operator, role) {
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
	if req.Avatar != "" {
		user.Avatar = req.Avatar
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
		return ErrOldPasswordWrong
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
	stats := &dto.UserStatsResponse{}

	// 用户上报的案件数
	totalCases, err := s.mpRepo.CountByReporter(ctx, id)
	if err != nil {
		logger.Error("Failed to count cases by reporter", logger.Err(err))
	} else {
		stats.TotalCases = totalCases
	}

	// 用户的任务统计
	taskStats, err := s.taskRepo.GetStats(ctx, id)
	if err != nil {
		logger.Error("Failed to get task stats for user", logger.Err(err))
	} else {
		stats.TotalTasks = taskStats.MyTasks
		stats.PendingTasks = taskStats.MyPending
	}

	// 用户上报的案件中已找到（found）和已团聚（reunited）的数量
	foundCases, err := s.mpRepo.CountByReporterAndStatus(ctx, id, entity.MissingStatusFound)
	if err != nil {
		logger.Error("Failed to count found cases", logger.Err(err))
	}
	reunitedCases, err := s.mpRepo.CountByReporterAndStatus(ctx, id, entity.MissingStatusReunited)
	if err != nil {
		logger.Error("Failed to count reunited cases", logger.Err(err))
	}
	stats.CompletedCases = foundCases + reunitedCases
	stats.ActiveCases = stats.TotalCases - stats.CompletedCases
	if stats.ActiveCases < 0 {
		stats.ActiveCases = 0
	}

	return stats, nil
}

// canModify check if can modify user
func (s *UserAppService) canModify(ctx context.Context, operator, target *entity.User) (bool, error) {
	decision, err := s.authz.CanModifyUser(ctx, operator, target)
	if err != nil {
		return false, err
	}
	return decision.Allowed, nil
}

// isValidRole check if role is valid
func isValidRole(role entity.Role) bool {
	switch role {
	case entity.RoleSuperAdmin, entity.RoleAdmin, entity.RoleManager, entity.RoleVolunteer:
		return true
	default:
		return false
	}
}
