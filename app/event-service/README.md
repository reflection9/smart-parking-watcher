# event-service

Event history service for manual parking events and reservation lifecycle events from Kafka.

## Endpoints

- `POST /events`
- `GET /events/zones/:zoneId`
- `GET /events/spots/:spotId`
- `GET /events/reservations/:reservationId`

## Notes

- Manual event creation over HTTP is still supported.
- When Kafka is configured, the service consumes reservation lifecycle events and stores them in MongoDB.
- Event history is idempotent by `event_id`, which allows safe repeated delivery from Kafka.
