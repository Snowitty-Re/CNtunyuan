// Package testutil 测试工具包
package testutil

import (
	"context"
	"errors"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/service"
)

// MockAuthService 模拟认证服务
type MockAuthService struct {
	ValidateTokenFunc func(ctx context.Context, token string) (*service.TokenClaims, error)
	GenerateTokenFunc func(ctx context.Context, user *entity.User) (*service.TokenPair, error)
}

// NewMockAuthService 创建模拟认证服务
func NewMockAuthService() *MockAuthService {
	return &MockAuthService{
		ValidateTokenFunc: func(ctx context.Context, token string) (*service.TokenClaims, error) {
			// 默认实现：根据token返回不同的结果
			switch token {
			case "valid-token":
				return &service.TokenClaims{
					UserID:   "user-123",
					Nickname: "Test User",
					Role:     string(entity.RoleVolunteer),
					OrgID:    "org-123",
				}, nil
			case "admin-token":
				return &service.TokenClaims{
					UserID:   "admin-123",
					Nickname: "Admin User",
					Role:     string(entity.RoleAdmin),
					OrgID:    "org-123",
				}, nil
			case "super-admin-token":
				return &service.TokenClaims{
					UserID:   "super-admin-123",
					Nickname: "Super Admin",
					Role:     string(entity.RoleSuperAdmin),
					OrgID:    "org-123",
				}, nil
			case "manager-token":
				return &service.TokenClaims{
					UserID:   "manager-123",
					Nickname: "Manager User",
					Role:     string(entity.RoleManager),
					OrgID:    "org-123",
				}, nil
			case "expired-token":
				return nil, errors.New("token expired")
			case "invalid-token":
				return nil, errors.New("invalid token")
			default:
				return nil, errors.New("invalid token")
			}
		},
	}
}

// ValidateToken 验证token
func (m *MockAuthService) ValidateToken(ctx context.Context, token string) (*service.TokenClaims, error) {
	return m.ValidateTokenFunc(ctx, token)
}

// GenerateTokenPair 生成token对（供测试使用）
func (m *MockAuthService) GenerateTokenPair(ctx context.Context, user *entity.User) (*service.TokenPair, error) {
	if m.GenerateTokenFunc != nil {
		return m.GenerateTokenFunc(ctx, user)
	}
	return &service.TokenPair{
		AccessToken:  "generated-token",
		RefreshToken: "refresh-token",
		ExpiresIn:    3600,
	}, nil
}

// RevokeToken 撤销token
func (m *MockAuthService) RevokeToken(ctx context.Context, token string) error {
	return nil
}

// MockTokenClaims 创建模拟的token claims
func MockTokenClaims(role entity.Role) *service.TokenClaims {
	return &service.TokenClaims{
		UserID:   "user-123",
		Nickname: "Test User",
		Role:     string(role),
		OrgID:    "org-123",
	}
}

// GenerateTestToken 生成测试用token
func GenerateTestToken(role entity.Role) string {
	tokens := map[entity.Role]string{
		entity.RoleVolunteer:    "valid-token",
		entity.RoleManager:      "manager-token",
		entity.RoleAdmin:        "admin-token",
		entity.RoleSuperAdmin:   "super-admin-token",
	}
	if token, ok := tokens[role]; ok {
		return token
	}
	return "valid-token"
}

// MockLogger 模拟日志记录器
type MockLogger struct {
	InfoLogs  []string
	WarnLogs  []string
	ErrorLogs []string
}

// NewMockLogger 创建模拟日志记录器
func NewMockLogger() *MockLogger {
	return &MockLogger{
		InfoLogs:  []string{},
		WarnLogs:  []string{},
		ErrorLogs: []string{},
	}
}

// Reset 重置日志
func (m *MockLogger) Reset() {
	m.InfoLogs = []string{}
	m.WarnLogs = []string{}
	m.ErrorLogs = []string{}
}

// MockAuditLogRepository 模拟审计日志仓储
type MockAuditLogRepository struct {
	Logs []*entity.AuditLog
}

// NewMockAuditLogRepository 创建模拟审计日志仓储
func NewMockAuditLogRepository() *MockAuditLogRepository {
	return &MockAuditLogRepository{
		Logs: make([]*entity.AuditLog, 0),
	}
}

// Create 创建审计日志
func (m *MockAuditLogRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	log.ID = GenerateTestID()
	log.CreatedAt = time.Now()
	m.Logs = append(m.Logs, log)
	return nil
}

// FindByID 根据ID查找
func (m *MockAuditLogRepository) FindByID(ctx context.Context, id string) (*entity.AuditLog, error) {
	for _, log := range m.Logs {
		if log.ID == id {
			return log, nil
		}
	}
	return nil, errors.New("audit log not found")
}

// List 列表查询 - 使用新的接口签名
func (m *MockAuditLogRepository) List(ctx context.Context, query *entity.AuditLogQuery) (*repository.PaginatedResult, error) {
	return &repository.PaginatedResult{
		List:     convertToAuditLogList(m.Logs),
		Total:    int64(len(m.Logs)),
		Page:     1,
		PageSize: 10,
	}, nil
}

// GetStats 获取统计信息
func (m *MockAuditLogRepository) GetStats(ctx context.Context, startTime, endTime *time.Time) (*entity.AuditLogStats, error) {
	return &entity.AuditLogStats{
		TotalCount: int64(len(m.Logs)),
	}, nil
}

// GetUserActivity 获取用户活动统计
func (m *MockAuditLogRepository) GetUserActivity(ctx context.Context, userID string, days int) ([]*repository.UserActivityItem, error) {
	return []*repository.UserActivityItem{}, nil
}

// GetModuleStats 获取模块统计
func (m *MockAuditLogRepository) GetModuleStats(ctx context.Context, startTime, endTime *time.Time) ([]*repository.ModuleStatItem, error) {
	return []*repository.ModuleStatItem{}, nil
}

// CleanupOldLogs 清理旧日志
func (m *MockAuditLogRepository) CleanupOldLogs(ctx context.Context, before time.Time) (int64, error) {
	return 0, nil
}

// GetRecentLogs 获取最近的日志
func (m *MockAuditLogRepository) GetRecentLogs(ctx context.Context, limit int) ([]entity.AuditLog, error) {
	result := make([]entity.AuditLog, 0, len(m.Logs))
	for _, log := range m.Logs {
		result = append(result, *log)
	}
	return result, nil
}

// convertToAuditLogList 转换为AuditLog列表
func convertToAuditLogList(logs []*entity.AuditLog) []entity.AuditLog {
	result := make([]entity.AuditLog, 0, len(logs))
	for _, log := range logs {
		result = append(result, *log)
	}
	return result
}

// GenerateTestID 生成测试ID
func GenerateTestID() string {
	return "test-id-" + time.Now().Format("20060102150405")
}
