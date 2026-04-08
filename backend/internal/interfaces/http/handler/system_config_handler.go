package handler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	appservice "github.com/Snowitty-Re/CNtunyuan/internal/application/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/middleware"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const maskedSecretValue = "******"

type SystemConfigHandler struct {
	authzService *appservice.AuthorizationService
}

type updateSystemConfigRequest struct {
	Config map[string]interface{} `json:"config"`
}

func NewSystemConfigHandler() *SystemConfigHandler {
	return &SystemConfigHandler{}
}

func NewSystemConfigHandlerWithAuthz(authzService *appservice.AuthorizationService) *SystemConfigHandler {
	return &SystemConfigHandler{authzService: authzService}
}

func (h *SystemConfigHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	system := router.Group("/system")
	system.Use(authMiddleware.Required(), middleware.RequireAdmin())
	{
		system.GET("/config", h.GetConfig)
		system.PUT("/config", h.UpdateConfig)
		system.GET("/authz/role-permissions", h.ListRolePermissions)
		system.PUT("/authz/role-permissions/:role", middleware.RequireSuperAdmin(), h.ReplaceRolePermissions)
		system.POST("/authz/preview/role-permissions/:role", h.PreviewRolePermissions)
		system.GET("/authz/policy-rules", h.ListPolicyRules)
		system.PUT("/authz/policy-rules/:permission_code", middleware.RequireSuperAdmin(), h.ReplacePolicyRules)
		system.POST("/authz/preview/policy-rules/:permission_code", h.PreviewPolicyRules)
		system.GET("/authz/decisions", h.ListAuthzDecisions)
		system.GET("/authz/change-logs", h.ListAuthzPolicyChanges)
		system.GET("/authz/requests", h.ListAuthzPolicyChangeRequests)
		system.POST("/authz/requests/role-permissions/:role", h.SubmitRolePermissionsChangeRequest)
		system.POST("/authz/requests/policy-rules/:permission_code", h.SubmitPolicyRulesChangeRequest)
		system.POST("/authz/requests/rollback/:change_id", h.SubmitRollbackChangeRequest)
		system.POST("/authz/requests/:request_id/review", h.ReviewPolicyChangeRequest)
		system.GET("/authz/notifications", h.ListAuthzNotifications)
		system.GET("/authz/notifications/unread-count", h.CountUnreadAuthzNotifications)
		system.POST("/authz/notifications/:notification_id/read", h.MarkAuthzNotificationRead)
		system.POST("/authz/notifications/read-all", h.MarkAllAuthzNotificationsRead)
		system.GET("/authz/rollback-preview/:change_id", middleware.RequireSuperAdmin(), h.PreviewRollbackAuthzPolicyChange)
		system.POST("/authz/rollback/:change_id", middleware.RequireSuperAdmin(), h.RollbackAuthzPolicyChange)
		system.POST("/authz/refresh", h.RefreshAuthzCache)
	}
}

type replaceRolePermissionsRequest struct {
	PermissionCodes []string `json:"permission_codes"`
	ApprovalCode    string   `json:"approval_code"`
}

type policyRuleItem struct {
	ResourceType string `json:"resource_type"`
	ScopeRule    string `json:"scope_rule"`
	Effect       string `json:"effect"`
	Priority     int    `json:"priority"`
	Enabled      bool   `json:"enabled"`
}

type replacePolicyRulesRequest struct {
	Rules        []policyRuleItem `json:"rules"`
	ApprovalCode string           `json:"approval_code"`
}

type listAuthzDecisionsRequest struct {
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
	Action       string `form:"action"`
	Allowed      string `form:"allowed"`
	OperatorID   string `form:"operator_id"`
	ResourceType string `form:"resource_type"`
	Reason       string `form:"reason"`
	StartTime    string `form:"start_time"`
	EndTime      string `form:"end_time"`
}

type listAuthzPolicyChangesRequest struct {
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
	Operation    string `form:"operation"`
	ChangeType   string `form:"change_type"`
	TargetKey    string `form:"target_key"`
	RollbackOfID string `form:"rollback_of_id"`
	OperatorID   string `form:"operator_id"`
	StartTime    string `form:"start_time"`
	EndTime      string `form:"end_time"`
}

type rollbackAuthzPolicyRequest struct {
	ChangeID     string `json:"change_id"`
	ApprovalCode string `json:"approval_code"`
}

type submitPolicyChangeRequest struct {
	RequestNote string `json:"request_note"`
	ScopeType   string `json:"scope_type"`
	TargetOrgID string `json:"target_org_id"`
}

type reviewPolicyChangeRequest struct {
	Approve    bool   `json:"approve"`
	ReviewNote string `json:"review_note"`
}

type listAuthzPolicyChangeRequestsRequest struct {
	Page        int    `form:"page"`
	PageSize    int    `form:"page_size"`
	RequestType string `form:"request_type"`
	Status      string `form:"status"`
	ScopeType   string `form:"scope_type"`
	RequestedBy string `form:"requested_by"`
	TargetOrgID string `form:"target_org_id"`
	TargetKey   string `form:"target_key"`
	StartTime   string `form:"start_time"`
	EndTime     string `form:"end_time"`
}

type listAuthzNotificationsRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
	Status    string `form:"status"`
	Category  string `form:"category"`
	RelatedID string `form:"related_id"`
}

func parseRole(input string) (entity.Role, bool) {
	role := entity.Role(strings.ToLower(strings.TrimSpace(input)))
	switch role {
	case entity.RoleSuperAdmin, entity.RoleAdmin, entity.RoleManager, entity.RoleVolunteer:
		return role, true
	default:
		return "", false
	}
}

