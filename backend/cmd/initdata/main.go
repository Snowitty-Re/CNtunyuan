package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/model"
	"golang.org/x/crypto/bcrypt"
)

const initSQLTemplate = `-- 团圆寻亲志愿者系统 - 数据库初始化脚本
-- 生成时间: %s
-- 超级管理员: %s
-- 默认密码: %s

BEGIN;

-- ============================================
-- 1. 根组织 (团圆志愿者总部)
-- ============================================
INSERT INTO ty_organizations (
    id, name, code, type, level, parent_id, leader_id,
    province, city, district, street, address,
    contact, phone, email, description, sort, status,
    volunteer_count, case_count, created_at, updated_at
) VALUES (
    '00000000-0000-0000-0000-000000000001',
    '团圆志愿者总部',
    'ROOT',
    'root',
    1,
    NULL,
    NULL,
    '全国',
    '',
    '',
    '',
    '',
    '',
    '',
    '',
    '团圆寻亲志愿者系统总部',
    0,
    'active',
    0,
    0,
    NOW(),
    NOW()
) ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    updated_at = NOW();

-- ============================================
-- 2. 超级管理员
-- ============================================
-- 先删除已存在的超级管理员（如果存在）
DELETE FROM ty_users WHERE phone = '%s' OR role = 'super_admin';

-- 插入超级管理员
INSERT INTO ty_users (
    id, union_id, open_id, nickname, avatar, phone, email,
    real_name, id_card, password, role, status, org_id,
    last_login, login_ip, created_at, updated_at
) VALUES (
    '00000000-0000-0000-0000-000000000002',
    '',
    '',
    '超级管理员',
    '',
    '%s',
    '%s',
    '系统管理员',
    '',
    '%s',
    'super_admin',
    'active',
    '00000000-0000-0000-0000-000000000001',
    NULL,
    '',
    NOW(),
    NOW()
);

-- ============================================
-- 3. 示例省级组织 (可选)
-- ============================================
INSERT INTO ty_organizations (
    id, name, code, type, level, parent_id, leader_id,
    province, city, district, street, address,
    contact, phone, email, description, sort, status,
    volunteer_count, case_count, created_at, updated_at
) VALUES 
('10000000-0000-0000-0000-000000000001', '北京志愿者协会', 'BJ-001', 'province', 2, '00000000-0000-0000-0000-000000000001', NULL, '北京市', '', '', '', '', '', '', '', '北京市团圆志愿者协会', 1, 'active', 0, 0, NOW(), NOW()),
('10000000-0000-0000-0000-000000000002', '上海志愿者协会', 'SH-001', 'province', 2, '00000000-0000-0000-0000-000000000001', NULL, '上海市', '', '', '', '', '', '', '', '上海市团圆志愿者协会', 2, 'active', 0, 0, NOW(), NOW()),
('10000000-0000-0000-0000-000000000003', '广东志愿者协会', 'GD-001', 'province', 2, '00000000-0000-0000-0000-000000000001', NULL, '广东省', '', '', '', '', '', '', '', '广东省团圆志愿者协会', 3, 'active', 0, 0, NOW(), NOW())
ON CONFLICT (code) DO NOTHING;

COMMIT;

-- 验证数据
SELECT '初始化完成' as status;
SELECT id, name, code, type, status FROM ty_organizations WHERE code = 'ROOT';
SELECT id, nickname, phone, email, role, status FROM ty_users WHERE role = 'super_admin';
`

