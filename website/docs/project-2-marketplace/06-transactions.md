---
title: "Транзакции"
sidebar_position: 6
---

import Quiz from '@site/src/components/Quiz';

# Транзакции и конкурентный доступ

Когда два продавца одновременно покупают последний товар, или платёж проходит а списание не происходит — нужны транзакции. В этой главе: ACID, уровни изоляции и конкурентный доступ в PostgreSQL.

## ACID за 30 секунд

- **Atomicity** — всё или ничего: либо все операции транзакции выполняются, либо ни одна
- **Consistency** — constraints выполняются после транзакции
- **Isolation** — параллельные транзакции не мешают друг другу
- **Durability** — закоммиченные данные не исчезнут

## Транзакции в pgx

```go
func (r *OrderRepo) CreateOrder(ctx context.Context, order Order) error {
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx) // откат если не закоммитили

    // 1. Проверить наличие товара
    var available bool
    tx.QueryRow(ctx, `SELECT quantity > 0 FROM products WHERE id=$1 FOR UPDATE`,
        order.ProductID).Scan(&available)
    if !available {
        return fmt.Errorf("товар недоступен")
    }

    // 2. Уменьшить количество
    tx.Exec(ctx, `UPDATE products SET quantity = quantity - 1 WHERE id=$1`,
        order.ProductID)

    // 3. Создать заказ
    tx.Exec(ctx, `INSERT INTO orders (product_id, buyer_id, price) VALUES ($1,$2,$3)`,
        order.ProductID, order.BuyerID, order.Price)

    return tx.Commit(ctx)
}
```

`defer tx.Rollback(ctx)` — страховка. Если `Commit` не вызван (ошибка, паника) — `Rollback` отменит все изменения.

## FOR UPDATE — блокировка строки

```sql
SELECT quantity FROM products WHERE id = $1 FOR UPDATE
```

`FOR UPDATE` блокирует строку до конца транзакции. Другая транзакция, которая тоже сделает `FOR UPDATE` на ту же строку, будет **ждать**. Это гарантирует, что два покупателя не купят последний товар.

:::tip Не злоупотребляй FOR UPDATE
`FOR UPDATE` сериализует доступ к строке — снижает конкурентность. Используй только когда действительно нужно атомарное чтение-запись.
:::

## Уровни изоляции

| Уровень | Грязное чтение | Неповторяемое чтение | Фантомы | Производительность |
|---------|---------------|---------------------|---------|-------------------|
| Read Uncommitted | Да | Да | Да | Максимальная |
| Read Committed | Нет | Да | Да | Высокая (default) |
| Repeatable Read | Нет | Нет | Нет* | Средняя |
| Serializable | Нет | Нет | Нет | Низкая |

PostgreSQL default — Read Committed. Для большинства операций этого достаточно. Для перевода денег между счетами — Repeatable Read или Serializable.

```go
tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{
    IsoLevel: pgx.RepeatableRead,
})
```

## Оптимистичная блокировка

Альтернатива `FOR UPDATE` — версионирование:

```sql
-- Добавляем колонку version
UPDATE products SET quantity = quantity - 1, version = version + 1
WHERE id = $1 AND version = $2
```

Если `RowsAffected() == 0` — кто-то уже изменил строку, нужно повторить. Это **оптимистичная** блокировка — не блокирует, а проверяет при записи.

## Ключевые выводы

- Транзакция: Begin → операции → Commit (или Rollback)
- `defer tx.Rollback()` всегда
- `FOR UPDATE` для атомарного чтения-записи
- Оптимистичная блокировка (version) для сценариев с низкой конкуренцией

<Quiz quizId="p2-06-transactions" questions={[
  {id:"q1",question:"Зачем нужен FOR UPDATE в транзакции?",options:["Для ускорения запроса","Чтобы заблокировать строку и гарантировать что другой запрос не изменит её до конца транзакции","Это синтаксический сахар","Для создания индекса"],correctIndex:1,explanation:"FOR UPDATE блокирует выбранные строки до COMMIT/ROLLBACK. Другая транзакция с FOR UPDATE на те же строки будет ждать — это гарантирует консистентность."},
  {id:"q2",question:"Почему всегда пишут defer tx.Rollback()?",options:["Традиция","Rollback ничего не делает после Commit, но если транзакция не закоммитилась (ошибка, паника) — откатит изменения","Это обязательно для PostgreSQL","Go требует defer для всех транзакций"],correctIndex:1,explanation:"Rollback после Commit — no-op. Но если произошла ошибка или паника до Commit, Rollback гарантирует что изменения не применятся."}
]} />