func (h *SystemConfigHandler) ensureAuthzService(c *gin.Context) bool {
	if h.authzService == nil {
		response.InternalServerError(c, "authz service not configured")
		return false
	}
	return true
}

func (h *SystemConfigHandler) ListRolePermissions(c *gin.Context) {
	if !h.ensureAuthzService(c) {
		return
	}
	items, err := h.authzService.ListRolePermissions(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "failed to load role permissions")
		return
	}
	response.Success(c, gin.H{"items": items})
}

func (h *SystemConfigHandler) ReplaceRolePermissions(c *gin.Context) {
	if !h.ensureAuthzService(c) {
		return
	}
	role, ok := parseRole(c.Param("role"))
	if !ok {
		response.BadRequest(c, "invalid role")
		return
	}
	var req replaceRolePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	if isAuthzApprovalWorkflowEnabled() {
		response.Forbidden(c, "approval workflow enabled; please submit change request instead")
		return
	}
	if !validateAuthzApproval(c, req.ApprovalCode) {
		return
	}
	ctx := withAuthzOperatorContext(c)
	if err := h.authzService.ReplaceRolePermissions(ctx, role, req.PermissionCodes); err != nil {
		response.InternalServerError(c, fmt.Sprintf("failed to replace role permissions: %v", err))
		return
	}
	response.SuccessWithMessage(c, "角色权限已更新", gin.H{
		"role":             role,
		"permission_codes": req.PermissionCodes,
	})
}

// PreviewRolePermissions 预检查角色权限变更
// @Summary 预检查角色权限变更
// @Description Dry-run 方式预览角色权限替换的增删项，不会实际落库
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param role path string true "角色：super_admin/admin/manager/volunteer"
// @Param request body replaceRolePermissionsRequest true "目标权限集合"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /system/authz/preview/role-permissions/{role} [post]
func (h *SystemConfigHandler) PreviewRolePermissions(c *gin.Context) {
	if !h.ensureAuthzService(c) {
		return
	}
	role, ok := parseRole(c.Param("role"))
	if !ok {
		response.BadRequest(c, "invalid role")
		return
	}
	var req replaceRolePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	result, err := h.authzService.PreviewReplaceRolePermissions(c.Request.Context(), role, req.PermissionCodes)
	if err != nil {
		response.InternalServerError(c, fmt.Sprintf("failed to preview role permissions: %v", err))
		return
	}
	response.Success(c, result)
}

func (h *SystemConfigHandler) ListPolicyRules(c *gin.Context) {
	if !h.ensureAuthzService(c) {
		return
	}
	items, err := h.authzService.ListPolicyRules(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "failed to load policy rules")
		return
	}
	response.Success(c, gin.H{"items": items})
}

func (h *SystemConfigHandler) ReplacePolicyRules(c *gin.Context) {
	if !h.ensureAuthzService(c) {
		return
	}
	permissionCode := strings.TrimSpace(c.Param("permission_code"))
	if permissionCode == "" {
		response.BadRequest(c, "permission_code is required")
		return
	}
	var req replacePolicyRulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	if isAuthzApprovalWorkflowEnabled() {
		response.Forbidden(c, "approval workflow enabled; please submit change request instead")
		return
	}
	if !validateAuthzApproval(c, req.ApprovalCode) {
		return
	}

	rules := make([]repository.PolicyRule, 0, len(req.Rules))
	for _, item := range req.Rules {
		rules = append(rules, repository.PolicyRule{
			PermissionCode: permissionCode,
			ResourceType:   strings.TrimSpace(item.ResourceType),
			ScopeRule:      strings.TrimSpace(strings.ToUpper(item.ScopeRule)),
			Effect:         strings.TrimSpace(strings.ToLower(item.Effect)),
			Priority:       item.Priority,
			Enabled:        item.Enabled,
		})
	}

	ctx := withAuthzOperatorContext(c)
	if err := h.authzService.ReplacePolicyRules(ctx, permissionCode, rules); err != nil {
		response.InternalServerError(c, fmt.Sprintf("failed to replace policy rules: %v", err))
		return
	}
	response.SuccessWithMessage(c, "策略规则已更新", gin.H{
		"permission_code": permissionCode,
		"rules":           rules,
	})
}

// PreviewPolicyRules 预检查策略规则变更
// @Summary 预检查策略规则变更
// @Description Dry-run 方式预览策略规则替换的差异与冲突校验，不会实际落库
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param permission_code path string true "权限编码，如 task:view"
// @Param request body replacePolicyRulesRequest true "目标规则集合"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /system/authz/preview/policy-rules/{permission_code} [post]
func (h *SystemConfigHandler) PreviewPolicyRules(c *gin.Context) {
	if !h.ensureAuthzService(c) {
		return
	}
	permissionCode := strings.TrimSpace(c.Param("permission_code"))
	if permissionCode == "" {
		response.BadRequest(c, "permission_code is required")
		return
	}
	var req replacePolicyRulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	rules := make([]repository.PolicyRule, 0, len(req.Rules))
	for _, item := range req.Rules {
		rules = append(rules, repository.PolicyRule{
			PermissionCode: permissionCode,
			ResourceType:   strings.TrimSpace(item.ResourceType),
			ScopeRule:      strings.TrimSpace(strings.ToUpper(item.ScopeRule)),
			Effect:         strings.TrimSpace(strings.ToLower(item.Effect)),
			Priority:       item.Priority,
			Enabled:        item.Enabled,
		})
	}
	result, err := h.authzService.PreviewReplacePolicyRules(c.Request.Context(), permissionCode, rules)
	if err != nil {
		response.InternalServerError(c, fmt.Sprintf("failed to preview policy rules: %v", err))
		return
	}
	response.Success(c, result)
}

