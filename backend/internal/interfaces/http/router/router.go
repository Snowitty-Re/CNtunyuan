package router

import (
	"github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/handler"
	"github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/middleware"
	pkgmiddleware "github.com/Snowitty-Re/CNtunyuan/pkg/middleware"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Router 路由管理器
type Router struct {
	engine                   *gin.Engine
	authHandler              *handler.AuthHandler
	userHandler              *handler.UserHandler
	organizationHandler      *handler.OrganizationHandler
	missingPersonHandler     *handler.MissingPersonHandler
	dialectHandler           *handler.DialectHandler
	taskHandler              *handler.TaskHandler
	uploadHandler            *handler.UploadHandler
	dashboardHandler         *handler.DashboardHandler
	auditHandler             *handler.AuditHandler
	workflowHandler          *handler.WorkflowHandler
	authMiddleware           *middleware.AuthMiddleware
	auditMiddleware          *middleware.AuditMiddleware
	dataPermissionMiddleware *middleware.DataPermissionMiddleware
}

// NewRouter 创建路由管理器
func NewRouter(
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	organizationHandler *handler.OrganizationHandler,
	missingPersonHandler *handler.MissingPersonHandler,
	dialectHandler *handler.DialectHandler,
	taskHandler *handler.TaskHandler,
	uploadHandler *handler.UploadHandler,
	dashboardHandler *handler.DashboardHandler,
	auditHandler *handler.AuditHandler,
	workflowHandler *handler.WorkflowHandler,
	authMiddleware *middleware.AuthMiddleware,
	auditMiddleware *middleware.AuditMiddleware,
	dataPermissionMiddleware *middleware.DataPermissionMiddleware,
) *Router {
	engine := gin.New()

	// 全局中间件（按执行顺序排列）
	// 1. 恢复中间件（捕获 panic）
	engine.Use(pkgmiddleware.RecoveryMiddleware())

	// 2. 追踪 ID 中间件
	engine.Use(pkgmiddleware.TraceIDMiddleware())

	// 3. 安全响应头中间件
	engine.Use(pkgmiddleware.SecurityHeadersMiddleware())

	// 4. CORS 中间件
	engine.Use(pkgmiddleware.CORSMiddleware())

	// 5. 请求大小限制（50MB）
	engine.Use(pkgmiddleware.RequestSizeMiddleware(50 * 1024 * 1024))

	// 6. 限流中间件（每秒100请求，突发200）
	engine.Use(pkgmiddleware.RateLimitMiddleware(100, 200))

	// 7. 结构化日志中间件
	engine.Use(pkgmiddleware.LoggingMiddleware())

	// 8. 统一错误处理中间件
	engine.Use(pkgmiddleware.ErrorHandlerMiddleware())

	return &Router{
		engine:                   engine,
		authHandler:              authHandler,
		userHandler:              userHandler,
		organizationHandler:      organizationHandler,
		missingPersonHandler:     missingPersonHandler,
		dialectHandler:           dialectHandler,
		taskHandler:              taskHandler,
		uploadHandler:            uploadHandler,
		dashboardHandler:         dashboardHandler,
		auditHandler:             auditHandler,
		workflowHandler:          workflowHandler,
		authMiddleware:           authMiddleware,
		auditMiddleware:          auditMiddleware,
		dataPermissionMiddleware: dataPermissionMiddleware,
	}
}

// Setup 设置路由
func (r *Router) Setup() {
	// API v1 路由组
	api := r.engine.Group("/api/v1")

	// 健康检查（不需要认证）
	api.GET("/health", r.healthCheck)
	api.GET("/health/detailed", r.detailedHealthCheck)

	// Prometheus 指标端点
	api.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 公开路由（不需要认证）
	public := api.Group("/")
	{
		public.GET("/", r.welcome)
	}

	// 注册各个模块路由
	r.authHandler.RegisterRoutes(api)
	r.userHandler.RegisterRoutes(api, r.authMiddleware)
	r.organizationHandler.RegisterRoutes(api, r.authMiddleware)
	r.missingPersonHandler.RegisterRoutes(api, r.authMiddleware)
	r.dialectHandler.RegisterRoutes(api, r.authMiddleware)
	r.taskHandler.RegisterRoutes(api, r.authMiddleware)
	r.uploadHandler.RegisterRoutes(api, r.authMiddleware)
	r.dashboardHandler.RegisterRoutes(api, r.authMiddleware)
	r.auditHandler.RegisterRoutes(api, r.authMiddleware)
	r.workflowHandler.RegisterRoutes(api, r.authMiddleware)

	// 404 处理
	r.engine.NoRoute(func(c *gin.Context) {
		response.NotFound(c, "route not found")
	})

	// 405 处理
	r.engine.NoMethod(func(c *gin.Context) {
		response.ErrorCodeWithMessage(c, 405, "method not allowed")
	})
}

// GetEngine 获取 gin 引擎
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// welcome 欢迎信息
func (r *Router) welcome(c *gin.Context) {
	response.Success(c, gin.H{
		"name":        "团圆寻亲志愿者系统",
		"version":     "2.0.0",
		"description": "帮助寻找走失人员的公益平台",
		"docs":        "/api/v1/docs",
		"health":      "/api/v1/health",
	})
}

// healthCheck 健康检查
func (r *Router) healthCheck(c *gin.Context) {
	response.Success(c, gin.H{
		"status": "UP",
		"time":   gin.H{},
	})
}

// detailedHealthCheck 详细健康检查
func (r *Router) detailedHealthCheck(c *gin.Context) {
	// 返回详细的系统状态
	response.Success(c, gin.H{
		"status":  "UP",
		"version": "2.0.0",
		"checks": gin.H{
			"api": gin.H{
				"status":  "UP",
				"message": "API服务正常运行",
			},
		},
	})
}
