package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad_RequiredVariables(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		wantErr     bool
		errContains string
	}{
		{
			name: "missing DATABASE_URL",
			setup: func() {
				os.Clearenv()
				os.Setenv("ANTHROPIC_API_KEY", "test-key")
			},
			wantErr:     true,
			errContains: "DATABASE_URL is required",
		},
		{
			name: "missing ANTHROPIC_API_KEY",
			setup: func() {
				os.Clearenv()
				os.Setenv("DATABASE_URL", "postgres://localhost/test")
			},
			wantErr:     true,
			errContains: "ANTHROPIC_API_KEY is required",
		},
		{
			name: "missing JWT_SIGNING_KEY in production",
			setup: func() {
				os.Clearenv()
				os.Setenv("DATABASE_URL", "postgres://localhost/test")
				os.Setenv("ANTHROPIC_API_KEY", "test-key")
				os.Setenv("GO_ENV", "production")
			},
			wantErr:     true,
			errContains: "JWT_SIGNING_KEY is required in production",
		},
		{
			name: "all required variables present",
			setup: func() {
				os.Clearenv()
				os.Setenv("DATABASE_URL", "postgres://localhost/test")
				os.Setenv("ANTHROPIC_API_KEY", "test-key")
				os.Setenv("JWT_SIGNING_KEY", "test-signing-key")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			cfg, err := Load()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Load() expected error containing %q, got nil", tt.errContains)
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Load() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Load() unexpected error = %v", err)
					return
				}
				if cfg == nil {
					t.Error("Load() returned nil config")
				}
			}
		})
	}
}

