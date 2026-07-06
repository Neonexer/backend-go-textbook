---
title: "gRPC Streaming"
sidebar_position: 9
---

import Quiz from '@site/src/components/Quiz';

# gRPC Streaming

Глава 2 показала unary gRPC: запрос → ответ. Но gRPC поддерживает ещё три режима стриминга через ключевое слово `stream` в proto-файле.

## Четыре режима gRPC

| Режим | Клиент | Сервер | Пример |
|-------|--------|--------|--------|
| **Unary** | 1 запрос | 1 ответ | GetVideo |
| **Server streaming** | 1 запрос | поток ответов | Стриминг видео по чанкам |
| **Client streaming** | поток запросов | 1 ответ | Загрузка большого файла чанками |
| **Bidirectional** | поток | поток | Чат, коллаборативное редактирование |

## Server-side streaming

Видео не влезает в один gRPC-ответ. Стримим чанками:

```protobuf
service VideoService {
    rpc StreamVideo(StreamVideoRequest) returns (stream VideoChunk);
}

message StreamVideoRequest {
    string video_id = 1;
    int32 quality = 2;  // 360, 720, 1080
}

message VideoChunk {
    bytes data = 1;
    int32 chunk_index = 2;
}
```

Сервер:

```go
func (s *VideoServer) StreamVideo(req *video.StreamVideoRequest, stream video.VideoService_StreamVideoServer) error {
    chunks, err := s.svc.GetVideoChunks(req.VideoId, req.Quality)
    if err != nil {
        return status.Errorf(codes.NotFound, "video not found")
    }

    for _, chunk := range chunks {
        if err := stream.Send(&video.VideoChunk{
            Data: chunk.Data, ChunkIndex: chunk.Index,
        }); err != nil {
            return err
        }
    }
    return nil
}
```

Клиент читает поток:

```go
stream, err := client.StreamVideo(ctx, &video.StreamVideoRequest{VideoId: "abc"})
for {
    chunk, err := stream.Recv()
    if err == io.EOF { break }
    // обрабатываем чанк
}
```

## Bidirectional streaming — чат

```protobuf
service ChatService {
    rpc Chat(stream ChatMessage) returns (stream ChatMessage);
}
```

Обе стороны одновременно шлют и принимают сообщения:

```go
func (s *ChatServer) Chat(stream video.ChatService_ChatServer) error {
    for {
        msg, err := stream.Recv()
        if err == io.EOF { return nil }
        if err != nil { return err }

        // Обрабатываем и отвечаем
        stream.Send(&video.ChatMessage{
            UserId: msg.UserId, Text: "echo: " + msg.Text,
        })
    }
}
```

<Quiz quizId="p3-09-streaming" questions={[
  {id:"q1",question:"Когда использовать server-side streaming вместо unary?",options:["Когда ответ не влезает в одно сообщение — большие файлы, потоки данных","Всегда","Только для видео","Unary deprecated"],correctIndex:0,explanation:"Server streaming — для больших ответов которые не должны ждать полной загрузки. Клиент начинает получать данные сразу, не дожидаясь всего ответа."}
]} />
