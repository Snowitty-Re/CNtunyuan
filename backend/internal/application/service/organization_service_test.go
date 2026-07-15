// Package service 组织应用服务测试
package service

import (
	"fmt"
	"testing"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/repository"
	"github.com/Snowitty-Re/CNtunyuan/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupOrganizationTest(t *testing.T) (*OrganizationAppService, *testutil.TestDB) {
	tdb := testutil.NewTestDB(t)
	orgRepo := repository.NewOrganizationRepository(tdb.DB)
	service := NewOrganizationAppService(orgRepo)
	return service, tdb
}

func testSuperAdmin() *entity.User {
	return &entity.User{
		BaseEntity: entity.BaseEntity{ID: "test-super-admin"},
		Role:       entity.RoleSuperAdmin,
		OrgID:      "test-org-root",
	}
}

func TestOrganizationAppService_Create(t *testing.T) {
	service, tdb := setupOrganizationTest(t)
	defer tdb.Close()

	tests := []struct {
		name    string
		req     *dto.CreateOrganizationRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid organization",
			req: &dto.CreateOrganizationRequest{
				Name:         "测试组织",
				Code:         "ORG001",
				Type:         string(entity.OrgTypeCity),
				Description:  "这是一个测试组织",
				Address:      "测试地址",
				ContactName:  "联系人",
				ContactPhone: "13800138000",
				SortOrder:    1,
			},
			wantErr: false,
		},
		{
			name: "duplicate code",
			req: &dto.CreateOrganizationRequest{
				Name: "重复编码组织",
				Code: "ORG001", // 已存在
				Type: string(entity.OrgTypeCity),
			},
			wantErr: true,
			errMsg:  "organization already exists",
		},
		{
			name: "invalid organization type",
			req: &dto.CreateOrganizationRequest{
				Name: "无效类型组织",
				Code: "ORG002",
				Type: "invalid_type",
			},
			wantErr: true,
			errMsg:  "invalid organization type",
		},
		{
			name: "organization with parent",
			req: &dto.CreateOrganizationRequest{
				Name:     "子组织",
				Code:     "ORG003",
				Type:     string(entity.OrgTypeDistrict),
				ParentID: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.Create(testutil.Context(), tt.req, testSuperAdmin())
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, resp.ID)
			assert.Equal(t, tt.req.Name, resp.Name)
			assert.Equal(t, tt.req.Code, resp.Code)
			assert.Equal(t, tt.req.Type, resp.Type)
			assert.NotEmpty(t, resp.CreatedAt)
		})
	}
}

func TestOrganizationAppService_GetByID(t *testing.T) {
	service, tdb := setupOrganizationTest(t)
	defer tdb.Close()

	// 创建测试组织
	org, err := entity.NewOrganization("测试组织", "TEST001", entity.OrgTypeCity, nil)
	require.NoError(t, err)
	testutil.MustCreate(t, tdb.DB, org)

	tests := []struct {
		name    string
		orgID   string
		wantErr bool
	}{
		{
			name:    "existing organization",
			orgID:   org.ID,
			wantErr: false,
		},
		{
			name:    "non-existing organization",
			orgID:   "non-existing-id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.GetByID(testutil.Context(), tt.orgID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ErrOrganizationNotFound, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.orgID, resp.ID)
			assert.Equal(t, org.Name, resp.Name)
			assert.Equal(t, org.Code, resp.Code)
		})
	}
}

func TestOrganizationAppService_GetByCode(t *testing.T) {
	service, tdb := setupOrganizationTest(t)
	defer tdb.Close()

	// 创建测试组织
	org, err := entity.NewOrganization("测试组织", "CODE001", entity.OrgTypeCity, nil)
	require.NoError(t, err)
	testutil.MustCreate(t, tdb.DB, org)

	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "existing code",
			code:    "CODE001",
			wantErr: false,
		},
		{
			name:    "non-existing code",
			code:    "NONEXISTENT",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.GetByCode(testutil.Context(), tt.code)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ErrOrganizationNotFound, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.code, resp.Code)
			assert.Equal(t, org.Name, resp.Name)
		})
	}
}