func (h *SystemConfigHandler) RefreshAuthzCache(c *gin.Context) {
	if !h.ensureAuthzService(c) {
		return
	}
	h.authzService.RefreshPolicies()
	response.SuccessWithMessage(c, "权限策略缓存已刷新", gin.H{
		"refreshed": true,
	})
}

// ListAuthzDecisions 获取权限决策日志
// @Summary 获取权限决策日志
// @Description 分页查询权限决策日志，支持按动作、是否允许、操作者、资源类型和时间范围筛选
// @Tags 系统配置
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码，默认1"
// @Param page_size query int false "每页数量，默认20，最大100"
// @Param action query string false "权限动作，如 user:modify"
// @Param allowed query string false "是否允许，true/false"
// @Param operator_id query string false "操作者ID"
// @Param resource_type query string false "资源类型，如 user/task"
// @Param reason query string false "决策原因，如 DENY_ORG_SCOPE"
// @Param start_time query string false "开始时间，格式：2006-01-02"
// @Param end_time query string false "结束时间，格式：2006-01-02"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /system/authz/decisions [get]
func (h *SystemConfigHandler) ListAuthzDecisions(c *gin.Context) {
	if !h.ensureAuthzService(c) {
		return
	}

	var req listAuthzDecisionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query")
		return
	}

	query := entity.NewAuthzDecisionQuery()
	if req.Page > 0 {
		query.Page = req.Page
	}
	if req.PageSize > 0 {
		query.PageSize = req.PageSize
	}
	query.Action = strings.TrimSpace(req.Action)
	query.OperatorID = strings.TrimSpace(req.OperatorID)
	query.ResourceType = strings.TrimSpace(req.ResourceType)
	query.Reason = strings.TrimSpace(req.Reason)

	if allowed := strings.TrimSpace(strings.ToLower(req.Allowed)); allowed != "" {
		if allowed != "true" && allowed != "false" {
			response.BadRequest(c, "allowed must be true or false")
			return
		}
		boolValue := allowed == "true"
		query.Allowed = &boolValue
	}

	if req.StartTime != "" {
		start, err := time.Parse("2006-01-02", req.StartTime)
		if err != nil {
			response.BadRequest(c, "invalid start_time")
			return
		}
		query.StartTime = &start
	}
	if req.EndTime != "" {
		end, err := time.Parse("2006-01-02", req.EndTime)
		if err != nil {
			response.BadRequest(c, "invalid end_time")
			return
		}
		endOfDay := end.Add(24*time.Hour - time.Second)
		query.EndTime = &endOfDay
	}

	result, err := h.authzService.ListDecisions(c.Request.Context(), query)
	if err != nil {
		response.InternalServerError(c, fmt.Sprintf("failed to list authz decisions: %v", err))
		return
	}
	response.Success(c, result)
}

// ListAuthzPolicyChanges 获取策略变更日志
// @Summary 获取策略变更日志
// @Description 分页查询权限策略变更日志，支持按变更类型、目标键、操作者和时间范围筛选
// @Tags 系统配置
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码，默认1"
// @Param page_size query int false "每页数量，默认20，最大100"
// @Param operation query string false "操作类型：apply/rollback"
// @Param change_type query string false "变更类型：role_permissions/policy_rules"
// @Param target_key query string false "目标键，如 admin 或 task:view"
// @Param rollback_of_id query string false "回滚来源变更ID"
// @Param operator_id query string false "操作者ID"
// @Param start_time query string false "开始时间，格式：2006-01-02"
// @Param end_time query string false "结束时间，格式：2006-01-02"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /system/authz/change-logs [get]
func (h *SystemConfigHandler) ListAuthzPolicyChanges(c *gin.Context) {
	if !h.ensureAuthzService(c) {
		return
	}
	var req listAuthzPolicyChangesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query")
		return
	}

	query := entity.NewAuthzPolicyChangeQuery()
	if req.Page > 0 {
		query.Page = req.Page
	}
	if req.PageSize > 0 {
		query.PageSize = req.PageSize
	}
	query.Operation = strings.TrimSpace(req.Operation)
	query.ChangeType = strings.TrimSpace(req.ChangeType)
	query.TargetKey = strings.TrimSpace(req.TargetKey)
	query.RollbackOfID = strings.TrimSpace(req.RollbackOfID)
	query.OperatorID = strings.TrimSpace(req.OperatorID)

	if req.StartTime != "" {
		start, err := time.Parse("2006-01-02", req.StartTime)
		if err != nil {
			response.BadRequest(c, "invalid start_time")
			return
		}
		query.StartTime = &start
	}
	if req.EndTime != "" {
		end, err := time.Parse("2006-01-02", req.EndTime)
		if err != nil {
			response.BadRequest(c, "invalid end_time")
			return
		}
		endOfDay := end.Add(24*time.Hour - time.Second)
		query.EndTime = &endOfDay
	}

	result, err := h.authzService.ListPolicyChanges(c.Request.Context(), query)
	if err != nil {
		response.InternalServerError(c, fmt.Sprintf("failed to list authz policy changes: %v", err))
		return
	}
	response.Success(c, result)
}

