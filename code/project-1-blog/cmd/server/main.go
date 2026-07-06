package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/go-course/project-1-blog/internal/handler"
	"github.com/go-course/project-1-blog/internal/repository"
	"github.com/go-course/project-1-blog/internal/service"
)

func main() {
	// Структурный логгер
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Слои
	repo := repository.NewMemory()
	svc := service.NewPost(repo, logger)
	h := handler.NewPost(svc)

	// Роутер
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	setupRoutes(r, h)

	// Сервер
	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Запуск в горутине
	go func() {
		slog.Info("сервер запущен", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("ошибка сервера", "err", err)
			os.Exit(1)
		}
	}()

	// Ожидание сигнала
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	slog.Info("выключаем сервер", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("сервер остановлен принудительно", "err", err)
		os.Exit(1)
	}

	slog.Info("сервер остановлен корректно")
}

func setupRoutes(r chi.Router, h *handler.PostHandler) {
	r.Group(func(r chi.Router) {
		r.Get("/posts", h.List)
		r.Get("/posts/{id:[0-9]+}", h.Get)
	})

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Post("/posts", h.Create)
		r.Put("/posts/{id:[0-9]+}", h.Update)
		r.Delete("/posts/{id:[0-9]+}", h.Delete)
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			handler.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	})
}
