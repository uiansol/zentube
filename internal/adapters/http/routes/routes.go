package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/uiansol/zentube/internal/adapters/http/handlers"
)

func RegisterRoutes(r *gin.Engine, h *handlers.YouTubeHandler) {
	r.Static("/static", "./web/static")
	r.GET("/", h.Home)
	r.POST("/search", h.Search)
}
