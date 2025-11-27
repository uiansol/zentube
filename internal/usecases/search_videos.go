package usecases

import (
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

func (s *SearchVideos) Execute(query string, maxResults int64) ([]entities.Video, error) {
	videos, err := s.ytClient.Search(query, maxResults)
	if err != nil {
		return nil, err
	}

	// Save search history
	history := &entities.SearchHistory{
		Query:     query,
		Results:   len(videos),
		CreatedAt: time.Now(),
	}

	// Don't fail the search if history save fails, just log it
	if err := s.historyRepo.Save(history); err != nil {
		// In production, you might want to use a proper logger here
		// For now, we'll silently ignore the error
		_ = err
	}

	return videos, nil
}
