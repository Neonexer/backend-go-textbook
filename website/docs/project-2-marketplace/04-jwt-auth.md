---
title: "JWT-аутентификация"
sidebar_position: 4
---

import Quiz from '@site/src/components/Quiz';

# JWT-аутентификация

Делаем настоящую аутентификацию: регистрация, логин, access и refresh токены на JWT.

## Как работает JWT

JWT — три base64-строки: `header.payload.signature`. Сервер создаёт токен при логине, клиент присылает в `Authorization: Bearer <token>`. Сервер проверяет подпись **без обращения к БД** — в этом суть JWT.

## Access + Refresh

| | Access | Refresh |
|---|---|---|
| Время жизни | 15 минут | 7 дней |
| Назначение | Доступ к API | Получить новый access |
| Хранение | Память (фронт) | HttpOnly cookie |

Access короткоживущий — если украдут, ущерб ограничен. Refresh в httpOnly cookie недоступен JavaScript.

## Генерация токена

```go
import "github.com/golang-jwt/jwt/v5"

func generateToken(userID int, role string, secret []byte, ttl time.Duration) (string, error) {
    claims := jwt.MapClaims{
        "user_id": userID,
        "role":    role,
        "exp":     time.Now().Add(ttl).Unix(),
        "iat":     time.Now().Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(secret)
}
```

## Проверка токена в middleware

```go
func authMiddleware(secret []byte) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            header := r.Header.Get("Authorization")
            tokenStr := strings.TrimPrefix(header, "Bearer ")

            token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
                return secret, nil
            })
            if err != nil || !token.Valid {
                writeError(w, http.StatusUnauthorized, "invalid token")
                return
            }

            claims := token.Claims.(jwt.MapClaims)
            ctx := context.WithValue(r.Context(), "user_id", int(claims["user_id"].(float64)))
            ctx = context.WithValue(ctx, "role", claims["role"].(string))
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

## Пароли: bcrypt

```go
import "golang.org/x/crypto/bcrypt"

// При регистрации
hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// При логине
if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
    return "", "", fmt.Errorf("invalid credentials")
}
```

## Ключевые выводы

- Access (15 мин) для API, Refresh (7 дней) для продления
- JWT проверяется без БД — stateless
- bcrypt для паролей, секреты через `JWT_SECRET`
- Никогда не раскрывай причину отказа при логине

<Quiz quizId="p2-04-jwt" questions={[
  {id:"q1",question:"Зачем два токена — access и refresh?",options:["Требование спецификации","Access короткоживущий — кража ограничена по времени, Refresh в httpOnly cookie только для обновления","Один для GET, другой для POST","Два быстрее чем один"],correctIndex:1,explanation:"Access 15 мин — если украли, ущерб ограничен. Refresh 7 дней в httpOnly cookie недоступен JS."},
  {id:"q2",question:"Почему проверка JWT не требует запроса к БД?",options:["JWT самодостаточен: подпись гарантирует подлинность payload","JWT проверяется через внешний сервис","Требует, просто мы экономим","JWT хранит копию БД"],correctIndex:0,explanation:"Payload содержит данные, подпись гарантирует что они не подделаны. Сервер проверяет подпись локально — в этом stateless-природа JWT."}
]} />
