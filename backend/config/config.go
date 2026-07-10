// Package config provides application configuration loading from environment variables
// with fail-fast validation. All configuration is loaded once at startup and validated
// before the application begins serving requests.
//
// Configuration is organized into logical groups (Server, Database, AI, Auth, Log)
// with sensible defaults for development. Production environments must explicitly
// provide required values (DATABASE_URL, JWT_SIGNING_KEY, and ANTHROPIC_API_KEY when
// AI_PROVIDER=anthropic).
//
// Usage:
//
//	cfg, err := config.Load()
//	if err != nil {
//	    log.Fatal("Failed to load configuration: %w", err)
//	}
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration loaded from environment variables.
// All fields are loaded at startup with fail-fast validation.
type Config struct {
	Log      LogConfig
	Server   ServerConfig
	Database DatabaseConfig
	AI       AIConfig
	Auth     AuthConfig
}

// ServerConfig contains HTTP server settings.
type ServerConfig struct {
	AllowedCORS  string        // ALLOWED_CORS_ORIGINS (default: http://localhost:5173)
	ReadTimeout  time.Duration // HTTP_READ_TIMEOUT (default: 30s)
	WriteTimeout time.Duration // HTTP_WRITE_TIMEOUT (default: 30s)
	IdleTimeout  time.Duration // HTTP_IDLE_TIMEOUT (default: 120s)
	Port         int           // HTTP_PORT (default: 8080)
}

// DatabaseConfig contains PostgreSQL connection settings.
type DatabaseConfig struct {
	URL            string // DATABASE_URL (required)
	MaxConnections int    // DB_MAX_CONNECTIONS (default: 25)
	MinConnections int    // DB_MIN_CONNECTIONS (default: 5)
	MaxIdleTime    time.Duration
	MaxLifetime    time.Duration
}

// AIConfig contains AI provider settings shared across both supported
// backends (Ollama, Anthropic), plus each backend's own settings.
type AIConfig struct {
	Provider       string        // AI_PROVIDER (default: "ollama" in development, "anthropic" in production)
	Timeout        time.Duration // AI_TIMEOUT (default: 60s)
	MaxRetries     int           // AI_MAX_RETRIES (default: 3)
	StreamingChunk int           // AI_STREAMING_CHUNK_SIZE (default: 4096)
	Anthropic      AnthropicConfig
	Ollama         OllamaConfig
}

// AnthropicConfig contains settings for the Anthropic Claude backend, used
// in staging/production (AI_PROVIDER=anthropic).
type AnthropicConfig struct {
	APIKey string // ANTHROPIC_API_KEY (required when AI_PROVIDER=anthropic)
	Model  string // ANTHROPIC_MODEL (default: claude-3-5-sonnet-20241022)
}

// OllamaConfig contains settings for the local Ollama backend, used for
// local development and MVP testing (AI_PROVIDER=ollama, the default).
type OllamaConfig struct {
	Host  string // OLLAMA_HOST (default: http://localhost:11434)
	Model string // OLLAMA_MODEL (default: gemma3:4b)
}

// AuthConfig contains JWT and session settings.
type AuthConfig struct {
	JWTSigningKey     string        // JWT_SIGNING_KEY (required in production)
	CookieDomain      string        // COOKIE_DOMAIN (default: localhost)
	JWTExpiration     time.Duration // JWT_EXPIRATION (default: 24h)
	RefreshExpiration time.Duration // REFRESH_TOKEN_EXPIRATION (default: 7 days)
	BcryptCost        int           // BCRYPT_COST (default: 12)
	CookieSecure      bool          // COOKIE_SECURE (default: false, true in production)
}

// LogConfig contains structured logging settings.
type LogConfig struct {
	Level  string // LOG_LEVEL (default: info)
	Format string // LOG_FORMAT (default: json)
}

// Load reads environment variables and returns a validated Config.
// Returns an error if required variables are missing or invalid.
func Load() (*Config, error) {
	cfg := &Config{
		Server:   loadServerConfig(),
		Database: loadDatabaseConfig(),
		AI:       loadAIConfig(),
		Auth:     loadAuthConfig(),
		Log:      loadLogConfig(),
	}

	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// loadServerConfig reads HTTP server configuration from environment variables.
func loadServerConfig() ServerConfig {
	return ServerConfig{
		Port:         getEnvInt("HTTP_PORT", 8080),
		ReadTimeout:  getEnvDuration("HTTP_READ_TIMEOUT", 30*time.Second),
		WriteTimeout: getEnvDuration("HTTP_WRITE_TIMEOUT", 30*time.Second),
		IdleTimeout:  getEnvDuration("HTTP_IDLE_TIMEOUT", 120*time.Second),
		AllowedCORS:  getEnv("ALLOWED_CORS_ORIGINS", "http://localhost:5173"),
	}
}

// loadDatabaseConfig reads PostgreSQL database configuration from environment variables.
func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		URL:            getEnv("DATABASE_URL", ""),
		MaxConnections: getEnvInt("DB_MAX_CONNECTIONS", 25),
		MinConnections: getEnvInt("DB_MIN_CONNECTIONS", 5),
		MaxIdleTime:    getEnvDuration("DB_MAX_IDLE_TIME", 15*time.Minute),
		MaxLifetime:    getEnvDuration("DB_MAX_LIFETIME", 1*time.Hour),
	}
}

