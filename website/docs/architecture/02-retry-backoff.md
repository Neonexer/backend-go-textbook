---
title: "Retry + Backoff"
sidebar_position: 2
---

import Quiz from '@site/src/components/Quiz';

# Retry + Exponential Backoff

Сетевой запрос упал. Повторить? Да, но не мгновенно — иначе лавина ретраев положит сервис ещё быстрее.

## Простой retry с backoff

```go
func withRetry(ctx context.Context, maxAttempts int, fn func() error) error {
    var err error
    for attempt := 0; attempt < maxAttempts; attempt++ {
        if attempt > 0 {
            backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
            select {
            case <-time.After(backoff):
            case <-ctx.Done():
                return ctx.Err()
            }
        }

        err = fn()
        if err == nil {
            return nil
        }

        // Не ретраим клиентские ошибки (4xx)
        if isClientError(err) {
            return err
        }
    }
    return fmt.Errorf("all %d attempts failed: %w", maxAttempts, err)
}
```

## Exponential backoff + jitter

Без jitter'а 100 клиентов одновременно упадут и одновременно повторят — снова нагрузка на сервис:

```go
backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
time.Sleep(backoff + jitter)
```

## Полная цепочка устойчивости

```
Request → Retry (exponential + jitter)
       → Circuit Breaker (стоп если сервис мёртв)
       → Timeout (не ждать бесконечно)
       → Fallback (кеш, дефолтное значение)
```

## Библиотеки

`cenkalti/backoff` — готовые стратегии:

```go
b := backoff.NewExponentialBackOff()
b.MaxElapsedTime = 30 * time.Second

backoff.Retry(func() error {
    return callExternalAPI()
}, b)
```

<Quiz quizId="arch-02-retry" questions={[
  {id:"q1",question:"Зачем нужен jitter к exponential backoff?",options:["Для красоты","Без jitter'а все клиенты повторяют одновременно в одни и те же интервалы — создают пиковую нагрузку. Jitter размазывает повторы по времени.","Jitter ускоряет retry","Это требование HTTP"],correctIndex:1,explanation:"100 клиентов без jitter'а = 100 одновременных ретраев через 1, 2, 4, 8 секунд. С jitter'ом они распределяются в окне и не создают пиков."}
]} />
