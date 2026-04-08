// Package task 简化版定时任务调度器（使用标准库）
package task

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	appservice "github.com/Snowitty-Re/CNtunyuan/internal/application/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	taskRepo  repository.TaskRepository
	mpRepo    repository.MissingPersonRepository
	userRepo  repository.UserRepository
	auditRepo repository.AuditLogRepository
	authzSvc  *appservice.AuthorizationService
	dbConfig  *config.DatabaseConfig
	backupCfg *config.BackupConfig
	systemCfg *config.SystemConfig
	ticker    *time.Ticker
	stopChan  chan struct{}
	isRunning atomic.Bool
	stopOnce  sync.Once
}

// NewScheduler 创建定时任务调度器
func NewScheduler(
	taskRepo repository.TaskRepository,
	mpRepo repository.MissingPersonRepository,
	userRepo repository.UserRepository,
	auditRepo repository.AuditLogRepository,
	authzSvc *appservice.AuthorizationService,
) *Scheduler {
	return &Scheduler{
		taskRepo:  taskRepo,
		mpRepo:    mpRepo,
		userRepo:  userRepo,
		auditRepo: auditRepo,
		authzSvc:  authzSvc,
		stopChan:  make(chan struct{}),
	}
}

// SetDatabaseConfig 设置数据库配置（用于备份）
func (s *Scheduler) SetDatabaseConfig(dbCfg *config.DatabaseConfig, backupCfg *config.BackupConfig) {
	s.dbConfig = dbCfg
	s.backupCfg = backupCfg
}

// SetSystemConfig 设置系统配置（用于审批申请超时处理）
func (s *Scheduler) SetSystemConfig(systemCfg *config.SystemConfig) {
	s.systemCfg = systemCfg
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

			// 每10分钟检查一次权限策略审批申请超时
			if time.Now().Minute()%10 == 0 {
				s.rejectExpiredAuthzPolicyRequests()
			}

		case <-s.stopChan:
			return
		}
	}
}

func (s *Scheduler) rejectExpiredAuthzPolicyRequests() {
	if s.authzSvc == nil || s.systemCfg == nil {
		return
	}
	if !s.systemCfg.AuthzPolicyChangeRequiresApproval {
		return
	}
	expireHours := s.systemCfg.AuthzPolicyRequestExpireHours
	if expireHours <= 0 {
		return
	}

	ctx := context.Background()
	updated, err := s.authzSvc.AutoRejectExpiredPolicyChangeRequests(ctx, expireHours, "system:auto-timeout")
	if err != nil {
		logger.Error("Failed to auto reject expired authz policy requests", logger.Err(err))
		return
	}
	if updated > 0 {
		logger.Info("Auto rejected expired authz policy requests", logger.Int64("count", updated))
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
			s.cleanupOldBackups()

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
				UserID:    task.CreatorID,
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
		logger.Int64("total_cases", func() int64 {
			if mpStats != nil {
				return mpStats.Total
			}
			return 0
		}()),
		logger.Int64("total_tasks", func() int64 {
			if taskStats != nil {
				return taskStats.Total
			}
			return 0
		}()),
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

// backupData 执行数据库备份
func (s *Scheduler) backupData() {
	if s.dbConfig == nil || s.backupCfg == nil || !s.backupCfg.Enabled {
		logger.Debug("Database backup is disabled or not configured, skipping")
		return
	}

	logger.Info("Starting database backup")

	// 确保备份目录存在
	backupDir := s.backupCfg.BackupDir
	if backupDir == "" {
		backupDir = "./backups"
	}
	if err := os.MkdirAll(backupDir, 0750); err != nil {
		logger.Error("Failed to create backup directory", logger.Err(err), logger.String("dir", backupDir))
		return
	}

	// 生成备份文件名
	timestamp := time.Now().Format("20060102_150405")
	var backupFile string
	var cmd *exec.Cmd

	switch s.dbConfig.Type {
	case config.DatabaseTypePostgres:
		backupFile = filepath.Join(backupDir, fmt.Sprintf("cntuanyuan_pg_%s.sql.gz", timestamp))
		cmd = s.buildPostgresBackupCmd(backupFile)
	case config.DatabaseTypeMySQL:
		backupFile = filepath.Join(backupDir, fmt.Sprintf("cntuanyuan_mysql_%s.sql.gz", timestamp))
		cmd = s.buildMySQLBackupCmd(backupFile)
	default:
		logger.Warn("Unknown database type for backup", logger.String("type", string(s.dbConfig.Type)))
		return
	}

	if cmd == nil {
		logger.Warn("Backup command tool not found, skipping backup")
		return
	}

	// 执行备份命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("Database backup failed",
			logger.Err(err),
			logger.String("output", string(output)),
			logger.String("file", backupFile),
		)
		// 清理可能的部分文件
		os.Remove(backupFile)
		return
	}

	// 验证备份文件
	info, err := os.Stat(backupFile)
	if err != nil || info.Size() == 0 {
		logger.Error("Backup file is empty or missing", logger.String("file", backupFile))
		os.Remove(backupFile)
		return
	}

	logger.Info("Database backup completed successfully",
		logger.String("file", backupFile),
		logger.Int64("size_bytes", info.Size()),
	)
}

