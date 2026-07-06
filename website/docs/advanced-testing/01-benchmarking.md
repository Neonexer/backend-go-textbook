---
title: "Бенчмаркинг"
sidebar_position: 1
---

import Quiz from '@site/src/components/Quiz';

# Бенчмаркинг

Правильный ли выбор: pgx или GORM? `sync.Mutex` или `atomic.Int64`? Бенчмарки отвечают цифрами, а не догадками.

## Первый бенчмарк

```go
func BenchmarkJSONMarshal(b *testing.B) {
    p := Post{ID: 1, Title: "Test", Body: "Lorem ipsum"}
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        json.Marshal(p)
    }
}
```

```bash
go test -bench=. -benchmem
# BenchmarkJSONMarshal-12    2000000    650 ns/op    240 B/op    3 allocs/op
```

Go сам подбирает `b.N` чтобы тест длился ~1 секунду. `-benchmem` показывает аллокации.

## Табличные бенчмарки

```go
func BenchmarkValidate(b *testing.B) {
    tests := []struct {
        name string
        post Post
    }{
        {"valid", Post{Title: "Go", Body: "Fast"}},
        {"empty", Post{}},
    }
    for _, tt := range tests {
        b.Run(tt.name, func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                tt.post.Validate()
            }
        })
    }
}
```

## Сравнение подходов

```go
func BenchmarkRepo_Pgx(b *testing.B) {
    pool, _ := pgxpool.New(ctx, dsn)
    repo := NewPgxProduct(pool)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        repo.FindByID(ctx, 1)
    }
}

func BenchmarkRepo_GORM(b *testing.B) {
    db, _ := gorm.Open(postgres.Open(dsn))
    repo := NewGormProduct(db)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        repo.FindByID(1)
    }
}
```

## Читаем результаты

| ns/op | B/op | allocs/op |
|-------|------|-----------|
| 650 ns | 240 B | 3 |
| Время на операцию | Байт на операцию | Аллокаций на операцию |

Чем меньше — тем лучше. 0 allocs/op — идеал (все на стеке).

<Quiz quizId="at-01-bench" questions={[
  {id:"q1",question:"Что означает 0 allocs/op в бенчмарке?",options:["Бенчмарк сломан","Память не выделяется в куче — все данные на стеке. Самый быстрый путь.","Аллокации не измеряются","Код не выполнился"],correctIndex:1,explanation:"0 allocs/op значит escape analysis оставил всё на стеке. Нет работы для GC — максимальная производительность."}
]} />
