---
title: "Idempotency Keys"
sidebar_position: 1
---

import Quiz from '@site/src/components/Quiz';

# Idempotency Keys

Пользователь нажал «Оплатить», запрос ушёл, ответ не пришёл (таймаут). Пользователь нажимает ещё раз. Без идемпотентности — двойное списание. Idempotency key решает эту проблему.

## Как это работает

Клиент генерирует уникальный ключ для каждой операции. Сервер по ключу понимает что запрос уже был обработан:

```
POST /orders
Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000
```

```go
func (s *OrderService) Create(ctx context.Context, key string, req CreateOrderReq) (Order, error) {
    // Проверяем: был ли уже запрос с таким ключом?
    existing, err := s.repo.FindByIdempotencyKey(ctx, key)
    if err == nil {
        return existing, nil // вернуть тот же результат
    }

    // Новый запрос — выполняем
    order, err := s.repo.Create(ctx, req)
    if err != nil {
        return Order{}, err
    }

    // Сохраняем ключ + результат
    s.repo.SaveIdempotencyKey(ctx, key, order.ID)

    return order, nil
}
```

## Хранение ключей

Таблица `idempotency_keys`:

```sql
CREATE TABLE idempotency_keys (
    key         UUID PRIMARY KEY,
    response    JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Автоочистка старых ключей (через 24 часа)
CREATE INDEX idx_idempotency_created ON idempotency_keys(created_at);
```

## Ключевые выводы

- Idempotency key защищает от дублирования при retry'ях
- Ключ генерирует клиент, сервер хранит ключ + результат
- Очищай старые ключи (TTL 24 часа)
- Особенно критично для: платежи, создание заказов, списание баланса

<Quiz quizId="pp-01-idempotency" questions={[
  {id:"q1",question:"Почему idempotency key должен генерировать клиент а не сервер?",options:["Сервер не умеет генерировать UUID","Клиент знает что это тот же самый запрос при ретрае. Сервер не может отличить повторный запрос от нового без ключа.","Серверные ключи не уникальны","Это требование HTTP"],correctIndex:1,explanation:"При таймауте сервер мог уже обработать запрос но клиент не получил ответ. Клиент повторяет с тем же ключом — сервер понимает что это ретрай."}
]} />
