---
title: "Middleware"
sidebar_position: 3
---

import Quiz from '@site/src/components/Quiz';

# Middleware

Middleware — это функции, которые оборачивают HTTP-обработчики и выполняются **до** и **после** них. В Go middleware реализуется через паттерн `func(http.Handler) http.Handler`. Chi даёт удобный способ вешать middleware на любой уровень: глобально, на группу, на конкретный маршрут.

## Как работает middleware в Go

Middleware в Go — это функция, которая принимает `http.Handler` и возвращает новый `http.Handler`:

```go
func myMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // ДО обработчика
        fmt.Println("before request")

        next.ServeHTTP(w, r) // вызываем следующий обработчик

        // ПОСЛЕ обработчика
        fmt.Println("after request")
    })
}
```

Цепочка middleware образует «луковицу» (onion): запрос проходит через слои снаружи внутрь до обработчика, а ответ — обратно.

## Встроенный middleware в chi

Chi идёт с набором готовых middleware в `chi/v5/middleware`:

```go
import "github.com/go-chi/chi/v5/middleware"
```

### Logger — логирование запросов

```go
r.Use(middleware.Logger)
```

Выводит каждый запрос в формате:

```
200 POST /posts 2.3ms
```

### Recoverer — защита от паники

```go
r.Use(middleware.Recoverer)
```

Если обработчик запаникует, `Recoverer` поймает панику, залогирует стектрейс и вернёт 500 вместо падения всего сервера. Всегда ставь первым или вторым в цепочке.

### RequestID — уникальный ID для каждого запроса

```go
r.Use(middleware.RequestID)
```

Добавляет заголовок `X-Request-Id` и кладёт ID в контекст. Полезно для отладки в логах.

### RealIP — правильный IP за прокси

```go
r.Use(middleware.RealIP)
```

Доверяет заголовкам `X-Forwarded-For` / `X-Real-IP`. Важно, если сервер стоит за nginx или load balancer.

### Timeout — таймаут на запрос

```go
r.Use(middleware.Timeout(30 * time.Second))
```

Если обработчик не уложился в 30 секунд — клиент получает 504.

### Heartbeat — health check

```go
r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("ok"))
})
```

Не встроенный, но общепринятый паттерн. Простой эндпоинт для проверки живости сервера.

## Порядок middleware имеет значение

Middleware выполняются в порядке регистрации. Правильный порядок для продакшен-сервера:

```go
r.Use(middleware.RequestID)  // 1. ID запроса
r.Use(middleware.RealIP)     // 2. Правильный IP
r.Use(middleware.Logger)     // 3. Логирование
r.Use(middleware.Recoverer)  // 4. Защита от паники
r.Use(middleware.Timeout(30 * time.Second)) // 5. Таймаут
```

:::warning Recoverer до Logger
Если поставить `Logger` до `Recoverer`, то при панике запрос не будет залогирован. Всегда `Recoverer` до `Logger`.
:::

## Пишем свой middleware

### Логирующий middleware

```go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        fmt.Printf("→ %s %s\n", r.Method, r.URL.Path)

        next.ServeHTTP(w, r)

        fmt.Printf("← %s %s (%s)\n", r.Method, r.URL.Path, time.Since(start))
    })
}
```

### Auth middleware

```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
            return
        }
        // Здесь проверили бы токен
        next.ServeHTTP(w, r)
    })
}
```

Если middleware не вызывает `next.ServeHTTP`, цепочка обрывается — клиент получает ответ прямо от middleware.

## Middleware на уровне групп и маршрутов

В chi можно применить middleware выборочно:

```go
r := chi.NewRouter()

// Глобально — для всех маршрутов
r.Use(middleware.Logger)
r.Use(middleware.Recoverer)

// Публичные маршруты — без авторизации
r.Group(func(r chi.Router) {
    r.Get("/posts", listPosts)
    r.Get("/posts/{id}", getPost)
})

// Приватные — с авторизацией
r.Group(func(r chi.Router) {
    r.Use(authMiddleware)

    r.Post("/posts", createPost)
    r.Put("/posts/{id}", updatePost)
    r.Delete("/posts/{id}", deletePost)
})
```

:::tip Group vs Route
`Group` создаёт новый роутер, но монтирует его без префикса. Это идеально для случаев, когда у тебя разные middleware на одном уровне пути, но без изменения URL.
:::

## Собираем всё вместе

Финальный `main.go` нашего блога:

```go
func main() {
    r := chi.NewRouter()

    // Глобальный middleware
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.Timeout(30 * time.Second))

    // Health check
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("ok"))
    })

    // REST API
    r.Route("/posts", func(r chi.Router) {
        r.Get("/", listPosts)
        r.Get("/{id}", getPost)

        // Только для создания/редактирования
        r.Group(func(r chi.Router) {
            r.Use(authMiddleware)
            r.Post("/", createPost)
            r.Put("/{id}", updatePost)
            r.Delete("/{id}", deletePost)
        })
    })

    server := &http.Server{
        Addr:    ":8080",
        Handler: r,
        // ... таймауты
    }
    server.ListenAndServe()
}
```

## Ключевые выводы

1. Middleware в Go — это `func(http.Handler) http.Handler`. Никакой магии.
2. Порядок важен: RequestID → RealIP → Recoverer → Logger → Timeout
3. `Group` позволяет применить middleware без изменения URL
4. Если middleware не вызывает `next.ServeHTTP` — цепочка обрывается
5. Всегда ставь `Recoverer` в продакшене

В следующей главе разберём JSON-сериализацию: `encoding/json`, кастомные поля и валидацию.

---

## Проверь себя

<Quiz
  quizId="03-middleware"
  questions={[
    {
      id: "q1",
      question: "Как middleware обрывает цепочку обработки запроса?",
      options: [
        "Вызывает panic()",
        "Не вызывает next.ServeHTTP(w, r) и возвращает ответ сам",
        "Вызывает r.Context().Done()",
        "Middleware не может оборвать цепочку"
      ],
      correctIndex: 1,
      explanation: "Middleware сам решает, вызывать ли следующий обработчик. Если не вызвать next.ServeHTTP — цепочка обрывается и middleware должен сам записать ответ."
    },
    {
      id: "q2",
      question: "Почему Recoverer должен идти до Logger?",
      options: [
        "Это неважно, можно в любом порядке",
        "Чтобы при панике запрос всё равно был залогирован",
        "Recoverer должен быть последним в цепочке",
        "Logger ломается при панике"
      ],
      correctIndex: 1,
      explanation: "Если Logger стоит до Recoverer, то при панике Logger не успеет отработать — запрос упадёт раньше. Recoverer первым ловит панику и передаёт управление дальше для корректного логирования."
    },
    {
      id: "q3",
      question: "Чем Group отличается от Route в chi?",
      options: [
        "Ничем, это синонимы",
        "Group создаёт новый роутер без префикса, Route группирует с префиксом",
        "Group только для middleware, Route для маршрутов",
        "Group медленнее, Route быстрее"
      ],
      correctIndex: 1,
      explanation: "Group создаёт отдельный chi.Router и монтирует его по тому же пути. Это нужно, когда хочешь разные middleware на одном уровне без изменения URL."
    }
  ]}
/>