// PreviewRollbackAuthzPolicyChange 回滚预览
// @Summary 回滚权限策略预览
// @Description 预览指定策略变更ID执行回滚后的影响，不会实际写入
// @Tags 系统配置
// @Produce json
// @Security ApiKeyAuth
// @Param change_id path string true "策略变更日志ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /system/authz/rollback-preview/{change_id} [get]
func (h *SystemConfigHandler) PreviewRollbackAuthzPolicyChange(c *gin.Context) {
	if !h.ensureAuthzService(c) {
		return
	}
	changeID := strings.TrimSpace(c.Param("change_id"))
	if changeID == "" {
		response.BadRequest(c, "change_id is required")
		return
	}
	result, err := h.authzService.PreviewRollbackPolicyChange(c.Request.Context(), changeID)
	if err != nil {
		response.InternalServerError(c, fmt.Sprintf("failed to preview rollback authz policy change: %v", err))
		return
	}
	response.Success(c, result)
}

// RollbackAuthzPolicyChange 回滚权限策略变更
// @Summary 回滚权限策略变更
// @Description 按策略变更日志 ID 回滚到变更前状态，仅超级管理员可操作
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param change_id path string true "策略变更日志ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /system/authz/rollback/{change_id} [post]
func (h *SystemConfigHandler) RollbackAuthzPolicyChange(c *gin.Context) {
	if !h.ensureAuthzService(c) {
		return
	}
	changeID := strings.TrimSpace(c.Param("change_id"))
	approvalCode := strings.TrimSpace(c.Query("approval_code"))
	if c.Request.ContentLength > 0 {
		var req rollbackAuthzPolicyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "invalid request body")
			return
		}
		if strings.TrimSpace(req.ChangeID) != "" {
			changeID = strings.TrimSpace(req.ChangeID)
		}
		if strings.TrimSpace(req.ApprovalCode) != "" {
			approvalCode = strings.TrimSpace(req.ApprovalCode)
		}
	}
	if !validateAuthzApproval(c, approvalCode) {
		return
	}
	if isAuthzApprovalWorkflowEnabled() {
		response.Forbidden(c, "approval workflow enabled; please submit rollback request instead")
		return
	}
	if changeID == "" {
		response.BadRequest(c, "change_id is required")
		return
	}

	ctx := withAuthzOperatorContext(c)
	result, err := h.authzService.RollbackPolicyChange(ctx, changeID)
	if err != nil {
		response.InternalServerError(c, fmt.Sprintf("failed to rollback authz policy change: %v", err))
		return
	}
	response.SuccessWithMessage(c, "权限策略已回滚", result)
}

// ListAuthzPolicyChangeRequests 查询策略变更申请
// @Summary 查询策略变更申请
// @Description 分页查询策略变更申请列表（提交、审批、执行状态）
// @Tags 系统配置
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码，默认1"
// @Param page_size query int false "每页数量，默认20，最大100"
// @Param request_type query string false "申请类型(role_permissions/policy_rules/rollback)"
// @Param status query string false "申请状态(pending/approved/rejected)"
// @Param scope_type query string false "作用域(global/org)"
// @Param requested_by query string false "申请人用户ID"
// @Param target_org_id query string false "目标组织ID（组织作用域时）"
// @Param target_key query string false "目标键（角色/权限码/变更ID）"
// @Param start_time query string false "开始时间，格式：2006-01-02"
// @Param end_time query string false "结束时间，格式：2006-01-02"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /system/authz/requests [get]
func (h *SystemConfigHandler) ListAuthzPolicyChangeRequests(c *gin.Context) {
	if !h.ensureAuthzService(c) {
		return
	}
	var req listAuthzPolicyChangeRequestsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query")
		return
	}
	query := entity.NewAuthzPolicyChangeRequestQuery()
	if req.Page > 0 {
		query.Page = req.Page
	}
	if req.PageSize > 0 {
		query.PageSize = req.PageSize
	}
	query.RequestType = strings.TrimSpace(req.RequestType)
	query.Status = strings.TrimSpace(req.Status)
	query.ScopeType = strings.TrimSpace(req.ScopeType)
	query.RequestedBy = strings.TrimSpace(req.RequestedBy)
	query.TargetOrgID = strings.TrimSpace(req.TargetOrgID)
	query.TargetKey = strings.TrimSpace(req.TargetKey)
	if req.StartTime != "" {
		start, err := time.Parse("2006-01-02", req.StartTime)
		if err != nil {
			response.BadRequest(c, "invalid start_time")
			return
		}
		query.StartTime = &start
	}
	if req.EndTime != "" {
		end, err := time.Parse("2006-01-02", req.EndTime)
		if err != nil {
			response.BadRequest(c, "invalid end_time")
			return
		}
		endOfDay := end.Add(24*time.Hour - time.Second)
		query.EndTime = &endOfDay
	}
	result, err := h.authzService.ListPolicyChangeRequests(c.Request.Context(), query)
	if err != nil {
		response.InternalServerError(c, fmt.Sprintf("failed to list authz policy change requests: %v", err))
		return
	}
	response.Success(c, result)
}

