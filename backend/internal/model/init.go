package model

import (
	"fmt"
	"log"
	"time"

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
		// 生产环境建议启用外键约束
		DisableForeignKeyConstraintWhenMigrating: false,
	})
	if err != nil {
		return nil, err
	}

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 连接池配置
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	DB = db
	return db, nil
}

// AutoMigrate 自动迁移数据库结构
func AutoMigrate(db *gorm.DB) error {
	// 先迁移基础表
	err := db.AutoMigrate(
		// 用户相关
		&User{},
		&UserProfile{},
		// 组织相关
		&Organization{},
		&OrgStats{},
		// 走失人员相关
		&MissingPerson{},
		&MissingPhoto{},
		&MissingPersonTrack{},
		// 方言相关
		&Dialect{},
		&DialectComment{},
		&DialectLike{},
		&DialectPlayLog{},
		// 任务相关
		&Task{},
		&TaskAttachment{},
		&TaskLog{},
		&TaskComment{},
		// 工作流相关
		&Workflow{},
		&WorkflowStep{},
		&WorkflowInstance{},
		&WorkflowHistory{},
		// 通用
		&Tag{},
		&Notification{},
		&OperationLog{},
		&Config{},
		&DashboardStats{},
	)
	if err != nil {
		return err
	}

	// 创建外键约束
	if err := createForeignKeys(db); err != nil {
		log.Printf("创建外键约束警告: %v", err)
		// 外键创建失败不阻塞迁移
	}

	// 创建性能优化索引
	if err := createPerformanceIndexes(db); err != nil {
		log.Printf("创建性能索引警告: %v", err)
	}

	return nil
}

// createForeignKeys 创建外键约束
func createForeignKeys(db *gorm.DB) error {
	foreignKeys := []struct {
		model       interface{}
		field       string
		references  string
		onDelete    string
	}{
		// 用户外键
		{&User{}, "org_id", "organizations(id)", "SET NULL"},
		{&UserProfile{}, "user_id", "users(id)", "CASCADE"},

		// 组织外键
		{&Organization{}, "parent_id", "organizations(id)", "CASCADE"},
		{&Organization{}, "leader_id", "users(id)", "SET NULL"},
		{&OrgStats{}, "org_id", "organizations(id)", "CASCADE"},

		// 走失人员外键
		{&MissingPerson{}, "reporter_id", "users(id)", "RESTRICT"},
		{&MissingPerson{}, "org_id", "organizations(id)", "RESTRICT"},
		{&MissingPhoto{}, "missing_person_id", "missing_persons(id)", "CASCADE"},
		{&MissingPersonTrack{}, "missing_person_id", "missing_persons(id)", "CASCADE"},
		{&MissingPersonTrack{}, "reporter_id", "users(id)", "RESTRICT"},

		// 方言外键
		{&Dialect{}, "collector_id", "users(id)", "RESTRICT"},
		{&Dialect{}, "org_id", "organizations(id)", "RESTRICT"},
		{&DialectComment{}, "dialect_id", "dialects(id)", "CASCADE"},
		{&DialectComment{}, "user_id", "users(id)", "RESTRICT"},
		{&DialectLike{}, "dialect_id", "dialects(id)", "CASCADE"},
		{&DialectLike{}, "user_id", "users(id)", "CASCADE"},
		{&DialectPlayLog{}, "dialect_id", "dialects(id)", "CASCADE"},
		{&DialectPlayLog{}, "user_id", "users(id)", "CASCADE"},

		// 任务外键
		{&Task{}, "missing_person_id", "missing_persons(id)", "SET NULL"},
		{&Task{}, "creator_id", "users(id)", "RESTRICT"},
		{&Task{}, "assignee_id", "users(id)", "SET NULL"},
		{&Task{}, "org_id", "organizations(id)", "RESTRICT"},
		{&Task{}, "workflow_id", "workflows(id)", "SET NULL"},
		{&TaskAttachment{}, "task_id", "tasks(id)", "CASCADE"},
		{&TaskLog{}, "task_id", "tasks(id)", "CASCADE"},
		{&TaskLog{}, "user_id", "users(id)", "RESTRICT"},
		{&TaskComment{}, "task_id", "tasks(id)", "CASCADE"},
		{&TaskComment{}, "user_id", "users(id)", "RESTRICT"},

		// 工作流外键
		{&Workflow{}, "creator_id", "users(id)", "RESTRICT"},
		{&WorkflowStep{}, "workflow_id", "workflows(id)", "CASCADE"},
		{&WorkflowInstance{}, "workflow_id", "workflows(id)", "CASCADE"},
		{&WorkflowInstance{}, "current_step_id", "workflow_steps(id)", "SET NULL"},
		{&WorkflowInstance{}, "starter_id", "users(id)", "RESTRICT"},
		{&WorkflowHistory{}, "instance_id", "workflow_instances(id)", "CASCADE"},
		{&WorkflowHistory{}, "step_id", "workflow_steps(id)", "RESTRICT"},
		{&WorkflowHistory{}, "operator_id", "users(id)", "SET NULL"},

		// 通知外键
		{&Notification{}, "sender_id", "users(id)", "SET NULL"},
		{&Notification{}, "receiver_id", "users(id)", "CASCADE"},
		{&OperationLog{}, "user_id", "users(id)", "SET NULL"},
	}

	// GORM 会自动创建外键，这里只记录需要手动创建的
	// 如需手动创建外键，可在此遍历 foreignKeys 并执行 SQL
	_ = foreignKeys

	return nil
}

