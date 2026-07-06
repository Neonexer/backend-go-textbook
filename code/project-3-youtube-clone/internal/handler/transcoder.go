package handler

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/go-course/project-3-youtube-clone/internal/model"
	"github.com/segmentio/kafka-go"
)

// Transcoder — Kafka consumer: слушает VideoUploaded и запускает транскодирование.
type Transcoder struct {
	reader *kafka.Reader
	logger *slog.Logger
}

func NewTranscoder(reader *kafka.Reader, logger *slog.Logger) *Transcoder {
	return &Transcoder{reader: reader, logger: logger}
}

func (t *Transcoder) Run(ctx context.Context) error {
	for {
		msg, err := t.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			t.logger.Error("ошибка чтения из Kafka", "err", err)
			continue
		}

		var event model.VideoUploaded
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			t.logger.Error("невалидное событие", "err", err)
			continue
		}

		t.logger.Info("транскодирование", "video_id", event.VideoID, "title", event.Title)
		// Здесь: реальное транскодирование (ffmpeg или облачный сервис)
		t.logger.Info("транскодирование завершено", "video_id", event.VideoID)
	}
}