// SubmitRolePermissionsChangeRequest 提交角色权限变更申请
// @Summary 提交角色权限变更申请
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Router /system/authz/requests/role-permissions/{role} [post]
func (h *SystemConfigHandler) SubmitRolePermissionsChangeRequest(c *gin.Context) {
	if !h.ensureAuthzService(c) {
		return
	}
	role, ok := parseRole(c.Param("role"))
	if !ok {
		response.BadRequest(c, "invalid role")
		return
	}
	var req struct {
		PermissionCodes []string `json:"permission_codes"`
		RequestNote     string   `json:"request_note"`
		ScopeType       string   `json:"scope_type"`
		TargetOrgID     string   `json:"target_org_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	ctx := withAuthzOperatorContext(c)
	result, err := h.authzService.SubmitRolePermissionsChangeRequestWithScope(
		ctx,
		role,
		req.PermissionCodes,
		req.RequestNote,
		req.ScopeType,
		req.TargetOrgID,
	)
	if err != nil {
		response.InternalServerError(c, fmt.Sprintf("failed to submit role permission change request: %v", err))
		return
	}
	response.SuccessWithMessage(c, "变更申请已提交", result)
}

// SubmitPolicyRulesChangeRequest 提交策略规则变更申请
// @Summary 提交策略规则变更申请
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Router /system/authz/requests/policy-rules/{permission_code} [post]
func (h *SystemConfigHandler) SubmitPolicyRulesChangeRequest(c *gin.Context) {
	if !h.ensureAuthzService(c) {
		return
	}
	permissionCode := strings.TrimSpace(c.Param("permission_code"))
	if permissionCode == "" {
		response.BadRequest(c, "permission_code is required")
		return
	}
	var req struct {
		Rules       []policyRuleItem `json:"rules"`
		RequestNote string           `json:"request_note"`
		ScopeType   string           `json:"scope_type"`
		TargetOrgID string           `json:"target_org_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	rules := make([]repository.PolicyRule, 0, len(req.Rules))
	for _, item := range req.Rules {
		rules = append(rules, repository.PolicyRule{
			PermissionCode: permissionCode,
			ResourceType:   strings.TrimSpace(item.ResourceType),
			ScopeRule:      strings.TrimSpace(strings.ToUpper(item.ScopeRule)),
			Effect:         strings.TrimSpace(strings.ToLower(item.Effect)),
			Priority:       item.Priority,
			Enabled:        item.Enabled,
		})
	}
	ctx := withAuthzOperatorContext(c)
	result, err := h.authzService.SubmitPolicyRulesChangeRequestWithScope(
		ctx,
		permissionCode,
		rules,
		req.RequestNote,
		req.ScopeType,
		req.TargetOrgID,
	)
	if err != nil {
		response.InternalServerError(c, fmt.Sprintf("failed to submit policy rules change request: %v", err))
		return
	}
	response.SuccessWithMessage(c, "变更申请已提交", result)
}

// SubmitRollbackChangeRequest 提交回滚申请
// @Summary 提交回滚申请
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Router /system/authz/requests/rollback/{change_id} [post]
func (h *SystemConfigHandler) SubmitRollbackChangeRequest(c *gin.Context) {
	if !h.ensureAuthzService(c) {
		return
	}
	changeID := strings.TrimSpace(c.Param("change_id"))
	if changeID == "" {
		response.BadRequest(c, "change_id is required")
		return
	}
	var req submitPolicyChangeRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "invalid request body")
			return
		}
	}
	ctx := withAuthzOperatorContext(c)
	result, err := h.authzService.SubmitRollbackChangeRequestWithScope(
		ctx,
		changeID,
		req.RequestNote,
		req.ScopeType,
		req.TargetOrgID,
	)
	if err != nil {
		response.InternalServerError(c, fmt.Sprintf("failed to submit rollback change request: %v", err))
		return
	}
	response.SuccessWithMessage(c, "回滚申请已提交", result)
}

// ReviewPolicyChangeRequest 审批策略变更申请
// @Summary 审批策略变更申请
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Router /system/authz/requests/{request_id}/review [post]
func (h *SystemConfigHandler) ReviewPolicyChangeRequest(c *gin.Context) {
	if !h.ensureAuthzService(c) {
		return
	}
	requestID := strings.TrimSpace(c.Param("request_id"))
	if requestID == "" {
		response.BadRequest(c, "request_id is required")
		return
	}
	var req reviewPolicyChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	ctx := withAuthzOperatorContext(c)
	result, err := h.authzService.ReviewPolicyChangeRequest(ctx, requestID, req.Approve, req.ReviewNote)
	if err != nil {
		response.InternalServerError(c, fmt.Sprintf("failed to review policy change request: %v", err))
		return
	}
	response.SuccessWithMessage(c, "审批完成", result)
}

// ListAuthzNotifications 查询权限相关通知
// @Summary 查询权限相关通知
// @Tags 系统配置
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码，默认1"
// @Param page_size query int false "每页数量，默认20，最大100"
// @Param status query string false "状态(unread/read)"
// @Param category query string false "分类(authz)"
// @Param related_id query string false "关联ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /system/authz/notifications [get]
func (h *SystemConfigHandler) ListAuthzNotifications(c *gin.Context) {
	if !h.ensureAuthzService(c) {
		return
	}
	var req listAuthzNotificationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query")
		return
	}
	query := entity.NewSystemNotificationQuery()
	if req.Page > 0 {
		query.Page = req.Page
	}
	if req.PageSize > 0 {
		query.PageSize = req.PageSize
	}
	query.Status = strings.TrimSpace(req.Status)
	query.Category = strings.TrimSpace(req.Category)
	query.RelatedID = strings.TrimSpace(req.RelatedID)
	query.UserID = strings.TrimSpace(middleware.GetUserID(c))
	query.UserRole = strings.TrimSpace(string(middleware.GetUserRole(c)))

	result, err := h.authzService.ListAuthzNotifications(c.Request.Context(), query)
	if err != nil {
		response.InternalServerError(c, fmt.Sprintf("failed to list authz notifications: %v", err))
		return
	}
	response.Success(c, result)
}

