package handler

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/database"
	infraRepo "github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/repository"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const initDefaultOrgID = "00000000-0000-0000-0000-000000000000"
const initSeedSuperAdminID = "00000000-0000-0000-0000-000000000001"

type BootstrapHandler struct {
	userRepo          repository.UserRepository
	healthService     *service.HealthService
	db                *gorm.DB
	managedConfigPath string
	initializeLock    sync.Mutex
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
		userRepo:          userRepo,
		healthService:     healthService,
		db:                db,
		managedConfigPath: "",
	}
}

func (h *BootstrapHandler) WithManagedConfigPath(path string) *BootstrapHandler {
	h.managedConfigPath = path
	return h
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
// @Description  检查系统是否已初始化，并返回数据库连通、配置可写、Schema 就绪等检测结果
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

	managed, _ := config.LoadManagedStartupConfig(h.getManagedConfigPath())
	configSource, startupConfigPath := config.GetStartupMetadata()
	healthStatus := "unknown"
	dbConnected := false
	schemaReady := false
	superAdminCount := int64(0)

	if h.healthService != nil {
		health := h.healthService.CheckHealth(c.Request.Context())
		healthStatus = strings.ToLower(string(health.Status))
		if dbCheck, ok := health.Checks["database"]; ok {
			dbConnected = strings.EqualFold(string(dbCheck.Status), "UP")
		}
	}

	if h.db != nil {
		schemaReady = requiredBootstrapTablesReady(h.db)
		superAdminCount = h.countSuperAdmins(c.Request.Context(), h.db)
	}

	initialized := false
	if managed != nil {
		initialized = managed.Initialized
	} else if superAdminCount > 0 {
		initialized = true
	}

	response.Success(c, gin.H{
		"initialized":       initialized,
		"startup_mode":      h.startupMode(dbConnected),
		"config_source":     configSource,
		"super_admin_count": superAdminCount,
		"checks": gin.H{
			"database_connected": dbConnected,
			"schema_ready":       schemaReady,
			"settings_storage":   settingsStorageLabel(configSource),
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
			"logo_url":                         cfg.System.LogoURL,
			"default_org_code":                 cfg.System.DefaultOrgCode,
			"enable_register":                  cfg.System.EnableRegister,
			"enable_wechat_login":              cfg.System.EnableWechatLogin,
			"enable_wechat_login_web":          cfg.System.EnableWechatLoginWeb,
			"enable_wechat_login_mini_program": cfg.System.EnableWechatLoginMiniProgram,
			"enable_sms_login":                 cfg.System.EnableSMSLogin,
		},
		"config_path": gin.H{
			"startup": startupPathLabel(configSource, startupConfigPath, h.getManagedConfigPath()),
			"runtime": "ty_system_settings",
		},
		"server_time": time.Now().Format(time.RFC3339),
	})
}

