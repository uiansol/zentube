package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/uiansol/zentube/internal/adapters/http/handlers"
	"github.com/uiansol/zentube/internal/adapters/http/middleware"
)

func RegisterRoutes(r *gin.Engine, h *handlers.YouTubeHandler) {
	// Apply middleware
	r.Use(middleware.Logger())
	r.Use(middleware.HTMX())

	// Routes
	r.Static("/static", "./web/static")
	r.GET("/", h.Home)
	r.POST("/search", h.Search)
}
