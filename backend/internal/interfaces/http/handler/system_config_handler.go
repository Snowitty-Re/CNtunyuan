package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/middleware"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const maskedSecretValue = "******"

type SystemConfigHandler struct{}

type updateSystemConfigRequest struct {
	Config map[string]interface{} `json:"config"`
}

func NewSystemConfigHandler() *SystemConfigHandler {
	return &SystemConfigHandler{}
}

func (h *SystemConfigHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	system := router.Group("/system")
	system.Use(authMiddleware.Required(), middleware.RequireAdmin())
	{
		system.GET("/config", h.GetConfig)
		system.PUT("/config", h.UpdateConfig)
	}
}

// GetConfig 获取系统配置
// @Summary 获取系统配置
// @Description 获取当前系统配置（敏感字段已掩码，仅管理员可访问）
// @Tags 系统配置
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /system/config [get]
func (h *SystemConfigHandler) GetConfig(c *gin.Context) {
	cfg := config.GetConfig()
	if cfg == nil {
		response.InternalServerError(c, "config not loaded")
		return
	}

	flat := flattenConfig(cfg)
	for _, key := range sensitiveKeys() {
		if v, ok := flat[key]; ok {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				flat[key] = maskedSecretValue
			}
		}
	}

	response.Success(c, gin.H{
		"config":           flat,
		"sensitive_fields": sensitiveKeys(),
	})
}

// UpdateConfig 更新系统配置
// @Summary 更新系统配置
// @Description 更新并持久化系统配置到 config.yaml（敏感字段传掩码将保留原值）
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body updateSystemConfigRequest true "配置项（扁平 key-value）"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /system/config [put]
func (h *SystemConfigHandler) UpdateConfig(c *gin.Context) {
	cfg := config.GetConfig()
	if cfg == nil {
		response.InternalServerError(c, "config not loaded")
		return
	}

	var req updateSystemConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	if req.Config == nil {
		response.BadRequest(c, "config is required")
		return
	}

	flat := flattenConfig(cfg)
	sensitive := make(map[string]struct{}, len(sensitiveKeys()))
	for _, key := range sensitiveKeys() {
		sensitive[key] = struct{}{}
	}

	for key, raw := range req.Config {
		current, ok := flat[key]
		if !ok {
			continue
		}

		if _, isSensitive := sensitive[key]; isSensitive {
			val := strings.TrimSpace(toString(raw))
			if val == "" || val == maskedSecretValue {
				continue
			}
			flat[key] = val
			continue
		}

		flat[key] = coerceValue(raw, current)
	}

	nested := unflattenConfig(flat)
	bytes, err := yaml.Marshal(nested)
	if err != nil {
		response.InternalServerError(c, "failed to encode config")
		return
	}

	configPath := viper.ConfigFileUsed()
	if strings.TrimSpace(configPath) == "" {
		configPath = filepath.Join("config", "config.yaml")
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		response.InternalServerError(c, "failed to ensure config dir")
		return
	}
	if err := os.WriteFile(configPath, bytes, 0o600); err != nil {
		response.InternalServerError(c, "failed to save config")
		return
	}

	var newCfg config.Config
	if err := yaml.Unmarshal(bytes, &newCfg); err != nil {
		response.InternalServerError(c, "failed to apply config")
		return
	}
	config.SetConfig(&newCfg)

	masked := flattenConfig(&newCfg)
	for _, key := range sensitiveKeys() {
		if v, ok := masked[key]; ok {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				masked[key] = maskedSecretValue
			}
		}
	}

	response.SuccessWithMessage(c, "配置已保存（重启后将完全生效）", gin.H{
		"config":           masked,
		"sensitive_fields": sensitiveKeys(),
	})
}

