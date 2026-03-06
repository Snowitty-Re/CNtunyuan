package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/database"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	var (
		all      = flag.Bool("all", false, "Import all seed data")
		orgs     = flag.Bool("orgs", false, "Import organizations only")
		users    = flag.Bool("users", false, "Import users only")
		cases    = flag.Bool("cases", false, "Import missing persons only")
		dialects = flag.Bool("dialects", false, "Import dialects only")
		tasks    = flag.Bool("tasks", false, "Import tasks only")
		clean    = flag.Bool("clean", false, "Clean data before import")
		count    = flag.Int("count", 50, "Number of records to generate per type")
	)
	flag.Parse()

	// 初始化日志
	logCfg := &config.LogConfig{
		Level:  "info",
		Format: "console",
	}
	if err := logger.Init(logCfg); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// 加载配置
	cfg, err := config.LoadConfig("")
	if err != nil {
		logger.Error("Failed to load config", logger.Err(err))
		os.Exit(1)
	}

	// 连接数据库
	db, err := database.NewDatabase(&cfg.Database)
	if err != nil {
		logger.Error("Failed to connect database", logger.Err(err))
		os.Exit(1)
	}

	// 如果需要，先清理数据
	if *clean {
		logger.Info("Cleaning existing data...")
		if err := cleanData(db); err != nil {
			logger.Error("Failed to clean data", logger.Err(err))
			os.Exit(1)
		}
	}

	logger.Info("Starting seed data import...", logger.Int("count", *count))

	// 导入数据
	imported := false

	if *all || *orgs {
		if err := importOrganizations(db, *count); err != nil {
			logger.Error("Failed to import organizations", logger.Err(err))
			os.Exit(1)
		}
		imported = true
	}

	if *all || *users {
		if err := importUsers(db, *count); err != nil {
			logger.Error("Failed to import users", logger.Err(err))
			os.Exit(1)
		}
		imported = true
	}

	if *all || *cases {
		if err := importMissingPersons(db, *count); err != nil {
			logger.Error("Failed to import missing persons", logger.Err(err))
			os.Exit(1)
		}
		imported = true
	}

	if *all || *dialects {
		if err := importDialects(db, *count); err != nil {
			logger.Error("Failed to import dialects", logger.Err(err))
			os.Exit(1)
		}
		imported = true
	}

	if *all || *tasks {
		if err := importTasks(db, *count); err != nil {
			logger.Error("Failed to import tasks", logger.Err(err))
			os.Exit(1)
		}
		imported = true
	}

	if !imported {
		logger.Info("No data type specified. Use -all or specific flags (-orgs, -users, etc.)")
		flag.Usage()
		os.Exit(1)
	}

	logger.Info("Seed data import completed successfully!")
}

// 数据生成辅助变量
var (
	provinces = []string{"北京", "上海", "广东", "浙江", "江苏", "山东", "河南", "四川", "湖北", "湖南"}
	cities    = []string{"北京市", "上海市", "广州市", "深圳市", "杭州市", "南京市", "济南市", "郑州市", "成都市", "武汉市", "长沙市"}
	districts = []string{"朝阳区", "海淀区", "浦东新区", "天河区", "南山区", "西湖区", "鼓楼区", "历下区", "金水区", "锦江区"}

	orgNames    = []string{"志愿者协会", "寻亲服务中心", "救助站", "公益组织", "救援队", "社区服务中心", "民政服务中心"}
	firstNames  = []string{"伟", "芳", "娜", "敏", "静", "强", "磊", "洋", "艳", "杰", "勇", "军", "平", "刚", "桂"}
	lastNames   = []string{"张", "王", "李", "刘", "陈", "杨", "黄", "赵", "周", "吴", "徐", "孙", "马", "朱", "胡"}
	genders     = []string{"男", "女"}
	roles       = []string{"volunteer", "manager", "admin"}
	statuses    = []string{"active", "inactive"}
	caseStatuses = []string{"missing", "searching", "found", "reunited"}
	urgencies   = []string{"low", "medium", "high", "critical"}
	taskTypes   = []string{"search", "verify", "assist", "follow", "interview"}
	taskStatuses = []string{"draft", "pending", "assigned", "processing", "completed"}
	dialectTypes = []string{"phrase", "story", "song", "daily"}
	regions     = []string{"北京话", "上海话", "粤语", "四川话", "河南话", "山东话", "东北话", "湖南话", "湖北话", "江浙话"}
)

// randomChoice 随机选择一个元素
func randomChoice(items []string) string {
	return items[rand.Intn(len(items))]
}

