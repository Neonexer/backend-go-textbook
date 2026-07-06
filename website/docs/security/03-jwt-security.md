---
title: "JWT Security"
sidebar_position: 3
---

import Quiz from '@site/src/components/Quiz';

# JWT-безопасность

JWT удобен, но легко допустить дыры. Разберём типичные ошибки.

## Алгоритм подписи

```go
// ❌ Никогда не принимай 'none'
token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
    return secret, nil // не проверяет алгоритм
})

// ✅ Явно проверяй алгоритм
token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
    if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
    }
    return secret, nil
})
```

## Время жизни

```go
claims := jwt.MapClaims{
    "exp": time.Now().Add(15 * time.Minute).Unix(), // access: короткий
    "iat": time.Now().Unix(),
    "nbf": time.Now().Add(-30 * time.Second).Unix(), // допуск 30с на разницу часов
}
```

- `exp` — после этого времени токен невалиден
- `iat` — issued at (не раньше)
- `nbf` — not before (допуск на clock skew)

## Ротация ключей

Меняй секрет подписи периодически:

```go
type KeyRotator struct {
    current []byte
    old     []byte
}

func (k *KeyRotator) Validate(tokenStr string) (jwt.MapClaims, error) {
    // Пробуем текущий ключ
    claims, err := parseToken(tokenStr, k.current)
    if err == nil {
        return claims, nil
    }
    // Фоллбек на старый (для токенов выпущенных до ротации)
    return parseToken(tokenStr, k.old)
}
```

<Quiz quizId="sec-03-jwt" questions={[
  {id:"q1",question:"Почему нужно явно проверять алгоритм подписи JWT?",options:["Для скорости","Злоумышленник может поменять alg на 'none' — тогда подпись не проверяется и любой токен считается валидным","Алгоритм не важен","JWT автоматически проверяет"],correctIndex:1,explanation:"Без явной проверки alg злоумышленник может создать токен с alg:'none' и произвольным payload. Парсер JWT должен отвергать такие токены."}
]} />
