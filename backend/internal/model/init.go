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
		Logger: logger.Default.LogMode(logger.Info),
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
// 如果根组织已存在，则跳过
func InitRootOrganization(db *gorm.DB) error {
	var count int64
	if err := db.Model(&Organization{}).Where("type = ?", OrgTypeRoot).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		log.Println("根组织已存在，跳过初始化")
		return nil
	}

	rootOrg := Organization{
		Name:    "团圆志愿者总部",
		Code:    "ROOT",
		Type:    OrgTypeRoot,
		Level:   1,
		Status:  OrgStatusActive,
		Province: "全国",
		Description: "团圆寻亲志愿者系统总部",
	}

	if err := db.Create(&rootOrg).Error; err != nil {
		return fmt.Errorf("创建根组织失败: %w", err)
	}

	log.Printf("根组织创建成功，ID: %s", rootOrg.ID)
	return nil
}

// CreateSuperAdmin 创建超级管理员
// 如果已存在超级管理员，则更新信息
func CreateSuperAdmin(db *gorm.DB, phone, email, password string) (*User, error) {
	// 获取根组织
	var rootOrg Organization
	if err := db.Where("type = ?", OrgTypeRoot).First(&rootOrg).Error; err != nil {
		return nil, fmt.Errorf("根组织不存在，请先执行 -init: %w", err)
	}

	// 生成密码哈希
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 检查是否已存在超级管理员
	var existing User
	result := db.Where("role = ?", RoleSuperAdmin).First(&existing)

	if result.Error == nil {
		// 更新现有超级管理员
		updates := map[string]interface{}{
			"phone":    phone,
			"email":    email,
			"password": string(passwordHash),
		}
		if err := db.Model(&existing).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("更新超级管理员失败: %w", err)
		}
		log.Printf("超级管理员信息已更新，手机号: %s", phone)
		return &existing, nil
	}

	// 创建新的超级管理员
	admin := User{
		Nickname: "超级管理员",
		RealName: "系统管理员",
		Phone:    phone,
		Email:    email,
		Password: string(passwordHash),
		Role:     RoleSuperAdmin,
		Status:   UserStatusActive,
		OrgID:    &rootOrg.ID,
	}

	if err := db.Create(&admin).Error; err != nil {
		return nil, fmt.Errorf("创建超级管理员失败: %w", err)
	}

	log.Printf("超级管理员创建成功，ID: %s, 手机号: %s", admin.ID, phone)
	return &admin, nil
}

// ResetAdminPassword 重置管理员密码
func ResetAdminPassword(db *gorm.DB, phone, newPassword string) error {
	var user User
	if err := db.Where("phone = ?", phone).First(&user).Error; err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	if err := db.Model(&user).Update("password", string(passwordHash)).Error; err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	log.Printf("密码重置成功，用户: %s", phone)
	return nil
}
