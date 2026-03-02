package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// 加载配置
	cfg, err := config.LoadConfig(".")
	if err != nil {
		logger.Error("Failed to load config", logger.Err(err))
		os.Exit(1)
	}

	// 处理命令行参数
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-migrate":
			if err := runMigration(cfg); err != nil {
				logger.Error("Migration failed", logger.Err(err))
				os.Exit(1)
			}
			return
		case "-seed":
			logger.Info("Seed data import not implemented yet")
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

// runMigration 执行数据库迁移
func runMigration(cfg *config.Config) error {
	db, err := database.NewDatabase(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	logger.Info("Starting database migration...")
	if err := database.AutoMigrate(db); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	logger.Info("Database migration completed")
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
