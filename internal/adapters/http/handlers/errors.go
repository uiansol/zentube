package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uiansol/zentube/internal/adapters/http/middleware"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// respondError handles error responses with structured logging
func respondError(c *gin.Context, statusCode int, message string) {
	// Get request ID for tracing
	requestID, _ := c.Get("request_id")

	// Log the error with structured fields
	slog.Error("request error",
		slog.String("message", message),
		slog.Int("status", statusCode),
		slog.String("request_id", requestID.(string)),
		slog.String("path", c.Request.URL.Path),
		slog.String("method", c.Request.Method),
	)

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
