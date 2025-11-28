package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uiansol/zentube/internal/adapters/database"
	"github.com/uiansol/zentube/internal/adapters/http/handlers"
	"github.com/uiansol/zentube/internal/adapters/http/middleware"
	"github.com/uiansol/zentube/internal/adapters/http/routes"
	"github.com/uiansol/zentube/internal/adapters/youtube"
	"github.com/uiansol/zentube/internal/config"
	"github.com/uiansol/zentube/internal/usecases"
	"github.com/uiansol/zentube/web/templates/pages"
	"golang.org/x/time/rate"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Determine environment and initialize structured logger
	env := config.GetEnvironment()
	logger := middleware.NewLogger(string(env))
	slog.SetDefault(logger) // Set as default logger for the application

	logger.Info("starting zentube",
		slog.String("env", string(env)),
		slog.String("version", "1.0.0"),
	)

	// Load environment variables (supports .env.<environment>)
	if err := config.LoadEnv(); err != nil {
		return fmt.Errorf("failed to load environment: %w", err)
	}

	// Load configuration (supports config.<environment>.yaml)
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logger.Info("loaded configuration",
		slog.String("environment", string(cfg.App.Environment)),
		slog.Int("port", cfg.App.Port),
		slog.String("database", cfg.Database.Path),
	)

	// Inject environment variables into config
	if err := config.InjectEnvVariables(cfg); err != nil {
		return fmt.Errorf("failed to inject env variables: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	logger.Info("configuration loaded successfully",
		slog.String("app_name", cfg.App.Name),
		slog.Int("port", cfg.App.Port),
	)

	// Initialize YouTube client
	ytClient, err := youtube.NewYouTubeClient(cfg.YouTube.APIKey)
	if err != nil {
		return fmt.Errorf("failed to create youtube client: %w", err)
	}

	// Initialize database
	// Ensure the database directory exists
	dbDir := filepath.Dir(cfg.Database.Path)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	dbRepo, err := database.NewSQLiteRepository(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	logger.Info("database initialized",
		slog.String("path", cfg.Database.Path),
	)

	// Initialize use cases
	searchVideos := usecases.NewSearchVideos(ytClient, dbRepo)
	ytHandler := handlers.NewYouTubeHandler(searchVideos, cfg.YouTube.MaxResults)
	healthHandler := handlers.NewHealthHandler(dbRepo.DB(), logger)

	// Setup Gin router (disable default middleware, we'll add our own)
	// Set Gin mode based on environment
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else if cfg.IsDevelopment() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.TestMode)
	}

	r := gin.New()

	// Apply custom middleware in order
	r.Use(middleware.Recovery(logger))                      // Recover from panics
	r.Use(middleware.RequestID())                           // Generate request IDs
	r.Use(middleware.Middleware(logger))                    // Structured logging
	r.Use(middleware.SecurityHeaders())                     // Security headers
	r.Use(middleware.RateLimit(rate.Limit(10), 20, logger)) // 10 req/sec, burst 20

	// Register routes
	routes.RegisterRoutes(r, ytHandler, healthHandler)

	// Ensure templates compile (helps catch errors early)
	_ = pages.HomePage("", nil)

	// Determine port
	port := cfg.App.Port
	if port == 0 {
		port = 8080
	}

	// Create HTTP server with timeouts
	srv := &http.Server{
		Addr:           ":" + strconv.Itoa(port),
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in a goroutine
	go func() {
		logger.Info("server started",
			slog.Int("port", port),
			slog.String("address", fmt.Sprintf("http://localhost:%d", port)),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed to start", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutdown signal received")

	// Give active connections 5 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown HTTP server first (stops accepting new requests, waits for active ones)
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", slog.Any("error", err))
	}

	// Now it's safe to close the database (all HTTP handlers have completed)
	logger.Info("closing database connection")
	if err := dbRepo.Close(); err != nil {
		logger.Error("error closing database", slog.Any("error", err))
	}

	logger.Info("server exited gracefully")
	return nil
}
