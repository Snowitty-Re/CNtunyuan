package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
)

// AuditService 审计服务
type AuditService struct {
	repo   repository.AuditLogRepository
	config *entity.AuditLogConfig
}

// NewAuditService 创建审计服务
func NewAuditService(repo repository.AuditLogRepository) *AuditService {
	return &AuditService{
		repo:   repo,
		config: entity.DefaultAuditLogConfig(),
	}
}

// SetConfig 设置配置
func (s *AuditService) SetConfig(config *entity.AuditLogConfig) {
	s.config = config
}

// ShouldAudit 检查是否应该审计
func (s *AuditService) ShouldAudit(path string) bool {
	return s.config.ShouldAudit(path)
}

// Log 记录审计日志
func (s *AuditService) Log(ctx context.Context, log *entity.AuditLog) error {
	if !s.config.Enabled {
		return nil
	}
	
	// 脱敏敏感数据
	if log.OldValues != nil {
		log.OldValues = entity.MaskSensitiveData(log.OldValues)
	}
	if log.NewValues != nil {
		log.NewValues = entity.MaskSensitiveData(log.NewValues)
	}
	
	// 计算变更字段
	if log.Action == entity.AuditActionUpdate {
		log.CalculateDelta()
	}
	
	// 异步记录日志（不阻塞主流程）
	go func() {
		bgCtx := context.Background()
		if err := s.repo.Create(bgCtx, log); err != nil {
			logger.Error("Failed to create audit log", logger.Err(err))
		}
	}()
	
	return nil
}

// LogFromRequest 从HTTP请求记录审计日志
func (s *AuditService) LogFromRequest(ctx context.Context, req *http.Request, resp *http.Response, userID, orgID string, duration int64) error {
	if !s.config.Enabled {
		return nil
	}
	
	action := s.parseAction(req.Method)
	resource := s.parseResource(req.URL.Path)
	
	log := entity.NewAuditLog(userID, orgID, action, resource).
		SetRequestInfo(req.Method, req.URL.String(), s.getClientIP(req), req.UserAgent()).
		SetTraceID(s.getTraceID(req)).
		SetDuration(duration)
	
	if resp != nil {
		log.SetStatus(resp.StatusCode)
	}
	
	return s.Log(ctx, log)
}

// LogCreate 记录创建操作
func (s *AuditService) LogCreate(ctx context.Context, userID, orgID, resource, resourceID, resourceName string, newValues map[string]interface{}) error {
	log := entity.NewAuditLog(userID, orgID, entity.AuditActionCreate, resource).
		SetResourceID(resourceID).
		SetResourceName(resourceName).
		SetDescription(fmt.Sprintf("创建%s: %s", resource, resourceName)).
		SetNewValues(newValues)
	
	return s.Log(ctx, log)
}

// LogUpdate 记录更新操作
func (s *AuditService) LogUpdate(ctx context.Context, userID, orgID, resource, resourceID, resourceName string, oldValues, newValues map[string]interface{}) error {
	log := entity.NewAuditLog(userID, orgID, entity.AuditActionUpdate, resource).
		SetResourceID(resourceID).
		SetResourceName(resourceName).
		SetDescription(fmt.Sprintf("更新%s: %s", resource, resourceName)).
		SetOldValues(oldValues).
		SetNewValues(newValues)
	
	return s.Log(ctx, log)
}

// LogDelete 记录删除操作
func (s *AuditService) LogDelete(ctx context.Context, userID, orgID, resource, resourceID, resourceName string, oldValues map[string]interface{}) error {
	log := entity.NewAuditLog(userID, orgID, entity.AuditActionDelete, resource).
		SetResourceID(resourceID).
		SetResourceName(resourceName).
		SetDescription(fmt.Sprintf("删除%s: %s", resource, resourceName)).
		SetOldValues(oldValues)
	
	return s.Log(ctx, log)
}

// LogLogin 记录登录操作
func (s *AuditService) LogLogin(ctx context.Context, userID, orgID, ip, userAgent string, success bool, errMsg string) error {
	action := entity.AuditActionLogin
	log := entity.NewAuditLog(userID, orgID, action, "session").
		SetRequestInfo("POST", "/api/v1/auth/login", ip, userAgent).
		SetDescription("用户登录")
	
	if success {
		log.SetStatus(http.StatusOK)
	} else {
		log.SetStatus(http.StatusUnauthorized).
			SetError(errMsg)
	}
	
	return s.Log(ctx, log)
}

// LogLogout 记录登出操作
func (s *AuditService) LogLogout(ctx context.Context, userID, orgID, ip, userAgent string) error {
	log := entity.NewAuditLog(userID, orgID, entity.AuditActionLogout, "session").
		SetRequestInfo("POST", "/api/v1/auth/logout", ip, userAgent).
		SetDescription("用户登出")
	
	return s.Log(ctx, log)
}

// LogApproval 记录审批操作
func (s *AuditService) LogApproval(ctx context.Context, userID, orgID, resource, resourceID, resourceName, decision, comment string) error {
	var action entity.AuditAction
	if decision == "approved" {
		action = entity.AuditActionApprove
	} else {
		action = entity.AuditActionReject
	}
	
	log := entity.NewAuditLog(userID, orgID, action, resource).
		SetResourceID(resourceID).
		SetResourceName(resourceName).
		SetDescription(fmt.Sprintf("%s: %s", action, resourceName)).
		AddExtra("decision", decision).
		AddExtra("comment", comment)
	
	return s.Log(ctx, log)
}

