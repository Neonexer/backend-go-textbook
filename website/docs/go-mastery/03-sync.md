---
title: "Примитивы синхронизации"
sidebar_position: 3
---

import Quiz from '@site/src/components/Quiz';

# Примитивы синхронизации

Горутины разделяют память. Когда две горутины пишут в одну переменную — гонка данных. `sync` пакет защищает от этого.

## sync.Mutex

```go
type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}
```

`defer Unlock` — обязательно. Если между Lock и Unlock паника, без defer мьютекс останется залочен навсегда.

:::tip RWMutex для read-heavy
`sync.RWMutex` даёт множественное чтение но эксклюзивную запись. В 100× больше чтений чем записей — RWLock быстрее обычного Mutex.
:::

## sync.WaitGroup — ждать группу горутин

```go
var wg sync.WaitGroup

for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        process(id)
    }(i)
}

wg.Wait() // блокируется пока все не вызовут Done()
```

## sync.Once — выполнить ровно один раз

```go
var (
    once   sync.Once
    client *http.Client
)

func GetClient() *http.Client {
    once.Do(func() {
        client = &http.Client{Timeout: 10 * time.Second}
    })
    return client
}
```

Идеально для ленивой инициализации. Потокобезопасно.

## errgroup — горутины с ошибками

```go
import "golang.org/x/sync/errgroup"

g, ctx := errgroup.WithContext(ctx)

g.Go(func() error { return fetchUser(ctx, 1) })
g.Go(func() error { return fetchUser(ctx, 2) })
g.Go(func() error { return fetchUser(ctx, 3) })

if err := g.Wait(); err != nil {
    // первая ошибка отменяет контекст для остальных
}
```

## atomic — без мьютекса

Для простых счётчиков:

```go
var requests atomic.Int64

requests.Add(1)         // инкремент
count := requests.Load() // чтение
```

Быстрее Mutex, но только для простых операций.

<Quiz quizId="gm-03-sync" questions={[
  {id:"q1",question:"Когда atomic быстрее Mutex?",options:["Всегда","Для простых операций (инкремент счётчика) — атомарные операции на уровне CPU быстрее чем блокировка. Но только для отдельных полей.","Никогда","Только для строк"],correctIndex:1,explanation:"atomic использует CPU-инструкции (LOCK XADD на x86) вместо блокировок. Для счётчика запросов — идеально. Для сложной структуры — только Mutex."}
]} />
