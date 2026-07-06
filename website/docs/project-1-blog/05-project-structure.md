---
title: "Структура проекта"
sidebar_position: 5
---

import Quiz from '@site/src/components/Quiz';

# Структура проекта

До сих пор весь код лежал в `main.go`. Это работает для демо, но не для реального проекта. В этой главе разберём слоёную архитектуру: **handler → service → repository** — и организуем код так, чтобы его можно было тестировать, расширять и поддерживать.

## Зачем нужны слои

Представь, что код делает три вещи одновременно:

```go
func createPost(w http.ResponseWriter, r *http.Request) {
    // 1. Парсит JSON из HTTP-запроса
    var p Post
    json.NewDecoder(r.Body).Decode(&p)

    // 2. Бизнес-логика: валидация, slug, уведомления
    if p.Title == "" { ... }

    // 3. Сохраняет в «базу» (пока in-memory)
    posts = append(posts, p)

    // 4. Сериализует ответ
    json.NewEncoder(w).Encode(p)
}
```

Четыре зоны ответственности в одной функции. При росте кода это превращается в кашу.

Решение — разделить на три слоя:

| Слой | Ответственность | Пример |
|------|----------------|--------|
| **Handler** | HTTP: парсинг запроса, статус-коды, заголовки | `createPost(w, r)` |
| **Service** | Бизнес-логика: валидация, правила, оркестрация | `CreatePost(title, body) (Post, error)` |
| **Repository** | Хранение: CRUD, запросы, транзакции | `Save(post)`, `FindByID(id)` |

## Структура директорий

```
cmd/server/
├── main.go            # Точка входа: роутер, middleware, запуск
internal/
├── handler/
│   └── posts.go       # HTTP-обработчики
├── service/
│   └── posts.go       # Бизнес-логика
├── repository/
│   └── memory.go      # In-memory хранилище
└── model/
    └── post.go        # Модели (Post, Status, ErrorResponse)
```

:::info Почему internal, а не pkg
`internal` — специальная директория Go. Компилятор запрещает импортировать пакеты из `internal/` другим модулям. Это защищает от случайного использования внутренних типов снаружи.
:::

## Слой Model

Вынесем все типы в отдельный пакет:

```go
// internal/model/post.go
package model

import (
    "encoding/json"
    "fmt"
    "strings"
    "time"
)

type Status int

const (
    StatusDraft     Status = 0
    StatusPublished Status = 1
)

func (s Status) MarshalJSON() ([]byte, error) { /* ... */ }
func (s *Status) UnmarshalJSON(data []byte) error { /* ... */ }

type Post struct {
    ID        int       `json:"id"`
    Title     string    `json:"title"`
    Body      string    `json:"body"`
    Status    Status    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
}

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

type ErrorResponse struct {
    Error string `json:"error"`
}
```

## Слой Repository

Репозиторий — единственное место, которое знает, **как** хранятся данные:

```go
// internal/repository/memory.go
package repository

import "github.com/go-course/project-1-blog/internal/model"

type MemoryRepo struct {
    posts  []model.Post
    nextID int
}

func NewMemory() *MemoryRepo {
    return &MemoryRepo{
        posts: []model.Post{
            {ID: 1, Title: "Первый пост", Body: "Привет!", Status: model.StatusPublished},
        },
        nextID: 2,
    }
}

func (r *MemoryRepo) FindAll() []model.Post {
    return r.posts
}

func (r *MemoryRepo) FindByID(id int) (model.Post, bool) {
    for _, p := range r.posts {
        if p.ID == id {
            return p, true
        }
    }
    return model.Post{}, false
}

func (r *MemoryRepo) Create(p model.Post) model.Post {
    p.ID = r.nextID
    r.nextID++
    r.posts = append(r.posts, p)
    return p
}

func (r *MemoryRepo) Update(id int, p model.Post) (model.Post, bool) {
    for i, existing := range r.posts {
        if existing.ID == id {
            p.ID = id
            r.posts[i] = p
            return p, true
        }
    }
    return model.Post{}, false
}

func (r *MemoryRepo) Delete(id int) bool {
    for i, p := range r.posts {
        if p.ID == id {
            r.posts = append(r.posts[:i], r.posts[i+1:]...)
            return true
        }
    }
    return false
}
```

:::tip Интерфейс перед реализацией
Позже мы заменим `MemoryRepo` на `PostgresRepo`. Чтобы не переписывать service, определи интерфейс:

```go
type PostRepository interface {
    FindAll() []model.Post
    FindByID(id int) (model.Post, bool)
    Create(p model.Post) model.Post
    Update(id int, p model.Post) (model.Post, bool)
    Delete(id int) bool
}
```

Service зависит от интерфейса, а не от конкретной реализации — это ключ к тестируемости.
:::

## Слой Service

Сервис содержит бизнес-логику и **не знает про HTTP**:

