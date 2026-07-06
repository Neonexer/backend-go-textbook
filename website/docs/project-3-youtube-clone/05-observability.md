---
title: "Observability"
sidebar_position: 5
---

import Quiz from '@site/src/components/Quiz';

# Observability

Микросервисов много. Когда видео не загружается — где проблема? Observability — это три столпа: **логи**, **метрики** и **трейсинг**. Вместе они позволяют понять что происходит в распределённой системе.

## Три столпа

| Столп | Вопрос | Инструмент |
|-------|--------|-----------|
| **Логи** | Что произошло? | slog → Loki / ELK |
| **Метрики** | Сколько и как быстро? | Prometheus + Grafana |
| **Трейсинг** | Где именно тормозит? | OpenTelemetry → Jaeger |

## OpenTelemetry

OpenTelemetry — единый стандарт для сбора телеметрии. Добавляем в Go:

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/trace"
)

func initTracer() (*trace.TracerProvider, error) {
    exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint("jaeger:4317"))
    tp := trace.NewTracerProvider(trace.WithBatcher(exporter))
    otel.SetTracerProvider(tp)
    return tp, nil
}
```

## Спаны в gRPC

Каждый запрос — span с атрибутами:

```go
func (s *VideoServer) GetVideo(ctx context.Context, req *video.GetVideoRequest) (*video.Video, error) {
    tracer := otel.Tracer("video-service")
    ctx, span := tracer.Start(ctx, "GetVideo")
    defer span.End()

    span.SetAttributes(attribute.String("video.id", req.Id))

    v, err := s.svc.Get(ctx, req.Id)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }
    return v, nil
}
```

Контекст передаётся через gRPC автоматически — span'ы связываются в цепочку от API Gateway до БД.

## Prometheus-метрики

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{Name: "http_requests_total"},
        []string{"method", "path", "status"},
    )
    httpDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{Name: "http_request_duration_seconds"},
        []string{"method", "path"},
    )
)
```

В middleware:

```go
func metricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        timer := prometheus.NewTimer(httpDuration.WithLabelValues(r.Method, r.URL.Path))
        defer timer.ObserveDuration()

        ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
        next.ServeHTTP(ww, r)

        httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path,
            strconv.Itoa(ww.Status())).Inc()
    })
}
```

## RED-метрики

Для каждого сервиса важны три метрики (RED):

- **R**ate — количество запросов в секунду
- **E**rrors — доля ошибок (4xx/5xx)
- **D**uration — время ответа (p50, p95, p99)

## Ключевые выводы

- Логи, метрики, трейсинг — не выбирай что-то одно, используй все три
- OpenTelemetry — стандарт, не привязывайся к конкретному вендору
- RED-метрики для каждого сервиса
- Span'ы передаются через контекст автоматически в gRPC

<Quiz quizId="p3-05-observability" questions={[
  {id:"q1",question:"Зачем нужен трейсинг если есть логи и метрики?",options:["Трейсинг заменяет логи","Трейсинг показывает путь запроса через все сервисы — видно где именно задержка в цепочке из 5+ микросервисов","Трейсинг бесплатный","Трейсинг генерирует метрики"],correctIndex:1,explanation:"Логи показывают что произошло в одном сервисе. Метрики — агрегированные показатели. Трейсинг связывает всё в одну цепочку: запрос прошёл API Gateway → Video → DB за 200ms, из которых 150ms занял Video → DB."}
]} />
