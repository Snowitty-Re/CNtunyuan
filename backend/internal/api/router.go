package api

import (
	"net/http"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/middleware"
	"github.com/Snowitty-Re/CNtunyuan/internal/model"
	"github.com/Snowitty-Re/CNtunyuan/internal/service"
	"github.com/Snowitty-Re/CNtunyuan/pkg/auth"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// Router API路由
type Router struct {
	router          *gin.Engine
	authHandler     *AuthHandler
	userHandler     *UserHandler
	orgHandler      *OrgHandler
	mpHandler       *MissingPersonHandler
	dialectHandler  *DialectHandler
	taskHandler     *TaskHandler
	workflowHandler *WorkflowHandler
	uploadHandler   *UploadHandler
	logHandler      *OperationLogHandler
	jwtAuth         *auth.JWTAuth
	storageConfig   *config.StorageConfig
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
	storageService *service.StorageService,
	storageConfig *config.StorageConfig,
	jwtAuth *auth.JWTAuth,
	redisClient *redis.Client,
	db *gorm.DB,
) *Router {
	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())

	// 限流中间件
	r.Use(middleware.IPBasedRateLimit(redisClient))

	// 操作日志中间件
	logMiddleware := middleware.NewAuditLogger(db)
	r.Use(logMiddleware.Logger())

	router := &Router{
		router:          r,
		authHandler:     NewAuthHandler(authService, wechatService, jwtAuth),
		userHandler:     NewUserHandler(userService),
		orgHandler:      NewOrgHandler(orgService),
		mpHandler:       NewMissingPersonHandler(mpService),
		dialectHandler:  NewDialectHandler(dialectService),
		taskHandler:     NewTaskHandler(taskService),
		workflowHandler: NewWorkflowHandler(workflowService),
		uploadHandler:   NewUploadHandler(storageService),
		logHandler:      NewOperationLogHandler(db),
		jwtAuth:         jwtAuth,
		storageConfig:   storageConfig,
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

	// 静态文件服务 - 本地存储的文件
	if r.storageConfig.Type == "local" {
		r.router.Static("/uploads", r.storageConfig.LocalPath)
	}

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
			// 当前用户相关（所有登录用户可用）
			auth := authorized.Group("/auth")
			{
				auth.GET("/me", r.authHandler.GetCurrentUser)
				auth.POST("/logout", r.authHandler.Logout)
			}

			// 文件上传（登录用户可用）
			uploads := authorized.Group("/upload")
			{
				uploads.POST("", r.uploadHandler.Upload)
				uploads.POST("/batch", r.uploadHandler.UploadMultiple)
				uploads.DELETE("", r.uploadHandler.DeleteFile)
			}

			// 用户管理（仅管理员）
			users := authorized.Group("/users")
			users.Use(middleware.RequireAdmin())
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
				// 公开查询（所有登录用户）
				orgs.GET("", r.orgHandler.ListOrgs)
				orgs.GET("/tree", r.orgHandler.GetOrgTree)
				orgs.GET("/:id", r.orgHandler.GetOrg)

				// 需要管理员权限
				adminOrgs := orgs.Group("")
				adminOrgs.Use(middleware.RequireAdmin())
				{
					adminOrgs.POST("", r.orgHandler.CreateOrg)
					adminOrgs.PUT("/:id", r.orgHandler.UpdateOrg)
					adminOrgs.DELETE("/:id", r.orgHandler.DeleteOrg)
				}
			}

			// 走失人员管理
			mps := authorized.Group("/missing-persons")
			{
				// 公开查询
				mps.GET("", r.mpHandler.List)
				mps.GET("/nearby", r.mpHandler.GetNearby)
				mps.GET("/statistics", r.mpHandler.GetStatistics)
				mps.GET("/:id", r.mpHandler.Get)
				mps.GET("/:id/tracks", r.mpHandler.GetTracks)

				// 需要管理者权限（创建、更新状态、添加轨迹）
				managerMps := mps.Group("")
				managerMps.Use(middleware.RequireManager())
				{
					managerMps.POST("", r.mpHandler.Create)
					managerMps.PUT("/:id/status", r.mpHandler.UpdateStatus)
					managerMps.POST("/:id/tracks", r.mpHandler.AddTrack)
				}
			}

			// 方言管理
			dialects := authorized.Group("/dialects")
			{
				// 公开查询
				dialects.GET("", r.dialectHandler.List)
				dialects.GET("/nearby", r.dialectHandler.GetNearby)
				dialects.GET("/statistics", r.dialectHandler.GetStatistics)
				dialects.GET("/:id", r.dialectHandler.Get)
				dialects.POST("/:id/play", r.dialectHandler.Play)
				dialects.POST("/:id/like", r.dialectHandler.Like)
				dialects.POST("/:id/unlike", r.dialectHandler.Unlike)

				// 需要登录用户权限（创建、更新、删除自己的）
				authDialects := dialects.Group("")
				authDialects.Use(middleware.RequireRole(model.RoleVolunteer))
				{
					authDialects.POST("", r.dialectHandler.Create)
					authDialects.PUT("/:id", r.dialectHandler.Update)
				}

				// 需要管理员权限（删除任意方言）
				adminDialects := dialects.Group("")
				adminDialects.Use(middleware.RequireAdmin())
				{
					adminDialects.DELETE("/:id", r.dialectHandler.Delete)
				}
			}

			// 任务管理
			tasks := authorized.Group("/tasks")
			{
				// 公开查询
				tasks.GET("", r.taskHandler.ListTasks)
				tasks.GET("/statistics", r.taskHandler.GetTaskStatistics)
				tasks.GET("/:id", r.taskHandler.GetTask)
				tasks.GET("/:id/logs", r.taskHandler.GetTaskLogs)
				tasks.GET("/:id/comments", r.taskHandler.GetComments)
				tasks.GET("/my", r.taskHandler.GetMyTasks)
				tasks.GET("/created", r.taskHandler.GetCreatedTasks)

				// 需要志愿者权限（创建任务、添加评论、更新进度）
				volunteerTasks := tasks.Group("")
				volunteerTasks.Use(middleware.RequireRole(model.RoleVolunteer))
				{
					volunteerTasks.POST("", r.taskHandler.CreateTask)
					volunteerTasks.POST("/:id/comments", r.taskHandler.AddComment)
					volunteerTasks.PUT("/:id/progress", r.taskHandler.UpdateProgress)
					volunteerTasks.POST("/:id/complete", r.taskHandler.CompleteTask)
				}

				// 需要管理者权限（分配、转派、取消）
				managerTasks := tasks.Group("")
				managerTasks.Use(middleware.RequireManager())
				{
					managerTasks.PUT("/:id", r.taskHandler.UpdateTask)
					managerTasks.DELETE("/:id", r.taskHandler.DeleteTask)
					managerTasks.POST("/:id/assign", r.taskHandler.AssignTask)
					managerTasks.POST("/:id/unassign", r.taskHandler.UnassignTask)
					managerTasks.POST("/:id/transfer", r.taskHandler.TransferTask)
					managerTasks.POST("/:id/cancel", r.taskHandler.CancelTask)
				}

				// 需要管理员权限（批量分配、自动分配）
				adminTasks := tasks.Group("")
				adminTasks.Use(middleware.RequireAdmin())
				{
					adminTasks.POST("/batch-assign", r.taskHandler.BatchAssign)
					adminTasks.POST("/auto-assign", r.taskHandler.AutoAssign)
				}
			}

			// 工作流管理
			workflows := authorized.Group("/workflows")
			{
				// 公开查询
				workflows.GET("", r.workflowHandler.ListWorkflows)
				workflows.GET("/:id", r.workflowHandler.GetWorkflow)

				// 需要管理员权限（管理工作流定义）
				adminWorkflows := workflows.Group("")
				adminWorkflows.Use(middleware.RequireAdmin())
				{
					adminWorkflows.POST("", r.workflowHandler.CreateWorkflow)
					adminWorkflows.PUT("/:id", r.workflowHandler.UpdateWorkflow)
					adminWorkflows.DELETE("/:id", r.workflowHandler.DeleteWorkflow)
					adminWorkflows.POST("/:id/steps", r.workflowHandler.CreateStep)
					adminWorkflows.PUT("/:id/steps/:step_id", r.workflowHandler.UpdateStep)
					adminWorkflows.DELETE("/:id/steps/:step_id", r.workflowHandler.DeleteStep)
					adminWorkflows.POST("/:id/steps/reorder", r.workflowHandler.ReorderSteps)
				}
			}

			// 工作流实例
			instances := authorized.Group("/workflow-instances")
			{
				// 公开查询
				instances.GET("", r.workflowHandler.ListInstances)
				instances.GET("/my", r.workflowHandler.GetMyInstances)
				instances.GET("/:id", r.workflowHandler.GetInstance)
				instances.GET("/:id/history", r.workflowHandler.GetInstanceHistory)

				// 需要志愿者权限
				volunteerInstances := instances.Group("")
				volunteerInstances.Use(middleware.RequireRole(model.RoleVolunteer))
				{
					volunteerInstances.POST("", r.workflowHandler.StartInstance)
				}

				// 需要管理者权限（审批、取消）
				managerInstances := instances.Group("")
				managerInstances.Use(middleware.RequireManager())
				{
					managerInstances.POST("/:id/approve", r.workflowHandler.Approve)
					managerInstances.POST("/:id/cancel", r.workflowHandler.CancelInstance)
				}
			}

			// 操作日志（仅超级管理员）
			logs := authorized.Group("/operation-logs")
			logs.Use(middleware.RequireSuperAdmin())
			{
				logs.GET("", r.logHandler.List)
				logs.DELETE("/cleanup", r.logHandler.Cleanup)
				logs.GET("/stats/user/:user_id", r.logHandler.GetUserStats)
				logs.GET("/stats/summary", r.logHandler.GetSummary)
			}
		}
	}

	// 404处理
	r.router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "请求的资源不存在",
		})
	})
}

// GetEngine 获取gin引擎
func (r *Router) GetEngine() *gin.Engine {
	return r.router
}
