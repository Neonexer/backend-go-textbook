---
title: "Clean / Hexagonal Architecture"
sidebar_position: 1
---

import Quiz from '@site/src/components/Quiz';

# Clean Architecture

Три слоя (handler → service → repository) — упрощённая версия Clean Architecture. Полная версия: домен в центре, инфраструктура снаружи, зависимости только внутрь.

## Принцип инверсии зависимостей

```
    ┌──────────────────────┐
    │       Domain         │  ← Правила, сущности (нет импортов извне)
    │  (Post, PostRepo)    │
    └──────────┬───────────┘
               │  implements
    ┌──────────▼───────────┐
    │    Use Cases         │  ← Бизнес-логика (service)
    │  (PostService)       │
    └──────────┬───────────┘
               │  implements
    ┌──────────▼───────────┐
    │   Adapters           │  ← HTTP, gRPC, БД, Kafka
    │ (handler, pgx repo)  │
    └──────────────────────┘
```

## Домен не зависит от БД

В классической трёхслойке сервис знает про репозиторий. В Clean Architecture — наоборот:

```go
// domain/post.go — НЕ импортирует ни handler, ни pgx
package domain

type Post struct { /* ... */ }

type PostRepository interface {
    FindByID(ctx context.Context, id int) (Post, error)
    Save(ctx context.Context, p Post) error
}
```

```go
// usecase/create_post.go
package usecase

type CreatePostUseCase struct {
    repo domain.PostRepository // зависит от интерфейса
}

func (uc *CreatePostUseCase) Execute(ctx context.Context, input CreatePostInput) (Post, error) {
    // бизнес-логика
}
```

```go
// adapter/postgres_repo.go — реализует интерфейс домена
package adapter

type PostgresPostRepo struct { pool *pgxpool.Pool }
func (r *PostgresPostRepo) FindByID(ctx context.Context, id int) (domain.Post, error) {
    // SQL
}
```

## Почему это важно

- Домен можно тестировать без БД и HTTP
- Меняем PostgreSQL на MongoDB — домен не трогаем
- Меняем HTTP на gRPC — use cases не трогаем

<Quiz quizId="arch-01-clean" questions={[
  {id:"q1",question:"Главное правило Clean Architecture?",options:["Все слои зависят друг от друга","Зависимости направлены внутрь: внешние слои знают о внутренних, но не наоборот. Домен не импортирует БД/HTTP.","Всегда использовать интерфейсы","Максимум 3 слоя"],correctIndex:1,explanation:"Стрелка зависимости всегда к центру. Домен не знает о Postgres. Репозиторий реализует доменный интерфейс а не наоборот."}
]} />