func main() {
	var (
		// 数据库配置
		configPath = flag.String("config", "config/config.yaml", "配置文件路径")

		// 超级管理员配置
		adminPhone    = flag.String("phone", "13800138000", "超级管理员手机号")
		adminEmail    = flag.String("email", "admin@cntunyuan.com", "超级管理员邮箱")
		adminPassword = flag.String("password", "admin123", "超级管理员初始密码")

		// 操作模式
		generateOnly = flag.Bool("gen", false, "仅生成SQL文件，不执行")
		outputFile   = flag.String("o", "sql/init_generated.sql", "生成SQL文件的输出路径")
		executeSQL   = flag.Bool("exec", false, "直接执行SQL初始化")
	)
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		fmt.Println("将使用默认数据库配置")
		cfg = &config.Config{
			Database: config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "postgres",
				Database: "cntuanyuan",
				SSLMode:  "disable",
			},
		}
	}

	// 生成密码哈希
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(*adminPassword), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("密码加密失败: %v\n", err)
		os.Exit(1)
	}

	// 生成SQL内容
	now := time.Now().Format("2006-01-02 15:04:05")
	sqlContent := fmt.Sprintf(
		initSQLTemplate,
		now,
		*adminPhone,
		*adminPassword,
		*adminPhone,
		*adminPhone,
		*adminEmail,
		string(passwordHash),
	)

	// 如果指定了生成SQL文件
	if *generateOnly || !*executeSQL {
		// 确保目录存在
		if err := os.MkdirAll("sql", 0755); err != nil {
			fmt.Printf("创建sql目录失败: %v\n", err)
		}
		
		if err := os.WriteFile(*outputFile, []byte(sqlContent), 0644); err != nil {
			fmt.Printf("写入SQL文件失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("SQL文件已生成: %s\n", *outputFile)
		fmt.Printf("\n超级管理员信息:\n")
		fmt.Printf("  手机号: %s\n", *adminPhone)
		fmt.Printf("  邮箱: %s\n", *adminEmail)
		fmt.Printf("  密码: %s\n", *adminPassword)

		if *generateOnly {
			fmt.Printf("\n请手动执行SQL:\n")
			fmt.Printf("  psql -h %s -p %d -U %s -d %s -f %s\n",
				cfg.Database.Host,
				cfg.Database.Port,
				cfg.Database.User,
				cfg.Database.Database,
				*outputFile,
			)
			return
		}
	}

	// 直接执行SQL
	if *executeSQL {
		fmt.Println("正在执行SQL初始化...")

		// 方法1: 通过psql命令执行
		psqlPath, err := exec.LookPath("psql")
		if err == nil {
			// 创建临时SQL文件
			tmpFile := *outputFile
			if !*generateOnly {
				tmpFile = "sql/init_temp.sql"
				if err := os.WriteFile(tmpFile, []byte(sqlContent), 0644); err != nil {
					fmt.Printf("创建临时SQL文件失败: %v\n", err)
					// 降级到使用GORM
					initWithGORM(cfg, *adminPhone, *adminEmail, string(passwordHash))
					return
				}
				defer os.Remove(tmpFile)
			}

			cmd := exec.Command(
				psqlPath,
				"-h", cfg.Database.Host,
				"-p", fmt.Sprintf("%d", cfg.Database.Port),
				"-U", cfg.Database.User,
				"-d", cfg.Database.Database,
				"-f", tmpFile,
			)
			cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", cfg.Database.Password))
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				fmt.Printf("psql执行失败: %v\n", err)
				fmt.Println("尝试使用GORM初始化...")
				initWithGORM(cfg, *adminPhone, *adminEmail, string(passwordHash))
			} else {
				fmt.Println("\n数据库初始化成功!")
				printAdminInfo(*adminPhone, *adminEmail, *adminPassword)
			}
		} else {
			// 方法2: 使用GORM执行
			initWithGORM(cfg, *adminPhone, *adminEmail, string(passwordHash))
		}
	}
}

// initWithGORM 使用GORM进行初始化
func initWithGORM(cfg *config.Config, phone, email, passwordHash string) {
	db, err := model.InitDB(&cfg.Database)
	if err != nil {
		fmt.Printf("连接数据库失败: %v\n", err)
		os.Exit(1)
	}

	// 执行迁移
	if err := model.AutoMigrate(db); err != nil {
		fmt.Printf("数据库迁移失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化根组织
	if err := model.InitRootOrganization(db); err != nil {
		fmt.Printf("初始化根组织失败: %v\n", err)
		os.Exit(1)
	}

	// 删除已存在的超级管理员
	db.Where("role = ? OR phone = ?", model.RoleSuperAdmin, phone).Delete(&model.User{})

	// 获取根组织
	var rootOrg model.Organization
	if err := db.Where("type = ?", model.OrgTypeRoot).First(&rootOrg).Error; err != nil {
		fmt.Printf("获取根组织失败: %v\n", err)
		os.Exit(1)
	}

	// 创建超级管理员
	admin := model.User{
		Nickname: "超级管理员",
		RealName: "系统管理员",
		Phone:    phone,
		Email:    email,
		Password: passwordHash,
		Role:     model.RoleSuperAdmin,
		Status:   model.UserStatusActive,
		OrgID:    &rootOrg.ID,
	}
	if err := db.Create(&admin).Error; err != nil {
		fmt.Printf("创建超级管理员失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n数据库初始化成功!")
	printAdminInfo(phone, email, "[设置的密码]")
}

func printAdminInfo(phone, email, password string) {
	fmt.Printf("\n================================\n")
	fmt.Printf("超级管理员信息:\n")
	fmt.Printf("  手机号: %s\n", phone)
	fmt.Printf("  邮箱: %s\n", email)
	if password != "" {
		fmt.Printf("  密码: %s\n", password)
	}
	fmt.Printf("================================\n")
}
