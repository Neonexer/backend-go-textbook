import React from "react";
import clsx from "clsx";
import Link from "@docusaurus/Link";
import Layout from "@theme/Layout";
import Heading from "@theme/Heading";
import styles from "./index.module.css";

const projects = [
  {
    title: "Проект 1: REST API для блога",
    description: "Пишем HTTP-сервер с нуля: net/http, chi, middleware, тесты и логирование.",
    link: "/project-1-blog/http-server",
  },
  {
    title: "Проект 2: Маркетплейс",
    description: "PostgreSQL, GORM, pgx, JWT-аутентификация, RBAC, транзакции и интеграционные тесты.",
    link: "/project-2-marketplace/sql-pgx",
  },
  {
    title: "Проект 3: YouTube-клон",
    description: "Микросервисы, gRPC, Kafka, Redis, observability, CI/CD и деплой в Kubernetes.",
    link: "/project-3-youtube-clone/microservices-arch",
  },
];

export default function Home() {
  return (
    <Layout
      title="Бэкенд на Go"
      description="Цифровой учебник по бэкенд-разработке на Go — от первого сервера до YouTube-клона"
    >
      <header className={clsx("hero hero--primary", styles.hero)}>
        <div className="container">
          <Heading as="h1" className="hero__title">
            Бэкенд на Go
          </Heading>
          <p className="hero__subtitle">
            Цифровой учебник для опытных разработчиков, которые хотят освоить Go.
            <br />
            От первого HTTP-сервера до YouTube-клона в микросервисах.
          </p>
          <div className={styles.buttons}>
            <Link
              className="button button--secondary button--lg"
              to="/intro"
            >
              Начать обучение →
            </Link>
            <Link
              className="button button--outline button--lg"
              to="https://github.com/go-course/backend-go-textbook"
            >
              GitHub
            </Link>
          </div>
        </div>
      </header>

      <main className={styles.main}>
        <div className="container">
          <div className="row">
            {projects.map((p) => (
              <div key={p.title} className="col col--4 margin-bottom--lg">
                <div className={styles.card}>
                  <Heading as="h3">{p.title}</Heading>
                  <p>{p.description}</p>
                  <Link to={p.link}>Читать →</Link>
                </div>
              </div>
            ))}
          </div>
        </div>
      </main>
    </Layout>
  );
}
