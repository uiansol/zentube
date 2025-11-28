package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID generates a unique request ID for tracing
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request already has an ID (e.g., from load balancer)
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store in context for use in handlers
		c.Set("request_id", requestID)

		// Return in response header for client-side tracing
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}
