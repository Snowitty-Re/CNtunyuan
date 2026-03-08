// Package entity 用户实体测试
package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name      string
		nickname  string
		phone     string
		orgID     string
		role      Role
		wantErr   bool
		errMsg    string
	}{
		{
			name:     "valid user",
			nickname: "测试用户",
			phone:    "13800138000",
			orgID:    "test-org-id",
			role:     RoleVolunteer,
			wantErr:  false,
		},
		{
			name:     "empty nickname",
			nickname: "",
			phone:    "13800138000",
			orgID:    "test-org-id",
			role:     RoleVolunteer,
			wantErr:  true,
			errMsg:   "昵称不能为空",
		},
		{
			name:     "invalid phone",
			nickname: "测试用户",
			phone:    "12345678901",
			orgID:    "test-org-id",
			role:     RoleVolunteer,
			wantErr:  true,
			errMsg:   "手机号格式不正确",
		},
		{
			name:     "invalid phone too short",
			nickname: "测试用户",
			phone:    "138001",
			orgID:    "test-org-id",
			role:     RoleVolunteer,
			wantErr:  true,
			errMsg:   "手机号格式不正确",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.nickname, tt.phone, tt.orgID, tt.role)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, user.ID)
			assert.Equal(t, tt.nickname, user.Nickname)
			assert.Equal(t, tt.phone, user.Phone)
			assert.Equal(t, tt.orgID, user.OrgID)
			assert.Equal(t, tt.role, user.Role)
			assert.Equal(t, UserStatusActive, user.Status)
		})
	}
}

func TestUser_SetPassword(t *testing.T) {
	user := &User{}

	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "password too short",
			password: "123",
			wantErr:  true,
			errMsg:   "密码至少需要6位",
		},
		{
			name:     "password empty",
			password: "",
			wantErr:  true,
			errMsg:   "密码至少需要6位",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := user.SetPassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, user.Password)
			// 验证密码哈希不为明文
			assert.NotEqual(t, tt.password, user.Password)
		})
	}
}

func TestUser_CheckPassword(t *testing.T) {
	user := &User{}
	password := "password123"
	
	err := user.SetPassword(password)
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "correct password",
			input:    password,
			expected: true,
		},
		{
			name:     "wrong password",
			input:    "wrongpassword",
			expected: false,
		},
		{
			name:     "empty password",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := user.CheckPassword(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUser_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status UserStatus
		want   bool
	}{
		{
			name:   "active user",
			status: UserStatusActive,
			want:   true,
		},
		{
			name:   "inactive user",
			status: UserStatusInactive,
			want:   false,
		},
		{
			name:   "banned user",
			status: UserStatusBanned,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Status: tt.status}
			assert.Equal(t, tt.want, user.IsActive())
		})
	}
}

func TestUser_IsAdmin(t *testing.T) {
	tests := []struct {
		name string
		role Role
		want bool
	}{
		{
			name: "super admin",
			role: RoleSuperAdmin,
			want: true,
		},
		{
			name: "admin",
			role: RoleAdmin,
			want: true,
		},
		{
			name: "manager",
			role: RoleManager,
			want: false,
		},
		{
			name: "volunteer",
			role: RoleVolunteer,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.role}
			assert.Equal(t, tt.want, user.IsAdmin())
		})
	}
}

func TestUser_HasPermission(t *testing.T) {
	tests := []struct {
		name     string
		userRole Role
		required Role
		want     bool
	}{
		{
			name:     "super admin can do anything",
			userRole: RoleSuperAdmin,
			required: RoleVolunteer,
			want:     true,
		},
		{
			name:     "admin can manage manager",
			userRole: RoleAdmin,
			required: RoleManager,
			want:     true,
		},
		{
			name:     "manager cannot manage admin",
			userRole: RoleManager,
			required: RoleAdmin,
			want:     false,
		},
		{
			name:     "volunteer can only manage self",
			userRole: RoleVolunteer,
			required: RoleVolunteer,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.userRole}
			assert.Equal(t, tt.want, user.HasPermission(tt.required))
		})
	}
}

func TestUser_ValidateEmail(t *testing.T) {
	user := &User{Nickname: "TestUser", Phone: "13800138000"}

	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{
			name:  "valid email",
			email: "test@example.com",
			valid: true,
		},
		{
			name:  "invalid email no at",
			email: "testexample.com",
			valid: false,
		},
		{
			name:  "invalid email no domain",
			email: "test@",
			valid: false,
		},
		{
			name:  "empty email is valid",
			email: "",
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user.Email = tt.email
			err := user.Validate()
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestUser_RecordLogin(t *testing.T) {
	user := &User{}
	ip := "192.168.1.1"

	require.Nil(t, user.LastLoginAt)
	require.Empty(t, user.LastLoginIP)

	user.RecordLogin(ip)

	assert.NotNil(t, user.LastLoginAt)
	assert.Equal(t, ip, user.LastLoginIP)
}

func TestGetRoleLevel(t *testing.T) {
	tests := []struct {
		role  Role
		level RoleHierarchy
	}{
		{RoleSuperAdmin, RoleLevelSuperAdmin},
		{RoleAdmin, RoleLevelAdmin},
		{RoleManager, RoleLevelManager},
		{RoleVolunteer, RoleLevelVolunteer},
		{"unknown", RoleLevelVolunteer},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			level := GetRoleLevel(tt.role)
			assert.Equal(t, tt.level, level)
		})
	}
}

func TestHasRole(t *testing.T) {
	tests := []struct {
		name     string
		userRole Role
		required Role
		want     bool
	}{
		{"super_admin has all", RoleSuperAdmin, RoleVolunteer, true},
		{"admin has manager", RoleAdmin, RoleManager, true},
		{"manager has volunteer", RoleManager, RoleVolunteer, true},
		{"volunteer has volunteer", RoleVolunteer, RoleVolunteer, true},
		{"volunteer cannot admin", RoleVolunteer, RoleAdmin, false},
		{"manager cannot admin", RoleManager, RoleAdmin, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasRole(tt.userRole, tt.required)
			assert.Equal(t, tt.want, got)
		})
	}
}
