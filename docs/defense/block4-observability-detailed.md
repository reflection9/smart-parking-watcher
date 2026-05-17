# Блок 4. Observability

## Что Требовалось По ТЗ

В четвертом блоке требовалось:

- настроить observability
- рассмотреть несколько вариантов стека
- аргументировать выбор
- покрыть метрики, логи и трейсы

## Что Реализовано

Выбран и реализован split-stack:

- `Prometheus`
- `Alertmanager`
- `Grafana`
- `Loki`
- `Promtail`
- `Jaeger`

Ключевые файлы:

- [shared/observability/observability.go](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/shared/observability/observability.go)
- [infra/gitops/platform/04-jaeger.yaml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/gitops/platform/04-jaeger.yaml)
- [infra/gitops/platform/05-smart-parking-observability.yaml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/gitops/platform/05-smart-parking-observability.yaml)
- [infra/k8s/platform/observability/00-prometheus-config.yaml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/k8s/platform/observability/00-prometheus-config.yaml)
- [infra/k8s/platform/observability/01-prometheus.yaml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/k8s/platform/observability/01-prometheus.yaml)
- [infra/k8s/platform/observability/02-alertmanager.yaml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/k8s/platform/observability/02-alertmanager.yaml)
- [infra/k8s/platform/observability/03-grafana.yaml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/k8s/platform/observability/03-grafana.yaml)
- [infra/k8s/platform/observability/04-loki.yaml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/k8s/platform/observability/04-loki.yaml)
- [infra/k8s/platform/observability/05-promtail.yaml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/k8s/platform/observability/05-promtail.yaml)
- [infra/k8s/platform/observability/06-ingresses.yaml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/k8s/platform/observability/06-ingresses.yaml)
- [infra/k8s/app/services/00-app-config.yaml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/k8s/app/services/00-app-config.yaml)

## Почему Выбран Именно Такой Стек

### Почему не ELK

`ELK` мощный, но для локальной среды:

- тяжелее по памяти и диску
- сложнее в эксплуатации
- избыточен для текущего стенда

### Почему не Victoria stack

`VictoriaMetrics` и `VictoriaLogs` хорошие инструменты, но:

- в проекте уже были Prometheus-style `/metrics`
- уже были существующие конфиги Loki/Promtail
- замена стека не давала критического выигрыша относительно объема переделок

### Почему не SigNoz

`SigNoz` это удобный all-in-one observability продукт, но:

- он поменял бы архитектуру с разделенного стека на монолитную платформу
- для локальной лабки это тяжелее
- проект уже был частично instrumented под другой подход

### Почему `Jaeger`

Потому что в коде уже есть OTLP-based tracing, и естественный endpoint для него:

- `Jaeger` collector / all-in-one

### Почему `Prometheus`

Потому что:

- сервисы уже отдают `/metrics`
- метрики оформлены в стиле Prometheus
- алерты уже формулируются на `PromQL`

## Как Идут Метрики

Каждый Go-сервис использует общий пакет:

- `shared/observability/observability.go`

Он:

- добавляет middleware для метрик
- поднимает endpoint `/metrics`
- считает:
  - `smart_parking_http_requests_total`
  - `smart_parking_http_request_duration_seconds`
  - `smart_parking_http_in_flight_requests`

Дальше схема такая:

1. Pod аннотирован `prometheus.io/scrape=true`
2. `Prometheus` обнаруживает pod через Kubernetes SD
3. `Prometheus` скрейпит `/metrics`
4. Метрики сохраняются во внутреннее TSDB хранилище Prometheus
5. `Grafana` берет их из `Prometheus`

## Как Считаются Алерты

Алерты объявлены в:

- [infra/k8s/platform/observability/00-prometheus-config.yaml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/k8s/platform/observability/00-prometheus-config.yaml)

Есть как минимум три класса алертов:

- сервис недоступен
- высокий процент `5xx`
- высокий `p95` latency

Flow такой:

