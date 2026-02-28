package main

// @title 团圆寻亲志愿者系统 API
// @version 1.0.0
// @description 团圆寻亲志愿者系统后端 API 文档
// @termsOfService https://github.com/Snowitty-Re/CNtunyuan

// @contact.name CNtunyuan Team
// @contact.url https://github.com/Snowitty-Re/CNtunyuan

// @license.name MIT

// @host localhost:8080
// @BasePath /api/v1

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/Snowitty-Re/CNtunyuan/docs"
	"github.com/Snowitty-Re/CNtunyuan/internal/api"
	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/model"
	"github.com/Snowitty-Re/CNtunyuan/internal/repository"
	"github.com/Snowitty-Re/CNtunyuan/internal/service"
	"github.com/Snowitty-Re/CNtunyuan/pkg/auth"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func main() {
	// 命令行参数
	var (
		migrate = flag.Bool("migrate", false, "执行数据库迁移")
		init    = flag.Bool("init", false, "初始化基础数据（根组织）")
	)
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 设置gin模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化数据库
	db, err := model.InitDB(&cfg.Database)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 自动迁移（仅当 -migrate 参数指定时）
	if *migrate {
		if err := model.AutoMigrate(db); err != nil {
			log.Fatalf("数据库迁移失败: %v", err)
		}
		log.Println("数据库迁移完成")
		log.Println("")
		log.Println("=======================================")
		log.Println("提示: 数据库结构已更新")
		log.Println("")
		log.Println("如需初始化超级管理员，请使用以下命令:")
		log.Println("  go run cmd/initdata/main.go -exec")
		log.Println("")
		log.Println("或生成SQL文件手动导入:")
		log.Println("  go run cmd/initdata/main.go -gen")
		log.Println("=======================================")
		return
	}

	// 仅初始化基础数据（根组织）
	if *init {
		if err := model.InitRootOrganization(db); err != nil {
			log.Fatalf("初始化根组织失败: %v", err)
		}
		log.Println("根组织初始化完成")
		log.Println("")
		log.Println("=======================================")
		log.Println("提示: 根组织已创建")
		log.Println("")
		log.Println("如需创建超级管理员，请使用:")
		log.Println("  go run cmd/initdata/main.go -exec")
		log.Println("=======================================")
		return
	}

	// 检查是否已有超级管理员（启动时检查，给出提示）
	go checkSuperAdmin(db)

	// 初始化Redis（可选）
	var rdb *redis.Client
	if cfg.Redis.Host != "" {
		rdb = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})

		// 测试Redis连接
		ctx := context.Background()
		if err := rdb.Ping(ctx).Err(); err != nil {
			log.Printf("Redis连接警告: %v，将使用内存缓存", err)
		}
	} else {
		log.Println("Redis未配置，将使用内存缓存")
	}

	// 初始化JWT
	jwtAuth := auth.NewJWTAuth(&cfg.JWT)

	// 初始化仓库
	userRepo := repository.NewUserRepository(db)
	orgRepo := repository.NewOrganizationRepository(db)
	mpRepo := repository.NewMissingPersonRepository(db)
	dialectRepo := repository.NewDialectRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	workflowRepo := repository.NewWorkflowRepository(db)

	// 初始化服务
	userService := service.NewUserService(userRepo, orgRepo)
	orgService := service.NewOrganizationService(orgRepo)
	mpService := service.NewMissingPersonService(mpRepo, orgRepo)
	dialectService := service.NewDialectService(dialectRepo)
	taskService := service.NewTaskService(taskRepo, userRepo, mpRepo, orgRepo)
	workflowService := service.NewWorkflowService(workflowRepo, userRepo)
	wechatService := service.NewWeChatService(cfg.WeChat.AppID, cfg.WeChat.AppSecret)

	// 创建路由
	router := api.NewRouter(
		userService,
		userService,
		orgService,
		mpService,
		dialectService,
		taskService,
		workflowService,
		wechatService,
		jwtAuth,
		rdb,
	)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router.GetEngine(),
	}

	// 启动服务器
	go func() {
		log.Printf("服务器启动，监听端口: %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")

	// 优雅关闭
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("服务器关闭失败: %v", err)
	}

	log.Println("服务器已退出")
}

// checkSuperAdmin 检查超级管理员是否存在
func checkSuperAdmin(db *gorm.DB) {
	var count int64
	if err := db.Model(&model.User{}).Where("role = ?", model.RoleSuperAdmin).Count(&count).Error; err != nil {
		log.Printf("检查超级管理员失败: %v", err)
		return
	}

	if count == 0 {
		log.Println("")
		log.Println("=======================================")
		log.Println("警告: 未检测到超级管理员账号")
		log.Println("")
		log.Println("请运行以下命令创建超级管理员:")
		log.Println("  go run cmd/initdata/main.go -exec")
		log.Println("")
		log.Println("或使用自定义参数:")
		log.Println("  go run cmd/initdata/main.go -exec -phone=13800138000 -password=admin123")
		log.Println("=======================================")
		log.Println("")
	}
}
