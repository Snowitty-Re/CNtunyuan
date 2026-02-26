package api

import (
	"github.com/Snowitty-Re/CNtunyuan/internal/middleware"
	"github.com/Snowitty-Re/CNtunyuan/internal/service"
	"github.com/Snowitty-Re/CNtunyuan/pkg/auth"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/files"
)

// Router API路由
type Router struct {
	router           *gin.Engine
	authHandler      *AuthHandler
	userHandler      *UserHandler
	orgHandler       *OrgHandler
	mpHandler        *MissingPersonHandler
	dialectHandler   *DialectHandler
	jwtAuth          *auth.JWTAuth
}

// NewRouter 创建路由
func NewRouter(
	authService *service.UserService,
	userService *service.UserService,
	orgService *service.OrganizationService,
	mpService *service.MissingPersonService,
	dialectService *service.DialectService,
	wechatService *service.WeChatService,
	jwtAuth *auth.JWTAuth,
	redisClient *redis.Client,
) *Router {
	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())

	// 限流中间件
	r.Use(middleware.IPBasedRateLimit(redisClient))

	router := &Router{
		router:         r,
		authHandler:    NewAuthHandler(authService, wechatService, jwtAuth),
		userHandler:    NewUserHandler(userService),
		orgHandler:     NewOrgHandler(orgService),
		mpHandler:      NewMissingPersonHandler(mpService),
		dialectHandler: NewDialectHandler(dialectService),
		jwtAuth:        jwtAuth,
	}

	router.setupRoutes()
	return router
}

// setupRoutes 设置路由
func (r *Router) setupRoutes() {
	// 健康检查
	r.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Swagger 文档
	r.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1
	v1 := r.router.Group("/api/v1")
	{
		// 认证相关(公开)
		auth := v1.Group("/auth")
		{
			auth.POST("/wechat-login", r.authHandler.WeChatLogin)
			auth.POST("/admin-login", r.authHandler.AdminLogin)
			auth.POST("/refresh", r.authHandler.RefreshToken)
		}

		// 需要认证的路由
		authorized := v1.Group("")
		authorized.Use(middleware.JWTAuth(r.jwtAuth))
		{
			// 用户相关
			auth := authorized.Group("/auth")
			{
				auth.GET("/me", r.authHandler.GetCurrentUser)
				auth.POST("/logout", r.authHandler.Logout)
			}

			// 用户管理
			users := authorized.Group("/users")
			{
				users.GET("", r.userHandler.ListUsers)
				users.GET("/statistics", r.userHandler.GetUserStatistics)
				users.POST("/assign-to-org", r.userHandler.AssignToOrg)
				users.GET("/:id", r.userHandler.GetUser)
				users.PUT("/:id", r.userHandler.UpdateUser)
				users.DELETE("/:id", r.userHandler.DeleteUser)
			}

			// 组织管理
			orgs := authorized.Group("/organizations")
			{
				orgs.GET("", r.orgHandler.ListOrgs)
				orgs.GET("/tree", r.orgHandler.GetOrgTree)
				orgs.POST("", r.orgHandler.CreateOrg)
				orgs.GET("/:id", r.orgHandler.GetOrg)
				orgs.PUT("/:id", r.orgHandler.UpdateOrg)
				orgs.DELETE("/:id", r.orgHandler.DeleteOrg)
			}

			// 走失人员管理
			mps := authorized.Group("/missing-persons")
			{
				mps.GET("", r.mpHandler.List)
				mps.POST("", r.mpHandler.Create)
				mps.GET("/nearby", r.mpHandler.GetNearby)
				mps.GET("/statistics", r.mpHandler.GetStatistics)
				mps.GET("/:id", r.mpHandler.Get)
				mps.PUT("/:id/status", r.mpHandler.UpdateStatus)
				mps.GET("/:id/tracks", r.mpHandler.GetTracks)
				mps.POST("/:id/tracks", r.mpHandler.AddTrack)
			}

			// 方言管理
			dialects := authorized.Group("/dialects")
			{
				dialects.GET("", r.dialectHandler.List)
				dialects.POST("", r.dialectHandler.Create)
				dialects.GET("/nearby", r.dialectHandler.GetNearby)
				dialects.GET("/statistics", r.dialectHandler.GetStatistics)
				dialects.GET("/:id", r.dialectHandler.Get)
				dialects.PUT("/:id", r.dialectHandler.Update)
				dialects.DELETE("/:id", r.dialectHandler.Delete)
				dialects.POST("/:id/play", r.dialectHandler.Play)
				dialects.POST("/:id/like", r.dialectHandler.Like)
				dialects.POST("/:id/unlike", r.dialectHandler.Unlike)
			}
		}
	}
}

// GetEngine 获取gin引擎
func (r *Router) GetEngine() *gin.Engine {
	return r.router
}
