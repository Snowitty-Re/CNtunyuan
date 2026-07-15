package service

import (
	"testing"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/repository"
	"github.com/Snowitty-Re/CNtunyuan/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanAssignRole_Ceiling(t *testing.T) {
	admin := &entity.User{Role: entity.RoleAdmin}
	super := &entity.User{Role: entity.RoleSuperAdmin}
	manager := &entity.User{Role: entity.RoleManager}

	assert.False(t, canAssignRole(admin, entity.RoleSuperAdmin))
	assert.False(t, canAssignRole(admin, entity.RoleAdmin))
	assert.True(t, canAssignRole(admin, entity.RoleManager))
	assert.True(t, canAssignRole(admin, entity.RoleVolunteer))

	assert.True(t, canAssignRole(super, entity.RoleSuperAdmin))
	assert.True(t, canAssignRole(super, entity.RoleAdmin))

	assert.False(t, canAssignRole(manager, entity.RoleAdmin))
	assert.False(t, canAssignRole(manager, entity.RoleManager))
	assert.True(t, canAssignRole(manager, entity.RoleVolunteer))
	assert.False(t, canAssignRole(nil, entity.RoleVolunteer))
}

func TestCanManageOrg_NotBoundToUserCreate(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	orgRepo := repository.NewOrganizationRepository(tdb.DB)
	authz := NewAuthorizationService(orgRepo)

	root := &entity.Organization{
		BaseEntity: entity.BaseEntity{ID: "org-root"},
		Name:       "根",
		Code:       "ROOT",
		Type:       entity.OrgTypeRoot,
	}
	testutil.MustCreate(t, tdb.DB, root)

	child := &entity.Organization{
		BaseEntity: entity.BaseEntity{ID: "org-child"},
		Name:       "子",
		Code:       "CHILD",
		Type:       entity.OrgTypeCity,
	}
	pid := root.ID
	child.ParentID = &pid
	testutil.MustCreate(t, tdb.DB, child)

	foreign := &entity.Organization{
		BaseEntity: entity.BaseEntity{ID: "org-foreign"},
		Name:       "外",
		Code:       "FOREIGN",
		Type:       entity.OrgTypeCity,
	}
	testutil.MustCreate(t, tdb.DB, foreign)

	// manager has no user:create but CanManageOrg should still allow own org subtree
	manager := &entity.User{
		BaseEntity: entity.BaseEntity{ID: "mgr-1"},
		Role:       entity.RoleManager,
		OrgID:      root.ID,
	}

	d, err := authz.CanManageOrg(testutil.Context(), manager, root.ID)
	require.NoError(t, err)
	assert.True(t, d.Allowed, "manager should manage own org")

	d, err = authz.CanManageOrg(testutil.Context(), manager, child.ID)
	require.NoError(t, err)
	assert.True(t, d.Allowed, "manager should manage descendant org")

	d, err = authz.CanManageOrg(testutil.Context(), manager, foreign.ID)
	require.NoError(t, err)
	assert.False(t, d.Allowed, "manager must not manage foreign org")

	// volunteer cannot manage org
	volunteer := &entity.User{
		BaseEntity: entity.BaseEntity{ID: "vol-1"},
		Role:       entity.RoleVolunteer,
		OrgID:      root.ID,
	}
	d, err = authz.CanManageOrg(testutil.Context(), volunteer, root.ID)
	require.NoError(t, err)
	assert.False(t, d.Allowed)
}
