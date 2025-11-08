package youtube

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

	"github.com/uiansol/zentube/internal/config"
)

// NewService receives youtube config and returns a YouTube service client
func NewService(cfg config.YouTube) (*youtube.Service, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("youtube: API key missing")
	}
	ctx := context.Background()
	svc, err := youtube.NewService(ctx, option.WithAPIKey(cfg.APIKey))
	if err != nil {
		return nil, fmt.Errorf("creating YouTube service: %w", err)
	}
	return svc, nil
}
