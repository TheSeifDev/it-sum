package routes

import (
	"github.com/TheSeifDev/it-sum/internal/http/handlers"
	"github.com/gin-gonic/gin"
)

// Config defines the handlers and dependencies needed to register application routes.
type Config struct {
	HealthHandler *handlers.HealthHandler
}

// RegisterRoutes registers all API routes onto the given Gin engine.
func RegisterRoutes(router *gin.Engine, cfg Config) {
	v1 := router.Group("/api/v1")
	{
		if cfg.HealthHandler != nil {
			v1.GET("/health", cfg.HealthHandler.Health)
			v1.GET("/health/db", cfg.HealthHandler.DatabaseHealth)
		}
	}
}
