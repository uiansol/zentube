package main

import (
	"context"
	"fmt"
	"log"
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
	"github.com/uiansol/zentube/internal/adapters/http/routes"
	"github.com/uiansol/zentube/internal/adapters/youtube"
	"github.com/uiansol/zentube/internal/config"
	"github.com/uiansol/zentube/internal/usecases"
	"github.com/uiansol/zentube/web/templates/pages"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load environment variables
	if err := config.LoadEnv(); err != nil {
		return fmt.Errorf("failed to load environment: %w", err)
	}

	// Load configuration
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Inject environment variables into config
	if err := config.InjectEnvVariables(cfg); err != nil {
		return fmt.Errorf("failed to inject env variables: %w", err)
	}

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
	defer func() {
		if err := dbRepo.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Initialize use cases
	searchVideos := usecases.NewSearchVideos(ytClient, dbRepo)
	ytHandler := handlers.NewYouTubeHandler(searchVideos, cfg.YouTube.MaxResults)

	// Setup Gin router
	r := gin.Default()
	routes.RegisterRoutes(r, ytHandler)

	// Ensure templates compile (helps catch errors early)
	_ = pages.HomePage("", nil)

	// Determine port
	port := cfg.App.Port
	if port == 0 {
		port = 8080
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 zentube running on http://localhost:%d", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Give active connections 5 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("✅ Server exited gracefully")
	return nil
}
