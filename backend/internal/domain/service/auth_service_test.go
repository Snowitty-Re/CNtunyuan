package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/service/mocks"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	pkgErrors "github.com/Snowitty-Re/CNtunyuan/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTokenService 令牌服务Mock
type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) Generate(userID, orgID string, role string) (string, error) {
	args := m.Called(userID, orgID, role)
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) Validate(token string) (*TokenClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TokenClaims), args.Error(1)
}

func (m *MockTokenService) Refresh(token string) (string, error) {
	args := m.Called(token)
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) Revoke(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

// TokenClaims 令牌声明
type TokenClaims struct {
	UserID string
	OrgID  string
	Role   string
}

func TestAuthService_Login(t *testing.T) {
	userRepo := new(mocks.UserRepositoryMock)
	tokenService := new(MockTokenService)
	authService := NewAuthService(userRepo, tokenService, nil, nil)
	ctx := context.Background()

	tests := []struct {
		name     string
		phone    string
		password string
		mock     func()
		wantErr  bool
		errCode  pkgErrors.ErrorCode
	}{
		{
			name:     "success - login with valid credentials",
			phone:    "13800138000",
			password: "password123",
			mock: func() {
				user := &entity.User{
					ID:       "user-001",
					Phone:    "13800138000",
					Password: "$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrzM8mU7sJ7oJ0.9gHzb8N8b3r8G8a", // bcrypt hash
					Role:     "volunteer",
					OrgID:    "org-001",
					Status:   entity.UserStatusActive,
				}
				user.SetPassword("password123")
				userRepo.On("FindByPhone", ctx, "13800138000").Return(user, nil).Once()
				userRepo.On("UpdateLastLogin", ctx, "user-001", mock.Anything).Return(nil).Once()
				tokenService.On("Generate", "user-001", "org-001", "volunteer").Return("test-token", nil).Once()
			},
			wantErr: false,
		},
		{
			name:     "fail - user not found",
			phone:    "13800138000",
			password: "password123",
			mock: func() {
				userRepo.On("FindByPhone", ctx, "13800138000").Return(nil, errors.New("user not found")).Once()
			},
			wantErr: true,
			errCode: pkgErrors.CodeUserNotFound,
		},
		{
			name:     "fail - invalid password",
			phone:    "13800138000",
			password: "wrongpassword",
			mock: func() {
				user := &entity.User{
					ID:       "user-001",
					Phone:    "13800138000",
					Password: "$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrzM8mU7sJ7oJ0.9gHzb8N8b3r8G8a",
					Status:   entity.UserStatusActive,
				}
				user.SetPassword("password123")
				userRepo.On("FindByPhone", ctx, "13800138000").Return(user, nil).Once()
			},
			wantErr: true,
			errCode: pkgErrors.CodeInvalidPassword,
		},
		{
			name:     "fail - account disabled",
			phone:    "13800138000",
			password: "password123",
			mock: func() {
				user := &entity.User{
					ID:       "user-001",
					Phone:    "13800138000",
					Password: "$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrzM8mU7sJ7oJ0.9gHzb8N8b3r8G8a",
					Status:   entity.UserStatusInactive,
				}
				user.SetPassword("password123")
				userRepo.On("FindByPhone", ctx, "13800138000").Return(user, nil).Once()
			},
			wantErr: true,
			errCode: pkgErrors.CodeAccountDisabled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			token, user, err := authService.Login(ctx, tt.phone, tt.password, "127.0.0.1")

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != 0 {
					assert.Equal(t, tt.errCode, pkgErrors.GetCode(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				assert.NotNil(t, user)
			}

			userRepo.AssertExpectations(t)
			tokenService.AssertExpectations(t)
		})
	}
}

func TestAuthService_Logout(t *testing.T) {
	userRepo := new(mocks.UserRepositoryMock)
	tokenService := new(MockTokenService)
	authService := NewAuthService(userRepo, tokenService, nil, nil)

	tests := []struct {
		name  string
		token string
		mock  func()
		wantErr bool
	}{
		{
			name:  "success - logout",
			token: "valid-token",
			mock: func() {
				tokenService.On("Revoke", "valid-token").Return(nil).Once()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := authService.Logout(tt.token)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tokenService.AssertExpectations(t)
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	userRepo := new(mocks.UserRepositoryMock)
	tokenService := new(MockTokenService)
	authService := NewAuthService(userRepo, tokenService, nil, nil)

	tests := []struct {
		name  string
		token string
		mock  func()
		wantErr bool
	}{
		{
			name:  "success - refresh token",
			token: "old-token",
			mock: func() {
				tokenService.On("Refresh", "old-token").Return("new-token", nil).Once()
			},
			wantErr: false,
		},
		{
			name:  "fail - invalid token",
			token: "invalid-token",
			mock: func() {
				tokenService.On("Refresh", "invalid-token").Return("", errors.New("invalid token")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			newToken, err := authService.RefreshToken(tt.token)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, newToken)
			}

			tokenService.AssertExpectations(t)
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	userRepo := new(mocks.UserRepositoryMock)
	tokenService := new(MockTokenService)
	authService := NewAuthService(userRepo, tokenService, nil, nil)

	tests := []struct {
		name  string
		token string
		mock  func()
		wantErr bool
	}{
		{
			name:  "success - validate token",
			token: "valid-token",
			mock: func() {
				claims := &TokenClaims{
					UserID: "user-001",
					OrgID:  "org-001",
					Role:   "volunteer",
				}
				tokenService.On("Validate", "valid-token").Return(claims, nil).Once()
			},
			wantErr: false,
		},
		{
			name:  "fail - invalid token",
			token: "invalid-token",
			mock: func() {
				tokenService.On("Validate", "invalid-token").Return(nil, errors.New("invalid token")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			claims, err := authService.ValidateToken(tt.token)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
			}

			tokenService.AssertExpectations(t)
		})
	}
}

func TestAuthService_GetCurrentUser(t *testing.T) {
	userRepo := new(mocks.UserRepositoryMock)
	tokenService := new(MockTokenService)
	authService := NewAuthService(userRepo, tokenService, nil, nil)
	ctx := context.Background()

	tests := []struct {
		name   string
		userID string
		mock   func()
		wantErr bool
	}{
		{
			name:   "success - get current user",
			userID: "user-001",
			mock: func() {
				user := &entity.User{
					ID:        "user-001",
					Nickname:  "Test User",
					Phone:     "13800138000",
					Role:      "volunteer",
					OrgID:     "org-001",
					Status:    entity.UserStatusActive,
					CreatedAt: time.Now(),
				}
				userRepo.On("FindByID", ctx, "user-001").Return(user, nil).Once()
			},
			wantErr: false,
		},
		{
			name:   "fail - user not found",
			userID: "user-notfound",
			mock: func() {
				userRepo.On("FindByID", ctx, "user-notfound").Return(nil, errors.New("user not found")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			user, err := authService.GetCurrentUser(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.userID, user.ID)
			}

			userRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_UpdatePassword(t *testing.T) {
	userRepo := new(mocks.UserRepositoryMock)
	tokenService := new(MockTokenService)
	authService := NewAuthService(userRepo, tokenService, nil, nil)
	ctx := context.Background()

	tests := []struct {
		name        string
		userID      string
		oldPassword string
		newPassword string
		mock        func()
		wantErr     bool
	}{
		{
			name:        "success - update password",
			userID:      "user-001",
			oldPassword: "oldpassword123",
			newPassword: "newpassword123",
			mock: func() {
				user := &entity.User{
					ID:       "user-001",
					Password: "$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrzM8mU7sJ7oJ0.9gHzb8N8b3r8G8a",
				}
				user.SetPassword("oldpassword123")
				userRepo.On("FindByID", ctx, "user-001").Return(user, nil).Once()
				userRepo.On("Update", ctx, mock.AnythingOfType("*entity.User")).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:        "fail - user not found",
			userID:      "user-notfound",
			oldPassword: "oldpassword123",
			newPassword: "newpassword123",
			mock: func() {
				userRepo.On("FindByID", ctx, "user-notfound").Return(nil, errors.New("user not found")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := authService.UpdatePassword(ctx, tt.userID, tt.oldPassword, tt.newPassword)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			userRepo.AssertExpectations(t)
		})
	}
}
