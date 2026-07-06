import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebars: SidebarsConfig = {
  textbook: [
    { type: "doc", id: "intro", label: "Введение" },

    // ── Проект 1 ──
    {
      type: "category",
      label: "Проект 1: REST API для блога",
      link: {
        type: "generated-index",
        title: "Проект 1: REST API для блога",
        description: "Основы Go и HTTP: пишем REST API с нуля",
        slug: "/project-1-blog",
      },
      items: [
        "project-1-blog/http-server",
        "project-1-blog/routing-chi",
        "project-1-blog/middleware",
        "project-1-blog/json-serialization",
        "project-1-blog/project-structure",
        "project-1-blog/testing",
        "project-1-blog/logging",
        "project-1-blog/graceful-shutdown",
        "project-1-blog/swagger",
        "project-1-blog/error-handling",
        "project-1-blog/viper-config",
        "project-1-blog/fuzz-testing",
      ],
    },

    // ── Проект 2 ──
    {
      type: "category",
      label: "Проект 2: Маркетплейс",
      link: {
        type: "generated-index",
        title: "Проект 2: Маркетплейс",
        description: "Базы данных и аутентификация на Go",
        slug: "/project-2-marketplace",
      },
      items: [
        "project-2-marketplace/sql-pgx",
        "project-2-marketplace/gorm",
        "project-2-marketplace/migrations",
        "project-2-marketplace/jwt-auth",
        "project-2-marketplace/rbac",
        "project-2-marketplace/transactions",
        "project-2-marketplace/pagination-search",
        "project-2-marketplace/file-upload",
        "project-2-marketplace/integration-tests",
        "project-2-marketplace/goose-migrations",
        "project-2-marketplace/sqlc",
      ],
    },

    // ── Проект 3 ──
    {
      type: "category",
      label: "Проект 3: YouTube-клон",
      link: {
        type: "generated-index",
        title: "Проект 3: YouTube-клон",
        description: "Микросервисы и продакшен на Go",
        slug: "/project-3-youtube-clone",
      },
      items: [
        "project-3-youtube-clone/microservices-arch",
        "project-3-youtube-clone/grpc",
        "project-3-youtube-clone/kafka",
        "project-3-youtube-clone/redis-caching",
        "project-3-youtube-clone/observability",
        "project-3-youtube-clone/rate-limiting",
        "project-3-youtube-clone/ci-cd",
        "project-3-youtube-clone/deployment",
        "project-3-youtube-clone/grpc-streaming",
        "project-3-youtube-clone/circuit-breaker",
        "project-3-youtube-clone/websockets",
      ],
    },

    // ── Системный дизайн ──
    {
      type: "category",
      label: "Системный дизайн",
      link: {
        type: "generated-index",
        title: "Системный дизайн",
        description: "Архитектурные паттерны для масштабируемых систем",
        slug: "/system-design",
      },
      items: [
        "system-design/intro",
        "system-design/load-balancing",
        "system-design/sharding",
        "system-design/cqrs",
        "system-design/saga",
        "system-design/event-sourcing",
      ],
    },

    // ── Безопасность ──
    {
      type: "category",
      label: "Безопасность",
      link: {
        type: "generated-index",
        title: "Безопасность",
        description: "Защита веб-приложений на Go",
        slug: "/security",
      },
      items: [
        "security/cors-csrf",
        "security/owasp-go",
        "security/jwt-security",
      ],
    },

    // ── Docker Compose ──
    {
      type: "category",
      label: "Docker Compose",
      link: {
        type: "generated-index",
        title: "Docker Compose",
        description: "Локальное окружение для всех проектов",
        slug: "/docker-compose",
      },
      items: ["docker-compose/local-env"],
    },

    // ── Go-мастерство ──
    {
      type: "category",
      label: "Go-мастерство",
      link: {
        type: "generated-index",
        title: "Go-мастерство",
        description: "Продвинутые темы языка: контексты, конкурентность, синхронизация",
        slug: "/go-mastery",
      },
      items: [
        "go-mastery/context",
        "go-mastery/concurrency",
        "go-mastery/sync",
      ],
    },

    // ── Продакшен-паттерны ──
    {
      type: "category",
      label: "Продакшен-паттерны",
      link: {
        type: "generated-index",
        title: "Продакшен-паттерны",
        description: "Паттерны для надёжных систем: идемпотентность, feature flags, версионирование API",
        slug: "/production-patterns",
      },
      items: [
        "production-patterns/idempotency",
        "production-patterns/feature-flags",
        "production-patterns/api-versioning",
      ],
    },

    // ── Инфраструктура ──
    {
      type: "category",
      label: "Инфраструктура",
      link: {
        type: "generated-index",
        title: "Инфраструктура",
        description: "Terraform, мониторинг, SLO и линтинг",
        slug: "/infrastructure",
      },
      items: [
        "infrastructure/terraform",
        "infrastructure/monitoring-slo",
        "infrastructure/golangci-lint",
      ],
    },
  ],
};

export default sidebars;
