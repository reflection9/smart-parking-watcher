# parking-service

Minimal parking management service.

## Endpoints

- `POST /zones`
- `GET /zones`
- `GET /zones/:id`
- `POST /zones/:zoneId/spots`
- `GET /spots/:spotId/zones/:zoneId`
- `PATCH /zones/:zoneId/spots/:spotId/status`
- `POST /zones/:zoneId/spots/:spotId/reserve`
- `POST /zones/:zoneId/spots/:spotId/release`
- `POST /zones/:zoneId/spots/:spotId/occupy`

## Notes

- When Kafka is configured, the service publishes `spot_reserved`, `spot_freed`, and `spot_occupied` events after successful status transitions.

