package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery creates a panic recovery middleware with structured logging
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get request ID for tracing
				requestID, _ := c.Get("request_id")

				// Log the panic with stack trace
				logger.Error("panic recovered",
					slog.Any("error", err),
					slog.String("request_id", requestID.(string)),
					slog.String("method", c.Request.Method),
					slog.String("path", c.Request.URL.Path),
					slog.String("client_ip", c.ClientIP()),
					slog.String("stack", string(debug.Stack())),
				)

				// Return generic error to client (don't expose internals)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":      "internal server error",
					"message":    "An unexpected error occurred. Please try again later.",
					"request_id": requestID,
				})
			}
		}()

		c.Next()
	}
}
