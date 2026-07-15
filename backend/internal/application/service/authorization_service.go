package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
)

// AuthzReason 授权决策原因
type AuthzReason string

const (
	AuthzAllow                AuthzReason = "ALLOW"
	AuthzDenyUnauthenticated              = "DENY_UNAUTHENTICATED"
	AuthzDenyRoleLevel                    = "DENY_ROLE_LEVEL"
	AuthzDenyTargetSuperAdmin             = "DENY_TARGET_SUPER_ADMIN"
	AuthzDenyOrgScope                     = "DENY_ORG_SCOPE"
	AuthzDenyOwnerOnly                    = "DENY_OWNER_ONLY"
	AuthzDenyAssigneeOnly                 = "DENY_ASSIGNEE_ONLY"
	AuthzDenyReporterOnly                 = "DENY_REPORTER_ONLY"
	AuthzDenyPolicy                       = "DENY_POLICY"
)

// AuthzDecision 授权决策结果
type AuthzDecision struct {
	Allowed bool
	Reason  AuthzReason
}

type RolePermissionsPreviewResult struct {
	Role    entity.Role `json:"role"`
	Before  []string    `json:"before"`
	After   []string    `json:"after"`
	Added   []string    `json:"added"`
	Removed []string    `json:"removed"`
	Changed bool        `json:"changed"`
}

type PolicyRulesPreviewResult struct {
	PermissionCode string                  `json:"permission_code"`
	Before         []repository.PolicyRule `json:"before"`
	After          []repository.PolicyRule `json:"after"`
	Added          []string                `json:"added"`
	Removed        []string                `json:"removed"`
	Changed        bool                    `json:"changed"`
}

type AuthzPolicyRollbackResult struct {
	ChangeID      string `json:"change_id"`
	ChangeType    string `json:"change_type"`
	TargetKey     string `json:"target_key"`
	Applied       bool   `json:"applied"`
	ExecutedLogID string `json:"executed_log_id,omitempty"`
}

type AuthzPolicyRollbackPreviewResult struct {
	ChangeID   string      `json:"change_id"`
	ChangeType string      `json:"change_type"`
	TargetKey  string      `json:"target_key"`
	Operation  string      `json:"operation"`
	Changed    bool        `json:"changed"`
	Impact     interface{} `json:"impact"`
}

type AuthzPolicyChangeRequestSubmitResult struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
}

type AuthzPolicyChangeReviewResult struct {
	RequestID  string `json:"request_id"`
	Status     string `json:"status"`
	Executed   bool   `json:"executed"`
	ExecutedID string `json:"executed_log_id,omitempty"`
}

type authzAction string

const (
	actionUserCreate    authzAction = "user:create"
	actionUserModify    authzAction = "user:modify"
	actionUserView      authzAction = "user:view"
	actionTaskManage    authzAction = "task:manage"
	actionTaskView      authzAction = "task:view"
	actionTaskEdit      authzAction = "task:edit"
	actionTaskExecute   authzAction = "task:execute"
	actionMissingModify authzAction = "missing:modify"
	actionMissingManage authzAction = "missing:manage"
	actionDialectModify authzAction = "dialect:modify"
	actionDialectManage authzAction = "dialect:manage"
)

var roleActions = map[entity.Role]map[authzAction]struct{}{
	entity.RoleSuperAdmin: {
		actionUserCreate:    {},
		actionUserModify:    {},
		actionUserView:      {},
		actionTaskManage:    {},
		actionTaskView:      {},
		actionTaskEdit:      {},
		actionTaskExecute:   {},
		actionMissingModify: {},
		actionMissingManage: {},
		actionDialectModify: {},
		actionDialectManage: {},
	},
	entity.RoleAdmin: {
		actionUserCreate:    {},
		actionUserModify:    {},
		actionUserView:      {},
		actionTaskManage:    {},
		actionTaskView:      {},
		actionTaskEdit:      {},
		actionTaskExecute:   {},
		actionMissingModify: {},
		actionMissingManage: {},
		actionDialectModify: {},
		actionDialectManage: {},
	},
	entity.RoleManager: {
		actionUserView:      {},
		actionUserModify:    {},
		actionTaskManage:    {},
		actionTaskView:      {},
		actionTaskEdit:      {},
		actionTaskExecute:   {},
		actionMissingModify: {},
		actionMissingManage: {},
		actionDialectModify: {},
		actionDialectManage: {},
	},
	entity.RoleVolunteer: {
		actionUserView:      {},
		actionUserModify:    {},
		actionTaskView:      {},
		actionTaskEdit:      {},
		actionTaskExecute:   {},
		actionMissingModify: {},
		actionDialectModify: {},
	},
}

func isValidAuthzRole(role entity.Role) bool {
	switch role {
	case entity.RoleSuperAdmin, entity.RoleAdmin, entity.RoleManager, entity.RoleVolunteer:
		return true
	default:
		return false
	}
}

var supportedPermissionCodes = map[string]struct{}{
	string(actionUserCreate):    {},
	string(actionUserModify):    {},
	string(actionUserView):      {},
	string(actionTaskManage):    {},
	string(actionTaskView):      {},
	string(actionTaskEdit):      {},
	string(actionTaskExecute):   {},
	string(actionMissingModify): {},
	string(actionMissingManage): {},
	string(actionDialectModify): {},
	string(actionDialectManage): {},
}

var supportedScopeRules = map[string]struct{}{
	"GLOBAL":         {},
	"SELF":           {},
	"ORG_DESCENDANT": {},
	"CREATOR":        {},
	"OWNER":          {},
	"ASSIGNEE":       {},
	"REPORTER":       {},
}

// AuthorizationService 统一权限服务（RBAC + ABAC）
type AuthorizationService struct {
	orgRepo          repository.OrganizationRepository
	policyRepo       repository.AuthzPolicyRepository
	notificationRepo repository.SystemNotificationRepository
	roleActions      sync.Map // map[entity.Role]map[authzAction]struct{}
	policyRules      sync.Map // map[string][]repository.PolicyRule
}

func NewAuthorizationService(orgRepo repository.OrganizationRepository, policyRepo ...repository.AuthzPolicyRepository) *AuthorizationService {
	var repo repository.AuthzPolicyRepository
	if len(policyRepo) > 0 {
		repo = policyRepo[0]
	}
	return &AuthorizationService{
		orgRepo:    orgRepo,
		policyRepo: repo,
	}
}

func (s *AuthorizationService) SetNotificationRepository(repo repository.SystemNotificationRepository) {
	s.notificationRepo = repo
}

func (s *AuthorizationService) RefreshPolicies() {
	s.roleActions = sync.Map{}
	s.policyRules = sync.Map{}
}

func (s *AuthorizationService) ListRolePermissions(ctx context.Context) ([]repository.RolePermission, error) {
	if s.policyRepo == nil {
		return nil, fmt.Errorf("policy repository not configured")
	}
	return s.policyRepo.ListAllRolePermissions(ctx)
}

func (s *AuthorizationService) ReplaceRolePermissions(ctx context.Context, role entity.Role, permissionCodes []string) error {
	_, err := s.replaceRolePermissionsWithOperation(ctx, role, permissionCodes, entity.AuthzPolicyOpApply, "")
	return err
}

func (s *AuthorizationService) replaceRolePermissionsWithOperation(ctx context.Context, role entity.Role, permissionCodes []string, operation entity.AuthzPolicyOperationType, targetOrgID string) (string, error) {
	if s.policyRepo == nil {
		return "", fmt.Errorf("policy repository not configured")
	}
	if !isValidAuthzRole(role) {
		return "", fmt.Errorf("invalid role: %s", role)
	}
	normalizedCodes, err := s.normalizePermissionCodes(permissionCodes)
	if err != nil {
		return "", err
	}
	targetOrgID = strings.TrimSpace(targetOrgID)
	beforeCodes, err := s.policyRepo.ListRolePermissionCodesForOrg(ctx, role, targetOrgID)
	if err != nil {
		return "", err
	}
	if err := s.policyRepo.ReplaceRolePermissionsForOrg(ctx, role, targetOrgID, normalizedCodes); err != nil {
		return "", err
	}
	logCtx := context.WithValue(ctx, "authz_scope_org_id", targetOrgID)
	logID := s.recordRolePermissionChangeWithOperation(logCtx, operation, role, beforeCodes, normalizedCodes)
	s.RefreshPolicies()
	return logID, nil
}

func (s *AuthorizationService) ListPolicyRules(ctx context.Context) ([]repository.PolicyRule, error) {
	if s.policyRepo == nil {
		return nil, fmt.Errorf("policy repository not configured")
	}
	return s.policyRepo.ListAllPolicyRules(ctx)
}

func (s *AuthorizationService) ReplacePolicyRules(ctx context.Context, permissionCode string, rules []repository.PolicyRule) error {
	_, err := s.replacePolicyRulesWithOperation(ctx, permissionCode, rules, entity.AuthzPolicyOpApply, "")
	return err
}

