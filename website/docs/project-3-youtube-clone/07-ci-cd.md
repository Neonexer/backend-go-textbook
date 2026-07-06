---
title: "CI/CD"
sidebar_position: 7
---

import Quiz from '@site/src/components/Quiz';

# CI/CD

Код без автоматической сборки и деплоя — ручной труд и ошибки. CI/CD автоматизирует проверку, сборку и доставку кода в продакшен.

## CI: Continuous Integration

На каждый push в main (и PR) GitHub Actions запускает:

```yaml
name: CI
on: [push, pull_request]

jobs:
  lint-and-test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_DB: test
          POSTGRES_PASSWORD: test
        ports: ["5432:5432"]

    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: "1.22" }

      - name: Lint
        run: go vet ./...

      - name: Test
        run: go test ./... -count=1
        env:
          DATABASE_URL: postgres://postgres:test@localhost:5432/test?sslmode=disable
```

## CD: Continuous Delivery

После прохождения CI — деплой. Для микросервисов:

```yaml
deploy:
  needs: lint-and-test
  runs-on: ubuntu-latest
  steps:
    - name: Build and push Docker image
      run: |
        docker build -t ghcr.io/go-course/video-service:${{ github.sha }} .
        docker push ghcr.io/go-course/video-service:${{ github.sha }}

    - name: Deploy to Kubernetes
      run: |
        kubectl set image deployment/video-service \
          video-service=ghcr.io/go-course/video-service:${{ github.sha }}
```

## Dockerfile для Go

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

Multi-stage build: бинарник собирается в `golang`-образе (800 MB), копируется в `alpine` (15 MB).

## Ключевые выводы

- CI проверяет код (lint, test) на каждый push
- CD деплоит после CI
- Multi-stage Dockerfile для минимального образа
- Service containers в GitHub Actions (Postgres, Kafka) для интеграционных тестов

<Quiz quizId="p3-07-cicd" questions={[
  {id:"q1",question:"Зачем нужен multi-stage Dockerfile?",options:["Для скорости сборки","Итоговый образ содержит только бинарник и сертификаты (15 MB), а не весь Go SDK (800 MB)","Это требование Docker Hub","Для совместимости с Kubernetes"],correctIndex:1,explanation:"Builder stage компилирует бинарник в golang-образе. Final stage копирует только бинарник в alpine — минимальный образ, быстрый деплой, меньше уязвимостей."}
]} />
