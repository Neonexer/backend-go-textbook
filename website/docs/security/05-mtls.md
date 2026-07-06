---
title: "mTLS"
sidebar_position: 5
---

import Quiz from '@site/src/components/Quiz';

# Mutual TLS (mTLS)

TLS защищает клиента: сервер предъявляет сертификат, клиент проверяет. mTLS добавляет обратную проверку: сервер тоже проверяет сертификат клиента. В микросервисах — стандарт безопасности.

## TLS vs mTLS

| | TLS | mTLS |
|---|---|---|
| Серверный сертификат | ✅ | ✅ |
| Клиентский сертификат | ❌ | ✅ |
| Кто проверяет | Клиент | Обе стороны |

## Go-реализация

```go
// Загружаем сертификаты
cert, _ := tls.LoadX509KeyPair("server.crt", "server.key")
caCert, _ := os.ReadFile("ca.crt")
caCertPool := x509.NewCertPool()
caCertPool.AppendCertsFromPEM(caCert)

tlsConfig := &tls.Config{
    Certificates: []tls.Certificate{cert},
    ClientAuth:   tls.RequireAndVerifyClientCert, // ← ключевая строка
    ClientCAs:    caCertPool,
    MinVersion:   tls.VersionTLS13,
}

server := &http.Server{
    Addr:      ":8443",
    TLSConfig: tlsConfig,
}
server.ListenAndServeTLS("", "")
```

## Service Mesh (проще)

Вместо ручной настройки mTLS для каждого сервиса — Istio/Linkerd делают это автоматически через sidecar-прокси.

<Quiz quizId="sec-05-mtls" questions={[
  {id:"q1",question:"Что mTLS даёт по сравнению с обычным TLS?",options:["Ускорение","Сервер проверяет что клиент — это действительно Video Service а не злоумышленник с украденным API ключом. Обе стороны аутентифицированы сертификатами.","mTLS бесплатный","TLS deprecated"],correctIndex:1,explanation:"mTLS гарантирует что обе стороны — те за кого себя выдают. API ключ можно украсть, сертификат без приватного ключа — нет."}
]} />
