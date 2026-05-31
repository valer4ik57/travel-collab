package config

import "testing"

func TestNormalizeEnvAliases(t *testing.T) {
	cases := map[string]string{
		"":             "development",
		"prod":         "production",
		"PRODUCTION":   "production",
		"development":  "development",
		"test":         "test",
		"custom-stage": "custom-stage",
	}
	for input, expected := range cases {
		if got := normalizeEnv(input); got != expected {
			t.Fatalf("normalizeEnv(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestIsOriginAllowedInProductionUsesOnlyConfiguredOrigins(t *testing.T) {
	cfg := Config{
		AppEnv:       "production",
		FrontendURL:  "https://travel.example.ru",
		FrontendURLs: []string{"https://travel.example.ru", "https://www.travel.example.ru"},
	}

	allowed := []string{
		"https://travel.example.ru",
		"https://travel.example.ru/",
		"https://www.travel.example.ru",
	}
	for _, origin := range allowed {
		if !cfg.IsOriginAllowed(origin) {
			t.Fatalf("expected origin %q to be allowed", origin)
		}
	}

	denied := []string{
		"http://localhost:5173",
		"http://127.0.0.1:5173",
		"http://192.168.1.10:5173",
		"https://evil.example.ru",
		"",
	}
	for _, origin := range denied {
		if cfg.IsOriginAllowed(origin) {
			t.Fatalf("expected origin %q to be denied in production", origin)
		}
	}
}

func TestIsOriginAllowedInDevelopmentAllowsLocalNetwork(t *testing.T) {
	cfg := Config{AppEnv: "development", FrontendURL: "http://localhost:5173"}

	allowed := []string{
		"http://localhost:5173",
		"http://127.0.0.1:5173",
		"http://192.168.0.25:5173",
		"http://10.0.0.15:5173",
		"http://172.20.0.7:5173",
		"capacitor://localhost",
	}
	for _, origin := range allowed {
		if !cfg.IsOriginAllowed(origin) {
			t.Fatalf("expected development origin %q to be allowed", origin)
		}
	}
}

func TestWeakJWTSecretDetection(t *testing.T) {
	weak := []string{
		"",
		developmentJWTSecret,
		"short-secret",
		"replace_this_secret_before_deploy",
		"generate-super-secret",
		"secret-but-too-short",
	}
	for _, secret := range weak {
		if !isWeakJWTSecret(secret) {
			t.Fatalf("expected secret %q to be weak", secret)
		}
	}

	strong := "9QkzG7bfKtm2wRrD6nLpXs4yV8hCaE3uMvP0sZdN1aFb"
	if isWeakJWTSecret(strong) {
		t.Fatalf("expected long random secret to be accepted")
	}
}

func TestParseCSVNormalizesOrigins(t *testing.T) {
	got := parseCSV(" https://travel.example.ru/ , , http://localhost:5173/ ")
	want := []string{"https://travel.example.ru", "http://localhost:5173"}
	if len(got) != len(want) {
		t.Fatalf("expected %d origins, got %d: %#v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("origin[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