// loadAIConfig reads AI provider configuration from environment variables.
// The default provider is "ollama" everywhere except production (mirrors
// the JWT_SIGNING_KEY-required-only-in-production pattern below): local
// development and MVP testing run against a free local Ollama server, while
// staging/production use Anthropic Claude.
func loadAIConfig() AIConfig {
	defaultProvider := "ollama"
	if getEnv("GO_ENV", "development") == "production" {
		defaultProvider = "anthropic"
	}

	return AIConfig{
		Provider:       getEnv("AI_PROVIDER", defaultProvider),
		Timeout:        getEnvDuration("AI_TIMEOUT", 60*time.Second),
		MaxRetries:     getEnvInt("AI_MAX_RETRIES", 3),
		StreamingChunk: getEnvInt("AI_STREAMING_CHUNK_SIZE", 4096),
		Anthropic: AnthropicConfig{
			APIKey: getEnv("ANTHROPIC_API_KEY", ""),
			Model:  getEnv("ANTHROPIC_MODEL", "claude-3-5-sonnet-20241022"),
		},
		Ollama: OllamaConfig{
			Host:  getEnv("OLLAMA_HOST", "http://localhost:11434"),
			Model: getEnv("OLLAMA_MODEL", "gemma3:4b"),
		},
	}
}

// loadAuthConfig reads authentication and session configuration from environment variables.
func loadAuthConfig() AuthConfig {
	return AuthConfig{
		JWTSigningKey:     getEnv("JWT_SIGNING_KEY", ""),
		JWTExpiration:     getEnvDuration("JWT_EXPIRATION", 24*time.Hour),
		RefreshExpiration: getEnvDuration("REFRESH_TOKEN_EXPIRATION", 7*24*time.Hour),
		CookieDomain:      getEnv("COOKIE_DOMAIN", "localhost"),
		CookieSecure:      getEnvBool("COOKIE_SECURE", false),
		BcryptCost:        getEnvInt("BCRYPT_COST", 12),
	}
}

// loadLogConfig reads structured logging configuration from environment variables.
func loadLogConfig() LogConfig {
	return LogConfig{
		Level:  getEnv("LOG_LEVEL", "info"),
		Format: getEnv("LOG_FORMAT", "json"),
	}
}

// validate performs fail-fast validation of the loaded configuration.
// It checks that:
//   - Required environment variables are present (DATABASE_URL)
//   - AI_PROVIDER is "ollama" or "anthropic"; ANTHROPIC_API_KEY is required only
//     when AI_PROVIDER=anthropic (Ollama has no required secret — Host and Model
//     both have defaults)
//   - JWT_SIGNING_KEY is set in production (not required in development for local testing)
//   - Database connection pool limits are logical (min ≤ max)
//   - HTTP port is within valid range (1-65535)
//   - Log level is a recognized value (debug, info, warn, error)
//
// Returns an error if any validation rule fails.
func validate(cfg *Config) error {
	if cfg.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	switch cfg.AI.Provider {
	case "anthropic":
		if cfg.AI.Anthropic.APIKey == "" {
			return fmt.Errorf("ANTHROPIC_API_KEY is required when AI_PROVIDER=anthropic")
		}
	case "ollama":
		// No required secret; Host and Model both have defaults.
	default:
		return fmt.Errorf("AI_PROVIDER must be 'ollama' or 'anthropic', got %q", cfg.AI.Provider)
	}

	if cfg.Auth.JWTSigningKey == "" && getEnv("GO_ENV", "development") == "production" {
		return fmt.Errorf("JWT_SIGNING_KEY is required in production")
	}

	if cfg.Database.MinConnections > cfg.Database.MaxConnections {
		return fmt.Errorf("DB_MIN_CONNECTIONS (%d) cannot exceed DB_MAX_CONNECTIONS (%d)",
			cfg.Database.MinConnections, cfg.Database.MaxConnections)
	}

	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("HTTP_PORT must be between 1 and 65535, got %d", cfg.Server.Port)
	}

	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[cfg.Log.Level] {
		return fmt.Errorf("LOG_LEVEL must be one of: debug, info, warn, error; got %s", cfg.Log.Level)
	}

	return nil
}

// Helper functions for environment variable parsing

// getEnv retrieves a string environment variable or returns the default if unset or empty.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt retrieves an integer environment variable or returns the default if unset, empty, or invalid.
// Invalid values (non-numeric strings) are silently ignored and the default is used.
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBool retrieves a boolean environment variable or returns the default if unset, empty, or invalid.
// Valid boolean strings: "1", "t", "T", "true", "True", "TRUE", "0", "f", "F", "false", "False", "FALSE".
// Invalid values are silently ignored and the default is used.
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvDuration retrieves a duration environment variable or returns the default if unset, empty, or invalid.
// Durations must be in Go duration format (e.g., "30s", "5m", "1h30m").
// Invalid values are silently ignored and the default is used.
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
