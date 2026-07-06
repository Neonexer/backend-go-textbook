---
title: "Маршрутизация с chi"
sidebar_position: 2
---

import Quiz from '@site/src/components/Quiz';

# Маршрутизация с chi

В первой главе мы использовали `DefaultServeMux` — стандартный мультиплексор Go. Для REST API этого недостаточно: нужны URL-параметры (`/posts/{id}`), группы маршрутов, method-based роутинг. Подключаем `chi` — лёгкий, идиоматичный роутер, который расширяет `net/http`, а не заменяет его.

## Почему chi

Стандартный `http.ServeMux` до Go 1.22 не поддерживал параметры в пути и method-based роутинг. На Go 1.22+ часть проблем решена, но `chi` всё ещё даёт больше:

- **URL-параметры** — `{id}` и regex: `{id:[0-9]+}`
- **Группы маршрутов** — общий префикс и middleware для группы
- **Method-based роутинг** — `r.Get()`, `r.Post()` вместо `switch r.Method`
- **Middleware на уровне роута** — цепочки обработчиков
- **Саброутеры** — композиция маршрутов
- **Совместимость** — chi-обработчик это `http.Handler`, работает со всей экосистемой

## Установка

```bash
go get github.com/go-chi/chi/v5
```

## Минимальный chi-сервер

Перепишем сервер из первой главы на chi:

```go
package main

import (
    "fmt"
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"
)

func main() {
    r := chi.NewRouter()

    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Привет, бэкенд!")
    })

    server := &http.Server{
        Addr:         ":8080",
        Handler:      r, // chi-роутер реализует http.Handler
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  120 * time.Second,
    }

    fmt.Println("Сервер запущен на http://localhost:8080")
    server.ListenAndServe()
}
```

Главное отличие: `chi.NewRouter()` передан в `server.Handler`. Chi-роутер — это просто `http.Handler`, поэтому он встраивается без магии.

## URL-параметры

В REST API нам нужны параметры в пути: `/posts/42`, `/users/ivan`. Chi использует синтаксис `{name}`:

```go
r.Get("/posts/{id}", func(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    fmt.Fprintf(w, "Запрошен пост #%s", id)
})
```

`chi.URLParam(r, "id")` извлекает значение из пути. `/posts/42` → `id = "42"`.

### Regex-ограничения

Можно ограничить формат параметра регуляркой:

```go
r.Get("/posts/{id:[0-9]+}", postByID)
r.Get("/posts/{slug:[a-z-]+}", postBySlug)
```

Первым сработает тот маршрут, чей паттерн совпадёт.

:::tip Числовой ID или slug?
Используй regex-ограничения, чтобы развести `/posts/42` и `/posts/my-awesome-post` в разные обработчики. Chi проверяет маршруты в порядке регистрации и выбирает первый подходящий.
:::

## Method-based роутинг

Вместо `switch r.Method` — отдельные методы роутера:

```go
r.Get("/posts", listPosts)       // GET    /posts
r.Post("/posts", createPost)     // POST   /posts
r.Get("/posts/{id}", getPost)    // GET    /posts/42
r.Put("/posts/{id}", updatePost) // PUT    /posts/42
r.Delete("/posts/{id}", deletePost) // DELETE /posts/42
r.Patch("/posts/{id}", patchPost)   // PATCH  /posts/42
```

Метод `Head` обрабатывается автоматически — chi генерирует HEAD из GET.

Для OPTIONS можно зарегистрировать явно:

```go
r.Options("/posts", optionsHandler)
```

## Группы маршрутов

Когда маршрутов много, появляется дублирование префиксов. Группы решают эту проблему:

```go
r.Route("/posts", func(r chi.Router) {
    r.Get("/", listPosts)         // GET /posts
    r.Post("/", createPost)       // POST /posts

    r.Route("/{id}", func(r chi.Router) {
        r.Get("/", getPost)       // GET /posts/{id}
        r.Put("/", updatePost)    // PUT /posts/{id}
        r.Delete("/", deletePost) // DELETE /posts/{id}
    })
})
```

:::info Разница между Route и Group
- `Route` — группирует маршруты с общим префиксом
- `Group` — то же, но создаёт **новый роутер**, который можно монтировать

В большинстве случаев `Route` достаточно. `Group` нужен, когда хочешь вынести подмаршруты в отдельную функцию или middleware.
:::

## Саброутеры (Mount)

Для больших проектов логично вынести группы маршрутов в отдельные функции:

