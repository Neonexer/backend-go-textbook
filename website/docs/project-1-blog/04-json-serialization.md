---
title: "JSON-сериализация"
sidebar_position: 4
---

import Quiz from '@site/src/components/Quiz';

# JSON-сериализация

JSON — язык общения REST API. Go даёт `encoding/json` из стандартной библиотеки. Без рефлексии, без кодогенерации, без аннотаций. Только структурные теги и явное преобразование. В этой главе разберём Marshal, Unmarshal, теги и кастомную сериализацию.

## Marshal: Go → JSON

`json.Marshal` превращает Go-структуру в JSON-байты:

```go
type Post struct {
    ID    int    `json:"id"`
    Title string `json:"title"`
    Body  string `json:"body"`
}

p := Post{ID: 1, Title: "Привет", Body: "Текст поста"}
data, err := json.Marshal(p)
// data = {"id":1,"title":"Привет","body":"Текст поста"}
```

В HTTP-обработчике это выглядит так:

```go
func listPosts(w http.ResponseWriter, r *http.Request) {
    posts := []Post{{ID: 1, Title: "Первый"}}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(posts)
}
```

`json.NewEncoder(w).Encode(posts)` эффективнее, чем `json.Marshal` + `w.Write` — энкодер пишет прямо в поток, не создавая промежуточный слайс байт.

## Unmarshal: JSON → Go

`json.Unmarshal` — обратное преобразование:

```go
var p Post
err := json.Unmarshal(data, &p)
```

В обработчике — через `json.NewDecoder`:

```go
func createPost(w http.ResponseWriter, r *http.Request) {
    var p Post
    if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
        http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
        return
    }
    // ... работаем с p
}
```

:::warning Декодировать только один раз
`r.Body` — это поток. `json.NewDecoder(r.Body).Decode(&p)` читает тело до конца. Повторный вызов Decode на том же Body вернёт `io.EOF`.
:::

## Структурные теги

Теги управляют тем, как поля отображаются в JSON:

```go
type Post struct {
    ID        int       `json:"id"`
    Title     string    `json:"title"`
    Body      string    `json:"body"`
    CreatedAt time.Time `json:"created_at"`
    AuthorID  int       `json:"author_id,omitempty"` // не выводить, если 0
    internal  string    `json:"-"`                   // игнорировать всегда
}
```

| Тег | Значение |
|-----|---------|
| `json:"fieldname"` | Имя поля в JSON |
| `json:"fieldname,omitempty"` | Не выводить, если значение zero-value |
| `json:"-"` | Всегда игнорировать (даже если публичное) |
| `json:"fieldname,string"` | Сериализовать как строку |

### omitempty и zero-value

```go
type Post struct {
    Title string `json:"title,omitempty"`
    Views int    `json:"views,omitempty"`
}

p := Post{Title: "", Views: 0}
data, _ := json.Marshal(p)
// {}
```

Для `string` zero-value — `""`, для `int` — `0`, для `bool` — `false`, для указателей и слайсов — `nil`.

:::tip omitempty и time.Time
`time.Time` — это структура, её zero-value не `nil`. Поэтому `omitempty` не скроет `"0001-01-01T00:00:00Z"`. Используй `*time.Time` (указатель) — тогда `nil` скроется через `omitempty`.
:::

## Кастомная сериализация

Если нужно изменить формат поля — реализуй интерфейсы `json.Marshaler` / `json.Unmarshaler`:

```go
type Status int

const (
    StatusDraft     Status = 0
    StatusPublished Status = 1
)

func (s Status) MarshalJSON() ([]byte, error) {
    switch s {
    case StatusDraft:
        return json.Marshal("draft")
    case StatusPublished:
        return json.Marshal("published")
    default:
        return json.Marshal("unknown")
    }
}

func (s *Status) UnmarshalJSON(data []byte) error {
    var str string
    if err := json.Unmarshal(data, &str); err != nil {
        return err
    }
    switch str {
    case "draft":
        *s = StatusDraft
    case "published":
        *s = StatusPublished
    default:
        *s = 0
    }
    return nil
}
```

Теперь `Status` сериализуется как `"draft"`/`"published"` вместо `0`/`1`.

## DisallowUnknownFields — защита от опечаток