func (s *AuthorizationService) replacePolicyRulesWithOperation(ctx context.Context, permissionCode string, rules []repository.PolicyRule, operation entity.AuthzPolicyOperationType, targetOrgID string) (string, error) {
	if s.policyRepo == nil {
		return "", fmt.Errorf("policy repository not configured")
	}
	permissionCode = strings.TrimSpace(permissionCode)
	if permissionCode == "" {
		return "", fmt.Errorf("permission_code is required")
	}
	permissionCode = strings.TrimSpace(permissionCode)
	if _, ok := supportedPermissionCodes[permissionCode]; !ok {
		return "", fmt.Errorf("unknown permission_code: %s", permissionCode)
	}
	normalizedRules, err := s.normalizeAndValidateRules(permissionCode, rules)
	if err != nil {
		return "", err
	}
	targetOrgID = strings.TrimSpace(targetOrgID)
	beforeRules, err := s.policyRepo.ListPolicyRulesForOrg(ctx, permissionCode, targetOrgID)
	if err != nil {
		return "", err
	}
	if err := s.policyRepo.ReplacePolicyRulesForOrg(ctx, permissionCode, targetOrgID, normalizedRules); err != nil {
		return "", err
	}
	logCtx := context.WithValue(ctx, "authz_scope_org_id", targetOrgID)
	logID := s.recordPolicyRulesChangeWithOperation(logCtx, operation, permissionCode, beforeRules, normalizedRules)
	s.RefreshPolicies()
	return logID, nil
}

func (s *AuthorizationService) ListDecisions(ctx context.Context, query *entity.AuthzDecisionQuery) (*repository.AuthzDecisionPaginatedResult, error) {
	if s.policyRepo == nil {
		return nil, fmt.Errorf("policy repository not configured")
	}
	return s.policyRepo.ListDecisions(ctx, query)
}

func (s *AuthorizationService) ListPolicyChanges(ctx context.Context, query *entity.AuthzPolicyChangeQuery) (*repository.AuthzPolicyChangePaginatedResult, error) {
	if s.policyRepo == nil {
		return nil, fmt.Errorf("policy repository not configured")
	}
	return s.policyRepo.ListPolicyChanges(ctx, query)
}

func (s *AuthorizationService) PreviewReplaceRolePermissions(ctx context.Context, role entity.Role, permissionCodes []string) (*RolePermissionsPreviewResult, error) {
	return s.previewReplaceRolePermissionsForOrg(ctx, role, permissionCodes, "")
}

func (s *AuthorizationService) previewReplaceRolePermissionsForOrg(ctx context.Context, role entity.Role, permissionCodes []string, orgID string) (*RolePermissionsPreviewResult, error) {
	if s.policyRepo == nil {
		return nil, fmt.Errorf("policy repository not configured")
	}
	if !isValidAuthzRole(role) {
		return nil, fmt.Errorf("invalid role: %s", role)
	}
	after, err := s.normalizePermissionCodes(permissionCodes)
	if err != nil {
		return nil, err
	}
	before, err := s.policyRepo.ListRolePermissionCodesForOrg(ctx, role, strings.TrimSpace(orgID))
	if err != nil {
		return nil, err
	}
	beforeNorm, err := s.normalizePermissionCodes(before)
	if err != nil {
		beforeNorm = before
	}

	added, removed := diffStringSlices(beforeNorm, after)
	return &RolePermissionsPreviewResult{
		Role:    role,
		Before:  beforeNorm,
		After:   after,
		Added:   added,
		Removed: removed,
		Changed: len(added) > 0 || len(removed) > 0,
	}, nil
}

func (s *AuthorizationService) PreviewReplacePolicyRules(ctx context.Context, permissionCode string, rules []repository.PolicyRule) (*PolicyRulesPreviewResult, error) {
	return s.previewReplacePolicyRulesForOrg(ctx, permissionCode, rules, "")
}

func (s *AuthorizationService) previewReplacePolicyRulesForOrg(ctx context.Context, permissionCode string, rules []repository.PolicyRule, orgID string) (*PolicyRulesPreviewResult, error) {
	if s.policyRepo == nil {
		return nil, fmt.Errorf("policy repository not configured")
	}
	permissionCode = strings.TrimSpace(permissionCode)
	if permissionCode == "" {
		return nil, fmt.Errorf("permission_code is required")
	}
	if _, ok := supportedPermissionCodes[permissionCode]; !ok {
		return nil, fmt.Errorf("unknown permission_code: %s", permissionCode)
	}

	after, err := s.normalizeAndValidateRules(permissionCode, rules)
	if err != nil {
		return nil, err
	}
	before, err := s.policyRepo.ListPolicyRulesForOrg(ctx, permissionCode, strings.TrimSpace(orgID))
	if err != nil {
		return nil, err
	}
	beforeNorm, err := s.normalizeAndValidateRules(permissionCode, before)
	if err != nil {
		beforeNorm = before
	}

	beforeKeys := make([]string, 0, len(beforeNorm))
	for _, rule := range beforeNorm {
		beforeKeys = append(beforeKeys, ruleSignature(rule))
	}
	afterKeys := make([]string, 0, len(after))
	for _, rule := range after {
		afterKeys = append(afterKeys, ruleSignature(rule))
	}
	added, removed := diffStringSlices(beforeKeys, afterKeys)

	return &PolicyRulesPreviewResult{
		PermissionCode: permissionCode,
		Before:         beforeNorm,
		After:          after,
		Added:          added,
		Removed:        removed,
		Changed:        len(added) > 0 || len(removed) > 0,
	}, nil
}

func (s *AuthorizationService) RollbackPolicyChange(ctx context.Context, changeID string) (*AuthzPolicyRollbackResult, error) {
	if s.policyRepo == nil {
		return nil, fmt.Errorf("policy repository not configured")
	}
	changeID = strings.TrimSpace(changeID)
	if changeID == "" {
		return nil, fmt.Errorf("change_id is required")
	}
	log, err := s.policyRepo.FindPolicyChangeByID(ctx, changeID)
	if err != nil {
		return nil, err
	}
	if log.Operation == entity.AuthzPolicyOpRollback {
		return nil, fmt.Errorf("change %s is already a rollback record", changeID)
	}
	alreadyRolledBack, err := s.policyRepo.ExistsRollbackForChange(ctx, changeID)
	if err != nil {
		return nil, err
	}
	if alreadyRolledBack {
		return nil, fmt.Errorf("change %s has already been rolled back", changeID)
	}
	result := &AuthzPolicyRollbackResult{
		ChangeID:   log.ID,
		ChangeType: string(log.ChangeType),
		TargetKey:  log.TargetKey,
		Applied:    false,
	}
	targetKeyBase, targetOrgID := decodeScopedTargetKey(log.TargetKey)
	preview, err := s.PreviewRollbackPolicyChange(ctx, changeID)
	if err != nil {
		return nil, err
	}
	ctxWithRollbackRef := context.WithValue(ctx, "authz_rollback_of_id", log.ID)
	switch log.ChangeType {
	case entity.AuthzPolicyChangeRolePermissions:
		role, ok := parseRoleForService(targetKeyBase)
		if !ok {
			return nil, fmt.Errorf("invalid rollback role target: %s", targetKeyBase)
		}
		targetCodes := make([]string, 0)
		if strings.TrimSpace(log.BeforeJSON) != "" {
			if err := json.Unmarshal([]byte(log.BeforeJSON), &targetCodes); err != nil {
				return nil, fmt.Errorf("invalid role permission rollback payload: %w", err)
			}
		}
		targetCodes, err = s.normalizePermissionCodes(targetCodes)
		if err != nil {
			return nil, err
		}
		currentCodes, err := s.policyRepo.ListRolePermissionCodesForOrg(ctx, role, targetOrgID)
		if err != nil {
			return nil, err
		}
		currentCodes, err = s.normalizePermissionCodes(currentCodes)
		if err != nil {
			return nil, err
		}
		if err := s.policyRepo.ReplaceRolePermissionsForOrg(ctx, role, targetOrgID, targetCodes); err != nil {
			return nil, err
		}
		logCtx := context.WithValue(ctxWithRollbackRef, "authz_scope_org_id", targetOrgID)
		result.ExecutedLogID = s.recordRolePermissionChangeWithOperation(logCtx, entity.AuthzPolicyOpRollback, role, currentCodes, targetCodes)
		s.RefreshPolicies()
		result.Applied = preview.Changed
		return result, nil
	case entity.AuthzPolicyChangePolicyRules:
		permissionCode := strings.TrimSpace(targetKeyBase)
		targetRules := make([]repository.PolicyRule, 0)
		if strings.TrimSpace(log.BeforeJSON) != "" {
			if err := json.Unmarshal([]byte(log.BeforeJSON), &targetRules); err != nil {
				return nil, fmt.Errorf("invalid policy rules rollback payload: %w", err)
			}
		}
		targetRules, err = s.normalizeAndValidateRules(permissionCode, targetRules)
		if err != nil {
			return nil, err
		}
		currentRules, err := s.policyRepo.ListPolicyRulesForOrg(ctx, permissionCode, targetOrgID)
		if err != nil {
			return nil, err
		}
		currentRules, err = s.normalizeAndValidateRules(permissionCode, currentRules)
		if err != nil {
			return nil, err
		}
		if err := s.policyRepo.ReplacePolicyRulesForOrg(ctx, permissionCode, targetOrgID, targetRules); err != nil {
			return nil, err
		}
		logCtx := context.WithValue(ctxWithRollbackRef, "authz_scope_org_id", targetOrgID)
		result.ExecutedLogID = s.recordPolicyRulesChangeWithOperation(logCtx, entity.AuthzPolicyOpRollback, permissionCode, currentRules, targetRules)
		s.RefreshPolicies()
		result.Applied = preview.Changed
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported change_type: %s", log.ChangeType)
	}
}

func (s *AuthorizationService) SubmitRolePermissionsChangeRequest(ctx context.Context, role entity.Role, permissionCodes []string, requestNote string) (*AuthzPolicyChangeRequestSubmitResult, error) {
	return s.SubmitRolePermissionsChangeRequestWithScope(ctx, role, permissionCodes, requestNote, string(entity.AuthzPolicyScopeGlobal), "")
}