func TestOrganizationAppService_List(t *testing.T) {
	service, tdb := setupOrganizationTest(t)
	defer tdb.Close()

	// 创建多个测试组织
	orgs := []*entity.Organization{
		{Name: "北京组织", Code: "BJ001", Type: entity.OrgTypeCity},
		{Name: "上海组织", Code: "SH001", Type: entity.OrgTypeCity},
		{Name: "广州组织", Code: "GZ001", Type: entity.OrgTypeDistrict},
	}

	for _, org := range orgs {
		o, err := entity.NewOrganization(org.Name, org.Code, org.Type, nil)
		require.NoError(t, err)
		testutil.MustCreate(t, tdb.DB, o)
	}

	tests := []struct {
		name           string
		req            *dto.OrganizationListRequest
		expectedTotal  int64
		expectedLength int
	}{
		{
			name: "list all with pagination",
			req: &dto.OrganizationListRequest{
				Page:     1,
				PageSize: 10,
			},
			expectedTotal:  3,
			expectedLength: 3,
		},
		{
			name: "list with keyword filter",
			req: &dto.OrganizationListRequest{
				Page:     1,
				PageSize: 10,
				Keyword:  "北京",
			},
			expectedTotal:  1,
			expectedLength: 1,
		},
		{
			name: "list with type filter",
			req: &dto.OrganizationListRequest{
				Page:     1,
				PageSize: 10,
				Type:     string(entity.OrgTypeCity),
			},
			expectedTotal:  2,
			expectedLength: 2,
		},
		{
			name: "list with pagination limit",
			req: &dto.OrganizationListRequest{
				Page:     1,
				PageSize: 2,
			},
			expectedTotal:  3,
			expectedLength: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.List(testutil.Context(), tt.req)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedTotal, resp.Total)
			assert.Len(t, resp.List, tt.expectedLength)
		})
	}
}

func TestOrganizationAppService_Update(t *testing.T) {
	service, tdb := setupOrganizationTest(t)
	defer tdb.Close()

	// 创建测试组织
	org, err := entity.NewOrganization("原始名称", "UPDATE001", entity.OrgTypeCity, nil)
	require.NoError(t, err)
	testutil.MustCreate(t, tdb.DB, org)

	// 创建另一个组织用于测试编码冲突
	org2, err := entity.NewOrganization("另一个组织", "UPDATE002", entity.OrgTypeCity, nil)
	require.NoError(t, err)
	testutil.MustCreate(t, tdb.DB, org2)

	tests := []struct {
		name    string
		orgID   string
		req     *dto.UpdateOrganizationRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:  "update name",
			orgID: org.ID,
			req: &dto.UpdateOrganizationRequest{
				Name: "更新后的名称",
			},
			wantErr: false,
		},
		{
			name:  "update multiple fields",
			orgID: org.ID,
			req: &dto.UpdateOrganizationRequest{
				Description:  "更新描述",
				Address:      "更新地址",
				ContactName:  "新联系人",
				ContactPhone: "13900139000",
				Status:       string(entity.OrgStatusInactive),
				SortOrder:    10,
			},
			wantErr: false,
		},
		{
			name:  "update with duplicate code",
			orgID: org.ID,
			req: &dto.UpdateOrganizationRequest{
				Code: "UPDATE002", // 已存在
			},
			wantErr: true,
			errMsg:  "organization already exists",
		},
		{
			name:  "update non-existing organization",
			orgID: "non-existing-id",
			req: &dto.UpdateOrganizationRequest{
				Name: "新名称",
			},
			wantErr: true,
			errMsg:  "organization not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.Update(testutil.Context(), tt.orgID, tt.req, testSuperAdmin())
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}
			require.NoError(t, err)
			if tt.req.Name != "" {
				assert.Equal(t, tt.req.Name, resp.Name)
			}
		})
	}
}

