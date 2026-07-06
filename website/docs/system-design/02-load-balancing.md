---
title: "Load Balancing"
sidebar_position: 2
---

import Quiz from '@site/src/components/Quiz';

# Load Balancing

Один сервер обрабатывает 1000 запросов/сек. Два — 2000. Но кто решает какой запрос на какой сервер отправить? Load Balancer.

## Алгоритмы балансировки

| Алгоритм | Как работает | Когда |
|----------|-------------|-------|
| **Round Robin** | По очереди: 1→2→3→1→... | Простой default |
| **Least Connections** | На сервер с наименьшим числом активных соединений | Долгие запросы |
| **IP Hash** | Хеш от IP клиента → всегда на один сервер | Sticky sessions |
| **Weighted** | Сервер с весом 2 получает вдвое больше трафика | Разные мощности серверов |

## Layer 4 vs Layer 7

- **L4 (TCP)** — балансировка на уровне IP:порта. Быстро, не видит HTTP. (HAProxy, AWS NLB)
- **L7 (HTTP)** — видит URL, заголовки, cookies. Может направить `/video/*` на один кластер, `/api/*` на другой. (Nginx, Envoy, AWS ALB)

## Health Checks

Балансировщик должен знать какие серверы живы:

```nginx
upstream backend {
    server 10.0.0.1:8080;
    server 10.0.0.2:8080;
    server 10.0.0.3:8080;

    health_check uri=/health interval=10s;
}
```

## Sticky Sessions

Проблема: пользователь залогинился на сервере 1, следующий запрос round-robin отправил на сервер 2 где сессии нет. Решения:

- **Sticky sessions** (cookie/ip hash) — привязка к серверу. Просто, но нарушает балансировку
- **Shared session store** (Redis) — сессия доступна всем серверам
- **Stateless JWT** — нет серверной сессии вообще

<Quiz quizId="sd-02-lb" questions={[
  {id:"q1",question:"Почему JWT лучше sticky sessions для масштабирования?",options:["JWT быстрее","JWT не привязывает пользователя к серверу — любой сервер может обработать запрос. Sticky sessions нарушают балансировку.","Sticky sessions deprecated","JWT бесплатный"],correctIndex:1,explanation:"JWT stateless: любой сервер проверяет подпись локально. Sticky sessions создают неравномерную нагрузку и проблемы при падении сервера."}
]} />
