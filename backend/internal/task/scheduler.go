// Package task 简化版定时任务调度器（使用标准库）
package task

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	taskRepo   repository.TaskRepository
	mpRepo     repository.MissingPersonRepository
	userRepo   repository.UserRepository
	auditRepo  repository.AuditLogRepository
	ticker     *time.Ticker
	stopChan   chan struct{}
	isRunning  atomic.Bool
	stopOnce   sync.Once
}

// NewScheduler 创建定时任务调度器
func NewScheduler(
	taskRepo repository.TaskRepository,
	mpRepo repository.MissingPersonRepository,
	userRepo repository.UserRepository,
	auditRepo repository.AuditLogRepository,
) *Scheduler {
	return &Scheduler{
		taskRepo:  taskRepo,
		mpRepo:    mpRepo,
		userRepo:  userRepo,
		auditRepo: auditRepo,
		stopChan:  make(chan struct{}),
	}
}

// Start 启动定时任务
func (s *Scheduler) Start() error {
	if s.isRunning.Load() {
		return nil
	}

	// 创建每分钟触发的ticker
	s.ticker = time.NewTicker(1 * time.Minute)
	s.isRunning.Store(true)

	// 启动定时任务循环
	go s.run()

	// 启动每日任务（在3点执行清理和备份）
	go s.runDailyTasks()

	logger.Info("Task scheduler started")
	return nil
}

// Stop 停止定时任务
func (s *Scheduler) Stop() {
	s.stopOnce.Do(func() {
		s.isRunning.Store(false)
		if s.ticker != nil {
			s.ticker.Stop()
		}
		close(s.stopChan)
		logger.Info("Task scheduler stopped")
	})
}

// IsRunning 是否正在运行
func (s *Scheduler) IsRunning() bool {
	return s.isRunning.Load()
}

// run 定时任务循环
func (s *Scheduler) run() {
	for {
		select {
		case <-s.ticker.C:
			// 每分钟检查逾期任务
			s.checkOverdueTasks()
			
			// 每小时更新统计（整点时）
			if time.Now().Minute() == 0 {
				s.updateStatistics()
			}
			
		case <-s.stopChan:
			return
		}
	}
}

// runDailyTasks 每日定时任务
func (s *Scheduler) runDailyTasks() {
	for {
		now := time.Now()
		// 计算到下一个3点的时长
		nextRun := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
		if nextRun.Before(now) {
			nextRun = nextRun.Add(24 * time.Hour)
		}
		
		timer := time.NewTimer(nextRun.Sub(now))
		
		select {
		case <-timer.C:
			// 执行每日任务
			s.cleanupOldLogs()
			s.backupData()
			
		case <-s.stopChan:
			timer.Stop()
			return
		}
	}
}

// checkOverdueTasks 检查逾期任务
func (s *Scheduler) checkOverdueTasks() {
	ctx := context.Background()

	// 获取所有活跃且逾期的任务
	overdueTasks, err := s.taskRepo.FindOverdueTasks(ctx)
	if err != nil {
		logger.Error("Failed to find overdue tasks", logger.Err(err))
		return
	}

	count := 0
	for _, task := range overdueTasks {
		if task.CheckAndUpdateOverdue() {
			if err := s.taskRepo.Update(ctx, &task); err != nil {
				logger.Error("Failed to update overdue task",
					logger.String("task_id", task.ID),
					logger.Err(err),
				)
				continue
			}
			count++

			// 记录任务日志
			log := &entity.TaskLog{
				TaskID:    task.ID,
				Action:    "auto_overdue",
				Content:   "任务自动标记为逾期",
				NewStatus: "overdue",
			}
			if err := s.taskRepo.AddLog(ctx, log); err != nil {
				logger.Error("Failed to add overdue log", logger.Err(err))
			}
		}
	}

	if count > 0 {
		logger.Info("Overdue tasks check completed", logger.Int("count", count))
	}
}

// updateStatistics 更新统计数据
func (s *Scheduler) updateStatistics() {
	logger.Info("Running statistics update")
	ctx := context.Background()

	userCount, err := s.userRepo.Count(ctx)
	if err != nil {
		logger.Error("Failed to count users", logger.Err(err))
	}

	mpStats, err := s.mpRepo.GetStats(ctx)
	if err != nil {
		logger.Error("Failed to get missing person stats", logger.Err(err))
	}

	taskStats, err := s.taskRepo.GetStats(ctx, "")
	if err != nil {
		logger.Error("Failed to get task stats", logger.Err(err))
	}

	logger.Info("Statistics update completed",
		logger.Int64("total_users", userCount),
		logger.Int64("total_cases", func() int64 { if mpStats != nil { return mpStats.Total }; return 0 }()),
		logger.Int64("total_tasks", func() int64 { if taskStats != nil { return taskStats.Total }; return 0 }()),
	)
}

// cleanupOldLogs 清理旧日志
func (s *Scheduler) cleanupOldLogs() {
	logger.Info("Running old logs cleanup")
	ctx := context.Background()

	// 清理90天前的审计日志
	before := time.Now().AddDate(0, 0, -90)
	count, err := s.auditRepo.CleanupOldLogs(ctx, before)
	if err != nil {
		logger.Error("Failed to cleanup old logs", logger.Err(err))
		return
	}

	logger.Info("Old logs cleanup completed", logger.Int64("deleted", count))
}

// backupData 备份数据
func (s *Scheduler) backupData() {
	logger.Info("Data backup reminder: application-level backup is not implemented. " +
		"Use infrastructure-level tools (e.g., pg_dump for PostgreSQL, mysqldump for MySQL) " +
		"configured via cron or your cloud provider's backup service.")
}