// buildPostgresBackupCmd 构建PostgreSQL备份命令
// 通过 PGPASSWORD 环境变量传递密码，避免密码出现在 shell 命令中（防止 shell 注入和进程列表泄露）
func (s *Scheduler) buildPostgresBackupCmd(outputFile string) *exec.Cmd {
	pgDump := "pg_dump"
	if runtime.GOOS == "windows" {
		pgDump = "pg_dump.exe"
	}

	// 检查pg_dump是否可用
	if _, err := exec.LookPath(pgDump); err != nil {
		logger.Warn("pg_dump not found in PATH, backup skipped")
		return nil
	}

	sslMode := s.dbConfig.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	// DSN 中不包含密码；密码通过 PGPASSWORD 环境变量传递
	dsn := fmt.Sprintf("postgresql://%s@%s:%d/%s?sslmode=%s",
		s.dbConfig.User,
		s.dbConfig.Host, s.dbConfig.Port,
		s.dbConfig.Database, sslMode,
	)
	pgpassEnv := "PGPASSWORD=" + s.dbConfig.Password

	// pg_dump | gzip > file
	if isGzipAvailable() {
		shellCmd := fmt.Sprintf("%s --dbname=%q --no-owner --no-privileges --clean --if-exists | gzip > %q",
			pgDump, dsn, outputFile)
		cmd := shellExec(shellCmd)
		if cmd != nil {
			cmd.Env = append(os.Environ(), pgpassEnv)
		}
		return cmd
	}

	// 不压缩
	outputFile = strings.TrimSuffix(outputFile, ".gz")
	cmd := exec.Command(pgDump,
		"--dbname="+dsn,
		"--no-owner",
		"--no-privileges",
		"--clean",
		"--if-exists",
		"--file="+outputFile,
	)
	cmd.Env = append(os.Environ(), pgpassEnv)
	return cmd
}

// buildMySQLBackupCmd 构建MySQL备份命令
// 通过 MYSQL_PWD 环境变量传递密码，避免密码出现在 shell 命令和进程列表中
func (s *Scheduler) buildMySQLBackupCmd(outputFile string) *exec.Cmd {
	mysqldump := "mysqldump"
	if runtime.GOOS == "windows" {
		mysqldump = "mysqldump.exe"
	}

	// 检查mysqldump是否可用
	if _, err := exec.LookPath(mysqldump); err != nil {
		logger.Warn("mysqldump not found in PATH, backup skipped")
		return nil
	}

	// 不在 args 中传递密码；通过 MYSQL_PWD 环境变量传递
	mysqlPwdEnv := "MYSQL_PWD=" + s.dbConfig.Password
	args := []string{
		fmt.Sprintf("--host=%s", s.dbConfig.Host),
		fmt.Sprintf("--port=%d", s.dbConfig.Port),
		fmt.Sprintf("--user=%s", s.dbConfig.User),
		"--single-transaction",
		"--routines",
		"--triggers",
		"--set-gtid-purged=OFF",
		s.dbConfig.Database,
	}

	if isGzipAvailable() {
		shellCmd := fmt.Sprintf("%s %s | gzip > %q",
			mysqldump, strings.Join(args, " "), outputFile)
		cmd := shellExec(shellCmd)
		if cmd != nil {
			cmd.Env = append(os.Environ(), mysqlPwdEnv)
		}
		return cmd
	}

	outputFile = strings.TrimSuffix(outputFile, ".gz")
	args = append(args, fmt.Sprintf("--result-file=%s", outputFile))
	cmd := exec.Command(mysqldump, args...)
	cmd.Env = append(os.Environ(), mysqlPwdEnv)
	return cmd
}

// cleanupOldBackups 清理过期备份文件
func (s *Scheduler) cleanupOldBackups() {
	if s.backupCfg == nil || !s.backupCfg.Enabled {
		return
	}

	backupDir := s.backupCfg.BackupDir
	if backupDir == "" {
		backupDir = "./backups"
	}

	retention := s.backupCfg.Retention
	if retention <= 0 {
		retention = 7
	}

	cutoff := time.Now().AddDate(0, 0, -retention)

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		logger.Error("Failed to read backup directory", logger.Err(err))
		return
	}

	// 收集并按时间排序
	type backupFile struct {
		name    string
		modTime time.Time
	}
	var backups []backupFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "cntuanyuan_") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		backups = append(backups, backupFile{name: name, modTime: info.ModTime()})
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].modTime.Before(backups[j].modTime)
	})

	deleted := 0
	for _, b := range backups {
		if b.modTime.Before(cutoff) {
			path := filepath.Join(backupDir, b.name)
			if err := os.Remove(path); err != nil {
				logger.Error("Failed to delete old backup", logger.Err(err), logger.String("file", path))
			} else {
				deleted++
			}
		}
	}

	if deleted > 0 {
		logger.Info("Old backups cleaned up", logger.Int("deleted", deleted), logger.Int("retention_days", retention))
	}
}

// isGzipAvailable 检查gzip是否可用
func isGzipAvailable() bool {
	_, err := exec.LookPath("gzip")
	return err == nil
}

// shellExec 通过shell执行命令（支持管道）
func shellExec(command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("cmd", "/C", command)
	}
	return exec.Command("sh", "-c", command)
}
