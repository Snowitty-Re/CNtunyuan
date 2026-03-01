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
		// 禁用外键约束自动创建，我们手动创建以处理循环依赖
		DisableForeignKeyConstraintWhenMigrating: true,
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
	// 由于存在循环外键依赖，需要使用原始DB创建一个新的连接来禁用外键约束
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// 获取数据库配置
	var dbName string
	sqlDB.QueryRow("SELECT CURRENT_DATABASE()").Scan(&dbName)

	// 临时禁用外键约束检查（PostgreSQL语法）
	// 注意：这只是为了让AutoMigrate顺利进行，实际外键会在之后手动创建
	models := []interface{}{
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
	}

	log.Println("创建数据库表...")
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("迁移表 %T 失败: %v", model, err)
		}
	}
	log.Println("数据库表创建完成")

	// 第二阶段：创建外键约束
	log.Println("创建外键约束...")
	if err := createForeignKeys(db); err != nil {
		log.Printf("创建外键约束警告: %v", err)
	}

	// 第三阶段：创建性能优化索引
	log.Println("创建性能索引...")
	if err := createPerformanceIndexes(db); err != nil {
		log.Printf("创建性能索引警告: %v", err)
	}

	return nil
}

// createForeignKeys 创建外键约束
func createForeignKeys(db *gorm.DB) error {
	// 手动创建外键约束（处理循环依赖和复杂关系）
	
	foreignKeys := []struct {
		table      string
		column     string
		refTable   string
		refColumn  string
		onDelete   string
		onUpdate   string
	}{
		// 用户外键
		{"ty_users", "org_id", "ty_organizations", "id", "SET NULL", "CASCADE"},
		{"ty_user_profiles", "user_id", "ty_users", "id", "CASCADE", "CASCADE"},
		
		// 组织外键（注意：leader_id 可能为 NULL）
		{"ty_organizations", "parent_id", "ty_organizations", "id", "CASCADE", "CASCADE"},
		{"ty_organizations", "leader_id", "ty_users", "id", "SET NULL", "CASCADE"},
		{"ty_org_stats", "org_id", "ty_organizations", "id", "CASCADE", "CASCADE"},
		
		// 走失人员外键
		{"ty_missing_persons", "reporter_id", "ty_users", "id", "RESTRICT", "CASCADE"},
		{"ty_missing_persons", "org_id", "ty_organizations", "id", "RESTRICT", "CASCADE"},
		{"ty_missing_photos", "missing_person_id", "ty_missing_persons", "id", "CASCADE", "CASCADE"},
		{"ty_missing_person_tracks", "missing_person_id", "ty_missing_persons", "id", "CASCADE", "CASCADE"},
		{"ty_missing_person_tracks", "reporter_id", "ty_users", "id", "RESTRICT", "CASCADE"},
		
		// 方言外键
		{"ty_dialects", "collector_id", "ty_users", "id", "RESTRICT", "CASCADE"},
		{"ty_dialects", "org_id", "ty_organizations", "id", "RESTRICT", "CASCADE"},
		{"ty_dialect_comments", "dialect_id", "ty_dialects", "id", "CASCADE", "CASCADE"},
		{"ty_dialect_comments", "user_id", "ty_users", "id", "RESTRICT", "CASCADE"},
		{"ty_dialect_likes", "dialect_id", "ty_dialects", "id", "CASCADE", "CASCADE"},
		{"ty_dialect_likes", "user_id", "ty_users", "id", "CASCADE", "CASCADE"},
		{"ty_dialect_play_logs", "dialect_id", "ty_dialects", "id", "CASCADE", "CASCADE"},
		{"ty_dialect_play_logs", "user_id", "ty_users", "id", "CASCADE", "CASCADE"},
		
		// 任务外键
		{"ty_tasks", "missing_person_id", "ty_missing_persons", "id", "SET NULL", "CASCADE"},
		{"ty_tasks", "creator_id", "ty_users", "id", "RESTRICT", "CASCADE"},
		{"ty_tasks", "assignee_id", "ty_users", "id", "SET NULL", "CASCADE"},
		{"ty_tasks", "org_id", "ty_organizations", "id", "RESTRICT", "CASCADE"},
		{"ty_tasks", "workflow_id", "ty_workflows", "id", "SET NULL", "CASCADE"},
		{"ty_task_attachments", "task_id", "ty_tasks", "id", "CASCADE", "CASCADE"},
		{"ty_task_logs", "task_id", "ty_tasks", "id", "CASCADE", "CASCADE"},
		{"ty_task_logs", "user_id", "ty_users", "id", "RESTRICT", "CASCADE"},
		{"ty_task_comments", "task_id", "ty_tasks", "id", "CASCADE", "CASCADE"},
		{"ty_task_comments", "user_id", "ty_users", "id", "RESTRICT", "CASCADE"},
		
		// 工作流外键
		{"ty_workflows", "creator_id", "ty_users", "id", "RESTRICT", "CASCADE"},
		{"ty_workflow_steps", "workflow_id", "ty_workflows", "id", "CASCADE", "CASCADE"},
		{"ty_workflow_instances", "workflow_id", "ty_workflows", "id", "CASCADE", "CASCADE"},
		{"ty_workflow_instances", "current_step_id", "ty_workflow_steps", "id", "SET NULL", "CASCADE"},
		{"ty_workflow_instances", "starter_id", "ty_users", "id", "RESTRICT", "CASCADE"},
		{"ty_workflow_histories", "instance_id", "ty_workflow_instances", "id", "CASCADE", "CASCADE"},
		{"ty_workflow_histories", "step_id", "ty_workflow_steps", "id", "RESTRICT", "CASCADE"},
		{"ty_workflow_histories", "operator_id", "ty_users", "id", "SET NULL", "CASCADE"},
		
		// 通知外键
		{"ty_notifications", "sender_id", "ty_users", "id", "SET NULL", "CASCADE"},
		{"ty_notifications", "receiver_id", "ty_users", "id", "CASCADE", "CASCADE"},
		
		// 操作日志外键（可选，通常不严格约束）
		{"ty_operation_logs", "user_id", "ty_users", "id", "SET NULL", "CASCADE"},
	}

	for _, fk := range foreignKeys {
		// 检查外键是否已存在
		var count int64
		checkSQL := `
			SELECT COUNT(*) FROM information_schema.table_constraints 
			WHERE constraint_name = ? AND table_name = ?
		`
		constraintName := fmt.Sprintf("fk_%s_%s", fk.table, fk.column)
		db.Raw(checkSQL, constraintName, fk.table).Scan(&count)
		
		if count > 0 {
			continue // 外键已存在
		}

		// 创建外键
		sql := fmt.Sprintf(
			"ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(%s) ON DELETE %s ON UPDATE %s",
			fk.table, constraintName, fk.column, fk.refTable, fk.refColumn, fk.onDelete, fk.onUpdate,
		)
		
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("创建外键 %s 失败: %v", constraintName, err)
			// 继续创建其他外键
		}
	}

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
