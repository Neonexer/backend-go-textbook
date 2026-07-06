---
title: "Деплой"
sidebar_position: 8
---

import Quiz from '@site/src/components/Quiz';

# Деплой в Kubernetes

Код написан, протестирован, собран в Docker-образ. Осталось запустить в продакшене. Используем Kubernetes — платформу для оркестрации контейнеров.

## Почему K8s

Docker Compose хорош для локальной разработки. Для продакшена K8s даёт:

- **Self-healing**: упавший контейнер перезапускается автоматически
- **Масштабирование**: `kubectl scale deployment/video-service --replicas=5`
- **Rolling updates**: обновление без простоя
- **Service discovery**: не нужно хардкодить адреса

## Основные объекты

| Объект | Назначение |
|--------|-----------|
| **Pod** | Один или несколько контейнеров, минимальная единица |
| **Deployment** | Управляет подами: сколько, какой образ, стратегия обновления |
| **Service** | Стабильный IP/DNS для подов |
| **ConfigMap/Secret** | Конфигурация и секреты |
| **Ingress** | Внешний доступ к сервисам |

## Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: video-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: video-service
  template:
    metadata:
      labels:
        app: video-service
    spec:
      containers:
        - name: video-service
          image: ghcr.io/go-course/video-service:v1.0.0
          ports:
            - containerPort: 8080
          env:
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: video-db
                  key: url
          readinessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 5
```

## Readiness и Liveness пробы

- **Readiness** — готов ли под принимать трафик. Если нет — исключается из Service.
- **Liveness** — жив ли под. Если нет — перезапускается.

```yaml
readinessProbe:
  httpGet: { path: /health, port: 8080 }
livenessProbe:
  httpGet: { path: /health, port: 8080 }
  initialDelaySeconds: 15
  periodSeconds: 20
```

## Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: video-service
spec:
  selector:
    app: video-service
  ports:
    - port: 8080
      targetPort: 8080
```

Теперь другие сервисы обращаются по DNS-имени `video-service:8080`.

## Helm

Для управления несколькими сервисами — Helm (менеджер пакетов для K8s):

```
video-app/
├── Chart.yaml
├── values.yaml          # конфигурация
└── templates/
    ├── deployment.yaml
    ├── service.yaml
    └── ingress.yaml
```

```bash
helm install video-app ./video-app --values values-prod.yaml
```

## Ключевые выводы

- K8s для продакшена: автоматический рестарт, масштабирование, rolling updates
- Readiness/Liveness пробы — обязательно
- Helm для управления несколькими сервисами
- Docker Compose для локальной разработки, K8s для продакшена

<Quiz quizId="p3-08-deployment" questions={[
  {id:"q1",question:"Чем отличается readiness от liveness пробы?",options:["Ничем","Readiness проверяет готовность принимать трафик, liveness — жив ли процесс. Если readiness упала — под исключается из балансировки. Если liveness — перезапускается.","Readiness для HTTP, liveness для TCP","Readiness проверяется при старте, liveness — постоянно"],correctIndex:1,explanation:"Readiness определяет попадает ли под в Service. Liveness — нужно ли перезапустить контейнер. Readiness может временно упасть (БД недоступна), liveness — признак что процесс мёртв."}
]} />
