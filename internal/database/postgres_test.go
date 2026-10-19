package database

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestDefaultPoolConfig(t *testing.T) {
	cfg := DefaultPoolConfig()

	if cfg.MaxConns != 5 {
		t.Errorf("expected MaxConns to be 5, got %d", cfg.MaxConns)
	}
	if cfg.MinConns != 0 {
		t.Errorf("expected MinConns to be 0, got %d", cfg.MinConns)
	}
	if cfg.MaxConnLifetime != 1*time.Hour {
		t.Errorf("expected MaxConnLifetime to be 1h, got %v", cfg.MaxConnLifetime)
	}
	if cfg.MaxConnIdleTime != 5*time.Minute {
		t.Errorf("expected MaxConnIdleTime to be 5m, got %v", cfg.MaxConnIdleTime)
	}
	if cfg.ConnectTimeout != 5*time.Second {
		t.Errorf("expected ConnectTimeout to be 5s, got %v", cfg.ConnectTimeout)
	}
}

func TestNewPoolValidation(t *testing.T) {
	ctx := context.Background()

	t.Run("empty URL returns error", func(t *testing.T) {
		pool, err := NewPool(ctx, "")
		if err == nil {
			t.Fatal("expected error with empty URL, got nil")
		}
		if pool != nil {
			t.Fatal("expected nil pool on error")
		}
		if err.Error() != "database URL cannot be empty" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("invalid URL format returns sanitized error without leaking credentials", func(t *testing.T) {
		sensitiveURL := "postgres://supersecretuser:supersecretpass@::invalid-host::/mydb"
		pool, err := NewPool(ctx, sensitiveURL)
		if err == nil {
			t.Fatal("expected error with invalid URL, got nil")
		}
		if pool != nil {
			t.Fatal("expected nil pool on error")
		}

		errMsg := err.Error()
		if strings.Contains(errMsg, "supersecretuser") || strings.Contains(errMsg, "supersecretpass") {
			t.Errorf("error message leaked credentials: %s", errMsg)
		}
		if errMsg != "invalid database URL format" {
			t.Errorf("expected sanitized error 'invalid database URL format', got %q", errMsg)
		}
	})

	t.Run("unreachable host returns sanitized error without leaking credentials", func(t *testing.T) {
		// Test unreachable address with a quick timeout
		cfg := DefaultPoolConfig()
		cfg.ConnectTimeout = 100 * time.Millisecond

		unreachableURL := "postgres://secretuser:secretpass@127.0.0.1:54321/nonexistent?sslmode=disable"
		ctxTimeout, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
		defer cancel()

		pool, err := NewPool(ctxTimeout, unreachableURL, cfg)
		if err == nil {
			if pool != nil {
				pool.Close()
			}
			t.Fatal("expected connection to fail for unreachable host")
		}

		errMsg := err.Error()
		if strings.Contains(errMsg, "secretuser") || strings.Contains(errMsg, "secretpass") {
			t.Errorf("error message leaked credentials: %s", errMsg)
		}
	})
}
