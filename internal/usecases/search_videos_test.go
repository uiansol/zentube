package usecases

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/uiansol/zentube/internal/entities"
)

// MockYouTubeClient for testing
type MockYouTubeClient struct {
	mock.Mock
}

func (m *MockYouTubeClient) Search(query string, maxResults int64) ([]entities.Video, error) {
	args := m.Called(query, maxResults)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.Video), args.Error(1)
}

func TestSearchVideos_Execute_Success(t *testing.T) {
	// Arrange
	mockClient := new(MockYouTubeClient)
	expectedVideos := []entities.Video{
		{
			ID:    "test123",
			Title: "Test Video",
		},
	}

	mockClient.On("Search", "golang", int64(10)).Return(expectedVideos, nil)

	uc := NewSearchVideos(mockClient)

	// Act
	videos, err := uc.Execute("golang", 10)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, videos, 1)
	assert.Equal(t, "test123", videos[0].ID)
	assert.Equal(t, "Test Video", videos[0].Title)
	mockClient.AssertExpectations(t)
}

func TestSearchVideos_Execute_Error(t *testing.T) {
	// Arrange
	mockClient := new(MockYouTubeClient)
	expectedError := errors.New("API error")

	mockClient.On("Search", "golang", int64(10)).Return(nil, expectedError)

	uc := NewSearchVideos(mockClient)

	// Act
	videos, err := uc.Execute("golang", 10)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, videos)
	assert.Equal(t, expectedError, err)
	mockClient.AssertExpectations(t)
}

func TestSearchVideos_Execute_EmptyQuery(t *testing.T) {
	// Arrange
	mockClient := new(MockYouTubeClient)
	mockClient.On("Search", "", int64(10)).Return([]entities.Video{}, nil)

	uc := NewSearchVideos(mockClient)

	// Act
	videos, err := uc.Execute("", 10)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, videos)
	mockClient.AssertExpectations(t)
}
