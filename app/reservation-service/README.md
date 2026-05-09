# reservation-service

Сервис временного бронирования парковочных мест.

## Endpoint'ы

- `POST /reservations`
- `GET /reservations/:id`
- `GET /reservations/users/:userId`
- `POST /reservations/:id/confirm`
- `POST /reservations/:id/cancel`
- `POST /reservations/:id/expire`

## Примечания

- Сервис хранит брони в `PostgreSQL`.
- Жизненный цикл брони оркестрируется через Kafka: создание, подтверждение, отмена и истечение публикуют команды, которые `parking-service` обрабатывает асинхронно.
- Существование пользователя проверяется через `user-service`.
- События жизненного цикла брони публикуются в Kafka, если брокер настроен.
- `TTL` брони хранится в записи через `expires_at` и дублируется в `Redis` для автоматического истечения.
- `POST /reservations/:id/expire` остается доступным как ручной запасной вариант и для тестирования, но штатное истечение теперь может происходить автоматически через `Redis TTL`.