func sensitiveKeys() []string {
	return []string{
		"database.password",
		"jwt.secret",
		"wechat.app_secret",
		"wechat.api_key",
		"storage.oss_access_key_secret",
		"storage.cos_secret_key",
		"sms.aliyun_access_key_secret",
		"sms.tencent_secret_key",
		"email.smtp_password",
		"map.key",
		"map.tencent_key",
		"map.amap_key",
		"map.baidu_key",
		"notification.getui_app_key",
		"notification.getui_master_secret",
		"notification.jpush_master_secret",
	}
}

func flattenConfig(cfg *config.Config) map[string]interface{} {
	return map[string]interface{}{
		"server.port":             cfg.Server.Port,
		"server.mode":             cfg.Server.Mode,
		"server.domain":           cfg.Server.Domain,
		"server.read_timeout":     cfg.Server.ReadTimeout,
		"server.write_timeout":    cfg.Server.WriteTimeout,
		"server.max_header_bytes": cfg.Server.MaxHeaderBytes,
		"server.cors_origins":     cfg.Server.CORSOrigins,

		"database.type":              string(cfg.Database.Type),
		"database.host":              cfg.Database.Host,
		"database.port":              cfg.Database.Port,
		"database.user":              cfg.Database.User,
		"database.password":          cfg.Database.Password,
		"database.database":          cfg.Database.Database,
		"database.ssl_mode":          cfg.Database.SSLMode,
		"database.timezone":          cfg.Database.Timezone,
		"database.charset":           cfg.Database.Charset,
		"database.max_idle_conns":    cfg.Database.MaxIdleConns,
		"database.max_open_conns":    cfg.Database.MaxOpenConns,
		"database.conn_max_lifetime": cfg.Database.ConnMaxLifetime,

		"redis.host":           cfg.Redis.Host,
		"redis.port":           cfg.Redis.Port,
		"redis.password":       cfg.Redis.Password,
		"redis.db":             cfg.Redis.DB,
		"redis.pool_size":      cfg.Redis.PoolSize,
		"redis.min_idle_conns": cfg.Redis.MinIdleConns,

		"jwt.secret":       cfg.JWT.Secret,
		"jwt.expire_time":  cfg.JWT.ExpireTime,
		"jwt.refresh_time": cfg.JWT.RefreshTime,

		"wechat.app_id":       cfg.WeChat.AppID,
		"wechat.app_secret":   cfg.WeChat.AppSecret,
		"wechat.enable_login": cfg.WeChat.EnableLogin,
		"wechat.mch_id":       cfg.WeChat.MchID,
		"wechat.api_key":      cfg.WeChat.APIKey,
		"wechat.notify_url":   cfg.WeChat.NotifyURL,

		"storage.type":                  cfg.Storage.Type,
		"storage.local_path":            cfg.Storage.LocalPath,
		"storage.base_url":              cfg.Storage.BaseURL,
		"storage.max_file_size":         cfg.Storage.MaxFileSize,
		"storage.allowed_types":         cfg.Storage.AllowedTypes,
		"storage.oss_access_key_id":     cfg.Storage.OSSAccessKeyID,
		"storage.oss_access_key_secret": cfg.Storage.OSSAccessKeySecret,
		"storage.oss_endpoint":          cfg.Storage.OSSEndpoint,
		"storage.oss_bucket":            cfg.Storage.OSSBucket,
		"storage.oss_region":            cfg.Storage.OSSRegion,
		"storage.cos_secret_id":         cfg.Storage.COSSecretID,
		"storage.cos_secret_key":        cfg.Storage.COSSecretKey,
		"storage.cos_bucket":            cfg.Storage.COSBucket,
		"storage.cos_region":            cfg.Storage.COSRegion,

		"sms.provider":                 cfg.SMS.Provider,
		"sms.sign_name":                cfg.SMS.SignName,
		"sms.dev_mode":                 cfg.SMS.DevMode,
		"sms.code_expiry":              cfg.SMS.CodeExpiry,
		"sms.aliyun_access_key_id":     cfg.SMS.AliyunAccessKeyID,
		"sms.aliyun_access_key_secret": cfg.SMS.AliyunAccessSecret,
		"sms.tencent_secret_id":        cfg.SMS.TencentSecretID,
		"sms.tencent_secret_key":       cfg.SMS.TencentSecretKey,
		"sms.tencent_app_id":           cfg.SMS.TencentAppID,

		"email.enabled":       cfg.Email.Enabled,
		"email.smtp_host":     cfg.Email.SMTPHost,
		"email.smtp_port":     cfg.Email.SMTPPort,
		"email.smtp_user":     cfg.Email.SMTPUser,
		"email.smtp_password": cfg.Email.SMTPPassword,
		"email.from_name":     cfg.Email.FromName,
		"email.use_tls":       cfg.Email.UseTLS,

		"map.provider":    cfg.Map.Provider,
		"map.key":         cfg.Map.Key,
		"map.tencent_key": cfg.Map.TencentKey,
		"map.amap_key":    cfg.Map.AmapKey,
		"map.baidu_key":   cfg.Map.BaiduKey,

		"log.level":       cfg.Log.Level,
		"log.format":      cfg.Log.Format,
		"log.output_path": cfg.Log.OutputPath,
		"log.file_name":   cfg.Log.FileName,
		"log.max_size":    cfg.Log.MaxSize,
		"log.max_backups": cfg.Log.MaxBackups,
		"log.max_age":     cfg.Log.MaxAge,
		"log.compress":    cfg.Log.Compress,

		"notification.push_enabled":        cfg.Notification.PushEnabled,
		"notification.getui_app_id":        cfg.Notification.GetuiAppID,
		"notification.getui_app_key":       cfg.Notification.GetuiAppKey,
		"notification.getui_master_secret": cfg.Notification.GetuiMasterSecret,
		"notification.jpush_app_key":       cfg.Notification.JPushAppKey,
		"notification.jpush_master_secret": cfg.Notification.JPushMasterSecret,

		"system.default_org_name":    cfg.System.DefaultOrgName,
		"system.default_org_code":    cfg.System.DefaultOrgCode,
		"system.enable_register":     cfg.System.EnableRegister,
		"system.enable_wechat_login": cfg.System.EnableWechatLogin,
		"system.enable_sms_login":    cfg.System.EnableSMSLogin,
		"system.admin_ips":           cfg.System.AdminIPs,
		"system.rate_limit":          cfg.System.RateLimit,

		"security.max_login_attempts": cfg.Security.MaxLoginAttempts,
		"security.lockout_duration":   cfg.Security.LockoutDuration,

		"backup.enabled":    cfg.Backup.Enabled,
		"backup.backup_dir": cfg.Backup.BackupDir,
		"backup.retention":  cfg.Backup.Retention,
	}
}

