# subscription-service

Сервис подписок пользователей на парковочные зоны.

## Endpoint'ы

- `POST /subscriptions`
- `GET /subscriptions/users/:userId`
- `GET /subscriptions/zones/:zoneId`
- `DELETE /subscriptions/users/:userId/zones/:zoneId`

## Примечания

- Перед созданием подписки сервис проверяет существование пользователя в `user-service`.
- Также сервис проверяет существование парковочной зоны в `parking-service`.
- Список подписчиков зоны и список подписок пользователя кэшируются в `Redis` для быстрых чтений.
