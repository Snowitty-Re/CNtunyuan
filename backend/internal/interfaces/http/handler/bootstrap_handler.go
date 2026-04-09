package handler

import (
	"strings"
	"sync"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/database"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const initDefaultOrgID = "00000000-0000-0000-0000-000000000000"

type BootstrapHandler struct {
	userRepo       repository.UserRepository
	healthService  *service.HealthService
	db             *gorm.DB
	initializeLock sync.Mutex
}

type BootstrapValidateDatabaseRequest struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"ssl_mode"`
	Timezone string `json:"timezone"`
}

type BootstrapInitializeRequest struct {
	Database   *BootstrapValidateDatabaseRequest `json:"database"`
	Site       *BootstrapInitializeSiteRequest   `json:"site"`
	SuperAdmin *BootstrapInitializeAdminRequest  `json:"super_admin"`
}

type BootstrapInitializeSiteRequest struct {
	Domain            string `json:"domain"`
	CORSOrigins       string `json:"cors_origins"`
	DefaultOrgName    string `json:"default_org_name"`
	DefaultOrgCode    string `json:"default_org_code"`
	EnableRegister    *bool  `json:"enable_register"`
	EnableWechatLogin *bool  `json:"enable_wechat_login"`
	EnableWechatWeb   *bool  `json:"enable_wechat_login_web"`
	EnableWechatMini  *bool  `json:"enable_wechat_login_mini_program"`
	EnableSMSLogin    *bool  `json:"enable_sms_login"`
}

type BootstrapInitializeAdminRequest struct {
	Nickname string `json:"nickname"`
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	Email    string `json:"email"`
}

func NewBootstrapHandler(userRepo repository.UserRepository, healthService *service.HealthService, db *gorm.DB) *BootstrapHandler {
	return &BootstrapHandler{
		userRepo:      userRepo,
		healthService: healthService,
		db:            db,
	}
}

func (h *BootstrapHandler) RegisterRoutes(router *gin.RouterGroup) {
	bootstrap := router.Group("/bootstrap")
	{
		bootstrap.GET("/status", h.GetStatus)
		bootstrap.POST("/validate-db", h.ValidateDatabase)
		bootstrap.POST("/initialize", h.Initialize)
	}
}

// GetStatus 获取初始化状态
// @Summary      获取初始化状态
// @Description  检查系统是否已初始化，并返回数据库连通、配置可写等检测结果
// @Tags         系统初始化
// @Produce      json
// @Success      200  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /bootstrap/status [get]
func (h *BootstrapHandler) GetStatus(c *gin.Context) {
	cfg := config.GetConfig()
	if cfg == nil {
		response.InternalServerError(c, "config not loaded")
		return
	}

	superAdminCount, err := h.userRepo.CountByRole(c.Request.Context(), entity.RoleSuperAdmin)
	if err != nil {
		response.InternalServerError(c, "failed to load user initialization status")
		return
	}

	healthStatus := "unknown"
	dbConnected := false
	if h.healthService != nil {
		health := h.healthService.CheckHealth(c.Request.Context())
		healthStatus = strings.ToLower(string(health.Status))
		if dbCheck, ok := health.Checks["database"]; ok {
			dbConnected = strings.EqualFold(string(dbCheck.Status), "UP")
		}
	}

	response.Success(c, gin.H{
		"initialized":       superAdminCount > 0,
		"super_admin_count": superAdminCount,
		"checks": gin.H{
			"database_connected": dbConnected,
			"settings_storage":   "database_overrides",
			"health_status":      healthStatus,
		},
		"database": gin.H{
			"type":     string(cfg.Database.Type),
			"host":     cfg.Database.Host,
			"port":     cfg.Database.Port,
			"user":     cfg.Database.User,
			"database": cfg.Database.Database,
			"ssl_mode": cfg.Database.SSLMode,
			"timezone": cfg.Database.Timezone,
		},
		"site": gin.H{
			"domain":                           cfg.Server.Domain,
			"cors_origins":                     cfg.Server.CORSOrigins,
			"default_org_name":                 cfg.System.DefaultOrgName,
			"default_org_code":                 cfg.System.DefaultOrgCode,
			"enable_register":                  cfg.System.EnableRegister,
			"enable_wechat_login":              cfg.System.EnableWechatLogin,
			"enable_wechat_login_web":          cfg.System.EnableWechatLoginWeb,
			"enable_wechat_login_mini_program": cfg.System.EnableWechatLoginMiniProgram,
			"enable_sms_login":                 cfg.System.EnableSMSLogin,
		},
		"config_path": "config.yaml(database only) + ty_system_settings(runtime overrides)",
		"server_time": time.Now().Format(time.RFC3339),
	})
}

// ValidateDatabase 校验数据库连接
// @Summary      校验数据库连接
// @Description  用提交的数据库配置进行连通性测试（不会落盘）
// @Tags         系统初始化
// @Accept       json
// @Produce      json
// @Param        request  body      BootstrapValidateDatabaseRequest  true  "数据库连接配置"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /bootstrap/validate-db [post]
func (h *BootstrapHandler) ValidateDatabase(c *gin.Context) {
	var req BootstrapValidateDatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	cfg := config.GetConfig()
	if cfg == nil {
		response.InternalServerError(c, "config not loaded")
		return
	}

	dbCfg := buildDatabaseConfigFromRequest(&cfg.Database, &req)
	if !dbCfg.IsValid() {
		response.BadRequest(c, "database config is invalid")
		return
	}

	begin := time.Now()
	if err := database.TestConnection(dbCfg); err != nil {
		response.BadRequest(c, "database validation failed: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"ok":         true,
		"latency_ms": time.Since(begin).Milliseconds(),
		"database": gin.H{
			"type":     string(dbCfg.Type),
			"host":     dbCfg.Host,
			"port":     dbCfg.Port,
			"user":     dbCfg.User,
			"database": dbCfg.Database,
			"ssl_mode": dbCfg.SSLMode,
			"timezone": dbCfg.Timezone,
		},
	})
}

// Initialize 执行首次初始化
// @Summary      执行首次初始化
// @Description  首次启动时写入配置并创建超级管理员；若已有 super_admin 则拒绝重复初始化
// @Tags         系统初始化
// @Accept       json
// @Produce      json
// @Param        request  body      BootstrapInitializeRequest  true  "初始化请求（数据库、站点、超级管理员）"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      403      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /bootstrap/initialize [post]
func (h *BootstrapHandler) Initialize(c *gin.Context) {
	h.initializeLock.Lock()
	defer h.initializeLock.Unlock()

	superAdminCount, err := h.userRepo.CountByRole(c.Request.Context(), entity.RoleSuperAdmin)
	if err != nil {
		response.InternalServerError(c, "failed to load user initialization status")
		return
	}
	if superAdminCount > 0 {
		response.Forbidden(c, "system already initialized")
		return
	}

	var req BootstrapInitializeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	if req.SuperAdmin == nil {
		response.BadRequest(c, "super_admin is required")
		return
	}

	cfg := config.GetConfig()
	if cfg == nil {
		response.InternalServerError(c, "config not loaded")
		return
	}

	working := *cfg
	if req.Database != nil {
		working.Database = *buildDatabaseConfigFromRequest(&cfg.Database, req.Database)
		if !working.Database.IsValid() {
			response.BadRequest(c, "database config is invalid")
			return
		}
		if err := database.TestConnection(&working.Database); err != nil {
			response.BadRequest(c, "database validation failed: "+err.Error())
			return
		}
	}
	if req.Site != nil {
		applySiteConfig(&working, req.Site)
	}

	adminPhone := strings.TrimSpace(req.SuperAdmin.Phone)
	existsPhone, err := h.userRepo.ExistsPhone(c.Request.Context(), adminPhone)
	if err != nil {
		response.InternalServerError(c, "failed to check super admin phone")
		return
	}
	if existsPhone {
		response.BadRequest(c, "phone already exists")
		return
	}

	superAdmin := &entity.User{
		BaseEntity: entity.BaseEntity{ID: uuid.New().String()},
		Nickname:   strings.TrimSpace(req.SuperAdmin.Nickname),
		Phone:      adminPhone,
		Email:      strings.TrimSpace(req.SuperAdmin.Email),
		Role:       entity.RoleSuperAdmin,
		Status:     entity.UserStatusActive,
		OrgID:      initDefaultOrgID,
	}
	if superAdmin.Nickname == "" {
		superAdmin.Nickname = "超级管理员"
	}
	if err := superAdmin.SetPassword(strings.TrimSpace(req.SuperAdmin.Password)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := superAdmin.Validate(); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Site != nil {
		siteFlat := buildSiteOverrideMap(req.Site)
		if len(siteFlat) > 0 {
			nextCfg, _, err := config.SaveRuntimeOverrides(c.Request.Context(), h.db, cfg, siteFlat)
			if err != nil {
				response.InternalServerError(c, "failed to save runtime settings: "+err.Error())
				return
			}
			config.SetConfig(nextCfg)
			working = *nextCfg
		}
	}

	if err := h.userRepo.Create(c.Request.Context(), superAdmin); err != nil {
		response.InternalServerError(c, "failed to create super admin: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "初始化完成", gin.H{
		"initialized":  true,
		"config_path":  "config.yaml(database only) + ty_system_settings(runtime overrides)",
		"super_admin":  gin.H{"phone": superAdmin.Phone, "nickname": superAdmin.Nickname},
		"server_time":  time.Now().Format(time.RFC3339),
		"next_actions": []string{"请使用超级管理员账号登录 Web 端", "建议重启后端以确保所有配置完全生效"},
	})
}

func buildDatabaseConfigFromRequest(current *config.DatabaseConfig, req *BootstrapValidateDatabaseRequest) *config.DatabaseConfig {
	next := *current
	if strings.TrimSpace(req.Type) != "" {
		next.Type = config.DatabaseType(strings.TrimSpace(req.Type))
	}
	if strings.TrimSpace(req.Host) != "" {
		next.Host = strings.TrimSpace(req.Host)
	}
	if req.Port > 0 {
		next.Port = req.Port
	}
	if strings.TrimSpace(req.User) != "" {
		next.User = strings.TrimSpace(req.User)
	}
	if req.Password != "" {
		next.Password = req.Password
	}
	if strings.TrimSpace(req.Database) != "" {
		next.Database = strings.TrimSpace(req.Database)
	}
	if strings.TrimSpace(req.SSLMode) != "" {
		next.SSLMode = strings.TrimSpace(req.SSLMode)
	}
	if strings.TrimSpace(req.Timezone) != "" {
		next.Timezone = strings.TrimSpace(req.Timezone)
	}
	return &next
}

func applySiteConfig(cfg *config.Config, site *BootstrapInitializeSiteRequest) {
	if strings.TrimSpace(site.Domain) != "" {
		cfg.Server.Domain = strings.TrimSpace(site.Domain)
	}
	if strings.TrimSpace(site.CORSOrigins) != "" {
		cfg.Server.CORSOrigins = strings.TrimSpace(site.CORSOrigins)
	}
	if strings.TrimSpace(site.DefaultOrgName) != "" {
		cfg.System.DefaultOrgName = strings.TrimSpace(site.DefaultOrgName)
	}
	if strings.TrimSpace(site.DefaultOrgCode) != "" {
		cfg.System.DefaultOrgCode = strings.TrimSpace(site.DefaultOrgCode)
	}
	if site.EnableRegister != nil {
		cfg.System.EnableRegister = *site.EnableRegister
	}
	if site.EnableWechatLogin != nil {
		cfg.System.EnableWechatLogin = *site.EnableWechatLogin
	}
	if site.EnableWechatWeb != nil {
		cfg.System.EnableWechatLoginWeb = *site.EnableWechatWeb
	}
	if site.EnableWechatMini != nil {
		cfg.System.EnableWechatLoginMiniProgram = *site.EnableWechatMini
	}
	if site.EnableSMSLogin != nil {
		cfg.System.EnableSMSLogin = *site.EnableSMSLogin
	}
}

func buildSiteOverrideMap(site *BootstrapInitializeSiteRequest) map[string]interface{} {
	if site == nil {
		return nil
	}
	overrides := map[string]interface{}{}
	if strings.TrimSpace(site.Domain) != "" {
		overrides["server.domain"] = strings.TrimSpace(site.Domain)
	}
	if strings.TrimSpace(site.CORSOrigins) != "" {
		overrides["server.cors_origins"] = strings.TrimSpace(site.CORSOrigins)
	}
	if strings.TrimSpace(site.DefaultOrgName) != "" {
		overrides["system.default_org_name"] = strings.TrimSpace(site.DefaultOrgName)
	}
	if strings.TrimSpace(site.DefaultOrgCode) != "" {
		overrides["system.default_org_code"] = strings.TrimSpace(site.DefaultOrgCode)
	}
	if site.EnableRegister != nil {
		overrides["system.enable_register"] = *site.EnableRegister
	}
	if site.EnableWechatLogin != nil {
		overrides["system.enable_wechat_login"] = *site.EnableWechatLogin
	}
	if site.EnableWechatWeb != nil {
		overrides["system.enable_wechat_login_web"] = *site.EnableWechatWeb
	}
	if site.EnableWechatMini != nil {
		overrides["system.enable_wechat_login_mini_program"] = *site.EnableWechatMini
	}
	if site.EnableSMSLogin != nil {
		overrides["system.enable_sms_login"] = *site.EnableSMSLogin
	}
	return overrides
}