// ValidateDatabase 校验数据库连接
// @Summary      校验数据库连接
// @Description  用提交的数据库配置进行连通性测试，并返回 schema 就绪状态（不会落盘）
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
	testDB, err := database.NewDatabase(dbCfg)
	if err != nil {
		response.BadRequest(c, "database validation failed: "+err.Error())
		return
	}
	sqlDB, _ := testDB.DB()
	if sqlDB != nil {
		defer sqlDB.Close()
	}

	response.Success(c, gin.H{
		"ok":           true,
		"latency_ms":   time.Since(begin).Milliseconds(),
		"schema_ready": requiredBootstrapTablesReady(testDB),
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
// @Description  首次启动时验证数据库、自动建表、写入启动配置与站点设置，并创建超级管理员
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
	}
	if !working.Database.IsValid() {
		response.BadRequest(c, "database config is invalid")
		return
	}
	if req.Site != nil {
		applySiteConfig(&working, req.Site)
	}
	config.SetConfig(&working)
	working.Storage.BaseURL = strings.TrimRight(working.Server.Domain, "/") + "/uploads"
	config.SetConfig(&working)

	db, err := database.NewDatabase(&working.Database)
	if err != nil {
		response.BadRequest(c, "database validation failed: "+err.Error())
		return
	}
	sqlDB, _ := db.DB()
	defer func() {
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	}()

	if !requiredBootstrapTablesReady(db) {
		if err := database.RunBootstrapMigration(db, working.Database.Type); err != nil {
			response.InternalServerError(c, "failed to bootstrap database schema: "+err.Error())
			return
		}
	}

	if err := h.initializeData(c.Request.Context(), db, &working, req); err != nil {
		if strings.Contains(err.Error(), "system already initialized") {
			response.Forbidden(c, err.Error())
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}

	siteFlat := buildSiteOverrideMap(req.Site)
	if len(siteFlat) > 0 {
		nextCfg, _, err := config.SaveRuntimeOverrides(c.Request.Context(), db, &working, siteFlat)
		if err != nil {
			response.InternalServerError(c, "failed to save runtime settings: "+err.Error())
			return
		}
		working = *nextCfg
	}

	managed := config.BuildManagedStartupConfig(&working, true)
	if err := config.SaveManagedStartupConfig(h.getManagedConfigPath(), managed); err != nil {
		response.InternalServerError(c, "failed to save managed startup config: "+err.Error())
		return
	}
	config.SetConfig(&working)

	response.SuccessWithMessage(c, "初始化完成", gin.H{
		"initialized": true,
		"startup_mode": gin.H{
			"current": "bootstrap",
			"next":    "full",
		},
		"config_path": gin.H{
			"startup": h.getManagedConfigPath(),
			"runtime": "ty_system_settings",
		},
		"super_admin": gin.H{
			"phone":    strings.TrimSpace(req.SuperAdmin.Phone),
			"nickname": defaultString(strings.TrimSpace(req.SuperAdmin.Nickname), "超级管理员"),
		},
		"server_time":  time.Now().Format(time.RFC3339),
		"next_actions": []string{"请重启后端进入完整运行模式", "重启后使用超级管理员账号登录 Web 端"},
	})
}

func (h *BootstrapHandler) initializeData(ctx context.Context, db *gorm.DB, working *config.Config, req BootstrapInitializeRequest) error {
	superAdminCount := h.countSuperAdmins(ctx, db)
	if superAdminCount > 0 {
		managed, _ := config.LoadManagedStartupConfig(h.getManagedConfigPath())
		if managed != nil && managed.Initialized {
			return fmt.Errorf("system already initialized")
		}
	}

	orgRepo := infraRepo.NewOrganizationRepository(db)
	userRepo := infraRepo.NewUserRepository(db)

	rootOrg, err := ensureRootOrganization(ctx, orgRepo, working)
	if err != nil {
		return fmt.Errorf("failed to ensure root organization: %w", err)
	}

	if err := ensureSuperAdmin(ctx, db, userRepo, rootOrg.ID, req.SuperAdmin); err != nil {
		return fmt.Errorf("failed to create super admin: %w", err)
	}

	return nil
}

func ensureRootOrganization(ctx context.Context, orgRepo repository.OrganizationRepository, cfg *config.Config) (*entity.Organization, error) {
	if orgRepo == nil {
		return nil, fmt.Errorf("organization repository not configured")
	}
	root, err := orgRepo.FindByID(ctx, initDefaultOrgID)
	if err == nil && root != nil {
		changed := false
		if strings.TrimSpace(cfg.System.DefaultOrgName) != "" && root.Name != strings.TrimSpace(cfg.System.DefaultOrgName) {
			root.Name = strings.TrimSpace(cfg.System.DefaultOrgName)
			changed = true
		}
		if strings.TrimSpace(cfg.System.DefaultOrgCode) != "" && root.Code != strings.TrimSpace(cfg.System.DefaultOrgCode) {
			root.Code = strings.TrimSpace(cfg.System.DefaultOrgCode)
			changed = true
		}
		root.Type = entity.OrgTypeRoot
		root.Level = 1
		root.Status = entity.OrgStatusActive
		if changed {
			if err := orgRepo.Update(ctx, root); err != nil {
				return nil, err
			}
		}
		return root, nil
	}

	root, err = entity.NewRootOrganization(
		defaultString(strings.TrimSpace(cfg.System.DefaultOrgName), "助力团圆志愿者协会"),
		defaultString(strings.TrimSpace(cfg.System.DefaultOrgCode), "ROOT"),
	)
	if err != nil {
		return nil, err
	}
	root.Description = "系统初始化创建的默认根组织"
	root.Address = "中国"
	root.ContactName = "系统管理员"
	root.ContactPhone = "13800000000"
	root.SortOrder = 0
	if err := orgRepo.Create(ctx, root); err != nil {
		return nil, err
	}
	return root, nil
}

func ensureSuperAdmin(ctx context.Context, db *gorm.DB, userRepo repository.UserRepository, orgID string, req *BootstrapInitializeAdminRequest) error {
	adminPhone := strings.TrimSpace(req.Phone)
	existsPhone, err := userRepo.ExistsPhone(ctx, adminPhone)
	if err != nil {
		return err
	}

	if existsPhone {
		existing, findErr := userRepo.FindByPhone(ctx, adminPhone)
		if findErr == nil && existing != nil {
			if existing.Role != entity.RoleSuperAdmin {
				return fmt.Errorf("phone already exists")
			}
			return overwriteSuperAdmin(ctx, existing, userRepo, orgID, req, false)
		}
		return fmt.Errorf("phone already exists")
	}

	var seed entity.User
	if err := db.WithContext(ctx).Where("id = ?", initSeedSuperAdminID).First(&seed).Error; err == nil {
		return overwriteSuperAdmin(ctx, &seed, userRepo, orgID, req, true)
	}

	superAdmin := &entity.User{
		BaseEntity: entity.BaseEntity{ID: uuid.New().String()},
		Nickname:   defaultString(strings.TrimSpace(req.Nickname), "超级管理员"),
		Phone:      adminPhone,
		Email:      strings.TrimSpace(req.Email),
		Role:       entity.RoleSuperAdmin,
		Status:     entity.UserStatusActive,
		OrgID:      orgID,
	}
	if err := superAdmin.SetPassword(strings.TrimSpace(req.Password)); err != nil {
		return err
	}
	if err := superAdmin.Validate(); err != nil {
		return err
	}
	return userRepo.Create(ctx, superAdmin)
}

func overwriteSuperAdmin(ctx context.Context, existing *entity.User, userRepo repository.UserRepository, orgID string, req *BootstrapInitializeAdminRequest, allowPhoneChange bool) error {
	existing.Nickname = defaultString(strings.TrimSpace(req.Nickname), "超级管理员")
	if allowPhoneChange {
		existing.Phone = strings.TrimSpace(req.Phone)
	}
	existing.Email = strings.TrimSpace(req.Email)
	existing.Role = entity.RoleSuperAdmin
	existing.Status = entity.UserStatusActive
	existing.OrgID = orgID
	if err := existing.SetPassword(strings.TrimSpace(req.Password)); err != nil {
		return err
	}
	if err := existing.Validate(); err != nil {
		return err
	}
	return userRepo.Update(ctx, existing)
}

func requiredBootstrapTablesReady(db *gorm.DB) bool {
	required := []string{
		"ty_organizations",
		"ty_users",
		"ty_system_settings",
	}
	for _, table := range required {
		exists, err := database.TableExists(db, table)
		if err != nil || !exists {
			return false
		}
	}
	return true
}

func (h *BootstrapHandler) countSuperAdmins(ctx context.Context, db *gorm.DB) int64 {
	if h.userRepo != nil && h.db != nil && db == h.db {
		count, err := h.userRepo.CountByRole(ctx, entity.RoleSuperAdmin)
		if err == nil {
			return count
		}
	}
	var count int64
	_ = db.WithContext(ctx).Model(&entity.User{}).Where("role = ?", entity.RoleSuperAdmin).Count(&count).Error
	return count
}

func (h *BootstrapHandler) startupMode(dbConnected bool) string {
	if h.db != nil && dbConnected {
		return "full"
	}
	return "bootstrap"
}

func (h *BootstrapHandler) getManagedConfigPath() string {
	if strings.TrimSpace(h.managedConfigPath) != "" {
		return h.managedConfigPath
	}
	return "config/" + config.ManagedStartupConfigFilename()
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

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func settingsStorageLabel(configSource string) string {
	switch strings.TrimSpace(configSource) {
	case "file-config":
		return "file_config + ty_system_settings"
	default:
		return "managed_startup + ty_system_settings"
	}
}

func startupPathLabel(configSource, runtimePath, managedPath string) string {
	switch strings.TrimSpace(configSource) {
	case "file-config":
		return runtimePath
	default:
		return managedPath
	}
}
