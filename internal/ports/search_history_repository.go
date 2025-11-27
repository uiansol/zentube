package ports

import "github.com/uiansol/zentube/internal/entities"

type SearchHistoryRepository interface {
	Save(history *entities.SearchHistory) error
	GetLast(limit int) ([]entities.SearchHistory, error)
}
