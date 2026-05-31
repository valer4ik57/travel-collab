package api

import (
	"net/http/httptest"
	"testing"

	"travel-collab/backend/internal/config"
)

func TestWebSocketOriginCheckAllowsConfiguredProductionOrigin(t *testing.T) {
	cfg := config.Config{
		AppEnv:       "production",
		FrontendURL:  "https://travel.example.ru",
		FrontendURLs: []string{"https://travel.example.ru"},
	}
	req := httptest.NewRequest("GET", "/api/v1/ws/trip-1", nil)
	req.Header.Set("Origin", "https://travel.example.ru")

	if !isWebSocketOriginAllowed(cfg, req) {
		t.Fatal("expected configured production origin to be allowed")
	}
}

func TestWebSocketOriginCheckRejectsUnknownProductionOrigin(t *testing.T) {
	cfg := config.Config{
		AppEnv:       "production",
		FrontendURL:  "https://travel.example.ru",
		FrontendURLs: []string{"https://travel.example.ru"},
	}
	req := httptest.NewRequest("GET", "/api/v1/ws/trip-1", nil)
	req.Header.Set("Origin", "https://evil.example.ru")

	if isWebSocketOriginAllowed(cfg, req) {
		t.Fatal("expected unknown production origin to be rejected")
	}
}

func TestWebSocketOriginCheckKeepsLocalDevelopmentConvenience(t *testing.T) {
	cfg := config.Config{AppEnv: "development", FrontendURL: "http://localhost:5173"}
	req := httptest.NewRequest("GET", "/api/v1/ws/trip-1", nil)
	req.Header.Set("Origin", "http://192.168.1.20:5173")

	if !isWebSocketOriginAllowed(cfg, req) {
		t.Fatal("expected local development network origin to be allowed")
	}
}
