package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Status — статус поста.
type Status int

const (
	StatusDraft     Status = 0
	StatusPublished Status = 1
)

func (s Status) MarshalJSON() ([]byte, error) {
	switch s {
	case StatusDraft:
		return json.Marshal("draft")
	case StatusPublished:
		return json.Marshal("published")
	default:
		return json.Marshal("unknown")
	}
}

func (s *Status) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	switch strings.ToLower(str) {
	case "draft":
		*s = StatusDraft
	case "published":
		*s = StatusPublished
	default:
		*s = StatusDraft
	}
	return nil
}

// Post — модель поста блога.
type Post struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Validate проверяет бизнес-правила.
func (p *Post) Validate() error {
	if strings.TrimSpace(p.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if len(p.Title) > 200 {
		return fmt.Errorf("title too long: %d chars (max 200)", len(p.Title))
	}
	if strings.TrimSpace(p.Body) == "" {
		return fmt.Errorf("body is required")
	}
	return nil
}

// ErrorResponse — структура ошибки API.
type ErrorResponse struct {
	Error string `json:"error"`
}