func (s *AuthorizationService) SubmitRolePermissionsChangeRequestWithScope(ctx context.Context, role entity.Role, permissionCodes []string, requestNote, scopeType, targetOrgID string) (*AuthzPolicyChangeRequestSubmitResult, error) {
	scope := normalizeAuthzScopeType(scopeType)
	targetOrg := normalizeAuthzTargetOrgID(targetOrgID)
	if scope == entity.AuthzPolicyScopeGlobal {
		targetOrg = ""
	} else if targetOrg == "" {
		return nil, fmt.Errorf("target_org_id is required for org scope")
	}
	preview, err := s.previewReplaceRolePermissionsForOrg(ctx, role, permissionCodes, targetOrg)
	if err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(preview.After)
	previewJSON, _ := json.Marshal(preview)
	req := entity.NewAuthzPolicyChangeRequest()
	req.RequestType = entity.AuthzPolicyRequestRolePermissions
	req.TargetKey = string(role)
	req.ScopeType = scope
	req.TargetOrgID = targetOrg
	req.PayloadJSON = string(payload)
	req.PreviewJSON = string(previewJSON)
	req.RequestNote = strings.TrimSpace(requestNote)
	fillRequestOperatorFromContext(ctx, req)
	if err := s.policyRepo.CreatePolicyChangeRequest(ctx, req); err != nil {
		return nil, err
	}
	s.notifyAuthzRequestSubmitted(ctx, req)
	return &AuthzPolicyChangeRequestSubmitResult{RequestID: req.ID, Status: string(req.Status)}, nil
}

func (s *AuthorizationService) SubmitPolicyRulesChangeRequest(ctx context.Context, permissionCode string, rules []repository.PolicyRule, requestNote string) (*AuthzPolicyChangeRequestSubmitResult, error) {
	return s.SubmitPolicyRulesChangeRequestWithScope(ctx, permissionCode, rules, requestNote, string(entity.AuthzPolicyScopeGlobal), "")
}

func (s *AuthorizationService) SubmitPolicyRulesChangeRequestWithScope(ctx context.Context, permissionCode string, rules []repository.PolicyRule, requestNote, scopeType, targetOrgID string) (*AuthzPolicyChangeRequestSubmitResult, error) {
	scope := normalizeAuthzScopeType(scopeType)
	targetOrg := normalizeAuthzTargetOrgID(targetOrgID)
	if scope == entity.AuthzPolicyScopeGlobal {
		targetOrg = ""
	} else if targetOrg == "" {
		return nil, fmt.Errorf("target_org_id is required for org scope")
	}
	preview, err := s.previewReplacePolicyRulesForOrg(ctx, permissionCode, rules, targetOrg)
	if err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(preview.After)
	previewJSON, _ := json.Marshal(preview)
	req := entity.NewAuthzPolicyChangeRequest()
	req.RequestType = entity.AuthzPolicyRequestPolicyRules
	req.TargetKey = strings.TrimSpace(permissionCode)
	req.ScopeType = scope
	req.TargetOrgID = targetOrg
	req.PayloadJSON = string(payload)
	req.PreviewJSON = string(previewJSON)
	req.RequestNote = strings.TrimSpace(requestNote)
	fillRequestOperatorFromContext(ctx, req)
	if err := s.policyRepo.CreatePolicyChangeRequest(ctx, req); err != nil {
		return nil, err
	}
	s.notifyAuthzRequestSubmitted(ctx, req)
	return &AuthzPolicyChangeRequestSubmitResult{RequestID: req.ID, Status: string(req.Status)}, nil
}

func (s *AuthorizationService) SubmitRollbackChangeRequest(ctx context.Context, changeID, requestNote string) (*AuthzPolicyChangeRequestSubmitResult, error) {
	return s.SubmitRollbackChangeRequestWithScope(ctx, changeID, requestNote, string(entity.AuthzPolicyScopeGlobal), "")
}

func (s *AuthorizationService) SubmitRollbackChangeRequestWithScope(ctx context.Context, changeID, requestNote, scopeType, targetOrgID string) (*AuthzPolicyChangeRequestSubmitResult, error) {
	preview, err := s.PreviewRollbackPolicyChange(ctx, changeID)
	if err != nil {
		return nil, err
	}
	changeLog, err := s.policyRepo.FindPolicyChangeByID(ctx, strings.TrimSpace(changeID))
	if err != nil {
		return nil, err
	}
	_, logTargetOrgID := decodeScopedTargetKey(changeLog.TargetKey)
	expectedScope := entity.AuthzPolicyScopeGlobal
	if strings.TrimSpace(logTargetOrgID) != "" {
		expectedScope = entity.AuthzPolicyScopeOrg
	}
	payload, _ := json.Marshal(ginH{"change_id": strings.TrimSpace(changeID)})
	previewJSON, _ := json.Marshal(preview)
	req := entity.NewAuthzPolicyChangeRequest()
	req.RequestType = entity.AuthzPolicyRequestRollback
	req.TargetKey = strings.TrimSpace(changeID)
	if normalizedScope := normalizeAuthzScopeType(scopeType); strings.TrimSpace(scopeType) != "" && normalizedScope != expectedScope {
		return nil, fmt.Errorf("rollback scope mismatch with target change")
	}
	if normalizedTargetOrg := normalizeAuthzTargetOrgID(targetOrgID); normalizedTargetOrg != "" && normalizedTargetOrg != strings.TrimSpace(logTargetOrgID) {
		return nil, fmt.Errorf("rollback target_org_id mismatch with target change")
	}
	req.ScopeType = expectedScope
	req.TargetOrgID = strings.TrimSpace(logTargetOrgID)
	req.PayloadJSON = string(payload)
	req.PreviewJSON = string(previewJSON)
	req.RequestNote = strings.TrimSpace(requestNote)
	fillRequestOperatorFromContext(ctx, req)
	if err := s.policyRepo.CreatePolicyChangeRequest(ctx, req); err != nil {
		return nil, err
	}
	s.notifyAuthzRequestSubmitted(ctx, req)
	return &AuthzPolicyChangeRequestSubmitResult{RequestID: req.ID, Status: string(req.Status)}, nil
}

func (s *AuthorizationService) ListPolicyChangeRequests(ctx context.Context, query *entity.AuthzPolicyChangeRequestQuery) (*repository.AuthzPolicyChangeRequestPaginatedResult, error) {
	if s.policyRepo == nil {
		return nil, fmt.Errorf("policy repository not configured")
	}
	return s.policyRepo.ListPolicyChangeRequests(ctx, query)
}

func (s *AuthorizationService) ReviewPolicyChangeRequest(ctx context.Context, requestID string, approve bool, reviewNote string) (*AuthzPolicyChangeReviewResult, error) {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return nil, fmt.Errorf("request_id is required")
	}
	req, err := s.policyRepo.FindPolicyChangeRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req.Status != entity.AuthzPolicyRequestPending {
		return nil, fmt.Errorf("request is not pending")
	}
	if requesterID := strings.TrimSpace(req.RequestedBy); requesterID != "" {
		if reviewerID, ok := ctx.Value("authz_operator_id").(string); ok && strings.TrimSpace(reviewerID) == requesterID {
			return nil, fmt.Errorf("requester cannot review own request")
		}
	}
	if err := s.validatePolicyRequestReviewScope(ctx, req); err != nil {
		return nil, err
	}

	now := time.Now()
	req.ReviewedAt = &now
	req.ReviewNote = strings.TrimSpace(reviewNote)
	if reviewerID, ok := ctx.Value("authz_operator_id").(string); ok {
		req.ReviewedBy = strings.TrimSpace(reviewerID)
	}

	result := &AuthzPolicyChangeReviewResult{
		RequestID: req.ID,
		Status:    string(req.Status),
		Executed:  false,
	}

	if !approve {
		req.Status = entity.AuthzPolicyRequestRejected
		if err := s.policyRepo.UpdatePolicyChangeRequest(ctx, req); err != nil {
			return nil, err
		}
		s.notifyAuthzRequestReviewed(ctx, req, false)
		result.Status = string(req.Status)
		return result, nil
	}

	req.Status = entity.AuthzPolicyRequestApproved
	execCtx := context.WithValue(ctx, "trace_id", fmt.Sprintf("authz_request_execute:%s", req.ID))
	targetOrgID := strings.TrimSpace(req.TargetOrgID)
	if req.ScopeType == entity.AuthzPolicyScopeGlobal {
		targetOrgID = ""
	}
	switch req.RequestType {
	case entity.AuthzPolicyRequestRolePermissions:
		role, ok := parseRoleForService(req.TargetKey)
		if !ok {
			return nil, fmt.Errorf("invalid role target")
		}
		var codes []string
		if err := json.Unmarshal([]byte(req.PayloadJSON), &codes); err != nil {
			return nil, fmt.Errorf("invalid payload for role permissions")
		}
		executedLogID, execErr := s.replaceRolePermissionsWithOperation(execCtx, role, codes, entity.AuthzPolicyOpApply, targetOrgID)
		if execErr != nil {
			return nil, execErr
		}
		req.ExecutedLogID = strings.TrimSpace(executedLogID)
	case entity.AuthzPolicyRequestPolicyRules:
		var rules []repository.PolicyRule
		if err := json.Unmarshal([]byte(req.PayloadJSON), &rules); err != nil {
			return nil, fmt.Errorf("invalid payload for policy rules")
		}
		executedLogID, execErr := s.replacePolicyRulesWithOperation(execCtx, req.TargetKey, rules, entity.AuthzPolicyOpApply, targetOrgID)
		if execErr != nil {
			return nil, execErr
		}
		req.ExecutedLogID = strings.TrimSpace(executedLogID)
	case entity.AuthzPolicyRequestRollback:
		rollbackRes, rollbackErr := s.RollbackPolicyChange(execCtx, req.TargetKey)
		if rollbackErr != nil {
			return nil, rollbackErr
		}
		req.ExecutedLogID = strings.TrimSpace(rollbackRes.ExecutedLogID)
	default:
		return nil, fmt.Errorf("unsupported request type: %s", req.RequestType)
	}

	req.Executed = true
	req.ExecutedAt = &now
	if err := s.policyRepo.UpdatePolicyChangeRequest(ctx, req); err != nil {
		return nil, err
	}
	s.notifyAuthzRequestReviewed(ctx, req, true)
	result.Status = string(req.Status)
	result.Executed = req.Executed
	result.ExecutedID = req.ExecutedLogID
	return result, nil
}

