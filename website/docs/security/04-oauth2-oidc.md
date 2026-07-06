---
title: "OAuth2 / OIDC"
sidebar_position: 4
---

import Quiz from '@site/src/components/Quiz';

# OAuth2 и OpenID Connect

JWT — хорошо. Но откуда брать пользователей? Регистрировать самому? OAuth2 позволяет пользователям входить через Google, GitHub, Яндекс. OpenID Connect (OIDC) добавляет identity поверх OAuth2.

## OAuth2 Flow (Authorization Code + PKCE)

```
User → "Войти через Google"
     → Редирект на accounts.google.com
     → Пользователь разрешает доступ
     → Google редиректит обратно с ?code=ABC
     → Сервер обменивает code на access_token (server-to-server)
     → Сервер получает userinfo с access_token
```

## Почему PKCE

PKCE (Proof Key for Code Exchange) — защита от перехвата authorization code. Даже если code украден, без `code_verifier` его не обменять на токен.

## Реализация на Go

```go
import "golang.org/x/oauth2"

var googleOauth = &oauth2.Config{
    ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
    ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
    RedirectURL:  "https://api.example.com/auth/google/callback",
    Scopes:       []string{"email", "profile"},
    Endpoint:     google.Endpoint,
}

// Шаг 1: редирект на Google
func loginHandler(w http.ResponseWriter, r *http.Request) {
    url := googleOauth.AuthCodeURL(state, oauth2.AccessTypeOffline)
    http.Redirect(w, r, url, http.StatusFound)
}

// Шаг 2: callback
func callbackHandler(w http.ResponseWriter, r *http.Request) {
    code := r.URL.Query().Get("code")
    token, _ := googleOauth.Exchange(ctx, code)
    userInfo, _ := fetchUserInfo(token)
    // Создать/найти пользователя, выпустить свой JWT
}
```

<Quiz quizId="sec-04-oauth" questions={[
  {id:"q1",question:"Зачем PKCE в OAuth2 flow?",options:["Для скорости","Code передаётся через браузер (фронт-канал) и может быть перехвачен. PKCE добавляет code_verifier который знает только сервер — без него code бесполезен.","PKCE обязателен для мобильных приложений","Это просто рекомендация"],correctIndex:1,explanation:"Authorization code идёт через URL редиректа. PKCE гарантирует что даже перехваченный code нельзя обменять на токен без code_verifier."}
]} />
