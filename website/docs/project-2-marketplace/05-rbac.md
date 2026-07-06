---
title: "RBAC-авторизация"
sidebar_position: 5
---

import Quiz from '@site/src/components/Quiz';

# RBAC-авторизация

JWT говорит **кто**. RBAC определяет **что можно**. В маркетплейсе: buyer (покупает), seller (продаёт), admin (управляет).

## Middleware по ролям

```go
func requireRole(roles ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            role := r.Context().Value("role").(string)
            for _, allowed := range roles {
                if role == allowed {
                    next.ServeHTTP(w, r)
                    return
                }
            }
            writeError(w, http.StatusForbidden, "insufficient permissions")
        })
    }
}
```

## Применение в роутере

```go
// Публично
r.Get("/products", h.List)

// Авторизованные
r.Group(func(r chi.Router) {
    r.Use(authMiddleware(secret))
    r.Use(requireRole("seller", "admin"))
    r.Post("/products", h.Create)
    r.Put("/products/{id}", h.Update)
    r.Delete("/products/{id}", h.Delete)
})

// Только админы
r.Group(func(r chi.Router) {
    r.Use(requireRole("admin"))
    r.Get("/admin/users", h.ListUsers)
})
```

## Проверка владения

Роли недостаточно — продавец не должен править чужие товары:

```go
if product.SellerID != userID && role != "admin" {
    writeError(w, http.StatusForbidden, "not your product")
    return
}
```

## Ключевые выводы

- Middleware проверяет роль, handler — владение
- Админ — исключение из всех проверок владения
- 403 Forbidden ≠ 401 Unauthorized

<Quiz quizId="p2-05-rbac" questions={[
  {id:"q1",question:"Почему проверку владения нельзя делать только в middleware?",options:["Middleware не имеет доступа к контексту","Middleware не знает данные из БД — seller_id конкретного товара доступен только в handler/service","Можно, это лучшая практика","Middleware только для заголовков"],correctIndex:1,explanation:"Middleware видит HTTP-запрос. Данные о владельце товара — в БД. Middleware проверяет роль, handler — владение."}
]} />
