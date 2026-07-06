# Бэкенд на Go

Цифровой учебник по бэкенд-разработке на Go — от первого HTTP-сервера до YouTube-клона в микросервисах.

## Структура

```
├── website/          # Docusaurus — платформа учебника
│   └── docs/         # Главы (Markdown/MDX)
├── code/             # Исходный код проектов
│   ├── project-1-blog/
│   ├── project-2-marketplace/
│   └── project-3-youtube-clone/
└── docs/             # PRD и документация
```

## Запуск локально

```bash
cd website
npm install
npm start
```

Открой `http://localhost:3000`.

## Три проекта

| # | Проект | Темы |
|---|--------|------|
| 1 | REST API для блога | `net/http`, `chi`, middleware, тесты, логирование |
| 2 | Маркетплейс | PostgreSQL, GORM, pgx, JWT, RBAC, транзакции |
| 3 | YouTube-клон | Микросервисы, gRPC, Kafka, Redis, observability, CI/CD |

## Статус

🟢 Проект 1: в разработке
⚪ Проект 2: запланирован
⚪ Проект 3: запланирован

## Лицензия

MIT
