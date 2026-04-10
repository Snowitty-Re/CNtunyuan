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
	"github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/handler"
	httpRouter "github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/router"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type startupOptions struct {
	ConfigPath    string
	UseConfigFile bool
	CheckDB       bool
	ShowHelp      bool
}

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

	opts, err := parseStartupOptions(os.Args[1:])
	if err != nil {
		logger.Error("Invalid startup options", logger.Err(err))
		printUsage()
		os.Exit(1)
	}
	if opts.ShowHelp {
		printUsage()
		return
	}

	cfg, mode, err := loadStartupConfiguration(opts)
	if err != nil {
		logger.Error("Failed to load config", logger.Err(err))
		os.Exit(1)
	}

	// 根据配置设置 Gin 运行模式，确保 server.mode 生效
	setGinMode(cfg.Server.Mode)
	config.SetStartupMetadata(mode, displayConfigPath(opts.ConfigPath))
	logger.Info("Startup configuration resolved",
		logger.String("mode", mode),
		logger.String("config_path", displayConfigPath(opts.ConfigPath)),
	)

	if opts.CheckDB {
		if err := checkDatabase(cfg); err != nil {
			logger.Error("Database check failed", logger.Err(err))
			os.Exit(1)
		}
		return
	}

	// 优先进入完整模式；数据库未配置或不可用时退化到初始化引导模式
	if cfg.Database.IsValid() {
		container, containerErr := di.NewContainer(cfg)
		if containerErr == nil {
			startServer(cfg, container.Router.GetEngine(), func(context.Context) {
				if container.Cache != nil {
					if err := container.Cache.Close(); err != nil {
						logger.Error("Failed to close cache", logger.Err(err))
					}
				}
			})
			return
		}
		if opts.UseConfigFile {
			logger.Error("Full startup failed in file-config mode", logger.Err(containerErr))
			os.Exit(1)
		}
		logger.Warn("Full startup unavailable, switching to bootstrap mode", logger.Err(containerErr))
	} else {
		if opts.UseConfigFile {
			logger.Error("Database config incomplete in file-config mode")
			os.Exit(1)
		}
		logger.Warn("Database config incomplete, switching to bootstrap mode")
	}

	bootstrapHandler := handler.NewBootstrapHandler(nil, nil, nil)
	engine := httpRouter.NewBootstrapEngine(bootstrapHandler)
	startServer(cfg, engine, nil)
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
		"ty_dialect_card_groups",
		"ty_dialect_cards",
		"ty_files",
		"ty_audit_logs",
		"ty_system_settings",
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
		logger.Info("Database table check passed: all required tables exist")
	} else {
		return fmt.Errorf("some tables are missing, please run SQL migration files first")
	}

	requiredDialectColumns := []string{
		"batch_id",
		"card_group_id",
		"card_id",
	}

	logger.Info("Checking dialect schema columns...")
	for _, col := range requiredDialectColumns {
		exists, colErr := columnExists(db, "ty_dialects", col)
		if colErr != nil {
			logger.Error("Failed to check column", logger.String("table", "ty_dialects"), logger.String("column", col), logger.Err(colErr))
			allExist = false
			continue
		}
		if exists {
			logger.Info("✓ Column exists", logger.String("table", "ty_dialects"), logger.String("column", col))
		} else {
			logger.Error("✗ Column missing", logger.String("table", "ty_dialects"), logger.String("column", col))
			allExist = false
		}
	}

	if !allExist {
		return fmt.Errorf("database schema is not aligned, please run latest migration files first")
	}

	logger.Info("Database check passed: tables and key schema columns are aligned")

	return nil
}

func columnExists(db *gorm.DB, tableName, columnName string) (bool, error) {
	var count int64
	err := db.Raw(
		"SELECT COUNT(*) FROM information_schema.columns WHERE table_name = ? AND column_name = ?",
		tableName,
		columnName,
	).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// startServer 启动服务器
func startServer(cfg *config.Config, engine http.Handler, onShutdown func(context.Context)) {
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
	if onShutdown != nil {
		onShutdown(ctx)
	}

	logger.Info("Server stopped")
}

func parseStartupOptions(args []string) (startupOptions, error) {
	opts := startupOptions{}
	for i := 0; i < len(args); i++ {
		arg := strings.TrimSpace(args[i])
		switch arg {
		case "-help", "--help", "-h":
			opts.ShowHelp = true
		case "-check-db":
			opts.CheckDB = true
		case "--config":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--config requires a file path")
			}
			i++
			opts.ConfigPath = strings.TrimSpace(args[i])
			opts.UseConfigFile = true
		default:
			return opts, fmt.Errorf("unknown option: %s", arg)
		}
	}
	return opts, nil
}

func loadStartupConfiguration(opts startupOptions) (*config.Config, string, error) {
	if opts.UseConfigFile {
		cfg, err := config.LoadConfigFileOnly(opts.ConfigPath)
		if err != nil {
			return nil, "", err
		}
		return cfg, "file-config", nil
	}

	cfg, err := config.LoadStartupConfig(opts.ConfigPath)
	if err != nil {
		return nil, "", err
	}
	return cfg, "bootstrap-managed", nil
}

func printUsage() {
	fmt.Println("Usage: go run cmd/app/main.go [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --config <path>  Force file-config mode and read only the specified config.yaml")
	fmt.Println("  -check-db        Check database connectivity and required tables")
	fmt.Println("  -help            Show this help message")
	fmt.Println()
	fmt.Println("Startup modes:")
	fmt.Println("  default          bootstrap-managed mode (managed startup config + init wizard)")
	fmt.Println("  --config         file-config mode (local config only, no managed startup config)")
}

func displayConfigPath(path string) string {
	if strings.TrimSpace(path) == "" {
		return "default"
	}
	return path
}
