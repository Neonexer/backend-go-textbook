---
title: "Пагинация и поиск"
sidebar_position: 7
---

import Quiz from '@site/src/components/Quiz';

# Пагинация и поиск

Когда товаров больше 20 — отдавать их все за один запрос нельзя. Пагинация разбивает результат на страницы, поиск фильтрует по запросу.

## Offset vs Cursor

### Offset-based

```sql
SELECT * FROM products ORDER BY created_at DESC LIMIT 20 OFFSET 40;
```

Простая, понятная. **Проблема**: при добавлении новых записей страницы «съезжают» (пропуск/дубликаты). Подходит для админок и небольших объёмов.

### Cursor-based

```sql
SELECT * FROM products WHERE created_at < $1 ORDER BY created_at DESC LIMIT 20;
```

Клиент передаёт курсор — значение сортировочного поля последнего элемента. Не съезжает при вставках. **Проблема**: нельзя перейти на произвольную страницу.

:::tip Когда что
**Offset** — для админок, search results, «страница 3 из 50». **Cursor** — для infinite scroll ленты (Twitter, Instagram).
:::

## Реализация offset-пагинации

```go
type Pagination struct {
    Page       int `json:"page"`
    PerPage    int `json:"per_page"`
    Total      int `json:"total"`
    TotalPages int `json:"total_pages"`
}

func (r *ProductRepo) FindPaginated(ctx context.Context, page, perPage int) ([]Product, Pagination, error) {
    // Общее количество
    var total int
    r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM products`).Scan(&total)

    offset := (page - 1) * perPage
    rows, err := r.pool.Query(ctx,
        `SELECT id, title, price FROM products ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
        perPage, offset,
    )
    // ... scan rows

    pagination := Pagination{
        Page: page, PerPage: perPage, Total: total,
        TotalPages: (total + perPage - 1) / perPage,
    }
    return products, pagination, err
}
```

## Полнотекстовый поиск

Для поиска по товарам Postgres даёт `tsvector`:

```sql
-- Добавляем колонку для поиска
ALTER TABLE products ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (to_tsvector('russian', title || ' ' || description)) STORED;

CREATE INDEX idx_products_search ON products USING GIN(search_vector);

-- Поиск
SELECT * FROM products WHERE search_vector @@ plainto_tsquery('russian', 'кроссовки nike');
```

В Go:

```go
func (r *ProductRepo) Search(ctx context.Context, query string) ([]Product, error) {
    rows, err := r.pool.Query(ctx,
        `SELECT id, title, price FROM products
         WHERE search_vector @@ plainto_tsquery('russian', $1)
         ORDER BY ts_rank(search_vector, plainto_tsquery('russian', $1)) DESC
         LIMIT 50`, query,
    )
    // ...
}
```

`ts_rank` ранжирует результаты по релевантности — самые подходящие сверху.

## Ключевые выводы

- Offset для страниц с номерами, Cursor для бесконечной ленты
- `COUNT(*) + LIMIT + OFFSET` = простая пагинация
- `tsvector` + GIN индекс для полнотекстового поиска
- Всегда ограничивай `LIMIT` — без него можно положить БД

<Quiz quizId="p2-07-pagination" questions={[
  {id:"q1",question:"В чём проблема offset-пагинации при частых вставках?",options:["Она медленнее","Страницы съезжают: новая запись сдвигает все offset'ы","OFFSET не работает в PostgreSQL","Это deprecated метод"],correctIndex:1,explanation:"При вставке новой записи в начало, offset=20 покажет ту же запись что была на offset=10 до вставки. Cursor-based пагинация лишена этой проблемы."}
]} />