func TestLoad_Validation(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		wantErr     bool
		errContains string
	}{
		{
			name: "invalid port - too high",
			setup: func() {
				setValidEnv()
				os.Setenv("HTTP_PORT", "99999")
			},
			wantErr:     true,
			errContains: "HTTP_PORT must be between 1 and 65535",
		},
		{
			name: "invalid port - zero",
			setup: func() {
				setValidEnv()
				os.Setenv("HTTP_PORT", "0")
			},
			wantErr:     true,
			errContains: "HTTP_PORT must be between 1 and 65535",
		},
		{
			name: "invalid log level",
			setup: func() {
				setValidEnv()
				os.Setenv("LOG_LEVEL", "verbose")
			},
			wantErr:     true,
			errContains: "LOG_LEVEL must be one of: debug, info, warn, error",
		},
		{
			name: "min connections exceeds max connections",
			setup: func() {
				setValidEnv()
				os.Setenv("DB_MIN_CONNECTIONS", "30")
				os.Setenv("DB_MAX_CONNECTIONS", "20")
			},
			wantErr:     true,
			errContains: "DB_MIN_CONNECTIONS (30) cannot exceed DB_MAX_CONNECTIONS (20)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := Load()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Load() expected error containing %q, got nil", tt.errContains)
					return
				}
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("Load() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Load() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestLoad_Defaults(t *testing.T) {
	setValidEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error = %v", err)
	}

	tests := []struct {
		name     string
		got      interface{}
		want     interface{}
		category string
	}{
		// Server defaults
		{"HTTP_PORT", cfg.Server.Port, 8080, "Server"},
		{"HTTP_READ_TIMEOUT", cfg.Server.ReadTimeout, 30 * time.Second, "Server"},
		{"HTTP_WRITE_TIMEOUT", cfg.Server.WriteTimeout, 30 * time.Second, "Server"},
		{"HTTP_IDLE_TIMEOUT", cfg.Server.IdleTimeout, 120 * time.Second, "Server"},
		{"ALLOWED_CORS_ORIGINS", cfg.Server.AllowedCORS, "http://localhost:5173", "Server"},

		// Database defaults
		{"DB_MAX_CONNECTIONS", cfg.Database.MaxConnections, 25, "Database"},
		{"DB_MIN_CONNECTIONS", cfg.Database.MinConnections, 5, "Database"},
		{"DB_MAX_IDLE_TIME", cfg.Database.MaxIdleTime, 15 * time.Minute, "Database"},
		{"DB_MAX_LIFETIME", cfg.Database.MaxLifetime, 1 * time.Hour, "Database"},

		// AI defaults
		{"AI_MODEL", cfg.AI.Model, "claude-3-5-sonnet-20241022", "AI"},
		{"AI_TIMEOUT", cfg.AI.Timeout, 60 * time.Second, "AI"},
		{"AI_MAX_RETRIES", cfg.AI.MaxRetries, 3, "AI"},
		{"AI_STREAMING_CHUNK_SIZE", cfg.AI.StreamingChunk, 4096, "AI"},

		// Auth defaults
		{"JWT_EXPIRATION", cfg.Auth.JWTExpiration, 24 * time.Hour, "Auth"},
		{"REFRESH_TOKEN_EXPIRATION", cfg.Auth.RefreshExpiration, 7 * 24 * time.Hour, "Auth"},
		{"COOKIE_DOMAIN", cfg.Auth.CookieDomain, "localhost", "Auth"},
		{"COOKIE_SECURE", cfg.Auth.CookieSecure, false, "Auth"},
		{"BCRYPT_COST", cfg.Auth.BcryptCost, 12, "Auth"},

		// Log defaults
		{"LOG_LEVEL", cfg.Log.Level, "info", "Log"},
		{"LOG_FORMAT", cfg.Log.Format, "json", "Log"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s default: got %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestLoad_CustomValues(t *testing.T) {
	setValidEnv()
	os.Setenv("HTTP_PORT", "3000")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("DB_MAX_CONNECTIONS", "50")
	os.Setenv("AI_TIMEOUT", "120s")
	os.Setenv("BCRYPT_COST", "14")
	os.Setenv("COOKIE_SECURE", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error = %v", err)
	}

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"HTTP_PORT", cfg.Server.Port, 3000},
		{"LOG_LEVEL", cfg.Log.Level, "debug"},
		{"DB_MAX_CONNECTIONS", cfg.Database.MaxConnections, 50},
		{"AI_TIMEOUT", cfg.AI.Timeout, 120 * time.Second},
		{"BCRYPT_COST", cfg.Auth.BcryptCost, 14},
		{"COOKIE_SECURE", cfg.Auth.CookieSecure, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s: got %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestLoad_RequiredValues(t *testing.T) {
	setValidEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error = %v", err)
	}

	if cfg.Database.URL != "postgres://localhost/test" {
		t.Errorf("DATABASE_URL: got %q, want %q", cfg.Database.URL, "postgres://localhost/test")
	}

	if cfg.AI.APIKey != "test-key" {
		t.Errorf("ANTHROPIC_API_KEY: got %q, want %q", cfg.AI.APIKey, "test-key")
	}
}

func TestValidLogLevels(t *testing.T) {
	validLevels := []string{"debug", "info", "warn", "error"}

	for _, level := range validLevels {
		t.Run(level, func(t *testing.T) {
			setValidEnv()
			os.Setenv("LOG_LEVEL", level)

			cfg, err := Load()
			if err != nil {
				t.Errorf("Load() with LOG_LEVEL=%q unexpected error = %v", level, err)
				return
			}

			if cfg.Log.Level != level {
				t.Errorf("LOG_LEVEL: got %q, want %q", cfg.Log.Level, level)
			}
		})
	}
}

func TestGetEnvHelpers(t *testing.T) {
	t.Run("getEnv with default", func(t *testing.T) {
		os.Clearenv()
		got := getEnv("MISSING_VAR", "default-value")
		if got != "default-value" {
			t.Errorf("getEnv() = %q, want %q", got, "default-value")
		}
	})

	t.Run("getEnv with value", func(t *testing.T) {
		os.Setenv("PRESENT_VAR", "actual-value")
		got := getEnv("PRESENT_VAR", "default-value")
		if got != "actual-value" {
			t.Errorf("getEnv() = %q, want %q", got, "actual-value")
		}
	})

	t.Run("getEnvInt with default", func(t *testing.T) {
		os.Clearenv()
		got := getEnvInt("MISSING_VAR", 42)
		if got != 42 {
			t.Errorf("getEnvInt() = %d, want %d", got, 42)
		}
	})

	t.Run("getEnvInt with value", func(t *testing.T) {
		os.Setenv("INT_VAR", "100")
		got := getEnvInt("INT_VAR", 42)
		if got != 100 {
			t.Errorf("getEnvInt() = %d, want %d", got, 100)
		}
	})

	t.Run("getEnvInt with invalid value uses default", func(t *testing.T) {
		os.Setenv("BAD_INT", "not-a-number")
		got := getEnvInt("BAD_INT", 42)
		if got != 42 {
			t.Errorf("getEnvInt() with invalid value = %d, want default %d", got, 42)
		}
	})

	t.Run("getEnvBool with default", func(t *testing.T) {
		os.Clearenv()
		got := getEnvBool("MISSING_VAR", true)
		if got != true {
			t.Errorf("getEnvBool() = %v, want %v", got, true)
		}
	})

	t.Run("getEnvBool with true", func(t *testing.T) {
		os.Setenv("BOOL_VAR", "true")
		got := getEnvBool("BOOL_VAR", false)
		if got != true {
			t.Errorf("getEnvBool() = %v, want %v", got, true)
		}
	})

	t.Run("getEnvBool with false", func(t *testing.T) {
		os.Setenv("BOOL_VAR", "false")
		got := getEnvBool("BOOL_VAR", true)
		if got != false {
			t.Errorf("getEnvBool() = %v, want %v", got, false)
		}
	})

	t.Run("getEnvDuration with default", func(t *testing.T) {
		os.Clearenv()
		got := getEnvDuration("MISSING_VAR", 5*time.Second)
		if got != 5*time.Second {
			t.Errorf("getEnvDuration() = %v, want %v", got, 5*time.Second)
		}
	})

	t.Run("getEnvDuration with value", func(t *testing.T) {
		os.Setenv("DURATION_VAR", "30s")
		got := getEnvDuration("DURATION_VAR", 5*time.Second)
		if got != 30*time.Second {
			t.Errorf("getEnvDuration() = %v, want %v", got, 30*time.Second)
		}
	})

	t.Run("getEnvDuration with invalid value uses default", func(t *testing.T) {
		os.Setenv("BAD_DURATION", "not-a-duration")
		got := getEnvDuration("BAD_DURATION", 5*time.Second)
		if got != 5*time.Second {
			t.Errorf("getEnvDuration() with invalid value = %v, want default %v", got, 5*time.Second)
		}
	})
}

// Helper functions

func setValidEnv() {
	os.Clearenv()
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	os.Setenv("ANTHROPIC_API_KEY", "test-key")
	os.Setenv("JWT_SIGNING_KEY", "test-signing-key")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
