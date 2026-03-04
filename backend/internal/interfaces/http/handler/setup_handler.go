package handler

import (
	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/database"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SetupHandler 系统初始化处理器
type SetupHandler struct {
	configPath string
}

// NewSetupHandler 创建初始化处理器
func NewSetupHandler(configPath string) *SetupHandler {
	return &SetupHandler{configPath: configPath}
}

// RegisterRoutes 注册路由
func (h *SetupHandler) RegisterRoutes(router *gin.RouterGroup) {
	setup := router.Group("/setup")
	{
		setup.GET("/status", h.GetStatus)
		setup.POST("/test-db", h.TestDatabase)
		setup.POST("/initialize", h.Initialize)
	}
}

// SetupStatus 初始化状态
type SetupStatus struct {
	Initialized bool `json:"initialized"`
}

// GetStatus 获取初始化状态
func (h *SetupHandler) GetStatus(c *gin.Context) {
	// 检查配置文件是否存在且包含有效的数据库配置
	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		response.Success(c, SetupStatus{Initialized: false})
		return
	}

	// 检查数据库配置是否有效
	if !cfg.Database.IsValid() {
		response.Success(c, SetupStatus{Initialized: false})
		return
	}

	// 尝试连接数据库
	db, err := database.NewDatabase(&cfg.Database)
	if err != nil {
		response.Success(c, SetupStatus{Initialized: false})
		return
	}

	// 检查是否已存在管理员用户
	var count int64
	if err := db.Model(&entity.User{}).Where("role = ?", entity.RoleSuperAdmin).Count(&count).Error; err != nil {
		response.Success(c, SetupStatus{Initialized: false})
		return
	}

	response.Success(c, SetupStatus{Initialized: count > 0})
}

// TestDBRequest 测试数据库请求
type TestDBRequest struct {
	Type     string `json:"type" binding:"required,oneof=postgres mysql"`
	Host     string `json:"host" binding:"required"`
	Port     int    `json:"port" binding:"required,min=1,max=65535"`
	User     string `json:"user" binding:"required"`
	Password string `json:"password" binding:"required"`
	Database string `json:"database" binding:"required"`
	SSLMode  string `json:"ssl_mode"` // PostgreSQL
	Charset  string `json:"charset"`  // MySQL
}

