package handlers

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler(db *sql.DB, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		db:     db,
		logger: logger,
	}
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Checks    map[string]string `json:"checks,omitempty"`
}

// Live checks if the server is running (liveness probe)
// Returns 200 if server can accept requests
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// Ready checks if the server is ready to handle requests (readiness probe)
// Returns 200 only if all dependencies are healthy
func (h *HealthHandler) Ready(c *gin.Context) {
	checks := make(map[string]string)
	allHealthy := true

	// Check database connection
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		checks["database"] = "unhealthy: " + err.Error()
		allHealthy = false
		h.logger.Error("database health check failed", slog.Any("error", err))
	} else {
		checks["database"] = "healthy"
	}

	// You can add more dependency checks here:
	// - Redis connection
	// - External API availability
	// - File system write permissions
	// - etc.

	status := "ok"
	statusCode := http.StatusOK

	if !allHealthy {
		status = "degraded"
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, HealthResponse{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    checks,
	})
}
