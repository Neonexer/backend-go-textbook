---
title: "Кеширование с Redis"
sidebar_position: 4
---

import Quiz from '@site/src/components/Quiz';

# Кеширование с Redis

Главная страница YouTube-клона показывает популярные видео. Каждый раз ходить в БД за одним и тем же списком — расточительно. Redis хранит данные в памяти и отдаёт их за микросекунды.

## Почему Redis

- **Скорость**: in-memory, `&lt;1ms` на операцию
- **Структуры данных**: не только key-value, но и списки, множества, sorted sets, hash maps
- **TTL**: автоматическое удаление устаревших данных
- **Pub/Sub**: можно использовать как message bus (но Kafka надёжнее)

## Подключение

```go
import "github.com/redis/go-redis/v9"

rdb := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

ctx := context.Background()
err := rdb.Set(ctx, "key", "value", 10*time.Minute).Err()
val, err := rdb.Get(ctx, "key").Result()
```

## Стратегии кеширования

### Cache-Aside (Look-Aside)

Самая распространённая:

```go
func (s *VideoService) GetVideo(ctx context.Context, id string) (Video, error) {
    // 1. Попробовать кеш
    cacheKey := "video:" + id
    cached, err := s.redis.Get(ctx, cacheKey).Result()
    if err == nil {
        var v Video
        json.Unmarshal([]byte(cached), &v)
        return v, nil
    }

    // 2. В кеше нет — запрос в БД
    v, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return Video{}, err
    }

    // 3. Сохранить в кеш
    data, _ := json.Marshal(v)
    s.redis.Set(ctx, cacheKey, data, 5*time.Minute)

    return v, nil
}
```

### Инвалидация

Самая сложная часть кеширования — понять когда данные устарели:

```go
// При обновлении видео — сбросить кеш
func (s *VideoService) Update(ctx context.Context, v Video) error {
    if err := s.repo.Update(ctx, v); err != nil {
        return err
    }
    s.redis.Del(ctx, "video:"+v.ID)
    return nil
}
```

### TTL и staggered expiry

```go
// Не ставь одинаковый TTL — это вызовет лавину запросов к БД
ttl := 5*time.Minute + time.Duration(rand.Intn(60))*time.Second
s.redis.Set(ctx, key, data, ttl)
```

## Что кешировать

- ✅ Главная страница (популярные видео) — TTL 1-5 минут
- ✅ Метаданные видео (название, описание) — TTL 10 минут, инвалидация при обновлении
- ✅ Сессии пользователей — TTL = время жизни токена
- ❌ Комментарии (слишком часто меняются)
- ❌ Счётчики просмотров (лучше атомарный INCR в Redis)

## Ключевые выводы

- Cache-Aside: кеш → БД → сохранить в кеш
- TTL обязателен, иначе утечка памяти
- Инвалидация при обновлении
- Staggered expiry чтобы избежать лавины

<Quiz quizId="p3-04-redis" questions={[
  {id:"q1",question:"Зачем добавлять случайную дельту к TTL?",options:["Для безопасности","Чтобы не все ключи истекали одновременно — иначе будет лавина запросов к БД","Это требование Redis","Для уникальности ключей"],correctIndex:1,explanation:"Если 1000 ключей истекут одновременно, 1000 запросов пойдут в БД в одну секунду. Staggered expiry размазывает их по минуте."}
]} />
