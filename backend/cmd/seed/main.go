package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/database"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
)

func main() {
	var (
		all  = flag.Bool("all", false, "Import all seed data")
		orgs = flag.Bool("orgs", false, "Import organizations only")
		users = flag.Bool("users", false, "Import users only")
		cases = flag.Bool("cases", false, "Import missing persons only")
		dialects = flag.Bool("dialects", false, "Import dialects only")
		tasks = flag.Bool("tasks", false, "Import tasks only")
		clean = flag.Bool("clean", false, "Clean data before import")
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
	cfg, err := config.LoadConfig(".")
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

	logger.Info("Starting seed data import...")

	// 如果需要，先清理数据
	if *clean {
		logger.Info("Cleaning existing data...")
		// TODO: 实现数据清理
	}

	// 导入数据
	imported := false

	if *all || *orgs {
		if err := importOrganizations(db); err != nil {
			logger.Error("Failed to import organizations", logger.Err(err))
			os.Exit(1)
		}
		imported = true
	}

	if *all || *users {
		if err := importUsers(db); err != nil {
			logger.Error("Failed to import users", logger.Err(err))
			os.Exit(1)
		}
		imported = true
	}

	if *all || *cases {
		if err := importMissingPersons(db); err != nil {
			logger.Error("Failed to import missing persons", logger.Err(err))
			os.Exit(1)
		}
		imported = true
	}

	if *all || *dialects {
		if err := importDialects(db); err != nil {
			logger.Error("Failed to import dialects", logger.Err(err))
			os.Exit(1)
		}
		imported = true
	}

	if *all || *tasks {
		if err := importTasks(db); err != nil {
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

// importOrganizations 导入组织数据
func importOrganizations(db interface{}) error {
	logger.Info("Importing organizations...")
	// TODO: 实现组织数据导入
	return nil
}

// importUsers 导入用户数据
func importUsers(db interface{}) error {
	logger.Info("Importing users...")
	// TODO: 实现用户数据导入
	return nil
}

// importMissingPersons 导入走失人员数据
func importMissingPersons(db interface{}) error {
	logger.Info("Importing missing persons...")
	// TODO: 实现走失人员数据导入
	return nil
}

// importDialects 导入方言数据
func importDialects(db interface{}) error {
	logger.Info("Importing dialects...")
	// TODO: 实现方言数据导入
	return nil
}

// importTasks 导入任务数据
func importTasks(db interface{}) error {
	logger.Info("Importing tasks...")
	// TODO: 实现任务数据导入
	return nil
}