func (s *AuthorizationService) AutoRejectExpiredPolicyChangeRequests(ctx context.Context, expireHours int, reviewedBy string) (int64, error) {
	if s.policyRepo == nil {
		return 0, fmt.Errorf("policy repository not configured")
	}
	if expireHours <= 0 {
		return 0, nil
	}
	cutoff := time.Now().Add(-time.Duration(expireHours) * time.Hour)
	note := fmt.Sprintf("auto rejected: pending request timeout (%d hours)", expireHours)
	updated, err := s.policyRepo.RejectExpiredPolicyChangeRequests(ctx, cutoff, reviewedBy, note)
	if err != nil {
		return 0, err
	}
	if updated > 0 {
		s.notifyAuthzTimeout(ctx, updated, expireHours)
	}
	return updated, nil
}

func (s *AuthorizationService) notifyAuthzRequestSubmitted(ctx context.Context, req *entity.AuthzPolicyChangeRequest) {
	if s.notificationRepo == nil || req == nil {
		return
	}
	notification := entity.NewSystemNotification()
	notification.Category = entity.SystemNotificationCategoryAuthz
	notification.RecipientRole = string(entity.RoleSuperAdmin)
	notification.RelatedType = "authz_policy_change_request"
	notification.RelatedID = req.ID
	notification.Title = "权限策略变更申请待审批"
	notification.Content = fmt.Sprintf("收到新的策略变更申请：type=%s, target=%s", req.RequestType, req.TargetKey)
	if v, ok := ctx.Value("authz_operator_id").(string); ok {
		notification.OperatorID = strings.TrimSpace(v)
	}
	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		logger.Warn("create authz submit notification failed", logger.Err(err))
	}
}

func (s *AuthorizationService) notifyAuthzRequestReviewed(ctx context.Context, req *entity.AuthzPolicyChangeRequest, approved bool) {
	if s.notificationRepo == nil || req == nil {
		return
	}
	recipientID := strings.TrimSpace(req.RequestedBy)
	if recipientID == "" {
		return
	}
	notification := entity.NewSystemNotification()
	notification.Category = entity.SystemNotificationCategoryAuthz
	notification.RecipientID = recipientID
	notification.RelatedType = "authz_policy_change_request"
	notification.RelatedID = req.ID
	if approved {
		notification.Title = "权限策略变更申请已通过"
		notification.Content = fmt.Sprintf("申请已通过并执行：type=%s, target=%s", req.RequestType, req.TargetKey)
	} else {
		notification.Title = "权限策略变更申请已驳回"
		notification.Content = fmt.Sprintf("申请已驳回：type=%s, target=%s", req.RequestType, req.TargetKey)
	}
	if v, ok := ctx.Value("authz_operator_id").(string); ok {
		notification.OperatorID = strings.TrimSpace(v)
	}
	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		logger.Warn("create authz review notification failed", logger.Err(err))
	}
}

func (s *AuthorizationService) notifyAuthzTimeout(ctx context.Context, count int64, expireHours int) {
	if s.notificationRepo == nil || count <= 0 {
		return
	}
	notification := entity.NewSystemNotification()
	notification.Category = entity.SystemNotificationCategoryAuthz
	notification.RecipientRole = string(entity.RoleSuperAdmin)
	notification.RelatedType = "authz_policy_change_request"
	notification.Title = "权限策略申请自动驳回"
	notification.Content = fmt.Sprintf("有 %d 条待审批申请已超时自动驳回（阈值 %d 小时）", count, expireHours)
	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		logger.Warn("create authz timeout notification failed", logger.Err(err))
	}
}

func (s *AuthorizationService) ListAuthzNotifications(ctx context.Context, query *entity.SystemNotificationQuery) (*repository.SystemNotificationPaginatedResult, error) {
	if s.notificationRepo == nil {
		return nil, fmt.Errorf("notification repository not configured")
	}
	return s.notificationRepo.List(ctx, query)
}

func (s *AuthorizationService) MarkAuthzNotificationRead(ctx context.Context, id, userID, userRole string) (bool, error) {
	if s.notificationRepo == nil {
		return false, fmt.Errorf("notification repository not configured")
	}
	return s.notificationRepo.MarkRead(ctx, id, userID, userRole)
}

func (s *AuthorizationService) MarkAllAuthzNotificationsRead(ctx context.Context, userID, userRole string) (int64, error) {
	if s.notificationRepo == nil {
		return 0, fmt.Errorf("notification repository not configured")
	}
	return s.notificationRepo.MarkAllRead(ctx, userID, userRole)
}

func (s *AuthorizationService) CountUnreadAuthzNotifications(ctx context.Context, userID, userRole string) (int64, error) {
	if s.notificationRepo == nil {
		return 0, fmt.Errorf("notification repository not configured")
	}
	return s.notificationRepo.CountUnread(ctx, userID, userRole)
}

type ginH map[string]interface{}

func fillRequestOperatorFromContext(ctx context.Context, req *entity.AuthzPolicyChangeRequest) {
	if req == nil {
		return
	}
	if v, ok := ctx.Value("authz_operator_id").(string); ok {
		req.RequestedBy = strings.TrimSpace(v)
	}
	if v, ok := ctx.Value("authz_operator_role").(string); ok {
		req.RequestedByRole = strings.TrimSpace(v)
	}
	if v, ok := ctx.Value("trace_id").(string); ok {
		req.TraceID = strings.TrimSpace(v)
	}
}

func normalizeAuthzScopeType(scopeType string) entity.AuthzPolicyScopeType {
	switch entity.AuthzPolicyScopeType(strings.ToLower(strings.TrimSpace(scopeType))) {
	case entity.AuthzPolicyScopeOrg:
		return entity.AuthzPolicyScopeOrg
	default:
		return entity.AuthzPolicyScopeGlobal
	}
}

func normalizeAuthzTargetOrgID(targetOrgID string) string {
	return strings.TrimSpace(targetOrgID)
}

func (s *AuthorizationService) validatePolicyRequestReviewScope(ctx context.Context, req *entity.AuthzPolicyChangeRequest) error {
	scopeType := req.ScopeType
	if scopeType == "" {
		scopeType = entity.AuthzPolicyScopeGlobal
	}
	targetOrgID := strings.TrimSpace(req.TargetOrgID)
	if req.RequestType == entity.AuthzPolicyRequestRollback {
		changeLog, err := s.policyRepo.FindPolicyChangeByID(ctx, strings.TrimSpace(req.TargetKey))
		if err != nil {
			return fmt.Errorf("failed to validate rollback target scope: %w", err)
		}
		_, logTargetOrgID := decodeScopedTargetKey(changeLog.TargetKey)
		if strings.TrimSpace(logTargetOrgID) == "" {
			scopeType = entity.AuthzPolicyScopeGlobal
			targetOrgID = ""
		} else {
			scopeType = entity.AuthzPolicyScopeOrg
			targetOrgID = strings.TrimSpace(logTargetOrgID)
		}
	}

	reviewerRole := entity.Role(strings.TrimSpace(getStringFromContext(ctx, "authz_operator_role")))
	reviewerID := strings.TrimSpace(getStringFromContext(ctx, "authz_operator_id"))
	reviewerOrgID := strings.TrimSpace(getStringFromContext(ctx, "authz_operator_org_id"))

	if scopeType == entity.AuthzPolicyScopeGlobal {
		if reviewerRole != entity.RoleSuperAdmin {
			return fmt.Errorf("global scoped request requires super_admin reviewer")
		}
		return nil
	}

	// 组织级审批：管理员及以上可审批；管理员仅限本组织及下级组织
	if reviewerRole != entity.RoleSuperAdmin && reviewerRole != entity.RoleAdmin {
		return fmt.Errorf("org scoped request requires admin reviewer")
	}
	if reviewerRole == entity.RoleSuperAdmin {
		return nil
	}

	if targetOrgID == "" {
		return fmt.Errorf("org scoped request missing target_org_id")
	}

	operator := &entity.User{
		BaseEntity: entity.BaseEntity{ID: reviewerID},
		Role:       reviewerRole,
		OrgID:      reviewerOrgID,
	}
	allowed, err := canManageTargetOrg(ctx, s.orgRepo, operator, targetOrgID)
	if err != nil {
		return fmt.Errorf("failed to validate org scoped review: %w", err)
	}
	if !allowed {
		return fmt.Errorf("reviewer org scope mismatch")
	}
	return nil
}

