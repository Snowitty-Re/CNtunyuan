package model

import (
	"fmt"
	"log"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// DB 全局数据库连接
var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := cfg.GetDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Info),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "ty_", // 表前缀: tunyuan
			SingularTable: false,
		},
		DisableForeignKeyConstraintWhenMigrating: true, // 禁用外键约束自动创建
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库实例失败: %w", err)
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	DB = db
	return db, nil
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate(db *gorm.DB) error {
	models := []interface{}{
		&User{},
		&UserProfile{},
		&Organization{},
		&OrgStats{},
		&MissingPerson{},
		&MissingPhoto{},
		&MissingPersonTrack{},
		&Dialect{},
		&DialectComment{},
		&DialectLike{},
		&DialectPlayLog{},
		&Task{},
		&TaskAttachment{},
		&TaskLog{},
		&TaskComment{},
		&Workflow{},
		&WorkflowStep{},
		&WorkflowInstance{},
		&WorkflowHistory{},
		&Tag{},
		&Notification{},
		&OperationLog{},
		&Config{},
		&DashboardStats{},
	}

	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			log.Printf("迁移表失败: %v", err)
			return err
		}
	}

	log.Println("数据库迁移完成")
	return nil
}

// InitRootOrg 初始化根组织
func InitRootOrg(db *gorm.DB) error {
	var count int64
	if err := db.Model(&Organization{}).Where("type = ?", OrgTypeRoot).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	rootOrg := Organization{
		Name:     "团圆志愿者总部",
		Code:     "ROOT",
		Type:     OrgTypeRoot,
		Level:    1,
		Province: "全国",
		Status:   "active",
	}

	if err := db.Create(&rootOrg).Error; err != nil {
		return fmt.Errorf("创建根组织失败: %w", err)
	}

	log.Println("根组织创建成功")
	return nil
}
