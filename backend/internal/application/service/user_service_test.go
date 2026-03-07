package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/application/service/mocks"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	pkgErrors "github.com/Snowitty-Re/CNtunyuan/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserAppService_Create(t *testing.T) {
	mockRepo := new(mocks.UserRepositoryMock)
	service := NewUserAppService(mockRepo)
	ctx := context.Background()

	tests := []struct {
		name    string
		req     *dto.CreateUserRequest
		mock    func()
		wantErr bool
		errCode pkgErrors.ErrorCode
	}{
		{
			name: "success - create user",
			req: &dto.CreateUserRequest{
				Nickname: "Test User",
				Phone:    "13800138000",
				Password: "password123",
				Role:     "volunteer",
				OrgID:    "org-001",
			},
			mock: func() {
				mockRepo.On("ExistsByPhone", ctx, "13800138000").Return(false, nil).Once()
				mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.User")).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "fail - phone already exists",
			req: &dto.CreateUserRequest{
				Nickname: "Test User",
				Phone:    "13800138000",
				Password: "password123",
				Role:     "volunteer",
				OrgID:    "org-001",
			},
			mock: func() {
				mockRepo.On("ExistsByPhone", ctx, "13800138000").Return(true, nil).Once()
			},
			wantErr: true,
			errCode: pkgErrors.CodeUserExists,
		},
		{
			name: "fail - invalid phone",
			req: &dto.CreateUserRequest{
				Nickname: "Test User",
				Phone:    "invalid",
				Password: "password123",
				Role:     "volunteer",
				OrgID:    "org-001",
			},
			mock:    func() {},
			wantErr: true,
			errCode: pkgErrors.CodeInvalidParam,
		},
		{
			name: "fail - password too short",
			req: &dto.CreateUserRequest{
				Nickname: "Test User",
				Phone:    "13800138001",
				Password: "123",
				Role:     "volunteer",
				OrgID:    "org-001",
			},
			mock:    func() {},
			wantErr: true,
			errCode: pkgErrors.CodeInvalidParam,
		},
		{
			name: "fail - empty nickname",
			req: &dto.CreateUserRequest{
				Nickname: "",
				Phone:    "13800138002",
				Password: "password123",
				Role:     "volunteer",
				OrgID:    "org-001",
			},
			mock:    func() {},
			wantErr: true,
			errCode: pkgErrors.CodeInvalidParam,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			resp, err := service.Create(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != 0 {
					assert.Equal(t, tt.errCode, pkgErrors.GetCode(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.req.Nickname, resp.Nickname)
				assert.Equal(t, tt.req.Phone, resp.Phone)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserAppService_GetByID(t *testing.T) {
	mockRepo := new(mocks.UserRepositoryMock)
	service := NewUserAppService(mockRepo)
	ctx := context.Background()

	tests := []struct {
		name    string
		userID  string
		mock    func()
		wantErr bool
		errCode pkgErrors.ErrorCode
	}{
		{
			name:   "success - get user by id",
			userID: "user-001",
			mock: func() {
				user := &entity.User{
					ID:       "user-001",
					Nickname: "Test User",
					Phone:    "13800138000",
					Role:     "volunteer",
					OrgID:    "org-001",
					Status:   entity.UserStatusActive,
				}
				mockRepo.On("FindByID", ctx, "user-001").Return(user, nil).Once()
			},
			wantErr: false,
		},
		{
			name:   "fail - user not found",
			userID: "user-notfound",
			mock: func() {
				mockRepo.On("FindByID", ctx, "user-notfound").Return(nil, errors.New("user not found")).Once()
			},
			wantErr: true,
			errCode: pkgErrors.CodeUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			resp, err := service.GetByID(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != 0 {
					assert.Equal(t, tt.errCode, pkgErrors.GetCode(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.userID, resp.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserAppService_Update(t *testing.T) {
	mockRepo := new(mocks.UserRepositoryMock)
	service := NewUserAppService(mockRepo)
	ctx := context.Background()

	tests := []struct {
		name    string
		userID  string
		req     *dto.UpdateUserRequest
		mock    func()
		wantErr bool
		errCode pkgErrors.ErrorCode
	}{
		{
			name:   "success - update user",
			userID: "user-001",
			req: &dto.UpdateUserRequest{
				Nickname: "Updated Name",
				Email:    "updated@example.com",
			},
			mock: func() {
				user := &entity.User{
					ID:       "user-001",
					Nickname: "Test User",
					Phone:    "13800138000",
				}
				mockRepo.On("FindByID", ctx, "user-001").Return(user, nil).Once()
				mockRepo.On("Update", ctx, mock.AnythingOfType("*entity.User")).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:   "fail - user not found",
			userID: "user-notfound",
			req: &dto.UpdateUserRequest{
				Nickname: "Updated Name",
			},
			mock: func() {
				mockRepo.On("FindByID", ctx, "user-notfound").Return(nil, errors.New("user not found")).Once()
			},
			wantErr: true,
			errCode: pkgErrors.CodeUserNotFound,
		},
		{
			name:   "fail - invalid email",
			userID: "user-001",
			req: &dto.UpdateUserRequest{
				Email: "invalid-email",
			},
			mock: func() {
				user := &entity.User{
					ID:    "user-001",
					Phone: "13800138000",
				}
				mockRepo.On("FindByID", ctx, "user-001").Return(user, nil).Once()
			},
			wantErr: true,
			errCode: pkgErrors.CodeInvalidParam,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			resp, err := service.Update(ctx, tt.userID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != 0 {
					assert.Equal(t, tt.errCode, pkgErrors.GetCode(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserAppService_Delete(t *testing.T) {
	mockRepo := new(mocks.UserRepositoryMock)
	service := NewUserAppService(mockRepo)
	ctx := context.Background()

	tests := []struct {
		name    string
		userID  string
		mock    func()
		wantErr bool
		errCode pkgErrors.ErrorCode
	}{
		{
			name:   "success - delete user",
			userID: "user-001",
			mock: func() {
				user := &entity.User{
					ID:     "user-001",
					Role:   "volunteer",
					Status: entity.UserStatusActive,
				}
				mockRepo.On("FindByID", ctx, "user-001").Return(user, nil).Once()
				mockRepo.On("Delete", ctx, "user-001").Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:   "fail - user not found",
			userID: "user-notfound",
			mock: func() {
				mockRepo.On("FindByID", ctx, "user-notfound").Return(nil, errors.New("user not found")).Once()
			},
			wantErr: true,
			errCode: pkgErrors.CodeUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := service.Delete(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != 0 {
					assert.Equal(t, tt.errCode, pkgErrors.GetCode(err))
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserAppService_List(t *testing.T) {
	mockRepo := new(mocks.UserRepositoryMock)
	service := NewUserAppService(mockRepo)
	ctx := context.Background()

	tests := []struct {
		name    string
		req     *dto.UserListRequest
		mock    func()
		wantErr bool
	}{
		{
			name: "success - list users",
			req: &dto.UserListRequest{
				Page:     1,
				PageSize: 10,
				Keyword:  "test",
			},
			mock: func() {
				users := []entity.User{
					{ID: "user-001", Nickname: "Test User 1", Phone: "13800138000"},
					{ID: "user-002", Nickname: "Test User 2", Phone: "13800138001"},
				}
				result := repository.NewPageResult(users, 2, 1, 10)
				mockRepo.On("List", ctx, mock.AnythingOfType("*repository.UserQuery")).Return(result, nil).Once()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			resp, err := service.List(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.GreaterOrEqual(t, len(resp.List), 0)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserAppService_UpdateStatus(t *testing.T) {
	mockRepo := new(mocks.UserRepositoryMock)
	service := NewUserAppService(mockRepo)
	ctx := context.Background()

	tests := []struct {
		name    string
		userID  string
		status  string
		mock    func()
		wantErr bool
	}{
		{
			name:   "success - activate user",
			userID: "user-001",
			status: "active",
			mock: func() {
				user := &entity.User{
					ID:     "user-001",
					Status: entity.UserStatusInactive,
				}
				mockRepo.On("FindByID", ctx, "user-001").Return(user, nil).Once()
				mockRepo.On("UpdateStatus", ctx, "user-001", entity.UserStatusActive).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:   "fail - invalid status",
			userID: "user-001",
			status: "invalid",
			mock: func() {
				user := &entity.User{
					ID:     "user-001",
					Status: entity.UserStatusActive,
				}
				mockRepo.On("FindByID", ctx, "user-001").Return(user, nil).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := service.UpdateStatus(ctx, tt.userID, tt.status)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserAppService_ChangePassword(t *testing.T) {
	mockRepo := new(mocks.UserRepositoryMock)
	service := NewUserAppService(mockRepo)
	ctx := context.Background()

	tests := []struct {
		name    string
		userID  string
		req     *dto.ChangePasswordRequest
		mock    func()
		wantErr bool
	}{
		{
			name:   "success - change password",
			userID: "user-001",
			req: &dto.ChangePasswordRequest{
				OldPassword: "oldpassword123",
				NewPassword: "newpassword123",
			},
			mock: func() {
				user := &entity.User{
					ID:       "user-001",
					Password: "$2a$10$...",
				}
				mockRepo.On("FindByID", ctx, "user-001").Return(user, nil).Once()
				mockRepo.On("Update", ctx, mock.AnythingOfType("*entity.User")).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:   "fail - user not found",
			userID: "user-notfound",
			req: &dto.ChangePasswordRequest{
				OldPassword: "oldpassword123",
				NewPassword: "newpassword123",
			},
			mock: func() {
				mockRepo.On("FindByID", ctx, "user-notfound").Return(nil, errors.New("user not found")).Once()
			},
			wantErr: true,
		},
		{
			name:   "fail - password too short",
			userID: "user-001",
			req: &dto.ChangePasswordRequest{
				OldPassword: "oldpassword123",
				NewPassword: "123",
			},
			mock: func() {
				user := &entity.User{
					ID:       "user-001",
					Password: "$2a$10$...",
				}
				mockRepo.On("FindByID", ctx, "user-001").Return(user, nil).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := service.ChangePassword(ctx, tt.userID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
