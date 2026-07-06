package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/go-course/project-3-youtube-clone/internal/model"
	"github.com/segmentio/kafka-go"
)

// VideoRepo — интерфейс хранилища видео.
type VideoRepo interface {
	FindByID(ctx context.Context, id string) (model.Video, error)
	FindAll(ctx context.Context, pageSize int, pageToken string) ([]model.Video, string, error)
	Create(ctx context.Context, v model.Video) (model.Video, error)
}

// VideoService — бизнес-логика видео.
type VideoService struct {
	repo       VideoRepo
	writer     *kafka.Writer
	logger     *slog.Logger
}

func NewVideo(repo VideoRepo, writer *kafka.Writer, logger *slog.Logger) *VideoService {
	return &VideoService{repo: repo, writer: writer, logger: logger}
}

func (s *VideoService) Get(ctx context.Context, id string) (model.Video, error) {
	v, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return model.Video{}, fmt.Errorf("get video: %w", err)
	}
	return v, nil
}

func (s *VideoService) List(ctx context.Context, pageSize int, pageToken string) ([]model.Video, string, error) {
	return s.repo.FindAll(ctx, pageSize, pageToken)
}

func (s *VideoService) Upload(ctx context.Context, title, description, url string) (model.Video, error) {
	v := model.Video{Title: title, Description: description, URL: url}
	created, err := s.repo.Create(ctx, v)
	if err != nil {
		return model.Video{}, fmt.Errorf("create video: %w", err)
	}

	// Публикуем событие в Kafka — транскодирование запустится асинхронно
	event := model.VideoUploaded{VideoID: created.ID, Title: title, URL: url}
	data, _ := json.Marshal(event)
	if err := s.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(created.ID),
		Value: data,
	}); err != nil {
		s.logger.Error("ошибка публикации VideoUploaded", "err", err, "video_id", created.ID)
		// не фейлим запрос — видео сохранено, событие можно переотправить
	}

	s.logger.Info("видео загружено", "video_id", created.ID)
	return created, nil
}
