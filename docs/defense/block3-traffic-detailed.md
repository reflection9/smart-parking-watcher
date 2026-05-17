# Блок 3. Service Mesh, Ingress И Rate Limiting

## Что Требовалось По ТЗ

В третьем блоке требовалось:

- организовать service mesh
- сделать внешнюю точку входа в систему
- реализовать rate limiting

## Что Реализовано

Выбран стек:

- `Istio` для east-west трафика
- `ingress-nginx` как Ingress Controller
- `gateway nginx` как прикладной API gateway и слой rate limiting

Ключевые файлы:

- [infra/gitops/platform/00-istio-base.yaml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/gitops/platform/00-istio-base.yaml)
- [infra/gitops/platform/01-istiod.yaml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/gitops/platform/01-istiod.yaml)
- [infra/gitops/platform/02-ingress-nginx.yaml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/gitops/platform/02-ingress-nginx.yaml)
- [infra/gitops/platform/03-smart-parking-traffic.yaml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/gitops/platform/03-smart-parking-traffic.yaml)
- [infra/k8s/platform/traffic/00-smart-parking-mesh.yaml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/k8s/platform/traffic/00-smart-parking-mesh.yaml)
- [infra/k8s/platform/traffic/01-edge-ingress.yaml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/k8s/platform/traffic/01-edge-ingress.yaml)
- [infra/k8s/app/services/07-gateway-config.yaml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/k8s/app/services/07-gateway-config.yaml)

## Как Устроен Трафик

Полный путь запроса такой:

1. Пользователь приходит извне в `ingress-nginx`
2. `ingress-nginx` прокидывает весь внешний HTTP-трафик в `gateway`
3. `gateway` маршрутизирует запрос по path в нужный backend
4. Внутри namespace трафик между сервисами проходит через `Istio` sidecar’ы

То есть у нас есть два разных слоя:

- edge ingress слой
- service-to-service mesh слой

## Почему Выбран `Istio`

`Istio` выбран вместо `Linkerd`, потому что в этом проекте было важно показать:

- `DestinationRule`
- `VirtualService`
- mTLS
- retries
- timeouts
- outlier detection

Для защиты это сильнее и нагляднее, чем более минималистичный mesh.

## Что Именно Делает Mesh

### mTLS

В namespace `smart-parking` включен:

- `PeerAuthentication` с `mtls.mode: STRICT`

Это означает:

- сервисы внутри namespace общаются только через mutual TLS
- plaintext east-west communication не является нормальным режимом

### DestinationRule

Для `gateway` и backend-сервисов заданы `DestinationRule`, которые включают:

- `LEAST_REQUEST`
- connection pool limits
- `outlierDetection`

Смысл:

- балансировка идет не случайно, а по нагрузке
- сервис с плохими ответами может временно исключаться из пула
- это и есть приближенная реализация circuit-breaker behavior на уровне сервисной политики

### VirtualService

Для read-only трафика настроены retry policy и timeout:

- retries только на `GET`
- `attempts: 2`
- `perTryTimeout: 1s`
- `timeout: 3s`

Почему только на `GET`:

- повторять write-запросы опасно
- повторный `POST` может создать дубли или менять состояние системы

## Почему Выбран `ingress-nginx`

Для локального `k3d` это самый практичный выбор:

- он легко встает через Helm
- хорошо работает с hostPort
- проще, чем отдельный `HAProxy + Keepalived` контур

В нашей локальной среде вход организован так:

- `k3d` load balancer принимает внешний порт
- `ingress-nginx` обслуживает ingress rules

## Почему Rate Limiting Живет В Gateway

Это очень важный момент защиты.

Rate limiting сделан не в `ingress-nginx`, а в `gateway-nginx-config`.

Причина:

- gateway знает конкретные бизнес-маршруты
- разные endpoint’ы требуют разных лимитов
- часть операций write-heavy и чувствительна к дублированию

Реализованные ограничения:

- `POST /users/register` -> `5 req/min`
- `POST /users/login` -> `10 req/min`
- `POST /subscriptions` -> `20 req/min`
- `POST /reservations` и state transition’ы -> `15 req/min`
- `POST /history/archive` -> `2 req/min`

Именно gateway здесь принимает решение, какой путь относится к какому бизнес-лимиту.

## Как Это Работает В Рантайме

При входящем запросе:

1. запрос приходит в `ingress-nginx`
2. попадает в `gateway`
3. в `gateway` отрабатывает `limit_req_zone`
4. если лимит не превышен, запрос проксируется в нужный сервис
5. дальше east-west вызовы уже идут через `Istio`

## Коварные Вопросы И Ответы

### Почему и Ingress, и Gateway одновременно

Потому что это разные задачи:

- `ingress-nginx` это точка входа в кластер
- `gateway` это прикладной reverse proxy с бизнес-маршрутизацией и rate limiting

### Почему не использовать только Istio Gateway

Потому что текущая архитектура уже имела отдельный `gateway`, и для проекта было выгоднее сохранить явный API layer вместо полной замены edge-архитектуры.

### Где здесь circuit breaker

Он реализован через `DestinationRule`:

- connection pool limits
- outlier detection
- ограничение запросов на соединение

Это не отдельная кнопка с именем “CircuitBreaker”, а набор traffic policy primitives, из которых он и собирается.

### Почему retry только на GET

Потому что retry на state-changing запросах может приводить к некорректной повторной записи или повторной бизнес-операции.

## Честные Ограничения

- внешний edge слой не High Availability в продовом смысле, а локальный учебный ingress
- rate limiting реализован на gateway уровне, а не как централизованный external ratelimit service
- это правильно для текущего проекта, но не единственный возможный production pattern

## Итог По Блоку

Блок 3 закрыт как двухслойная traffic-модель:

- внешний вход через `ingress-nginx`
- прикладная маршрутизация и rate limiting в `gateway`
- mesh-политики и mTLS через `Istio`

Это логичная и хорошо защищаемая архитектура для локального Kubernetes-проекта.