// randomPhone 生成随机手机号
func randomPhone() string {
	prefixes := []string{"138", "139", "137", "136", "135", "134", "159", "158", "157", "150", "151", "152"}
	prefix := prefixes[rand.Intn(len(prefixes))]
	return fmt.Sprintf("%s%08d", prefix, rand.Intn(100000000))
}

// randomName 生成随机姓名
func randomName() string {
	return lastNames[rand.Intn(len(lastNames))] + firstNames[rand.Intn(len(firstNames))]
}

// randomTime 生成随机时间（过去1年内）
func randomTime() time.Time {
	days := rand.Intn(365)
	return time.Now().AddDate(0, 0, -days)
}

// cleanData 清理现有数据
func cleanData(db *gorm.DB) error {
	tables := []string{
		"ty_task_comments", "ty_task_logs", "ty_task_attachments", "ty_tasks",
		"ty_dialect_play_logs", "ty_dialect_likes", "ty_dialect_comments", "ty_dialects",
		"ty_missing_person_tracks", "ty_missing_persons",
		"ty_user_permissions", "ty_permissions", "ty_users",
		"ty_org_stats", "ty_organizations",
		"ty_files",
	}
	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error; err != nil {
			logger.Warn("Failed to truncate table", logger.String("table", table), logger.Err(err))
		}
	}
	logger.Info("Data cleaned successfully")
	return nil
}

// importOrganizations 导入组织数据
func importOrganizations(db *gorm.DB, count int) error {
	logger.Info("Importing organizations...")

	// 创建根组织
	rootOrg := &entity.Organization{
		BaseEntity: entity.BaseEntity{ID: uuid.New().String()},
		Name:       "团圆寻亲志愿者总会",
		Code:       "ROOT",
		Type:       "root",
		Level:      1,
		Status:     "active",
	}
	if err := db.Create(rootOrg).Error; err != nil {
		return err
	}

	// 创建省级组织
	for i := 0; i < count && i < len(provinces); i++ {
		org := &entity.Organization{
			BaseEntity: entity.BaseEntity{ID: uuid.New().String()},
			Name:       fmt.Sprintf("%s%s", provinces[i], orgNames[0]),
			Code:       fmt.Sprintf("PROV_%02d", i+1),
			Type:       "province",
			Level:      2,
			ParentID:   &rootOrg.ID,
			Status:     "active",
			Address:    fmt.Sprintf("%s省/市", provinces[i]),
		}
		if err := db.Create(org).Error; err != nil {
			return err
		}
	}

	// 创建市级组织
	var parentOrgs []entity.Organization
	if err := db.Where("type = ?", "province").Find(&parentOrgs).Error; err != nil {
		return err
	}

	orgCount := count
	for i, parent := range parentOrgs {
		if orgCount <= 0 {
			break
		}
		for j := 0; j < 3 && orgCount > 0; j++ {
			cityIndex := (i*3 + j) % len(cities)
			org := &entity.Organization{
				BaseEntity: entity.BaseEntity{ID: uuid.New().String()},
				Name:       fmt.Sprintf("%s%s", cities[cityIndex], orgNames[rand.Intn(len(orgNames))]),
				Code:       fmt.Sprintf("CITY_%02d%02d", i+1, j+1),
				Type:       "city",
				Level:      3,
				ParentID:   &parent.ID,
				Status:     "active",
				Address:    fmt.Sprintf("%s%s", cities[cityIndex], districts[rand.Intn(len(districts))]),
			}
			if err := db.Create(org).Error; err != nil {
				return err
			}
			orgCount--
		}
	}

	logger.Info("Organizations imported successfully")
	return nil
}

// importUsers 导入用户数据
func importUsers(db *gorm.DB, count int) error {
	logger.Info("Importing users...")

	// 获取组织列表
	var orgs []entity.Organization
	if err := db.Find(&orgs).Error; err != nil {
		return err
	}

	if len(orgs) == 0 {
		return fmt.Errorf("no organizations found, please import organizations first")
	}

	// 创建超级管理员
	adminUser := &entity.User{
		BaseEntity: entity.BaseEntity{ID: uuid.New().String()},
		Nickname:   "超级管理员",
		Phone:      "13800138000",
		Email:      "admin@cntuanyuan.com",
		Password:   "",
		Role:       entity.RoleSuperAdmin,
		Status:     entity.UserStatusActive,
		OrgID:      orgs[0].ID,
	}
	adminUser.SetPassword("admin123")
	if err := db.Create(adminUser).Error; err != nil {
		return err
	}

	// 创建普通用户
	for i := 0; i < count; i++ {
		role := entity.RoleVolunteer
		if i < 5 {
			role = entity.RoleAdmin
		} else if i < 15 {
			role = entity.RoleManager
		}

		user := &entity.User{
			BaseEntity: entity.BaseEntity{ID: uuid.New().String()},
			Nickname:   randomName(),
			Phone:      randomPhone(),
			Email:      fmt.Sprintf("user%d@example.com", i+1),
			Password:   "",
			Role:       role,
			Status:     entity.UserStatusActive,
			OrgID:      orgs[rand.Intn(len(orgs))].ID,
			RealName:   randomName(),
			Gender:     randomChoice(genders),
			Address:    fmt.Sprintf("%s%s%s", randomChoice(provinces), randomChoice(cities), randomChoice(districts)),
		}
		user.SetPassword("123456")
		if err := db.Create(user).Error; err != nil {
			return err
		}
	}

	logger.Info("Users imported successfully")
	return nil
}

