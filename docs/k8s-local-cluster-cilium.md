# Локальный Kubernetes-кластер с k3d/k3s и Cilium

## Цель ветки

Эта ветка закрывает первую часть блока подготовки Kubernetes-инфраструктуры: поднимается локальный Kubernetes-кластер на базе k3s через k3d, а штатная сеть k3s заменяется на Cilium.

На этом этапе приложение еще не переносится из Docker Compose в Kubernetes. Цель — получить чистую платформенную основу, на которую следующим шагом будут добавляться Kubernetes-манифесты сервисов SmartParkingWatcher.

## Что выбрано

Для локальной разработки выбран `k3d`, потому что он запускает k3s-кластер внутри Docker-контейнеров и не требует отдельных виртуальных машин. В кластере создается один control-plane node и две worker-ноды.

В качестве CNI используется `Cilium`. Для этого при создании k3s отключаются встроенные сетевые компоненты, которые конфликтуют с внешним CNI:

- `flannel`;
- встроенный `network-policy` controller;
- `traefik`;
- `servicelb`.

`kube-proxy` пока оставлен включенным. Это упрощает локальную разработку и будущую интеграцию с Istio: Cilium уже является CNI, но полный kube-proxy replacement можно включать отдельным осознанным этапом.

## Что создается

Файл кластера:

```text
infra/k8s/local/k3d-cilium-cluster.yaml
```

Он поднимает:

- кластер `spw-local`;
- 1 server node;
- 2 agent node;
- локальный registry `localhost:5001` для будущих образов сервисов;
- проброс портов `8080 -> 80` и `8443 -> 443` для будущего ingress/gateway слоя.

Файл значений Cilium:

```text
infra/k8s/local/cilium-values.yaml
```

Он включает:

- Kubernetes IPAM;
- Cilium как CNI;
- Hubble relay для сетевой наблюдаемости;
- один экземпляр Cilium Operator для локального кластера.

## Требования на машине разработчика

Нужны установленные инструменты:

- Docker Desktop;
- k3d;
- kubectl;
- Helm;
- опционально Cilium CLI.

На Windows их можно поставить через `winget`:

```powershell
winget install Docker.DockerDesktop
winget install k3d.k3d
winget install Kubernetes.kubectl
winget install Helm.Helm
```

После установки Docker Desktop должен быть запущен.

## Создание кластера

PowerShell:

```powershell
.\scripts\k8s\create-local-cilium-cluster.ps1
```

Bash:

```bash
./scripts/k8s/create-local-cilium-cluster.sh
```

После успешного выполнения текущий `kubectl` context будет переключен на:

```text
k3d-spw-local
```

Проверка нод:

```bash
kubectl get nodes -o wide
```

Ожидаемый результат: одна server-нода и две agent-ноды в статусе `Ready`.

Проверка Cilium:

```bash
kubectl -n kube-system get pods -l k8s-app=cilium
kubectl -n kube-system get deployment cilium-operator
```

Если установлен Cilium CLI:

```bash
cilium status --wait
```

## Проверка Cilium Network Policy

Для проверки есть smoke-манифест:

```text
infra/k8s/local/cilium-policy-smoke.yaml
```

Он создает тестовый namespace `spw-smoke`, nginx-сервис и два клиента:

- `client-allowed` — должен иметь доступ к nginx;
- `client-blocked` — должен быть заблокирован CiliumNetworkPolicy.

Запуск проверки:

PowerShell:

```powershell
.\scripts\k8s\validate-local-cilium-cluster.ps1
```

Bash:

```bash
./scripts/k8s/validate-local-cilium-cluster.sh
```

Если проверка прошла, можно удалить тестовый namespace:

```bash
kubectl delete namespace spw-smoke
```

## Удаление кластера

PowerShell:

```powershell
.\scripts\k8s\delete-local-cilium-cluster.ps1
```

Bash:

```bash
./scripts/k8s/delete-local-cilium-cluster.sh
```

## Что делать дальше

Следующая ветка после создания локального кластера — перенос приложения в Kubernetes:

- namespaces;
- ConfigMap и Secret;
- PostgreSQL, MongoDB, Redis, Kafka/Strimzi, MinIO как инфраструктурные компоненты;
- Deployment и Service для Go-сервисов;
- health checks;
- подготовка образов через локальный registry `localhost:5001`.

