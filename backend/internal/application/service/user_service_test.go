// Package service 用户应用服务测试
package service

import (
	"testing"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/repository"
	"github.com/Snowitty-Re/CNtunyuan/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserTest(t *testing.T) (*UserAppService, *testutil.TestDB) {
	tdb := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(tdb.DB)
	taskRepo := repository.NewTaskRepository(tdb.DB)
	mpRepo := repository.NewMissingPersonRepository(tdb.DB)
	service := NewUserAppService(userRepo, taskRepo, mpRepo, nil)
	return service, tdb
}

func TestUserAppService_Create(t *testing.T) {
	service, tdb := setupUserTest(t)
	defer tdb.Close()

	// 创建测试组织
	org := &entity.Organization{
		BaseEntity: entity.BaseEntity{ID: "test-org-id"},
		Name:       "测试组织",
		Code:       "TEST001",
		Type:       entity.OrgTypeCity,
	}
	testutil.MustCreate(t, tdb.DB, org)

	tests := []struct {
		name    string
		req     *dto.CreateUserRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid user",
			req: &dto.CreateUserRequest{
				Nickname: "测试用户",
				Phone:    "13800138001",
				Password: "password123",
				OrgID:    org.ID,
				Role:     entity.RoleVolunteer,
			},
			wantErr: false,
		},
		{
			name: "duplicate phone",
			req: &dto.CreateUserRequest{
				Nickname: "重复用户",
				Phone:    "13800138001", // 已存在
				Password: "password123",
				OrgID:    org.ID,
				Role:     entity.RoleVolunteer,
			},
			wantErr: true,
			errMsg:  "phone already exists",
		},
		{
			name: "invalid role",
			req: &dto.CreateUserRequest{
				Nickname: "测试用户2",
				Phone:    "13800138002",
				Password: "password123",
				OrgID:    org.ID,
				Role:     "invalid_role",
			},
			wantErr: true,
			errMsg:  "invalid role",
		},
	}

	operator := &entity.User{
		BaseEntity: entity.BaseEntity{ID: "op-super"},
		Role:       entity.RoleSuperAdmin,
		OrgID:      org.ID,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.Create(testutil.Context(), tt.req, operator)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, resp.ID)
			assert.Equal(t, tt.req.Nickname, resp.Nickname)
			assert.Equal(t, tt.req.Phone, resp.Phone)
		})
	}
}

func TestUserAppService_Create_RoleCeiling(t *testing.T) {
	service, tdb := setupUserTest(t)
	defer tdb.Close()

	org := &entity.Organization{
		BaseEntity: entity.BaseEntity{ID: "test-org-ceiling"},
		Name:       "测试组织",
		Code:       "CEIL001",
		Type:       entity.OrgTypeCity,
	}
	testutil.MustCreate(t, tdb.DB, org)

	admin := &entity.User{
		BaseEntity: entity.BaseEntity{ID: "op-admin"},
		Role:       entity.RoleAdmin,
		OrgID:      org.ID,
	}

	// admin cannot create super_admin
	_, err := service.Create(testutil.Context(), &dto.CreateUserRequest{
		Nickname: "越权超管",
		Phone:    "13900000001",
		Password: "password123",
		OrgID:    org.ID,
		Role:     entity.RoleSuperAdmin,
	}, admin)
	assert.ErrorIs(t, err, ErrCannotModify)

	// admin cannot create peer admin (strict ceiling)
	_, err = service.Create(testutil.Context(), &dto.CreateUserRequest{
		Nickname: "同级管理员",
		Phone:    "13900000002",
		Password: "password123",
		OrgID:    org.ID,
		Role:     entity.RoleAdmin,
	}, admin)
	assert.ErrorIs(t, err, ErrCannotModify)

	// admin can create manager
	resp, err := service.Create(testutil.Context(), &dto.CreateUserRequest{
		Nickname: "合法管理者",
		Phone:    "13900000003",
		Password: "password123",
		OrgID:    org.ID,
		Role:     entity.RoleManager,
	}, admin)
	require.NoError(t, err)
	assert.Equal(t, string(entity.RoleManager), resp.Role)
}

func TestUserAppService_GetByID(t *testing.T) {
	service, tdb := setupUserTest(t)
	defer tdb.Close()

	// 创建测试用户
	user, err := entity.NewUser("测试用户", "13800138000", "test-org", entity.RoleVolunteer)
	require.NoError(t, err)
	user.SetPassword("password123")
	testutil.MustCreate(t, tdb.DB, user)

	tests := []struct {
		name    string
		userID  string
		wantErr bool
	}{
		{
			name:    "existing user",
			userID:  user.ID,
			wantErr: false,
		},
		{
			name:    "non-existing user",
			userID:  "non-existing-id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.GetByID(testutil.Context(), tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.userID, resp.ID)
		})
	}
}

