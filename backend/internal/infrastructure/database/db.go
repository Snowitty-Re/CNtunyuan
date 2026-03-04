package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
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
	sslMode := cfg.SSLMode
	if sslMode == "" {
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

	// 连接池配置
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	logger.Info("Database connected successfully",
		logger.String("type", string(cfg.Type)),
		logger.String("host", cfg.Host),
		logger.Int("port", cfg.Port),
		logger.String("database", cfg.Database),
	)

	return gormDB, nil
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

	return sqlDB.Ping()
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
	logger.Error("GORM Error", logger.String("msg", fmt.Sprintf(msg, data...)))
}

func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	if err != nil {
		logger.Error("SQL Trace",
			logger.String("sql", sql),
			logger.Int64("rows", rows),
			logger.Duration("elapsed", elapsed),
			logger.Err(err),
		)
	} else if elapsed > time.Second {
		logger.Warn("Slow SQL",
			logger.String("sql", sql),
			logger.Int64("rows", rows),
			logger.Duration("elapsed", elapsed),
		)
	} else {
		logger.Debug("SQL Trace",
			logger.String("sql", sql),
			logger.Int64("rows", rows),
			logger.Duration("elapsed", elapsed),
		)
	}
}

// Ensure gormLogger implements gormlogger.Interface
var _ gormlogger.Interface = (*gormLogger)(nil)

// AutoMigrate 自动迁移数据库结构
func AutoMigrate(db *gorm.DB) error {
	logger.Info("Starting database auto-migration...")

	// 自动迁移所有实体模型
	err := db.AutoMigrate(
		// 用户相关
		&entity.User{},
		&entity.Permission{},
		&entity.UserPermission{},

		// 组织相关
		&entity.Organization{},
		&entity.OrgStats{},

		// 走失人员相关
		&entity.MissingPerson{},
		&entity.MissingPersonTrack{},

		// 方言相关
		&entity.Dialect{},
		&entity.DialectComment{},
		&entity.DialectLike{},
		&entity.DialectPlayLog{},

		// 任务相关
		&entity.Task{},
		&entity.TaskAttachment{},
		&entity.TaskLog{},
		&entity.TaskComment{},

		// 文件相关
		&entity.File{},
	)

	if err != nil {
		return fmt.Errorf("auto migration failed: %w", err)
	}

	logger.Info("Database auto-migration completed successfully")
	return nil
}
