package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uiansol/zentube/internal/adapters/http/middleware"
	"github.com/uiansol/zentube/internal/usecases"
	"github.com/uiansol/zentube/web/templates/components"
	"github.com/uiansol/zentube/web/templates/pages"
)

type YouTubeHandler struct {
	searchUC   *usecases.SearchVideos
	maxResults int64
}

func NewYouTubeHandler(searchUC *usecases.SearchVideos, maxResults int64) *YouTubeHandler {
	return &YouTubeHandler{searchUC: searchUC, maxResults: maxResults}
}

func (h *YouTubeHandler) Home(c *gin.Context) {
	pages.HomePage("", nil).Render(c.Request.Context(), c.Writer)
}

func (h *YouTubeHandler) Search(c *gin.Context) {
	query := c.PostForm("q")
	if query == "" {
		components.SearchResults(nil).Render(c.Request.Context(), c.Writer)
		return
	}

	videos, err := h.searchUC.Execute(query, h.maxResults)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to search videos")
		return
	}

	// Check if it's an HTMX request - return only results fragment
	if middleware.IsHTMXRequest(c) {
		components.SearchResults(videos).Render(c.Request.Context(), c.Writer)
	} else {
		// Regular request - return full page
		pages.HomePage(query, videos).Render(c.Request.Context(), c.Writer)
	}
}