1. `Prometheus` выполняет правила
2. при срабатывании формирует alert
3. отправляет alert в `Alertmanager`
4. `Alertmanager` группирует и маршрутизирует

В нашей локальной версии routing упрощенный, потому что задача блока была не про интеграцию с почтой или PagerDuty, а про observability plumbing.

## Как Идут Трейсы

### Что Делает Код

В `shared/observability/observability.go`:

- создается OTLP HTTP exporter
- включается `otelgin` middleware
- HTTP client-ы оборачиваются через `otelhttp.NewTransport`

Это значит:

- входящий HTTP-запрос в сервис становится root/server span
- исходящие HTTP-вызовы к другим сервисам тоже трассируются
- контекст трассы пробрасывается дальше

### Куда Идут Трейсы

Трейсы идут в:

- `jaeger.observability.svc.cluster.local:4318`

Это значение приходит из:

- `OTEL_EXPORTER_OTLP_ENDPOINT`

Важно:

- в коде используется `otlptracehttp.WithEndpoint(...)`
- значит endpoint ожидается как `host:port`, а не как полный URL с `http://`

Это хороший каверзный вопрос, на котором часто ловят.

### Как Дальше Обрабатываются Трейсы

1. сервис отправляет OTLP trace в `Jaeger`
2. `Jaeger` принимает и хранит trace data
3. `Grafana` может показывать trace UI через datasource `Jaeger`
4. сам `Jaeger` тоже доступен как отдельный UI

## Как Идут Логи

Логический pipeline такой:

1. контейнер пишет логи в stdout/stderr
2. kubelet складывает их в pod logs на ноде
3. `Promtail` как DaemonSet читает эти файлы с `hostPath`
4. `Promtail` пушит логи в `Loki`
5. `Grafana` читает их из `Loki`

То есть `Promtail` у нас агент сбора, а `Loki` хранилище и query backend.

## Почему Grafana Является Единой Точкой Визуализации

Потому что в ней одновременно заведены datasource:

- `Prometheus`
- `Loki`
- `Jaeger`

Именно поэтому можно объяснять observability как три сигнала с единым UI, а не как три разрозненных инструмента.

## GitOps Размещение

Здесь observability разделен на две части:

- `Jaeger` приходит как upstream chart через отдельный Argo app
- остальной observability stack хранится как локальные манифесты

Это нормально, потому что:

- `Jaeger` удобно взять как готовый chart
- `Prometheus/Loki/Grafana/Promtail` у нас настроены под конкретные локальные assumptions и проще поддерживаются локально

## Коварные Вопросы И Ответы

### Куда конкретно идут трейсы

В `Jaeger` по OTLP/HTTP на `jaeger.observability.svc.cluster.local:4318`.

### Откуда Prometheus знает, какие сервисы скрейпить

Он не скрейпит вручную перечисленный список Pod IP. Он использует Kubernetes service discovery и фильтрует pod’ы по `prometheus.io/scrape`.

### Почему не использовали Prometheus Operator

Потому что для локальной среды нужен более легкий стек. Полный operator stack тяжелее для `k3d + Docker Desktop`.

### Почему не Tempo

Потому что код уже был ориентирован на Jaeger-based tracing flow, и в текущем проекте смена backend-а не давала практического выигрыша.

### Почему Promtail, если он deprecated

Потому что для локального учебного контура важнее совместимость и минимальная перестройка архитектуры. Мы не скрываем этот компромисс.

## Честные Ограничения

- observability stack локальный и ephemeral
- storage в основном временное, не production-grade
- часть runtime rollout-а может быть чувствительна к состоянию локального `k3d/Cilium`
- AI monitoring специально не добавлялся

## Итог По Блоку

Блок 4 закрыт как полноценный split-stack observability:

- метрики идут в `Prometheus`
- алерты маршрутизируются через `Alertmanager`
- логи идут в `Loki`
- трейсы идут в `Jaeger`
- `Grafana` дает единый UI

Это хорошо объясняется на защите, потому что по каждому сигналу понятно:

- кто собирает
- кто хранит
- кто визуализирует
