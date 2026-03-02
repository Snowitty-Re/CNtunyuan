package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// DB 数据库连接管理器
type DB struct {
	*gorm.DB
	config *config.DatabaseConfig
}

// NewDatabase 创建数据库连接
func NewDatabase(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
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
		logger.String("host", cfg.Host),
		logger.Int("port", cfg.Port),
		logger.String("database", cfg.Database),
	)

	return gormDB, nil
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
