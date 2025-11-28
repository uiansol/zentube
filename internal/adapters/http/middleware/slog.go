package middleware

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// NewLogger creates a new structured logger based on environment
func NewLogger(env string) *slog.Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	// Use JSON in production, human-readable in development
	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

// Middleware creates a Gin middleware for structured request logging
func Middleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Get or generate request ID
		requestID, exists := c.Get("request_id")
		if !exists {
			requestID = ""
		}

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// Log with structured fields
		logger.LogAttrs(
			context.Background(),
			slog.LevelInfo,
			"request completed",
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", statusCode),
			slog.Duration("duration", duration),
			slog.String("request_id", requestID.(string)),
			slog.String("client_ip", c.ClientIP()),
		)
	}
}

// FromContext retrieves logger from context, falls back to default
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value("logger").(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// WithContext adds logger to context
func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, "logger", logger)
}
