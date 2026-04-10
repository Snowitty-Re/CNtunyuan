package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config 全局配置
type Config struct {
	Server       ServerConfig       `mapstructure:"server"`
	Database     DatabaseConfig     `mapstructure:"database"`
	Redis        RedisConfig        `mapstructure:"redis"`
	JWT          JWTConfig          `mapstructure:"jwt"`
	WeChat       WeChatConfig       `mapstructure:"wechat"`
	Storage      StorageConfig      `mapstructure:"storage"`
	SMS          SMSConfig          `mapstructure:"sms"`
	Email        EmailConfig        `mapstructure:"email"`
	Map          MapConfig          `mapstructure:"map"`
	Log          LogConfig          `mapstructure:"log"`
	Notification NotificationConfig `mapstructure:"notification"`
	System       SystemConfig       `mapstructure:"system"`
	Security     SecurityConfig     `mapstructure:"security"`
	Backup       BackupConfig       `mapstructure:"backup"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port           string `mapstructure:"port"`
	Mode           string `mapstructure:"mode"`
	Domain         string `mapstructure:"domain"`
	ReadTimeout    int    `mapstructure:"read_timeout"`
	WriteTimeout   int    `mapstructure:"write_timeout"`
	MaxHeaderBytes int    `mapstructure:"max_header_bytes"`
	CORSOrigins    string `mapstructure:"cors_origins"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	MaxLoginAttempts int `mapstructure:"max_login_attempts"`
	LockoutDuration  int `mapstructure:"lockout_duration"` // 锁定时长（秒）
}

// DatabaseType 数据库类型
type DatabaseType string

const (
	DatabaseTypePostgres DatabaseType = "postgres"
	DatabaseTypeMySQL    DatabaseType = "mysql"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type            DatabaseType `mapstructure:"type"` // 数据库类型: postgres/mysql
	Host            string       `mapstructure:"host"`
	Port            int          `mapstructure:"port"`
	User            string       `mapstructure:"user"`
	Password        string       `mapstructure:"password"`
	Database        string       `mapstructure:"database"`
	SSLMode         string       `mapstructure:"ssl_mode"` // PostgreSQL 使用
	Timezone        string       `mapstructure:"timezone"` // 数据库会话时区
	Charset         string       `mapstructure:"charset"`  // MySQL 使用，默认 utf8mb4
	MaxIdleConns    int          `mapstructure:"max_idle_conns"`
	MaxOpenConns    int          `mapstructure:"max_open_conns"`
	ConnMaxLifetime int          `mapstructure:"conn_max_lifetime"`
}

// IsValid 检查数据库配置是否有效
func (c *DatabaseConfig) IsValid() bool {
	return c.Type != "" && c.Host != "" && c.Port > 0 &&
		c.User != "" && c.Database != ""
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	timezone := strings.TrimSpace(c.Timezone)
	if timezone == "" {
		timezone = "Asia/Shanghai"
	}

	switch c.Type {
	case DatabaseTypeMySQL:
		charset := c.Charset
		if charset == "" {
			charset = "utf8mb4"
		}
		// MySQL DSN: user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			c.User, c.Password, c.Host, c.Port, c.Database, charset)
	case DatabaseTypePostgres:
		// PostgreSQL DSN
		sslMode := c.SSLMode
		if sslMode == "" {
			sslMode = "disable"
		}
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s client_encoding=UTF8 TimeZone=%s",
			c.Host, c.Port, c.User, c.Password, c.Database, sslMode, timezone)
		return dsn
	default:
		return ""
	}
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireTime  int64  `mapstructure:"expire_time"`
	RefreshTime int64  `mapstructure:"refresh_time"`
}

