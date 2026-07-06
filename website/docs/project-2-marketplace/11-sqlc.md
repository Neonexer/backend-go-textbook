---
title: "sqlc — кодогенерация SQL"
sidebar_position: 11
---

import Quiz from '@site/src/components/Quiz';

# sqlc — кодогенерация SQL

pgx — ручной SQL. GORM — ORM. **sqlc** — третий путь: ты пишешь обычный SQL, sqlc генерирует типобезопасный Go-код на этапе компиляции. Никакой рефлексии в рантайме.

## Как это работает

1. Пишешь SQL-запросы в `.sql` файлах с аннотациями
2. Запускаешь `sqlc generate`
3. Получаешь Go-код с готовыми функциями

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

## Конфигурация

```yaml
# sqlc.yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "queries/"
    schema: "migrations/"
    gen:
      go:
        package: "db"
        out: "internal/db"
        sql_package: "pgx/v5"
```

## SQL-запросы

```sql
-- queries/products.sql

-- name: GetProduct :one
SELECT * FROM products WHERE id = $1;

-- name: ListProducts :many
SELECT * FROM products ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: CreateProduct :one
INSERT INTO products (title, description, price, seller_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateProduct :execrows
UPDATE products SET title=$2, price=$3 WHERE id=$1 AND seller_id=$4;

-- name: DeleteProduct :execrows
DELETE FROM products WHERE id=$1 AND seller_id=$2;
```

Аннотации:
- `:one` — одна запись
- `:many` — слайс
- `:exec` — без возврата
- `:execrows` — возвращает количество затронутых строк

## Сгенерированный код

После `sqlc generate`:

```go
// internal/db/products.sql.go
func (q *Queries) GetProduct(ctx context.Context, id int32) (Product, error) {
    // sqlc генерирует pgx-вызов с правильными типами
}

func (q *Queries) ListProducts(ctx context.Context, arg ListProductsParams) ([]Product, error) {
    // ...
}
```

Использование:

```go
pool, _ := pgxpool.New(ctx, dsn)
queries := db.New(pool)

product, err := queries.GetProduct(ctx, 42)
products, err := queries.ListProducts(ctx, db.ListProductsParams{Limit: 20, Offset: 0})
```

## sqlc vs pgx vs GORM

| | pgx (ручной) | GORM (ORM) | sqlc (кодоген) |
|---|---|---|---|
| Где SQL | В Go-коде строками | Генерируется ORM | В `.sql` файлах |
| Типобезопасность | Ручная (Scan) | Автоматически | Автоматически |
| Сложные JOIN'ы | Легко | Трудно | Легко (чистый SQL) |
| Скорость | Максимальная | Оверхед ORM | Как pgx (нет оверхеда) |

<Quiz quizId="p2-11-sqlc" questions={[
  {id:"q1",question:"В чём отличие sqlc от ORM вроде GORM?",options:["Ни в чём","sqlc работает на этапе компиляции — генерирует код из SQL без рефлексии. GORM работает в рантайме через рефлексию.","sqlc быстрее GORM в 100 раз","sqlc не поддерживает PostgreSQL"],correctIndex:1,explanation:"sqlc компилирует SQL в Go-код. В рантайме — обычные вызовы pgx, никакой рефлексии. GORM в рантайме анализирует структуры и строит SQL."}
]} />
