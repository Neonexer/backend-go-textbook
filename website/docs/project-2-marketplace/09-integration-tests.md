---
title: "Интеграционные тесты"
sidebar_position: 9
---

import Quiz from '@site/src/components/Quiz';

# Интеграционные тесты

Юнит-тесты с моками быстрые, но не проверяют реальную БД. Интеграционные тесты поднимают настоящий PostgreSQL и проверяют запросы. Используем `testcontainers-go` — Docker-контейнеры в тестах.

## testcontainers-go

```bash
go get github.com/testcontainers/testcontainers-go
go get github.com/testcontainers/testcontainers-go/modules/postgres
```

## Поднимаем PostgreSQL в тесте

```go
func TestProductRepo_Integration(t *testing.T) {
    ctx := context.Background()

    // Запускаем контейнер с PostgreSQL
    pgContainer, err := postgres.Run(ctx,
        "postgres:16-alpine",
        postgres.WithDatabase("marketplace_test"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections"),
        ),
    )
    require.NoError(t, err)
    defer pgContainer.Terminate(ctx)

    // Строка подключения
    connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
    require.NoError(t, err)

    pool, err := pgxpool.New(ctx, connStr)
    require.NoError(t, err)
    defer pool.Close()

    // Применяем миграции
    runMigrations(connStr)

    // Тестируем репозиторий
    repo := NewProduct(pool)
    p, err := repo.Create(ctx, model.Product{Title: "Test", Price: 1000})
    require.NoError(t, err)
    assert.Equal(t, "Test", p.Title)
    assert.NotZero(t, p.ID)

    // Находим созданный товар
    found, err := repo.FindByID(ctx, p.ID)
    require.NoError(t, err)
    assert.Equal(t, "Test", found.Title)
}
```

## Что тестировать интеграционно

- **Репозиторий** — SQL-запросы на реальной БД
- **Транзакции** — конкурентный доступ
- **Миграции** — up и down

## Структура интеграционных тестов

```go
// product_pg_test.go
//go:build integration
// +build integration

package repository_test
```

Build tag `integration` позволяет запускать интеграционные тесты отдельно:

```bash
go test ./... -tags=integration    # с интеграционными
go test ./...                       # только юнит-тесты (быстрые)
```

## Ключевые выводы

- testcontainers дают реальный PostgreSQL в Docker
- Интеграционные тесты проверяют SQL и миграции
- Build tags разделяют быстрые и медленные тесты
- Всегда `defer pgContainer.Terminate(ctx)` — убираем за собой

<Quiz quizId="p2-09-integration" questions={[
  {id:"q1",question:"Зачем разделять юнит и интеграционные тесты через build tags?",options:["Чтобы не запускать Docker-зависимые тесты при каждом go test — они медленные и требуют Docker","Это требование Go","Интеграционные тесты не работают без build tags","Для совместимости с CI"],correctIndex:0,explanation:"Интеграционные тесты требуют Docker и работают секунды/десятки секунд. Юнит-тесты с моками — миллисекунды. Build tags позволяют запускать их раздельно."}
]} />
