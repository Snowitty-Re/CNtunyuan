// @title 助力团圆志愿者系统 API
// @version 1.0.0
// @description 助力团圆志愿者系统后端 API 文档
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
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/Snowitty-Re/CNtunyuan/docs"
	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/di"
	"github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/database"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/gin-gonic/gin"
	_ "gorm.io/gorm"
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

	// 根据配置设置 Gin 运行模式，确保 server.mode 生效
	setGinMode(cfg.Server.Mode)

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
		case "-help", "--help", "-h":
			fmt.Println("Usage: go run cmd/app/main.go [option]")
			fmt.Println()
			fmt.Println("Options:")
			fmt.Println("  -check-db   Check database connectivity and required tables")
			fmt.Println("  -help       Show this help message")
			return
		default:
			logger.Error("Unknown option", logger.String("option", os.Args[1]))
			fmt.Println("Use -help to see available options.")
			os.Exit(1)
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

func setGinMode(mode string) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "release":
		gin.SetMode(gin.ReleaseMode)
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		// 配置校验已约束为 debug/release，这里保底用 release
		gin.SetMode(gin.ReleaseMode)
	}
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
		"ty_missing_person_tracks",
		"ty_tasks",
		"ty_task_follow_ups",
		"ty_task_follow_up_comments",
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

// startServer 启动服务器
func startServer(cfg *config.Config, container *di.Container) {
	engine := container.Router.GetEngine()

	// 创建 HTTP 服务器
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	logger.Info("Starting server", logger.String("port", port))

	// 使用 http.Server 以支持优雅关闭
	srv := &http.Server{
		Addr:           ":" + port,
		Handler:        engine,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", logger.Err(err))
			os.Exit(1)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 优雅关闭：等待正在处理的请求完成
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", logger.Err(err))
	}

	// 关闭缓存连接
	if container.Cache != nil {
		if err := container.Cache.Close(); err != nil {
			logger.Error("Failed to close cache", logger.Err(err))
		}
	}

	logger.Info("Server stopped")
}
