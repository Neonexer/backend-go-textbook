---
title: "API Versioning"
sidebar_position: 3
---

import Quiz from '@site/src/components/Quiz';

# API Versioning

API эволюционирует. Поле `price` меняется с копеек на рубли. Клиенты не могут обновиться мгновенно. Нужна версионность.

## Три стратегии

### URL versioning (самая популярная)

```
GET /api/v1/posts
GET /api/v2/posts
```

Явно, легко тестировать, кешируется. Не Restful-пуризм но работает надёжно.

### Header versioning

```
GET /posts
Accept: application/vnd.blog.v2+json
```

URL чистый. Но не кешируется CDN и сложнее отлаживать.

### Query parameter

```
GET /posts?version=2
```

Просто, но засоряет query params.

## Реализация в chi

```go
r.Route("/api", func(r chi.Router) {
    r.Route("/v1", func(r chi.Router) {
        r.Get("/posts", hV1.List)
        r.Post("/posts", hV1.Create)
    })
    r.Route("/v2", func(r chi.Router) {
        r.Get("/posts", hV2.List)
        r.Post("/posts", hV2.Create)
    })
})
```

Сервисный слой — общий. Handler конвертирует между версиями:

```go
func (h *V2Handler) Create(w http.ResponseWriter, r *http.Request) {
    var req CreatePostV2Request // цена в рублях
    json.NewDecoder(r.Body).Decode(&req)

    // v2 → сервис (конвертация рублей в копейки)
    post, _ := h.svc.Create(req.Title, req.Body, req.Price*100)

    // сервис → v2 ответ
    writeJSON(w, 201, toV2Response(post))
}
```

## Deprecation

```go
w.Header().Set("Deprecation", "true")
w.Header().Set("Sunset", "Sat, 01 Jan 2027 00:00:00 GMT")
w.Header().Set("Link", "</api/v2/posts>; rel=successor-version")
```

Клиенты видят что версия устарела и когда отключится.

<Quiz quizId="pp-03-versioning" questions={[
  {id:"q1",question:"Почему URL versioning (v1, v2) самый популярный несмотря на критику?",options:["Это единственный способ","Явный, кешируется CDN, видно в логах, легко тестировать. Теоретически не RESTful, на практике удобнее всего.","Header versioning deprecated","Query parameter не работает"],correctIndex:1,explanation:"URL versioning прост и надёжен. CDN кеширует v1 и v2 раздельно. Логи сразу показывают кто какую версию использует. REST-пуристы недовольны но индустрия выбрала этот путь."}
]} />
