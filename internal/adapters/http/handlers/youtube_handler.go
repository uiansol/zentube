package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/uiansol/zentube/internal/adapters/http/middleware"
	appErrors "github.com/uiansol/zentube/internal/errors"
	"github.com/uiansol/zentube/internal/usecases"
	"github.com/uiansol/zentube/internal/validation"
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
		respondError(c, appErrors.NewInternalError("Failed to render page", err), "Failed to render page")
		return
	}
}

func (h *YouTubeHandler) Search(c *gin.Context) {
	query := c.PostForm("q")

	// Validate and sanitize input
	input, err := validation.ValidateSearchQuery(query, h.maxResults)
	if err != nil {
		// Validation errors are AppErrors with proper status codes
		respondError(c, err, "Invalid search query")
		return
	}

	// Execute search with validated input
	videos, err := h.searchUC.Execute(c.Request.Context(), input.Query, input.MaxResults)
	if err != nil {
		respondError(c, appErrors.NewInternalError("Failed to search videos", err), "Failed to search videos")
		return
	}

	// Check if it's an HTMX request - return only results fragment
	if middleware.IsHTMXRequest(c) {
		if err := components.SearchResults(videos).Render(c.Request.Context(), c.Writer); err != nil {
			respondError(c, appErrors.NewInternalError("Failed to render search results", err), "Failed to render search results")
		}
	} else {
		// Regular request - return full page
		if err := pages.HomePage(input.Query, videos).Render(c.Request.Context(), c.Writer); err != nil {
			respondError(c, appErrors.NewInternalError("Failed to render page", err), "Failed to render page")
		}
	}
}
