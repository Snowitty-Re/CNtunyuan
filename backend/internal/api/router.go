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
	taskHandler      *TaskHandler
	workflowHandler  *WorkflowHandler
	jwtAuth          *auth.JWTAuth
}

// NewRouter 创建路由
func NewRouter(
	authService *service.UserService,
	userService *service.UserService,
	orgService *service.OrganizationService,
	mpService *service.MissingPersonService,
	dialectService *service.DialectService,
	taskService *service.TaskService,
	workflowService *service.WorkflowService,
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
		taskHandler:    NewTaskHandler(taskService),
		workflowHandler: NewWorkflowHandler(workflowService),
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

			// 任务管理
			tasks := authorized.Group("/tasks")
			{
				tasks.GET("", r.taskHandler.ListTasks)
				tasks.POST("", r.taskHandler.CreateTask)
				tasks.GET("/my", r.taskHandler.GetMyTasks)
				tasks.GET("/created", r.taskHandler.GetCreatedTasks)
				tasks.GET("/statistics", r.taskHandler.GetTaskStatistics)
				tasks.POST("/batch-assign", r.taskHandler.BatchAssign)
				tasks.POST("/auto-assign", r.taskHandler.AutoAssign)
				tasks.GET("/:id", r.taskHandler.GetTask)
				tasks.PUT("/:id", r.taskHandler.UpdateTask)
				tasks.DELETE("/:id", r.taskHandler.DeleteTask)
				tasks.POST("/:id/assign", r.taskHandler.AssignTask)
				tasks.POST("/:id/unassign", r.taskHandler.UnassignTask)
				tasks.POST("/:id/transfer", r.taskHandler.TransferTask)
				tasks.POST("/:id/complete", r.taskHandler.CompleteTask)
				tasks.POST("/:id/cancel", r.taskHandler.CancelTask)
				tasks.PUT("/:id/progress", r.taskHandler.UpdateProgress)
				tasks.GET("/:id/logs", r.taskHandler.GetTaskLogs)
				tasks.GET("/:id/comments", r.taskHandler.GetComments)
				tasks.POST("/:id/comments", r.taskHandler.AddComment)
			}

			// 工作流管理
			workflows := authorized.Group("/workflows")
			{
				workflows.GET("", r.workflowHandler.ListWorkflows)
				workflows.POST("", r.workflowHandler.CreateWorkflow)
				workflows.GET("/:id", r.workflowHandler.GetWorkflow)
				workflows.PUT("/:id", r.workflowHandler.UpdateWorkflow)
				workflows.DELETE("/:id", r.workflowHandler.DeleteWorkflow)
				workflows.POST("/:id/steps", r.workflowHandler.CreateStep)
				workflows.PUT("/:id/steps/:step_id", r.workflowHandler.UpdateStep)
				workflows.DELETE("/:id/steps/:step_id", r.workflowHandler.DeleteStep)
				workflows.POST("/:id/steps/reorder", r.workflowHandler.ReorderSteps)
			}

			// 工作流实例
			instances := authorized.Group("/workflow-instances")
			{
				instances.GET("", r.workflowHandler.ListInstances)
				instances.POST("", r.workflowHandler.StartInstance)
				instances.GET("/my", r.workflowHandler.GetMyInstances)
				instances.GET("/:id", r.workflowHandler.GetInstance)
				instances.POST("/:id/approve", r.workflowHandler.Approve)
				instances.POST("/:id/cancel", r.workflowHandler.CancelInstance)
				instances.GET("/:id/history", r.workflowHandler.GetInstanceHistory)
			}
		}
	}
}

// GetEngine 获取gin引擎
func (r *Router) GetEngine() *gin.Engine {
	return r.router
}
