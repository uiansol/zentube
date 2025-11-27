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
	if err := pages.HomePage("", nil).Render(c.Request.Context(), c.Writer); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to render page")
		return
	}
}

func (h *YouTubeHandler) Search(c *gin.Context) {
	query := c.PostForm("q")
	if query == "" {
		if err := components.SearchResults(nil).Render(c.Request.Context(), c.Writer); err != nil {
			respondError(c, http.StatusInternalServerError, "Failed to render search results")
		}
		return
	}

	videos, err := h.searchUC.Execute(c.Request.Context(), query, h.maxResults)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to search videos")
		return
	}

	// Check if it's an HTMX request - return only results fragment
	if middleware.IsHTMXRequest(c) {
		if err := components.SearchResults(videos).Render(c.Request.Context(), c.Writer); err != nil {
			respondError(c, http.StatusInternalServerError, "Failed to render search results")
		}
	} else {
		// Regular request - return full page
		if err := pages.HomePage(query, videos).Render(c.Request.Context(), c.Writer); err != nil {
			respondError(c, http.StatusInternalServerError, "Failed to render page")
		}
	}
}