// importMissingPersons 导入走失人员数据
func importMissingPersons(db *gorm.DB, count int) error {
	logger.Info("Importing missing persons...")

	// 获取用户和组织列表
	var users []entity.User
	var orgs []entity.Organization
	if err := db.Find(&users).Error; err != nil {
		return err
	}
	if err := db.Find(&orgs).Error; err != nil {
		return err
	}

	if len(users) == 0 || len(orgs) == 0 {
		return fmt.Errorf("no users or organizations found")
	}

	clothes := []string{"红色外套", "蓝色T恤", "黑色裤子", "白色衬衫", "灰色运动服", "黄色夹克"}
	features := []string{"戴眼镜", "短发", "长发", "有胎记", "个子较高", "体型偏瘦"}

	for i := 0; i < count; i++ {
		status := entity.MissingStatus(randomChoice(caseStatuses))
		mp := &entity.MissingPerson{
			BaseEntity:   entity.BaseEntity{ID: uuid.New().String()},
			Name:         randomName(),
			Gender:       randomChoice(genders),
			Age:          10 + rand.Intn(80),
			Height:       150 + rand.Intn(50),
			Weight:       40 + rand.Intn(60),
			Description:  fmt.Sprintf("走失人员特征描述 %d", i+1),
			PhotoUrl:     fmt.Sprintf("https://example.com/photo%d.jpg", i+1),
			MissingTime:  randomTime(),
			Province:     randomChoice(provinces),
			City:         randomChoice(cities),
			District:     randomChoice(districts),
			Address:      fmt.Sprintf("%s街道%d号", randomChoice(districts), rand.Intn(1000)),
			Clothes:      clothes[rand.Intn(len(clothes))],
			Features:     features[rand.Intn(len(features))],
			ContactName:  randomName(),
			ContactPhone: randomPhone(),
			ContactRel:   randomChoice([]string{"父亲", "母亲", "儿子", "女儿", "配偶"}),
			Status:       status,
			Urgency:      entity.UrgencyLevel(randomChoice(urgencies)),
			ReporterID:   users[rand.Intn(len(users))].ID,
			OrgID:        orgs[rand.Intn(len(orgs))].ID,
		}

		// 如果已找到，添加找到信息
		if status == entity.MissingStatusFound || status == entity.MissingStatusReunited {
			foundTime := mp.MissingTime.AddDate(0, 0, rand.Intn(30)+1)
			mp.FoundTime = &foundTime
			mp.FoundLocation = randomChoice(cities)
			mp.FoundNote = "在热心市民帮助下成功找到"
		}

		if err := db.Create(mp).Error; err != nil {
			return err
		}
	}

	logger.Info("Missing persons imported successfully")
	return nil
}

