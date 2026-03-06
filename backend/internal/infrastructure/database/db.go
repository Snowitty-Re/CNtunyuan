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

// AutoMigrate 自动迁移数据库（仅用于开发环境）
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

// TableExists 检查表是否存在
func TableExists(db *gorm.DB, tableName string) (bool, error) {
	var count int64
	err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ?", tableName).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Seed 插入种子数据（仅用于开发环境）
func Seed(db *gorm.DB) error {
	logger.Info("Starting database seeding...")

	// 检查是否已有数据
	var count int64
	if err := db.Model(&entity.Organization{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check organizations: %w", err)
	}
	if count > 0 {
		logger.Info("Database already has data, skipping seed")
		return nil
	}

	// 创建根组织
	rootOrg, err := entity.NewRootOrganization("团圆寻亲志愿者协会", "ROOT")
	if err != nil {
		return fmt.Errorf("failed to create root organization: %w", err)
	}
	if err := db.Create(rootOrg).Error; err != nil {
		return fmt.Errorf("failed to save root organization: %w", err)
	}
	logger.Info("Created root organization")

	// 创建超级管理员
	admin, err := entity.NewSuperAdmin("超级管理员", "13800138000", "admin123")
	if err != nil {
		return fmt.Errorf("failed to create super admin: %w", err)
	}
	if err := db.Create(admin).Error; err != nil {
		return fmt.Errorf("failed to save super admin: %w", err)
	}
	logger.Info("Created super admin")

	logger.Info("Database seeding completed successfully")
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
