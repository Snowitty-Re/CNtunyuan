package config

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type SystemSetting struct {
	ID           string         `gorm:"column:id;primaryKey"`
	CreatedAt    time.Time      `gorm:"column:created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index"`
	Category     string         `gorm:"column:category"`
	SettingKey   string         `gorm:"column:setting_key;uniqueIndex"`
	SettingValue string         `gorm:"column:setting_value"`
	ValueType    string         `gorm:"column:value_type"`
}

func (SystemSetting) TableName() string {
	return "ty_system_settings"
}

var databaseManagedConfigKeys = map[string]struct{}{
	"server.domain":                                {},
	"server.cors_origins":                          {},
	"redis.host":                                   {},
	"redis.port":                                   {},
	"redis.password":                               {},
	"redis.db":                                     {},
	"redis.pool_size":                              {},
	"redis.min_idle_conns":                         {},
	"jwt.secret":                                   {},
	"jwt.expire_time":                              {},
	"jwt.refresh_time":                             {},
	"wechat.app_id":                                {},
	"wechat.app_secret":                            {},
	"wechat.enable_login":                          {},
	"wechat.mch_id":                                {},
	"wechat.api_key":                               {},
	"wechat.notify_url":                            {},
	"storage.type":                                 {},
	"storage.local_path":                           {},
	"storage.base_url":                             {},
	"storage.max_file_size":                        {},
	"storage.allowed_types":                        {},
	"storage.oss_access_key_id":                    {},
	"storage.oss_access_key_secret":                {},
	"storage.oss_endpoint":                         {},
	"storage.oss_bucket":                           {},
	"storage.oss_region":                           {},
	"storage.cos_secret_id":                        {},
	"storage.cos_secret_key":                       {},
	"storage.cos_bucket":                           {},
	"storage.cos_region":                           {},
	"sms.provider":                                 {},
	"sms.sign_name":                                {},
	"sms.dev_mode":                                 {},
	"sms.code_expiry":                              {},
	"sms.template_verify_code":                     {},
	"sms.template_reset_password":                  {},
	"sms.template_change_phone":                    {},
	"sms.aliyun_access_key_id":                     {},
	"sms.aliyun_access_key_secret":                 {},
	"sms.tencent_secret_id":                        {},
	"sms.tencent_secret_key":                       {},
	"sms.tencent_app_id":                           {},
	"email.enabled":                                {},
	"email.smtp_host":                              {},
	"email.smtp_port":                              {},
	"email.smtp_user":                              {},
	"email.smtp_password":                          {},
	"email.from_name":                              {},
	"email.use_tls":                                {},
	"map.provider":                                 {},
	"map.key":                                      {},
	"map.tencent_key":                              {},
	"map.amap_key":                                 {},
	"map.baidu_key":                                {},
	"log.level":                                    {},
	"log.format":                                   {},
	"log.output_path":                              {},
	"log.file_name":                                {},
	"log.max_size":                                 {},
	"log.max_backups":                              {},
	"log.max_age":                                  {},
	"log.compress":                                 {},
	"notification.push_enabled":                    {},
	"notification.getui_app_id":                    {},
	"notification.getui_app_key":                   {},
	"notification.getui_master_secret":             {},
	"notification.jpush_app_key":                   {},
	"notification.jpush_master_secret":             {},
	"system.default_org_name":                      {},
	"system.default_org_code":                      {},
	"system.enable_register":                       {},
	"system.enable_wechat_login":                   {},
	"system.enable_wechat_login_web":               {},
	"system.enable_wechat_login_mini_program":      {},
	"system.enable_sms_login":                      {},
	"system.authz_policy_change_requires_approval": {},
	"system.authz_policy_change_approval_code":     {},
	"system.authz_policy_request_expire_hours":     {},
	"system.admin_ips":                             {},
	"system.rate_limit":                            {},
	"security.max_login_attempts":                  {},
	"security.lockout_duration":                    {},
	"backup.enabled":                               {},
	"backup.backup_dir":                            {},
	"backup.retention":                             {},
}

func FlattenConfig(cfg *Config) map[string]interface{} {
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
		"sms.template_verify_code":     cfg.SMS.TemplateVerifyCode,
		"sms.template_reset_password":  cfg.SMS.TemplateResetPwd,
		"sms.template_change_phone":    cfg.SMS.TemplateChangePh,
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

		"system.default_org_name":                      cfg.System.DefaultOrgName,
		"system.default_org_code":                      cfg.System.DefaultOrgCode,
		"system.enable_register":                       cfg.System.EnableRegister,
		"system.enable_wechat_login":                   cfg.System.EnableWechatLogin,
		"system.enable_wechat_login_web":               ResolveWebWechatLoginEnabled(cfg),
		"system.enable_wechat_login_mini_program":      ResolveMiniProgramWechatLoginEnabled(cfg),
		"system.enable_sms_login":                      cfg.System.EnableSMSLogin,
		"system.authz_policy_change_requires_approval": cfg.System.AuthzPolicyChangeRequiresApproval,
		"system.authz_policy_change_approval_code":     cfg.System.AuthzPolicyChangeApprovalCode,
		"system.authz_policy_request_expire_hours":     cfg.System.AuthzPolicyRequestExpireHours,
		"system.admin_ips":                             cfg.System.AdminIPs,
		"system.rate_limit":                            cfg.System.RateLimit,

		"security.max_login_attempts": cfg.Security.MaxLoginAttempts,
		"security.lockout_duration":   cfg.Security.LockoutDuration,

		"backup.enabled":    cfg.Backup.Enabled,
		"backup.backup_dir": cfg.Backup.BackupDir,
		"backup.retention":  cfg.Backup.Retention,
	}
}

