package usecases

import (
	"context"
	"time"

	"github.com/uiansol/zentube/internal/cache"
	"github.com/uiansol/zentube/internal/entities"
	"github.com/uiansol/zentube/internal/ports"
)

type SearchVideos struct {
	ytClient    ports.YouTubeClient
	historyRepo ports.SearchHistoryRepository
	cache       *cache.Cache // Optional cache for reducing API calls
}

// NewSearchVideos creates a new SearchVideos use case
// Cache is optional - pass nil to disable caching
func NewSearchVideos(ytClient ports.YouTubeClient, historyRepo ports.SearchHistoryRepository) *SearchVideos {
	return &SearchVideos{
		ytClient:    ytClient,
		historyRepo: historyRepo,
		// Initialize cache with 1000 entries max, 5-minute TTL
		// This prevents hammering the YouTube API with duplicate searches
		cache: cache.NewCache(1000, 5*time.Minute),
	}
}

func (s *SearchVideos) Execute(ctx context.Context, query string, maxResults int64) ([]entities.Video, error) {
	// Generate cache key from query and maxResults
	cacheKey := cache.GenerateKey("search", query, maxResults)

	// Try to get from cache first
	if s.cache != nil {
		if cached, found := s.cache.Get(cacheKey); found {
			// Cache hit! Return cached results
			if videos, ok := cached.([]entities.Video); ok {
				return videos, nil
			}
		}
	}

	// Cache miss - fetch from YouTube API
	videos, err := s.ytClient.Search(query, maxResults)
	if err != nil {
		return nil, err
	}

	// Store in cache for future requests
	if s.cache != nil {
		s.cache.Set(cacheKey, videos)
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
