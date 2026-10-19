package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// DatabasePinger defines the capability to check database liveness/readiness.
type DatabasePinger interface {
	Ping(ctx context.Context) error
}

// HealthHandler manages application and database health endpoints.
type HealthHandler struct {
	db DatabasePinger
}

// NewHealthHandler creates a new HealthHandler instance with constructor-based dependency injection.
func NewHealthHandler(db DatabasePinger) *HealthHandler {
	return &HealthHandler{db: db}
}

// Health handles GET /api/v1/health (general application liveness).
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "it-sum-api",
		"status":  "ok",
	})
}

// DatabaseHealth handles GET /api/v1/health/db (database readiness).
// It performs a bounded connectivity check and returns sanitized responses without leaking raw errors.
func (h *HealthHandler) DatabaseHealth(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   "error",
			"database": "unavailable",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   "error",
			"database": "unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"database": "connected",
	})
}
