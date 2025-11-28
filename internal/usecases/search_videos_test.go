package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

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

// MockSearchHistoryRepository for testing
type MockSearchHistoryRepository struct {
	mock.Mock
}

func (m *MockSearchHistoryRepository) Save(ctx context.Context, history *entities.SearchHistory) error {
	args := m.Called(ctx, history)
	return args.Error(0)
}

func (m *MockSearchHistoryRepository) GetLast(ctx context.Context, limit int) ([]entities.SearchHistory, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.SearchHistory), args.Error(1)
}

func TestSearchVideos_Execute_Success(t *testing.T) {
	// Arrange
	mockClient := new(MockYouTubeClient)
	mockRepo := new(MockSearchHistoryRepository)
	expectedVideos := []entities.Video{
		{
			ID:    "test123",
			Title: "Test Video",
		},
	}

	mockClient.On("Search", "golang", int64(10)).Return(expectedVideos, nil)
	mockRepo.On("Save", mock.Anything, mock.Anything).Return(nil)

	uc := NewSearchVideos(mockClient, mockRepo)

	// Act
	ctx := context.Background()
	videos, err := uc.Execute(ctx, "golang", 10)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, videos, 1)
	assert.Equal(t, "test123", videos[0].ID)
	assert.Equal(t, "Test Video", videos[0].Title)
	mockClient.AssertExpectations(t)

	// Wait a bit for async save to complete
	time.Sleep(100 * time.Millisecond)
	mockRepo.AssertExpectations(t)
}

func TestSearchVideos_Execute_Error(t *testing.T) {
	// Arrange
	mockClient := new(MockYouTubeClient)
	mockRepo := new(MockSearchHistoryRepository)
	expectedError := errors.New("API error")

	mockClient.On("Search", "golang", int64(10)).Return(nil, expectedError)

	uc := NewSearchVideos(mockClient, mockRepo)

	// Act
	ctx := context.Background()
	videos, err := uc.Execute(ctx, "golang", 10)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, videos)
	assert.Equal(t, expectedError, err)
	mockClient.AssertExpectations(t)
	// History should not be saved if search fails
	mockRepo.AssertNotCalled(t, "Save")
}

func TestSearchVideos_Execute_EmptyQuery(t *testing.T) {
	// Arrange
	mockClient := new(MockYouTubeClient)
	mockRepo := new(MockSearchHistoryRepository)
	mockClient.On("Search", "", int64(10)).Return([]entities.Video{}, nil)
	mockRepo.On("Save", mock.Anything, mock.Anything).Return(nil)

	uc := NewSearchVideos(mockClient, mockRepo)

	// Act
	ctx := context.Background()
	videos, err := uc.Execute(ctx, "", 10)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, videos)
	mockClient.AssertExpectations(t)

	// Wait a bit for async save to complete
	time.Sleep(100 * time.Millisecond)
	mockRepo.AssertExpectations(t)
}

func TestSearchVideos_Execute_CacheHit(t *testing.T) {
	// Arrange
	mockClient := new(MockYouTubeClient)
	mockRepo := new(MockSearchHistoryRepository)
	expectedVideos := []entities.Video{
		{
			ID:    "cached123",
			Title: "Cached Video",
		},
	}

	// First call should hit the API
	mockClient.On("Search", "golang", int64(10)).Return(expectedVideos, nil).Once()
	mockRepo.On("Save", mock.Anything, mock.Anything).Return(nil)

	uc := NewSearchVideos(mockClient, mockRepo)
	ctx := context.Background()

	// Act - First call (cache miss)
	videos1, err1 := uc.Execute(ctx, "golang", 10)

	// Assert first call
	assert.NoError(t, err1)
	assert.Len(t, videos1, 1)
	assert.Equal(t, "cached123", videos1[0].ID)

	// Wait for async save
	time.Sleep(100 * time.Millisecond)

	// Act - Second call with same parameters (cache hit)
	videos2, err2 := uc.Execute(ctx, "golang", 10)

	// Assert second call
	assert.NoError(t, err2)
	assert.Len(t, videos2, 1)
	assert.Equal(t, "cached123", videos2[0].ID)

	// Client should only be called once (first call was cached)
	mockClient.AssertExpectations(t)
	mockClient.AssertNumberOfCalls(t, "Search", 1)
}

func TestSearchVideos_Execute_CacheMissDifferentParams(t *testing.T) {
	// Arrange
	mockClient := new(MockYouTubeClient)
	mockRepo := new(MockSearchHistoryRepository)
	videos1 := []entities.Video{{ID: "vid1", Title: "Video 1"}}
	videos2 := []entities.Video{{ID: "vid2", Title: "Video 2"}}

	// Different queries should result in cache misses
	mockClient.On("Search", "golang", int64(10)).Return(videos1, nil).Once()
	mockClient.On("Search", "python", int64(10)).Return(videos2, nil).Once()
	mockRepo.On("Save", mock.Anything, mock.Anything).Return(nil)

	uc := NewSearchVideos(mockClient, mockRepo)
	ctx := context.Background()

	// Act
	result1, err1 := uc.Execute(ctx, "golang", 10)
	result2, err2 := uc.Execute(ctx, "python", 10)

	// Assert
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, "vid1", result1[0].ID)
	assert.Equal(t, "vid2", result2[0].ID)

	// Both calls should hit the API (different cache keys)
	mockClient.AssertNumberOfCalls(t, "Search", 2)
}
