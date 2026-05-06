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
- Spot state changes are delegated to `parking-service` over HTTP.
- User existence is validated through `user-service`.
- Reservation TTL is stored in the reservation record through `expires_at`.
- In this iteration expiration can be triggered explicitly through `POST /reservations/:id/expire`; automatic expiration will later move to Redis/Kafka based orchestration.

