// Package service 审计日志应用服务
package service

import (
	"context"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
)

// AuditLogService 审计日志应用服务
type AuditLogService struct {
	auditRepo repository.AuditLogRepository
}

// NewAuditLogService 创建审计日志服务
func NewAuditLogService(auditRepo repository.AuditLogRepository) *AuditLogService {
	return &AuditLogService{auditRepo: auditRepo}
}

// List 查询审计日志列表
func (s *AuditLogService) List(ctx context.Context, query *entity.AuditLogQuery) (*repository.PaginatedResult, error) {
	return s.auditRepo.List(ctx, query)
}

// GetByID 根据ID获取审计日志
func (s *AuditLogService) GetByID(ctx context.Context, id string) (*entity.AuditLog, error) {
	return s.auditRepo.FindByID(ctx, id)
}

// GetStats 获取统计信息
func (s *AuditLogService) GetStats(ctx context.Context, startTime, endTime *time.Time) (*entity.AuditLogStats, error) {
	return s.auditRepo.GetStats(ctx, startTime, endTime)
}

// GetUserActivity 获取用户活动统计
func (s *AuditLogService) GetUserActivity(ctx context.Context, userID string, days int) ([]*repository.UserActivityItem, error) {
	return s.auditRepo.GetUserActivity(ctx, userID, days)
}

// GetModuleStats 获取模块统计
func (s *AuditLogService) GetModuleStats(ctx context.Context, startTime, endTime *time.Time) ([]*repository.ModuleStatItem, error) {
	return s.auditRepo.GetModuleStats(ctx, startTime, endTime)
}

// CleanupOldLogs 清理旧日志
func (s *AuditLogService) CleanupOldLogs(ctx context.Context, before time.Time) (int64, error) {
	return s.auditRepo.CleanupOldLogs(ctx, before)
}

// GetRecentLogs 获取最近的日志
func (s *AuditLogService) GetRecentLogs(ctx context.Context, limit int) ([]entity.AuditLog, error) {
	return s.auditRepo.GetRecentLogs(ctx, limit)
}

// LogLogin 记录登录日志
func (s *AuditLogService) LogLogin(ctx context.Context, userID, username, role, ip, userAgent string, success bool, errMsg string) {
	log := entity.NewAuditLog("认证", "登录", entity.AuditLogTypeLogin)
	log.UserID = userID
	log.Username = username
	log.UserRole = role
	log.RequestIP = ip
	log.UserAgent = userAgent
	
	if success {
		log.MarkSuccess()
	} else {
		log.MarkFailure(errMsg)
	}

	if err := s.auditRepo.Create(ctx, log); err != nil {
		logger.Error("Failed to create login audit log", logger.Err(err))
	}
}

// LogOperation 记录操作日志
func (s *AuditLogService) LogOperation(ctx context.Context, userID, username, module, action string, logType entity.AuditLogType, description string, success bool) {
	log := entity.NewAuditLog(module, action, logType)
	log.UserID = userID
	log.Username = username
	log.Description = description
	
	if success {
		log.MarkSuccess()
	} else {
		log.MarkFailure("")
	}

	if err := s.auditRepo.Create(ctx, log); err != nil {
		logger.Error("Failed to create operation audit log", logger.Err(err))
	}
}
