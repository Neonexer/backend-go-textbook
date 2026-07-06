---
title: "Apache Kafka"
sidebar_position: 3
---

import Quiz from '@site/src/components/Quiz';

# Apache Kafka

Видео загружено — нужно запустить транскодирование. Комментарий оставлен — уведомить автора. Эти операции не требуют мгновенного ответа. Используем Kafka — асинхронный брокер сообщений.

## Зачем вообще брокер

Можно было бы слать HTTP-запрос из Video Service в Transcoder напрямую. Но:

- **Transcoder упал** — Video Service должен запомнить задачу и повторить позже. Kafka хранит сообщения на диске.
- **Пиковая нагрузка** — 100 видео загружены одновременно. Transcoder не справляется. Kafka работает как буфер.
- **Новый consumer** — добавили Notification Service. С Kafka он просто подписывается на тот же топик.

## Основные понятия

| Термин | Значение |
|--------|---------|
| **Topic** | Категория сообщений: `video.uploaded`, `comment.created` |
| **Producer** | Отправляет сообщения в топик |
| **Consumer** | Читает сообщения из топика |
| **Consumer Group** | Группа consumer'ов, которые делят партиции между собой |
| **Partition** | Часть топика, лежит на диске. Сообщения в партиции строго упорядочены. |

## Producer на Go

```go
import "github.com/segmentio/kafka-go"

func produceEvent(topic string, key, value []byte) error {
    writer := &kafka.Writer{
        Addr:     kafka.TCP("localhost:9092"),
        Topic:    topic,
        Balancer: &kafka.Hash{},
    }
    defer writer.Close()

    return writer.WriteMessages(context.Background(), kafka.Message{
        Key:   key,
        Value: value,
    })
}
```

## Consumer на Go

```go
func consumeEvents(topic, groupID string, handler func(kafka.Message) error) {
    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers: []string{"localhost:9092"},
        Topic:   topic,
        GroupID: groupID,
    })
    defer reader.Close()

    for {
        msg, err := reader.ReadMessage(context.Background())
        if err != nil {
            slog.Error("ошибка чтения", "err", err)
            continue
        }
        if err := handler(msg); err != nil {
            slog.Error("ошибка обработки", "err", err)
            // Сообщение не закоммитится и будет обработано снова
            continue
        }
    }
}
```

## Exactly-once семантика

По умолчанию Kafka гарантирует **at-least-once**: сообщение может быть обработано повторно. Для exactly-once:

- **Idempotent producer** — не дублирует сообщения при ретраях
- **Транзакции** — atomic запись в несколько топиков

В большинстве случаев at-least-once достаточно, если обработчик идемпотентен.

## Ключевые выводы

- Kafka — брокер сообщений с хранением на диске
- Topics + Partitions + Consumer Groups = масштабирование
- At-least-once по умолчанию, пиши идемпотентные обработчики
- Для exactly-once: idempotent producer + транзакции

<Quiz quizId="p3-03-kafka" questions={[
  {id:"q1",question:"Почему Kafka, а не прямой HTTP-запрос между сервисами?",options:["HTTP медленнее","Kafka буферизует сообщения, сохраняет их на диск и гарантирует доставку даже если consumer временно недоступен","Kafka бесплатный","HTTP не поддерживает JSON"],correctIndex:1,explanation:"Если Transcoder упал, HTTP-запрос потерян. Kafka хранит сообщения на диске — consumer прочитает их когда восстановится. Это главное отличие очереди от прямого вызова."}
]} />
