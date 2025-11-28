package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uiansol/zentube/internal/adapters/http/middleware"
	appErrors "github.com/uiansol/zentube/internal/errors"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	Code      string `json:"code,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// respondError handles error responses with structured logging
// Supports both standard errors and custom AppError types
func respondError(c *gin.Context, err error, fallbackMessage string) {
	// Get request ID for tracing
	requestID, _ := c.Get("request_id")
	reqID := ""
	if requestID != nil {
		reqID = requestID.(string)
	}

	// Determine status code and error details
	var statusCode int
	var errorCode string
	var message string

	if appErr, ok := err.(*appErrors.AppError); ok {
		// Use AppError details
		statusCode = appErr.StatusCode
		errorCode = appErr.Code
		message = appErr.Message
	} else {
		// Fallback for standard errors
		statusCode = http.StatusInternalServerError
		errorCode = appErrors.ErrCodeInternal
		message = fallbackMessage
	}

	// Log the error with structured fields
	slog.Error("request error",
		slog.String("message", message),
		slog.String("error_code", errorCode),
		slog.Int("status", statusCode),
		slog.String("request_id", reqID),
		slog.String("path", c.Request.URL.Path),
		slog.String("method", c.Request.Method),
		slog.Any("error", err),
	)

	// For HTMX requests, return simple error message
	if middleware.IsHTMXRequest(c) {
		c.String(statusCode, message)
		return
	}

	// For regular requests, return structured JSON
	c.JSON(statusCode, ErrorResponse{
		Error:     http.StatusText(statusCode),
		Message:   message,
		Code:      errorCode,
		RequestID: reqID,
	})
}

// respondAppError is a convenience wrapper for AppError
func respondAppError(c *gin.Context, err *appErrors.AppError) {
	respondError(c, err, err.Message)
}

// respondSuccess sends a successful response with optional data
func respondSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}
