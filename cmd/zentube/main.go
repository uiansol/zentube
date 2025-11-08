package main

import (
	"fmt"
	"log"
	"os"

	"github.com/uiansol/zentube/internal/adapters/youtube"
	"github.com/uiansol/zentube/internal/config"
	"github.com/uiansol/zentube/internal/usecases"
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

	// Initialize use cases
	searchVideos := usecases.NewSearchVideos(ytClient)

	// Example search
	videos, err := searchVideos.Execute("golang tutorial", cfg.YouTube.MaxResults)
	if err != nil {
		return fmt.Errorf("failed to search videos: %w", err)
	}

	// Display results
	for _, video := range videos {
		link := "https://www.youtube.com/watch?v=" + video.ID
		log.Printf("%s – %s", video.Title, link)
	}

	return nil
}
