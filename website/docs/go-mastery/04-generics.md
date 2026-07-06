---
title: "Дженерики в Go"
sidebar_position: 4
---

import Quiz from '@site/src/components/Quiz';

# Дженерики в Go

Go 1.18 добавил дженерики. Не для ООП-наследования, а для обобщённых алгоритмов: функция работает с любым типом, удовлетворяющим constraint.

## Синтаксис

```go
func Max[T constraints.Ordered](a, b T) T {
    if a > b { return a }
    return b
}

x := Max[int](3, 5)       // явно
y := Max(3.14, 2.71)       // вывод типа
```

## Constraints

```go
// Любой тип
func Print[T any](v T) { fmt.Println(v) }

// Сравнимые (==, !=)
func Contains[T comparable](slice []T, item T) bool {
    for _, v := range slice {
        if v == item { return true }
    }
    return false
}

// Числовые
import "golang.org/x/exp/constraints"
func Sum[T constraints.Integer | constraints.Float](nums []T) T {
    var total T
    for _, n := range nums { total += n }
    return total
}
```

## Когда использовать

- ✅ Обобщённые структуры данных: `Stack[T]`, `Set[T]`, `Result[T]`
- ✅ Утилиты: `Ptr[T any](v T) *T` (указатель на значение)
- ✅ Алгоритмы: `Map`, `Filter`, `Reduce` для слайсов
- ❌ Замена интерфейсов — интерфейсы всё ещё предпочтительнее
- ❌ Сложные constraint'ы с десятком типов — признак что что-то не так

## Дженерики и производительность

Дженерики в Go мономорфизируются: для каждого типа генерируется отдельная функция на этапе компиляции. Нет boxing/unboxing как в Java, нет оверхеда в рантайме.

<Quiz quizId="gm-04-generics" questions={[
  {id:"q1",question:"Когда дженерики в Go полезны а когда избыточны?",options:["Всегда","Для обобщённых структур (Set[T], Stack[T]) — да. Для замены интерфейсов — нет. Если constraint содержит >5 типов — сигнал пересмотреть дизайн.","Никогда","Только для числовых типов"],correctIndex:1,explanation:"Дженерики хороши для контейнеров и утилит. Интерфейсы + динамический полиморфизм всё ещё основной механизм абстракции в Go."}
]} />
