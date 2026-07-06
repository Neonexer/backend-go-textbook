---
title: "Event Sourcing"
sidebar_position: 6
---

import Quiz from '@site/src/components/Quiz';

# Event Sourcing

Традиционно мы храним текущее состояние: `product.price = 5000`. Event Sourcing хранит **события**: `PriceSet(4000) → PriceIncreased(500) → PriceDiscounted(500)`. Текущее состояние вычисляется из журнала событий.

## Как это работает

```go
type Event interface {
    Apply(state *Product) error
}

type PriceChanged struct {
    NewPrice int
}

func (e PriceChanged) Apply(p *Product) error {
    p.Price = e.NewPrice
    return nil
}

// Восстановление состояния
func rebuildState(events []Event) *Product {
    p := &Product{}
    for _, e := range events {
        e.Apply(p)
    }
    return p
}
```

## Преимущества

- **Полный аудит** — каждое изменение сохранено
- **Отладка** — можно пересобрать состояние на любой момент времени
- **CQRS** — события идеально стыкуются с read-моделью

## Недостатки

- **Объём хранения** — все события vs одно текущее состояние
- **Сложность запросов** — «покажи текущую цену» требует пересчёта всех событий (решается снапшотами)
- **Eventual consistency** — read-модель обновляется асинхронно

<Quiz quizId="sd-06-es" questions={[
  {id:"q1",question:"Когда Event Sourcing стоит использовать а когда нет?",options:["Всегда","Когда нужен полный аудит (финансы) или полная история изменений — да. Для блога с ценой товара — избыточно.","Только для микросервисов","Только для монолитов"],correctIndex:1,explanation:"Event Sourcing даёт аудит и возможность отката, но требует хранения всех событий. Для простого CRUD это overengineering."}
]} />
