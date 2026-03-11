package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
)

// DashboardService 仪表盘服务
type DashboardService struct {
	userRepo    repository.UserRepository
	orgRepo     repository.OrganizationRepository
	mpRepo      repository.MissingPersonRepository
	taskRepo    repository.TaskRepository
	dialectRepo repository.DialectRepository
	fileRepo    repository.FileRepository
}

// NewDashboardService 创建仪表盘服务
func NewDashboardService(
	userRepo repository.UserRepository,
	orgRepo repository.OrganizationRepository,
	mpRepo repository.MissingPersonRepository,
	taskRepo repository.TaskRepository,
	dialectRepo repository.DialectRepository,
	fileRepo repository.FileRepository,
) *DashboardService {
	return &DashboardService{
		userRepo:    userRepo,
		orgRepo:     orgRepo,
		mpRepo:      mpRepo,
		taskRepo:    taskRepo,
		dialectRepo: dialectRepo,
		fileRepo:    fileRepo,
	}
}

// DashboardStats 仪表盘统计
type DashboardStats struct {
	Users          UserStats          `json:"users"`
	Organizations  OrgStats           `json:"organizations"`
	MissingPersons MissingPersonStats `json:"missing_persons"`
	Tasks          TaskStats          `json:"tasks"`
	Dialects       DialectStats       `json:"dialects"`
	Files          FileStats          `json:"files"`
	RecentActivity []Activity         `json:"recent_activity"`
}

// UserStats 用户统计
type UserStats struct {
	Total    int64 `json:"total"`
	Active   int64 `json:"active"`
	NewToday int64 `json:"new_today"`
	NewWeek  int64 `json:"new_week"`
	NewMonth int64 `json:"new_month"`
}

// OrgStats 组织统计
type OrgStats struct {
	Total     int64 `json:"total"`
	Provinces int64 `json:"provinces"`
	Cities    int64 `json:"cities"`
	Districts int64 `json:"districts"`
}

// MissingPersonStats 走失人员统计
type MissingPersonStats struct {
	Total     int64 `json:"total"`
	Missing   int64 `json:"missing"`
	Searching int64 `json:"searching"`
	Found     int64 `json:"found"`
	Reunited  int64 `json:"reunited"`
	NewToday  int64 `json:"new_today"`
	NewWeek   int64 `json:"new_week"`
}

// TaskStats 任务统计
type TaskStats struct {
	Total      int64 `json:"total"`
	Pending    int64 `json:"pending"`
	Processing int64 `json:"processing"`
	Completed  int64 `json:"completed"`
	Overdue    int64 `json:"overdue"`
}

// DialectStats 方言统计
type DialectStats struct {
	Total    int64 `json:"total"`
	Featured int64 `json:"featured"`
	Plays    int64 `json:"plays"`
	Likes    int64 `json:"likes"`
}

// FileStats 文件统计
type FileStats struct {
	TotalCount int64 `json:"total_count"`
	TotalSize  int64 `json:"total_size"`
}

// Activity 活动记录
type Activity struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Title      string    `json:"title"`
	UserID     string    `json:"user_id"`
	UserName   string    `json:"user_name"`
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// GetDashboardStats 获取仪表盘统计
func (s *DashboardService) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// 用户统计
	userTotal, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user count: %w", err)
	}
	stats.Users.Total = userTotal

	// 走失人员统计
	mpStats, err := s.mpRepo.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get missing person stats: %w", err)
	}
	stats.MissingPersons.Total = mpStats.Total
	stats.MissingPersons.Missing = mpStats.Missing
	stats.MissingPersons.Searching = mpStats.Searching
	stats.MissingPersons.Found = mpStats.Found
	stats.MissingPersons.Reunited = mpStats.Reunited

	// 任务统计
	taskStats, err := s.taskRepo.GetStats(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get task stats: %w", err)
	}
	stats.Tasks.Total = taskStats.Total
	stats.Tasks.Pending = taskStats.Pending
	stats.Tasks.Processing = taskStats.Processing
	stats.Tasks.Completed = taskStats.Completed
	stats.Tasks.Overdue = taskStats.Overdue

	// 方言统计
	dialectStats, err := s.dialectRepo.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get dialect stats: %w", err)
	}
	stats.Dialects.Total = dialectStats.Total
	stats.Dialects.Featured = dialectStats.Featured
	stats.Dialects.Plays = dialectStats.TotalPlays
	stats.Dialects.Likes = dialectStats.TotalLikes

	// 文件统计
	fileStats, err := s.fileRepo.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get file stats: %w", err)
	}
	stats.Files.TotalCount = fileStats.TotalCount
	stats.Files.TotalSize = fileStats.TotalSize

	return stats, nil
}

