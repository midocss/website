package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds every runtime setting of the monolith. Values are read from the
// environment so the same binary can run in any environment without rebuild.
type Config struct {
	App      App
	HTTP     HTTP
	Database Database
	JWT      JWT
	Storage  Storage
	Payment  Payment
}

type App struct {
	Name        string
	Environment string
	Debug       bool
	DefaultLang string
	BaseURL     string
}

type HTTP struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	TrustedProxies  []string
	AllowedOrigins  []string
}

type Database struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type JWT struct {
	Secret          string
	Issuer          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type Storage struct {
	Endpoint        string
	AccessKey       string
	SecretKey       string
	UseSSL          bool
	Region          string
	PublicBucket    string
	PrivateBucket   string
	SignedURLExpiry time.Duration
}

type Payment struct {
	ZainCash ZainCash
}

type ZainCash struct {
	MerchantID string
	MSISDN     string
	Secret     string
	BaseURL    string
	RedirectURL
}

type RedirectURL struct {
	SuccessURL string
	FailureURL string
}

// DSN builds the PostgreSQL connection string.
func (d Database) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

// Load reads the configuration from the environment and validates it.
func Load() (*Config, error) {
	cfg := &Config{
		App: App{
			Name:        env("APP_NAME", "midocss-platform"),
			Environment: env("APP_ENV", "development"),
			Debug:       envBool("APP_DEBUG", true),
			DefaultLang: env("APP_DEFAULT_LANG", "ar"),
			BaseURL:     env("APP_BASE_URL", "http://localhost:8080"),
		},
		HTTP: HTTP{
			Host:            env("HTTP_HOST", "0.0.0.0"),
			Port:            envInt("HTTP_PORT", 8080),
			ReadTimeout:     envDuration("HTTP_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    envDuration("HTTP_WRITE_TIMEOUT", 30*time.Second),
			ShutdownTimeout: envDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),
			TrustedProxies:  envList("HTTP_TRUSTED_PROXIES", nil),
			AllowedOrigins:  envList("HTTP_ALLOWED_ORIGINS", []string{"*"}),
		},
		Database: Database{
			Host:            env("DB_HOST", "127.0.0.1"),
			Port:            envInt("DB_PORT", 5432),
			User:            env("DB_USER", "postgres"),
			Password:        env("DB_PASSWORD", "postgres"),
			Name:            env("DB_NAME", "midocss"),
			SSLMode:         env("DB_SSLMODE", "disable"),
			MaxOpenConns:    envInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    envInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: envDuration("DB_CONN_MAX_LIFETIME", time.Hour),
		},
		JWT: JWT{
			Secret:          env("JWT_SECRET", ""),
			Issuer:          env("JWT_ISSUER", "midocss-platform"),
			AccessTokenTTL:  envDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTokenTTL: envDuration("JWT_REFRESH_TTL", 720*time.Hour),
		},
		Storage: Storage{
			Endpoint:        env("MINIO_ENDPOINT", "127.0.0.1:9000"),
			AccessKey:       env("MINIO_ACCESS_KEY", ""),
			SecretKey:       env("MINIO_SECRET_KEY", ""),
			UseSSL:          envBool("MINIO_USE_SSL", false),
			Region:          env("MINIO_REGION", "us-east-1"),
			PublicBucket:    env("MINIO_PUBLIC_BUCKET", "public"),
			PrivateBucket:   env("MINIO_PRIVATE_BUCKET", "private"),
			SignedURLExpiry: envDuration("MINIO_SIGNED_URL_EXPIRY", 15*time.Minute),
		},
		Payment: Payment{
			ZainCash: ZainCash{
				MerchantID: env("ZAINCASH_MERCHANT_ID", ""),
				MSISDN:     env("ZAINCASH_MSISDN", ""),
				Secret:     env("ZAINCASH_SECRET", ""),
				BaseURL:    env("ZAINCASH_BASE_URL", "https://test.zaincash.iq"),
				RedirectURL: RedirectURL{
					SuccessURL: env("ZAINCASH_SUCCESS_URL", ""),
					FailureURL: env("ZAINCASH_FAILURE_URL", ""),
				},
			},
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) IsProduction() bool {
	return strings.EqualFold(c.App.Environment, "production")
}

func (c *Config) validate() error {
	if c.JWT.Secret == "" {
		return fmt.Errorf("config: JWT_SECRET is required")
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("config: JWT_SECRET must be at least 32 characters")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("config: DB_NAME is required")
	}
	if c.IsProduction() && c.App.Debug {
		return fmt.Errorf("config: APP_DEBUG must be false in production")
	}
	return nil
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		if parsed, err := strconv.ParseBool(v); err == nil {
			return parsed
		}
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if parsed, err := time.ParseDuration(v); err == nil {
			return parsed
		}
	}
	return fallback
}

func envList(key string, fallback []string) []string {
	v, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(v) == "" {
		return fallback
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
