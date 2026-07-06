---
title: "WebSockets"
sidebar_position: 11
---

import Quiz from '@site/src/components/Quiz';

# WebSockets

Уведомления о новых комментариях, лайки в реальном времени, прогресс загрузки видео — всё это требует двусторонней связи между клиентом и сервером. WebSockets дают постоянное соединение.

## HTTP vs WebSocket

| | HTTP | WebSocket |
|---|---|---|
| Соединение | На каждый запрос | Постоянное |
| Инициатор | Только клиент | Клиент и сервер |
| Накладные расходы | Заголовки каждый запрос | Минимальные |
| Протокол | HTTP/1.1, HTTP/2 | WebSocket (над HTTP Upgrade) |

## Сервер на Go

```go
import "github.com/gorilla/websocket"

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        slog.Error("upgrade failed", "err", err)
        return
    }
    defer conn.Close()

    for {
        msgType, msg, err := conn.ReadMessage()
        if err != nil {
            break
        }
        // Эхо — в реальном приложении: broadcast другим клиентам
        conn.WriteMessage(msgType, msg)
    }
}
```

## Hub-паттерн

Одно соединение → один пользователь. Hub управляет всеми:

```go
type Hub struct {
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.clients[client] = true
        case client := <-h.unregister:
            delete(h.clients, client)
            close(client.send)
        case msg := <-h.broadcast:
            for client := range h.clients {
                client.send <- msg
            }
        }
    }
}
```

## Когда НЕ использовать WebSockets

- Простые уведомления → Server-Sent Events (SSE) проще
- Запрос-ответ → HTTP/gRPC
- Высокая частота сообщений с гарантией доставки → Kafka + polling

<Quiz quizId="p3-11-websockets" questions={[
  {id:"q1",question:"В чём преимущество WebSocket перед HTTP polling?",options:["WebSocket быстрее HTTP","Постоянное соединение без накладных расходов на заголовки + сервер может инициировать отправку","WebSocket бесплатный","HTTP polling не работает в браузере"],correctIndex:1,explanation:"HTTP polling шлёт заголовки с каждым запросом даже если данных нет. WebSocket держит одно соединение и шлёт данные только когда они появляются."}
]} />
