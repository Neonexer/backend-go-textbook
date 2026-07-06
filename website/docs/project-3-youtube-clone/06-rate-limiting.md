---
title: "Rate Limiting"
sidebar_position: 6
---

import Quiz from '@site/src/components/Quiz';

# Rate Limiting

Без ограничений один пользователь может положить API тысячей запросов в секунду. Rate limiting защищает сервис от злоупотребления и обеспечивает честное распределение ресурсов.

## Token Bucket алгоритм

Самый распространённый алгоритм:

- «Ведро» вмещает N токенов
- Токены добавляются с постоянной скоростью (например, 10 в секунду)
- Каждый запрос забирает 1 токен
- Если токенов нет — запрос отклоняется (429 Too Many Requests)

```go
import "golang.org/x/time/rate"

// 10 запросов/сек, burst до 20
limiter := rate.NewLimiter(10, 20)

if !limiter.Allow() {
    http.Error(w, "слишком много запросов", http.StatusTooManyRequests)
    return
}
```

## Middleware для rate limiting

```go
func rateLimitMiddleware(rps int) func(http.Handler) http.Handler {
    limiters := sync.Map{} // user_id → *rate.Limiter

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID, _ := r.Context().Value("user_id").(int)
            key := strconv.Itoa(userID)

            lim, _ := limiters.LoadOrStore(key, rate.NewLimiter(rate.Limit(rps), rps*2))
            if !lim.(*rate.Limiter).Allow() {
                writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

## Rate limiting в распределённой системе

In-memory ограничения работают только на одном инстансе. Для кластера нужно внешнее хранилище:

```go
// Redis-реализация
func isAllowed(ctx context.Context, rdb *redis.Client, key string, limit int, window time.Duration) bool {
    pipe := rdb.Pipeline()
    now := time.Now().UnixNano()
    windowStart := now - window.Nanoseconds()

    pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))
    pipe.ZCard(ctx, key)
    pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})

    cmds, _ := pipe.Exec(ctx)
    count := cmds[1].(*redis.IntCmd).Val()

    if count >= int64(limit) {
        return false
    }
    return true
}
```

## Заголовки rate limit

Хороший тон — сообщать клиенту лимиты:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 73
X-RateLimit-Reset: 1620000000
Retry-After: 5
```

## Ключевые выводы

- Token bucket — простой и эффективный
- In-memory для одного инстанса, Redis для кластера
- Заголовки `X-RateLimit-*` — уважай клиента
- Rate limiting защищает от злоупотреблений, а не от DDoS (для этого внешний CDN/WAF)

<Quiz quizId="p3-06-rate-limit" questions={[
  {id:"q1",question:"Почему in-memory rate limit не работает в кластере из 3 инстансов?",options:["Работает, это лучший вариант","Каждый инстанс считает свои запросы — пользователь может получить 3× лимит распределяя запросы между инстансами","In-memory быстрее кластерного","Кластер не поддерживает rate limiting"],correctIndex:1,explanation:"Счётчик в памяти одного инстанса не видит запросы на других. Пользователь может обойти лимит, распределив запросы. Для кластера — Redis или другой общий счётчик."}
]} />
