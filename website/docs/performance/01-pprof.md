---
title: "pprof — профилирование"
sidebar_position: 1
---

import Quiz from '@site/src/components/Quiz';

# pprof — профилирование Go

Код работает, но медленно. pprof показывает куда уходит CPU, где выделяется память и сколько горутин запущено.

## CPU-профиль

```go
import (
    "os"
    "runtime/pprof"
)

func main() {
    f, _ := os.Create("cpu.prof")
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    // ... код для профилирования
}
```

```bash
go tool pprof -http=:8081 cpu.prof
```

Открывается flame graph — интерактивная визуализация где каждый прямоугольник = функция, ширина = время.

## HTTP-эндпоинт (живое профилирование)

```go
import _ "net/http/pprof"

go func() {
    log.Println(http.ListenAndServe(":6060", nil))
}()
```

```bash
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

## Виды профилей

| Профиль | Что показывает |
|---------|---------------|
| `profile` | CPU — где тратится время |
| `heap` | Память — где выделяется |
| `goroutine` | Горутины — где зависли |
| `block` | Блокировки — кто кого ждёт |
| `mutex` | Мьютексы — contention |

## Heap-профиль и утечки

```bash
go tool pprof http://localhost:6060/debug/pprof/heap
```

Показывает: `inuse_space` (текущее использование) и `alloc_space` (всего выделено с момента старта). Если `alloc_space` растёт а `inuse_space` не снижается — утечка.

<Quiz quizId="perf-01-pprof" questions={[
  {id:"q1",question:"Какая разница между inuse_space и alloc_space в heap профиле?",options:["Это одно и то же","inuse_space — память занятая сейчас, alloc_space — вся память выделенная с момента старта. Если inuse растёт бесконечно — утечка.","alloc_space всегда меньше","inuse_space только для стека"],correctIndex:1,explanation:"inuse_space растёт и падает с GC. alloc_space только растёт. Если inuse_space постоянно растёт — GC не может освободить память."}
]} />
