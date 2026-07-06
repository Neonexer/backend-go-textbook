---
title: "HTTP-сервер на Go"
sidebar_position: 1
---

import Quiz from '@site/src/components/Quiz';

# HTTP-сервер на Go

В этой главе мы напишем первый HTTP-сервер. Без внешних библиотек — только стандартный пакет `net/http`.

## Почему `net/http`

В Go стандартная библиотека покрывает 80% задач бэкенда. В отличие от Python (где нужен Flask/Django) или JavaScript (Express/Fastify), Go даёт готовый продакшен-сервер из коробки. Это философия языка: **меньше зависимостей — меньше проблем**.

## Минимальный сервер

Создай файл `main.go`:

```go
package main

import (
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Привет, бэкенд!")
    })

    fmt.Println("Сервер запущен на http://localhost:8080")
    http.ListenAndServe(":8080", nil)
}
```

Запусти:

```bash
go run main.go
```

Открой `http://localhost:8080` — увидишь «Привет, бэкенд!». Ты только что написал веб-сервер **на трёх строках кода без фреймворков**.

## Как это работает

Разберём по частям.

### `http.HandleFunc`

```go
http.HandleFunc("/", handler)
```

Регистрирует функцию-обработчик для URL-пути `/`. Когда приходит HTTP-запрос на `/`, Go вызывает эту функцию. `HandleFunc` — это синтаксический сахар над стандартным мультиплексором `DefaultServeMux`.

### `http.ResponseWriter` и `*http.Request`

Каждый обработчик получает два аргумента:

- **`w http.ResponseWriter`** — интерфейс для записи ответа. Мы пишем в него через `fmt.Fprintln(w, ...)`.
- **`r *http.Request`** — структура со всей информацией о запросе: метод, URL, заголовки, тело.

### `http.ListenAndServe`

```go
http.ListenAndServe(":8080", nil)
```

Запускает TCP-сервер на порту 8080. Второй аргумент — маршрутизатор (`nil` означает использовать `DefaultServeMux`).

:::tip Порт 8080
Порты ниже 1024 требуют прав суперпользователя. 8080 — стандартный порт для разработки.
:::

## Обработка разных методов

Добавим поддержку GET и POST:

```go
package main

import (
    "fmt"
    "net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        fmt.Fprintln(w, "GET запрос принят")
    case http.MethodPost:
        fmt.Fprintln(w, "POST запрос принят")
    default:
        http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
    }
}

func main() {
    http.HandleFunc("/", handler)
    fmt.Println("Сервер запущен на http://localhost:8080")
    http.ListenAndServe(":8080", nil)
}
```

Проверь POST-запрос через curl:

```bash
curl -X POST http://localhost:8080
```

## Чтение тела запроса

При POST-запросе клиент отправляет данные в теле. Прочитаем JSON:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Только POST", http.StatusMethodNotAllowed)
        return
    }

    // Читаем тело
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Ошибка чтения тела", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    fmt.Fprintf(w, "Получено: %s", body)
}
```

:::warning Всегда закрывай тело запроса
`r.Body.Close()` освобождает ресурсы соединения. Если не закрыть — соединение не вернётся в пул.
:::

## Настройка таймаутов

`ListenAndServe` запускает сервер без таймаутов. В продакшене это опасно: медленный клиент может занять соединение навсегда. Добавим `http.Server` с таймаутами:

```go
server := &http.Server{
    Addr:         ":8080",
    Handler:      nil, // используем DefaultServeMux
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
}

fmt.Println("Сервер запущен на http://localhost:8080")
if err := server.ListenAndServe(); err != nil {
    fmt.Printf("Ошибка сервера: %v\n", err)
}
```

| Параметр | Назначение |
|----------|-----------|
| `ReadTimeout` | Максимальное время чтения запроса (включая тело) |
| `WriteTimeout` | Максимальное время записи ответа |
| `IdleTimeout` | Максимальное время keep-alive соединения без запросов |

## Ключевые выводы

1. Go даёт готовый HTTP-сервер в стандартной библиотеке — `net/http`
2. `HandleFunc` регистрирует обработчик, `ListenAndServe` запускает сервер
3. В продакшене всегда настраивай таймауты через `http.Server`
4. Тело запроса нужно закрывать: `defer r.Body.Close()`

В следующей главе подключим роутер `chi` и научимся работать с URL-параметрами и группами маршрутов.

---

## Проверь себя

<Quiz
  quizId="01-http-server"
  questions={[
    {
      id: "q1",
      question: "Какой пакет стандартной библиотеки Go отвечает за HTTP-сервер?",
      options: ["net", "net/http", "http/server", "fmt"],
      correctIndex: 1,
      explanation: "`net/http` — основной пакет для HTTP в Go. Он содержит http.Server, http.Handler, DefaultServeMux и всё необходимое."
    },
    {
      id: "q2",
      question: "Что произойдёт, если не закрыть `r.Body`?",
      options: [
        "Ничего страшного, Go сам закроет",
        "Соединение не вернётся в пул, возможна утечка ресурсов",
        "Программа упадёт с паникой",
        "Компилятор выдаст ошибку"
      ],
      correctIndex: 1,
      explanation: "Go не закрывает тело запроса автоматически. Если не вызвать Close(), соединение не вернётся в пул для повторного использования — это утечка ресурсов."
    },
    {
      id: "q3",
      question: "Зачем настраивать ReadTimeout на сервере?",
      options: [
        "Чтобы ускорить чтение данных",
        "Чтобы ограничить время на чтение запроса и защититься от медленных клиентов",
        "Чтобы сервер быстрее перезагружался",
        "Это необязательно, можно пропустить"
      ],
      correctIndex: 1,
      explanation: "ReadTimeout защищает сервер от slowloris-атак и медленных клиентов. Без него соединение может висеть открытым неограниченно долго."
    }
  ]}
/>
