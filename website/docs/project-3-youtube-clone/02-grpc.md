---
title: "gRPC"
sidebar_position: 2
---

import Quiz from '@site/src/components/Quiz';

# gRPC — межсервисное общение

Микросервисы общаются друг с другом. HTTP+JSON — медленно и многословно. gRPC — бинарный протокол от Google на HTTP/2 с кодогенерацией из `.proto`-файлов.

## Почему gRPC, а не REST

| | REST/JSON | gRPC/Protobuf |
|---|---|---|
| Формат | Текстовый JSON | Бинарный Protobuf |
| Скорость | Медленнее | В 3-10× быстрее |
| Контракт | OpenAPI (опционально) | `.proto` (обязательно) |
| Потоки | Нет | Streaming (server, client, bidirectional) |
| Кодогенерация | Ручная | Автоматическая |

## Protobuf-спецификация

```protobuf
// api/video.proto
syntax = "proto3";
package video;
option go_package = "github.com/go-course/youtube-clone/api/video";

service VideoService {
    rpc GetVideo(GetVideoRequest) returns (Video);
    rpc ListVideos(ListVideosRequest) returns (ListVideosResponse);
    rpc UploadVideo(stream UploadVideoRequest) returns (UploadVideoResponse);
}

message Video {
    string id = 1;
    string title = 2;
    string description = 3;
    string url = 4;
    int64 views = 5;
}

message GetVideoRequest {
    string id = 1;
}

message ListVideosRequest {
    int32 page_size = 1;
    string page_token = 2;
}

message ListVideosResponse {
    repeated Video videos = 1;
    string next_page_token = 2;
}
```

## Генерация Go-кода

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

protoc --go_out=. --go-grpc_out=. api/video.proto
```

## gRPC-сервер на Go

```go
type VideoServer struct {
    video.UnimplementedVideoServiceServer
    svc *service.VideoService
}

func (s *VideoServer) GetVideo(ctx context.Context, req *video.GetVideoRequest) (*video.Video, error) {
    v, err := s.svc.Get(ctx, req.Id)
    if err != nil {
        return nil, status.Errorf(codes.NotFound, "video not found")
    }
    return &video.Video{
        Id: v.ID, Title: v.Title, Url: v.URL, Views: v.Views,
    }, nil
}

func main() {
    lis, _ := net.Listen("tcp", ":50051")
    grpcServer := grpc.NewServer()
    video.RegisterVideoServiceServer(grpcServer, &VideoServer{})
    grpcServer.Serve(lis)
}
```

## gRPC-клиент

```go
conn, _ := grpc.Dial("video-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
client := video.NewVideoServiceClient(conn)

resp, err := client.GetVideo(ctx, &video.GetVideoRequest{Id: "abc"})
```

## Ключевые выводы

- `.proto` — единый источник правды для контракта
- Бинарный Protobuf в 3-10× быстрее JSON
- Кодогенерация даёт типобезопасный клиент и сервер
- gRPC использует HTTP/2 — мультиплексирование, header compression

<Quiz quizId="p3-02-grpc" questions={[
  {id:"q1",question:"Главное преимущество gRPC перед REST/JSON?",options:["Проще в отладке","Бинарный Protobuf быстрее, есть кодогенерация, поддержка стриминга","gRPC работает без интернета","Бесплатный"],correctIndex:1,explanation:"Бинарный формат быстрее парсится, меньше размер. Кодогенерация даёт типобезопасность. Стриминг недоступен в REST."}
]} />