func getStringFromContext(ctx context.Context, key string) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(key).(string); ok {
		return v
	}
	return ""
}

func encodeScopedTargetKey(baseKey, orgID string) string {
	baseKey = strings.TrimSpace(baseKey)
	orgID = strings.TrimSpace(orgID)
	if orgID == "" {
		return baseKey
	}
	return baseKey + "||org=" + orgID
}

func decodeScopedTargetKey(targetKey string) (baseKey, orgID string) {
	targetKey = strings.TrimSpace(targetKey)
	parts := strings.SplitN(targetKey, "||org=", 2)
	baseKey = strings.TrimSpace(parts[0])
	if len(parts) == 2 {
		orgID = strings.TrimSpace(parts[1])
	}
	return baseKey, orgID
}

func (s *AuthorizationService) PreviewRollbackPolicyChange(ctx context.Context, changeID string) (*AuthzPolicyRollbackPreviewResult, error) {
	if s.policyRepo == nil {
		return nil, fmt.Errorf("policy repository not configured")
	}
	changeID = strings.TrimSpace(changeID)
	if changeID == "" {
		return nil, fmt.Errorf("change_id is required")
	}
	log, err := s.policyRepo.FindPolicyChangeByID(ctx, changeID)
	if err != nil {
		return nil, err
	}
	result := &AuthzPolicyRollbackPreviewResult{
		ChangeID:   log.ID,
		ChangeType: string(log.ChangeType),
		TargetKey:  log.TargetKey,
		Operation:  string(entity.AuthzPolicyOpRollback),
		Changed:    false,
	}
	targetKeyBase, targetOrgID := decodeScopedTargetKey(log.TargetKey)
	switch log.ChangeType {
	case entity.AuthzPolicyChangeRolePermissions:
		role, ok := parseRoleForService(targetKeyBase)
		if !ok {
			return nil, fmt.Errorf("invalid rollback role target: %s", targetKeyBase)
		}
		targetCodes := make([]string, 0)
		if strings.TrimSpace(log.BeforeJSON) != "" {
			if err := json.Unmarshal([]byte(log.BeforeJSON), &targetCodes); err != nil {
				return nil, fmt.Errorf("invalid role permission rollback payload: %w", err)
			}
		}
		preview, err := s.previewReplaceRolePermissionsForOrg(ctx, role, targetCodes, targetOrgID)
		if err != nil {
			return nil, err
		}
		result.Changed = preview.Changed
		result.Impact = preview
		return result, nil
	case entity.AuthzPolicyChangePolicyRules:
		permissionCode := strings.TrimSpace(targetKeyBase)
		targetRules := make([]repository.PolicyRule, 0)
		if strings.TrimSpace(log.BeforeJSON) != "" {
			if err := json.Unmarshal([]byte(log.BeforeJSON), &targetRules); err != nil {
				return nil, fmt.Errorf("invalid policy rules rollback payload: %w", err)
			}
		}
		preview, err := s.previewReplacePolicyRulesForOrg(ctx, permissionCode, targetRules, targetOrgID)
		if err != nil {
			return nil, err
		}
		result.Changed = preview.Changed
		result.Impact = preview
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported change_type: %s", log.ChangeType)
	}
}

func parseRoleForService(input string) (entity.Role, bool) {
	role := entity.Role(strings.ToLower(strings.TrimSpace(input)))
	switch role {
	case entity.RoleSuperAdmin, entity.RoleAdmin, entity.RoleManager, entity.RoleVolunteer:
		return role, true
	default:
		return "", false
	}
}

func diffStringSlices(before, after []string) (added []string, removed []string) {
	beforeSet := make(map[string]struct{}, len(before))
	afterSet := make(map[string]struct{}, len(after))
	for _, item := range before {
		beforeSet[item] = struct{}{}
	}
	for _, item := range after {
		afterSet[item] = struct{}{}
	}
	added = make([]string, 0)
	removed = make([]string, 0)
	for item := range afterSet {
		if _, ok := beforeSet[item]; !ok {
			added = append(added, item)
		}
	}
	for item := range beforeSet {
		if _, ok := afterSet[item]; !ok {
			removed = append(removed, item)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	return added, removed
}

func ruleSignature(rule repository.PolicyRule) string {
	return fmt.Sprintf("%s|%s|%s|%d|%t",
		strings.TrimSpace(rule.ResourceType),
		strings.TrimSpace(strings.ToUpper(rule.ScopeRule)),
		strings.TrimSpace(strings.ToLower(rule.Effect)),
		rule.Priority,
		rule.Enabled,
	)
}

func (s *AuthorizationService) normalizePermissionCodes(permissionCodes []string) ([]string, error) {
	unique := make(map[string]struct{}, len(permissionCodes))
	normalized := make([]string, 0, len(permissionCodes))
	for _, code := range permissionCodes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		if _, ok := supportedPermissionCodes[code]; !ok {
			return nil, fmt.Errorf("unknown permission code: %s", code)
		}
		if _, exists := unique[code]; exists {
			continue
		}
		unique[code] = struct{}{}
		normalized = append(normalized, code)
	}
	sort.Strings(normalized)
	return normalized, nil
}

func (s *AuthorizationService) normalizeAndValidateRules(permissionCode string, rules []repository.PolicyRule) ([]repository.PolicyRule, error) {
	normalized := make([]repository.PolicyRule, 0, len(rules))
	type conflictState struct {
		effect   string
		priority int
	}
	seen := make(map[string]conflictState, len(rules))
	for _, item := range rules {
		resourceType := strings.TrimSpace(item.ResourceType)
		if resourceType == "" {
			resourceType = "*"
		}
		scope := strings.TrimSpace(strings.ToUpper(item.ScopeRule))
		if scope == "" {
			return nil, fmt.Errorf("scope_rule is required")
		}
		if _, ok := supportedScopeRules[scope]; !ok {
			return nil, fmt.Errorf("unsupported scope_rule: %s", scope)
		}
		effect := strings.TrimSpace(strings.ToLower(item.Effect))
		if effect == "" {
			effect = "allow"
		}
		if effect != "allow" && effect != "deny" {
			return nil, fmt.Errorf("invalid effect: %s", item.Effect)
		}
		priority := item.Priority
		if priority == 0 {
			priority = 100
		}
		key := resourceType + "|" + scope + "|" + fmt.Sprintf("%d", priority)
		if prev, exists := seen[key]; exists && prev.effect != effect {
			return nil, fmt.Errorf("policy conflict at resource=%s scope=%s priority=%d: both allow and deny", resourceType, scope, priority)
		}
		seen[key] = conflictState{effect: effect, priority: priority}
		normalized = append(normalized, repository.PolicyRule{
			PermissionCode: permissionCode,
			ResourceType:   resourceType,
			ScopeRule:      scope,
			Effect:         effect,
			Priority:       priority,
			Enabled:        item.Enabled,
		})
	}
	sort.SliceStable(normalized, func(i, j int) bool {
		if normalized[i].Priority != normalized[j].Priority {
			return normalized[i].Priority < normalized[j].Priority
		}
		if normalized[i].ResourceType != normalized[j].ResourceType {
			return normalized[i].ResourceType < normalized[j].ResourceType
		}
		if normalized[i].ScopeRule != normalized[j].ScopeRule {
			return normalized[i].ScopeRule < normalized[j].ScopeRule
		}
		return normalized[i].Effect < normalized[j].Effect
	})
	return normalized, nil
}

func (s *AuthorizationService) recordRolePermissionChange(ctx context.Context, role entity.Role, before, after []string) {
	s.recordRolePermissionChangeWithOperation(ctx, entity.AuthzPolicyOpApply, role, before, after)
}

func (s *AuthorizationService) recordRolePermissionChangeWithOperation(ctx context.Context, operation entity.AuthzPolicyOperationType, role entity.Role, before, after []string) string {
	if s.policyRepo == nil {
		return ""
	}
	beforeNorm, err := s.normalizePermissionCodes(before)
	if err != nil {
		beforeNorm = before
	}
	afterNorm, err := s.normalizePermissionCodes(after)
	if err != nil {
		afterNorm = after
	}
	beforeBytes, _ := json.Marshal(beforeNorm)
	afterBytes, _ := json.Marshal(afterNorm)
	if string(beforeBytes) == string(afterBytes) {
		return ""
	}
	log := entity.NewAuthzPolicyChangeLog()
	log.Operation = operation
	log.ChangeType = entity.AuthzPolicyChangeRolePermissions
	log.TargetKey = encodeScopedTargetKey(string(role), getStringFromContext(ctx, "authz_scope_org_id"))
	log.BeforeJSON = string(beforeBytes)
	log.AfterJSON = string(afterBytes)
	if v, ok := ctx.Value("authz_operator_id").(string); ok {
		log.OperatorID = strings.TrimSpace(v)
	}
	if v, ok := ctx.Value("authz_operator_role").(string); ok {
		log.OperatorRole = strings.TrimSpace(v)
	}
	if v, ok := ctx.Value("trace_id").(string); ok {
		log.TraceID = strings.TrimSpace(v)
	}
	if v, ok := ctx.Value("authz_rollback_of_id").(string); ok {
		log.RollbackOfID = strings.TrimSpace(v)
	}
	if err := s.policyRepo.CreatePolicyChange(ctx, log); err != nil {
		logger.Warn("persist role permission change failed", logger.String("role", string(role)), logger.Err(err))
		return ""
	}
	return log.ID
}

func (s *AuthorizationService) recordPolicyRulesChange(ctx context.Context, permissionCode string, before, after []repository.PolicyRule) {
	s.recordPolicyRulesChangeWithOperation(ctx, entity.AuthzPolicyOpApply, permissionCode, before, after)
}

func (s *AuthorizationService) recordPolicyRulesChangeWithOperation(ctx context.Context, operation entity.AuthzPolicyOperationType, permissionCode string, before, after []repository.PolicyRule) string {
	if s.policyRepo == nil {
		return ""
	}
	beforeBytes, _ := json.Marshal(before)
	afterBytes, _ := json.Marshal(after)
	if string(beforeBytes) == string(afterBytes) {
		return ""
	}
	log := entity.NewAuthzPolicyChangeLog()
	log.Operation = operation
	log.ChangeType = entity.AuthzPolicyChangePolicyRules
	log.TargetKey = encodeScopedTargetKey(permissionCode, getStringFromContext(ctx, "authz_scope_org_id"))
	log.BeforeJSON = string(beforeBytes)
	log.AfterJSON = string(afterBytes)
	if v, ok := ctx.Value("authz_operator_id").(string); ok {
		log.OperatorID = strings.TrimSpace(v)
	}
	if v, ok := ctx.Value("authz_operator_role").(string); ok {
		log.OperatorRole = strings.TrimSpace(v)
	}
	if v, ok := ctx.Value("trace_id").(string); ok {
		log.TraceID = strings.TrimSpace(v)
	}
	if v, ok := ctx.Value("authz_rollback_of_id").(string); ok {
		log.RollbackOfID = strings.TrimSpace(v)
	}
	if err := s.policyRepo.CreatePolicyChange(ctx, log); err != nil {
		logger.Warn("persist policy rules change failed", logger.String("permission_code", permissionCode), logger.Err(err))
		return ""
	}
	return log.ID
}

func (s *AuthorizationService) recordDecision(ctx context.Context, action, resourceType, resourceID string, operator *entity.User, decision AuthzDecision) {
	if strings.TrimSpace(action) == "" {
		return
	}
	if !decision.Allowed && strings.TrimSpace(string(decision.Reason)) == "" {
		return
	}

	s.logDecision(action, operator, resourceID, decision)
	if s.policyRepo == nil {
		return
	}

	log := entity.NewAuthzDecisionLog()
	log.Action = action
	log.ResourceType = resourceType
	log.ResourceID = resourceID
	log.Allowed = decision.Allowed
	log.Reason = string(decision.Reason)
	if operator != nil {
		log.OperatorID = operator.ID
		log.OperatorRole = string(operator.Role)
		log.OperatorOrgID = operator.OrgID
	}
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		log.TraceID = strings.TrimSpace(traceID)
	}

	if err := s.policyRepo.CreateDecision(ctx, log); err != nil {
		logger.Warn("persist authz decision failed", logger.String("action", action), logger.Err(err))
	}
}

