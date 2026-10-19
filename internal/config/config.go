package config

import (
	"errors"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all application configuration values.
type Config struct {
	Port           string
	DatabaseURL    string
	AllowedOrigins []string
	Environment    string
}

// Load reads configuration from environment variables, loading a local .env
// file if present. It validates required fields and sets sensible defaults.
func Load() (*Config, error) {
	// Attempt to load .env; ignore error if file is not found (e.g. in production/CI).
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	databaseURL = strings.TrimSpace(databaseURL)

	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	allowedOrigins := parseAllowedOrigins(allowedOriginsStr)

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("ENV")
	}
	if env == "" {
		env = "development"
	}

	cfg := &Config{
		Port:           port,
		DatabaseURL:    databaseURL,
		AllowedOrigins: allowedOrigins,
		Environment:    env,
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks whether mandatory configuration fields are populated.
func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL environment variable is required")
	}
	return nil
}

// parseAllowedOrigins splits comma-separated origins and falls back to localhost if none provided.
func parseAllowedOrigins(originsStr string) []string {
	if originsStr == "" {
		return []string{"http://localhost:3000"}
	}

	var origins []string
	parts := strings.Split(originsStr, ",")
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	if len(origins) == 0 {
		return []string{"http://localhost:3000"}
	}

	return origins
}
