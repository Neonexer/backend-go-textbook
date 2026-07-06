package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"

	"github.com/go-course/project-3-youtube-clone/internal/handler"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Kafka writer (producer)
	writer := &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "video.uploaded",
		Balancer: &kafka.Hash{},
	}
	defer writer.Close()

	// Запускаем Transcoder (Kafka consumer) в фоне
	transcoderReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "video.uploaded",
		GroupID: "transcoder",
	})
	defer transcoderReader.Close()

	transcoder := handler.NewTranscoder(transcoderReader, logger)
	go func() {
		if err := transcoder.Run(context.Background()); err != nil {
			slog.Error("transcoder остановлен", "err", err)
		}
	}()

	// gRPC сервер
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		slog.Error("не удалось запустить listener", "err", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	// video.RegisterVideoServiceServer(grpcServer, videoHandler) — после генерации proto

	go func() {
		slog.Info("gRPC сервер запущен", "addr", ":50051")
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("ошибка gRPC", "err", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	slog.Info("выключаем сервер", "signal", sig.String())
	grpcServer.GracefulStop()
	_ = transcoderReader.Close()
	_ = writer.Close()

	slog.Info("сервер остановлен корректно")
}
