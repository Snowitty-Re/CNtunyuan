package router

import (
	"net/http"
	"strings"
	"time"

	_ "github.com/Snowitty-Re/CNtunyuan/docs"
	"github.com/Snowitty-Re/CNtunyuan/internal/application/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	domainService "github.com/Snowitty-Re/CNtunyuan/internal/domain/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/handler"
	"github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/middleware"
	pkgmiddleware "github.com/Snowitty-Re/CNtunyuan/pkg/middleware"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Router 路由管理器
type Router struct {
	engine               *gin.Engine
	authHandler          *handler.AuthHandler
	userHandler          *handler.UserHandler
	organizationHandler  *handler.OrganizationHandler
	missingPersonHandler *handler.MissingPersonHandler
	dialectHandler       *handler.DialectHandler
	taskHandler          *handler.TaskHandler
	uploadHandler        *handler.UploadHandler
	dashboardHandler     *handler.DashboardHandler
	auditHandler         *handler.AuditHandler
	systemConfigHandler  *handler.SystemConfigHandler
	authMiddleware       *middleware.AuthMiddleware
	healthService        *service.HealthService
	cache                domainService.Cache
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
	systemConfigHandler *handler.SystemConfigHandler,
	authMiddleware *middleware.AuthMiddleware,
	auditMiddleware *middleware.AuditMiddleware,
	healthService *service.HealthService,
	cache domainService.Cache,
) *Router {
	engine := gin.New()

	// 全局中间件（按执行顺序排列）
	// 1. 恢复中间件（捕获 panic）
	engine.Use(pkgmiddleware.RecoveryMiddleware())

	// 2. 追踪 ID 中间件
	engine.Use(pkgmiddleware.TraceIDMiddleware())

	// 3. 安全响应头中间件
	engine.Use(pkgmiddleware.SecurityHeadersMiddleware())

	// 4. CORS 中间件（从配置读取允许的源）
	var corsOrigins []string
	cfg := config.GetConfig()
	if cfg != nil && cfg.Server.CORSOrigins != "" {
		corsOrigins = strings.Split(cfg.Server.CORSOrigins, ",")
		for i := range corsOrigins {
			corsOrigins[i] = strings.TrimSpace(corsOrigins[i])
		}
	}
	engine.Use(pkgmiddleware.CORSMiddleware(corsOrigins...))

	// 5. 请求大小限制（50MB）
	engine.Use(pkgmiddleware.RequestSizeMiddleware(50 * 1024 * 1024))

	// 6. 限流中间件
	// 生产优先使用 Redis 分布式限流；无缓存时回退到进程内限流。
	if cache != nil {
		engine.Use(middleware.DistributedRateLimitMiddleware(cache, 100, time.Second))
	} else {
		engine.Use(pkgmiddleware.RateLimitMiddleware(100, 200))
	}

	// 7. 审计日志中间件（记录所有请求）
	if auditMiddleware != nil {
		engine.Use(auditMiddleware.AutoAudit())
	}

	// 8. 结构化日志中间件
	engine.Use(pkgmiddleware.LoggingMiddleware())

	// 9. 统一错误处理中间件
	engine.Use(pkgmiddleware.ErrorHandlerMiddleware())

	return &Router{
		engine:               engine,
		authHandler:          authHandler,
		userHandler:          userHandler,
		organizationHandler:  organizationHandler,
		missingPersonHandler: missingPersonHandler,
		dialectHandler:       dialectHandler,
		taskHandler:          taskHandler,
		uploadHandler:        uploadHandler,
		dashboardHandler:     dashboardHandler,
		auditHandler:         auditHandler,
		systemConfigHandler:  systemConfigHandler,
		authMiddleware:       authMiddleware,
		healthService:        healthService,
		cache:                cache,
	}
}

// Setup 设置路由
func (r *Router) Setup() {
	// 本地存储静态文件路由（仅 local 模式）
	if cfg := config.GetConfig(); cfg != nil && cfg.Storage.Type == "local" && cfg.Storage.LocalPath != "" {
		r.engine.Static("/uploads", cfg.Storage.LocalPath)
	}

	// Swagger UI - 公开访问
	r.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 路由组
	api := r.engine.Group("/api/v1")

	// 健康检查（不需要认证）
	api.GET("/health", r.healthCheck)
	api.GET("/health/detailed", r.detailedHealthCheck)

	// Prometheus 指标端点（仅管理员可访问，避免泄露运维数据）
	api.GET("/metrics", r.authMiddleware.Required(), middleware.RequireAdmin(), r.metrics)

	// Swagger JSON/YAML 文档端点
	api.GET("/docs", r.redirectDocs)

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
	if r.systemConfigHandler != nil {
		r.systemConfigHandler.RegisterRoutes(api, r.authMiddleware)
	}

	// 注册审计日志路由（如果配置了）
	if r.auditHandler != nil {
		r.auditHandler.RegisterRoutes(api, r.authMiddleware)
	}

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
// @Summary      欢迎信息
// @Description  获取系统基础信息与文档入口
// @Tags         健康检查
// @Produce      json
// @Success      200  {object}  response.Response
// @Router       / [get]
func (r *Router) welcome(c *gin.Context) {
	response.Success(c, gin.H{
		"name":        "助力团圆志愿者系统",
		"version":     "2.0.0",
		"description": "帮助走失人员寻找亲属、助力团圆的公益平台",
		"docs":        "/api/v1/docs",
		"health":      "/api/v1/health",
	})
}

// redirectDocs 跳转到 Swagger UI
// @Summary      API 文档入口
// @Description  跳转到 Swagger UI 页面
// @Tags         健康检查
// @Produce      json
// @Success      301  {string}  string  "Moved Permanently"
// @Router       /docs [get]
func (r *Router) redirectDocs(c *gin.Context) {
	c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
}

// healthCheck 健康检查
// @Summary      健康检查
// @Description  检查服务是否可用
// @Tags         健康检查
// @Produce      json
// @Success      200  {object}  response.Response
// @Failure      503  {object}  response.Response
// @Router       /health [get]
func (r *Router) healthCheck(c *gin.Context) {
	if r.healthService == nil {
		response.Success(c, gin.H{
			"status": "UP",
			"time":   gin.H{},
		})
		return
	}

	result := r.healthService.CheckHealth(c.Request.Context())
	if result.Status == service.HealthStatusUP {
		response.Success(c, result)
	} else {
		response.ErrorCodeWithMessage(c, http.StatusServiceUnavailable, "service unavailable")
	}
}

// detailedHealthCheck 详细健康检查
// @Summary      详细健康检查
// @Description  返回服务及依赖组件健康状态详情
// @Tags         健康检查
// @Produce      json
// @Success      200  {object}  response.Response
// @Router       /health/detailed [get]
func (r *Router) detailedHealthCheck(c *gin.Context) {
	if r.healthService == nil {
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
		return
	}

	result := r.healthService.CheckHealth(c.Request.Context())
	response.Success(c, result)
}

// metrics Prometheus 指标
// @Summary      Prometheus 指标
// @Description  获取系统运行指标（管理员权限）
// @Tags         健康检查
// @Produce      plain
// @Success      200  {string}  string  "metrics"
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /metrics [get]
// @Security     Bearer
func (r *Router) metrics(c *gin.Context) {
	gin.WrapH(promhttp.Handler())(c)
}
