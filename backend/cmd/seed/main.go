package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 种子数据结构
type SeedData struct {
	Organizations []model.Organization
	Users         []model.User
	MissingPersons []model.MissingPerson
	Dialects      []model.Dialect
	Tasks         []model.Task
}

func main() {
	var (
		configPath = flag.String("config", "config/config.yaml", "配置文件路径")
		all        = flag.Bool("all", false, "导入所有种子数据")
		orgs       = flag.Bool("orgs", false, "导入示例组织")
		users      = flag.Bool("users", false, "导入示例用户")
		cases      = flag.Bool("cases", false, "导入示例走失人员")
		dialects   = flag.Bool("dialects", false, "导入示例方言")
		tasks      = flag.Bool("tasks", false, "导入示例任务")
		clean      = flag.Bool("clean", false, "清空现有数据（危险！）")
	)
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 连接数据库
	db, err := model.InitDB(&cfg.Database)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 执行迁移
	log.Println("执行数据库迁移...")
	if err := model.AutoMigrate(db); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	log.Println("数据库迁移完成")

	// 清空数据（如果指定）
	if *clean {
		if err := cleanData(db); err != nil {
			log.Fatalf("清空数据失败: %v", err)
		}
		log.Println("数据已清空")
	}

	// 确保根组织存在
	if err := model.InitRootOrganization(db); err != nil {
		log.Fatalf("初始化根组织失败: %v", err)
	}

	// 导入种子数据
	if *all || *orgs {
		if err := seedOrganizations(db); err != nil {
			log.Printf("导入组织失败: %v", err)
		}
	}

	if *all || *users {
		if err := seedUsers(db); err != nil {
			log.Printf("导入用户失败: %v", err)
		}
	}

	if *all || *cases {
		if err := seedMissingPersons(db); err != nil {
			log.Printf("导入走失人员失败: %v", err)
		}
	}

	if *all || *dialects {
		if err := seedDialects(db); err != nil {
			log.Printf("导入方言失败: %v", err)
		}
	}

	if *all || *tasks {
		if err := seedTasks(db); err != nil {
			log.Printf("导入任务失败: %v", err)
		}
	}

	log.Println("\n种子数据导入完成!")
}

// cleanData 清空所有业务数据（保留根组织和超级管理员）
func cleanData(db *gorm.DB) error {
	log.Println("警告: 正在清空业务数据...")
	
	// 按依赖顺序删除
	tables := []interface{}{
		&model.TaskComment{},
		&model.TaskLog{},
		&model.TaskAttachment{},
		&model.Task{},
		&model.WorkflowHistory{},
		&model.WorkflowInstance{},
		&model.WorkflowStep{},
		&model.Workflow{},
		&model.DialectPlayLog{},
		&model.DialectLike{},
		&model.DialectComment{},
		&model.Dialect{},
		&model.MissingPersonTrack{},
		&model.MissingPhoto{},
		&model.MissingPerson{},
		&model.UserProfile{},
		&model.User{},
		&model.OrgStats{},
		&model.Organization{},
	}

	for _, table := range tables {
		if err := db.Where("1 = 1").Delete(table).Error; err != nil {
			return err
		}
	}

	return nil
}

// seedOrganizations 导入示例组织
func seedOrganizations(db *gorm.DB) error {
	log.Println("导入示例组织...")

	// 获取根组织
	var rootOrg model.Organization
	if err := db.Where("type = ?", model.OrgTypeRoot).First(&rootOrg).Error; err != nil {
		return err
	}

	orgs := []model.Organization{
		{
			Name:     "北京志愿者协会",
			Code:     "BJ-001",
			Type:     model.OrgTypeProvince,
			Level:    2,
			ParentID: &rootOrg.ID,
			Province: "北京市",
			Status:   model.OrgStatusActive,
		},
		{
			Name:     "上海志愿者协会",
			Code:     "SH-001",
			Type:     model.OrgTypeProvince,
			Level:    2,
			ParentID: &rootOrg.ID,
			Province: "上海市",
			Status:   model.OrgStatusActive,
		},
		{
			Name:     "广东志愿者协会",
			Code:     "GD-001",
			Type:     model.OrgTypeProvince,
			Level:    2,
			ParentID: &rootOrg.ID,
			Province: "广东省",
			Status:   model.OrgStatusActive,
		},
		{
			Name:     "深圳市志愿者协会",
			Code:     "GD-SZ-001",
			Type:     model.OrgTypeCity,
			Level:    3,
			Province: "广东省",
			City:     "深圳市",
			Status:   model.OrgStatusActive,
		},
	}

	for _, org := range orgs {
		if org.ParentID == nil {
			org.ParentID = &rootOrg.ID
		}
		
		// 检查是否已存在
		var existing model.Organization
		result := db.Where("code = ?", org.Code).First(&existing)
		if result.Error == nil {
			log.Printf("  组织 %s 已存在，跳过", org.Code)
			continue
		}

		if err := db.Create(&org).Error; err != nil {
			log.Printf("  创建组织 %s 失败: %v", org.Code, err)
			continue
		}
		log.Printf("  创建组织: %s", org.Name)
	}

	return nil
}