// CountUnreadAuthzNotifications 统计未读权限通知
// @Summary 统计未读权限通知
// @Tags 系统配置
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /system/authz/notifications/unread-count [get]
func (h *SystemConfigHandler) CountUnreadAuthzNotifications(c *gin.Context) {
	if !h.ensureAuthzService(c) {
		return
	}
	userID := strings.TrimSpace(middleware.GetUserID(c))
	userRole := strings.TrimSpace(string(middleware.GetUserRole(c)))
	count, err := h.authzService.CountUnreadAuthzNotifications(c.Request.Context(), userID, userRole)
	if err != nil {
		response.InternalServerError(c, fmt.Sprintf("failed to count unread notifications: %v", err))
		return
	}
	response.Success(c, gin.H{"unread_count": count})
}

// MarkAuthzNotificationRead 标记权限通知已读
// @Summary 标记权限通知已读
// @Tags 系统配置
// @Produce json
// @Security ApiKeyAuth
// @Param notification_id path string true "通知ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /system/authz/notifications/{notification_id}/read [post]
func (h *SystemConfigHandler) MarkAuthzNotificationRead(c *gin.Context) {
	if !h.ensureAuthzService(c) {
		return
	}
	notificationID := strings.TrimSpace(c.Param("notification_id"))
	if notificationID == "" {
		response.BadRequest(c, "notification_id is required")
		return
	}
	userID := strings.TrimSpace(middleware.GetUserID(c))
	userRole := strings.TrimSpace(string(middleware.GetUserRole(c)))
	updated, err := h.authzService.MarkAuthzNotificationRead(c.Request.Context(), notificationID, userID, userRole)
	if err != nil {
		response.InternalServerError(c, fmt.Sprintf("failed to mark notification read: %v", err))
		return
	}
	response.Success(c, gin.H{"updated": updated})
}

// MarkAllAuthzNotificationsRead 全部标记已读
// @Summary 全部标记已读
// @Tags 系统配置
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /system/authz/notifications/read-all [post]
func (h *SystemConfigHandler) MarkAllAuthzNotificationsRead(c *gin.Context) {
	if !h.ensureAuthzService(c) {
		return
	}
	userID := strings.TrimSpace(middleware.GetUserID(c))
	userRole := strings.TrimSpace(string(middleware.GetUserRole(c)))
	count, err := h.authzService.MarkAllAuthzNotificationsRead(c.Request.Context(), userID, userRole)
	if err != nil {
		response.InternalServerError(c, fmt.Sprintf("failed to mark all notifications read: %v", err))
		return
	}
	response.Success(c, gin.H{"updated_count": count})
}

func validateAuthzApproval(c *gin.Context, approvalCode string) bool {
	cfg := config.GetConfig()
	if cfg == nil {
		response.InternalServerError(c, "config not loaded")
		return false
	}
	if !cfg.System.AuthzPolicyChangeRequiresApproval {
		return true
	}
	expected := strings.TrimSpace(cfg.System.AuthzPolicyChangeApprovalCode)
	if expected == "" {
		response.InternalServerError(c, "authz policy approval gate enabled but approval code is not configured")
		return false
	}
	if strings.TrimSpace(approvalCode) != expected {
		response.Forbidden(c, "approval code invalid")
		return false
	}
	return true
}

func isAuthzApprovalWorkflowEnabled() bool {
	cfg := config.GetConfig()
	if cfg == nil {
		return false
	}
	return cfg.System.AuthzPolicyChangeRequiresApproval
}

func withAuthzOperatorContext(c *gin.Context) context.Context {
	ctx := c.Request.Context()
	if uid := middleware.GetUserID(c); strings.TrimSpace(uid) != "" {
		ctx = context.WithValue(ctx, "authz_operator_id", uid)
	}
	if role := middleware.GetUserRole(c); strings.TrimSpace(string(role)) != "" {
		ctx = context.WithValue(ctx, "authz_operator_role", string(role))
	}
	if orgID := middleware.GetOrgID(c); strings.TrimSpace(orgID) != "" {
		ctx = context.WithValue(ctx, "authz_operator_org_id", orgID)
	}
	if traceID, exists := c.Get("trace_id"); exists {
		if tid, ok := traceID.(string); ok && strings.TrimSpace(tid) != "" {
			ctx = context.WithValue(ctx, "trace_id", tid)
		}
	}
	return ctx
}

// GetConfig 获取系统配置
// @Summary 获取系统配置
// @Description 获取当前系统配置（敏感字段已掩码，仅管理员可访问）
// @Tags 系统配置
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /system/config [get]
func (h *SystemConfigHandler) GetConfig(c *gin.Context) {
	cfg := config.GetConfig()
	if cfg == nil {
		response.InternalServerError(c, "config not loaded")
		return
	}

	flat := flattenConfig(cfg)
	for _, key := range sensitiveKeys() {
		if v, ok := flat[key]; ok {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				flat[key] = maskedSecretValue
			}
		}
	}

	response.Success(c, gin.H{
		"config":           flat,
		"sensitive_fields": sensitiveKeys(),
	})
}

