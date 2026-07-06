---
title: "Swagger / OpenAPI"
sidebar_position: 9
---

import Quiz from '@site/src/components/Quiz';

# Swagger / OpenAPI-документация

REST API работает, но клиенты не знают какие эндпоинты существуют и какие параметры принимать. OpenAPI (бывший Swagger) — стандарт описания REST API в машиночитаемом формате. В Go — через `swaggo/swag` с кодогенерацией из аннотаций.

## Установка

```bash
go install github.com/swaggo/swag/cmd/swag@latest
go get github.com/swaggo/http-swagger
```

## Аннотации в коде

```go
// @title           Blog API
// @version         1.0
// @description     REST API для блога на Go
// @host            localhost:8080
// @BasePath        /

// @Summary         Список постов
// @Description     Возвращает все посты
// @Tags            posts
// @Produce         json
// @Success         200  {array}   model.Post
// @Router          /posts [get]
func (h *PostHandler) List(w http.ResponseWriter, r *http.Request) {
    // ...
}

// @Summary         Создать пост
// @Tags            posts
// @Accept          json
// @Produce         json
// @Param           post  body      CreatePostInput  true  "Данные поста"
// @Success         201   {object}  model.Post
// @Failure         400   {object}  model.ErrorResponse
// @Failure         401   {object}  model.ErrorResponse
// @Router          /posts [post]
func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
    // ...
}
```

## Генерация и поднятие

```bash
swag init -g cmd/server/main.go -o docs/
```

```go
import httpSwagger "github.com/swaggo/http-swagger"
import _ "github.com/go-course/project-1-blog/docs"

r.Get("/swagger/*", httpSwagger.WrapHandler)
```

Swagger UI доступен на `http://localhost:8080/swagger/index.html` — интерактивная документация с возможностью отправлять запросы.

## Почему это важно

- Клиенты видят все эндпоинты без чтения кода
- Можно тестировать API прямо из браузера
- Генерация клиентских SDK (TypeScript, Python, Java) из спеки

<Quiz quizId="p1-09-swagger" questions={[
  {id:"q1",question:"Зачем нужен Swagger если есть README?",options:["README не интерактивный — Swagger даёт живую документацию с возможностью отправить запрос","Swagger заменяет тесты","Это требование закона","Swagger автоматически исправляет ошибки"],correctIndex:0,explanation:"Swagger UI позволяет не только читать документацию, но и отправлять запросы к API. Спецификация машиночитаема — можно генерировать клиентские SDK."}
]} />