// seedUsers 导入示例用户
func seedUsers(db *gorm.DB) error {
	log.Println("导入示例用户...")

	// 获取一个组织
	var org model.Organization
	if err := db.Where("code = ?", "BJ-001").First(&org).Error; err != nil {
		log.Println("  未找到示例组织，跳过用户导入")
		return nil
	}

	users := []struct {
		model.User
		PlainPassword string
	}{
		{
			User: model.User{
				Nickname: "张管理员",
				RealName: "张三",
				Phone:    "13800138001",
				Email:    "admin1@cntunyuan.com",
				Role:     model.RoleAdmin,
				Status:   model.UserStatusActive,
				OrgID:    &org.ID,
			},
			PlainPassword: "admin123",
		},
		{
			User: model.User{
				Nickname: "李管理",
				RealName: "李四",
				Phone:    "13800138002",
				Email:    "manager1@cntunyuan.com",
				Role:     model.RoleManager,
				Status:   model.UserStatusActive,
				OrgID:    &org.ID,
			},
			PlainPassword: "manager123",
		},
		{
			User: model.User{
				Nickname: "王志愿者",
				RealName: "王五",
				Phone:    "13800138003",
				Email:    "volunteer1@cntunyuan.com",
				Role:     model.RoleVolunteer,
				Status:   model.UserStatusActive,
				OrgID:    &org.ID,
			},
			PlainPassword: "volunteer123",
		},
		{
			User: model.User{
				Nickname: "赵志愿者",
				RealName: "赵六",
				Phone:    "13800138004",
				Email:    "volunteer2@cntunyuan.com",
				Role:     model.RoleVolunteer,
				Status:   model.UserStatusActive,
				OrgID:    &org.ID,
			},
			PlainPassword: "volunteer123",
		},
	}

	for _, u := range users {
		// 检查是否已存在
		var existing model.User
		result := db.Where("phone = ?", u.Phone).First(&existing)
		if result.Error == nil {
			log.Printf("  用户 %s 已存在，跳过", u.Phone)
			continue
		}

		// 设置密码
		u.User.Password = u.PlainPassword

		if err := db.Create(&u.User).Error; err != nil {
			log.Printf("  创建用户 %s 失败: %v", u.Phone, err)
			continue
		}
		log.Printf("  创建用户: %s (%s)", u.Nickname, u.Role)
	}

	return nil
}

// seedMissingPersons 导入示例走失人员
func seedMissingPersons(db *gorm.DB) error {
	log.Println("导入示例走失人员...")

	// 获取一个用户
	var user model.User
	if err := db.Where("role = ?", model.RoleManager).First(&user).Error; err != nil {
		log.Println("  未找到示例用户，跳过导入")
		return nil
	}

	// 获取组织
	var org model.Organization
	if err := db.Where("id = ?", user.OrgID).First(&org).Error; err != nil {
		return err
	}

	cases := []model.MissingPerson{
		{
			Name:         "测试走失人员1",
			Gender:       "male",
			Age:          8,
			Height:       130,
			CaseType:     "lost_child",
			Status:       "searching",
			MissingTime:  time.Now().AddDate(0, -1, 0),
			Province:     org.Province,
			City:         org.City,
			MissingLocation: "某某公园",
			Appearance:   "身穿蓝色外套，黑色裤子",
			Clothing:     "蓝色外套，黑色裤子，运动鞋",
			ReporterName: user.RealName,
			ReporterPhone: user.Phone,
			ReporterID:   &user.ID,
			OrgID:        user.OrgID,
		},
		{
			Name:         "测试走失人员2",
			Gender:       "female",
			Age:          75,
			Height:       160,
			CaseType:     "lost_elderly",
			Status:       "searching",
			MissingTime:  time.Now().AddDate(0, 0, -15),
			Province:     org.Province,
			City:         org.City,
			MissingLocation: "某某小区",
			Appearance:   "白发，戴眼镜，穿灰色毛衣",
			Clothing:     "灰色毛衣，深色裤子",
			SpecialFeatures: "患有阿尔茨海默病",
			ReporterName: user.RealName,
			ReporterPhone: user.Phone,
			ReporterID:   &user.ID,
			OrgID:        user.OrgID,
		},
	}

	for _, c := range cases {
		// 检查是否已存在
		var existing model.MissingPerson
		result := db.Where("name = ? AND reporter_id = ?", c.Name, c.ReporterID).First(&existing)
		if result.Error == nil {
			log.Printf("  走失人员 %s 已存在，跳过", c.Name)
			continue
		}

		if err := db.Create(&c).Error; err != nil {
			log.Printf("  创建走失人员 %s 失败: %v", c.Name, err)
			continue
		}
		log.Printf("  创建走失人员: %s", c.Name)
	}

	return nil
}