func ResolveWebWechatLoginEnabled(cfg *Config) bool {
	if cfg == nil {
		return false
	}
	if globalConfig != nil && viper.IsSet("system.enable_wechat_login_web") {
		return cfg.System.EnableWechatLoginWeb
	}
	return cfg.System.EnableWechatLogin
}

func ResolveMiniProgramWechatLoginEnabled(cfg *Config) bool {
	if cfg == nil {
		return false
	}
	if globalConfig != nil && viper.IsSet("system.enable_wechat_login_mini_program") {
		return cfg.System.EnableWechatLoginMiniProgram
	}
	return cfg.System.EnableWechatLogin
}

func UnflattenConfig(flat map[string]interface{}) map[string]interface{} {
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

func CoerceValue(raw interface{}, current interface{}) interface{} {
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

func ApplyFlatOverrides(cfg *Config, overrides map[string]interface{}) (*Config, error) {
	flat := FlattenConfig(cfg)
	for key, raw := range overrides {
		current, ok := flat[key]
		if !ok {
			continue
		}
		flat[key] = CoerceValue(raw, current)
	}

	var next Config
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          "mapstructure",
		Result:           &next,
		WeaklyTypedInput: true,
	})
	if err != nil {
		return nil, err
	}
	if err := decoder.Decode(UnflattenConfig(flat)); err != nil {
		return nil, err
	}
	return &next, nil
}

func PersistableConfigKeys() map[string]struct{} {
	keys := make(map[string]struct{}, len(databaseManagedConfigKeys))
	for key := range databaseManagedConfigKeys {
		keys[key] = struct{}{}
	}
	return keys
}

func LoadRuntimeOverrides(ctx context.Context, db *gorm.DB, base *Config) (*Config, map[string]interface{}, error) {
	if db == nil || base == nil {
		return base, map[string]interface{}{}, nil
	}
	if !db.Migrator().HasTable((&SystemSetting{}).TableName()) {
		return base, map[string]interface{}{}, nil
	}

	var rows []SystemSetting
	if err := db.WithContext(ctx).Where("deleted_at IS NULL").Find(&rows).Error; err != nil {
		return nil, nil, err
	}
	overrides := make(map[string]interface{}, len(rows))
	for _, row := range rows {
		if _, ok := databaseManagedConfigKeys[row.SettingKey]; !ok {
			continue
		}
		var decoded interface{}
		if err := json.Unmarshal([]byte(row.SettingValue), &decoded); err != nil {
			decoded = row.SettingValue
		}
		overrides[row.SettingKey] = decoded
	}

	next, err := ApplyFlatOverrides(base, overrides)
	if err != nil {
		return nil, nil, err
	}
	return next, overrides, nil
}

func SaveRuntimeOverrides(ctx context.Context, db *gorm.DB, cfg *Config, updates map[string]interface{}) (*Config, map[string]interface{}, error) {
	if db == nil {
		return nil, nil, fmt.Errorf("db is nil")
	}
	if cfg == nil {
		return nil, nil, fmt.Errorf("config is nil")
	}
	if !db.Migrator().HasTable((&SystemSetting{}).TableName()) {
		return nil, nil, fmt.Errorf("runtime settings table missing")
	}

	currentFlat := FlattenConfig(cfg)
	persisted := make(map[string]interface{})
	for key, raw := range updates {
		if _, ok := databaseManagedConfigKeys[key]; !ok {
			continue
		}
		current, ok := currentFlat[key]
		if !ok {
			continue
		}
		persisted[key] = CoerceValue(raw, current)
	}

	if len(persisted) == 0 {
		next, _, err := LoadRuntimeOverrides(ctx, db, cfg)
		return next, map[string]interface{}{}, err
	}

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for key, value := range persisted {
			payload, err := json.Marshal(value)
			if err != nil {
				return err
			}

			updates := map[string]interface{}{
				"category":      categoryFromSettingKey(key),
				"setting_value": string(payload),
				"value_type":    detectValueType(value),
				"updated_at":    time.Now(),
				"deleted_at":    nil,
			}

			var existing SystemSetting
			err = tx.Where("setting_key = ? AND deleted_at IS NULL", key).Take(&existing).Error
			if err == nil {
				if err := tx.Model(&existing).Updates(updates).Error; err != nil {
					return err
				}
				continue
			}
			if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}

			row := &SystemSetting{
				ID:           uuid.NewString(),
				Category:     categoryFromSettingKey(key),
				SettingKey:   key,
				SettingValue: string(payload),
				ValueType:    detectValueType(value),
			}
			if err := tx.Create(row).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return LoadRuntimeOverrides(ctx, db, cfg)
}

func categoryFromSettingKey(key string) string {
	parts := strings.SplitN(key, ".", 2)
	if len(parts) == 0 {
		return "system"
	}
	return parts[0]
}

func detectValueType(value interface{}) string {
	switch value.(type) {
	case bool:
		return "bool"
	case int, int32, int64, float32, float64:
		return "number"
	default:
		return "string"
	}
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
