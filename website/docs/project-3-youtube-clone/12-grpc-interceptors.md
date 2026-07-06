---
title: "gRPC Interceptors"
sidebar_position: 12
---

import Quiz from '@site/src/components/Quiz';

# gRPC Interceptors

В REST мы используем middleware. В gRPC — interceptors. Тот же принцип: перехватить запрос до/после обработчика.

## Unary Interceptor

```go
func loggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
    start := time.Now()
    slog.Info("gRPC запрос", "method", info.FullMethod)

    resp, err := handler(ctx, req)

    slog.Info("gRPC ответ",
        "method", info.FullMethod,
        "duration", time.Since(start),
        "err", err,
    )
    return resp, err
}

grpcServer := grpc.NewServer(
    grpc.UnaryInterceptor(loggingInterceptor),
)
```

## Stream Interceptor

```go
func streamInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
    slog.Info("стрим", "method", info.FullMethod)
    return handler(srv, ss)
}

grpcServer := grpc.NewServer(
    grpc.UnaryInterceptor(loggingInterceptor),
    grpc.StreamInterceptor(streamInterceptor),
)
```

## Цепочка интерсепторов

В отличие от chi где middleware вешаются последовательно, gRPC требует явную цепочку:

```go
import "google.golang.org/grpc/chain"

grpcServer := grpc.NewServer(
    grpc.ChainUnaryInterceptor(
        recoveryInterceptor,
        loggingInterceptor,
        authInterceptor,
        validationInterceptor,
    ),
)
```

## Типичные интерсепторы

- **Логирование** — метод, длительность, ошибка
- **Аутентификация** — проверка JWT в метаданных
- **Трейсинг** — создание span'ов
- **Recovery** — ловля паник
- **Валидация** — protobuf validation

<Quiz quizId="p3-12-interceptors" questions={[
  {id:"q1",question:"Чем gRPC interceptor отличается от HTTP middleware?",options:["Ничем","Тем же, но для gRPC: перехватывает unary и stream вызовы вместо HTTP-запросов. Имеет доступ к gRPC-метаданным и статус-кодам.","Interceptor быстрее","Middleware работает только в HTTP/1"],correctIndex:1,explanation:"Interceptor — это middleware для gRPC. Он видит protobuf-сообщения, gRPC-статусы, метаданные. HTTP middleware видит заголовки и байты."}
]} />
