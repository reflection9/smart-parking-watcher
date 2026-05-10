# Описание всех 7 C4-диаграмм

Этот файл фиксирует состав и смысл всех семи C4-диаграмм проекта `SmartParkingWatcher`.

## 1. System Context

### Название

`C4 Level 1 — System Context`

### Что показывает

Система показывается как единый черный ящик, а вокруг нее располагаются внешние акторы и внешние системы.

### Основные элементы

- `Пользователь`
- `Администратор`
- `SmartParkingWatcher`
- `SMTP / Email Provider`

### Основная логика

- `Пользователь`:
  - регистрируется
  - входит в систему
  - подписывается на зоны
  - создает бронь
  - получает уведомления

- `Администратор`:
  - создает парковочные зоны
  - добавляет парковочные места
  - управляет статусами мест

- `SmartParkingWatcher`:
  - отправляет email через `SMTP / Email Provider`

### Важное замечание

Роль администратора в проекте сейчас логическая, а не role-based access control на уровне кода.

## 2. Container

### Название

`C4 Level 2 — Container`

### Что показывает

Всю систему на уровне сервисов, баз данных, брокера, кэша, cold storage и gateway.

### Основные контейнеры

- `API Gateway`
- `user-service`
- `parking-service`
- `subscription-service`
- `reservation-service`
- `history-service`
- `notification-service`

### Основные хранилища и инфраструктура

- `User DB (PostgreSQL)`
- `Parking DB (PostgreSQL)`
- `Subscription DB (PostgreSQL)`
- `Reservation DB (PostgreSQL)`
- `Notification DB (PostgreSQL)`
- `History DB (MongoDB)`
- `Kafka`
- `Redis`
- `MinIO`

### Основная логика

- `API Gateway` — единая точка входа
- `user-service` — пользователи
- `parking-service` — зоны и места
- `subscription-service` — подписки
- `reservation-service` — бронирование и orchestration saga
- `history-service` — теплая история и архивация
- `notification-service` — уведомления

## 3. Component — reservation-service

### Название

`C4 Level 3 — Reservation Service Components`

### Что показывает

Внутреннее устройство `reservation-service`.

### Основные компоненты

- `Reservation Handler`
- `Reservation Orchestrator`
- `Reservation Repository`
- `User Lookup Client`
- `Parking Spot Client`
- `Reservation TTL Tracker`
- `Parking Command Publisher`
- `Spot Event Consumer`

### Почему диаграмма важна

Она лучше всего показывает:

- orchestration saga
- Kafka-команды и события
- интеграцию с `Redis`
- работу с `PostgreSQL`

## 4. Component — notification-service

### Название

`C4 Level 3 — Notification Service Components`

### Что показывает

Внутреннее устройство `notification-service`.

### Основные компоненты

- `Notification Handler`
- `Notification Service`
- `Notification Repository`
- `Subscription Lookup Client`
- `User Lookup Client`
- `Email Sender`
- `Spot Event Consumer`

### Внешние зависимости

- `subscription-service`
- `user-service`
- `Notification DB`
- `Kafka`
- `SMTP / Email Provider`

### Почему диаграмма важна

Она показывает:

- реакцию на Kafka-события
- межсервисные HTTP-вызовы
- отправку уведомлений наружу

## 5. Component — history-service

### Название

`C4 Level 3 — History Service Components`

### Что показывает

Внутреннее устройство `history-service`.

### Основные компоненты

- `History Handler`
- `History Service`
- `History Repository`
- `Reservation Event Consumer`
- `Spot Event Consumer`
- `Archive Scheduler`
- `Cold Storage Client`

### Внешние зависимости

- `Kafka`
- `History DB (MongoDB)`
- `MinIO`

### Почему диаграмма важна

Она показывает:

- warm history
- cold storage
- data-oriented роль сервиса

## 6. Container — Data Flow and Storage

### Название

`C4 Level 2 — Data Flow and Storage`

### Что показывает

Систему с точки зрения хранения данных и их жизненного цикла.

### Основная логика

- `PostgreSQL` хранит `hot data`
- `Redis` хранит TTL и cache layer
- `MongoDB` хранит `warm history`
- `MinIO` хранит `cold data`
- `Kafka` транспортирует события

### Почему диаграмма важна

Она отдельно объясняет:

- горячие данные
- теплые данные
- холодные данные
- роль каждого хранилища

## 7. Container — Infrastructure and Platform View

### Название

`C4 Level 2 — Infrastructure and Platform View`

### Что показывает

Систему с акцентом на инфраструктурные компоненты.

### Основные элементы

- `API Gateway`
- `Kafka`
- `Redis`
- `MinIO`
- `Prometheus`
- `Alertmanager`
- `Grafana`
- `Jaeger`
- `Loki`
- `Promtail`
- backend-сервисы системы

### Почему диаграмма важна

Она отдельно показывает:

- routing и rate limiting
- observability
- платформенный слой системы

## Итоговый набор диаграмм

1. `System Context`
2. `Container`
3. `Component — reservation-service`
4. `Component — notification-service`
5. `Component — history-service`
6. `Container — Data Flow and Storage`
7. `Container — Infrastructure and Platform View`
