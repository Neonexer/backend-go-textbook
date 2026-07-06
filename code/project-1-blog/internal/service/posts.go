package service

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/go-course/project-1-blog/internal/model"
)

// PostRepository — интерфейс хранилища постов.
type PostRepository interface {
	FindAll() []model.Post
	FindByID(id int) (model.Post, bool)
	Create(p model.Post) model.Post
	Update(id int, p model.Post) (model.Post, bool)
	Delete(id int) bool
}

// PostService — бизнес-логика постов.
type PostService struct {
	repo   PostRepository
	logger *slog.Logger
}

// NewPost создаёт сервис с репозиторием и логгером.
func NewPost(repo PostRepository, logger *slog.Logger) *PostService {
	return &PostService{repo: repo, logger: logger}
}

func (s *PostService) List() []model.Post {
	return s.repo.FindAll()
}

func (s *PostService) Get(id int) (model.Post, error) {
	p, ok := s.repo.FindByID(id)
	if !ok {
		return model.Post{}, fmt.Errorf("post not found")
	}
	return p, nil
}

func (s *PostService) Create(title, body string) (model.Post, error) {
	p := model.Post{
		Title:     title,
		Body:      body,
		Status:    model.StatusDraft,
		CreatedAt: time.Now(),
	}
	if err := p.Validate(); err != nil {
		s.logger.Warn("валидация не пройдена", "title", title, "err", err)
		return model.Post{}, err
	}
	created := s.repo.Create(p)
	s.logger.Info("пост создан", "post_id", created.ID)
	return created, nil
}

func (s *PostService) Update(id int, title, body string) (model.Post, error) {
	p, ok := s.repo.FindByID(id)
	if !ok {
		return model.Post{}, fmt.Errorf("post not found")
	}
	p.Title = title
	p.Body = body
	if err := p.Validate(); err != nil {
		s.logger.Warn("валидация не пройдена", "post_id", id, "err", err)
		return model.Post{}, err
	}
	updated, ok := s.repo.Update(id, p)
	if !ok {
		return model.Post{}, fmt.Errorf("post not found")
	}
	s.logger.Info("пост обновлён", "post_id", id)
	return updated, nil
}

func (s *PostService) Delete(id int) error {
	if !s.repo.Delete(id) {
		return fmt.Errorf("post not found")
	}
	s.logger.Info("пост удалён", "post_id", id)
	return nil
}
