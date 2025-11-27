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

// MockSearchHistoryRepository for testing
type MockSearchHistoryRepository struct {
	mock.Mock
}

func (m *MockSearchHistoryRepository) Save(history *entities.SearchHistory) error {
	args := m.Called(history)
	return args.Error(0)
}

func (m *MockSearchHistoryRepository) GetLast(limit int) ([]entities.SearchHistory, error) {
	args := m.Called(limit)
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
	mockRepo.On("Save", mock.Anything).Return(nil)

	uc := NewSearchVideos(mockClient, mockRepo)

	// Act
	videos, err := uc.Execute("golang", 10)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, videos, 1)
	assert.Equal(t, "test123", videos[0].ID)
	assert.Equal(t, "Test Video", videos[0].Title)
	mockClient.AssertExpectations(t)
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
	videos, err := uc.Execute("golang", 10)

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
	mockRepo.On("Save", mock.Anything).Return(nil)

	uc := NewSearchVideos(mockClient, mockRepo)

	// Act
	videos, err := uc.Execute("", 10)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, videos)
	mockClient.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}
