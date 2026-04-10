package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const managedStartupConfigFilename = "bootstrap.runtime.json"

func ManagedStartupConfigFilename() string {
	return managedStartupConfigFilename
}

type ManagedStartupConfig struct {
	Initialized   bool           `json:"initialized"`
	InitializedAt string         `json:"initialized_at,omitempty"`
	Database      DatabaseConfig `json:"database"`
	Redis         RedisConfig    `json:"redis"`
	JWT           JWTConfig      `json:"jwt"`
	Storage       StorageConfig  `json:"storage"`
	Server        ServerConfig   `json:"server"`
}

func resolveManagedStartupConfigPath(configPath string) string {
	if strings.TrimSpace(configPath) == "" {
		return filepath.Join("config", managedStartupConfigFilename)
	}

	info, err := os.Stat(configPath)
	if err == nil && info.IsDir() {
		return filepath.Join(configPath, managedStartupConfigFilename)
	}

	if err == nil {
		return filepath.Join(filepath.Dir(configPath), managedStartupConfigFilename)
	}

	if filepath.Ext(configPath) != "" {
		return filepath.Join(filepath.Dir(configPath), managedStartupConfigFilename)
	}

	return filepath.Join(configPath, managedStartupConfigFilename)
}

func LoadManagedStartupConfig(path string) (*ManagedStartupConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("读取托管启动配置失败: %w", err)
	}

	var stored ManagedStartupConfig
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, fmt.Errorf("解析托管启动配置失败: %w", err)
	}

	return &stored, nil
}

func SaveManagedStartupConfig(path string, stored *ManagedStartupConfig) error {
	if stored == nil {
		return fmt.Errorf("startup config is nil")
	}
	if strings.TrimSpace(stored.JWT.Secret) == "" || len(strings.TrimSpace(stored.JWT.Secret)) < 32 {
		stored.JWT.Secret = generateRandomHex(32)
	}
	if strings.TrimSpace(path) == "" {
		path = filepath.Join("config", managedStartupConfigFilename)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("创建启动配置目录失败: %w", err)
	}
	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化托管启动配置失败: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("写入托管启动配置失败: %w", err)
	}
	return nil
}

func BuildManagedStartupConfig(cfg *Config, initialized bool) *ManagedStartupConfig {
	if cfg == nil {
		return nil
	}
	stored := &ManagedStartupConfig{
		Initialized: initialized,
		Database:    cfg.Database,
		Redis:       cfg.Redis,
		JWT:         cfg.JWT,
		Storage:     cfg.Storage,
		Server: ServerConfig{
			Port:           cfg.Server.Port,
			Mode:           cfg.Server.Mode,
			Domain:         cfg.Server.Domain,
			ReadTimeout:    cfg.Server.ReadTimeout,
			WriteTimeout:   cfg.Server.WriteTimeout,
			MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
			CORSOrigins:    cfg.Server.CORSOrigins,
		},
	}
	if initialized {
		stored.InitializedAt = time.Now().Format(time.RFC3339)
	}
	return stored
}

func MergeManagedStartupConfig(base *Config, stored *ManagedStartupConfig) Config {
	if base == nil {
		base = &Config{}
	}
	merged := *base

	merged.Database = stored.Database
	merged.Redis = stored.Redis
	merged.JWT = stored.JWT
	merged.Storage = stored.Storage

	if strings.TrimSpace(stored.Server.Port) != "" {
		merged.Server.Port = stored.Server.Port
	}
	if strings.TrimSpace(stored.Server.Mode) != "" {
		merged.Server.Mode = stored.Server.Mode
	}
	if strings.TrimSpace(stored.Server.Domain) != "" {
		merged.Server.Domain = stored.Server.Domain
	}
	if stored.Server.ReadTimeout > 0 {
		merged.Server.ReadTimeout = stored.Server.ReadTimeout
	}
	if stored.Server.WriteTimeout > 0 {
		merged.Server.WriteTimeout = stored.Server.WriteTimeout
	}
	if stored.Server.MaxHeaderBytes > 0 {
		merged.Server.MaxHeaderBytes = stored.Server.MaxHeaderBytes
	}
	if strings.TrimSpace(stored.Server.CORSOrigins) != "" {
		merged.Server.CORSOrigins = stored.Server.CORSOrigins
	}

	return merged
}