// importDialects 导入方言数据
func importDialects(db *gorm.DB, count int) error {
	logger.Info("Importing dialects...")

	// 获取用户和组织列表
	var users []entity.User
	var orgs []entity.Organization
	if err := db.Find(&users).Error; err != nil {
		return err
	}
	if err := db.Find(&orgs).Error; err != nil {
		return err
	}

	if len(users) == 0 || len(orgs) == 0 {
		return fmt.Errorf("no users or organizations found")
	}

	titles := []string{"家乡问候", "日常用语", "童年记忆", "地方童谣", "方言故事", "求助录音", "亲情呼唤"}
	contents := []string{
		"吃了吗？您呐！",
		"这孩子长得真俊！",
		"天儿不早了，赶紧回家吧",
		"咱们一块儿去赶集",
		"这饭做得真香啊",
		"你慢点儿走，小心路滑",
	}

	for i := 0; i < count; i++ {
		dialect := &entity.Dialect{
			BaseEntity: entity.BaseEntity{ID: uuid.New().String()},
			Title:      fmt.Sprintf("%s-%d", titles[rand.Intn(len(titles))], i+1),
			Content:    contents[rand.Intn(len(contents))],
			Region:     randomChoice(regions),
			Province:   randomChoice(provinces),
			City:       randomChoice(cities),
			DialectType: entity.DialectType(randomChoice(dialectTypes)),
			AudioUrl:   fmt.Sprintf("https://example.com/audio%d.mp3", i+1),
			Duration:   10 + rand.Intn(280),
			FileSize:   100000 + rand.Intn(900000),
			Format:     "mp3",
			Status:     entity.DialectStatusActive,
			IsFeatured: rand.Intn(10) == 0, // 10% 概率设为精选
			PlayCount:  rand.Intn(1000),
			LikeCount:  rand.Intn(500),
			Tags:       `["方言", "寻亲", "语音"]`,
			Description: fmt.Sprintf("这是一段%s的方言录音，用于寻亲识别", randomChoice(regions)),
			UploaderID: users[rand.Intn(len(users))].ID,
			OrgID:      orgs[rand.Intn(len(orgs))].ID,
		}
		if err := db.Create(dialect).Error; err != nil {
			return err
		}
	}

	logger.Info("Dialects imported successfully")
	return nil
}

// importTasks 导入任务数据
func importTasks(db *gorm.DB, count int) error {
	logger.Info("Importing tasks...")

	// 获取用户、组织和走失人员列表
	var users []entity.User
	var orgs []entity.Organization
	var missingPersons []entity.MissingPerson
	if err := db.Find(&users).Error; err != nil {
		return err
	}
	if err := db.Find(&orgs).Error; err != nil {
		return err
	}
	if err := db.Find(&missingPersons).Error; err != nil {
		return err
	}

	if len(users) == 0 || len(orgs) == 0 {
		return fmt.Errorf("no users or organizations found")
	}

	titles := []string{"走访调查", "信息核实", "线索排查", "家属沟通", "区域搜索", "资料整理", "志愿者培训"}
	descriptions := []string{
		"需要对目标区域进行详细走访调查",
		"核实走失人员相关信息的真实性",
		"排查热心群众提供的线索",
		"与走失人员家属保持沟通",
		"在指定区域内进行搜索",
		"整理案件相关资料",
	}

	for i := 0; i < count; i++ {
		status := entity.TaskStatus(randomChoice(taskStatuses))
		creator := users[rand.Intn(len(users))]
		var assigneeID *string
		if status != entity.TaskStatusDraft && status != entity.TaskStatusPending {
			assignee := users[rand.Intn(len(users))].ID
			assigneeID = &assignee
		}

		task := &entity.Task{
			BaseEntity: entity.BaseEntity{ID: uuid.New().String()},
			Title:      fmt.Sprintf("%s-%d", titles[rand.Intn(len(titles))], i+1),
			Description: descriptions[rand.Intn(len(descriptions))],
			Type:       entity.TaskType(randomChoice(taskTypes)),
			Priority:   entity.TaskPriority(randomChoice(urgencies)),
			Status:     status,
			CreatorID:  creator.ID,
			AssigneeID: assigneeID,
			OrgID:      orgs[rand.Intn(len(orgs))].ID,
			Location:   fmt.Sprintf("%s%s", randomChoice(cities), randomChoice(districts)),
			Province:   randomChoice(provinces),
			City:       randomChoice(cities),
			District:   randomChoice(districts),
			Progress:   rand.Intn(101),
		}

		// 关联走失人员（50%概率）
		if len(missingPersons) > 0 && rand.Intn(2) == 0 {
			mpID := missingPersons[rand.Intn(len(missingPersons))].ID
			task.MissingPersonID = &mpID
		}

		// 设置 JSON 字段
		task.ResultPhotos = "[]"

		// 设置时间
		if status == entity.TaskStatusProcessing || status == entity.TaskStatusCompleted {
			startedAt := randomTime()
			task.StartedAt = &startedAt
		}
		if status == entity.TaskStatusCompleted {
			completedAt := time.Now()
			task.CompletedAt = &completedAt
			task.Result = "任务已完成，目标达成"
			task.Progress = 100
		}
		// 设置截止日期
		deadline := time.Now().AddDate(0, 0, rand.Intn(30)+7)
		task.Deadline = &deadline

		if err := db.Create(task).Error; err != nil {
			return err
		}
	}

	logger.Info("Tasks imported successfully")
	return nil
}
