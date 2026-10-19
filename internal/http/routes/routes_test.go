package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TheSeifDev/it-sum/internal/http/handlers"
	"github.com/gin-gonic/gin"
)

type dummyPinger struct{}

func (d *dummyPinger) Ping(ctx context.Context) error {
	return nil
}

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRegisterRoutes(t *testing.T) {
	router := gin.New()
	healthHandler := handlers.NewHealthHandler(&dummyPinger{})

	RegisterRoutes(router, Config{
		HealthHandler: healthHandler,
	})

	t.Run("GET /api/v1/health is registered", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("GET /api/v1/health/db is registered", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health/db", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("unregistered route returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/nonexistent", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}
