package config

import (
	"fmt"

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
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port           string `mapstructure:"port"`
	Mode           string `mapstructure:"mode"`
	ReadTimeout    int    `mapstructure:"read_timeout"`
	WriteTimeout   int    `mapstructure:"write_timeout"`
	MaxHeaderBytes int    `mapstructure:"max_header_bytes"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	SSLMode         string `mapstructure:"ssl_mode"`
	Charset         string `mapstructure:"charset"` // 字符集，默认 UTF8
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
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
	AppID     string `mapstructure:"app_id"`
	AppSecret string `mapstructure:"app_secret"`
	MchID     string `mapstructure:"mch_id"`
	APIKey    string `mapstructure:"api_key"`
	NotifyURL string `mapstructure:"notify_url"`
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
	Provider          string `mapstructure:"provider"`
	SignName          string `mapstructure:"sign_name"`
	AliyunAccessKeyID string `mapstructure:"aliyun_access_key_id"`
	AliyunAccessSecret string `mapstructure:"aliyun_access_key_secret"`
	TencentSecretID   string `mapstructure:"tencent_secret_id"`
	TencentSecretKey  string `mapstructure:"tencent_secret_key"`
	TencentAppID      string `mapstructure:"tencent_app_id"`
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
	Provider    string `mapstructure:"provider"`
	Key         string `mapstructure:"key"`
	TencentKey  string `mapstructure:"tencent_key"`
	AmapKey     string `mapstructure:"amap_key"`
	BaiduKey    string `mapstructure:"baidu_key"`
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
	DefaultOrgName    string `mapstructure:"default_org_name"`
	DefaultOrgCode    string `mapstructure:"default_org_code"`
	EnableRegister    bool   `mapstructure:"enable_register"`
	EnableWechatLogin bool   `mapstructure:"enable_wechat_login"`
	EnableSMSLogin    bool   `mapstructure:"enable_sms_login"`
	AdminIPs          string `mapstructure:"admin_ips"`
	RateLimit         int    `mapstructure:"rate_limit"`
}

var globalConfig *Config

// LoadConfig 加载配置
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigType("yaml")

	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.AddConfigPath("./config")
		viper.SetConfigName("config")
	}

	// 设置默认值
	setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	globalConfig = &config
	return &config, nil
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	return globalConfig
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	charset := c.Charset
	if charset == "" {
		charset = "UTF8"
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s client_encoding=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode, charset)
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "release")
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)
	viper.SetDefault("server.max_header_bytes", 1048576)

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.database", "cntunyuan")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.charset", "UTF8")
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.conn_max_lifetime", 3600)

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conns", 2)

	// JWT defaults
	viper.SetDefault("jwt.expire_time", 604800) // 7天
	viper.SetDefault("jwt.refresh_time", 2592000) // 30天

	// Storage defaults
	viper.SetDefault("storage.type", "local")
	viper.SetDefault("storage.local_path", "./uploads")
	viper.SetDefault("storage.base_url", "http://localhost:8080/uploads")
	viper.SetDefault("storage.max_file_size", 52428800) // 50MB
	viper.SetDefault("storage.allowed_types", "jpg,png,gif,mp4,mp3,wav")

	// SMS defaults
	viper.SetDefault("sms.provider", "aliyun")
	viper.SetDefault("sms.sign_name", "团圆寻亲")

	// Email defaults
	viper.SetDefault("email.enabled", false)
	viper.SetDefault("email.smtp_host", "smtp.qq.com")
	viper.SetDefault("email.smtp_port", 587)
	viper.SetDefault("email.use_tls", true)

	// Map defaults
	viper.SetDefault("map.provider", "tencent")

	// Log defaults
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output_path", "./logs")
	viper.SetDefault("log.file_name", "app.log")
	viper.SetDefault("log.max_size", 100)
	viper.SetDefault("log.max_backups", 10)
	viper.SetDefault("log.max_age", 30)
	viper.SetDefault("log.compress", true)

	// Notification defaults
	viper.SetDefault("notification.push_enabled", false)

	// System defaults
	viper.SetDefault("system.default_org_name", "团圆寻亲志愿者协会")
	viper.SetDefault("system.default_org_code", "ROOT")
	viper.SetDefault("system.enable_register", true)
	viper.SetDefault("system.enable_wechat_login", true)
	viper.SetDefault("system.enable_sms_login", false)
	viper.SetDefault("system.rate_limit", 100)
}