// UpdateConfig 更新系统配置
// @Summary 更新系统配置
// @Description 更新并持久化系统配置到 config.yaml（敏感字段传掩码将保留原值）
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body updateSystemConfigRequest true "配置项（扁平 key-value）"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /system/config [put]
func (h *SystemConfigHandler) UpdateConfig(c *gin.Context) {
	cfg := config.GetConfig()
	if cfg == nil {
		response.InternalServerError(c, "config not loaded")
		return
	}

	var req updateSystemConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	if req.Config == nil {
		response.BadRequest(c, "config is required")
		return
	}

	flat := flattenConfig(cfg)
	sensitive := make(map[string]struct{}, len(sensitiveKeys()))
	for _, key := range sensitiveKeys() {
		sensitive[key] = struct{}{}
	}

	for key, raw := range req.Config {
		current, ok := flat[key]
		if !ok {
			continue
		}

		if _, isSensitive := sensitive[key]; isSensitive {
			val := strings.TrimSpace(toString(raw))
			if val == "" || val == maskedSecretValue {
				continue
			}
			flat[key] = val
			continue
		}

		flat[key] = coerceValue(raw, current)
	}

	nested := unflattenConfig(flat)
	bytes, err := yaml.Marshal(nested)
	if err != nil {
		response.InternalServerError(c, "failed to encode config")
		return
	}

	configPath := viper.ConfigFileUsed()
	if strings.TrimSpace(configPath) == "" {
		configPath = filepath.Join("config", "config.yaml")
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		response.InternalServerError(c, "failed to ensure config dir")
		return
	}
	if err := os.WriteFile(configPath, bytes, 0o600); err != nil {
		response.InternalServerError(c, "failed to save config")
		return
	}

	var newCfg config.Config
	if err := yaml.Unmarshal(bytes, &newCfg); err != nil {
		response.InternalServerError(c, "failed to apply config")
		return
	}
	config.SetConfig(&newCfg)

	masked := flattenConfig(&newCfg)
	for _, key := range sensitiveKeys() {
		if v, ok := masked[key]; ok {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				masked[key] = maskedSecretValue
			}
		}
	}

	response.SuccessWithMessage(c, "配置已保存（重启后将完全生效）", gin.H{
		"config":           masked,
		"sensitive_fields": sensitiveKeys(),
	})
}

func sensitiveKeys() []string {
	return []string{
		"database.password",
		"jwt.secret",
		"wechat.app_secret",
		"wechat.api_key",
		"storage.oss_access_key_secret",
		"storage.cos_secret_key",
		"sms.aliyun_access_key_secret",
		"sms.tencent_secret_key",
		"email.smtp_password",
		"map.key",
		"map.tencent_key",
		"map.amap_key",
		"map.baidu_key",
		"notification.getui_app_key",
		"notification.getui_master_secret",
		"notification.jpush_master_secret",
		"system.authz_policy_change_approval_code",
	}
}

