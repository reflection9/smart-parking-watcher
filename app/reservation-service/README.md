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
- Reservation lifecycle is orchestrated through Kafka: creation, confirmation, cancellation, and expiration all publish commands that `parking-service` handles asynchronously.
- User existence is validated through `user-service`.
- Reservation lifecycle events are published to Kafka when the broker is configured.
- Reservation TTL is stored in the reservation record through `expires_at` and mirrored in Redis for automatic expiration.
- `POST /reservations/:id/expire` remains available for manual fallback and testing, but normal expiration can now happen automatically through Redis TTL.

