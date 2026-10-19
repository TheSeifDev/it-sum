package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// mockDatabasePinger implements DatabasePinger for testing.
type mockDatabasePinger struct {
	err error
}

func (m *mockDatabasePinger) Ping(ctx context.Context) error {
	return m.err
}

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHealth(t *testing.T) {
	handler := NewHealthHandler(&mockDatabasePinger{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.Health(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal JSON response: %v", err)
	}

	if response["service"] != "it-sum-api" {
		t.Errorf("expected service 'it-sum-api', got %q", response["service"])
	}
	if response["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", response["status"])
	}
}

func TestDatabaseHealth_Success(t *testing.T) {
	mockDB := &mockDatabasePinger{err: nil}
	handler := NewHealthHandler(mockDB)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/db", nil)
	c.Request = req

	handler.DatabaseHealth(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal JSON response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", response["status"])
	}
	if response["database"] != "connected" {
		t.Errorf("expected database 'connected', got %q", response["database"])
	}
}

func TestDatabaseHealth_Failure(t *testing.T) {
	mockDB := &mockDatabasePinger{err: errors.New("connection reset by peer")}
	handler := NewHealthHandler(mockDB)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/db", nil)
	c.Request = req

	handler.DatabaseHealth(c)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal JSON response: %v", err)
	}

	if response["status"] != "error" {
		t.Errorf("expected status 'error', got %q", response["status"])
	}
	if response["database"] != "unavailable" {
		t.Errorf("expected database 'unavailable', got %q", response["database"])
	}
	// Verify raw internal error is never leaked in JSON
	if _, exists := response["error"]; exists {
		t.Error("sanitized response must not contain raw error details")
	}
}

func TestDatabaseHealth_NilDB(t *testing.T) {
	handler := NewHealthHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/db", nil)
	c.Request = req

	handler.DatabaseHealth(c)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal JSON response: %v", err)
	}

	if response["database"] != "unavailable" {
		t.Errorf("expected database 'unavailable', got %q", response["database"])
	}
}