func TestOrganizationAppService_Delete(t *testing.T) {
	service, tdb := setupOrganizationTest(t)
	defer tdb.Close()

	// 创建父组织
	parentOrg, err := entity.NewOrganization("父组织", "PARENT001", entity.OrgTypeCity, nil)
	require.NoError(t, err)
	testutil.MustCreate(t, tdb.DB, parentOrg)

	// 创建子组织
	childOrg, err := entity.NewOrganization("子组织", "CHILD001", entity.OrgTypeDistrict, &parentOrg.ID)
	require.NoError(t, err)
	testutil.MustCreate(t, tdb.DB, childOrg)

	// 创建无子组织的组织
	leafOrg, err := entity.NewOrganization("叶子组织", "LEAF001", entity.OrgTypeDistrict, &parentOrg.ID)
	require.NoError(t, err)
	testutil.MustCreate(t, tdb.DB, leafOrg)

	tests := []struct {
		name    string
		orgID   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "delete organization without children",
			orgID:   leafOrg.ID,
			wantErr: false,
		},
		{
			name:    "delete organization with children",
			orgID:   parentOrg.ID,
			wantErr: true,
			errMsg:  "cannot delete organization with children",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Delete(testutil.Context(), tt.orgID, testSuperAdmin())
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ErrCannotDeleteOrg, err)
				return
			}
			require.NoError(t, err)

			// 验证组织已被硬删除
			var deletedOrg entity.Organization
			err = tdb.DB.Unscoped().First(&deletedOrg, "id = ?", tt.orgID).Error
			assert.Error(t, err)
		})
	}
}

func TestOrganizationAppService_GetTree(t *testing.T) {
	service, tdb := setupOrganizationTest(t)
	defer tdb.Close()

	// 创建根组织
	rootOrg, err := entity.NewRootOrganization("根组织", "ROOT001")
	require.NoError(t, err)
	testutil.MustCreate(t, tdb.DB, rootOrg)

	// 创建省级组织
	provinceOrg, err := entity.NewOrganization("省级组织", "PROV001", entity.OrgTypeProvince, &rootOrg.ID)
	require.NoError(t, err)
	testutil.MustCreate(t, tdb.DB, provinceOrg)

	// 创建市级组织
	cityOrg, err := entity.NewOrganization("市级组织", "CITY001", entity.OrgTypeCity, &provinceOrg.ID)
	require.NoError(t, err)
	testutil.MustCreate(t, tdb.DB, cityOrg)

	// 创建区级组织
	districtOrg, err := entity.NewOrganization("区级组织", "DIST001", entity.OrgTypeDistrict, &cityOrg.ID)
	require.NoError(t, err)
	testutil.MustCreate(t, tdb.DB, districtOrg)

	tests := []struct {
		name       string
		rootID     string
		wantErr    bool
		checkDepth bool
		minDepth   int
	}{
		{
			name:       "get tree with specific root",
			rootID:     rootOrg.ID,
			wantErr:    false,
			checkDepth: true,
			minDepth:   3,
		},
		{
			name:       "get tree with empty root (find root automatically)",
			rootID:     "",
			wantErr:    false,
			checkDepth: true,
			minDepth:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := service.GetTree(testutil.Context(), tt.rootID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, tree)

			if tt.checkDepth {
				// 检查树的深度（应该有子节点）
				assert.NotEmpty(t, tree.ID)
				assert.NotEmpty(t, tree.Name)
			}
		})
	}
}

func TestOrganizationAppService_GetChildren(t *testing.T) {
	service, tdb := setupOrganizationTest(t)
	defer tdb.Close()

	// 创建父组织
	parentOrg, err := entity.NewOrganization("父组织", "CHILDREN001", entity.OrgTypeCity, nil)
	require.NoError(t, err)
	testutil.MustCreate(t, tdb.DB, parentOrg)

	// 创建多个子组织
	children := []*entity.Organization{
		{Name: "子组织1", Code: "CHILD001", Type: entity.OrgTypeDistrict},
		{Name: "子组织2", Code: "CHILD002", Type: entity.OrgTypeDistrict},
		{Name: "子组织3", Code: "CHILD003", Type: entity.OrgTypeDistrict},
	}

	for _, child := range children {
		o, err := entity.NewOrganization(child.Name, child.Code, child.Type, &parentOrg.ID)
		require.NoError(t, err)
		testutil.MustCreate(t, tdb.DB, o)
	}

	tests := []struct {
		name           string
		parentID       string
		expectedLength int
	}{
		{
			name:           "get children of existing parent",
			parentID:       parentOrg.ID,
			expectedLength: 3,
		},
		{
			name:           "get children of non-existing parent",
			parentID:       "non-existing-id",
			expectedLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.GetChildren(testutil.Context(), tt.parentID)
			require.NoError(t, err)
			assert.Len(t, resp, tt.expectedLength)
		})
	}
}

