package youtube

import (
	"context"
	"time"

	"github.com/uiansol/zentube/internal/entities"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type YouTubeClient struct {
	service *youtube.Service
}

func NewYouTubeClient(apiKey string) (*YouTubeClient, error) {
	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	return &YouTubeClient{service: service}, nil
}

func (c *YouTubeClient) Search(query string, maxResults int64) ([]entities.Video, error) {
	call := c.service.Search.List([]string{"snippet"}).
		Q(query).
		Type("video").
		MaxResults(maxResults)

	resp, err := call.Do()
	if err != nil {
		return nil, err
	}

	videos := make([]entities.Video, 0, len(resp.Items))
	for _, item := range resp.Items {
		if item.Id == nil || item.Snippet == nil {
			continue
		}
		pubTime, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		v := entities.Video{
			ID:          item.Id.VideoId,
			Title:       item.Snippet.Title,
			Channel:     item.Snippet.ChannelTitle,
			PublishedAt: pubTime,
			Thumbnail:   item.Snippet.Thumbnails.Default.Url,
		}
		videos = append(videos, v)
	}

	return videos, nil
}