// seedDialects 导入示例方言
func seedDialects(db *gorm.DB) error {
	log.Println("导入示例方言...")

	// 获取一个用户
	var user model.User
	if err := db.Where("role = ?", model.RoleVolunteer).First(&user).Error; err != nil {
		log.Println("  未找到示例用户，跳过导入")
		return nil
	}

	dialects := []model.Dialect{
		{
			Title:       "北京话示例",
			Description: "北京方言语音示例",
			Province:    "北京市",
			Address:     "北京市朝阳区",
			AudioURL:    "https://example.com/audio1.mp3",
			Duration:    18,
			CollectorID: &user.ID,
			OrgID:       user.OrgID,
			Status:      "active",
		},
		{
			Title:       "上海话示例",
			Description: "上海方言语音示例",
			Province:    "上海市",
			Address:     "上海市浦东新区",
			AudioURL:    "https://example.com/audio2.mp3",
			Duration:    20,
			CollectorID: &user.ID,
			OrgID:       user.OrgID,
			Status:      "active",
		},
		{
			Title:       "粤语示例",
			Description: "广东方言语音示例",
			Province:    "广东省",
			City:        "广州市",
			Address:     "广州市天河区",
			AudioURL:    "https://example.com/audio3.mp3",
			Duration:    15,
			CollectorID: &user.ID,
			OrgID:       user.OrgID,
			Status:      "active",
		},
	}

	for _, d := range dialects {
		// 检查是否已存在
		var existing model.Dialect
		result := db.Where("title = ? AND collector_id = ?", d.Title, d.CollectorID).First(&existing)
		if result.Error == nil {
			log.Printf("  方言 %s 已存在，跳过", d.Title)
			continue
		}

		if err := db.Create(&d).Error; err != nil {
			log.Printf("  创建方言 %s 失败: %v", d.Title, err)
			continue
		}
		log.Printf("  创建方言: %s", d.Title)
	}

	return nil
}

// seedTasks 导入示例任务
func seedTasks(db *gorm.DB) error {
	log.Println("导入示例任务...")

	// 获取一个用户
	var user model.User
	if err := db.Where("role = ?", model.RoleManager).First(&user).Error; err != nil {
		log.Println("  未找到示例用户，跳过导入")
		return nil
	}

	// 获取一个走失人员
	var mp model.MissingPerson
	if err := db.Where("reporter_id = ?", user.ID).First(&mp).Error; err != nil {
		log.Println("  未找到示例走失人员，跳过导入")
		return nil
	}

	tasks := []model.Task{
		{
			Title:           "寻找" + mp.Name,
			Description:     "协助寻找走失人员，请在附近区域进行排查",
			Type:            "search",
			Priority:        "high",
			Status:          "pending",
			MissingPersonID: &mp.ID,
			CreatorID:       user.ID,
			OrgID:           *user.OrgID,
			Deadline:        time.Now().AddDate(0, 0, 7),
		},
	}

	for _, t := range tasks {
		// 检查是否已存在
		var existing model.Task
		result := db.Where("title = ? AND creator_id = ?", t.Title, t.CreatorID).First(&existing)
		if result.Error == nil {
			log.Printf("  任务 %s 已存在，跳过", t.Title)
			continue
		}

		if err := db.Create(&t).Error; err != nil {
			log.Printf("  创建任务 %s 失败: %v", t.Title, err)
			continue
		}
		log.Printf("  创建任务: %s", t.Title)
	}

	return nil
}