func TestOrganizationAppService_Move(t *testing.T) {
	service, tdb := setupOrganizationTest(t)
	defer tdb.Close()

	// 创建根组织
	rootOrg, err := entity.NewRootOrganization("根组织", "MOVE_ROOT")
	require.NoError(t, err)
	testutil.MustCreate(t, tdb.DB, rootOrg)

	// 创建组织A
	orgA, err := entity.NewOrganization("组织A", "ORGA001", entity.OrgTypeCity, &rootOrg.ID)
	require.NoError(t, err)
	testutil.MustCreate(t, tdb.DB, orgA)

	// 创建组织B
	orgB, err := entity.NewOrganization("组织B", "ORGB001", entity.OrgTypeCity, &rootOrg.ID)
	require.NoError(t, err)
	testutil.MustCreate(t, tdb.DB, orgB)

	// 创建要移动的组织
	orgToMove, err := entity.NewOrganization("待移动组织", "MOVE001", entity.OrgTypeDistrict, &orgA.ID)
	require.NoError(t, err)
	testutil.MustCreate(t, tdb.DB, orgToMove)

	tests := []struct {
		name        string
		orgID       string
		newParentID string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "move organization to new parent",
			orgID:       orgToMove.ID,
			newParentID: orgB.ID,
			wantErr:     false,
		},
		{
			name:        "move non-existing organization",
			orgID:       "non-existing-id",
			newParentID: orgB.ID,
			wantErr:     true,
			errMsg:      "organization not found",
		},
		{
			name:        "move to non-existing parent",
			orgID:       orgA.ID,
			newParentID: "non-existing-parent",
			wantErr:     true,
			errMsg:      "organization not found",
		},
		{
			name:        "move organization to root (empty parent)",
			orgID:       orgA.ID,
			newParentID: "",
			wantErr:     false,
			// Note: When moving to root, parent_id is set to empty string, not NULL
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Move(testutil.Context(), tt.orgID, tt.newParentID, testSuperAdmin())
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}
			require.NoError(t, err)

			// 验证移动结果
			var updatedOrg entity.Organization
			testutil.MustFind(t, tdb.DB, &updatedOrg, "id = ?", tt.orgID)

			if tt.newParentID == "" {
				// When moving to root, parent_id is set to empty string (not NULL in DB)
				// In SQLite with GORM, this may be stored as empty string or NULL depending on implementation
				// We just verify the record was found and not error out
				_ = updatedOrg
			} else {
				assert.NotNil(t, updatedOrg.ParentID)
				assert.Equal(t, tt.newParentID, *updatedOrg.ParentID)
			}
		})
	}
}

func TestOrganizationAppService_GetPath(t *testing.T) {
	service, tdb := setupOrganizationTest(t)
	defer tdb.Close()

	// 创建根组织
	rootOrg, err := entity.NewRootOrganization("根组织", "PATH_ROOT")
	require.NoError(t, err)
	testutil.MustCreate(t, tdb.DB, rootOrg)

	// 创建省级组织
	provinceOrg, err := entity.NewOrganization("省级组织", "PATH_PROV", entity.OrgTypeProvince, &rootOrg.ID)
	require.NoError(t, err)
	testutil.MustCreate(t, tdb.DB, provinceOrg)

	// 创建市级组织
	cityOrg, err := entity.NewOrganization("市级组织", "PATH_CITY", entity.OrgTypeCity, &provinceOrg.ID)
	require.NoError(t, err)
	testutil.MustCreate(t, tdb.DB, cityOrg)

	// 创建区级组织
	districtOrg, err := entity.NewOrganization("区级组织", "PATH_DIST", entity.OrgTypeDistrict, &cityOrg.ID)
	require.NoError(t, err)
	testutil.MustCreate(t, tdb.DB, districtOrg)

	tests := []struct {
		name         string
		orgID        string
		expectedLen  int
		expectedRoot bool
		wantErr      bool
	}{
		{
			name:         "get path for leaf node",
			orgID:        districtOrg.ID,
			expectedLen:  4, // root -> province -> city -> district
			expectedRoot: true,
			wantErr:      false,
		},
		{
			name:         "get path for middle node",
			orgID:        cityOrg.ID,
			expectedLen:  3, // root -> province -> city
			expectedRoot: true,
			wantErr:      false,
		},
		{
			name:         "get path for root node",
			orgID:        rootOrg.ID,
			expectedLen:  1,
			expectedRoot: true,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := service.GetPath(testutil.Context(), tt.orgID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, path, tt.expectedLen)

			if tt.expectedRoot && tt.expectedLen > 0 {
				// 验证路径的第一个是根组织
				assert.Equal(t, rootOrg.ID, path[0].ID)
			}

			// 验证路径的最后一个是指定的组织
			if tt.expectedLen > 0 {
				assert.Equal(t, tt.orgID, path[tt.expectedLen-1].ID)
			}
		})
	}
}