func flattenConfig(cfg *config.Config) map[string]interface{} {
	return map[string]interface{}{
		"server.port":             cfg.Server.Port,
		"server.mode":             cfg.Server.Mode,
		"server.domain":           cfg.Server.Domain,
		"server.read_timeout":     cfg.Server.ReadTimeout,
		"server.write_timeout":    cfg.Server.WriteTimeout,
		"server.max_header_bytes": cfg.Server.MaxHeaderBytes,
		"server.cors_origins":     cfg.Server.CORSOrigins,

		"database.type":              string(cfg.Database.Type),
		"database.host":              cfg.Database.Host,
		"database.port":              cfg.Database.Port,
		"database.user":              cfg.Database.User,
		"database.password":          cfg.Database.Password,
		"database.database":          cfg.Database.Database,
		"database.ssl_mode":          cfg.Database.SSLMode,
		"database.timezone":          cfg.Database.Timezone,
		"database.charset":           cfg.Database.Charset,
		"database.max_idle_conns":    cfg.Database.MaxIdleConns,
		"database.max_open_conns":    cfg.Database.MaxOpenConns,
		"database.conn_max_lifetime": cfg.Database.ConnMaxLifetime,

		"redis.host":           cfg.Redis.Host,
		"redis.port":           cfg.Redis.Port,
		"redis.password":       cfg.Redis.Password,
		"redis.db":             cfg.Redis.DB,
		"redis.pool_size":      cfg.Redis.PoolSize,
		"redis.min_idle_conns": cfg.Redis.MinIdleConns,

		"jwt.secret":       cfg.JWT.Secret,
		"jwt.expire_time":  cfg.JWT.ExpireTime,
		"jwt.refresh_time": cfg.JWT.RefreshTime,

		"wechat.app_id":       cfg.WeChat.AppID,
		"wechat.app_secret":   cfg.WeChat.AppSecret,
		"wechat.enable_login": cfg.WeChat.EnableLogin,
		"wechat.mch_id":       cfg.WeChat.MchID,
		"wechat.api_key":      cfg.WeChat.APIKey,
		"wechat.notify_url":   cfg.WeChat.NotifyURL,

		"storage.type":                  cfg.Storage.Type,
		"storage.local_path":            cfg.Storage.LocalPath,
		"storage.base_url":              cfg.Storage.BaseURL,
		"storage.max_file_size":         cfg.Storage.MaxFileSize,
		"storage.allowed_types":         cfg.Storage.AllowedTypes,
		"storage.oss_access_key_id":     cfg.Storage.OSSAccessKeyID,
		"storage.oss_access_key_secret": cfg.Storage.OSSAccessKeySecret,
		"storage.oss_endpoint":          cfg.Storage.OSSEndpoint,
		"storage.oss_bucket":            cfg.Storage.OSSBucket,
		"storage.oss_region":            cfg.Storage.OSSRegion,
		"storage.cos_secret_id":         cfg.Storage.COSSecretID,
		"storage.cos_secret_key":        cfg.Storage.COSSecretKey,
		"storage.cos_bucket":            cfg.Storage.COSBucket,
		"storage.cos_region":            cfg.Storage.COSRegion,

		"sms.provider":                 cfg.SMS.Provider,
		"sms.sign_name":                cfg.SMS.SignName,
		"sms.dev_mode":                 cfg.SMS.DevMode,
		"sms.code_expiry":              cfg.SMS.CodeExpiry,
		"sms.aliyun_access_key_id":     cfg.SMS.AliyunAccessKeyID,
		"sms.aliyun_access_key_secret": cfg.SMS.AliyunAccessSecret,
		"sms.tencent_secret_id":        cfg.SMS.TencentSecretID,
		"sms.tencent_secret_key":       cfg.SMS.TencentSecretKey,
		"sms.tencent_app_id":           cfg.SMS.TencentAppID,

		"email.enabled":       cfg.Email.Enabled,
		"email.smtp_host":     cfg.Email.SMTPHost,
		"email.smtp_port":     cfg.Email.SMTPPort,
		"email.smtp_user":     cfg.Email.SMTPUser,
		"email.smtp_password": cfg.Email.SMTPPassword,
		"email.from_name":     cfg.Email.FromName,
		"email.use_tls":       cfg.Email.UseTLS,

		"map.provider":    cfg.Map.Provider,
		"map.key":         cfg.Map.Key,
		"map.tencent_key": cfg.Map.TencentKey,
		"map.amap_key":    cfg.Map.AmapKey,
		"map.baidu_key":   cfg.Map.BaiduKey,

		"log.level":       cfg.Log.Level,
		"log.format":      cfg.Log.Format,
		"log.output_path": cfg.Log.OutputPath,
		"log.file_name":   cfg.Log.FileName,
		"log.max_size":    cfg.Log.MaxSize,
		"log.max_backups": cfg.Log.MaxBackups,
		"log.max_age":     cfg.Log.MaxAge,
		"log.compress":    cfg.Log.Compress,

		"notification.push_enabled":        cfg.Notification.PushEnabled,
		"notification.getui_app_id":        cfg.Notification.GetuiAppID,
		"notification.getui_app_key":       cfg.Notification.GetuiAppKey,
		"notification.getui_master_secret": cfg.Notification.GetuiMasterSecret,
		"notification.jpush_app_key":       cfg.Notification.JPushAppKey,
		"notification.jpush_master_secret": cfg.Notification.JPushMasterSecret,

		"system.default_org_name":                      cfg.System.DefaultOrgName,
		"system.default_org_code":                      cfg.System.DefaultOrgCode,
		"system.enable_register":                       cfg.System.EnableRegister,
		"system.enable_wechat_login":                   cfg.System.EnableWechatLogin,
		"system.enable_wechat_login_web":               resolveWebWechatLoginEnabled(cfg),
		"system.enable_wechat_login_mini_program":      resolveMiniProgramWechatLoginEnabled(cfg),
		"system.enable_sms_login":                      cfg.System.EnableSMSLogin,
		"system.authz_policy_change_requires_approval": cfg.System.AuthzPolicyChangeRequiresApproval,
		"system.authz_policy_change_approval_code":     cfg.System.AuthzPolicyChangeApprovalCode,
		"system.authz_policy_request_expire_hours":     cfg.System.AuthzPolicyRequestExpireHours,
		"system.admin_ips":                             cfg.System.AdminIPs,
		"system.rate_limit":                            cfg.System.RateLimit,

		"security.max_login_attempts": cfg.Security.MaxLoginAttempts,
		"security.lockout_duration":   cfg.Security.LockoutDuration,

		"backup.enabled":    cfg.Backup.Enabled,
		"backup.backup_dir": cfg.Backup.BackupDir,
		"backup.retention":  cfg.Backup.Retention,
	}
}

func resolveWebWechatLoginEnabled(cfg *config.Config) bool {
	if cfg == nil {
		return false
	}
	if viper.IsSet("system.enable_wechat_login_web") {
		return cfg.System.EnableWechatLoginWeb
	}
	return cfg.System.EnableWechatLogin
}

func resolveMiniProgramWechatLoginEnabled(cfg *config.Config) bool {
	if cfg == nil {
		return false
	}
	if viper.IsSet("system.enable_wechat_login_mini_program") {
		return cfg.System.EnableWechatLoginMiniProgram
	}
	return cfg.System.EnableWechatLogin
}

func unflattenConfig(flat map[string]interface{}) map[string]interface{} {
	root := map[string]interface{}{}
	for key, value := range flat {
		parts := strings.Split(key, ".")
		if len(parts) < 2 {
			continue
		}
		cur := root
		for i := 0; i < len(parts)-1; i++ {
			p := parts[i]
			next, ok := cur[p]
			if !ok {
				m := map[string]interface{}{}
				cur[p] = m
				cur = m
				continue
			}
			m, ok := next.(map[string]interface{})
			if !ok {
				m = map[string]interface{}{}
				cur[p] = m
			}
			cur = m
		}
		cur[parts[len(parts)-1]] = value
	}
	return root
}

func toString(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case fmt.Stringer:
		return x.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func coerceValue(raw interface{}, current interface{}) interface{} {
	switch current.(type) {
	case bool:
		if v, ok := raw.(bool); ok {
			return v
		}
		s := strings.TrimSpace(strings.ToLower(toString(raw)))
		return s == "1" || s == "true" || s == "yes" || s == "on"
	case int:
		if n, err := strconv.Atoi(strings.TrimSpace(toString(raw))); err == nil {
			return n
		}
		return current
	case int64:
		if n, err := strconv.ParseInt(strings.TrimSpace(toString(raw)), 10, 64); err == nil {
			return n
		}
		return current
	default:
		return toString(raw)
	}
}
