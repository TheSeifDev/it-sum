package main

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
)

func TestSanitizeError(t *testing.T) {
	t.Run("nil error returns nil", func(t *testing.T) {
		if sanitizeError(nil) != nil {
			t.Error("expected nil for nil input")
		}
	})

	t.Run("redacts postgres database URL with credentials", func(t *testing.T) {
		rawErr := errors.New("failed connecting to postgres://user:secret123@ep-cool-db.aws.neon.tech/neondb: connection timed out")
		sanitized := sanitizeError(rawErr)

		errMsg := sanitized.Error()
		if strings.Contains(errMsg, "secret123") || strings.Contains(errMsg, "ep-cool-db") {
			t.Errorf("error leaked sensitive details: %s", errMsg)
		}
		if !strings.Contains(errMsg, "postgres://[REDACTED]") {
			t.Errorf("expected redacted URL placeholder, got: %s", errMsg)
		}
	})
}

func TestIsNoMigrationsError(t *testing.T) {
	if isNoMigrationsError(nil) {
		t.Error("expected false for nil error")
	}
	if !isNoMigrationsError(os.ErrNotExist) {
		t.Error("expected true for os.ErrNotExist")
	}
	if !isNoMigrationsError(migrate.ErrNilVersion) {
		t.Error("expected true for migrate.ErrNilVersion")
	}
	if isNoMigrationsError(errors.New("generic error")) {
		t.Error("expected false for generic error")
	}
}
