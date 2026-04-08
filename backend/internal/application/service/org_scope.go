package service

import (
	"context"
	"strings"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
)

const unboundOrgID = "00000000-0000-0000-0000-000000000000"

func normalizeOrgID(orgID string) string {
	return strings.TrimSpace(orgID)
}

func isSuperAdmin(user *entity.User) bool {
	return user != nil && user.Role == entity.RoleSuperAdmin
}

func collectManageableOrgIDs(ctx context.Context, orgRepo repository.OrganizationRepository, operator *entity.User) ([]string, error) {
	if operator == nil {
		return nil, nil
	}
	if isSuperAdmin(operator) {
		return nil, nil
	}

	operatorOrgID := normalizeOrgID(operator.OrgID)
	if operatorOrgID == "" {
		return []string{unboundOrgID}, nil
	}
	if orgRepo == nil {
		return []string{operatorOrgID}, nil
	}

	children, err := orgRepo.FindChildren(ctx, operatorOrgID)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(children)+1)
	ids = append(ids, operatorOrgID)
	for _, child := range children {
		childID := normalizeOrgID(child.ID)
		if childID == "" {
			continue
		}
		ids = append(ids, childID)
	}

	return ids, nil
}

func canManageTargetOrg(ctx context.Context, orgRepo repository.OrganizationRepository, operator *entity.User, targetOrgID string) (bool, error) {
	if isSuperAdmin(operator) {
		return true, nil
	}

	targetOrgID = normalizeOrgID(targetOrgID)
	if targetOrgID == "" {
		targetOrgID = unboundOrgID
	}
	allowedOrgIDs, err := collectManageableOrgIDs(ctx, orgRepo, operator)
	if err != nil {
		return false, err
	}

	for _, orgID := range allowedOrgIDs {
		if orgID == targetOrgID {
			return true, nil
		}
	}
	return false, nil
}
