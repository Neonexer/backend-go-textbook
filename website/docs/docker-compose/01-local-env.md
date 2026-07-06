---
title: "Docker Compose для локальной разработки"
sidebar_position: 1
---

import Quiz from '@site/src/components/Quiz';

# Docker Compose для локальной разработки

Три проекта, PostgreSQL, Redis, Kafka — запускать всё вручную больно. Docker Compose поднимает всё одной командой.

## docker-compose.yaml

```yaml
version: "3.9"

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: marketplace
      POSTGRES_USER: app
      POSTGRES_PASSWORD: secret
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  kafka:
    image: bitnami/kafka:3.6
    environment:
      KAFKA_CFG_NODE_ID: 0
      KAFKA_CFG_PROCESS_ROLES: controller,broker
      KAFKA_CFG_LISTENERS: PLAINTEXT://:9092,CONTROLLER://:9093
      KAFKA_CFG_CONTROLLER_QUORUM_VOTERS: 0@kafka:9093
      KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP: CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT
    ports:
      - "9092:9092"

  # Проект 1: блог
  blog:
    build: ../code/project-1-blog
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://app:secret@postgres:5432/blog?sslmode=disable
    depends_on:
      - postgres

  # Проект 2: маркетплейс
  marketplace:
    build: ../code/project-2-marketplace
    ports:
      - "8081:8080"
    environment:
      DATABASE_URL: postgres://app:secret@postgres:5432/marketplace?sslmode=disable
    depends_on:
      - postgres

volumes:
  pgdata:
```

## Dockerfile для Go (единый)

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /server /server
COPY migrations/ /migrations/
EXPOSE 8080
CMD ["/server"]
```

## Запуск

```bash
docker compose up -d          # запустить всё
docker compose logs -f blog   # логи конкретного сервиса
docker compose down           # остановить и удалить
```

<Quiz quizId="docker-01" questions={[
  {id:"q1",question:"Зачем volumes для PostgreSQL в Docker Compose?",options:["Для скорости","Без volumes данные исчезнут при docker compose down. Volume сохраняет БД между перезапусками.","Это требование Docker","Для бекапов"],correctIndex:1,explanation:"Контейнеры stateless по умолчанию. Volume монтирует папку на хосте в контейнер — данные переживают docker compose down и пересоздание контейнера."}
]} />
