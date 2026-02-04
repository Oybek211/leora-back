package config

import (
	"fmt"
	"os"
	"strconv"
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
	Name              string        `mapstructure:"name"`
	Env               string        `mapstructure:"env"`
	Port              int           `mapstructure:"port"`
	BasePath          string        `mapstructure:"basePath"`
	JWTSecret         string        `mapstructure:"jwtSecret"`
	JWTAccessTTL      time.Duration `mapstructure:"jwtAccessTTL"`
	JWTRefreshTTL     time.Duration `mapstructure:"jwtRefreshTTL"`
	GoogleOAuthClient string        `mapstructure:"googleOAuthClient"`
	AppleBundleID     string        `mapstructure:"appleBundleID"`
	CORSOrigins       string        `mapstructure:"corsOrigins"`
}

// DatabaseConfig stores PostgreSQL connection info.
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	URL      string // populated from DATABASE_URL env var
}

// RedisConfig stores Redis connection info.
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	URL      string // populated from REDIS_URL env var
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

	// Override port from PORT env var (Render sets this).
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.App.Port = p
		}
	}

	// Override app env from APP_ENV.
	if env := os.Getenv("APP_ENV"); env != "" {
		cfg.App.Env = env
	}

	// Override JWT secret from env.
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		cfg.App.JWTSecret = secret
	}

	if cfg.App.JWTSecret == "" {
		return nil, fmt.Errorf("jwt secret not configured")
	}

	// Override Google OAuth client from env.
	if gc := os.Getenv("GOOGLE_OAUTH_CLIENT_ID"); gc != "" {
		cfg.App.GoogleOAuthClient = gc
	}

	// Override Apple config from env.
	if ab := os.Getenv("APPLE_BUNDLE_ID"); ab != "" {
		cfg.App.AppleBundleID = ab
	}

	// CORS origins from env.
	if co := os.Getenv("CORS_ORIGINS"); co != "" {
		cfg.App.CORSOrigins = co
	}

	// DATABASE_URL takes precedence over individual fields.
	cfg.Database.URL = os.Getenv("DATABASE_URL")

	// REDIS_URL takes precedence over individual fields.
	cfg.Redis.URL = os.Getenv("REDIS_URL")

	accessTTL := 30 * time.Minute
	if cfg.App.JWTAccessTTL == 0 || cfg.App.JWTAccessTTL > accessTTL {
		cfg.App.JWTAccessTTL = accessTTL
	}
	if cfg.App.JWTRefreshTTL == 0 {
		cfg.App.JWTRefreshTTL = 24 * time.Hour
	}

	return &cfg, nil
}
