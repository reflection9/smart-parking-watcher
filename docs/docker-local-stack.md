# Local Docker Stack

## What Starts

- PostgreSQL
- MongoDB
- Redis
- Kafka
- Kafka UI
- `user-service`
- `parking-service`
- `subscription-service`
- `reservation-service`
- `event-service`
- `notification-service`

## How To Run

From the project root:

```bash
docker compose up --build
```

## Exposed Ports

- `8081` - `user-service`
- `8082` - `subscription-service`
- `8083` - `parking-service`
- `8084` - `event-service`
- `8085` - `notification-service`
- `8086` - `reservation-service`
- `8088` - Kafka UI
- `5432` - PostgreSQL
- `6379` - Redis
- `27017` - MongoDB
- `9094` - Kafka from the host

## Notes

- Inside Docker, services use Kafka at `kafka:9092`.
- From the host machine, Kafka is available at `localhost:9094`.
- Redis is used by `reservation-service` for reservation TTL and automatic expiration.
- Redis is also used by `subscription-service` to cache zone subscribers for hot reads.
- PostgreSQL migrations run through one-shot containers before the services start.
