package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthzPolicyRepositoryImpl struct {
	db *gorm.DB
}

func NewAuthzPolicyRepository(db *gorm.DB) repository.AuthzPolicyRepository {
	return &AuthzPolicyRepositoryImpl{db: db}
}

type rolePermissionRecord struct {
	ID             string  `gorm:"column:id"`
	Role           string  `gorm:"column:role"`
	PermissionCode string  `gorm:"column:permission_code"`
	OrgID          *string `gorm:"column:org_id"`
	Effect         string  `gorm:"column:effect"`
	Enabled        bool    `gorm:"column:enabled"`
	Priority       int     `gorm:"column:priority"`
}

func (rolePermissionRecord) TableName() string {
	return "ty_role_permissions"
}

func (r *AuthzPolicyRepositoryImpl) ListRolePermissionCodes(ctx context.Context, role entity.Role) ([]string, error) {
	return r.ListRolePermissionCodesForOrg(ctx, role, "")
}

func (r *AuthzPolicyRepositoryImpl) ListRolePermissionCodesForOrg(ctx context.Context, role entity.Role, orgID string) ([]string, error) {
	scopeChain, err := r.resolveOrgScopeChain(ctx, orgID)
	if err != nil {
		return nil, err
	}
	for _, scopeOrgID := range scopeChain {
		var records []rolePermissionRecord
		db := r.db.WithContext(ctx).
			Model(&rolePermissionRecord{}).
			Where("role = ? AND enabled = ?", string(role), true)
		if strings.TrimSpace(scopeOrgID) == "" {
			db = db.Where("org_id IS NULL")
		} else {
			db = db.Where("org_id = ?", scopeOrgID)
		}
		if err := db.Order("priority ASC, permission_code ASC").Find(&records).Error; err != nil {
			return nil, err
		}
		if len(records) == 0 {
			continue
		}
		codes := make([]string, 0, len(records))
		seen := make(map[string]struct{}, len(records))
		for _, item := range records {
			code := strings.TrimSpace(item.PermissionCode)
			if code == "" {
				continue
			}
			if _, ok := seen[code]; ok {
				continue
			}
			seen[code] = struct{}{}
			codes = append(codes, code)
		}
		return codes, nil
	}

	var records []rolePermissionRecord
	if err := r.db.WithContext(ctx).
		Model(&rolePermissionRecord{}).
		Where("role = ? AND enabled = ? AND org_id IS NULL", string(role), true).
		Order("permission_code ASC").
		Find(&records).Error; err != nil {
		return nil, err
	}

	codes := make([]string, 0, len(records))
	for _, item := range records {
		if item.PermissionCode == "" {
			continue
		}
		codes = append(codes, item.PermissionCode)
	}
	return codes, nil
}

func (r *AuthzPolicyRepositoryImpl) ListAllRolePermissions(ctx context.Context) ([]repository.RolePermission, error) {
	var records []rolePermissionRecord
	if err := r.db.WithContext(ctx).
		Model(&rolePermissionRecord{}).
		Where("enabled = ?", true).
		Order("role ASC, priority ASC, permission_code ASC").
		Find(&records).Error; err != nil {
		return nil, err
	}

	result := make([]repository.RolePermission, 0, len(records))
	for _, item := range records {
		result = append(result, repository.RolePermission{
			ID:             item.ID,
			Role:           entity.Role(item.Role),
			PermissionCode: item.PermissionCode,
			OrgID:          derefString(item.OrgID),
			Effect:         item.Effect,
			Enabled:        item.Enabled,
			Priority:       item.Priority,
		})
	}
	return result, nil
}

func (r *AuthzPolicyRepositoryImpl) ReplaceRolePermissions(ctx context.Context, role entity.Role, permissionCodes []string) error {
	return r.ReplaceRolePermissionsForOrg(ctx, role, "", permissionCodes)
}

