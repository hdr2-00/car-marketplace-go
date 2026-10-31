package config

import (
	"strings"
	"testing"
)

func validConfig() *Config {
	return &Config{
		Server:   ServerConfig{Port: "8080", Mode: "release"},
		Database: DatabaseConfig{Path: "x.db"},
		JWT:      JWTConfig{SecretKey: strings.Repeat("k", 40)},
	}
}

func TestValidateRejectsDefaultSecretInRelease(t *testing.T) {
	cfg := validConfig()
	cfg.JWT.SecretKey = defaultJWTSecret

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected release mode to reject the placeholder JWT secret")
	}
}

func TestValidateAllowsDefaultSecretInDebug(t *testing.T) {
	cfg := validConfig()
	cfg.Server.Mode = "debug"
	cfg.JWT.SecretKey = defaultJWTSecret

	if err := cfg.Validate(); err != nil {
		t.Fatalf("debug mode should allow the placeholder secret: %v", err)
	}
}

func TestLoadReadsListSettings(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", strings.Repeat("k", 40))
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://a.example, https://b.example")
	t.Setenv("TRUSTED_PROXIES", "10.0.0.1")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.CORS.AllowedOrigins; len(got) != 2 || got[1] != "https://b.example" {
		t.Errorf("AllowedOrigins = %q", got)
	}
	if got := cfg.Server.TrustedProxies; len(got) != 1 || got[0] != "10.0.0.1" {
		t.Errorf("TrustedProxies = %q", got)
	}
}

func TestCookieSecureDefaultsOnInRelease(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", strings.Repeat("k", 40))
	t.Setenv("GIN_MODE", "release")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Server.CookieSecure {
		t.Error("expected secure cookies by default in release mode")
	}
}
