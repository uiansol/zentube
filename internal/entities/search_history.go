package entities

import "time"

type SearchHistory struct {
	ID        int64     `db:"id"`
	Query     string    `db:"query"`
	Results   int       `db:"results"`
	CreatedAt time.Time `db:"created_at"`
}