func (s *AuthorizationService) logDecision(action string, operator *entity.User, resourceID string, decision AuthzDecision) {
	operatorID := ""
	operatorRole := ""
	if operator != nil {
		operatorID = operator.ID
		operatorRole = string(operator.Role)
	}
	logger.Info("authz decision",
		logger.String("action", action),
		logger.String("operator_id", operatorID),
		logger.String("operator_role", operatorRole),
		logger.String("resource_id", resourceID),
		logger.Bool("allowed", decision.Allowed),
		logger.String("reason", string(decision.Reason)),
	)
}

func allow() AuthzDecision {
	return AuthzDecision{Allowed: true, Reason: AuthzAllow}
}

func deny(reason AuthzReason) AuthzDecision {
	return AuthzDecision{Allowed: false, Reason: reason}
}

func canRoleDoFrom(actions map[authzAction]struct{}, action authzAction) bool {
	if len(actions) == 0 {
		return false
	}
	_, ok := actions[action]
	return ok
}

func (s *AuthorizationService) getRoleActions(ctx context.Context, role entity.Role, orgID string) map[authzAction]struct{} {
	cacheKey := string(role) + "::" + strings.TrimSpace(orgID)
	if cached, ok := s.roleActions.Load(cacheKey); ok {
		if actionSet, ok := cached.(map[authzAction]struct{}); ok {
			return actionSet
		}
	}

	// 数据库策略优先
	if s.policyRepo != nil {
		codes, err := s.policyRepo.ListRolePermissionCodesForOrg(ctx, role, strings.TrimSpace(orgID))
		if err != nil {
			logger.Warn("load role permissions failed, fallback to builtin map", logger.String("role", string(role)), logger.Err(err))
		} else {
			actionSet := make(map[authzAction]struct{}, len(codes))
			for _, code := range codes {
				actionSet[authzAction(code)] = struct{}{}
			}
			s.roleActions.Store(cacheKey, actionSet)
			return actionSet
		}
	}

	// 兜底：内置映射
	fallback, ok := roleActions[role]
	if !ok {
		return map[authzAction]struct{}{}
	}
	clone := make(map[authzAction]struct{}, len(fallback))
	for action := range fallback {
		clone[action] = struct{}{}
	}
	s.roleActions.Store(cacheKey, clone)
	return clone
}

func (s *AuthorizationService) getPolicyRules(ctx context.Context, permissionCode, orgID string) []repository.PolicyRule {
	if permissionCode == "" {
		return nil
	}
	cacheKey := permissionCode + "::" + strings.TrimSpace(orgID)
	if cached, ok := s.policyRules.Load(cacheKey); ok {
		if rules, ok := cached.([]repository.PolicyRule); ok {
			return rules
		}
	}
	if s.policyRepo == nil {
		return nil
	}
	rules, err := s.policyRepo.ListPolicyRulesForOrg(ctx, permissionCode, strings.TrimSpace(orgID))
	if err != nil {
		logger.Warn("load policy rules failed", logger.String("permission_code", permissionCode), logger.Err(err))
		return nil
	}
	s.policyRules.Store(cacheKey, rules)
	return rules
}

func (s *AuthorizationService) evaluatePolicy(ctx context.Context, permissionCode, resourceType string, candidates map[string]bool, orgID string) *AuthzDecision {
	rules := s.getPolicyRules(ctx, permissionCode, orgID)
	if len(rules) == 0 {
		return nil
	}

	allowMatched := false
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		if strings.TrimSpace(rule.ResourceType) != "" &&
			rule.ResourceType != "*" &&
			!strings.EqualFold(rule.ResourceType, resourceType) {
			continue
		}
		scope := strings.ToUpper(strings.TrimSpace(rule.ScopeRule))
		if scope == "" {
			continue
		}
		if !candidates[scope] {
			continue
		}

		if strings.EqualFold(rule.Effect, "deny") {
			denyDecision := deny(AuthzDenyPolicy)
			return &denyDecision
		}
		if strings.EqualFold(rule.Effect, "allow") {
			allowMatched = true
		}
	}

	if allowMatched {
		allowDecision := allow()
		return &allowDecision
	}

	denyDecision := deny(AuthzDenyPolicy)
	return &denyDecision
}

// CanCreateUserInOrg 用户创建（目标组织范围）
func (s *AuthorizationService) CanCreateUserInOrg(ctx context.Context, operator *entity.User, targetOrgID string) (decision AuthzDecision, err error) {
	defer func() {
		s.recordDecision(ctx, string(actionUserCreate), "user", targetOrgID, operator, decision)
	}()
	if operator == nil {
		return deny(AuthzDenyUnauthenticated), nil
	}
	if !canRoleDoFrom(s.getRoleActions(ctx, operator.Role, operator.OrgID), actionUserCreate) {
		return deny(AuthzDenyRoleLevel), nil
	}
	if operator.IsSuperAdmin() {
		return allow(), nil
	}

	policyDecision := s.evaluatePolicy(ctx, string(actionUserCreate), "user", map[string]bool{
		"GLOBAL":         operator.IsSuperAdmin(),
		"ORG_DESCENDANT": true, // 组织范围在后续 canManageTargetOrg 严格判定
	}, operator.OrgID)
	if policyDecision != nil && !policyDecision.Allowed {
		return *policyDecision, nil
	}

	ok, err := canManageTargetOrg(ctx, s.orgRepo, operator, targetOrgID)
	if err != nil {
		return AuthzDecision{}, err
	}
	if !ok {
		return deny(AuthzDenyOrgScope), nil
	}
	return allow(), nil
}

// CanManageOrg 组织范围通用校验（manager+ 且目标组织在可管理范围内）
// 注意：不再复用 user:create，避免 manager 按 org 筛选用户时被误拒
func (s *AuthorizationService) CanManageOrg(ctx context.Context, operator *entity.User, targetOrgID string) (decision AuthzDecision, err error) {
	defer func() {
		s.recordDecision(ctx, "org:manage", "organization", targetOrgID, operator, decision)
	}()
	if operator == nil {
		return deny(AuthzDenyUnauthenticated), nil
	}
	if operator.IsSuperAdmin() {
		return allow(), nil
	}
	if !entity.HasRole(operator.Role, entity.RoleManager) {
		return deny(AuthzDenyRoleLevel), nil
	}
	ok, err := canManageTargetOrg(ctx, s.orgRepo, operator, targetOrgID)
	if err != nil {
		return AuthzDecision{}, err
	}
	if !ok {
		return deny(AuthzDenyOrgScope), nil
	}
	return allow(), nil
}

