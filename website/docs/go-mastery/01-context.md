---
title: "Контексты глубоко"
sidebar_position: 1
---

import Quiz from '@site/src/components/Quiz';

# context.Context — глубокое погружение

Контекст — краеугольный камень Go-приложений. Он несёт дедлайны, cancellation сигналы и values. Но его неправильное использование — источник багов. Разберёмся глубоко.

## Что на самом деле делает контекст

```go
type Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key any) any
}
```

Четыре метода. Никакой магии.

## Дерево контекстов

Контексты — дерево. Отмена родителя отменяет всех потомков:

```go
root := context.Background()
ctx1, cancel1 := context.WithCancel(root)
ctx2, cancel2 := context.WithTimeout(ctx1, 5*time.Second)
ctx3 := context.WithValue(ctx2, "user_id", 42)

cancel1() // ctx1, ctx2, ctx3 — все отменены
```

## Правила

**1. Контекст — первый параметр**

```go
func GetPost(ctx context.Context, id int) (Post, error) // ✅
func GetPost(id int, ctx context.Context) (Post, error) // ❌
```

**2. Не храни контекст в структуре**

```go
type Service struct {
    ctx context.Context // ❌ антипаттерн
}

// ✅ Контекст передаётся в метод
func (s *Service) GetPost(ctx context.Context, id int) (Post, error)
```

**3. ctx.Value — только для request-scoped данных**

```go
// ✅ OK: request_id, user_id, tracer
ctx = context.WithValue(ctx, "request_id", reqID)

// ❌ НЕ OK: параметры конфигурации, зависимости
ctx = context.WithValue(ctx, "db_url", dbURL) // так не делай
```

**4. Не игнорируй cancel**

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel() // всегда! даже если таймаут ещё не истёк — освобождает ресурсы
```

## HTTP-сервер и контекст

`r.Context()` — контекст запроса. Отменяется когда клиент разрывает соединение:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    // Если клиент ушёл — ctx.Done() закроется
    // Долгие операции должны проверять ctx
    select {
    case <-ctx.Done():
        return
    case result := <-longOperation():
        json.NewEncoder(w).Encode(result)
    }
}
```

<Quiz quizId="gm-01-context" questions={[
  {id:"q1",question:"Почему нельзя хранить context.Context в структуре?",options:["Это медленно","Контекст привязан к запросу/операции а не к сервису. Структура живёт дольше запроса — контекст устареет","Компилятор запрещает","Это просто стиль"],correctIndex:1,explanation:"Сервис создаётся один раз при старте. Контекст запроса живёт только пока длится запрос. Хранить ctx в сервисе = использовать мёртвый контекст."}
]} />