// GetOverview 获取概览数据
func (s *DashboardService) GetOverview(ctx context.Context) (map[string]interface{}, error) {
	overview := make(map[string]interface{})

	// 关键数据
	userCount, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user count: %w", err)
	}
	mpStats, err := s.mpRepo.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get missing person stats: %w", err)
	}
	taskStats, err := s.taskRepo.GetStats(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get task stats: %w", err)
	}

	overview["total_users"] = userCount
	overview["total_cases"] = mpStats.Total
	overview["active_cases"] = mpStats.Missing + mpStats.Searching
	overview["resolved_cases"] = mpStats.Found + mpStats.Reunited
	overview["pending_tasks"] = taskStats.Pending
	overview["processing_tasks"] = taskStats.Processing

	return overview, nil
}

// GetTrendData 获取趋势数据
func (s *DashboardService) GetTrendData(ctx context.Context, days int) ([]TrendData, error) {
	if days <= 0 {
		days = 7
	}
	if days > 90 {
		days = 90
	}

	trends := make([]TrendData, 0, days)
	now := time.Now()

	for i := days - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		dayStart := dateStr + " 00:00:00"
		dayEnd := dateStr + " 23:59:59"

		trend := TrendData{Date: dateStr}

		// 当日新增案件
		newCases, err := s.mpRepo.CountByDateRange(ctx, dayStart, dayEnd)
		if err != nil {
			return nil, fmt.Errorf("failed to count new cases for %s: %w", dateStr, err)
		}
		trend.NewCases = newCases

		// 当日解决案件
		resolvedCases, err := s.mpRepo.CountResolvedByDateRange(ctx, dayStart, dayEnd)
		if err != nil {
			return nil, fmt.Errorf("failed to count resolved cases for %s: %w", dateStr, err)
		}
		trend.ResolvedCases = resolvedCases

		// 当日新增任务
		newTasks, err := s.taskRepo.CountByDateRange(ctx, dayStart, dayEnd)
		if err != nil {
			return nil, fmt.Errorf("failed to count new tasks for %s: %w", dateStr, err)
		}
		trend.NewTasks = newTasks

		// 当日完成任务
		completedTasks, err := s.taskRepo.CountCompletedByDateRange(ctx, dayStart, dayEnd)
		if err != nil {
			return nil, fmt.Errorf("failed to count completed tasks for %s: %w", dateStr, err)
		}
		trend.CompletedTasks = completedTasks

		// 当日新增用户
		newUsers, err := s.userRepo.CountByDateRange(ctx, dayStart, dayEnd)
		if err != nil {
			return nil, fmt.Errorf("failed to count new users for %s: %w", dateStr, err)
		}
		trend.NewUsers = newUsers

		trends = append(trends, trend)
	}

	return trends, nil
}

// TrendData 趋势数据
type TrendData struct {
	Date           string `json:"date"`
	NewCases       int64  `json:"new_cases"`
	ResolvedCases  int64  `json:"resolved_cases"`
	NewTasks       int64  `json:"new_tasks"`
	CompletedTasks int64  `json:"completed_tasks"`
	NewUsers       int64  `json:"new_users"`
}
