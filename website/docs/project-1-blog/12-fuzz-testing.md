---
title: "Fuzz-тестирование"
sidebar_position: 12
---

import Quiz from '@site/src/components/Quiz';

# Fuzz-тестирование

Обычные тесты проверяют конкретные входные данные. Fuzz-тесты генерируют случайные входы и ищут паники, бесконечные циклы и asserts. Встроено в Go с 1.18.

## Как это работает

Fuzzer мутирует входные данные и смотрит чтобы тест не упал:

```go
func FuzzValidatePost(f *testing.F) {
    // Начальные значения (seed corpus)
    f.Add("Valid Title", "Valid body content")
    f.Add("", "") // пустой ввод

    f.Fuzz(func(t *testing.T, title string, body string) {
        p := model.Post{Title: title, Body: body}
        err := p.Validate()

        // Validate не должна паниковать
        // Если title валидный — body тоже должен быть валидным?
        // Это уже зависит от бизнес-логики
        _ = err
    })
}
```

Запуск:

```bash
go test -fuzz=FuzzValidatePost -fuzztime=30s ./internal/model/
```

Go будет генерировать случайные строки, числа, слайсы и проверять что тест не паникует.

## Что фаззить

- **Парсеры**: JSON, URL-query, multipart forms — хорошие кандидаты
- **Валидация**: на странных Unicode-строках может упасть
- **Бинарные форматы**: protobuf, custom binary protocols
- **Конвертеры**: string → int, любые трансформации данных

```go
func FuzzParsePagination(f *testing.F) {
    f.Add("1", "20")    // нормальные значения
    f.Add("-1", "0")    // отрицательные
    f.Add("abc", "xyz") // не числа

    f.Fuzz(func(t *testing.T, pageStr string, perPageStr string) {
        page, _ := strconv.Atoi(pageStr)
        perPage, _ := strconv.Atoi(perPageStr)
        // Проверяем что функция не паникует при любом вводе
        _ = parsePagination(page, perPage)
    })
}
```

## Fuzzing в CI

Fuzz-тесты запускаются ограниченное время при каждом PR:

```yaml
- name: Fuzz (30 seconds)
  run: go test -fuzz=. -fuzztime=30s ./...
```

А полноценный длительный fuzzing — отдельной задачей раз в сутки.

<Quiz quizId="p1-12-fuzz" questions={[
  {id:"q1",question:"Чем fuzz-тестирование отличается от юнит-тестирования?",options:["Ничем","Юнит-тесты проверяют конкретные входы, fuzz-тесты генерируют случайные входные данные и ищут крайние случаи которые разработчик не предусмотрел","Fuzz-тесты быстрее","Fuzz-тесты не требуют написания кода"],correctIndex:1,explanation:"Разработчик пишет тесты для ожидаемых сценариев. Fuzzer генерирует миллионы неожиданных входов и находит баги, которые человек не предвидел."}
]} />