```go
// internal/service/posts.go
package service

import (
    "fmt"
    "time"

    "github.com/go-course/project-1-blog/internal/model"
)

type PostRepository interface {
    FindAll() []model.Post
    FindByID(id int) (model.Post, bool)
    Create(p model.Post) model.Post
    Update(id int, p model.Post) (model.Post, bool)
    Delete(id int) bool
}

type PostService struct {
    repo PostRepository
}

func NewPost(repo PostRepository) *PostService {
    return &PostService{repo: repo}
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
        return model.Post{}, err
    }
    return s.repo.Create(p), nil
}

func (s *PostService) Update(id int, title, body string) (model.Post, error) {
    p, ok := s.repo.FindByID(id)
    if !ok {
        return model.Post{}, fmt.Errorf("post not found")
    }
    p.Title = title
    p.Body = body
    if err := p.Validate(); err != nil {
        return model.Post{}, err
    }
    return s.repo.Update(id, p)
}

func (s *PostService) Delete(id int) error {
    if ok := s.repo.Delete(id); !ok {
        return fmt.Errorf("post not found")
    }
    return nil
}
```

Сервис принимает простые типы (`string`, `int`) и возвращает `(model.Post, error)`. Никаких `http.ResponseWriter`.

## Слой Handler

Самый тонкий слой — только HTTP-конвертация:

```go
// internal/handler/posts.go
package handler

import (
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/go-course/project-1-blog/internal/model"
    "github.com/go-course/project-1-blog/internal/service"
)

type PostHandler struct {
    svc *service.PostService
}

func NewPost(svc *service.PostService) *PostHandler {
    return &PostHandler{svc: svc}
}

func (h *PostHandler) List(w http.ResponseWriter, r *http.Request) {
    posts := h.svc.List()
    writeJSON(w, http.StatusOK, posts)
}

func (h *PostHandler) Get(w http.ResponseWriter, r *http.Request) {
    // строку в int — на совести handler'а
    id, err := strconv.Atoi(chi.URLParam(r, "id"))
    if err != nil {
        writeError(w, http.StatusBadRequest, "invalid id")
        return
    }
    p, err := h.svc.Get(id)
    if err != nil {
        writeError(w, http.StatusNotFound, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, p)
}
```

:::info Handler — переводчик
Handler переводит HTTP-термины (статус-коды, заголовки, URL-параметры) в вызовы сервиса и обратно. Никакой бизнес-логики.
:::

## main.go — собираем всё вместе

```go
func main() {
    // Инициализация слоёв
    repo := repository.NewMemory()
    svc := service.NewPost(repo)
    h := handler.NewPost(svc)

    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

    r.Get("/posts", h.List)
    r.Get("/posts/{id}", h.Get)
    // ...

    server.ListenAndServe()
}
```

Зависимости передаются явно (через конструктор) — никаких глобальных переменных.

## Почему это лучше

1. **Тестируемость** — сервис можно тестировать с мок-репозиторием, без HTTP-сервера
2. **Замена хранилища** — меняем `MemoryRepo` на `PostgresRepo`, handler и service не трогаем
3. **Переиспользование** — бизнес-логику можно дёргать из CLI, gRPC, очереди — не только из HTTP
4. **Читаемость** — каждый файл отвечает за что-то одно

## Ключевые выводы

1. Слои: **handler** (HTTP) → **service** (логика) → **repository** (хранение)
2. Зависимости через интерфейсы — не через конкретные типы
3. Service не знает про HTTP, handler не знает про хранение
4. `internal/` защищает от внешних импортов
5. Явная передача зависимостей через конструкторы

В следующей главе замокаем репозиторий и напишем юнит-тесты для сервиса.

---

## Проверь себя

<Quiz
  quizId="05-project-structure"
  questions={[
    {
      id: "q1",
      question: "Какую проблему решает слоёная архитектура handler → service → repository?",
      options: [
        "Ускоряет компиляцию Go-кода",
        "Разделяет ответственность: HTTP, бизнес-логика и хранение не смешаны",
        "Позволяет использовать несколько HTTP-роутеров одновременно",
        "Это требование стандартной библиотеки Go"
      ],
      correctIndex: 1,
      explanation: "Без слоёв HTTP-парсинг, бизнес-правила и запросы к БД смешиваются в одной функции. Слои разделяют зоны ответственности и делают код тестируемым."
    },
    {
      id: "q2",
      question: "Почему service должен зависеть от интерфейса, а не от конкретного репозитория?",
      options: [
        "Интерфейсы работают быстрее",
        "Чтобы можно было заменить реализацию (Memory → Postgres) без изменения service",
        "Go требует интерфейсы для всех зависимостей",
        "Только ради тестов, в продакшене это неважно"
      ],
      correctIndex: 1,
      explanation: "Интерфейс — контракт. Service знает «что» делает репозиторий, но не «как». Это позволяет подменить MemoryRepo на PostgresRepo или на мок для тестов."
    },
    {
      id: "q3",
      question: "Зачем нужна директория internal/?",
      options: [
        "Это просто соглашение, никакой разницы с pkg/",
        "Go запрещает импорт internal-пакетов из других модулей — защита от внешних зависимостей",
        "internal/ быстрее компилируется",
        "Там хранятся конфиденциальные данные"
      ],
      correctIndex: 1,
      explanation: "Компилятор Go запрещает импорт пакетов из internal/ другим модулям. Это гарантирует, что внутренние типы не «утекут» наружу и не станут частью публичного API."
    }
  ]}
/>
