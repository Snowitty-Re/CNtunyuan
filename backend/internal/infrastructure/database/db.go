package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/Snowitty-Re/CNtunyuan/pkg/metrics"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// CreateDatabaseIfNotExists 如果数据库不存在则创建
func CreateDatabaseIfNotExists(cfg *config.DatabaseConfig) error {
	switch cfg.Type {
	case config.DatabaseTypeMySQL:
		return createMySQLDatabaseIfNotExists(cfg)
	case config.DatabaseTypePostgres:
		return createPostgresDatabaseIfNotExists(cfg)
	default:
		return fmt.Errorf("unsupported database type: %s", cfg.Type)
	}
}

// createPostgresDatabaseIfNotExists 创建 PostgreSQL 数据库
func createPostgresDatabaseIfNotExists(cfg *config.DatabaseConfig) error {
	// 连接到 postgres 数据库（不指定具体数据库）
	// 强制使用 disable 避免 TLS 连接问题
	sslMode := cfg.SSLMode
	if sslMode == "" || sslMode == "require" || sslMode == "prefer" {
		sslMode = "disable"
	}
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s client_encoding=UTF8",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, sslMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: &gormLogger{},
	})
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	defer sqlDB.Close()

	// 检查数据库是否存在
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '%s')", cfg.Database)
	if err := db.Raw(query).Scan(&exists).Error; err != nil {
		return fmt.Errorf("failed to check database exists: %w", err)
	}

	if !exists {
		logger.Info("Database does not exist, creating...", logger.String("database", cfg.Database))
		createSQL := fmt.Sprintf("CREATE DATABASE \"%s\" WITH ENCODING = 'UTF8'", cfg.Database)
		if err := db.Exec(createSQL).Error; err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		logger.Info("Database created successfully", logger.String("database", cfg.Database))
	}

	return nil
}

// createMySQLDatabaseIfNotExists 创建 MySQL 数据库
func createMySQLDatabaseIfNotExists(cfg *config.DatabaseConfig) error {
	// 连接到 MySQL（不指定具体数据库）
	charset := cfg.Charset
	if charset == "" {
		charset = "utf8mb4"
	}

	// 构建不包含数据库名的 DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=%s&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, charset)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: &gormLogger{},
	})
	if err != nil {
		return fmt.Errorf("failed to connect to mysql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	defer sqlDB.Close()

	// 检查数据库是否存在
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = '%s')", cfg.Database)
	if err := db.Raw(query).Scan(&exists).Error; err != nil {
		return fmt.Errorf("failed to check database exists: %w", err)
	}

	if !exists {
		logger.Info("Database does not exist, creating...", logger.String("database", cfg.Database))
		createSQL := fmt.Sprintf("CREATE DATABASE `%s` CHARACTER SET %s COLLATE %s_unicode_ci",
			cfg.Database, charset, charset)
		if err := db.Exec(createSQL).Error; err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		logger.Info("Database created successfully", logger.String("database", cfg.Database))
	}

	return nil
}

// DB 数据库连接管理器
type DB struct {
	*gorm.DB
	config *config.DatabaseConfig
}

// NewDatabase 创建数据库连接
func NewDatabase(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	if !cfg.IsValid() {
		return nil, fmt.Errorf("invalid database configuration")
	}

	dsn := cfg.GetDSN()
	if dsn == "" {
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	var dialector gorm.Dialector
	switch cfg.Type {
	case config.DatabaseTypeMySQL:
		dialector = mysql.Open(dsn)
	case config.DatabaseTypePostgres:
		dialector = postgres.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	gormDB, err := gorm.Open(dialector, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "ty_",
			SingularTable: false,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger: &gormLogger{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 连接池配置优化
	// 设置最大空闲连接数（建议等于最大连接数）
	maxIdle := cfg.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = 10
	}
	sqlDB.SetMaxIdleConns(maxIdle)

	// 设置最大打开连接数
	maxOpen := cfg.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 100
	}
	sqlDB.SetMaxOpenConns(maxOpen)

	// 设置连接最大生命周期
	maxLifetime := cfg.ConnMaxLifetime
	if maxLifetime <= 0 {
		maxLifetime = 3600 // 默认1小时
	}
	sqlDB.SetConnMaxLifetime(time.Duration(maxLifetime) * time.Second)

	// 设置连接最大空闲时间（Go 1.15+）
	// 防止空闲连接长时间占用资源
	sqlDB.SetConnMaxIdleTime(time.Duration(maxLifetime/2) * time.Second)

	logger.Info("Database connected successfully",
		logger.String("type", string(cfg.Type)),
		logger.String("host", cfg.Host),
		logger.Int("port", cfg.Port),
		logger.String("database", cfg.Database),
		logger.Int("max_idle", maxIdle),
		logger.Int("max_open", maxOpen),
		logger.Int("max_lifetime", maxLifetime),
	)

	// 启动连接池监控
	go monitorConnectionPool(sqlDB)

	return gormDB, nil
}

// monitorConnectionPool 监控连接池状态
func monitorConnectionPool(sqlDB *sql.DB) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := sqlDB.Stats()
		metrics.SetDBConnections(float64(stats.OpenConnections))

		// 记录连接池状态日志
		logger.Debug("Database connection pool stats",
			logger.Int("open", stats.OpenConnections),
			logger.Int("in_use", stats.InUse),
			logger.Int("idle", stats.Idle),
			logger.Int64("wait_count", stats.WaitCount),
		)

		// 告警：如果等待连接数过多
		if stats.WaitCount > 100 {
			logger.Warn("High database wait count detected",
				logger.Int64("wait_count", stats.WaitCount),
				logger.String("suggestion", "Consider increasing MaxOpenConns"),
			)
		}
	}
}

// TestConnection 测试数据库连接
func TestConnection(cfg *config.DatabaseConfig) error {
	db, err := NewDatabase(cfg)
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return sqlDB.PingContext(ctx)
}

// AutoMigrate 自动迁移数据库
func AutoMigrate(db *gorm.DB) error {
	logger.Info("Starting database migration...")

	// 迁移所有实体
	err := db.AutoMigrate(
		&entity.User{},
		&entity.Organization{},
		&entity.MissingPerson{},
		&entity.Dialect{},
		&entity.Task{},
		&entity.File{},
	)

	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	logger.Info("Database migration completed successfully")
	return nil
}

// gormLogger GORM 日志适配器
type gormLogger struct{}

func (l *gormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return l
}

func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	logger.Info(fmt.Sprintf(msg, data...))
}

func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	logger.Warn(fmt.Sprintf(msg, data...))
}

func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	logger.Error(fmt.Sprintf(msg, data...))
}

func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	// 记录慢查询（超过 100ms）
	if elapsed > 100*time.Millisecond {
		logger.Warn("Slow query detected",
			logger.Duration("elapsed", elapsed),
			logger.String("sql", sql),
			logger.Int64("rows", rows),
		)
	}

	// 记录 Prometheus 指标
	metrics.RecordDBQuery("execute", "unknown", elapsed.Seconds())
}
