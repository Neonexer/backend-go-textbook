---
title: "Тестирование"
sidebar_position: 6
---

import Quiz from '@site/src/components/Quiz';

# Тестирование

Тестирование в Go — first-class citizen. Никаких плагинов, никаких аннотаций. Просто функция с именем `Test*` в файле `*_test.go`. В этой главе разберём тестирование HTTP-обработчиков, табличные тесты и моки через интерфейсы.

## Как Go запускает тесты

```bash
go test ./...        # все тесты в пакете и подпакетах
go test -v           # подробный вывод
go test -run TestGet # запустить конкретный тест
go test -cover       # с замером покрытия
```

Тесты живут в том же пакете, что и код, но в файлах с суффиксом `_test.go`. Go компилирует их отдельно — в продакшен-билд они не попадают.

## httptest — тестирование HTTP без сервера

Пакет `net/http/httptest` позволяет тестировать обработчики без поднятия реального сервера:

```go
func TestGetPost(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/posts/1", nil)
    rec := httptest.NewRecorder()

    handler := setupTestRouter()
    handler.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Errorf("ожидался 200, получен %d", rec.Code)
    }
}
```

Никаких портов, никаких `time.Sleep`. `httptest.NewRecorder` реализует `http.ResponseWriter` и записывает всё в память.

## Табличные тесты

Вместо копипасты — одна таблица с вариантами:

```go
func TestCreatePost_Validation(t *testing.T) {
    tests := []struct {
        name   string
        body   string
        status int
    }{
        {"empty title", `{"title":"","body":"text"}`, http.StatusBadRequest},
        {"empty body", `{"title":"ok","body":""}`, http.StatusBadRequest},
        {"too long", `{"title":"` + strings.Repeat("x", 201) + `","body":"text"}`, http.StatusBadRequest},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(tt.body))
            req.Header.Set("Authorization", "Bearer test")
            rec := httptest.NewRecorder()

            setupTestRouter().ServeHTTP(rec, req)

            if rec.Code != tt.status {
                t.Errorf("ожидался %d, получен %d", tt.status, rec.Code)
            }
        })
    }
}
```

:::tip Почему табличные тесты
- Легко добавить новый случай — ещё одна строка
- `t.Run` даёт имя каждому случаю в выводе
- Можно запустить конкретный вариант: `go test -run TestCreatePost_Validation/empty_title`
:::

## Моки через интерфейсы

Благодаря слоям, сервис зависит от интерфейса `PostRepository`, а не от конкретного `MemoryRepo`:

```go
// mockRepo — тестовая реализация PostRepository
type mockRepo struct {
    posts []model.Post
}

func (m *mockRepo) FindAll() []model.Post { return m.posts }
func (m *mockRepo) FindByID(id int) (model.Post, bool) {
    for _, p := range m.posts {
        if p.ID == id { return p, true }
    }
    return model.Post{}, false
}
// ... остальные методы интерфейса

func TestList(t *testing.T) {
    repo := &mockRepo{
        posts: []model.Post{{ID: 1, Title: "Test"}},
    }
    svc := service.NewPost(repo)
    posts := svc.List()
    if len(posts) != 1 {
        t.Errorf("expected 1, got %d", len(posts))
    }
}
```

Интерфейс + ручная реализация = мок. Никаких фреймворков. Для сложных случаев есть `testify/mock`, но для большинства задач хватает ручных моков.

## Тестирование с авторизацией

Просто добавь заголовок в запрос:

```go
// Без токена — 401
func TestCreatePost_NoAuth(t *testing.T) {
    req := httptest.NewRequest(http.MethodPost, "/posts", body)
    rec := httptest.NewRecorder()
    setupTestRouter().ServeHTTP(rec, req)
    if rec.Code != http.StatusUnauthorized {
        t.Errorf("ожидался 401, получен %d", rec.Code)
    }
}

// С токеном — 201
func TestCreatePost_WithAuth(t *testing.T) {
    req := httptest.NewRequest(http.MethodPost, "/posts", body)
    req.Header.Set("Authorization", "Bearer test-token")
    rec := httptest.NewRecorder()
    setupTestRouter().ServeHTTP(rec, req)
    if rec.Code != http.StatusCreated {
        t.Errorf("ожидался 201, получен %d", rec.Code)
    }
}
```

## Что тестировать в первую очередь

| Приоритет | Слой | Почему |
|-----------|------|--------|
| 🔴 Высокий | Service | Бизнес-логика, правила, валидация. Ошибка = баг. |
| 🟡 Средний | Handler | Сериализация, статус-коды, парсинг URL. |
| 🟢 Низкий | Repository (in-memory) | Простые CRUD, покрываются косвенно. |

## Покрытие

```bash
go test ./... -cover
```

Не гонись за 100% везде. 100% сервиса — отлично. 100% main.go — бессмысленно (там только сборка).

## Ключевые выводы

1. `httptest` тестирует HTTP без сети — быстро и надёжно
2. Табличные тесты — идиоматичный Go-way для многих случаев
3. Интерфейсы = встроенный механизм моков без фреймворков
4. Service-тесты дают максимальную отдачу
5. `t.Run` для подтестов, `-run` для фильтрации

В следующей главе подключим структурное логирование с `log/slog`.

---

## Проверь себя

<Quiz
  quizId="06-testing"
  questions={[
    {
      id: "q1",
      question: "Зачем httptest.NewRecorder вместо поднятия реального сервера?",
      options: [
        "Тесты быстрее и не требуют свободного порта",
        "Реальный сервер нельзя тестировать в Go",
        "NewRecorder автоматически генерирует тестовые данные",
        "Это требование go vet"
      ],
      correctIndex: 0,
      explanation: "httptest.NewRecorder записывает ответ в память без сетевого стека. Тесты запускаются мгновенно и не конкурируют за порты."
    },
    {
      id: "q2",
      question: "Почему интерфейсы в Go — это встроенный механизм моков?",
      options: [
        "Компилятор автоматически генерирует моки",
        "Интерфейс — контракт. Подменяем реализацию на тестовую, код не меняется",
        "Интерфейсы не используются для тестирования",
        "Только из-за особенностей компилятора"
      ],
      correctIndex: 1,
      explanation: "Service зависит от PostRepository (интерфейс), а не от MemoryRepo (тип). В тестах передаём mockRepo — service не замечает разницы."
    },
    {
      id: "q3",
      question: "Какой слой даёт наибольшую отдачу от тестирования?",
      options: [
        "main.go — точка входа",
        "Handler — HTTP-слой",
        "Service — бизнес-логика и правила",
        "Repository — хранение данных"
      ],
      correctIndex: 2,
      explanation: "Ошибка в бизнес-логике = баг. Сервисный слой содержит правила, валидацию и оркестрацию. Тесты сервиса дают максимальную уверенность."
    }
  ]}
/>
