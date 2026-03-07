package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			got := user.IsActive()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUser_IsAdmin(t *testing.T) {
	tests := []struct {
		name string
		role string
		want bool
	}{
		{
			name: "super admin",
			role: string(RoleSuperAdmin),
			want: true,
		},
		{
			name: "admin",
			role: string(RoleAdmin),
			want: true,
		},
		{
			name: "manager",
			role: string(RoleManager),
			want: false,
		},
		{
			name: "volunteer",
			role: string(RoleVolunteer),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.role}
			got := user.IsAdmin()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUser_IsSuperAdmin(t *testing.T) {
	tests := []struct {
		name string
		role string
		want bool
	}{
		{
			name: "super admin",
			role: string(RoleSuperAdmin),
			want: true,
		},
		{
			name: "admin",
			role: string(RoleAdmin),
			want: false,
		},
		{
			name: "volunteer",
			role: string(RoleVolunteer),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.role}
			got := user.IsSuperAdmin()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr bool
	}{
		{
			name: "valid user",
			user: User{
				Nickname: "Test User",
				Phone:    "13800138000",
				Email:    "test@example.com",
			},
			wantErr: false,
		},
		{
			name: "empty nickname",
			user: User{
				Nickname: "",
				Phone:    "13800138000",
			},
			wantErr: true,
		},
		{
			name: "nickname too long",
			user: User{
				Nickname: string(make([]byte, 101)),
				Phone:    "13800138000",
			},
			wantErr: true,
		},
		{
			name: "invalid phone",
			user: User{
				Nickname: "Test User",
				Phone:    "123456",
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			user: User{
				Nickname: "Test User",
				Phone:    "13800138000",
				Email:    "invalid-email",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUser_ValidatePhone(t *testing.T) {
	tests := []struct {
		name    string
		phone   string
		wantErr bool
	}{
		{
			name:    "valid phone",
			phone:   "13800138000",
			wantErr: false,
		},
		{
			name:    "valid phone 199",
			phone:   "19900138000",
			wantErr: false,
		},
		{
			name:    "invalid phone - too short",
			phone:   "138001",
			wantErr: true,
		},
		{
			name:    "invalid phone - starts with 2",
			phone:   "23800138000",
			wantErr: true,
		},
		{
			name:    "invalid phone - contains letters",
			phone:   "1380013800a",
			wantErr: true,
		},
		{
			name:    "empty phone",
			phone:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Phone: tt.phone}
			err := user.ValidatePhone()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUser_SetPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "password too short",
			password: "12345",
			wantErr:  true,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{}
			err := user.SetPassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, user.Password)
				// Verify password can be checked
				assert.True(t, user.CheckPassword(tt.password))
			}
		})
	}
}

func TestUser_CheckPassword(t *testing.T) {
	user := &User{}
	user.SetPassword("correctpassword")

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{
			name:     "correct password",
			password: "correctpassword",
			want:     true,
		},
		{
			name:     "wrong password",
			password: "wrongpassword",
			want:     false,
		},
		{
			name:     "empty password",
			password: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := user.CheckPassword(tt.password)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUser_HasPermission(t *testing.T) {
	tests := []struct {
		name     string
		userRole string
		required string
		want     bool
	}{
		{
			name:     "super admin has all permissions",
			userRole: string(RoleSuperAdmin),
			required: string(RoleVolunteer),
			want:     true,
		},
		{
			name:     "admin has manager permission",
			userRole: string(RoleAdmin),
			required: string(RoleManager),
			want:     true,
		},
		{
			name:     "manager does not have admin permission",
			userRole: string(RoleManager),
			required: string(RoleAdmin),
			want:     false,
		},
		{
			name:     "volunteer has only volunteer permission",
			userRole: string(RoleVolunteer),
			required: string(RoleVolunteer),
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.userRole}
			got := user.HasPermission(tt.required)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUser_CanModify(t *testing.T) {
	tests := []struct {
		name     string
		operator User
		target   User
		want     bool
	}{
		{
			name:     "super admin can modify anyone",
			operator: User{Role: string(RoleSuperAdmin)},
			target:   User{Role: string(RoleAdmin)},
			want:     true,
		},
		{
			name:     "admin cannot modify super admin",
			operator: User{Role: string(RoleAdmin)},
			target:   User{Role: string(RoleSuperAdmin)},
			want:     false,
		},
		{
			name:     "admin can modify volunteer",
			operator: User{Role: string(RoleAdmin)},
			target:   User{Role: string(RoleVolunteer)},
			want:     true,
		},
		{
			name:     "manager cannot modify admin",
			operator: User{Role: string(RoleManager)},
			target:   User{Role: string(RoleAdmin)},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.operator.CanModify(&tt.target)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetRoleLevel(t *testing.T) {
	tests := []struct {
		name string
		role LegacyRole
		want RoleHierarchy
	}{
		{
			name: "super admin level",
			role: RoleSuperAdmin,
			want: RoleLevelSuperAdmin,
		},
		{
			name: "admin level",
			role: RoleAdmin,
			want: RoleLevelAdmin,
		},
		{
			name: "manager level",
			role: RoleManager,
			want: RoleLevelManager,
		},
		{
			name: "volunteer level",
			role: RoleVolunteer,
			want: RoleLevelVolunteer,
		},
		{
			name: "unknown role defaults to volunteer",
			role: LegacyRole("unknown"),
			want: RoleLevelVolunteer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetRoleLevel(string(tt.role))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHasRole(t *testing.T) {
	tests := []struct {
		name         string
		userRole     string
		requiredRole string
		want         bool
	}{
		{
			name:         "super admin has all roles",
			userRole:     string(RoleSuperAdmin),
			requiredRole: string(RoleVolunteer),
			want:         true,
		},
		{
			name:         "admin has manager role",
			userRole:     string(RoleAdmin),
			requiredRole: string(RoleManager),
			want:         true,
		},
		{
			name:         "manager does not have admin role",
			userRole:     string(RoleManager),
			requiredRole: string(RoleAdmin),
			want:         false,
		},
		{
			name:         "same role has permission",
			userRole:     string(RoleVolunteer),
			requiredRole: string(RoleVolunteer),
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasRole(tt.userRole, tt.requiredRole)
			assert.Equal(t, tt.want, got)
		})
	}
}