func TestUserAppService_UpdateStatus(t *testing.T) {
	service, tdb := setupUserTest(t)
	defer tdb.Close()

	// 创建操作者（管理员）
	operator := &entity.User{
		BaseEntity: entity.BaseEntity{ID: "operator-id"},
		Nickname:   "管理员",
		Phone:      "13800138000",
		Email:      "admin@test.com",
		WxOpenID:   "wx-openid-operator",
		Role:       entity.RoleAdmin,
		Status:     entity.UserStatusActive,
		OrgID:      "test-org",
	}
	operator.SetPassword("password123")
	testutil.MustCreate(t, tdb.DB, operator)

	// 创建目标用户
	targetUser := &entity.User{
		BaseEntity: entity.BaseEntity{ID: "target-id"},
		Nickname:   "目标用户",
		Phone:      "13800138001",
		Email:      "target@test.com",
		WxOpenID:   "wx-openid-target",
		Role:       entity.RoleVolunteer,
		Status:     entity.UserStatusActive,
		OrgID:      "test-org",
	}
	targetUser.SetPassword("password123")
	testutil.MustCreate(t, tdb.DB, targetUser)

	err := service.UpdateStatus(testutil.Context(), targetUser.ID, entity.UserStatusInactive, operator)
	require.NoError(t, err)

	// 验证状态已更新
	var updatedUser entity.User
	testutil.MustFind(t, tdb.DB, &updatedUser, "id = ?", targetUser.ID)
	assert.Equal(t, entity.UserStatusInactive, updatedUser.Status)
}

func TestUserAppService_UpdateRole(t *testing.T) {
	service, tdb := setupUserTest(t)
	defer tdb.Close()

	// 创建管理员
	admin := &entity.User{
		BaseEntity: entity.BaseEntity{ID: "admin-id"},
		Nickname:   "管理员",
		Phone:      "13800138000",
		Email:      "admin2@test.com",
		WxOpenID:   "wx-openid-admin",
		Role:       entity.RoleAdmin,
		Status:     entity.UserStatusActive,
		OrgID:      "test-org",
	}
	admin.SetPassword("password123")
	testutil.MustCreate(t, tdb.DB, admin)

	// 创建普通用户
	user := &entity.User{
		BaseEntity: entity.BaseEntity{ID: "user-id"},
		Nickname:   "普通用户",
		Phone:      "13800138001",
		Email:      "user@test.com",
		WxOpenID:   "wx-openid-user",
		Role:       entity.RoleVolunteer,
		Status:     entity.UserStatusActive,
		OrgID:      "test-org",
	}
	user.SetPassword("password123")
	testutil.MustCreate(t, tdb.DB, user)

	tests := []struct {
		name    string
		newRole entity.Role
		wantErr bool
	}{
		{
			name:    "promote to manager",
			newRole: entity.RoleManager,
			wantErr: false,
		},
		{
			name:    "invalid role",
			newRole: "invalid_role",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateRole(testutil.Context(), user.ID, tt.newRole, admin)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			var updatedUser entity.User
			testutil.MustFind(t, tdb.DB, &updatedUser, "id = ?", user.ID)
			assert.Equal(t, tt.newRole, updatedUser.Role)
		})
	}
}

func TestUserAppService_ChangePassword(t *testing.T) {
	service, tdb := setupUserTest(t)
	defer tdb.Close()

	// 创建用户
	user := &entity.User{
		BaseEntity: entity.BaseEntity{ID: "user-id"},
		Nickname:   "测试用户",
		Phone:      "13800138000",
		Email:      "changepwd@test.com",
		WxOpenID:   "wx-openid-changepwd",
		Role:       entity.RoleVolunteer,
		Status:     entity.UserStatusActive,
		OrgID:      "test-org",
	}
	user.SetPassword("oldpassword123")
	testutil.MustCreate(t, tdb.DB, user)

	tests := []struct {
		name        string
		oldPassword string
		newPassword string
		wantErr     bool
	}{
		{
			name:        "correct old password",
			oldPassword: "oldpassword123",
			newPassword: "newpassword123",
			wantErr:     false,
		},
		{
			name:        "wrong old password",
			oldPassword: "wrongpassword",
			newPassword: "newpassword123",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &dto.ChangePasswordRequest{
				OldPassword: tt.oldPassword,
				NewPassword: tt.newPassword,
			}
			err := service.ChangePassword(testutil.Context(), user.ID, req)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestUserAppService_canModify(t *testing.T) {
	service, _ := setupUserTest(t)

	tests := []struct {
		name     string
		operator *entity.User
		target   *entity.User
		want     bool
	}{
		{
			name:     "super admin can modify anyone",
			operator: &entity.User{Role: entity.RoleSuperAdmin},
			target:   &entity.User{Role: entity.RoleAdmin},
			want:     true,
		},
		{
			name:     "admin can modify volunteer",
			operator: &entity.User{Role: entity.RoleAdmin},
			target:   &entity.User{Role: entity.RoleVolunteer},
			want:     true,
		},
		{
			name:     "manager can modify volunteer",
			operator: &entity.User{Role: entity.RoleManager},
			target:   &entity.User{Role: entity.RoleVolunteer},
			want:     true,
		},
		{
			name:     "manager cannot modify admin",
			operator: &entity.User{Role: entity.RoleManager},
			target:   &entity.User{Role: entity.RoleAdmin},
			want:     false,
		},
		{
			name:     "volunteer can modify self",
			operator: &entity.User{BaseEntity: entity.BaseEntity{ID: "user-1"}, Role: entity.RoleVolunteer},
			target:   &entity.User{BaseEntity: entity.BaseEntity{ID: "user-1"}, Role: entity.RoleVolunteer},
			want:     true,
		},
		{
			name:     "volunteer cannot modify others",
			operator: &entity.User{BaseEntity: entity.BaseEntity{ID: "user-1"}, Role: entity.RoleVolunteer},
			target:   &entity.User{BaseEntity: entity.BaseEntity{ID: "user-2"}, Role: entity.RoleVolunteer},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.canModify(testutil.Context(), tt.operator, tt.target)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
