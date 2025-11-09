package usecases

import (
	"testing"

	"github.com/uiansol/zentube/internal/entities"
)

// MockYouTubeClient for testing
type MockYouTubeClient struct {
	SearchFunc func(query string, maxResults int64) ([]entities.Video, error)
}

func (m *MockYouTubeClient) Search(query string, maxResults int64) ([]entities.Video, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(query, maxResults)
	}
	return nil, nil
}

func TestSearchVideos_Execute(t *testing.T) {
	mockClient := &MockYouTubeClient{
		SearchFunc: func(query string, maxResults int64) ([]entities.Video, error) {
			return []entities.Video{
				{
					ID:    "test123",
					Title: "Test Video",
				},
			}, nil
		},
	}

	uc := NewSearchVideos(mockClient)
	videos, err := uc.Execute("golang", 10)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(videos) != 1 {
		t.Fatalf("expected 1 video, got %d", len(videos))
	}

	if videos[0].ID != "test123" {
		t.Errorf("expected video ID 'test123', got '%s'", videos[0].ID)
	}
}
