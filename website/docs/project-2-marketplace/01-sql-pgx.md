---
title: "SQL и pgx"
sidebar_position: 1
---

import Quiz from '@site/src/components/Quiz';

# SQL и pgx

В Проекте 1 мы хранили посты в памяти. Для реального маркетплейса нужна база данных. В этой главе подключим PostgreSQL через `pgx` — быстрый, нативный драйвер с поддержкой всех фич Postgres.

## Почему PostgreSQL, а не MySQL

Для бэкенд-приложений Postgres объективно лучший выбор по умолчанию:

- **Расширяемость**: JSONB, полнотекстовый поиск, PostGIS — из коробки
- **Конкурентность**: MVCC без блокировок чтения
- **Стандарты**: ближе всех к стандартному SQL
- **Экосистема Go**: `pgx` — самый быстрый и полный драйвер

## Почему pgx, а не database/sql

`database/sql` — общий интерфейс, работает с любой БД. `pgx` — драйвер специально под Postgres:

| | database/sql + pq | pgx |
|---|---|---|
| Скорость | Средняя | Высокая (нативный протокол) |
| Postgres-типы | Только базовые | Все: arrays, hstore, jsonb, uuid, inet |
| COPY protocol | Нет | Да |
| Connection pooling | Отдельный пакет | Встроен (`pgxpool`) |
| Prepared statements | Неявные | Явный контроль |

Для Postgres берём `pgx`. Если когда-нибудь понадобится абстракция над разными БД — есть `sqlx` и `sqlc`.

## Установка и подключение

```bash
go get github.com/jackc/pgx/v5
```

Простое подключение:

```go
import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5"
)

func main() {
    ctx := context.Background()

    conn, err := pgx.Connect(ctx, "postgres://user:pass@localhost:5432/marketplace")
    if err != nil {
        panic(err)
    }
    defer conn.Close(ctx)

    var version string
    err = conn.QueryRow(ctx, "SELECT version()").Scan(&version)
    if err != nil {
        panic(err)
    }
    fmt.Println(version)
}
```

Строка подключения (DSN):

```
postgres://user:password@host:port/dbname?sslmode=disable
```

:::warning Не хардкодь DSN
Строку подключения бери из переменной окружения `DATABASE_URL`. Никогда не комить пароли в коде.
:::

## Connection pool

Одно соединение (`pgx.Connect`) — только для простых скриптов. Для веб-сервера нужен пул:

```go
import "github.com/jackc/pgx/v5/pgxpool"

func main() {
    ctx := context.Background()

    pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
    if err != nil {
        panic(err)
    }
    defer pool.Close()

    // Пул готов к использованию
    var count int
    err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM products").Scan(&count)
}
```

`pgxpool` управляет соединениями автоматически: открывает по необходимости, переиспользует, закрывает простаивающие. По умолчанию максимум 4 соединения — для продакшена обычно `max_connections / 2` (если несколько инстансов приложения).

## CRUD с pgx

### Create

```go
func (r *ProductRepo) Create(ctx context.Context, p Product) (Product, error) {
    err := r.pool.QueryRow(ctx,
        `INSERT INTO products (title, description, price, seller_id)
         VALUES ($1, $2, $3, $4)
         RETURNING id, created_at`,
        p.Title, p.Description, p.Price, p.SellerID,
    ).Scan(&p.ID, &p.CreatedAt)

    return p, err
}
```

`$1, $2, ...` — плейсхолдеры pgx. Не `?` как в MySQL, не `:name` как в sqlx. Никакой конкатенации строк — SQL-инъекции исключены.

`RETURNING` — фича Postgres: вставляем строку и сразу получаем сгенерированные поля. Без второго запроса.

### Read (один)

```go
func (r *ProductRepo) FindByID(ctx context.Context, id int) (Product, error) {
    var p Product
    err := r.pool.QueryRow(ctx,
        `SELECT id, title, description, price, seller_id, created_at
         FROM products WHERE id = $1`, id,
    ).Scan(&p.ID, &p.Title, &p.Description, &p.Price, &p.SellerID, &p.CreatedAt)

    if err == pgx.ErrNoRows {
        return Product{}, fmt.Errorf("product not found")
    }
    return p, err
}
```

:::tip pgx.ErrNoRows
`database/sql` возвращает `sql.ErrNoRows`. `pgx` возвращает `pgx.ErrNoRows`. Не перепутай при миграции с `database/sql` на pgx.
:::

### Read (много)

```go
func (r *ProductRepo) FindAll(ctx context.Context) ([]Product, error) {
    rows, err := r.pool.Query(ctx,
        `SELECT id, title, price, created_at FROM products ORDER BY created_at DESC`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var products []Product
    for rows.Next() {
        var p Product
        if err := rows.Scan(&p.ID, &p.Title, &p.Price, &p.CreatedAt); err != nil {
            return nil, err
        }
        products = append(products, p)
    }
    return products, rows.Err()
}
```

