package router

import (
	"github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/handler"
	"github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

// Router 路由管理器
type Router struct {
	engine               *gin.Engine
	authHandler          *handler.AuthHandler
	setupHandler         *handler.SetupHandler
	userHandler          *handler.UserHandler
	organizationHandler  *handler.OrganizationHandler
	missingPersonHandler *handler.MissingPersonHandler
	dialectHandler       *handler.DialectHandler
	taskHandler          *handler.TaskHandler
	uploadHandler        *handler.UploadHandler
	dashboardHandler     *handler.DashboardHandler
	authMiddleware       *middleware.AuthMiddleware
}

// NewRouter 创建路由管理器
func NewRouter(
	authHandler *handler.AuthHandler,
	setupHandler *handler.SetupHandler,
	userHandler *handler.UserHandler,
	organizationHandler *handler.OrganizationHandler,
	missingPersonHandler *handler.MissingPersonHandler,
	dialectHandler *handler.DialectHandler,
	taskHandler *handler.TaskHandler,
	uploadHandler *handler.UploadHandler,
	dashboardHandler *handler.DashboardHandler,
	authMiddleware *middleware.AuthMiddleware,
) *Router {
	engine := gin.New()

	// 全局中间件
	engine.Use(middleware.RecoveryMiddleware())
	engine.Use(middleware.CORSMiddleware())
	engine.Use(middleware.RequestLoggerMiddleware())

	return &Router{
		engine:               engine,
		authHandler:          authHandler,
		setupHandler:         setupHandler,
		userHandler:          userHandler,
		organizationHandler:  organizationHandler,
		missingPersonHandler: missingPersonHandler,
		dialectHandler:       dialectHandler,
		taskHandler:          taskHandler,
		uploadHandler:        uploadHandler,
		dashboardHandler:     dashboardHandler,
		authMiddleware:       authMiddleware,
	}
}

// Setup 设置路由
func (r *Router) Setup() {
	// API v1 路由组
	api := r.engine.Group("/api/v1")

	// 健康检查
	api.GET("/health", r.healthCheck)

	// 注册初始化路由（不需要认证）
	r.setupHandler.RegisterRoutes(api)

	// 注册各个模块路由
	r.authHandler.RegisterRoutes(api)
	r.userHandler.RegisterRoutes(api, r.authMiddleware)
	r.organizationHandler.RegisterRoutes(api, r.authMiddleware)
	r.missingPersonHandler.RegisterRoutes(api, r.authMiddleware)
	r.dialectHandler.RegisterRoutes(api, r.authMiddleware)
	r.taskHandler.RegisterRoutes(api, r.authMiddleware)
	r.uploadHandler.RegisterRoutes(api, r.authMiddleware)
	r.dashboardHandler.RegisterRoutes(api, r.authMiddleware)

	// 404 处理
	r.engine.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": 404, "message": "route not found"})
	})
}

// GetEngine 获取 gin 引擎
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// healthCheck 健康检查
func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"code":    0,
		"message": "healthy",
		"data": gin.H{
			"status": "ok",
		},
	})
}
