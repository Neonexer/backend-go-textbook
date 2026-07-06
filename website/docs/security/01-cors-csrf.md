---
title: "CORS и CSRF"
sidebar_position: 1
---

import Quiz from '@site/src/components/Quiz';

# CORS и CSRF

Когда фронтенд с одного домена (`app.example.com`) вызывает API на другом (`api.example.com`) — браузер блокирует запрос. Это Same-Origin Policy. CORS ослабляет это правило, CSRF-токены защищают от атак.

## CORS (Cross-Origin Resource Sharing)

Браузер перед cross-origin запросом шлёт preflight (`OPTIONS`):

```go
import "github.com/go-chi/cors"

r.Use(cors.Handler(cors.Options{
    AllowedOrigins:   []string{"https://app.example.com"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders:   []string{"Authorization", "Content-Type"},
    ExposedHeaders:   []string{"X-RateLimit-Limit"},
    AllowCredentials: true,
    MaxAge:           300, // кешировать preflight на 5 минут
}))
```

:::warning Никогда не используй `*` с credentials
`AllowedOrigins: ["*"]` не работает с `AllowCredentials: true`. Указывай конкретные домены.
:::

## CSRF (Cross-Site Request Forgery)

Злоумышленник заставляет пользователя выполнить нежелательное действие: `<img src="https://bank.com/transfer?to=hacker&amount=1000">`. Если пользователь залогинен в банке — перевод выполнится.

**Защита**: CSRF-токен — случайная строка, которую сервер генерирует и проверяет:

```go
import "github.com/gorilla/csrf"

csrfMiddleware := csrf.Protect(
    []byte("32-byte-secret-key"),
    csrf.Secure(false), // true в production (HTTPS only)
)

r.Use(csrfMiddleware)
```

Токен передаётся в заголовке `X-CSRF-Token` и проверяется для state-changing запросов (POST, PUT, DELETE).

<Quiz quizId="sec-01-cors" questions={[
  {id:"q1",question:"Почему Same-Origin Policy блокирует запросы между app.example.com и api.example.com?",options:["Это баг","Это защита: без CORS злоумышленный сайт мог бы вызывать API от имени пользователя. CORS говорит браузеру что кросс-доменные запросы разрешены.","Разные домены всегда запрещены","Это ограничение HTTP"],correctIndex:1,explanation:"Same-Origin Policy — фундаментальная защита браузера. CORS — способ сервера явно разрешить конкретным доменам доступ."}
]} />
