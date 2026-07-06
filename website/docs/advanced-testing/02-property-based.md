---
title: "Property-based тестирование"
sidebar_position: 2
---

import Quiz from '@site/src/components/Quiz';

# Property-based тестирование

Fuzz-тесты ищут паники случайными данными. Property-based тесты проверяют **свойства**: «после Marshal → Unmarshal получаем исходный объект» — для любых входных данных.

## rapid (стандартная библиотека)

```go
import "testing/rapid"

func TestMarshalRoundtrip(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // Генерируем случайный Post
        title := rapid.StringMatching("[a-zA-Z ]{1,200}").Draw(t, "title")
        body := rapid.String().Draw(t, "body")
        original := Post{Title: title, Body: body}

        data, err := json.Marshal(original)
        if err != nil {
            t.Fatal(err)
        }

        var decoded Post
        json.Unmarshal(data, &decoded)

        if original != decoded {
            t.Fatalf("roundtrip: %+v != %+v", original, decoded)
        }
    })
}
```

## Свойства для проверки

- **Roundtrip**: `decode(encode(x)) == x`
- **Идемпотентность**: `f(f(x)) == f(x)`
- **Монотонность**: `a < b → f(a) <= f(b)`
- **Коммутативность**: `f(a, b) == f(b, a)`

## Когда property-based лучше табличных тестов

Табличные тесты проверяют конкретные примеры. Property-based проверяет инварианты для любых входов. Одно свойство покрывает миллионы случаев.

<Quiz quizId="at-02-property" questions={[
  {id:"q1",question:"Чем property-based тест отличается от fuzz-теста?",options:["Ничем","Fuzz ищет паники. Property-based проверяет логические свойства (roundtrip, идемпотентность) для случайных входов.","Property-based тесты быстрее","Это одно и то же в разных языках"],correctIndex:1,explanation:"Fuzz: 'не упади'. Property-based: 'для любых входов выполняется свойство X'. Разные цели, дополняют друг друга."}
]} />
