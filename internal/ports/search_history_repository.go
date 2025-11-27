package ports

import (
	"context"

	"github.com/uiansol/zentube/internal/entities"
)

type SearchHistoryRepository interface {
	Save(ctx context.Context, history *entities.SearchHistory) error
	GetLast(ctx context.Context, limit int) ([]entities.SearchHistory, error)
}