// CanModifyUser 用户修改（角色+组织+对象）
func (s *AuthorizationService) CanModifyUser(ctx context.Context, operator, target *entity.User) (decision AuthzDecision, err error) {
	targetID := ""
	if target != nil {
		targetID = target.ID
	}
	defer func() {
		s.recordDecision(ctx, string(actionUserModify), "user", targetID, operator, decision)
	}()
	if operator == nil {
		return deny(AuthzDenyUnauthenticated), nil
	}
	if !canRoleDoFrom(s.getRoleActions(ctx, operator.Role, operator.OrgID), actionUserModify) {
		return deny(AuthzDenyRoleLevel), nil
	}
	if target == nil {
		return deny(AuthzDenyOwnerOnly), nil
	}
	if operator.IsSuperAdmin() {
		return allow(), nil
	}
	if target.IsSuperAdmin() {
		return deny(AuthzDenyTargetSuperAdmin), nil
	}
	if operator.IsAdmin() {
		policyDecision := s.evaluatePolicy(ctx, string(actionUserModify), "user", map[string]bool{
			"GLOBAL":         operator.IsSuperAdmin(),
			"SELF":           operator.ID == target.ID,
			"ORG_DESCENDANT": true,
		}, operator.OrgID)
		if policyDecision != nil && !policyDecision.Allowed {
			return *policyDecision, nil
		}
		ok, err := canManageTargetOrg(ctx, s.orgRepo, operator, target.OrgID)
		if err != nil {
			return AuthzDecision{}, err
		}
		if !ok {
			return deny(AuthzDenyOrgScope), nil
		}
		return allow(), nil
	}
	if operator.Role == entity.RoleManager {
		if target.Role != entity.RoleVolunteer {
			return deny(AuthzDenyRoleLevel), nil
		}
		policyDecision := s.evaluatePolicy(ctx, string(actionUserModify), "user", map[string]bool{
			"GLOBAL":         operator.IsSuperAdmin(),
			"SELF":           operator.ID == target.ID,
			"ORG_DESCENDANT": true,
		}, operator.OrgID)
		if policyDecision != nil && !policyDecision.Allowed {
			return *policyDecision, nil
		}
		ok, err := canManageTargetOrg(ctx, s.orgRepo, operator, target.OrgID)
		if err != nil {
			return AuthzDecision{}, err
		}
		if !ok {
			return deny(AuthzDenyOrgScope), nil
		}
		return allow(), nil
	}
	if operator.ID == target.ID {
		policyDecision := s.evaluatePolicy(ctx, string(actionUserModify), "user", map[string]bool{
			"SELF": true,
		}, operator.OrgID)
		if policyDecision != nil {
			return *policyDecision, nil
		}
		return allow(), nil
	}
	return deny(AuthzDenyOwnerOnly), nil
}

// CanViewUser 查看用户详情
func (s *AuthorizationService) CanViewUser(ctx context.Context, operator, target *entity.User) (decision AuthzDecision, err error) {
	targetID := ""
	if target != nil {
		targetID = target.ID
	}
	defer func() {
		s.recordDecision(ctx, string(actionUserView), "user", targetID, operator, decision)
	}()
	if operator == nil {
		return deny(AuthzDenyUnauthenticated), nil
	}
	if !canRoleDoFrom(s.getRoleActions(ctx, operator.Role, operator.OrgID), actionUserView) {
		return deny(AuthzDenyRoleLevel), nil
	}
	if target == nil {
		return deny(AuthzDenyOwnerOnly), nil
	}
	if operator.IsSuperAdmin() {
		return allow(), nil
	}
	if operator.ID == target.ID {
		policyDecision := s.evaluatePolicy(ctx, string(actionUserView), "user", map[string]bool{
			"SELF": true,
		}, operator.OrgID)
		if policyDecision != nil {
			return *policyDecision, nil
		}
		return allow(), nil
	}
	ok, err := canManageTargetOrg(ctx, s.orgRepo, operator, target.OrgID)
	if err != nil {
		return AuthzDecision{}, err
	}
	if !ok {
		return deny(AuthzDenyOrgScope), nil
	}
	policyDecision := s.evaluatePolicy(ctx, string(actionUserView), "user", map[string]bool{
		"ORG_DESCENDANT": true,
		"GLOBAL":         operator.IsSuperAdmin(),
	}, operator.OrgID)
	if policyDecision != nil {
		return *policyDecision, nil
	}
	return allow(), nil
}

// CanManageTask 管理任务（删除/分配/取消/审核）
func (s *AuthorizationService) CanManageTask(ctx context.Context, operator *entity.User, task *entity.Task) (decision AuthzDecision, err error) {
	taskID := ""
	if task != nil {
		taskID = task.ID
	}
	defer func() {
		s.recordDecision(ctx, string(actionTaskManage), "task", taskID, operator, decision)
	}()
	if operator == nil {
		return deny(AuthzDenyUnauthenticated), nil
	}
	if !canRoleDoFrom(s.getRoleActions(ctx, operator.Role, operator.OrgID), actionTaskManage) {
		return deny(AuthzDenyRoleLevel), nil
	}
	if task == nil {
		return deny(AuthzDenyOwnerOnly), nil
	}
	if operator.IsSuperAdmin() {
		return allow(), nil
	}
	if entity.GetRoleLevel(operator.Role) < entity.RoleLevelManager {
		return deny(AuthzDenyRoleLevel), nil
	}
	ok, err := canManageTargetOrg(ctx, s.orgRepo, operator, task.OrgID)
	if err != nil {
		return AuthzDecision{}, err
	}
	if !ok {
		return deny(AuthzDenyOrgScope), nil
	}
	policyDecision := s.evaluatePolicy(ctx, string(actionTaskManage), "task", map[string]bool{
		"ORG_DESCENDANT": true,
		"GLOBAL":         operator.IsSuperAdmin(),
	}, operator.OrgID)
	if policyDecision != nil {
		return *policyDecision, nil
	}
	return allow(), nil
}

// CanViewTask 查看任务（创建者/执行人/管理者范围）
func (s *AuthorizationService) CanViewTask(ctx context.Context, operator *entity.User, task *entity.Task) (decision AuthzDecision, err error) {
	taskID := ""
	if task != nil {
		taskID = task.ID
	}
	defer func() {
		s.recordDecision(ctx, string(actionTaskView), "task", taskID, operator, decision)
	}()
	if operator == nil {
		return deny(AuthzDenyUnauthenticated), nil
	}
	if !canRoleDoFrom(s.getRoleActions(ctx, operator.Role, operator.OrgID), actionTaskView) {
		return deny(AuthzDenyRoleLevel), nil
	}
	if task == nil {
		return deny(AuthzDenyOwnerOnly), nil
	}
	if operator.IsSuperAdmin() {
		return allow(), nil
	}
	if operator.ID == task.CreatorID {
		policyDecision := s.evaluatePolicy(ctx, string(actionTaskView), "task", map[string]bool{
			"CREATOR": true,
			"OWNER":   true,
		}, operator.OrgID)
		if policyDecision != nil {
			return *policyDecision, nil
		}
		return allow(), nil
	}
	if task.AssigneeID != nil && operator.ID == *task.AssigneeID {
		policyDecision := s.evaluatePolicy(ctx, string(actionTaskView), "task", map[string]bool{
			"ASSIGNEE": true,
		}, operator.OrgID)
		if policyDecision != nil {
			return *policyDecision, nil
		}
		return allow(), nil
	}
	if entity.GetRoleLevel(operator.Role) >= entity.RoleLevelManager {
		ok, err := canManageTargetOrg(ctx, s.orgRepo, operator, task.OrgID)
		if err != nil {
			return AuthzDecision{}, err
		}
		if ok {
			policyDecision := s.evaluatePolicy(ctx, string(actionTaskView), "task", map[string]bool{
				"ORG_DESCENDANT": true,
				"GLOBAL":         operator.IsSuperAdmin(),
			}, operator.OrgID)
			if policyDecision != nil {
				return *policyDecision, nil
			}
			return allow(), nil
		}
		return deny(AuthzDenyOrgScope), nil
	}
	return deny(AuthzDenyOwnerOnly), nil
}

// CanEditTask 编辑任务（创建者/执行人/管理者范围）
func (s *AuthorizationService) CanEditTask(ctx context.Context, operator *entity.User, task *entity.Task) (decision AuthzDecision, err error) {
	taskID := ""
	if task != nil {
		taskID = task.ID
	}
	defer func() {
		s.recordDecision(ctx, string(actionTaskEdit), "task", taskID, operator, decision)
	}()
	if operator == nil {
		return deny(AuthzDenyUnauthenticated), nil
	}
	if !canRoleDoFrom(s.getRoleActions(ctx, operator.Role, operator.OrgID), actionTaskEdit) {
		return deny(AuthzDenyRoleLevel), nil
	}
	if task == nil {
		return deny(AuthzDenyOwnerOnly), nil
	}
	if operator.IsSuperAdmin() {
		return allow(), nil
	}
	if operator.ID == task.CreatorID {
		policyDecision := s.evaluatePolicy(ctx, string(actionTaskEdit), "task", map[string]bool{
			"CREATOR": true,
			"OWNER":   true,
		}, operator.OrgID)
		if policyDecision != nil {
			return *policyDecision, nil
		}
		return allow(), nil
	}
	if task.AssigneeID != nil && operator.ID == *task.AssigneeID {
		policyDecision := s.evaluatePolicy(ctx, string(actionTaskEdit), "task", map[string]bool{
			"ASSIGNEE": true,
		}, operator.OrgID)
		if policyDecision != nil {
			return *policyDecision, nil
		}
		return allow(), nil
	}
	if entity.GetRoleLevel(operator.Role) >= entity.RoleLevelManager {
		ok, orgErr := canManageTargetOrg(ctx, s.orgRepo, operator, task.OrgID)
		if orgErr != nil {
			return AuthzDecision{}, orgErr
		}
		if ok {
			policyDecision := s.evaluatePolicy(ctx, string(actionTaskEdit), "task", map[string]bool{
				"ORG_DESCENDANT": true,
				"GLOBAL":         operator.IsSuperAdmin(),
			}, operator.OrgID)
			if policyDecision != nil {
				return *policyDecision, nil
			}
			return allow(), nil
		}
		return deny(AuthzDenyOrgScope), nil
	}
	return deny(AuthzDenyOwnerOnly), nil
}

