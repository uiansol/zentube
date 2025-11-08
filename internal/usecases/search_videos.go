package usecases

import (
	"github.com/uiansol/zentube/internal/entities"
	"github.com/uiansol/zentube/internal/ports"
)

type SearchVideos struct {
	ytClient ports.YouTubeClient
}

func NewSearchVideos(ytClient ports.YouTubeClient) *SearchVideos {
	return &SearchVideos{ytClient: ytClient}
}

func (s *SearchVideos) Execute(query string, maxResults int64) ([]entities.Video, error) {
	videos, err := s.ytClient.Search(query, maxResults)
	if err != nil {
		return nil, err
	}

	return videos, nil
}
