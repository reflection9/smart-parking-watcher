# Поток Kafka для бронирования

## Цель

Первая Kafka-итерация переводит жизненный цикл бронирования в событийную модель без отказа от существующего HTTP flow.

## Топик

- `reservation-lifecycle-events`

## Производитель

- `reservation-service`

Публикует события:

- `reservation.created`
- `reservation.confirmed`
- `reservation.cancelled`
- `reservation.expired`

## Потребитель

- `history-service`

Читает события из Kafka и сохраняет историю в MongoDB. Это делает `history-service` первым сервисом истории и read-model во второй части архитектуры.

## Содержимое события

Событие содержит:

- `event_id`
- `event_type`
- `source`
- `occurred_at`
- `reservation_id`
- `user_id`
- `zone_id`
- `spot_id`
- `status`
- `expires_at`
- `confirmed_at`

## Идемпотентность

На стороне `history-service` история дедуплицируется по `event_id`. Это позволяет безопасно обрабатывать повторную доставку Kafka-сообщений.
