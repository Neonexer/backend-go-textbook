---
title: "OWASP Top 10 для Go"
sidebar_position: 2
---

import Quiz from '@site/src/components/Quiz';

# OWASP Top 10 для Go

OWASP Top 10 — список самых критичных уязвимостей веб-приложений. Разберём как они выглядят в Go и как защититься.

## 1. Broken Access Control

Проверка что пользователь имеет право на действие:

```go
// ❌ Уязвимо — не проверяется владение
func updatePost(w http.ResponseWriter, r *http.Request) {
    postID, _ := strconv.Atoi(chi.URLParam(r, "id"))
    // любой авторизованный может редактировать любой пост
}

// ✅ Проверка владения
func updatePost(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(int)
    post := svc.Get(postID)
    if post.AuthorID != userID && role != "admin" {
        writeError(w, http.StatusForbidden, "not your post")
        return
    }
}
```

## 2. Cryptographic Failures

Пароли — bcrypt, не SHA256. Токены — HMAC-SHA256. Ключи — не в коде:

```go
// ❌ Уязвимо
hash := sha256.Sum256([]byte(password))
secret := "my-secret-key"

// ✅ Правильно
hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
secret := os.Getenv("JWT_SECRET")
```

## 3. Injection

Параметризованные запросы, а не конкатенация строк:

```go
// ❌ SQL Injection — НИКОГДА
db.Query("SELECT * FROM users WHERE email = '" + email + "'")

// ✅ Параметризованный запрос
db.Query("SELECT * FROM users WHERE email = $1", email)
```

## 4. Insecure Design

Ограничение частоты запросов, валидация на сервере (не на клиенте):

```go
// Rate limiting на уровне middleware (см. главу 6 Проекта 3)
// Валидация на сервере ВСЕГДА
```

## 5. Security Misconfiguration

Выключай дебаг в продакшене, ставь HTTPS, убирай лишние заголовки:

```go
// ❌ Информация о сервере
w.Header().Set("X-Powered-By", "Go/1.22")

// ✅ Убираем или маскируем
```

## 6. Vulnerable Components

`go list -m -u all` проверяет устаревшие зависимости. Dependabot обновляет автоматически.

<Quiz quizId="sec-02-owasp" questions={[
  {id:"q1",question:"Как защититься от SQL injection в Go?",options:["Экранировать кавычки","Использовать плейсхолдеры ($1, $2) — значения передаются отдельно от запроса","Писать SQL заглавными буквами","Проверять длину строки"],correctIndex:1,explanation:"Параметризованные запросы отделяют SQL-код от данных. Даже если значение содержит SQL-команды, они будут восприняты как литерал."}
]} />
