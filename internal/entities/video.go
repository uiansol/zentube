package entities

import "time"

type Video struct {
	ID          string
	Title       string
	Channel     string
	PublishedAt time.Time
	Thumbnail   string
}
