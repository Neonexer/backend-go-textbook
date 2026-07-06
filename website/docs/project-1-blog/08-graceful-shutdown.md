---
title: "Graceful Shutdown"
sidebar_position: 8
---

import Quiz from '@site/src/components/Quiz';

# Graceful Shutdown

`server.ListenAndServe()` блокирует main навсегда. При `Ctrl+C` (SIGINT) или `kill` (SIGTERM) сервер падает мгновенно, обрывая активные запросы. Это плохо: клиент получает ошибку, данные могут быть потеряны. Graceful shutdown решает эту проблему — сервер перестаёт принимать новые запросы, дожидается завершения текущих и только потом выключается.

## Сигналы ОС

Операционная система посылает процессу сигналы для управления:

| Сигнал | Источник | Значение |
|--------|---------|----------|
| `SIGINT` | Ctrl+C в терминале | Пользователь хочет остановить |
| `SIGTERM` | `kill <pid>`, Docker, K8s | Платформа просит завершиться |

Go ловит сигналы через `os/signal`:

```go
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
sig := <-sigCh // блокируется, пока не придёт сигнал
fmt.Printf("получен сигнал %v, завершаемся\n", sig)
```

## Базовая схема graceful shutdown

```go
func main() {
    // ... настройка роутера и middleware

    server := &http.Server{
        Addr:    ":8080",
        Handler: r,
    }

    // Запускаем сервер в горутине
    go func() {
        fmt.Println("сервер запущен на :8080")
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            fmt.Printf("ошибка сервера: %v\n", err)
        }
    }()

    // Ждём сигнал
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    fmt.Println("выключаем сервер...")

    // Даём активным запросам завершиться
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        fmt.Printf("сервер остановлен принудительно: %v\n", err)
    }

    fmt.Println("сервер остановлен")
}
```

Разберём:

1. Сервер запускается в отдельной горутине — `main` не блокируется
2. `signal.Notify` направляет OS-сигналы в канал
3. `<-quit` блокируется до получения сигнала
4. `server.Shutdown(ctx)` перестаёт принимать новые запросы и ждёт завершения текущих
5. Если за 30 секунд запросы не завершились — `context.WithTimeout` отменяет ожидание

:::tip ListenAndServe и ErrServerClosed
`ListenAndServe()` возвращает `http.ErrServerClosed` после `Shutdown()`. Это нормально, не ошибка — поэтому проверяем `err != http.ErrServerClosed`.
:::

## Graceful shutdown с логгером

Добавим информативные логи:

```go
slog.Info("сервер запущен", "addr", ":8080")

// ... горутина с сервером ...

sig := <-quit
slog.Info("выключаем сервер", "signal", sig.String())

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := server.Shutdown(ctx); err != nil {
    slog.Error("сервер остановлен принудительно", "err", err)
} else {
    slog.Info("сервер остановлен корректно")
}
```

## Полный main.go

Собираем всё вместе:

```go
package main

import (
    "context"
    "fmt"
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
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
    slog.SetDefault(logger)

    repo := repository.NewMemory()
    svc := service.NewPost(repo, logger)
    h := handler.NewPost(svc)

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

    server := &http.Server{
        Addr:         ":8080",
        Handler:      r,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  120 * time.Second,
    }

    go func() {
        slog.Info("сервер запущен", "addr", server.Addr)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            slog.Error("ошибка сервера", "err", err)
            os.Exit(1)
        }
    }()

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
```

## Как это работает в контейнерах

Docker и Kubernetes при остановке пода посылают `SIGTERM`. Если приложение не завершилось за `terminationGracePeriodSeconds` (по умолчанию 30 секунд) — посылают `SIGKILL` (мгновенное убийство). Поэтому:

- `server.Shutdown(ctx)` должен иметь таймаут **меньше**, чем terminationGracePeriodSeconds
- 30 секунд на shutdown и 30 секунд grace period в K8s — ок
- Health check должен возвращать `503` во время shutdown, чтобы load balancer перестал слать трафик

## Ключевые выводы

1. `signal.Notify` ловит SIGINT/SIGTERM — сервер знает, когда пора выключаться
2. `server.Shutdown(ctx)` ждёт завершения активных запросов, не принимая новых
3. `context.WithTimeout` страхует от зависших запросов
4. Сервер запускается в горутине, main ждёт сигнала
5. В K8s таймаут shutdown должен быть меньше terminationGracePeriodSeconds

---

Поздравляю! Ты написал production-ready REST API на Go: chi-роутер, middleware, JSON-сериализация, слоистая архитектура, тесты, логирование и graceful shutdown. В Проекте 2 добавим PostgreSQL, GORM и аутентификацию.

---

## Проверь себя

<Quiz
  quizId="08-graceful-shutdown"
  questions={[
    {
      id: "q1",
      question: "Почему нельзя просто убить процесс при остановке сервера?",
      options: [
        "Можно, это самый быстрый способ",
        "Активные запросы оборвутся — клиенты получат ошибки, данные могут потеряться",
        "Go запрещает убивать процессы",
        "Это нарушает стандарт HTTP"
      ],
      correctIndex: 1,
      explanation: "При мгновенном убийстве все незавершённые запросы обрываются. Клиент получает connection reset, транзакции не завершены. Graceful shutdown даёт запросам время завершиться."
    },
    {
      id: "q2",
      question: "Зачем server.Shutdown() принимает context?",
      options: [
        "Для передачи request_id",
        "Чтобы ограничить время ожидания — если запросы не завершились за N секунд, сервер всё равно выключается",
        "Для логирования",
        "Context не нужен, можно передать nil"
      ],
      correctIndex: 1,
      explanation: "Context с таймаутом — страховка от вечно висящих запросов. Если через 30 секунд соединения всё ещё открыты, Shutdown завершится с ошибкой, а не зависнет навсегда."
    },
    {
      id: "q3",
      question: "Почему сервер запускается в отдельной горутине?",
      options: [
        "Чтобы использовать несколько ядер CPU",
        "ListenAndServe блокирует горутину — main должен освободиться для ожидания сигнала",
        "Это требование chi",
        "Горутины быстрее обрабатывают запросы"
      ],
      correctIndex: 1,
      explanation: "ListenAndServe() — блокирующий вызов. Если запустить его в main, код после него никогда не выполнится. В горутине main продолжает работу и может ждать сигнал."
    }
  ]}
/>
