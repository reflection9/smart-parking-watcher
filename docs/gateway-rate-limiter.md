# Шлюз API, балансировка и ограничение частоты запросов

## Выбор стека

В проекте используется `Nginx` как единая инфраструктурная точка входа перед Go-микросервисами.

Этот шлюз отвечает за:

- маршрутизацию запросов в нужный сервис
- готовность к балансировке через `upstream`-группы
- ограничение частоты запросов для чувствительных write-endpoint'ов

## Внешняя точка входа

- `http://localhost:8080`

## Префиксы маршрутов

- `/users/*` -> `user-service`
- `/zones/*` и `/spots/*` -> `parking-service`
- `/subscriptions/*` -> `subscription-service`
- `/reservations/*` -> `reservation-service`
- `/history/*` и legacy `/events/*` -> `history-service`
- `/notifications/*` -> `notification-service`

## Endpoint'ы с ограничением частоты

- `POST /users/register` -> `5 запросов / минута`
- `POST /users/login` -> `10 запросов / минута`
- `POST /subscriptions` -> `20 запросов / минута`
- `POST /reservations` и переходы состояния брони -> `15 запросов / минута`
- `POST /history/archive` -> `2 запроса / минута`

## Модель балансировки

Каждый сервис расположен за `Nginx upstream`-блоком. Сейчас в каждом upstream находится один экземпляр сервиса, но gateway уже подготовлен к горизонтальному масштабированию: достаточно добавить новые backend-серверы в ту же upstream-группу.