func (r *AuthzPolicyRepositoryImpl) ReplaceRolePermissionsForOrg(ctx context.Context, role entity.Role, orgID string, permissionCodes []string) error {
	scopeOrgID := strings.TrimSpace(orgID)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		deleteDB := tx.Where("role = ?", string(role))
		if scopeOrgID == "" {
			deleteDB = deleteDB.Where("org_id IS NULL")
		} else {
			deleteDB = deleteDB.Where("org_id = ?", scopeOrgID)
		}
		if err := deleteDB.Delete(&rolePermissionRecord{}).Error; err != nil {
			return err
		}
		unique := make(map[string]struct{}, len(permissionCodes))
		for _, code := range permissionCodes {
			code = strings.TrimSpace(code)
			if code == "" {
				continue
			}
			if _, exists := unique[code]; exists {
				continue
			}
			unique[code] = struct{}{}
			record := rolePermissionRecord{
				ID:             uuid.New().String(),
				Role:           string(role),
				PermissionCode: code,
				OrgID:          nil,
				Effect:         "allow",
				Enabled:        true,
				Priority:       100,
			}
			if scopeOrgID != "" {
				record.OrgID = ptrString(scopeOrgID)
			}
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

type policyRuleRecord struct {
	ID             string  `gorm:"column:id"`
	PermissionCode string  `gorm:"column:permission_code"`
	OrgID          *string `gorm:"column:org_id"`
	ResourceType   string  `gorm:"column:resource_type"`
	ScopeRule      string  `gorm:"column:scope_rule"`
	Effect         string  `gorm:"column:effect"`
	Priority       int     `gorm:"column:priority"`
	Enabled        bool    `gorm:"column:enabled"`
}

func (policyRuleRecord) TableName() string {
	return "ty_policy_rules"
}

func (r *AuthzPolicyRepositoryImpl) ListPolicyRules(ctx context.Context, permissionCode string) ([]repository.PolicyRule, error) {
	return r.ListPolicyRulesForOrg(ctx, permissionCode, "")
}

func (r *AuthzPolicyRepositoryImpl) ListPolicyRulesForOrg(ctx context.Context, permissionCode, orgID string) ([]repository.PolicyRule, error) {
	scopeChain, err := r.resolveOrgScopeChain(ctx, orgID)
	if err != nil {
		return nil, err
	}
	for _, scopeOrgID := range scopeChain {
		var records []policyRuleRecord
		db := r.db.WithContext(ctx).
			Model(&policyRuleRecord{}).
			Where("permission_code = ? AND enabled = ?", permissionCode, true)
		if strings.TrimSpace(scopeOrgID) == "" {
			db = db.Where("org_id IS NULL")
		} else {
			db = db.Where("org_id = ?", scopeOrgID)
		}
		if err := db.Order("priority ASC").Find(&records).Error; err != nil {
			return nil, err
		}
		if len(records) == 0 {
			continue
		}
		rules := make([]repository.PolicyRule, 0, len(records))
		for _, item := range records {
			rules = append(rules, repository.PolicyRule{
				ID:             item.ID,
				PermissionCode: item.PermissionCode,
				OrgID:          derefString(item.OrgID),
				ResourceType:   item.ResourceType,
				ScopeRule:      item.ScopeRule,
				Effect:         item.Effect,
				Priority:       item.Priority,
				Enabled:        item.Enabled,
			})
		}
		return rules, nil
	}

	var records []policyRuleRecord
	if err := r.db.WithContext(ctx).
		Model(&policyRuleRecord{}).
		Where("permission_code = ? AND enabled = ? AND org_id IS NULL", permissionCode, true).
		Order("priority ASC").
		Find(&records).Error; err != nil {
		return nil, err
	}

	rules := make([]repository.PolicyRule, 0, len(records))
	for _, item := range records {
		rules = append(rules, repository.PolicyRule{
			ID:             item.ID,
			PermissionCode: item.PermissionCode,
			OrgID:          derefString(item.OrgID),
			ResourceType:   item.ResourceType,
			ScopeRule:      item.ScopeRule,
			Effect:         item.Effect,
			Priority:       item.Priority,
			Enabled:        item.Enabled,
		})
	}
	return rules, nil
}

func (r *AuthzPolicyRepositoryImpl) ListAllPolicyRules(ctx context.Context) ([]repository.PolicyRule, error) {
	var records []policyRuleRecord
	if err := r.db.WithContext(ctx).
		Model(&policyRuleRecord{}).
		Where("enabled = ?", true).
		Order("permission_code ASC, priority ASC, resource_type ASC, scope_rule ASC").
		Find(&records).Error; err != nil {
		return nil, err
	}
	rules := make([]repository.PolicyRule, 0, len(records))
	for _, item := range records {
		rules = append(rules, repository.PolicyRule{
			ID:             item.ID,
			PermissionCode: item.PermissionCode,
			OrgID:          derefString(item.OrgID),
			ResourceType:   item.ResourceType,
			ScopeRule:      item.ScopeRule,
			Effect:         item.Effect,
			Priority:       item.Priority,
			Enabled:        item.Enabled,
		})
	}
	return rules, nil
}

func (r *AuthzPolicyRepositoryImpl) ReplacePolicyRules(ctx context.Context, permissionCode string, rules []repository.PolicyRule) error {
	return r.ReplacePolicyRulesForOrg(ctx, permissionCode, "", rules)
}

func (r *AuthzPolicyRepositoryImpl) ReplacePolicyRulesForOrg(ctx context.Context, permissionCode, orgID string, rules []repository.PolicyRule) error {
	scopeOrgID := strings.TrimSpace(orgID)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		deleteDB := tx.Where("permission_code = ?", permissionCode)
		if scopeOrgID == "" {
			deleteDB = deleteDB.Where("org_id IS NULL")
		} else {
			deleteDB = deleteDB.Where("org_id = ?", scopeOrgID)
		}
		if err := deleteDB.Delete(&policyRuleRecord{}).Error; err != nil {
			return err
		}
		for _, item := range rules {
			record := policyRuleRecord{
				ID:             uuid.New().String(),
				PermissionCode: permissionCode,
				OrgID:          nil,
				ResourceType:   strings.TrimSpace(item.ResourceType),
				ScopeRule:      strings.TrimSpace(strings.ToUpper(item.ScopeRule)),
				Effect:         strings.TrimSpace(strings.ToLower(item.Effect)),
				Priority:       item.Priority,
				Enabled:        item.Enabled,
			}
			if record.ResourceType == "" {
				record.ResourceType = "*"
			}
			if record.ScopeRule == "" {
				continue
			}
			if record.Effect == "" {
				record.Effect = "allow"
			}
			if record.Priority == 0 {
				record.Priority = 100
			}
			if scopeOrgID != "" {
				record.OrgID = ptrString(scopeOrgID)
			}
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *AuthzPolicyRepositoryImpl) resolveOrgScopeChain(ctx context.Context, orgID string) ([]string, error) {
	orgID = strings.TrimSpace(orgID)
	chain := make([]string, 0, 6)
	if orgID == "" {
		return append(chain, ""), nil
	}

	visited := map[string]struct{}{}
	current := orgID
	for current != "" {
		if _, ok := visited[current]; ok {
			break
		}
		visited[current] = struct{}{}
		chain = append(chain, current)

		var row struct {
			ID       string  `gorm:"column:id"`
			ParentID *string `gorm:"column:parent_id"`
		}
		err := r.db.WithContext(ctx).Table("ty_organizations").Select("id, parent_id").Where("id = ?", current).Take(&row).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break
			}
			return nil, err
		}
		if row.ParentID == nil {
			break
		}
		current = strings.TrimSpace(*row.ParentID)
	}
	chain = append(chain, "")
	return chain, nil
}

func ptrString(s string) *string {
	v := s
	return &v
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

type authzDecisionRecord struct {
	ID            string    `gorm:"column:id"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
	OperatorID    string    `gorm:"column:operator_id"`
	OperatorRole  string    `gorm:"column:operator_role"`
	OperatorOrgID string    `gorm:"column:operator_org_id"`
	Action        string    `gorm:"column:action"`
	ResourceType  string    `gorm:"column:resource_type"`
	ResourceID    string    `gorm:"column:resource_id"`
	Allowed       bool      `gorm:"column:allowed"`
	Reason        string    `gorm:"column:reason"`
	TraceID       string    `gorm:"column:trace_id"`
}

func (authzDecisionRecord) TableName() string {
	return "ty_authz_decisions"
}

func (r *AuthzPolicyRepositoryImpl) CreateDecision(ctx context.Context, decision *entity.AuthzDecisionLog) error {
	if decision == nil {
		return nil
	}
	record := authzDecisionRecord{
		ID:            decision.ID,
		CreatedAt:     decision.CreatedAt,
		UpdatedAt:     decision.UpdatedAt,
		OperatorID:    decision.OperatorID,
		OperatorRole:  decision.OperatorRole,
		OperatorOrgID: decision.OperatorOrgID,
		Action:        decision.Action,
		ResourceType:  decision.ResourceType,
		ResourceID:    decision.ResourceID,
		Allowed:       decision.Allowed,
		Reason:        decision.Reason,
		TraceID:       decision.TraceID,
	}
	if strings.TrimSpace(record.ID) == "" {
		record.ID = uuid.New().String()
	}
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now()
	}
	if record.UpdatedAt.IsZero() {
		record.UpdatedAt = record.CreatedAt
	}
	return r.db.WithContext(ctx).Create(&record).Error
}

func (r *AuthzPolicyRepositoryImpl) ListDecisions(ctx context.Context, query *entity.AuthzDecisionQuery) (*repository.AuthzDecisionPaginatedResult, error) {
	if query == nil {
		query = entity.NewAuthzDecisionQuery()
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}

	db := r.db.WithContext(ctx).Model(&authzDecisionRecord{})
	if action := strings.TrimSpace(query.Action); action != "" {
		db = db.Where("action = ?", action)
	}
	if query.Allowed != nil {
		db = db.Where("allowed = ?", *query.Allowed)
	}
	if operatorID := strings.TrimSpace(query.OperatorID); operatorID != "" {
		db = db.Where("operator_id = ?", operatorID)
	}
	if resourceType := strings.TrimSpace(query.ResourceType); resourceType != "" {
		db = db.Where("resource_type = ?", resourceType)
	}
	if reason := strings.TrimSpace(query.Reason); reason != "" {
		db = db.Where("reason = ?", reason)
	}
	if query.StartTime != nil {
		db = db.Where("created_at >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("created_at <= ?", *query.EndTime)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	var records []authzDecisionRecord
	offset := (query.Page - 1) * query.PageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&records).Error; err != nil {
		return nil, err
	}

	items := make([]entity.AuthzDecisionLog, 0, len(records))
	for _, item := range records {
		items = append(items, entity.AuthzDecisionLog{
			BaseEntity: entity.BaseEntity{
				ID:        item.ID,
				CreatedAt: item.CreatedAt,
				UpdatedAt: item.UpdatedAt,
			},
			OperatorID:    item.OperatorID,
			OperatorRole:  item.OperatorRole,
			OperatorOrgID: item.OperatorOrgID,
			Action:        item.Action,
			ResourceType:  item.ResourceType,
			ResourceID:    item.ResourceID,
			Allowed:       item.Allowed,
			Reason:        item.Reason,
			TraceID:       item.TraceID,
		})
	}

	return &repository.AuthzDecisionPaginatedResult{
		List:     items,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

type authzPolicyChangeRecord struct {
	ID           string    `gorm:"column:id"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
	OperatorID   *string   `gorm:"column:operator_id"`
	OperatorRole string    `gorm:"column:operator_role"`
	Operation    string    `gorm:"column:operation"`
	ChangeType   string    `gorm:"column:change_type"`
	TargetKey    string    `gorm:"column:target_key"`
	RollbackOfID *string   `gorm:"column:rollback_of_id"`
	BeforeJSON   string    `gorm:"column:before_json"`
	AfterJSON    string    `gorm:"column:after_json"`
	TraceID      string    `gorm:"column:trace_id"`
}

func (authzPolicyChangeRecord) TableName() string {
	return "ty_authz_policy_changes"
}

func (r *AuthzPolicyRepositoryImpl) CreatePolicyChange(ctx context.Context, log *entity.AuthzPolicyChangeLog) error {
	if log == nil {
		return nil
	}
	record := authzPolicyChangeRecord{
		ID:           log.ID,
		CreatedAt:    log.CreatedAt,
		UpdatedAt:    log.UpdatedAt,
		OperatorID:   nullableString(log.OperatorID),
		OperatorRole: log.OperatorRole,
		Operation:    string(log.Operation),
		ChangeType:   string(log.ChangeType),
		TargetKey:    log.TargetKey,
		RollbackOfID: nullableString(log.RollbackOfID),
		BeforeJSON:   log.BeforeJSON,
		AfterJSON:    log.AfterJSON,
		TraceID:      log.TraceID,
	}
	if strings.TrimSpace(record.Operation) == "" {
		record.Operation = string(entity.AuthzPolicyOpApply)
	}
	if strings.TrimSpace(record.ID) == "" {
		record.ID = uuid.New().String()
	}
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now()
	}
	if record.UpdatedAt.IsZero() {
		record.UpdatedAt = record.CreatedAt
	}
	return r.db.WithContext(ctx).Create(&record).Error
}

func (r *AuthzPolicyRepositoryImpl) ListPolicyChanges(ctx context.Context, query *entity.AuthzPolicyChangeQuery) (*repository.AuthzPolicyChangePaginatedResult, error) {
	if query == nil {
		query = entity.NewAuthzPolicyChangeQuery()
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}

	db := r.db.WithContext(ctx).Model(&authzPolicyChangeRecord{})
	if operation := strings.TrimSpace(query.Operation); operation != "" {
		db = db.Where("operation = ?", operation)
	}
	if changeType := strings.TrimSpace(query.ChangeType); changeType != "" {
		db = db.Where("change_type = ?", changeType)
	}
	if targetKey := strings.TrimSpace(query.TargetKey); targetKey != "" {
		db = db.Where("target_key = ?", targetKey)
	}
	if rollbackOfID := strings.TrimSpace(query.RollbackOfID); rollbackOfID != "" {
		db = db.Where("rollback_of_id = ?", rollbackOfID)
	}
	if operatorID := strings.TrimSpace(query.OperatorID); operatorID != "" {
		db = db.Where("operator_id = ?", operatorID)
	}
	if query.StartTime != nil {
		db = db.Where("created_at >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("created_at <= ?", *query.EndTime)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	var records []authzPolicyChangeRecord
	offset := (query.Page - 1) * query.PageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&records).Error; err != nil {
		return nil, err
	}

	items := make([]entity.AuthzPolicyChangeLog, 0, len(records))
	for _, item := range records {
		items = append(items, entity.AuthzPolicyChangeLog{
			BaseEntity: entity.BaseEntity{
				ID:        item.ID,
				CreatedAt: item.CreatedAt,
				UpdatedAt: item.UpdatedAt,
			},
			OperatorID:   derefString(item.OperatorID),
			OperatorRole: item.OperatorRole,
			Operation:    entity.AuthzPolicyOperationType(item.Operation),
			ChangeType:   entity.AuthzPolicyChangeType(item.ChangeType),
			TargetKey:    item.TargetKey,
			RollbackOfID: derefString(item.RollbackOfID),
			BeforeJSON:   item.BeforeJSON,
			AfterJSON:    item.AfterJSON,
			TraceID:      item.TraceID,
		})
	}

	return &repository.AuthzPolicyChangePaginatedResult{
		List:     items,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

func (r *AuthzPolicyRepositoryImpl) FindPolicyChangeByID(ctx context.Context, id string) (*entity.AuthzPolicyChangeLog, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var record authzPolicyChangeRecord
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&record).Error; err != nil {
		return nil, err
	}
	return &entity.AuthzPolicyChangeLog{
		BaseEntity: entity.BaseEntity{
			ID:        record.ID,
			CreatedAt: record.CreatedAt,
			UpdatedAt: record.UpdatedAt,
		},
		OperatorID:   derefString(record.OperatorID),
		OperatorRole: record.OperatorRole,
		Operation:    entity.AuthzPolicyOperationType(record.Operation),
		ChangeType:   entity.AuthzPolicyChangeType(record.ChangeType),
		TargetKey:    record.TargetKey,
		RollbackOfID: derefString(record.RollbackOfID),
		BeforeJSON:   record.BeforeJSON,
		AfterJSON:    record.AfterJSON,
		TraceID:      record.TraceID,
	}, nil
}

func (r *AuthzPolicyRepositoryImpl) ExistsRollbackForChange(ctx context.Context, changeID string) (bool, error) {
	changeID = strings.TrimSpace(changeID)
	if changeID == "" {
		return false, nil
	}
	var count int64
	err := r.db.WithContext(ctx).
		Model(&authzPolicyChangeRecord{}).
		Where("operation = ? AND rollback_of_id = ?", string(entity.AuthzPolicyOpRollback), changeID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

type authzPolicyChangeRequestRecord struct {
	ID              string     `gorm:"column:id"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
	RequestType     string     `gorm:"column:request_type"`
	Status          string     `gorm:"column:status"`
	ScopeType       string     `gorm:"column:scope_type"`
	TargetOrgID     *string    `gorm:"column:target_org_id"`
	TargetKey       string     `gorm:"column:target_key"`
	PayloadJSON     string     `gorm:"column:payload_json"`
	PreviewJSON     string     `gorm:"column:preview_json"`
	RequestNote     string     `gorm:"column:request_note"`
	RequestedBy     *string    `gorm:"column:requested_by"`
	RequestedByRole string     `gorm:"column:requested_by_role"`
	ReviewNote      string     `gorm:"column:review_note"`
	ReviewedBy      *string    `gorm:"column:reviewed_by"`
	ReviewedAt      *time.Time `gorm:"column:reviewed_at"`
	Executed        bool       `gorm:"column:executed"`
	ExecutedAt      *time.Time `gorm:"column:executed_at"`
	ExecutedLogID   *string    `gorm:"column:executed_log_id"`
	TraceID         string     `gorm:"column:trace_id"`
}

func (authzPolicyChangeRequestRecord) TableName() string {
	return "ty_authz_policy_change_requests"
}

func (r *AuthzPolicyRepositoryImpl) CreatePolicyChangeRequest(ctx context.Context, req *entity.AuthzPolicyChangeRequest) error {
	if req == nil {
		return nil
	}
	record := authzPolicyChangeRequestRecord{
		ID:              req.ID,
		CreatedAt:       req.CreatedAt,
		UpdatedAt:       req.UpdatedAt,
		RequestType:     string(req.RequestType),
		Status:          string(req.Status),
		ScopeType:       string(req.ScopeType),
		TargetOrgID:     nullableString(req.TargetOrgID),
		TargetKey:       req.TargetKey,
		PayloadJSON:     req.PayloadJSON,
		PreviewJSON:     req.PreviewJSON,
		RequestNote:     req.RequestNote,
		RequestedBy:     nullableString(req.RequestedBy),
		RequestedByRole: req.RequestedByRole,
		ReviewNote:      req.ReviewNote,
		ReviewedBy:      nullableString(req.ReviewedBy),
		ReviewedAt:      req.ReviewedAt,
		Executed:        req.Executed,
		ExecutedAt:      req.ExecutedAt,
		ExecutedLogID:   nullableString(req.ExecutedLogID),
		TraceID:         req.TraceID,
	}
	if strings.TrimSpace(record.ID) == "" {
		record.ID = uuid.New().String()
	}
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now()
	}
	if record.UpdatedAt.IsZero() {
		record.UpdatedAt = record.CreatedAt
	}
	if strings.TrimSpace(record.Status) == "" {
		record.Status = string(entity.AuthzPolicyRequestPending)
	}
	if strings.TrimSpace(record.ScopeType) == "" {
		record.ScopeType = string(entity.AuthzPolicyScopeGlobal)
	}
	return r.db.WithContext(ctx).Create(&record).Error
}

func (r *AuthzPolicyRepositoryImpl) ListPolicyChangeRequests(ctx context.Context, query *entity.AuthzPolicyChangeRequestQuery) (*repository.AuthzPolicyChangeRequestPaginatedResult, error) {
	if query == nil {
		query = entity.NewAuthzPolicyChangeRequestQuery()
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}

	db := r.db.WithContext(ctx).Model(&authzPolicyChangeRequestRecord{})
	if requestType := strings.TrimSpace(query.RequestType); requestType != "" {
		db = db.Where("request_type = ?", requestType)
	}
	if status := strings.TrimSpace(query.Status); status != "" {
		db = db.Where("status = ?", status)
	}
	if scopeType := strings.TrimSpace(query.ScopeType); scopeType != "" {
		db = db.Where("scope_type = ?", scopeType)
	}
	if requestedBy := strings.TrimSpace(query.RequestedBy); requestedBy != "" {
		db = db.Where("requested_by = ?", requestedBy)
	}
	if targetOrgID := strings.TrimSpace(query.TargetOrgID); targetOrgID != "" {
		db = db.Where("target_org_id = ?", targetOrgID)
	}
	if targetKey := strings.TrimSpace(query.TargetKey); targetKey != "" {
		db = db.Where("target_key = ?", targetKey)
	}
	if query.StartTime != nil {
		db = db.Where("created_at >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("created_at <= ?", *query.EndTime)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}
	offset := (query.Page - 1) * query.PageSize
	var records []authzPolicyChangeRequestRecord
	if err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&records).Error; err != nil {
		return nil, err
	}
	list := make([]entity.AuthzPolicyChangeRequest, 0, len(records))
	for _, record := range records {
		list = append(list, entity.AuthzPolicyChangeRequest{
			BaseEntity: entity.BaseEntity{
				ID:        record.ID,
				CreatedAt: record.CreatedAt,
				UpdatedAt: record.UpdatedAt,
			},
			RequestType:     entity.AuthzPolicyRequestType(record.RequestType),
			Status:          entity.AuthzPolicyRequestStatus(record.Status),
			ScopeType:       entity.AuthzPolicyScopeType(record.ScopeType),
			TargetOrgID:     derefString(record.TargetOrgID),
			TargetKey:       record.TargetKey,
			PayloadJSON:     record.PayloadJSON,
			PreviewJSON:     record.PreviewJSON,
			RequestNote:     record.RequestNote,
			RequestedBy:     derefString(record.RequestedBy),
			RequestedByRole: record.RequestedByRole,
			ReviewNote:      record.ReviewNote,
			ReviewedBy:      derefString(record.ReviewedBy),
			ReviewedAt:      record.ReviewedAt,
			Executed:        record.Executed,
			ExecutedAt:      record.ExecutedAt,
			ExecutedLogID:   derefString(record.ExecutedLogID),
			TraceID:         record.TraceID,
		})
	}
	return &repository.AuthzPolicyChangeRequestPaginatedResult{
		List:     list,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

func (r *AuthzPolicyRepositoryImpl) FindPolicyChangeRequestByID(ctx context.Context, id string) (*entity.AuthzPolicyChangeRequest, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var record authzPolicyChangeRequestRecord
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&record).Error; err != nil {
		return nil, err
	}
	return &entity.AuthzPolicyChangeRequest{
		BaseEntity: entity.BaseEntity{
			ID:        record.ID,
			CreatedAt: record.CreatedAt,
			UpdatedAt: record.UpdatedAt,
		},
		RequestType:     entity.AuthzPolicyRequestType(record.RequestType),
		Status:          entity.AuthzPolicyRequestStatus(record.Status),
		ScopeType:       entity.AuthzPolicyScopeType(record.ScopeType),
		TargetOrgID:     derefString(record.TargetOrgID),
		TargetKey:       record.TargetKey,
		PayloadJSON:     record.PayloadJSON,
		PreviewJSON:     record.PreviewJSON,
		RequestNote:     record.RequestNote,
		RequestedBy:     derefString(record.RequestedBy),
		RequestedByRole: record.RequestedByRole,
		ReviewNote:      record.ReviewNote,
		ReviewedBy:      derefString(record.ReviewedBy),
		ReviewedAt:      record.ReviewedAt,
		Executed:        record.Executed,
		ExecutedAt:      record.ExecutedAt,
		ExecutedLogID:   derefString(record.ExecutedLogID),
		TraceID:         record.TraceID,
	}, nil
}

func (r *AuthzPolicyRepositoryImpl) UpdatePolicyChangeRequest(ctx context.Context, req *entity.AuthzPolicyChangeRequest) error {
	if req == nil {
		return nil
	}
	record := authzPolicyChangeRequestRecord{
		ID:              req.ID,
		CreatedAt:       req.CreatedAt,
		UpdatedAt:       req.UpdatedAt,
		RequestType:     string(req.RequestType),
		Status:          string(req.Status),
		ScopeType:       string(req.ScopeType),
		TargetOrgID:     nullableString(req.TargetOrgID),
		TargetKey:       req.TargetKey,
		PayloadJSON:     req.PayloadJSON,
		PreviewJSON:     req.PreviewJSON,
		RequestNote:     req.RequestNote,
		RequestedBy:     nullableString(req.RequestedBy),
		RequestedByRole: req.RequestedByRole,
		ReviewNote:      req.ReviewNote,
		ReviewedBy:      nullableString(req.ReviewedBy),
		ReviewedAt:      req.ReviewedAt,
		Executed:        req.Executed,
		ExecutedAt:      req.ExecutedAt,
		ExecutedLogID:   nullableString(req.ExecutedLogID),
		TraceID:         req.TraceID,
	}
	return r.db.WithContext(ctx).Model(&authzPolicyChangeRequestRecord{}).Where("id = ?", req.ID).Updates(record).Error
}

func nullableString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func (r *AuthzPolicyRepositoryImpl) RejectExpiredPolicyChangeRequests(ctx context.Context, before time.Time, reviewedBy, reviewNote string) (int64, error) {
	var reviewedByValue interface{}
	trimmedReviewedBy := strings.TrimSpace(reviewedBy)
	if trimmedReviewedBy != "" {
		if _, err := uuid.Parse(trimmedReviewedBy); err == nil {
			reviewedByValue = trimmedReviewedBy
		}
	}
	updates := map[string]interface{}{
		"status":      string(entity.AuthzPolicyRequestRejected),
		"reviewed_by": reviewedByValue,
		"reviewed_at": time.Now(),
		"review_note": strings.TrimSpace(reviewNote),
		"updated_at":  time.Now(),
	}
	tx := r.db.WithContext(ctx).
		Model(&authzPolicyChangeRequestRecord{}).
		Where("status = ? AND created_at < ?", string(entity.AuthzPolicyRequestPending), before).
		Updates(updates)
	return tx.RowsAffected, tx.Error
}