// CanOperateTaskExecution 执行任务动作（开始/完成/进度/跟进）执行人或管理者
func (s *AuthorizationService) CanOperateTaskExecution(ctx context.Context, operator *entity.User, task *entity.Task) (decision AuthzDecision, err error) {
	taskID := ""
	if task != nil {
		taskID = task.ID
	}
	defer func() {
		s.recordDecision(ctx, string(actionTaskExecute), "task", taskID, operator, decision)
	}()
	if operator == nil {
		return deny(AuthzDenyUnauthenticated), nil
	}
	if !canRoleDoFrom(s.getRoleActions(ctx, operator.Role, operator.OrgID), actionTaskExecute) {
		return deny(AuthzDenyRoleLevel), nil
	}
	if task == nil {
		return deny(AuthzDenyAssigneeOnly), nil
	}
	if task.AssigneeID != nil && operator.ID == *task.AssigneeID {
		policyDecision := s.evaluatePolicy(ctx, string(actionTaskExecute), "task", map[string]bool{
			"ASSIGNEE": true,
		}, operator.OrgID)
		if policyDecision != nil {
			return *policyDecision, nil
		}
		return allow(), nil
	}
	if entity.GetRoleLevel(operator.Role) >= entity.RoleLevelManager {
		ok, err := canManageTargetOrg(ctx, s.orgRepo, operator, task.OrgID)
		if err != nil {
			return AuthzDecision{}, err
		}
		if ok {
			policyDecision := s.evaluatePolicy(ctx, string(actionTaskExecute), "task", map[string]bool{
				"ORG_DESCENDANT": true,
				"GLOBAL":         operator.IsSuperAdmin(),
			}, operator.OrgID)
			if policyDecision != nil {
				return *policyDecision, nil
			}
			return allow(), nil
		}
		return deny(AuthzDenyOrgScope), nil
	}
	return deny(AuthzDenyAssigneeOnly), nil
}

// CanModifyMissingPerson 修改走失人员（上报者或管理者范围）
func (s *AuthorizationService) CanModifyMissingPerson(ctx context.Context, operator *entity.User, mp *entity.MissingPerson) (decision AuthzDecision, err error) {
	mpID := ""
	if mp != nil {
		mpID = mp.ID
	}
	defer func() {
		s.recordDecision(ctx, string(actionMissingModify), "missing_person", mpID, operator, decision)
	}()
	if operator == nil {
		return deny(AuthzDenyUnauthenticated), nil
	}
	if !canRoleDoFrom(s.getRoleActions(ctx, operator.Role, operator.OrgID), actionMissingModify) {
		return deny(AuthzDenyRoleLevel), nil
	}
	if mp == nil {
		return deny(AuthzDenyOwnerOnly), nil
	}
	if operator.IsSuperAdmin() {
		return allow(), nil
	}
	if mp.ReporterID == operator.ID {
		policyDecision := s.evaluatePolicy(ctx, string(actionMissingModify), "missing_person", map[string]bool{
			"REPORTER": true,
			"OWNER":    true,
		}, operator.OrgID)
		if policyDecision != nil {
			return *policyDecision, nil
		}
		return allow(), nil
	}
	if entity.GetRoleLevel(operator.Role) >= entity.RoleLevelManager {
		ok, err := canManageTargetOrg(ctx, s.orgRepo, operator, mp.OrgID)
		if err != nil {
			return AuthzDecision{}, err
		}
		if ok {
			policyDecision := s.evaluatePolicy(ctx, string(actionMissingModify), "missing_person", map[string]bool{
				"ORG_DESCENDANT": true,
				"GLOBAL":         operator.IsSuperAdmin(),
			}, operator.OrgID)
			if policyDecision != nil {
				return *policyDecision, nil
			}
			return allow(), nil
		}
		return deny(AuthzDenyOrgScope), nil
	}
	return deny(AuthzDenyReporterOnly), nil
}

// CanManageMissingPerson 管理走失人员状态（manager+ 且组织范围）
func (s *AuthorizationService) CanManageMissingPerson(ctx context.Context, operator *entity.User, mp *entity.MissingPerson) (decision AuthzDecision, err error) {
	mpID := ""
	if mp != nil {
		mpID = mp.ID
	}
	defer func() {
		s.recordDecision(ctx, string(actionMissingManage), "missing_person", mpID, operator, decision)
	}()
	if operator == nil {
		return deny(AuthzDenyUnauthenticated), nil
	}
	if !canRoleDoFrom(s.getRoleActions(ctx, operator.Role, operator.OrgID), actionMissingManage) {
		return deny(AuthzDenyRoleLevel), nil
	}
	if mp == nil {
		return deny(AuthzDenyOwnerOnly), nil
	}
	if operator.IsSuperAdmin() {
		return allow(), nil
	}
	if entity.GetRoleLevel(operator.Role) < entity.RoleLevelManager {
		return deny(AuthzDenyRoleLevel), nil
	}
	ok, err := canManageTargetOrg(ctx, s.orgRepo, operator, mp.OrgID)
	if err != nil {
		return AuthzDecision{}, err
	}
	if !ok {
		return deny(AuthzDenyOrgScope), nil
	}
	policyDecision := s.evaluatePolicy(ctx, string(actionMissingManage), "missing_person", map[string]bool{
		"ORG_DESCENDANT": true,
		"GLOBAL":         operator.IsSuperAdmin(),
	}, operator.OrgID)
	if policyDecision != nil {
		return *policyDecision, nil
	}
	return allow(), nil
}

// CanModifyDialect 修改方言（上传者或管理者范围）
func (s *AuthorizationService) CanModifyDialect(ctx context.Context, operator *entity.User, d *entity.Dialect) (decision AuthzDecision, err error) {
	dialectID := ""
	if d != nil {
		dialectID = d.ID
	}
	defer func() {
		s.recordDecision(ctx, string(actionDialectModify), "dialect", dialectID, operator, decision)
	}()
	if operator == nil {
		return deny(AuthzDenyUnauthenticated), nil
	}
	if !canRoleDoFrom(s.getRoleActions(ctx, operator.Role, operator.OrgID), actionDialectModify) {
		return deny(AuthzDenyRoleLevel), nil
	}
	if d == nil {
		return deny(AuthzDenyOwnerOnly), nil
	}
	if operator.IsSuperAdmin() {
		return allow(), nil
	}
	if d.UploaderID == operator.ID {
		policyDecision := s.evaluatePolicy(ctx, string(actionDialectModify), "dialect", map[string]bool{
			"OWNER": true,
		}, operator.OrgID)
		if policyDecision != nil {
			return *policyDecision, nil
		}
		return allow(), nil
	}
	if entity.GetRoleLevel(operator.Role) >= entity.RoleLevelManager {
		ok, err := canManageTargetOrg(ctx, s.orgRepo, operator, d.OrgID)
		if err != nil {
			return AuthzDecision{}, err
		}
		if ok {
			policyDecision := s.evaluatePolicy(ctx, string(actionDialectModify), "dialect", map[string]bool{
				"ORG_DESCENDANT": true,
				"GLOBAL":         operator.IsSuperAdmin(),
			}, operator.OrgID)
			if policyDecision != nil {
				return *policyDecision, nil
			}
			return allow(), nil
		}
		return deny(AuthzDenyOrgScope), nil
	}
	return deny(AuthzDenyOwnerOnly), nil
}

// CanManageDialectStatus 管理方言状态/精选（manager+）
func (s *AuthorizationService) CanManageDialectStatus(ctx context.Context, operator *entity.User, d *entity.Dialect) (decision AuthzDecision, err error) {
	dialectID := ""
	if d != nil {
		dialectID = d.ID
	}
	defer func() {
		s.recordDecision(ctx, string(actionDialectManage), "dialect", dialectID, operator, decision)
	}()
	if operator == nil {
		return deny(AuthzDenyUnauthenticated), nil
	}
	if !canRoleDoFrom(s.getRoleActions(ctx, operator.Role, operator.OrgID), actionDialectManage) {
		return deny(AuthzDenyRoleLevel), nil
	}
	if d == nil {
		return deny(AuthzDenyOwnerOnly), nil
	}
	if operator.IsSuperAdmin() {
		return allow(), nil
	}
	if entity.GetRoleLevel(operator.Role) < entity.RoleLevelManager {
		return deny(AuthzDenyRoleLevel), nil
	}
	ok, err := canManageTargetOrg(ctx, s.orgRepo, operator, d.OrgID)
	if err != nil {
		return AuthzDecision{}, err
	}
	if !ok {
		return deny(AuthzDenyOrgScope), nil
	}
	policyDecision := s.evaluatePolicy(ctx, string(actionDialectManage), "dialect", map[string]bool{
		"ORG_DESCENDANT": true,
		"GLOBAL":         operator.IsSuperAdmin(),
	}, operator.OrgID)
	if policyDecision != nil {
		return *policyDecision, nil
	}
	return allow(), nil
}
