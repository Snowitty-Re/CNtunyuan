package model

import (
	"fmt"
	"log"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"golang.org/x/crypto/bcrypt"
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
		return nil, err
	}

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	DB = db
	return db, nil
}

// AutoMigrate 自动迁移数据库结构
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&UserProfile{},
		&Organization{},
		&MissingPerson{},
		&MissingPhoto{},
		&Dialect{},
		&Task{},
		&Workflow{},
		&OperationLog{},
	)
}

// InitRootOrganization 初始化根组织
func InitRootOrganization(db *gorm.DB) error {
	var count int64
	if err := db.Model(&Organization{}).Where("type = ?", OrgTypeRoot).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	rootOrg := Organization{
		Name:    "团圆寻亲志愿者系统",
		Code:    "ROOT",
		Type:    OrgTypeRoot,
		Level:   0,
		Status:  OrgStatusActive,
		Address: "中国",
	}

	if err := db.Create(&rootOrg).Error; err != nil {
		return err
	}

	log.Println("根组织创建成功")
	return nil
}

// InitSuperAdmin 初始化超级管理员
func InitSuperAdmin(db *gorm.DB) error {
	var superAdmin User
	err := db.Where("role = ?", RoleSuperAdmin).First(&superAdmin).Error

	// 生成密码哈希（默认密码: admin123）
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	if err == nil {
		// 超级管理员已存在，检查密码是否为空
		if superAdmin.Password == "" {
			// 重置密码
			if err := db.Model(&superAdmin).Update("password", string(passwordHash)).Error; err != nil {
				return fmt.Errorf("重置超级管理员密码失败: %w", err)
			}
			log.Printf("超级管理员密码已重置，手机号: %s, 默认密码: admin123", superAdmin.Phone)
		}
		return nil
	}

	if err != gorm.ErrRecordNotFound {
		return err
	}

	// 获取根组织ID
	var rootOrg Organization
	if err := db.Where("type = ?", OrgTypeRoot).First(&rootOrg).Error; err != nil {
		return fmt.Errorf("获取根组织失败: %w", err)
	}

	superAdmin = User{
		Nickname: "超级管理员",
		RealName: "系统管理员",
		Phone:    "13800138000",
		Email:    "admin@cntunyuan.com",
		Password: string(passwordHash),
		Role:     RoleSuperAdmin,
		Status:   UserStatusActive,
		OrgID:    &rootOrg.ID,
	}

	if err := db.Create(&superAdmin).Error; err != nil {
		return fmt.Errorf("创建超级管理员失败: %w", err)
	}

	log.Printf("超级管理员创建成功，ID: %s, 手机号: %s, 默认密码: admin123", superAdmin.ID, superAdmin.Phone)
	return nil
}

// ResetSuperAdminPassword 重置超级管理员密码
func ResetSuperAdminPassword(db *gorm.DB, newPassword string) error {
	var superAdmin User
	if err := db.Where("role = ?", RoleSuperAdmin).First(&superAdmin).Error; err != nil {
		return fmt.Errorf("超级管理员不存在: %w", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	if err := db.Model(&superAdmin).Update("password", string(passwordHash)).Error; err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	log.Printf("超级管理员密码已重置，手机号: %s", superAdmin.Phone)
	return nil
}
