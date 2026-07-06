---
title: "Saga"
sidebar_position: 5
---

import Quiz from '@site/src/components/Quiz';

# Saga — распределённые транзакции

В монолите: `BEGIN → INSERT order → UPDATE inventory → CHARGE payment → COMMIT`. Одна ACID-транзакция. В микросервисах: три разных сервиса, три разных БД. Нужна Saga.

## Два типа Saga

### Choreography (хореография)

Каждый сервис слушает события и реагирует:

```
Order Service:     "OrderCreated" → Kafka
Payment Service:   видит OrderCreated → списывает → "PaymentSucceeded"
Inventory Service: видит PaymentSucceeded → резервирует → "InventoryReserved"
```

Без оркестратора. Децентрализованно. Сложно отследить общий статус.

### Orchestration (оркестрация)

Оркестратор управляет процессом:

```go
func (s *SagaOrchestrator) CreateOrder(ctx context.Context, order Order) error {
    tx := s.saga.Begin("CreateOrder")

    tx.Step("reserve-inventory", func() error {
        return s.inventory.Reserve(ctx, order.Items)
    }).Compensate(func() error {
        return s.inventory.Release(ctx, order.Items)
    })

    tx.Step("charge-payment", func() error {
        return s.payment.Charge(ctx, order.Amount)
    }).Compensate(func() error {
        return s.payment.Refund(ctx, order.Amount)
    })

    return tx.Execute(ctx)
}
```

Каждый шаг имеет компенсацию — обратное действие если что-то пошло не так.

## Компенсации вместо отката

В распределённой системе нельзя просто `ROLLBACK`. Если списание прошло, а резервирование нет — нужно вернуть деньги (компенсация):

- INSERT → DELETE (логическое удаление)
- Списание → Возврат
- Отправка email → Извинительный email

<Quiz quizId="sd-05-saga" questions={[
  {id:"q1",question:"Почему в микросервисах нельзя использовать обычные ACID-транзакции?",options:["Можно, это лучшая практика","Транзакция охватывает одну БД. В микросервисах у каждого своя БД — нужна распределённая координация.","ACID не работает в Go","Микросервисы не используют БД"],correctIndex:1,explanation:"ACID работает в пределах одной БД. Saga координирует транзакцию через несколько сервисов и БД с помощью компенсаций вместо ROLLBACK."}
]} />
