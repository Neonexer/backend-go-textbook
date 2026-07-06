---
title: "Circuit Breaker"
sidebar_position: 10
---

import Quiz from '@site/src/components/Quiz';

# Circuit Breaker

Микросервисы вызывают друг друга. Если Video Service упал — Comment Service не должен бесконечно пытаться достучаться и исчерпывать ресурсы. Circuit Breaker (автоматический выключатель) предотвращает каскадные отказы.

## Состояния

```
         ┌──────────────┐
    ┌───▶│   CLOSED     │───┐ ошибок > порог
    │    │ (всё ок)     │   │
    │    └──────────────┘   │
    │                       ▼
    │               ┌──────────────┐
    │    таймаут    │    OPEN      │
    │    истёк      │ (запросы     │
    │               │  не идут)    │
    │               └──────┬───────┘
    │                      │
    └──────────────────────┘
              HALF-OPEN
          (пробный запрос)
```

- **Closed** — нормальная работа. Считаем ошибки.
- **Open** — ошибок > N за интервал. Запросы сразу отклоняются, не доходя до сервиса.
- **Half-Open** — через таймаут пропускаем 1 пробный запрос. Если успех → Closed, если ошибка → Open.

## Реализация в Go

```go
import "github.com/sony/gobreaker"

var cb = gobreaker.NewCircuitBreaker(gobreaker.Settings{
    Name:        "video-service",
    MaxRequests: 3,                     // запросов в half-open
    Interval:    60 * time.Second,      // окно подсчёта ошибок
    Timeout:     30 * time.Second,      // таймаут в open
    ReadyToTrip: func(counts gobreaker.Counts) bool {
        failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
        return counts.Requests >= 5 && failureRatio >= 0.5
    },
})

func GetVideo(ctx context.Context, id string) (*Video, error) {
    result, err := cb.Execute(func() (any, error) {
        return client.GetVideo(ctx, &video.GetVideoRequest{Id: id})
    })
    if err != nil {
        return nil, err
    }
    return result.(*Video), nil
}
```

## Где ставить circuit breaker

- **На исходящие вызовы** (клиентская сторона): Comment Service → Video Service
- **Перед внешними API**: платежи, отправка SMS
- **Не на health checks** — они должны проходить всегда

## Ключевые выводы

- Circuit breaker предотвращает каскадные отказы
- Не даёт тратить ресурсы на заведомо мёртвый сервис
- Half-open — самоисцеление без ручного вмешательства

<Quiz quizId="p3-10-circuit-breaker" questions={[
  {id:"q1",question:"Что произойдёт без circuit breaker'а если Video Service упадёт?",options:["Ничего","Comment Service будет ждать таймаута для каждого запроса, исчерпает connection pool и тоже упадёт — каскадный отказ","Запросы автоматически перенаправятся","gRPC сам обработает"],correctIndex:1,explanation:"Без CB каждый запрос к мёртвому сервису висит до таймаута, занимая горутину и соединение. Через минуту весь пул занят — сервис не отвечает никому."}
]} />