```go
func postsRouter() chi.Router {
    r := chi.NewRouter()
    r.Get("/", listPosts)
    r.Post("/", createPost)
    r.Get("/{id}", getPost)
    r.Put("/{id}", updatePost)
    r.Delete("/{id}", deletePost)
    return r
}

func main() {
    r := chi.NewRouter()
    r.Mount("/posts", postsRouter())    // монтируем /posts/*
    r.Mount("/users", usersRouter())    // монтируем /users/*
    server.ListenAndServe()
}
```

`Mount` — это композиция. Каждый саброутер независим: свои middleware, свои обработчики.

## Middleware на уровне роута

Chi позволяет вешать middleware на любой уровень — глобально, на группу, на конкретный маршрут:

```go
// Глобальный middleware
r.Use(loggingMiddleware)
r.Use(recoverMiddleware)

// Middleware только для /posts/*
r.Route("/posts", func(r chi.Router) {
    r.Use(authMiddleware) // все обработчики внутри требуют аутентификации

    r.Get("/", listPosts)
    r.Post("/", createPost) // только авторизованные
})
```

Middleware в chi — это обычная функция `func(http.Handler) http.Handler`. О них подробно в следующей главе.

## Наш блог: структура маршрутов

Спроектируем REST API для блога:

| Метод | Путь | Обработчик |
|-------|------|-----------|
| `GET` | `/posts` | Список постов |
| `POST` | `/posts` | Создать пост |
| `GET` | `/posts/{id}` | Получить пост |
| `PUT` | `/posts/{id}` | Обновить пост |
| `DELETE` | `/posts/{id}` | Удалить пост |

В коде это выглядит так:

```go
func setupRoutes(r chi.Router) {
    r.Route("/posts", func(r chi.Router) {
        r.Get("/", listPosts)
        r.Post("/", createPost)
        r.Route("/{id}", func(r chi.Router) {
            r.Get("/", getPost)
            r.Put("/", updatePost)
            r.Delete("/", deletePost)
        })
    })
}
```

Пока обработчики будут возвращать JSON-заглушки. Реальную логику напишем после глав про JSON и структуру проекта.

## Ключевые выводы

1. Chi расширяет `net/http`, а не заменяет — это тот же `http.Handler`
2. `chi.URLParam(r, "id")` для параметров пути — просто и без рефлексии
3. `Route` для группировки, `Mount` для композиции саброутеров
4. Middleware вешается на любой уровень: глобально → группа → маршрут

В следующей главе разберём middleware подробно: логирование, panic recovery, CORS и аутентификацию.

---

## Проверь себя

<Quiz
  quizId="02-routing-chi"
  questions={[
    {
      id: "q1",
      question: "Чем chi-роутер отличается от стандартного ServeMux?",
      options: [
        "Это совершенно другой протокол, не HTTP",
        "Chi заменяет весь net/http на свой движок",
        "Chi реализует http.Handler и расширяет стандартную маршрутизацию параметрами и группами",
        "Chi быстрее, но не совместим с middleware из коробки"
      ],
      correctIndex: 2,
      explanation: "Chi сознательно остаётся в экосистеме net/http. Его роутер — это http.Handler, совместимый со всей стандартной библиотекой."
    },
    {
      id: "q2",
      question: "Как извлечь параметр {id} из пути /posts/42?",
      options: [
        "r.URL.Query().Get(\"id\")",
        "chi.URLParam(r, \"id\")",
        "r.PathValue(\"id\")",
        "r.Param(\"id\")"
      ],
      correctIndex: 1,
      explanation: "chi.URLParam(r, \"id\") — основной способ. r.PathValue работает в Go 1.22+, но chi.URLParam работает на всех версиях и поддерживает regex-ограничения."
    },
    {
      id: "q3",
      question: "В чём разница между Route и Mount в chi?",
      options: [
        "Никакой разницы, это синонимы",
        "Route — для GET/POST, Mount — для PUT/DELETE",
        "Route группирует маршруты с префиксом, Mount монтирует отдельный chi-роутер",
        "Route работает только с middleware, Mount без middleware"
      ],
      correctIndex: 2,
      explanation: "Route группирует внутри текущего роутера, Mount подключает внешний chi.Router. Mount даёт полную изоляцию middleware и обработчиков."
    },
    {
      id: "q4",
      question: "Можно ли использовать стандартные http.Handler с chi?",
      options: [
        "Нет, chi требует свои типы обработчиков",
        "Да, chi полностью совместим с http.Handler",
        "Только через адаптер chi.ToStdHandler",
        "Да, но только для GET-запросов"
      ],
      correctIndex: 1,
      explanation: "Chi-обработчики — это обычные http.Handler. Любой middleware или обработчик из экосистемы Go работает с chi без адаптеров."
    }
  ]}
/>
