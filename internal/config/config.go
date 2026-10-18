package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// defaultJWTSecret is a development-only placeholder; Validate rejects it in release mode.
const defaultJWTSecret = "your-super-secret-key-change-this-in-production"

// Config holds all application configuration
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	JWT       JWTConfig
	RateLimit RateLimitConfig
	CORS      CORSConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host         string
	Port         string
	Mode         string // "debug", "release", "test"
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	// CookieSecure sets the Secure flag on auth cookies (HTTPS only).
	CookieSecure bool
	// TrustedProxies lists proxy IPs/CIDRs allowed to set X-Forwarded-For.
	// Empty means the client IP is always the TCP peer address.
	TrustedProxies []string
}

// CORSConfig holds cross-origin settings
type CORSConfig struct {
	AllowedOrigins []string
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Path            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	SecretKey       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled         bool
	RequestsPerSec  float64
	Burst           int
	CleanupInterval time.Duration
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using default values or system environment variables")
	}

	config := &Config{
		Server:    loadServerConfig(),
		Database:  loadDatabaseConfig(),
		JWT:       loadJWTConfig(),
		RateLimit: loadRateLimitConfig(),
		CORS:      loadCORSConfig(),
	}

	// Validate required fields
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// loadServerConfig loads server configuration
func loadServerConfig() ServerConfig {
	mode := getEnv("GIN_MODE", "debug")
	return ServerConfig{
		Host:           getEnv("SERVER_HOST", "0.0.0.0"),
		Port:           getEnv("SERVER_PORT", "8080"),
		Mode:           mode,
		ReadTimeout:    getDurationEnv("SERVER_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:   getDurationEnv("SERVER_WRITE_TIMEOUT", 10*time.Second),
		IdleTimeout:    getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
		CookieSecure:   getBoolEnv("COOKIE_SECURE", mode == "release"),
		TrustedProxies: getListEnv("TRUSTED_PROXIES", nil),
	}
}

// loadCORSConfig loads cross-origin configuration
func loadCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: getListEnv("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:5173"}),
	}
}

// loadDatabaseConfig loads database configuration
func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Path:            getEnv("DATABASE_PATH", "./data/marketplace.db"),
		MaxOpenConns:    getIntEnv("DATABASE_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getIntEnv("DATABASE_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: getDurationEnv("DATABASE_MAX_LIFETIME", 5*time.Minute),
	}
}

// loadJWTConfig loads JWT configuration
func loadJWTConfig() JWTConfig {
	return JWTConfig{
		SecretKey:       getEnv("JWT_SECRET_KEY", defaultJWTSecret),
		AccessTokenTTL:  getDurationEnv("JWT_ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL: getDurationEnv("JWT_REFRESH_TOKEN_TTL", 7*24*time.Hour), // 7 days
	}
}

// loadRateLimitConfig loads rate limiting configuration
func loadRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Enabled:         getBoolEnv("RATE_LIMIT_ENABLED", true),
		RequestsPerSec:  getFloat64Env("RATE_LIMIT_RPS", 10.0),
		Burst:           getIntEnv("RATE_LIMIT_BURST", 20),
		CleanupInterval: getDurationEnv("RATE_LIMIT_CLEANUP", 5*time.Minute),
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	if c.Database.Path == "" {
		return fmt.Errorf("database path is required")
	}

	if c.JWT.SecretKey == "" {
		return fmt.Errorf("JWT secret key is required")
	}

	if len(c.JWT.SecretKey) < 32 {
		return fmt.Errorf("JWT secret key must be at least 32 characters long")
	}

	if c.Server.Mode == "release" && c.JWT.SecretKey == defaultJWTSecret {
		return fmt.Errorf("JWT_SECRET_KEY must be set to a real secret in release mode")
	}

	return nil
}

// GetServerAddress returns the full server address
func (c *Config) GetServerAddress() string {
	return c.Server.Host + ":" + c.Server.Port
}

// Helper functions

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getIntEnv gets an integer environment variable or returns a default value
func getIntEnv(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// getFloat64Env gets a float64 environment variable or returns a default value
func getFloat64Env(key string, defaultValue float64) float64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return defaultValue
	}

	return value
}

// getBoolEnv gets a boolean environment variable or returns a default value
func getBoolEnv(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// getDurationEnv gets a duration environment variable or returns a default value
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// getListEnv gets a comma-separated environment variable or returns a default value
func getListEnv(key string, defaultValue []string) []string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	var values []string
	for _, v := range strings.Split(valueStr, ",") {
		if v = strings.TrimSpace(v); v != "" {
			values = append(values, v)
		}
	}

	return values
}
