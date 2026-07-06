---
title: "Миграции БД"
sidebar_position: 3
---

import Quiz from '@site/src/components/Quiz';

# Миграции базы данных

Миграции — это версионирование схемы БД. Каждое изменение схемы — SQL-файл с номером, который применяется ровно один раз. Используем `golang-migrate`.

## Зачем нужны миграции

Без миграций схема БД живёт только в голове разработчика. Миграции дают:
- Воспроизводимость: `migrate up` → схема готова
- Откат: `migrate down` → возврат к предыдущей версии
- Аудит: git log показывает кто и когда изменил схему

```bash
migrate create -ext sql -dir migrations -seq add_products_table
# → 000001_add_products_table.up.sql
# → 000001_add_products_table.down.sql
```

## Запуск из Go

```go
import (
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func runMigrations(dbURL string) error {
    m, err := migrate.New("file://migrations", dbURL)
    if err != nil {
        return err
    }
    defer m.Close()

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }
    return nil
}
```

`m.Up()` при старте приложения — схема всегда актуальна.

## Правила

1. Одна миграция — одно изменение
2. Всегда пиши down (или пустой с комментарием)
3. Идемпотентность: `IF NOT EXISTS`, `IF EXISTS`
4. Не меняй применённые миграции — создавай новую

## Ключевые выводы

- Миграции — версионирование схемы, как git для кода
- `m.Up()` при старте: fail fast если БД недоступна
- Down для отката, идемпотентный SQL

<Quiz quizId="p2-03-migrations" questions={[
  {id:"q1",question:"Почему нельзя редактировать уже применённую миграцию?",options:["Файл заблокирован","Схемы окружений разъедутся — стейджинг уже выполнил старую версию","golang-migrate запрещает","Это просто правило хорошего тона"],correctIndex:1,explanation:"Изменённая миграция не будет повторно применена на окружениях где она уже прошла. Создавай новую."},
  {id:"q2",question:"Зачем запускать миграции при старте приложения?",options:["Это единственный способ","Fail fast: если БД недоступна, приложение не стартует и проблема видна сразу","Чтобы не создавать отдельный CI шаг","Миграции при старте быстрее"],correctIndex:1,explanation:"При старте миграции гарантируют актуальную схему. Нет схемы — нет приложения."}
]} />