Всегда проверяй `rows.Err()` после цикла — он мог завершиться по ошибке, а не потому что строки кончились.

### Update

```go
func (r *ProductRepo) Update(ctx context.Context, p Product) error {
    tag, err := r.pool.Exec(ctx,
        `UPDATE products SET title=$1, description=$2, price=$3
         WHERE id=$4 AND seller_id=$5`,
        p.Title, p.Description, p.Price, p.ID, p.SellerID,
    )
    if err != nil {
        return err
    }
    if tag.RowsAffected() == 0 {
        return fmt.Errorf("product not found or not owned by seller")
    }
    return nil
}
```

`Exec` возвращает `CommandTag` с методом `RowsAffected()`. Проверяем, что обновилась ровно одна строка.

### Delete

```go
func (r *ProductRepo) Delete(ctx context.Context, id, sellerID int) error {
    tag, err := r.pool.Exec(ctx,
        `DELETE FROM products WHERE id=$1 AND seller_id=$2`,
        id, sellerID,
    )
    if err != nil {
        return err
    }
    if tag.RowsAffected() == 0 {
        return fmt.Errorf("product not found")
    }
    return nil
}
```

## NULL-поля

В реальной БД поля могут быть NULL. Go-типы не бывают nil (кроме указателей и интерфейсов). pgx даёт типы-обёртки:

```go
type Product struct {
    ID          int
    Title       string
    Description pgtype.Text // может быть NULL
    Price       int         // NOT NULL
}

// Сканирование
var p Product
err := row.Scan(&p.ID, &p.Title, &p.Description, &p.Price)

// Использование
if p.Description.Valid {
    fmt.Println(p.Description.String)
}
```

Для простых случаев можно использовать указатели:

```go
type Product struct {
    Description *string // nil = NULL
}
```

## Миграции: первый взгляд

Вместо ручного создания таблиц через `psql`, используем `golang-migrate`:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Создаём первую миграцию:

```bash
migrate create -ext sql -dir migrations -seq create_products
```

Файл `000001_create_products.up.sql`:

```sql
CREATE TABLE products (
    id          SERIAL PRIMARY KEY,
    title       VARCHAR(200) NOT NULL,
    description TEXT,
    price       INTEGER NOT NULL CHECK (price >= 0),
    seller_id   INTEGER NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Подробно миграции разберём в главе 3.

## Ключевые выводы

1. `pgx` — нативный Postgres-драйвер, быстрее и полнее `database/sql`
2. `pgxpool` для веб-серверов — управляет пулом соединений
3. `$1, $2` — параметризованные запросы, защита от инъекций
4. `RETURNING` — получаем сгенерированные поля без второго запроса
5. Всегда проверяй `RowsAffected()` в Update/Delete и `rows.Err()` после цикла

В следующей главе посмотрим на GORM — ORM, который прячет SQL за методами, и сравним оба подхода.

---

## Проверь себя

<Quiz
  quizId="p2-01-sql-pgx"
  questions={[
    {
      id: "q1",
      question: "Почему pgx предпочтительнее database/sql для работы с PostgreSQL?",
      options: [
        "pgx — единственный драйвер, который работает с PostgreSQL",
        "pgx быстрее, поддерживает все Postgres-типы, COPY-протокол и connection pooling из коробки",
        "database/sql запрещён в production",
        "pgx написан на Rust и вызывается через CGo"
      ],
      correctIndex: 1,
      explanation: "pgx использует нативный бинарный протокол PostgreSQL, а не текстовый. Он быстрее, поддерживает специфичные Postgres-типы (arrays, jsonb, uuid) и имеет встроенный connection pool."
    },
    {
      id: "q2",
      question: "Зачем нужен RETURNING в INSERT-запросах?",
      options: [
        "Чтобы проверить, что данные вставились",
        "Чтобы получить сгенерированные поля (id, created_at) в одном запросе вместо двух",
        "Это синтаксический сахар, не влияет на производительность",
        "RETURNING нужен только для транзакций"
      ],
      correctIndex: 1,
      explanation: "Без RETURNING нужен отдельный SELECT для получения id и временных меток. RETURNING экономит один round-trip к БД."
    },
    {
      id: "q3",
      question: "Как pgx защищает от SQL-инъекций?",
      options: [
        "Никак, это ответственность разработчика",
        "Через $1, $2 плейсхолдеры — значения передаются отдельно от запроса, злоумышленник не может изменить структуру SQL",
        "Через ORM-надстройку над SQL",
        "pgx автоматически экранирует строки в запросе"
      ],
      correctIndex: 1,
      explanation: "Плейсхолдеры ($1, $2) передают значения как параметры prepared statement. Даже если значение содержит SQL-код, он будет воспринят как литерал, а не как часть запроса."
    }
  ]}
/>
