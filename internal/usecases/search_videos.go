package usecases

import (
	"context"
	"time"

	"github.com/uiansol/zentube/internal/entities"
	"github.com/uiansol/zentube/internal/ports"
)

type SearchVideos struct {
	ytClient    ports.YouTubeClient
	historyRepo ports.SearchHistoryRepository
}

func NewSearchVideos(ytClient ports.YouTubeClient, historyRepo ports.SearchHistoryRepository) *SearchVideos {
	return &SearchVideos{
		ytClient:    ytClient,
		historyRepo: historyRepo,
	}
}

func (s *SearchVideos) Execute(ctx context.Context, query string, maxResults int64) ([]entities.Video, error) {
	videos, err := s.ytClient.Search(query, maxResults)
	if err != nil {
		return nil, err
	}

	// Save search history asynchronously with a timeout
	// Use background context with timeout to avoid blocking the response
	go func() {
		saveCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		history := &entities.SearchHistory{
			Query:     query,
			Results:   len(videos),
			CreatedAt: time.Now(),
		}

		// Don't fail the search if history save fails
		_ = s.historyRepo.Save(saveCtx, history)
	}()

	return videos, nil
}
