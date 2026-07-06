---
title: "Горутины и каналы"
sidebar_position: 2
---

import Quiz from '@site/src/components/Quiz';

# Горутины и каналы

Горутины — лёгкие потоки (2KB стека), не OS-треды. Тысячи горутин — норма. Каналы — типобезопасный способ общения между ними.

## Горутины

```go
go doSomething()    // запускает в новой горутине, main не ждёт
go func() { ... }() // анонимная горутина
```

## Каналы

```go
ch := make(chan int)      // небуферизованный
ch := make(chan int, 10)  // буферизованный (10 элементов)

ch <- 42                  // отправить
v := <-ch                 // получить
close(ch)                 // закрыть (отправитель!)
```

## Паттерны

### Pipeline

```go
func gen(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        for _, n := range nums {
            out <- n
        }
        close(out)
    }()
    return out
}

func sq(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            out <- n * n
        }
        close(out)
    }()
    return out
}

result := sq(gen(2, 3, 4)) // 4, 9, 16
```

### Fan-out / Fan-in

```go
func fanOut(in <-chan int, workers int) []<-chan int {
    outs := make([]<-chan int, workers)
    for i := 0; i < workers; i++ {
        outs[i] = worker(in)
    }
    return outs
}

func fanIn(channels ...<-chan int) <-chan int {
    var wg sync.WaitGroup
    out := make(chan int)
    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for v := range c { out <- v }
        }(ch)
    }
    go func() { wg.Wait(); close(out) }()
    return out
}
```

### Select — множественное ожидание

```go
select {
case msg := <-ch1:
    // обработать ch1
case msg := <-ch2:
    // обработать ch2
case <-ctx.Done():
    // контекст отменён — выходим
    return ctx.Err()
case <-time.After(5 * time.Second):
    // таймаут
}
```

## Типичные ошибки

- **Запись в закрытый канал → panic**
- **Чтение из закрытого канала → zero value (ok)**
- **Забыть закрыть канал → горутина висит вечно в range**
- **Deadlock**: все горутины спят, никто не пишет/читает

<Quiz quizId="gm-02-concurrency" questions={[
  {id:"q1",question:"Кто должен закрывать канал — отправитель или получатель?",options:["Получатель — он знает когда данные кончились","Отправитель — только он знает когда данных больше не будет. Закрытие получателем = panic при следующей отправке.","Оба могут","Каналы не нужно закрывать"],correctIndex:1,explanation:"Только отправитель закрывает канал. Получатель использует range или val, ok := <-ch. Закрытие получателем вызовет панику при следующей отправке."}
]} />