// parseAction 解析HTTP方法为审计操作
func (s *AuditService) parseAction(method string) entity.AuditAction {
	switch strings.ToUpper(method) {
	case http.MethodPost:
		return entity.AuditActionCreate
	case http.MethodPut, http.MethodPatch:
		return entity.AuditActionUpdate
	case http.MethodDelete:
		return entity.AuditActionDelete
	case http.MethodGet:
		return entity.AuditActionQuery
	default:
		return entity.AuditActionOther
	}
}

// parseResource 从URL路径解析资源类型
func (s *AuditService) parseResource(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 && parts[0] == "api" && parts[1] == "v1" {
		return parts[2]
	}
	return "unknown"
}

// getClientIP 获取客户端IP
func (s *AuditService) getClientIP(req *http.Request) string {
	// 检查X-Forwarded-For头
	xff := req.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	
	// 检查X-Real-IP头
	xri := req.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}
	
	// 使用RemoteAddr
	return req.RemoteAddr
}

// getTraceID 获取追踪ID
func (s *AuditService) getTraceID(req *http.Request) string {
	return req.Header.Get("X-Request-ID")
}

// List 查询审计日志列表
func (s *AuditService) List(ctx context.Context, req *dto.AuditLogListRequest) (*dto.AuditLogListResponse, error) {
	query := &entity.AuditLogQuery{
		UserID:     req.UserID,
		OrgID:      req.OrgID,
		Action:     entity.AuditAction(req.Action),
		Resource:   req.Resource,
		ResourceID: req.ResourceID,
		IPAddress:  req.IPAddress,
		TraceID:    req.TraceID,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}
	
	if req.StartTime != "" {
		if startTime, err := time.Parse(time.RFC3339, req.StartTime); err == nil {
			query.StartTime = &startTime
		}
	}
	if req.EndTime != "" {
		if endTime, err := time.Parse(time.RFC3339, req.EndTime); err == nil {
			query.EndTime = &endTime
		}
	}
	
	logs, total, err := s.repo.List(ctx, query)
	if err != nil {
		logger.Error("Failed to list audit logs", logger.Err(err))
		return nil, err
	}
	
	items := make([]dto.AuditLogResponse, len(logs))
	for i, log := range logs {
		items[i] = dto.ToAuditLogResponse(log)
	}
	
	return &dto.AuditLogListResponse{
		List:     items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetByID 根据ID获取审计日志
func (s *AuditService) GetByID(ctx context.Context, id string) (*dto.AuditLogResponse, error) {
	log, err := s.repo.FindByID(ctx, id)
	if err != nil {
		logger.Error("Failed to find audit log", logger.String("id", id), logger.Err(err))
		return nil, err
	}
	
	if log == nil {
		return nil, nil
	}
	
	resp := dto.ToAuditLogResponse(log)
	return &resp, nil
}

// GetStats 获取审计统计
func (s *AuditService) GetStats(ctx context.Context, orgID string, days int) (*dto.AuditLogStatsResponse, error) {
	var startTime, endTime *time.Time
	
	if days > 0 {
		st := time.Now().AddDate(0, 0, -days)
		startTime = &st
		et := time.Now()
		endTime = &et
	}
	
	stats, err := s.repo.GetStats(ctx, orgID, startTime, endTime)
	if err != nil {
		logger.Error("Failed to get audit stats", logger.Err(err))
		return nil, err
	}
	
	actionStats, err := s.repo.GetActionStats(ctx, orgID, startTime, endTime)
	if err != nil {
		logger.Error("Failed to get action stats", logger.Err(err))
		return nil, err
	}
	
	resourceStats, err := s.repo.GetResourceStats(ctx, orgID, startTime, endTime)
	if err != nil {
		logger.Error("Failed to get resource stats", logger.Err(err))
		return nil, err
	}
	
	return &dto.AuditLogStatsResponse{
		TotalCount:    stats.TotalCount,
		TodayCount:    stats.TodayCount,
		ActionStats:   actionStats,
		ResourceStats: resourceStats,
		PeriodDays:    days,
	}, nil
}

// GetUserActivity 获取用户活动统计
func (s *AuditService) GetUserActivity(ctx context.Context, userID string, days int) (map[string]interface{}, error) {
	return s.repo.GetUserActivityStats(ctx, userID, days)
}

// CleanupOldLogs 清理过期日志
func (s *AuditService) CleanupOldLogs(ctx context.Context, days int) (int64, error) {
	if days <= 0 {
		days = s.config.RetentionDays
	}
	
	before := time.Now().AddDate(0, 0, -days)
	deleted, err := s.repo.CleanupOldLogs(ctx, before)
	if err != nil {
		logger.Error("Failed to cleanup old logs", logger.Err(err))
		return 0, err
	}
	
	logger.Info("Audit logs cleaned up", logger.Int64("deleted", deleted), logger.String("before", before.Format(time.RFC3339)))
	return deleted, nil
}

// GetRetentionStats 获取保留统计
func (s *AuditService) GetRetentionStats(ctx context.Context) (map[string]interface{}, error) {
	return s.repo.GetRetentionStats(ctx, s.config.RetentionDays)
}

// ReadRequestBody 读取请求体（用于审计）
func (s *AuditService) ReadRequestBody(req *http.Request) map[string]interface{} {
	if req.Body == nil || req.ContentLength == 0 {
		return nil
	}
	
	// 限制读取大小
	limit := int64(s.config.MaxBodySize)
	if req.ContentLength > limit {
		return nil
	}
	
	body, err := io.ReadAll(io.LimitReader(req.Body, limit))
	if err != nil {
		return nil
	}
	
	// 恢复请求体
	req.Body = io.NopCloser(strings.NewReader(string(body)))
	
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil
	}
	
	return data
}

// ReadResponseBody 读取响应体（用于审计）
func (s *AuditService) ReadResponseBody(body []byte) map[string]interface{} {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil
	}
	return data
}
