# notification-service

Сервис email-уведомлений для пользователей, подписанных на парковочные зоны.

## Endpoint'ы

- `POST /notifications/spot-freed`
- `GET /notifications`
- `GET /notifications/:id`

## Примечания

- Сервис хранит историю уведомлений в `PostgreSQL`.
- Повторные уведомления блокируются уникальной парой `event_id + user_id`.
- Идентификаторы подписчиков запрашиваются из `subscription-service` по HTTP.
- Email-адреса получателей запрашиваются из `user-service`.
- Когда Kafka настроена, сервис напрямую читает события `spot_freed` из брокера.
- Если `EMAIL_TRANSPORT=log`, отправка писем логируется вместо реальной отправки через SMTP.