По умолчанию `json.Unmarshal` игнорирует неизвестные поля. Если клиент пришлёт `{"body": "...", "titel": "опечатка"}`, поле `titel` будет молча проигнорировано. Это плохо.

```go
decoder := json.NewDecoder(r.Body)
decoder.DisallowUnknownFields()

var p Post
if err := decoder.Decode(&p); err != nil {
    // вернёт ошибку при неизвестных полях
    http.Error(w, `{"error":"unknown field"}`, http.StatusBadRequest)
    return
}
```

Всегда включай `DisallowUnknownFields()` в продакшене.

## Валидация после Unmarshal

`json.Unmarshal` проверяет только типы, но не бизнес-правила. Валидацию делаем сами:

```go
func (p *Post) Validate() error {
    if p.Title == "" {
        return fmt.Errorf("title is required")
    }
    if len(p.Title) > 200 {
        return fmt.Errorf("title too long: %d chars", len(p.Title))
    }
    if len(p.Body) == 0 {
        return fmt.Errorf("body is required")
    }
    return nil
}

func createPost(w http.ResponseWriter, r *http.Request) {
    var p Post
    if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
        http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
        return
    }
    if err := p.Validate(); err != nil {
        http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
        return
    }
    // ...
}
```

:::tip Когда стоит подключать библиотеку валидации
Для простых API ручная валидация достаточна. Если полей >10 и правил много — смотри `go-playground/validator` (структурные теги вроде `validate:"required,min=1,max=200"`). Мы будем использовать её в Проекте 2, когда появятся сложные структуры.
:::

## Паттерн: строгий ответ

Возвращать сырые сообщения об ошибках неудобно. Сделаем структуру ответа:

```go
type ErrorResponse struct {
    Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
    writeJSON(w, status, ErrorResponse{Error: msg})
}
```

Использование:

```go
writeJSON(w, http.StatusCreated, p)
writeError(w, http.StatusNotFound, "post not found")
```

## Ключевые выводы

1. `json.NewEncoder(w).Encode(v)` в обработчиках — эффективнее, чем Marshal + Write
2. `DisallowUnknownFields()` всегда включай в продакшене
3. Валидация — отдельный шаг после Unmarshal, не возлагай всё на JSON-теги
4. `omitempty` не работает для `time.Time` — используй `*time.Time`
5. Унифицируй ответы через хелперы `writeJSON` / `writeError`

В следующей главе наведём порядок в структуре проекта: handler → service → repository.

---

## Проверь себя

<Quiz
  quizId="04-json-serialization"
  questions={[
    {
      id: "q1",
      question: "Почему json.NewEncoder(w).Encode(v) лучше, чем json.Marshal + w.Write?",
      options: [
        "Marshal работает только с маленькими структурами",
        "Encoder пишет прямо в поток, не создавая промежуточный слайс байт в памяти",
        "Encoder автоматически сжимает JSON",
        "Никакой разницы, это вопрос стиля"
      ],
      correctIndex: 1,
      explanation: "Encoder пишет прямо в io.Writer. Marshal создаёт []byte в памяти, который потом копируется в ResponseWriter. Для больших ответов разница существенна."
    },
    {
      id: "q2",
      question: "Что делает DisallowUnknownFields()?",
      options: [
        "Запрещает любые поля, кроме указанных в структуре",
        "Возвращает ошибку при неизвестных полях в JSON вместо их игнорирования",
        "Удаляет неизвестные поля из JSON перед Unmarshal",
        "Требует, чтобы все поля структуры были заполнены"
      ],
      correctIndex: 1,
      explanation: "По умолчанию json.Unmarshal молча игнорирует поля, которых нет в структуре. DisallowUnknownFields() заставляет его возвращать ошибку — защита от опечаток."
    },
    {
      id: "q3",
      question: "Почему omitempty не работает для time.Time?",
      options: [
        "Это баг в Go, его ещё не исправили",
        "time.Time — структура, её zero-value не nil, поэтому omitempty не срабатывает",
        "Для time.Time нужно использовать отдельный тег timeomitempty",
        "omitempty не поддерживается для не-строковых полей"
      ],
      correctIndex: 1,
      explanation: "time.Time — это struct{...}, а не указатель. Его zero-value — конкретная дата 0001-01-01, что не равно nil. Решение: использовать *time.Time."
    }
  ]}
/>
