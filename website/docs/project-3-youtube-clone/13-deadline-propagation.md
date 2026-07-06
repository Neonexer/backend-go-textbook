---
title: "Deadline Propagation"
sidebar_position: 13
---

import Quiz from '@site/src/components/Quiz';

# Deadline Propagation в gRPC

Клиент вызывает API Gateway с таймаутом 5 секунд. API Gateway вызывает Video Service, Video Service — БД. Без propagation каждый шаг не знает общий дедлайн и может затянуть цепочку.

## Как это работает

Контекст запроса содержит дедлайн. gRPC автоматически передаёт его в метаданных:

```go
// Клиент: "весь запрос должен занять не больше 3 секунд"
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

resp, err := client.GetVideo(ctx, &video.GetVideoRequest{Id: "42"})
```

Сервер получает дедлайн из контекста:

```go
func (s *VideoServer) GetVideo(ctx context.Context, req *video.GetVideoRequest) (*video.Video, error) {
    deadline, ok := ctx.Deadline()
    if ok {
        remaining := time.Until(deadline)
        slog.Info("запрос", "deadline_in", remaining)
    }

    // Контекст автоматически отменится когда дедлайн истечёт
    v, err := s.repo.FindByID(ctx, req.Id)
    // ...
}
```

## Настройка на клиенте

```go
conn, _ := grpc.Dial("video-service:50051",
    grpc.WithTimeout(2*time.Second), // таймаут на установку соединения
)
```

## Правила дедлайнов

- API Gateway: 5 секунд
- Video Service: 5s − (время на Gateway) ≈ 4.5s
- БД: 4.5s − (время на Video) ≈ 4s

Каждый сервис в цепочке должен проверять `ctx.Deadline()` и не превышать оставшееся время.

<Quiz quizId="p3-13-deadline" questions={[
  {id:"q1",question:"Что будет если сервис проигнорирует дедлайн из контекста?",options:["Ничего","Клиент уже ушёл по таймауту — сервис делает бесполезную работу. Ресурсы тратятся впустую. gRPC отправит CANCELLED но если не проверять ctx — запрос продолжит выполняться.","gRPC автоматически остановит","Контекст не содержит дедлайн"],correctIndex:1,explanation:"Без проверки ctx.Done() или ctx.Err() сервис продолжит выполнять запрос, даже если клиент уже отвалился. Всегда проверяй ctx в долгих операциях."}
]} />
