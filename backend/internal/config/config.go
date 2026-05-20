package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL        string
	JWTSecret          string
	Port               string
	FrontendURL        string
	GitHubClientID     string
	GitHubClientSecret string
	GitHubCallbackURL  string
}

func Load() (Config, error) {
	_ = godotenv.Load(".env", "../.env", "../../.env")

	cfg := Config{
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/travel_collab?sslmode=disable"),
		JWTSecret:          getEnv("JWT_SECRET", "dev_secret_change_me"),
		Port:               getEnv("PORT", "8080"),
		FrontendURL:        strings.TrimRight(getEnv("FRONTEND_URL", "http://localhost:5173"), "/"),
		GitHubClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		GitHubCallbackURL:  getEnv("GITHUB_CALLBACK_URL", "http://localhost:8080/api/v1/auth/github/callback"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" || cfg.JWTSecret == "dev_secret_change_me" {
		fmt.Println("warning: JWT_SECRET uses a development value; change it before real deployment")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
