---
title: "Миграции с Goose"
sidebar_position: 10
---

import Quiz from '@site/src/components/Quiz';

# Миграции с Goose

В главе 3 мы использовали `golang-migrate`. Альтернатива — **Goose**: миграции на чистом Go (не только SQL), встроенная поддержка переменных окружения и транзакций. Выбор зависит от проекта.

## golang-migrate vs Goose

| | golang-migrate | Goose |
|---|---|---|
| Язык миграций | Только SQL | SQL + Go |
| Транзакции | Автоматически на каждую миграцию | Явный контроль |
| Переменные окружения | Нет (чистый SQL) | Встроены |
| Версионирование | Числовое | Временные метки |
| CLI | `migrate` | `goose` |

Goose удобнее когда нужна логика в миграциях (data migration, не только DDL).

## Установка

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

## SQL-миграции с Goose

```bash
goose create add_products sql
# → 20260707120000_add_products.sql
```

```sql
-- +goose Up
CREATE TABLE products (
    id   SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL
);

-- +goose Down
DROP TABLE products;
```

Аннотации `-- +goose Up` и `-- +goose Down` разделяют направления в одном файле (в отличие от golang-migrate где два файла).

## Go-миграции

Главное преимущество Goose — миграции на Go:

```go
// 20260707120000_migrate_prices.go
// +goose Up
func Up(tx *sql.Tx) error {
    rows, _ := tx.Query("SELECT id, price FROM products")
    for rows.Next() {
        var id, oldPrice int
        rows.Scan(&id, &oldPrice)
        // Переводим цены из рублей в копейки
        tx.Exec("UPDATE products SET price = $1 WHERE id = $2",
            oldPrice*100, id)
    }
    return nil
}

// +goose Down
func Down(tx *sql.Tx) error {
    _, err := tx.Exec("UPDATE products SET price = price / 100")
    return err
}
```

Go-миграции незаменимы когда нужно перенести данные с преобразованием.

## Запуск из кода

```go
import "github.com/pressly/goose/v3"

func runMigrations(db *sql.DB) error {
    return goose.Up(db, "migrations")
}
```

## Переменные окружения в SQL-миграциях

Goose подставляет переменные в SQL:

```sql
-- +goose Up
CREATE TABLE products (
    owner VARCHAR(255) DEFAULT '${DEFAULT_OWNER}'
);
```

```bash
DEFAULT_OWNER=admin goose up
```

<Quiz quizId="p2-10-goose" questions={[
  {id:"q1",question:"Когда выбирать Goose вместо golang-migrate?",options:["Всегда","Когда нужны Go-миграции (data migration с логикой) или переменные окружения в SQL","Только для маленьких проектов","Goose deprecated"],correctIndex:1,explanation:"Goose shines когда миграция не просто DDL, а содержит логику: перенос данных, трансформация, вызов внешнего API."}
]} />
