package config

import (
	"os"
	"reflect"
	"testing"
)

func TestParseAllowedOrigins(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty input defaults to localhost",
			input:    "",
			expected: []string{"http://localhost:3000"},
		},
		{
			name:     "single origin trimmed",
			input:    " https://app.example.com ",
			expected: []string{"https://app.example.com"},
		},
		{
			name:     "multiple comma-separated origins with whitespace",
			input:    " http://localhost:3000 , https://preview.example.com,https://app.example.com ",
			expected: []string{"http://localhost:3000", "https://preview.example.com", "https://app.example.com"},
		},
		{
			name:     "string with only commas and whitespace",
			input:    " , ,   ",
			expected: []string{"http://localhost:3000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAllowedOrigins(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("parseAllowedOrigins(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	t.Run("missing DATABASE_URL returns error", func(t *testing.T) {
		cfg := &Config{
			Port:        "8080",
			DatabaseURL: "",
		}
		if err := cfg.Validate(); err == nil {
			t.Error("expected error for empty DATABASE_URL, got nil")
		}
	})

	t.Run("valid configuration passes validation", func(t *testing.T) {
		cfg := &Config{
			Port:        "8080",
			DatabaseURL: "postgres://user:pass@localhost:5432/db",
		}
		if err := cfg.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestLoad(t *testing.T) {
	// Backup original env vars
	origPort := os.Getenv("PORT")
	origDB := os.Getenv("DATABASE_URL")
	origOrigins := os.Getenv("ALLOWED_ORIGINS")
	origEnv := os.Getenv("APP_ENV")

	defer func() {
		os.Setenv("PORT", origPort)
		os.Setenv("DATABASE_URL", origDB)
		os.Setenv("ALLOWED_ORIGINS", origOrigins)
		os.Setenv("APP_ENV", origEnv)
	}()

	t.Run("fails when DATABASE_URL is unset", func(t *testing.T) {
		os.Unsetenv("DATABASE_URL")
		_, err := Load()
		if err == nil {
			t.Error("expected error when DATABASE_URL is unset")
		}
	})

	t.Run("succeeds with custom environment variables", func(t *testing.T) {
		os.Setenv("PORT", "9090")
		os.Setenv("DATABASE_URL", "postgres://testuser:testpass@localhost:5432/testdb")
		os.Setenv("ALLOWED_ORIGINS", "http://localhost:3000,https://app.it-sum.com")
		os.Setenv("APP_ENV", "production")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error loading config: %v", err)
		}

		if cfg.Port != "9090" {
			t.Errorf("expected port 9090, got %s", cfg.Port)
		}
		if cfg.DatabaseURL != "postgres://testuser:testpass@localhost:5432/testdb" {
			t.Errorf("unexpected database URL: %s", cfg.DatabaseURL)
		}
		if len(cfg.AllowedOrigins) != 2 || cfg.AllowedOrigins[1] != "https://app.it-sum.com" {
			t.Errorf("unexpected allowed origins: %v", cfg.AllowedOrigins)
		}
		if cfg.Environment != "production" {
			t.Errorf("expected production env, got %s", cfg.Environment)
		}
	})
}
