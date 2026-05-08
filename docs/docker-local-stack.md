# Local Docker Stack

## Что поднимается

- PostgreSQL
- MongoDB
- Kafka
- Kafka UI
- `user-service`
- `parking-service`
- `subscription-service`
- `reservation-service`
- `event-service`
- `notification-service`

## Как запустить

Из корня проекта:

```bash
docker compose up --build
```

## Доступные порты

- `8081` — `user-service`
- `8082` — `subscription-service`
- `8083` — `parking-service`
- `8084` — `event-service`
- `8085` — `notification-service`
- `8086` — `reservation-service`
- `8088` — Kafka UI
- `5432` — PostgreSQL
- `27017` — MongoDB
- `9094` — Kafka с хоста

## Важно

- Внутри Docker-сети сервисы используют Kafka по адресу `kafka:9092`.
- С хоста Kafka доступна по адресу `localhost:9094`.
- Postgres-миграции применяются отдельными one-shot контейнерами перед запуском сервисов.
