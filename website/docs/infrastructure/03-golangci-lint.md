---
title: "golangci-lint"
sidebar_position: 3
---

import Quiz from '@site/src/components/Quiz';

# golangci-lint — статический анализ

`go vet` проверяет базовые ошибки. `golangci-lint` запускает десятки линтеров одновременно и ловит баги, стилистические проблемы и потенциальные уязвимости до того как они попадут в продакшен.

## Установка и запуск

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run ./...
```

## Ключевые линтеры

| Линтер | Что проверяет |
|--------|--------------|
| `errcheck` | Необработанные ошибки |
| `govet` | Подозрительные конструкции |
| `staticcheck` | Неправильное использование API |
| `gosec` | Уязвимости безопасности |
| `bodyclose` | Незакрытые HTTP-тела |
| `sqlclosecheck` | Незакрытые SQL-rows |
| `gocritic` | Упрощение кода |
| `gocyclo` | Слишком сложные функции |

## Конфигурация

```yaml
# .golangci.yml
linters:
  enable:
    - errcheck
    - govet
    - staticcheck
    - gosec
    - bodyclose
    - gocritic
    - gocyclo

linters-settings:
  gocyclo:
    min-complexity: 15
  gosec:
    excludes:
      - G404 # слабый генератор случайных чисел (разрешён для небезопасных контекстов)
```

## В CI

```yaml
- name: Lint
  uses: golangci/golangci-lint-action@v6
  with:
    version: latest
    args: --timeout=5m
```

<Quiz quizId="infra-03-lint" questions={[
  {id:"q1",question:"Зачем golangci-lint если есть go vet?",options:["golangci-lint быстрее","go vet проверяет ~5 правил. golangci-lint запускает 50+ линтеров включая security (gosec), стиль (gocritic), утечки (bodyclose).","go vet deprecated","golangci-lint заменяет компилятор"],correctIndex:1,explanation:"go vet — только подозрительные конструкции. golangci-lint — это агрегатор: от errcheck (пропущенные ошибки) до gosec (SQL injection, weak crypto)."}
]} />
