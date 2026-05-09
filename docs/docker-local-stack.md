# Локальный Docker-стек

## Что поднимается

- Шлюз API (`Nginx`)
- `Grafana`
- `Prometheus`
- `Alertmanager`
- `Jaeger`
- `Loki`
- `Promtail`
- `PostgreSQL`
- `MongoDB`
- `Redis`
- `MinIO`
- `Kafka`
- `Kafka UI`
- `user-service`
- `parking-service`
- `subscription-service`
- `reservation-service`
- `history-service`
- `notification-service`

## Как запустить

Из корня проекта:

```bash
docker compose up --build
```

## Открытые порты

- `8080` — API Gateway
- `3000` — Grafana
- `8081` — `user-service`
- `8082` — `subscription-service`
- `8083` — `parking-service`
- `8084` — `history-service`
- `8085` — `notification-service`
- `8086` — `reservation-service`
- `8088` — Kafka UI
- `9090` — Prometheus
- `9093` — Alertmanager
- `16686` — Jaeger UI
- `3100` — Loki
- `5432` — PostgreSQL
- `6379` — Redis
- `27017` — MongoDB
- `9000` — MinIO API
- `9001` — MinIO Console
- `9094` — Kafka с хоста

## Примечания

- Основная внешняя точка входа в API: `http://localhost:8080`.
- `Nginx` маршрутизирует запросы во внутренние сервисы и применяет rate limiting к чувствительным endpoint'ам.
- `upstream`-блоки уже подготовлены для балансировки, поэтому позже можно добавить несколько экземпляров сервиса без изменения клиентских URL.
- `Grafana` доступна по адресу `http://localhost:3000`, логин/пароль: `admin/admin`.
- `Prometheus` доступен по адресу `http://localhost:9090`.
- `Alertmanager` доступен по адресу `http://localhost:9093`.
- `Jaeger UI` доступен по адресу `http://localhost:16686`.
- Каждый Go-сервис отдает `/metrics` и экспортирует трейсы по `OTLP HTTP` в `Jaeger`.
- `Loki` и `Promtail` собирают контейнерные логи для централизованного просмотра в `Grafana`.
- Внутри Docker сервисы используют Kafka по адресу `kafka:9092`.
- С хоста Kafka доступна по адресу `localhost:9094`.
- `Redis` используется в `reservation-service` для `TTL` брони и автоматического истечения.
- `Redis` также используется в `subscription-service` для кэша подписчиков зоны и подписок пользователя.
- `MinIO` используется в `history-service` как cold storage для архивной истории событий.
- Миграции PostgreSQL запускаются через одноразовые контейнеры до старта сервисов.
