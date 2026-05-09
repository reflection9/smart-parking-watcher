# Kafka Reservation Flow

## Цель

Первая Kafka-итерация переводит жизненный цикл бронирования в событийную модель без отказа от существующего HTTP flow.

## Topic

- `reservation-lifecycle-events`

## Producer

- `reservation-service`

Публикует события:

- `reservation.created`
- `reservation.confirmed`
- `reservation.cancelled`
- `reservation.expired`

## Consumer

- `history-service`

Читает события из Kafka и сохраняет историю в MongoDB. Это делает `history-service` первым history/read-model сервисом во второй части архитектуры.

## Payload

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
