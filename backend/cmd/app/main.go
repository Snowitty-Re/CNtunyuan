// @title 团圆寻亲志愿者系统 API
// @version 1.0.0
// @description 团圆寻亲志愿者系统后端 API 文档
// @termsOfService https://github.com/Snowitty-Re/CNtunyuan

// @contact.name CNtunyuan Team
// @contact.url https://github.com/Snowitty-Re/CNtunyuan
// @contact.email support@cntunyuan.org

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description 请输入 JWT Token，格式：Bearer {token}

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/Snowitty-Re/CNtunyuan/docs"
	_ "gorm.io/gorm"
	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/di"
	"github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/database"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
)

func main() {
	// 初始化日志
	logCfg := &config.LogConfig{
		Level:  "info",
		Format: "console",
	}
	if err := logger.Init(logCfg); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// 加载配置（从 ./config 目录加载 config.yaml）
	cfg, err := config.LoadConfig("")
	if err != nil {
		logger.Error("Failed to load config", logger.Err(err))
		os.Exit(1)
	}

	// 验证配置
	if !config.ValidateAndPrint(cfg) {
		logger.Error("Configuration validation failed")
		os.Exit(1)
	}

	// 处理命令行参数
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-check-db":
			// 检查数据库连接和表结构
			if err := checkDatabase(cfg); err != nil {
				logger.Error("Database check failed", logger.Err(err))
				os.Exit(1)
			}
			return
		case "-migrate":
			// 使用 GORM AutoMigrate（开发环境用）
			logger.Warn("Using GORM AutoMigrate is deprecated. Please use SQL migration files instead.")
			if err := runAutoMigrate(cfg); err != nil {
				logger.Error("Migration failed", logger.Err(err))
				os.Exit(1)
			}
			return
		case "-seed":
			// 使用 GORM 插入种子数据（开发环境用）
			logger.Warn("Using GORM seed is deprecated. Please use SQL seed files instead.")
			if err := runSeed(cfg); err != nil {
				logger.Error("Seed failed", logger.Err(err))
				os.Exit(1)
			}
			return
		}
	}

	// 创建依赖容器
	container, err := di.NewContainer(cfg)
	if err != nil {
		logger.Error("Failed to create container", logger.Err(err))
		os.Exit(1)
	}

	// 启动 HTTP 服务器
	startServer(cfg, container)
}

// checkDatabase 检查数据库连接和表结构
func checkDatabase(cfg *config.Config) error {
	db, err := database.NewDatabase(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// 检查关键表是否存在
	tables := []string{
		"ty_organizations",
		"ty_users",
		"ty_permissions",
		"ty_missing_persons",
		"ty_tasks",
		"ty_dialects",
		"ty_files",
		"ty_audit_logs",
	}

	logger.Info("Checking database tables...")
	allExist := true
	for _, table := range tables {
		exists, err := database.TableExists(db, table)
		if err != nil {
			logger.Error("Failed to check table", logger.String("table", table), logger.Err(err))
			allExist = false
			continue
		}
		if exists {
			logger.Info("✓ Table exists", logger.String("table", table))
		} else {
			logger.Error("✗ Table missing", logger.String("table", table))
			allExist = false
		}
	}

	if allExist {
		logger.Info("Database check passed: all tables exist")
	} else {
		return fmt.Errorf("some tables are missing, please run SQL migration files first")
	}

	return nil
}

// runAutoMigrate 使用 GORM AutoMigrate（仅用于开发环境）
func runAutoMigrate(cfg *config.Config) error {
	db, err := database.NewDatabase(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	logger.Info("Starting database migration with GORM AutoMigrate...")
	if err := database.AutoMigrate(db); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	logger.Info("Database migration completed")
	return nil
}

// runSeed 使用 GORM 插入种子数据（仅用于开发环境）
func runSeed(cfg *config.Config) error {
	db, err := database.NewDatabase(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	logger.Info("Starting database seeding with GORM...")
	if err := database.Seed(db); err != nil {
		return fmt.Errorf("seed failed: %w", err)
	}
	logger.Info("Database seeding completed")
	return nil
}

// startServer 启动服务器
func startServer(cfg *config.Config, container *di.Container) {
	engine := container.Router.GetEngine()

	// 创建 HTTP 服务器
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	logger.Info("Starting server", logger.String("port", port))

	// 优雅关闭
	srv := engine
	go func() {
		if err := srv.Run(":" + port); err != nil {
			logger.Error("Server error", logger.Err(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// TODO: 实现优雅关闭逻辑
	_ = ctx

	logger.Info("Server stopped")
}