// createPerformanceIndexes 创建性能优化索引
func createPerformanceIndexes(db *gorm.DB) error {
	indexes := []struct {
		table   string
		columns string
		name    string
	}{
		// 用户索引
		{"ty_users", "status,created_at", "idx_users_status_created"},
		{"ty_users", "role,org_id", "idx_users_role_org"},

		// 走失人员索引
		{"ty_missing_persons", "status,missing_time", "idx_mp_status_time"},
		{"ty_missing_persons", "case_type,status", "idx_mp_type_status"},
		{"ty_missing_persons", "org_id,status", "idx_mp_org_status"},
		{"ty_missing_persons", "reporter_id,created_at", "idx_mp_reporter"},
		{"ty_missing_persons", "missing_location", "idx_mp_location"},
		// 地理位置索引（使用 GiST）
		{"ty_missing_persons", "missing_latitude,missing_longitude", "idx_mp_geo"},

		// 方言索引
		{"ty_dialects", "province,city,district", "idx_dialect_region"},
		{"ty_dialects", "collector_id,created_at", "idx_dialect_collector"},
		{"ty_dialects", "status,play_count", "idx_dialect_status_play"},
		{"ty_dialects", "latitude,longitude", "idx_dialect_geo"},

		// 任务索引
		{"ty_tasks", "status,assignee_id", "idx_task_status_assignee"},
		{"ty_tasks", "org_id,status", "idx_task_org_status"},
		{"ty_tasks", "creator_id,created_at", "idx_task_creator"},
		{"ty_tasks", "missing_person_id,status", "idx_task_mp_status"},
		{"ty_tasks", "deadline,status", "idx_task_deadline"},
		{"ty_tasks", "priority,status", "idx_task_priority"},

		// 工作流实例索引
		{"ty_workflow_instances", "starter_id,status", "idx_wfi_starter_status"},
		{"ty_workflow_instances", "business_id,business_type", "idx_wfi_business"},

		// 操作日志索引
		{"ty_operation_logs", "user_id,created_at", "idx_oplog_user_time"},
		{"ty_operation_logs", "created_at", "idx_oplog_time"},
	}

	for _, idx := range indexes {
		sql := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s(%s)", idx.name, idx.table, idx.columns)
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("创建索引 %s 失败: %v", idx.name, err)
		}
	}

	// 创建全文搜索索引（PostgreSQL）
	fullTextIndexes := []struct {
		table   string
		columns string
		name    string
	}{
		{"ty_missing_persons", "name,appearance,clothing,special_features", "idx_mp_search"},
		{"ty_dialects", "title,description,address", "idx_dialect_search"},
		{"ty_tasks", "title,description", "idx_task_search"},
	}

	for _, idx := range fullTextIndexes {
		sql := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s USING gin(to_tsvector('chinese', COALESCE(%s, '')))`,
			idx.name, idx.table, idx.columns)
		if err := db.Exec(sql).Error; err != nil {
			// 全文搜索索引创建失败不阻塞
			log.Printf("创建全文索引 %s 失败: %v", idx.name, err)
		}
	}

	return nil
}

// InitRootOrganization 初始化根组织
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
		Name:        "团圆志愿者总部",
		Code:        "ROOT",
		Type:        OrgTypeRoot,
		Level:       1,
		Status:      OrgStatusActive,
		Province:    "全国",
		Description: "团圆寻亲志愿者系统总部",
	}

	if err := db.Create(&rootOrg).Error; err != nil {
		return fmt.Errorf("创建根组织失败: %w", err)
	}

	log.Printf("根组织创建成功，ID: %s", rootOrg.ID)
	return nil
}

// CreateSuperAdmin 创建超级管理员
func CreateSuperAdmin(db *gorm.DB, phone, email, password string) (*User, error) {
	var rootOrg Organization
	if err := db.Where("type = ?", OrgTypeRoot).First(&rootOrg).Error; err != nil {
		return nil, fmt.Errorf("根组织不存在，请先执行 -init: %w", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	var existing User
	result := db.Where("role = ?", RoleSuperAdmin).First(&existing)

	if result.Error == nil {
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
