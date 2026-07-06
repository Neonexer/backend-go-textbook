import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebars: SidebarsConfig = {
  textbook: [
    {
      type: "doc",
      id: "intro",
      label: "Введение",
    },
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
      ],
    },
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
      ],
    },
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
      ],
    },
  ],
};

export default sidebars;
