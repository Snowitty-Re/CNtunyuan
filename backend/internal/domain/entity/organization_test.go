// Package entity 组织实体测试
package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOrganization(t *testing.T) {
	tests := []struct {
		name     string
		orgName  string
		code     string
		orgType  OrgType
		parentID *string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid organization",
			orgName:  "测试组织",
			code:     "TEST001",
			orgType:  OrgTypeCity,
			parentID: nil,
			wantErr:  false,
		},
		{
			name:     "empty name",
			orgName:  "",
			code:     "TEST002",
			orgType:  OrgTypeCity,
			wantErr:  true,
			errMsg:   "组织名称不能为空",
		},
		{
			name:     "empty code",
			orgName:  "测试组织",
			code:     "",
			orgType:  OrgTypeCity,
			wantErr:  true,
			errMsg:   "组织编码不能为空",
		},
		{
			name:     "invalid type",
			orgName:  "测试组织",
			code:     "TEST003",
			orgType:  "invalid_type",
			wantErr:  true,
			errMsg:   "无效的组织类型",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			org, err := NewOrganization(tt.orgName, tt.code, tt.orgType, tt.parentID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, org.ID)
			assert.Equal(t, tt.orgName, org.Name)
			assert.Equal(t, tt.code, org.Code)
			assert.Equal(t, tt.orgType, org.Type)
			assert.Equal(t, OrgStatusActive, org.Status)
			assert.True(t, org.IsActive())
		})
	}
}

func TestNewRootOrganization(t *testing.T) {
	org, err := NewRootOrganization("根组织", "ROOT")
	require.NoError(t, err)

	assert.Equal(t, "00000000-0000-0000-0000-000000000000", org.ID)
	assert.Equal(t, "根组织", org.Name)
	assert.Equal(t, "ROOT", org.Code)
	assert.Equal(t, OrgTypeRoot, org.Type)
	assert.Equal(t, 1, org.Level)
	assert.Nil(t, org.ParentID)
	assert.True(t, org.IsRoot())
}

func TestOrganization_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status OrgStatus
		want   bool
	}{
		{"active", OrgStatusActive, true},
		{"inactive", OrgStatusInactive, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			org := &Organization{Status: tt.status}
			assert.Equal(t, tt.want, org.IsActive())
		})
	}
}

func TestOrganization_IsRoot(t *testing.T) {
	tests := []struct {
		name     string
		orgType  OrgType
		parentID *string
		want     bool
	}{
		{"root type", OrgTypeRoot, nil, true},
		{"nil parent", OrgTypeCity, nil, true},
		{"with parent", OrgTypeCity, strPtr("parent-id"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			org := &Organization{Type: tt.orgType, ParentID: tt.parentID}
			assert.Equal(t, tt.want, org.IsRoot())
		})
	}
}

func TestOrganization_CanHaveChildren(t *testing.T) {
	tests := []struct {
		name    string
		orgType OrgType
		want    bool
	}{
		{"root", OrgTypeRoot, true},
		{"province", OrgTypeProvince, true},
		{"city", OrgTypeCity, true},
		{"district", OrgTypeDistrict, true},
		{"street", OrgTypeStreet, true},
		{"community", OrgTypeCommunity, false},
		{"team", OrgTypeTeam, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			org := &Organization{Type: tt.orgType}
			assert.Equal(t, tt.want, org.CanHaveChildren())
		})
	}
}

func TestOrganization_GetLevelName(t *testing.T) {
	tests := []struct {
		orgType OrgType
		want    string
	}{
		{OrgTypeRoot, "总部"},
		{OrgTypeProvince, "省级"},
		{OrgTypeCity, "市级"},
		{OrgTypeDistrict, "区级"},
		{OrgTypeStreet, "街道"},
		{OrgTypeCommunity, "社区"},
		{OrgTypeTeam, "团队"},
		{"unknown", "未知"},
	}

	for _, tt := range tests {
		t.Run(string(tt.orgType), func(t *testing.T) {
			org := &Organization{Type: tt.orgType}
			assert.Equal(t, tt.want, org.GetLevelName())
		})
	}
}

func TestOrganization_SetParent(t *testing.T) {
	parentID := "parent-id"
	
	tests := []struct {
		name     string
		orgType  OrgType
		expected int
	}{
		{"root", OrgTypeRoot, 1},
		{"province", OrgTypeProvince, 2},
		{"city", OrgTypeCity, 3},
		{"district", OrgTypeDistrict, 4},
		{"street", OrgTypeStreet, 5},
		{"community", OrgTypeCommunity, 6},
		{"team", OrgTypeTeam, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			org := &Organization{Type: tt.orgType}
			org.SetParent(parentID)
			assert.Equal(t, &parentID, org.ParentID)
			assert.Equal(t, tt.expected, org.Level)
		})
	}
}

func strPtr(s string) *string {
	return &s
}
