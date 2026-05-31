package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const developmentJWTSecret = "dev_secret_change_me"

type Config struct {
	AppEnv                string
	DatabaseURL           string
	JWTSecret             string
	Port                  string
	FrontendURL           string
	FrontendURLs          []string
	AuthRateLimitRequests int
	AuthRateLimitWindow   time.Duration
	MaxRequestBodyBytes   int64
	GitHubClientID        string
	GitHubClientSecret    string
	GitHubCallbackURL     string
}

func Load() (Config, error) {
	_ = godotenv.Overload(".env", "../.env", "../../.env")

	cfg := Config{
		AppEnv:                normalizeEnv(getEnv("APP_ENV", "development")),
		DatabaseURL:           getEnv("DATABASE_URL", "postgres://postgres:postgres@127.0.0.1:5433/travel_collab?sslmode=disable"),
		JWTSecret:             getEnv("JWT_SECRET", developmentJWTSecret),
		Port:                  getEnv("PORT", "8080"),
		FrontendURL:           normalizeOrigin(getEnv("FRONTEND_URL", "http://localhost:5173")),
		FrontendURLs:          parseCSV(getEnv("FRONTEND_URLS", "http://localhost:5173,capacitor://localhost,http://localhost,http://127.0.0.1:5173")),
		AuthRateLimitRequests: getEnvAsInt("AUTH_RATE_LIMIT_REQUESTS", 20),
		AuthRateLimitWindow:   getEnvAsDuration("AUTH_RATE_LIMIT_WINDOW", time.Minute),
		MaxRequestBodyBytes:   getEnvAsInt64("MAX_REQUEST_BODY_BYTES", 1<<20),
		GitHubClientID:        os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret:    os.Getenv("GITHUB_CLIENT_SECRET"),
		GitHubCallbackURL:     getEnv("GITHUB_CALLBACK_URL", "http://localhost:8080/api/v1/auth/github/callback"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.AuthRateLimitRequests < 1 {
		return Config{}, fmt.Errorf("AUTH_RATE_LIMIT_REQUESTS must be greater than 0")
	}
	if cfg.AuthRateLimitWindow <= 0 {
		return Config{}, fmt.Errorf("AUTH_RATE_LIMIT_WINDOW must be greater than 0")
	}
	if cfg.MaxRequestBodyBytes < 1024 {
		return Config{}, fmt.Errorf("MAX_REQUEST_BODY_BYTES must be at least 1024 bytes")
	}
	if cfg.IsProduction() {
		if isWeakJWTSecret(cfg.JWTSecret) {
			return Config{}, fmt.Errorf("JWT_SECRET is weak or uses a development value; set a long random secret before production start")
		}
		if containsWildcard(cfg.FrontendURL, cfg.FrontendURLs) {
			return Config{}, fmt.Errorf("wildcard frontend origin is not allowed in production")
		}
	} else if cfg.JWTSecret == "" || cfg.JWTSecret == developmentJWTSecret {
		fmt.Println("warning: JWT_SECRET uses a development value; change it before real deployment")
	}
	return cfg, nil
}

func (c Config) IsProduction() bool {
	return c.AppEnv == "production"
}

func (c Config) AllowLocalDevelopmentOrigins() bool {
	return !c.IsProduction()
}

func (c Config) IsOriginAllowed(origin string) bool {
	cleaned := normalizeOrigin(origin)
	if cleaned == "" {
		return false
	}
	if normalizeOrigin(c.FrontendURL) == cleaned {
		return true
	}
	for _, allowed := range c.FrontendURLs {
		if allowed == "*" && !c.IsProduction() {
			return true
		}
		if normalizeOrigin(allowed) == cleaned {
			return true
		}
	}
	return c.AllowLocalDevelopmentOrigins() && IsLocalDevelopmentOrigin(cleaned)
}

func IsLocalDevelopmentOrigin(origin string) bool {
	return strings.HasPrefix(origin, "capacitor://localhost") ||
		strings.HasPrefix(origin, "http://localhost") ||
		strings.HasPrefix(origin, "http://127.0.0.1") ||
		strings.HasPrefix(origin, "http://192.168.") ||
		strings.HasPrefix(origin, "http://10.") ||
		strings.HasPrefix(origin, "http://172.16.") ||
		strings.HasPrefix(origin, "http://172.17.") ||
		strings.HasPrefix(origin, "http://172.18.") ||
		strings.HasPrefix(origin, "http://172.19.") ||
		strings.HasPrefix(origin, "http://172.20.") ||
		strings.HasPrefix(origin, "http://172.21.") ||
		strings.HasPrefix(origin, "http://172.22.") ||
		strings.HasPrefix(origin, "http://172.23.") ||
		strings.HasPrefix(origin, "http://172.24.") ||
		strings.HasPrefix(origin, "http://172.25.") ||
		strings.HasPrefix(origin, "http://172.26.") ||
		strings.HasPrefix(origin, "http://172.27.") ||
		strings.HasPrefix(origin, "http://172.28.") ||
		strings.HasPrefix(origin, "http://172.29.") ||
		strings.HasPrefix(origin, "http://172.30.") ||
		strings.HasPrefix(origin, "http://172.31.")
}

func parseCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		cleaned := normalizeOrigin(part)
		if cleaned != "" {
			result = append(result, cleaned)
		}
	}
	return result
}

func normalizeOrigin(value string) string {
	return strings.TrimRight(strings.TrimSpace(value), "/")
}

func normalizeEnv(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "prod":
		return "production"
	case "production", "development", "test":
		return strings.ToLower(strings.TrimSpace(value))
	case "":
		return "development"
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvAsInt64(key string, fallback int64) int64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func isWeakJWTSecret(secret string) bool {
	trimmed := strings.TrimSpace(secret)
	lowered := strings.ToLower(trimmed)
	return len(trimmed) < 32 ||
		trimmed == developmentJWTSecret ||
		strings.Contains(lowered, "change_me") ||
		strings.Contains(lowered, "replace") ||
		strings.Contains(lowered, "generate") ||
		strings.Contains(lowered, "dev_secret") ||
		strings.Contains(lowered, "secret") && len(trimmed) < 48
}

func containsWildcard(frontendURL string, frontendURLs []string) bool {
	if normalizeOrigin(frontendURL) == "*" {
		return true
	}
	for _, origin := range frontendURLs {
		if normalizeOrigin(origin) == "*" {
			return true
		}
	}
	return false
}