// WeChatConfig 微信小程序配置
type WeChatConfig struct {
	AppID       string `mapstructure:"app_id"`
	AppSecret   string `mapstructure:"app_secret"`
	MchID       string `mapstructure:"mch_id"`
	APIKey      string `mapstructure:"api_key"`
	NotifyURL   string `mapstructure:"notify_url"`
	EnableLogin bool   `mapstructure:"enable_login"` // 是否启用微信登录
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type         string `mapstructure:"type"`
	LocalPath    string `mapstructure:"local_path"`
	BaseURL      string `mapstructure:"base_url"`
	MaxFileSize  int64  `mapstructure:"max_file_size"`
	AllowedTypes string `mapstructure:"allowed_types"`
	// OSS配置
	OSSAccessKeyID     string `mapstructure:"oss_access_key_id"`
	OSSAccessKeySecret string `mapstructure:"oss_access_key_secret"`
	OSSEndpoint        string `mapstructure:"oss_endpoint"`
	OSSBucket          string `mapstructure:"oss_bucket"`
	OSSRegion          string `mapstructure:"oss_region"`
	// COS配置
	COSSecretID  string `mapstructure:"cos_secret_id"`
	COSSecretKey string `mapstructure:"cos_secret_key"`
	COSBucket    string `mapstructure:"cos_bucket"`
	COSRegion    string `mapstructure:"cos_region"`
}

