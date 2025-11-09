package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uiansol/zentube/internal/adapters/http/middleware"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// respondError handles error responses
func respondError(c *gin.Context, statusCode int, message string) {
	log.Printf("Error: %s (status: %d)", message, statusCode)

	// For HTMX requests, return simple error message
	if middleware.IsHTMXRequest(c) {
		c.String(statusCode, message)
		return
	}

	// For regular requests, return JSON
	c.JSON(statusCode, ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	})
}
