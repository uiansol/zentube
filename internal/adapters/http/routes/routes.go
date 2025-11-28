package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/uiansol/zentube/internal/adapters/http/handlers"
	"github.com/uiansol/zentube/internal/adapters/http/middleware"
)

func RegisterRoutes(r *gin.Engine, h *handlers.YouTubeHandler, health *handlers.HealthHandler) {
	// Apply HTMX middleware for all routes
	r.Use(middleware.HTMX())

	// Health check endpoints (no rate limiting)
	r.GET("/health/live", health.Live)
	r.GET("/health/ready", health.Ready)

	// Static files
	r.Static("/static", "./web/static")

	// Application routes
	r.GET("/", h.Home)
	r.POST("/search", h.Search)
}
