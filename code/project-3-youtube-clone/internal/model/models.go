package model

import "time"

// Video — модель видео.
type Video struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	Views       int64     `json:"views"`
	AuthorID    string    `json:"author_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// VideoUploaded — событие Kafka.
type VideoUploaded struct {
	VideoID string `json:"video_id"`
	Title   string `json:"title"`
	URL     string `json:"url"`
}
