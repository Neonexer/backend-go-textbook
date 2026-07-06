---
title: "Обработка ошибок"
sidebar_position: 10
---

import Quiz from '@site/src/components/Quiz';

# Обработка ошибок в Go

Ошибки в Go — это значения. Не исключения, не try/catch. В этой главе разберём идиоматичные паттерны: sentinel errors, кастомные типы, wrapping и multi-error.

## Ошибки как значения

```go
func FindByID(id int) (Post, error) {
    if id <= 0 {
        return Post{}, errors.New("id must be positive")
    }
    // ...
}
```

Вызывающий проверяет:

```go
post, err := repo.FindByID(42)
if err != nil {
    // обработка
}
```

Никакой магии. `error` — это интерфейс с одним методом `Error() string`.

## Sentinel errors

Для сравнения ошибок — предопределённые значения:

```go
var ErrNotFound = errors.New("not found")
var ErrPermissionDenied = errors.New("permission denied")

func (r *ProductRepo) FindByID(id int) (Product, error) {
    // ...
    if err == pgx.ErrNoRows {
        return Product{}, ErrNotFound
    }
}

// На вызывающей стороне
if errors.Is(err, ErrNotFound) {
    // 404
}
```

`errors.Is` сравнивает по цепочке, даже если ошибка обёрнута.

## Кастомный тип ошибки

Когда нужно передать контекст:

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation: %s %s", e.Field, e.Message)
}

// Использование
func (p *Post) Validate() error {
    if p.Title == "" {
        return &ValidationError{Field: "title", Message: "required"}
    }
    return nil
}
```

## Wrapping: добавляем контекст

`fmt.Errorf` с `%w` оборачивает ошибку:

```go
func (s *PostService) Get(id int) (Post, error) {
    p, err := s.repo.FindByID(id)
    if err != nil {
        return Post{}, fmt.Errorf("get post %d: %w", id, err)
    }
    return p, nil
}
```

```go
err := svc.Get(42)
// err.Error() = "get post 42: not found"

errors.Is(err, ErrNotFound)     // true — распаковывает цепочку
var valErr *ValidationError
errors.As(err, &valErr)          // true — извлекает тип
```

:::tip Всегда используй %w, не %v
`%w` оборачивает ошибку с сохранением цепочки. `%v` превращает в строку — `errors.Is` перестаёт работать.
:::

## Multi-error

Несколько ошибок — не останавливаемся на первой:

```go
import "errors"

func (p *Post) Validate() error {
    var errs []error
    if p.Title == "" {
        errs = append(errs, fmt.Errorf("title required"))
    }
    if len(p.Title) > 200 {
        errs = append(errs, fmt.Errorf("title too long"))
    }
    if p.Body == "" {
        errs = append(errs, fmt.Errorf("body required"))
    }
    return errors.Join(errs...) // Go 1.20+
}
```

## Логирование ошибок

```go
if err != nil {
    slog.Error("ошибка создания поста", "err", err, "user_id", userID)
    // Не логируй пароли, токены и персональные данные
}
```

<Quiz quizId="p1-10-errors" questions={[
  {id:"q1",question:"В чём разница между %w и %v в fmt.Errorf?",options:["Никакой","%w оборачивает ошибку сохраняя цепочку для errors.Is/As, %v превращает в строку и разрывает цепочку","%w быстрее","%v не работает с ошибками"],correctIndex:1,explanation:"Только %w сохраняет связь с исходной ошибкой. errors.Is(err, ErrNotFound) работает через %w, но не через %v."},
  {id:"q2",question:"Когда использовать sentinel errors vs кастомный тип?",options:["Sentinel — для простых проверок (is it 'not found'?), кастомный тип — когда нужен контекст (какое поле не прошло валидацию?)","Всегда кастомные типы","Всегда sentinel errors","В Go только один тип ошибок"],correctIndex:0,explanation:"Sentinel (ErrNotFound) — простое условие. Кастомный тип (ValidationError с полем Field) — когда нужна дополнительная информация для обработки."}
]} />
