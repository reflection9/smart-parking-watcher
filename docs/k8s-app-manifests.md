# Kubernetes-манифесты SmartParkingWatcher

## Цель ветки

Эта ветка переносит локальный Docker Compose runtime в Kubernetes runtime. На предыдущем этапе был поднят локальный кластер `k3d/k3s` с Cilium. На этом этапе в кластер добавляются базовые Kubernetes-манифесты для приложения и инфраструктурных зависимостей.

Это еще не GitOps и не production-grade Helm chart. Цель ветки — получить воспроизводимую локальную точку: собрать образы сервисов, запушить их в локальный registry `localhost:5001`, поднять инфраструктуру, выполнить миграции и запустить backend-сервисы через Kubernetes.

## Что добавлено

Основные директории:

```text
infra/k8s/app/infra       # инфраструктурные компоненты
infra/k8s/app/migrations  # ConfigMap с SQL и Kubernetes Jobs миграций
infra/k8s/app/services    # Go-сервисы и gateway
scripts/k8s               # скрипты сборки, деплоя, проверки и удаления
```

Namespace приложения:

```text
smart-parking
```

## Что поднимается в Kubernetes

Инфраструктура:

- PostgreSQL как `StatefulSet` + PVC;
- MongoDB как `StatefulSet` + PVC;
- MinIO как `StatefulSet` + PVC;
- Redis как `Deployment`;
- Kafka как локальный single-node `StatefulSet` в KRaft-режиме.

Backend-сервисы:

- `user-service`;
- `parking-service`;
- `subscription-service`;
- `reservation-service`;
- `history-service`;
- `notification-service`;
- `gateway` на Nginx.

Миграции PostgreSQL запускаются отдельными `Job`:

- `user-db-migrate`;
- `parking-db-migrate`;
- `subscription-db-migrate`;
- `reservation-db-migrate`;
- `notification-db-migrate`.

## Важное архитектурное уточнение по Kafka

На этом этапе Kafka поднимается обычным single-node `StatefulSet`, чтобы приложение можно было локально запустить в Kubernetes уже сейчас.

Это временный dev-вариант. В следующем блоке по IaC/GitOps Kafka будет переноситься на Strimzi, потому что в задании отдельно указан Kubernetes-оператор Strimzi и Ansible Role для его установки.

То есть текущий Kafka-манифест нужен как промежуточный локальный runtime, а не как финальная production-модель Kafka в Kubernetes.

## Требования

Перед деплоем должен быть уже поднят локальный кластер из предыдущей ветки:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\k8s\create-local-cilium-cluster.ps1
```

Также нужен локальный registry `localhost:5001`, который создается k3d-конфигом предыдущего этапа.

Проверить контекст:

```powershell
kubectl config current-context
kubectl get nodes
```

Ожидаемый context:

```text
k3d-spw-local
```

## Деплой одной командой

PowerShell:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\k8s\deploy-local-app.ps1 -BuildImages
```

Bash:

```bash
./scripts/k8s/deploy-local-app.sh --build-images
```

Эта команда:

1. собирает Docker-образы Go-сервисов;
2. пушит их в `localhost:5001`;
3. применяет Kubernetes-манифесты инфраструктуры;
4. ждет готовности PostgreSQL, MongoDB, Redis, MinIO и Kafka;
5. пересоздает и запускает Jobs миграций;
6. применяет Deployments/Services приложения;
7. ждет готовности всех backend-сервисов и gateway.

## Если образы уже собраны

Можно не пересобирать образы:

PowerShell:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\k8s\deploy-local-app.ps1
```

Bash:

```bash
./scripts/k8s/deploy-local-app.sh
```

## Проверка состояния

```powershell
kubectl -n smart-parking get pods -o wide
kubectl -n smart-parking get svc
kubectl -n smart-parking get jobs
```

Все основные pod-ы должны быть в статусе `Running`, а migration jobs — `Completed`.

## Smoke-test приложения

PowerShell:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\k8s\validate-local-app.ps1
```

Bash:

```bash
./scripts/k8s/validate-local-app.sh
```

Проверка делает две вещи:

1. вызывает `http://gateway:8080/health` внутри кластера;
2. регистрирует тестового пользователя через gateway.

Если оба запроса прошли, значит gateway, user-service, PostgreSQL и базовый Kubernetes runtime работают.

## Локальный доступ к gateway из браузера

Пока ingress-слой еще не добавлен. Поэтому для доступа с хоста используется port-forward:

```powershell
kubectl -n smart-parking port-forward svc/gateway 8080:8080
```

После этого можно открыть:

```text
http://localhost:8080/health
```

## Удаление приложения из Kubernetes

PowerShell:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\k8s\delete-local-app.ps1
```

Bash:

```bash
./scripts/k8s/delete-local-app.sh
```

Команда удаляет namespace `smart-parking` со всеми объектами приложения.

## Что важно понимать

В этой ветке приложение уже запускается через Kubernetes, но это еще не финальная platform-архитектура.

Что уже есть:

- Kubernetes `Namespace`;
- `Deployment`, `StatefulSet`, `Service`;
- `ConfigMap`, `Secret`;
- `PersistentVolumeClaim`;
- migration `Job`;
- локальная сборка и загрузка образов в k3d registry;
- smoke-test через gateway.

Что будет отдельными следующими этапами:

- autoscaling через HPA / Cluster Autoscaler;
- Terraform для базовой инфраструктуры;
- ArgoCD и GitOps sync;
- Kafka через Strimzi;
- Istio service mesh;
- Ingress/Gateway API, HAProxy/Keepalived и rate limiting.
