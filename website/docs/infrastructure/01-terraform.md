---
title: "Terraform"
sidebar_position: 1
---

import Quiz from '@site/src/components/Quiz';

# Terraform — инфраструктура как код

Ручной запуск инстансов в облаке не масштабируется. Terraform описывает инфраструктуру в `.tf` файлах. `terraform apply` — и всё создано, одинаково на staging и production.

## Основные концепты

- **Provider** — облако: AWS, GCP, Yandex Cloud, Kubernetes
- **Resource** — конкретный объект: VM, БД, бакет
- **State** — что уже создано (хранится в `.tfstate`, S3 или Terraform Cloud)

## Пример для нашего проекта

```hcl
# main.tf
terraform {
  required_providers {
    kubernetes = { source = "hashicorp/kubernetes" }
  }
}

resource "kubernetes_deployment" "video_service" {
  metadata {
    name = "video-service"
    labels = { app = "video-service" }
  }
  spec {
    replicas = 3
    selector { match_labels = { app = "video-service" } }
    template {
      metadata { labels = { app = "video-service" } }
      spec {
        container {
          name  = "video-service"
          image = "ghcr.io/go-course/video-service:${var.image_tag}"
          port  { container_port = 8080 }
          env {
            name  = "DATABASE_URL"
            value = var.database_url
          }
        }
      }
    }
  }
}

variable "image_tag" { default = "latest" }
variable "database_url" { sensitive = true }
```

## Plan → Apply

```bash
terraform plan    # что будет создано/изменено/удалено? Превью.
terraform apply   # применить
terraform destroy # удалить всё
```

<Quiz quizId="infra-01-terraform" questions={[
  {id:"q1",question:"Зачем Terraform если есть kubectl apply?",options:["Terraform быстрее","Terraform управляет всей инфраструктурой (K8s, БД, DNS, S3), а не только K8s. kubectl только для кластера.","kubectl не работает без Terraform","Terraform бесплатный"],correctIndex:1,explanation:"Terraform — единый язык для всей инфраструктуры. K8s поды + RDS база + S3 бакет + DNS запись = один terraform apply. Без него — 4 разных инструмента."}
]} />
