// Package config loads all environment-backed settings for the service.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

const defaultJWTSecret = "gigmatch-dev-secret-change-me"

// knownWeakSecrets are example/placeholder values that must never be used in
// production.
var knownWeakSecrets = []string{
	defaultJWTSecret,
	"please-change-me-to-a-long-random-string",
	"change-me",
	"changeme",
	"secret",
	"jwt-secret",
}

// MySQLConfig holds database connection settings.
type MySQLConfig struct {
	Host     string `env:"DB_HOST" envDefault:"127.0.0.1"`
	Port     int    `env:"DB_PORT" envDefault:"33301"`
	User     string `env:"DB_USER" envDefault:"gigmatch"`
	Password string `env:"DB_PASSWORD" envDefault:"gigmatch123"`
	Database string `env:"DB_NAME" envDefault:"gigmatch"`
}

// Config aggregates every environment-backed setting.
type Config struct {
	Env        string `env:"APP_ENV" envDefault:"development"`
	ServerPort int    `env:"SERVER_PORT" envDefault:"8080"`
	JWTSecret  string `env:"JWT_SECRET" envDefault:"gigmatch-dev-secret-change-me"`
	JWTExpire  int    `env:"JWT_EXPIRE_HOURS" envDefault:"72"`
	UploadDir  string `env:"UPLOAD_DIR" envDefault:"/tmp/gigmatch-uploads"`

	// CORS_ALLOWED_ORIGINS is a comma-separated origin list.
	CORSAllowedOrigins string `env:"CORS_ALLOWED_ORIGINS" envDefault:"http://localhost:28030"`

	// Rate limits are applied per client IP.
	AuthRateLimit int `env:"AUTH_RATE_LIMIT" envDefault:"10"`
	APIRateLimit  int `env:"API_RATE_LIMIT" envDefault:"120"`

	// Connection pool tuning.
	DBMaxOpenConns            int `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
	DBMaxIdleConns            int `env:"DB_MAX_IDLE_CONNS" envDefault:"5"`
	DBConnMaxLifetimeMin      int `env:"DB_CONN_MAX_LIFETIME_MIN" envDefault:"5"`
	DBConnectRetries          int `env:"DB_CONNECT_RETRIES" envDefault:"10"`
	DBConnectRetryIntervalSec int `env:"DB_CONNECT_RETRY_INTERVAL_SEC" envDefault:"3"`

	MySQL MySQLConfig
}

// Load reads configuration from environment variables and validates it.
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.ServerPort < 1 || c.ServerPort > 65535 {
		return fmt.Errorf("SERVER_PORT must be between 1 and 65535")
	}
	if c.JWTExpire <= 0 {
		return fmt.Errorf("JWT_EXPIRE_HOURS must be greater than 0")
	}
	if c.AuthRateLimit <= 0 || c.APIRateLimit <= 0 {
		return fmt.Errorf("AUTH_RATE_LIMIT and API_RATE_LIMIT must be greater than 0")
	}
	if c.DBMaxOpenConns <= 0 || c.DBMaxIdleConns < 0 {
		return fmt.Errorf("DB_MAX_OPEN_CONNS must be > 0 and DB_MAX_IDLE_CONNS must be >= 0")
	}
	if c.DBConnectRetries < 0 || c.DBConnectRetryIntervalSec < 0 {
		return fmt.Errorf("DB_CONNECT_RETRIES and DB_CONNECT_RETRY_INTERVAL_SEC must be >= 0")
	}
	if c.Env == "production" {
		if c.JWTSecret == "" || len(c.JWTSecret) < 32 || isWeakSecret(c.JWTSecret) {
			return fmt.Errorf("JWT_SECRET must be replaced with a value of at least 32 characters in production")
		}
		if strings.TrimSpace(c.CORSAllowedOrigins) == "*" {
			return fmt.Errorf("CORS_ALLOWED_ORIGINS must not be '*' in production")
		}
	}
	return nil
}

// AllowedOrigins parses CORS_ALLOWED_ORIGINS. Empty or '*' disables CORS.
func (c *Config) AllowedOrigins() []string {
	raw := strings.TrimSpace(c.CORSAllowedOrigins)
	if raw == "" || raw == "*" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// DSN builds the MySQL connection string.
func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.MySQL.User, c.MySQL.Password, c.MySQL.Host, c.MySQL.Port, c.MySQL.Database)
}

// JWTExpireDuration returns the access token lifetime.
func (c *Config) JWTExpireDuration() time.Duration {
	return time.Duration(c.JWTExpire) * time.Hour
}

func isWeakSecret(v string) bool {
	for _, weak := range knownWeakSecrets {
		if strings.EqualFold(v, weak) {
			return true
		}
	}
	return false
}
