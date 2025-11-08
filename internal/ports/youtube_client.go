package ports

import "github.com/uiansol/zentube/internal/entities"

type YouTubeClient interface {
	Search(query string, maxResults int64) ([]entities.Video, error)
}
