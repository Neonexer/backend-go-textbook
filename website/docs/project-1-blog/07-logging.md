---
title: "Логирование с slog"
sidebar_position: 7
---

import Quiz from '@site/src/components/Quiz';

# Логирование с slog

До сих пор мы использовали `fmt.Println`. В продакшене этого недостаточно: нужны уровни (debug, info, warn, error), структурированные поля и возможность отправить логи в систему агрегации. С Go 1.21 в стандартную библиотеку вошёл `log/slog` — структурный логгер, который решает эти задачи.

## Почему slog, а не сторонние библиотеки

До slog стандартом де-факто были `zap` (Uber) и `zerolog`. Они быстрее и функциональнее, но:

- **slog в стандартной библиотеке** — не тянет зависимостей
- **Единый интерфейс** — все новые библиотеки переходят на slog-совместимость
- **Достаточная производительность** — для 95% проектов хватает

Если проект вырастет до миллионов запросов в секунду — заменишь handler на zap, не меняя код логирования.

## Первый structured log

```go
import "log/slog"

func main() {
    slog.Info("сервер запущен", "port", 8080, "env", "production")
}
```

Вывод:

```
2026/07/06 18:00:00 INFO сервер запущен port=8080 env=production
```

Это текстовый формат по умолчанию. В продакшене лучше JSON:

```go
func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    slog.SetDefault(logger)

    slog.Info("сервер запущен", "port", 8080)
}
```

Вывод:

```json
{"time":"2026-07-06T18:00:00Z","level":"INFO","msg":"сервер запущен","port":8080}
```

JSON удобно парсить в Loki, ELK, CloudWatch и других системах агрегации логов.

## Уровни логирования

```go
slog.Debug("детальная инфа для отладки", "request_id", "abc")
slog.Info("штатная операция", "user_id", 42)
slog.Warn("что-то подозрительное", "retry", 3)
slog.Error("ошибка сохранения", "err", err)
```

Уровни в порядке возрастания: Debug → Info → Warn → Error.

### Настройка уровня

```go
handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo, // Debug-сообщения не выводятся
})
```

В dev-окружении ставь `LevelDebug`, в production — `LevelInfo`.

## Структурированные поля

Вместо того чтобы склеивать строки — передавай пары ключ-значение:

```go
// ❌ Плохо
slog.Info(fmt.Sprintf("пользователь %d создал пост %d", userID, postID))

// ✅ Хорошо
slog.Info("пост создан",
    "user_id", userID,
    "post_id", postID,
    "title", title,
)
```

Структурированные поля можно фильтровать и агрегировать: «покажи все логи где `user_id=42` и `level=ERROR`».

## Логирование в middleware

Заменим `middleware.Logger` на кастомный слоггер:

```go
func loggingMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()

            // Оборачиваем ResponseWriter чтобы поймать статус-код
            ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

            next.ServeHTTP(ww, r)

            logger.Info("request",
                "method", r.Method,
                "path", r.URL.Path,
                "status", ww.Status(),
                "duration", time.Since(start).String(),
            )
        })
    }
}
```

`middleware.NewWrapResponseWriter` из chi перехватывает `WriteHeader()` и запоминает статус-код.

## Добавление контекста в логи

RequestID должен быть в каждом логе — так можно связать все сообщения одного запроса:

```go
func loggingMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            reqID := middleware.GetReqID(r.Context())
            logger := logger.With("request_id", reqID)

            // ... логируем через logger, а не глобальный slog
            logger.Info("request", "method", r.Method)
            next.ServeHTTP(w, r)
        })
    }
}
```

`logger.With` создаёт дочерний логгер с предустановленными полями. Все сообщения этого логгера будут содержать `request_id`.

## Логирование в сервисном слое

Сервис принимает логгер через конструктор:

```go
type PostService struct {
    repo   PostRepository
    logger *slog.Logger
}

func NewPost(repo PostRepository, logger *slog.Logger) *PostService {
    return &PostService{repo: repo, logger: logger}
}

func (s *PostService) Create(title, body string) (model.Post, error) {
    p := model.Post{Title: title, Body: body, Status: model.StatusDraft}

    if err := p.Validate(); err != nil {
        s.logger.Warn("валидация не пройдена", "title", title, "err", err)
        return model.Post{}, err
    }

    created := s.repo.Create(p)
    s.logger.Info("пост создан", "post_id", created.ID)
    return created, nil
}
```

:::tip Не логируй каждый чих
Info — значимые операции (создание, удаление). Debug — детали. Warn — подозрительное. Error — ошибка, требующая внимания. Если логировать каждый вызов `FindByID` — логи станут бесполезным шумом.
:::

## Сборка в main.go

```go
func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
    slog.SetDefault(logger)

    repo := repository.NewMemory()
    svc := service.NewPost(repo, logger)
    h := handler.NewPost(svc)

    // ...
    slog.Info("сервер запущен", "port", 8080)
    server.ListenAndServe()
}
```

## Ключевые выводы

1. `log/slog` — стандартный структурированный логгер, без внешних зависимостей
2. JSON для продакшена, текст для разработки
3. `logger.With` для добавления контекста (request_id, user_id)
4. Уровни: Debug < Info < Warn < Error — фильтруются в `HandlerOptions`
5. Сервис получает логгер через конструктор — не через глобальную переменную

В последней главе проекта настроим graceful shutdown — чтобы сервер завершался корректно, а не обрывал запросы.

---

## Проверь себя

<Quiz
  quizId="07-logging"
  questions={[
    {
      id: "q1",
      question: "Зачем использовать структурированные поля вместо fmt.Sprintf в логах?",
      options: [
        "fmt.Sprintf быстрее работает",
        "Структурированные поля можно фильтровать и агрегировать в системах сбора логов",
        "slog не поддерживает строки",
        "Это просто стилистическое предпочтение"
      ],
      correctIndex: 1,
      explanation: "Когда логи в JSON, можно найти все записи с user_id=42 или посчитать количество ERROR за последний час. Со строками это был бы grep."
    },
    {
      id: "q2",
      question: "Что делает logger.With('request_id', id)?",
      options: [
        "Отправляет лог с полем request_id",
        "Создаёт дочерний логгер, который добавляет request_id ко всем последующим сообщениям",
        "Фильтрует логи по request_id",
        "Меняет глобальный уровень логирования"
      ],
      correctIndex: 1,
      explanation: "With создаёт новый логгер с предустановленными полями. Все вызовы этого логгера будут автоматически включать request_id — не нужно передавать его в каждом slog.Info."
    },
    {
      id: "q3",
      question: "Какой уровень логирования подходит для продакшена?",
      options: [
        "Debug — чем больше логов, тем лучше",
        "Info — значимые операции, Warn для подозрительного, Error для ошибок",
        "Только Error — всё остальное шум",
        "Уровни не важны, логируй всё подряд"
      ],
      correctIndex: 1,
      explanation: "Info покрывает ключевые операции. Debug включается только при отладке. Warn сигнализирует о потенциальных проблемах. Error — то, что требует немедленного внимания."
    }
  ]}
/>