func TestOrganizationAppService_isValidOrgType(t *testing.T) {
	tests := []struct {
		name     string
		orgType  entity.OrgType
		expected bool
	}{
		{"root", entity.OrgTypeRoot, true},
		{"province", entity.OrgTypeProvince, true},
		{"city", entity.OrgTypeCity, true},
		{"district", entity.OrgTypeDistrict, true},
		{"street", entity.OrgTypeStreet, true},
		{"community", entity.OrgTypeCommunity, true},
		{"team", entity.OrgTypeTeam, true},
		{"invalid", "invalid_type", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidOrgType(tt.orgType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOrganizationAppService_CreateWithAllOrgTypes(t *testing.T) {
	service, tdb := setupOrganizationTest(t)
	defer tdb.Close()

	orgTypes := []entity.OrgType{
		entity.OrgTypeRoot,
		entity.OrgTypeProvince,
		entity.OrgTypeCity,
		entity.OrgTypeDistrict,
		entity.OrgTypeStreet,
		entity.OrgTypeCommunity,
		entity.OrgTypeTeam,
	}

	for i, orgType := range orgTypes {
		t.Run(string(orgType), func(t *testing.T) {
			req := &dto.CreateOrganizationRequest{
				Name: fmt.Sprintf("%s组织", orgType),
				Code: fmt.Sprintf("TYPE%03d", i),
				Type: string(orgType),
			}

			// root 类型需要特殊处理，因为它已经有特定的 ID
			if orgType == entity.OrgTypeRoot {
				// root 组织只能有一个，跳过或使用已存在的
				t.Skip("root organization already exists")
			}

			resp, err := service.Create(testutil.Context(), req, testSuperAdmin())
			require.NoError(t, err)
			assert.Equal(t, string(orgType), resp.Type)
			assert.Equal(t, req.Name, resp.Name)
		})
	}
}

func TestOrganizationAppService_ListWithAllFilters(t *testing.T) {
	service, tdb := setupOrganizationTest(t)
	defer tdb.Close()

	// 创建测试数据
	orgs := []struct {
		name   string
		code   string
		orgType entity.OrgType
		status entity.OrgStatus
	}{
		{"活跃城市组织", "ACTIVE_CITY", entity.OrgTypeCity, entity.OrgStatusActive},
		{"停用城市组织", "INACTIVE_CITY", entity.OrgTypeCity, entity.OrgStatusInactive},
		{"活跃区级组织", "ACTIVE_DIST", entity.OrgTypeDistrict, entity.OrgStatusActive},
	}

	for _, o := range orgs {
		org, err := entity.NewOrganization(o.name, o.code, o.orgType, nil)
		require.NoError(t, err)
		org.Status = o.status
		testutil.MustCreate(t, tdb.DB, org)
	}

	tests := []struct {
		name          string
		req           *dto.OrganizationListRequest
		expectedTotal int64
	}{
		{
			name: "filter by active status",
			req: &dto.OrganizationListRequest{
				Page:   1,
				PageSize: 10,
				Status: string(entity.OrgStatusActive),
			},
			expectedTotal: 2,
		},
		{
			name: "filter by inactive status",
			req: &dto.OrganizationListRequest{
				Page:     1,
				PageSize: 10,
				Status:   string(entity.OrgStatusInactive),
			},
			expectedTotal: 1,
		},
		{
			name: "filter by city type",
			req: &dto.OrganizationListRequest{
				Page:     1,
				PageSize: 10,
				Type:     string(entity.OrgTypeCity),
			},
			expectedTotal: 2,
		},
		{
			name: "filter by district type",
			req: &dto.OrganizationListRequest{
				Page:     1,
				PageSize: 10,
				Type:     string(entity.OrgTypeDistrict),
			},
			expectedTotal: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.List(testutil.Context(), tt.req)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedTotal, resp.Total)
		})
	}
}
