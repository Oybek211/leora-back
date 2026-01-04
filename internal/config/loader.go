package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds global configuration values.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
}

// AppConfig stores application-level settings.
type AppConfig struct {
	Name          string        `mapstructure:"name"`
	Env           string        `mapstructure:"env"`
	Port          int           `mapstructure:"port"`
	BasePath      string        `mapstructure:"basePath"`
	JWTSecret     string        `mapstructure:"jwtSecret"`
	JWTAccessTTL  time.Duration `mapstructure:"jwtAccessTTL"`
	JWTRefreshTTL time.Duration `mapstructure:"jwtRefreshTTL"`
}

// DatabaseConfig stores PostgreSQL connection info.
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

// RedisConfig stores Redis connection info.
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// Load reads configuration from env and config file.
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(path)
	v.AutomaticEnv()
	v.SetEnvPrefix("LEORA")

	v.SetDefault("app.basePath", "/api/v1")
	v.SetDefault("app.port", 9090)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.App.JWTSecret == "" {
		return nil, fmt.Errorf("jwt secret not configured")
	}

	accessTTL := 7 * 24 * time.Hour
	if cfg.App.JWTAccessTTL != accessTTL {
		cfg.App.JWTAccessTTL = accessTTL
	}
	if cfg.App.JWTRefreshTTL == 0 {
		cfg.App.JWTRefreshTTL = 24 * time.Hour
	}

	return &cfg, nil
}