func unflattenConfig(flat map[string]interface{}) map[string]interface{} {
	root := map[string]interface{}{}
	for key, value := range flat {
		parts := strings.Split(key, ".")
		if len(parts) < 2 {
			continue
		}
		cur := root
		for i := 0; i < len(parts)-1; i++ {
			p := parts[i]
			next, ok := cur[p]
			if !ok {
				m := map[string]interface{}{}
				cur[p] = m
				cur = m
				continue
			}
			m, ok := next.(map[string]interface{})
			if !ok {
				m = map[string]interface{}{}
				cur[p] = m
			}
			cur = m
		}
		cur[parts[len(parts)-1]] = value
	}
	return root
}

func toString(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case fmt.Stringer:
		return x.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func coerceValue(raw interface{}, current interface{}) interface{} {
	switch current.(type) {
	case bool:
		if v, ok := raw.(bool); ok {
			return v
		}
		s := strings.TrimSpace(strings.ToLower(toString(raw)))
		return s == "1" || s == "true" || s == "yes" || s == "on"
	case int:
		if n, err := strconv.Atoi(strings.TrimSpace(toString(raw))); err == nil {
			return n
		}
		return current
	case int64:
		if n, err := strconv.ParseInt(strings.TrimSpace(toString(raw)), 10, 64); err == nil {
			return n
		}
		return current
	default:
		return toString(raw)
	}
}