// SMSConfig 短信配置
type SMSConfig struct {
	Provider           string `mapstructure:"provider"`
	SignName           string `mapstructure:"sign_name"`
	DevMode            bool   `mapstructure:"dev_mode"`
	CodeExpiry         int    `mapstructure:"code_expiry"` // 验证码有效期（秒）
	TemplateVerifyCode string `mapstructure:"template_verify_code"`
	TemplateResetPwd   string `mapstructure:"template_reset_password"`
	TemplateChangePh   string `mapstructure:"template_change_phone"`
	AliyunAccessKeyID  string `mapstructure:"aliyun_access_key_id"`
	AliyunAccessSecret string `mapstructure:"aliyun_access_key_secret"`
	TencentSecretID    string `mapstructure:"tencent_secret_id"`
	TencentSecretKey   string `mapstructure:"tencent_secret_key"`
	TencentAppID       string `mapstructure:"tencent_app_id"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	SMTPHost     string `mapstructure:"smtp_host"`
	SMTPPort     int    `mapstructure:"smtp_port"`
	SMTPUser     string `mapstructure:"smtp_user"`
	SMTPPassword string `mapstructure:"smtp_password"`
	FromName     string `mapstructure:"from_name"`
	UseTLS       bool   `mapstructure:"use_tls"`
}

// MapConfig 地图配置
type MapConfig struct {
	Provider   string `mapstructure:"provider"`
	Key        string `mapstructure:"key"`
	TencentKey string `mapstructure:"tencent_key"`
	AmapKey    string `mapstructure:"amap_key"`
	BaiduKey   string `mapstructure:"baidu_key"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	OutputPath string `mapstructure:"output_path"`
	FileName   string `mapstructure:"file_name"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// NotificationConfig 消息推送配置
type NotificationConfig struct {
	PushEnabled       bool   `mapstructure:"push_enabled"`
	GetuiAppID        string `mapstructure:"getui_app_id"`
	GetuiAppKey       string `mapstructure:"getui_app_key"`
	GetuiMasterSecret string `mapstructure:"getui_master_secret"`
	JPushAppKey       string `mapstructure:"jpush_app_key"`
	JPushMasterSecret string `mapstructure:"jpush_master_secret"`
}

// SystemConfig 系统配置
type SystemConfig struct {
	DefaultOrgName                    string `mapstructure:"default_org_name"`
	LogoURL                           string `mapstructure:"logo_url"`
	DefaultOrgCode                    string `mapstructure:"default_org_code"`
	EnableRegister                    bool   `mapstructure:"enable_register"`
	EnableWechatLogin                 bool   `mapstructure:"enable_wechat_login"`
	EnableWechatLoginWeb              bool   `mapstructure:"enable_wechat_login_web"`
	EnableWechatLoginMiniProgram      bool   `mapstructure:"enable_wechat_login_mini_program"`
	EnableSMSLogin                    bool   `mapstructure:"enable_sms_login"`
	AuthzPolicyChangeRequiresApproval bool   `mapstructure:"authz_policy_change_requires_approval"`
	AuthzPolicyChangeApprovalCode     string `mapstructure:"authz_policy_change_approval_code"`
	AuthzPolicyRequestExpireHours     int    `mapstructure:"authz_policy_request_expire_hours"`
	AdminIPs                          string `mapstructure:"admin_ips"`
	RateLimit                         int    `mapstructure:"rate_limit"`
}

// BackupConfig 备份配置
type BackupConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	BackupDir string `mapstructure:"backup_dir"` // 备份文件存放目录
	Retention int    `mapstructure:"retention"`  // 保留天数
}

var globalConfig *Config
var globalStartupSource = "bootstrap-managed"
var globalStartupConfigPath = "default"

type fileConfigOptions struct {
	Required          bool
	UseManagedStartup bool
}

// LoadConfig 加载配置
func LoadConfig(configPath string) (*Config, error) {
	return loadConfig(configPath, fileConfigOptions{Required: true, UseManagedStartup: true})
}

// LoadStartupConfig 加载启动配置；允许 config.yaml 缺失，并叠加系统托管的启动配置。
func LoadStartupConfig(configPath string) (*Config, error) {
	return loadConfig(configPath, fileConfigOptions{Required: false, UseManagedStartup: true})
}

// LoadConfigFileOnly 强制仅读取本地配置文件，不使用托管启动配置。
func LoadConfigFileOnly(configPath string) (*Config, error) {
	return loadConfig(configPath, fileConfigOptions{Required: true, UseManagedStartup: false})
}

func loadConfig(configPath string, opts fileConfigOptions) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	applyConfigPath(v, configPath)

	// 设置默认值
	setDefaultsFor(v)

	// 启用环境变量覆盖（格式: CNTY_DATABASE_PASSWORD 覆盖 database.password）
	v.SetEnvPrefix("CNTY")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok || opts.Required {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	if opts.UseManagedStartup {
		loaded, err := LoadManagedStartupConfig(resolveManagedStartupConfigPath(configPath))
		if err != nil {
			return nil, err
		}
		if loaded != nil {
			cfg = MergeManagedStartupConfig(&cfg, loaded)
		}
	}

	normalizeConfig(&cfg)

	globalConfig = &cfg
	return &cfg, nil
}

func applyConfigPath(v *viper.Viper, configPath string) {
	if configPath != "" {
		// 检查是否是文件路径
		info, err := os.Stat(configPath)
		if err == nil && info.IsDir() {
			// 是目录，在该目录下查找 config.yaml
			v.AddConfigPath(configPath)
			v.SetConfigName("config")
		} else if err == nil {
			// 是文件
			v.SetConfigFile(configPath)
		} else {
			// 路径不存在，尝试作为目录处理
			v.AddConfigPath(configPath)
			v.SetConfigName("config")
		}
	} else {
		v.AddConfigPath("./config")
		v.SetConfigName("config")
	}
}

func normalizeConfig(cfg *Config) {
	if cfg == nil {
		return
	}
	if strings.TrimSpace(cfg.Server.Port) == "" {
		cfg.Server.Port = "8080"
	}
	if strings.TrimSpace(cfg.Server.Mode) == "" {
		cfg.Server.Mode = "release"
	}
	if strings.TrimSpace(cfg.Server.Domain) == "" {
		cfg.Server.Domain = "http://localhost:" + cfg.Server.Port
	}

	if strings.TrimSpace(cfg.JWT.Secret) == "" || len(strings.TrimSpace(cfg.JWT.Secret)) < 32 {
		cfg.JWT.Secret = generateRandomHex(32)
	}
	if cfg.JWT.ExpireTime <= 0 {
		cfg.JWT.ExpireTime = 604800
	}
	if cfg.JWT.RefreshTime <= 0 {
		cfg.JWT.RefreshTime = 2592000
	}

	if cfg.Storage.Type == "" {
		cfg.Storage.Type = "local"
	}
	if strings.EqualFold(cfg.Storage.Type, "local") {
		if strings.TrimSpace(cfg.Storage.LocalPath) == "" {
			cfg.Storage.LocalPath = "./uploads"
		}
		if !filepath.IsAbs(cfg.Storage.LocalPath) {
			if abs, err := filepath.Abs(cfg.Storage.LocalPath); err == nil {
				cfg.Storage.LocalPath = abs
			}
		}
		if strings.TrimSpace(cfg.Storage.BaseURL) == "" {
			cfg.Storage.BaseURL = strings.TrimRight(cfg.Server.Domain, "/") + "/uploads"
		}
	}

	if strings.TrimSpace(cfg.Log.OutputPath) == "" {
		cfg.Log.OutputPath = "./logs"
	}
	if strings.TrimSpace(cfg.Backup.BackupDir) == "" {
		cfg.Backup.BackupDir = "./backups"
	}
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	return globalConfig
}

// SetConfig 设置全局配置（用于配置热更新）
func SetConfig(cfg *Config) {
	globalConfig = cfg
}

func SetStartupMetadata(source, configPath string) {
	if strings.TrimSpace(source) != "" {
		globalStartupSource = strings.TrimSpace(source)
	}
	if strings.TrimSpace(configPath) == "" {
		globalStartupConfigPath = "default"
	} else {
		globalStartupConfigPath = strings.TrimSpace(configPath)
	}
}

func GetStartupMetadata() (source string, configPath string) {
	return globalStartupSource, globalStartupConfigPath
}

func setDefaults() {
	setDefaultsFor(viper.GetViper())
}

func setDefaultsFor(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.mode", "release")
	v.SetDefault("server.domain", "http://localhost:8080")
	v.SetDefault("server.read_timeout", 30)
	v.SetDefault("server.write_timeout", 30)
	v.SetDefault("server.max_header_bytes", 1048576)
	v.SetDefault("server.cors_origins", "http://localhost:3000,http://localhost:5173")

	// Database defaults
	v.SetDefault("database.type", "postgres")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.database", "cntuanyuan")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.charset", "UTF8")
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.max_open_conns", 100)
	v.SetDefault("database.conn_max_lifetime", 3600)

	// Redis defaults
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.min_idle_conns", 2)

	// JWT defaults
	v.SetDefault("jwt.expire_time", 604800)   // 7天
	v.SetDefault("jwt.refresh_time", 2592000) // 30天

	// Storage defaults
	v.SetDefault("storage.type", "local")
	v.SetDefault("storage.local_path", "./uploads")
	v.SetDefault("storage.base_url", "http://localhost:8080/uploads")
	v.SetDefault("storage.max_file_size", 52428800) // 50MB
	v.SetDefault("storage.allowed_types", "jpg,jpeg,png,gif,webp,mp4,webm,mp3,wav,m4a,aac,ogg,opus,flac")

	// SMS defaults
	v.SetDefault("sms.provider", "aliyun")
	v.SetDefault("sms.sign_name", "助力团圆")
	v.SetDefault("sms.dev_mode", false)
	v.SetDefault("sms.code_expiry", 300)
	v.SetDefault("sms.template_verify_code", "verify_code") // 注册验证码模板（参数：{1}=code）
	v.SetDefault("sms.template_reset_password", "")
	v.SetDefault("sms.template_change_phone", "")

	// Email defaults
	v.SetDefault("email.enabled", false)
	v.SetDefault("email.smtp_host", "smtp.qq.com")
	v.SetDefault("email.smtp_port", 587)
	v.SetDefault("email.use_tls", true)

	// Map defaults
	v.SetDefault("map.provider", "tencent")

	// Log defaults
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("log.output_path", "./logs")
	v.SetDefault("log.file_name", "app.log")
	v.SetDefault("log.max_size", 100)
	v.SetDefault("log.max_backups", 10)
	v.SetDefault("log.max_age", 30)
	v.SetDefault("log.compress", true)

	// Notification defaults
	v.SetDefault("notification.push_enabled", false)

	// System defaults
	v.SetDefault("system.default_org_name", "助力团圆志愿者协会")
	v.SetDefault("system.logo_url", "")
	v.SetDefault("system.default_org_code", "ROOT")
	v.SetDefault("system.enable_register", true)
	v.SetDefault("system.enable_wechat_login", true)
	v.SetDefault("system.enable_sms_login", false)
	v.SetDefault("system.authz_policy_change_requires_approval", false)
	v.SetDefault("system.authz_policy_change_approval_code", "")
	v.SetDefault("system.authz_policy_request_expire_hours", 72)
	v.SetDefault("system.rate_limit", 100)

	// Security defaults
	v.SetDefault("security.max_login_attempts", 5)
	v.SetDefault("security.lockout_duration", 1800)

	// Backup defaults
	v.SetDefault("backup.enabled", false)
	v.SetDefault("backup.backup_dir", "./backups")
	v.SetDefault("backup.retention", 7)
}

func generateRandomHex(byteLen int) string {
	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err != nil {
		return strings.Repeat("a", byteLen*2)
	}
	return hex.EncodeToString(buf)
}