// TestDatabase 测试数据库连接
func (h *SetupHandler) TestDatabase(c *gin.Context) {
	var req TestDBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("TestDatabase bind error", logger.Err(err))
		response.BadRequest(c, err.Error())
		return
	}

	// Debug log
	logger.Info("TestDatabase request",
		logger.String("type", req.Type),
		logger.String("host", req.Host),
		logger.Int("port", req.Port),
		logger.String("database", req.Database),
		logger.String("ssl_mode_received", req.SSLMode),
	)

	// 设置 SSLMode 默认值 - 强制使用 disable 避免 TLS 连接问题
	sslMode := req.SSLMode
	if sslMode == "" || sslMode == "require" {
		sslMode = "disable"
	}

	cfg := config.DatabaseConfig{
		Type:     config.DatabaseType(req.Type),
		Host:     req.Host,
		Port:     req.Port,
		User:     req.User,
		Password: req.Password,
		Database: req.Database,
		SSLMode:  sslMode,
		Charset:  req.Charset,
	}

	if cfg.Charset == "" && cfg.Type == config.DatabaseTypeMySQL {
		cfg.Charset = "utf8mb4"
	}

	// 测试连接
	if err := database.TestConnection(&cfg); err != nil {
		logger.Error("Database connection test failed", logger.Err(err))
		response.BadRequest(c, "数据库连接失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{"message": "数据库连接成功"})
}

// InitializeRequest 初始化请求
type InitializeRequest struct {
	// 数据库配置
	DBType     string `json:"db_type" binding:"required,oneof=postgres mysql"`
	DBHost     string `json:"db_host" binding:"required"`
	DBPort     int    `json:"db_port" binding:"required,min=1,max=65535"`
	DBUser     string `json:"db_user" binding:"required"`
	DBPassword string `json:"db_password" binding:"required"`
	DBName     string `json:"db_name" binding:"required"`
	DBSSLMode  string `json:"db_ssl_mode"` // PostgreSQL
	DBCharset  string `json:"db_charset"`  // MySQL

	// 管理员配置
	AdminPhone    string `json:"admin_phone" binding:"required"`
	AdminPassword string `json:"admin_password" binding:"required,min=6"`
	AdminNickname string `json:"admin_nickname" binding:"required"`
}

// Initialize 初始化系统
func (h *SetupHandler) Initialize(c *gin.Context) {
	var req InitializeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 设置 SSLMode 默认值
	sslMode := req.DBSSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	// 1. 构建数据库配置
	cfg := config.DatabaseConfig{
		Type:            config.DatabaseType(req.DBType),
		Host:            req.DBHost,
		Port:            req.DBPort,
		User:            req.DBUser,
		Password:        req.DBPassword,
		Database:        req.DBName,
		SSLMode:         sslMode,
		Charset:         req.DBCharset,
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: 3600,
	}

	if cfg.Charset == "" && cfg.Type == config.DatabaseTypeMySQL {
		cfg.Charset = "utf8mb4"
	}

	// 2. 创建数据库（如果不存在）
	if err := database.CreateDatabaseIfNotExists(&cfg); err != nil {
		logger.Error("Failed to create database", logger.Err(err))
		response.InternalServerError(c, "创建数据库失败: "+err.Error())
		return
	}

	// 3. 连接数据库
	db, err := database.NewDatabase(&cfg)
	if err != nil {
		logger.Error("Failed to connect database", logger.Err(err))
		response.InternalServerError(c, "连接数据库失败: "+err.Error())
		return
	}

	// 4. 执行数据库迁移
	if err := database.AutoMigrate(db); err != nil {
		logger.Error("Failed to migrate database", logger.Err(err))
		response.InternalServerError(c, "数据库迁移失败: "+err.Error())
		return
	}

	// 5. 创建根组织
	rootOrg := &entity.Organization{
		BaseEntity: entity.BaseEntity{ID: uuid.New().String()},
		Name:       "团圆寻亲志愿者总会",
		Code:       "ROOT",
		Type:       entity.OrgTypeRoot,
		Level:      1,
		Status:     entity.OrgStatusActive,
	}
	if err := db.Create(rootOrg).Error; err != nil {
		logger.Error("Failed to create root organization", logger.Err(err))
		response.InternalServerError(c, "创建根组织失败")
		return
	}

	// 6. 创建超级管理员
	admin := &entity.User{
		BaseEntity: entity.BaseEntity{ID: uuid.New().String()},
		Nickname:   req.AdminNickname,
		Phone:      req.AdminPhone,
		OrgID:      rootOrg.ID,
		Role:       entity.RoleSuperAdmin,
		Status:     entity.UserStatusActive,
	}
	if err := admin.SetPassword(req.AdminPassword); err != nil {
		logger.Error("Failed to set admin password", logger.Err(err))
		response.InternalServerError(c, "设置管理员密码失败")
		return
	}

	if err := db.Create(admin).Error; err != nil {
		logger.Error("Failed to create admin user", logger.Err(err))
		response.InternalServerError(c, "创建管理员失败")
		return
	}

	// 7. TODO: 保存配置到配置文件
	// 这里需要将配置写入 config.yaml 文件
	// 暂时跳过，实际部署时需要实现

	logger.Info("System initialized successfully",
		logger.String("admin_phone", req.AdminPhone),
		logger.String("admin_nickname", req.AdminNickname),
		logger.String("db_type", req.DBType),
	)

	response.Success(c, gin.H{
		"message": "系统初始化成功",
		"admin": dto.UserResponse{
			ID:        admin.ID,
			Nickname:  admin.Nickname,
			Phone:     admin.Phone,
			Role:      string(admin.Role),
			OrgID:     admin.OrgID,
			CreatedAt: admin.CreatedAt,
		},
	})
}

// CheckAndInitialize 检查是否需要初始化
func CheckAndInitialize(db *gorm.DB) error {
	// 检查是否已存在管理员用户
	var count int64
	if err := db.Model(&entity.User{}).Where("role = ?", entity.RoleSuperAdmin).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		logger.Info("No super admin found, system needs initialization")
	}

	return nil
}
