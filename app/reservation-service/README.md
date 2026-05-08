# reservation-service

Reservation service for temporary parking spot bookings.

## Endpoints

- `POST /reservations`
- `GET /reservations/:id`
- `GET /reservations/users/:userId`
- `POST /reservations/:id/confirm`
- `POST /reservations/:id/cancel`
- `POST /reservations/:id/expire`

## Notes

- The service stores reservations in PostgreSQL.
- Reservation creation is orchestrated through Kafka: a new reservation starts in `PENDING`, `parking-service` reserves the spot asynchronously, and then the reservation moves to `ACTIVE` or `FAILED`.
- Confirmation, cancellation, and expiration still delegate spot state changes to `parking-service` over HTTP.
- User existence is validated through `user-service`.
- Reservation lifecycle events are published to Kafka when the broker is configured.
- Reservation TTL is stored in the reservation record through `expires_at`.
- In this iteration expiration can be triggered explicitly through `POST /reservations/:id/expire`; automatic expiration will later move to Redis/Kafka based orchestration.

