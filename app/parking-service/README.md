# parking-service

Сервис управления парковочными зонами и парковочными местами.

## Endpoint'ы

- `POST /zones`
- `GET /zones`
- `GET /zones/:id`
- `POST /zones/:zoneId/spots`
- `GET /spots/:spotId/zones/:zoneId`
- `PATCH /zones/:zoneId/spots/:spotId/status`
- `POST /zones/:zoneId/spots/:spotId/reserve`
- `POST /zones/:zoneId/spots/:spotId/release`
- `POST /zones/:zoneId/spots/:spotId/occupy`

## Примечания

- Когда Kafka настроена, сервис публикует события `spot_reserved`, `spot_freed` и `spot_occupied` после успешных переходов статуса.
- В orchestration flow сервис также читает команды `spot_reserve_requested` из Kafka и отвечает связанными событиями парковочных мест для `reservation-service`.
