# Local Docker Stack

## What Starts

- API Gateway (Nginx)
- Grafana
- Prometheus
- Alertmanager
- Jaeger
- Loki
- Promtail
- PostgreSQL
- MongoDB
- Redis
- MinIO
- Kafka
- Kafka UI
- `user-service`
- `parking-service`
- `subscription-service`
- `reservation-service`
- `history-service`
- `notification-service`

## How To Run

From the project root:

```bash
docker compose up --build
```

## Exposed Ports

- `8080` - API Gateway
- `3000` - Grafana
- `8081` - `user-service`
- `8082` - `subscription-service`
- `8083` - `parking-service`
- `8084` - `history-service`
- `8085` - `notification-service`
- `8086` - `reservation-service`
- `8088` - Kafka UI
- `9090` - Prometheus
- `9093` - Alertmanager
- `16686` - Jaeger UI
- `3100` - Loki
- `5432` - PostgreSQL
- `6379` - Redis
- `27017` - MongoDB
- `9000` - MinIO API
- `9001` - MinIO Console
- `9094` - Kafka from the host

## Notes

- The preferred entry point for the application API is `http://localhost:8080`.
- Nginx routes requests to the internal services and applies rate limiting to sensitive endpoints.
- Upstream blocks are prepared for balancing, so the gateway can fan out to multiple service replicas later without changing client URLs.
- Grafana is available at `http://localhost:3000` with `admin/admin`.
- Prometheus is available at `http://localhost:9090`.
- Alertmanager is available at `http://localhost:9093`.
- Jaeger UI is available at `http://localhost:16686`.
- Every Go service exposes `/metrics` and exports traces over OTLP HTTP to Jaeger.
- Loki and Promtail collect container logs for centralized log browsing in Grafana.
- Inside Docker, services use Kafka at `kafka:9092`.
- From the host machine, Kafka is available at `localhost:9094`.
- Redis is used by `reservation-service` for reservation TTL and automatic expiration.
- Redis is also used by `subscription-service` to cache zone subscribers for hot reads.
- MinIO is used by `history-service` as cold storage for archived event history.
- PostgreSQL migrations run through one-shot containers before the services start.
