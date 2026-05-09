# API Gateway, Balancing, and Rate Limiting

## Stack Choice

The project uses **Nginx** as a single infrastructure entry point in front of the Go microservices.

This gateway is responsible for:

- request routing to the correct service
- basic load-balancer readiness through `upstream` groups
- rate limiting for sensitive write endpoints

## Public Entry Point

- `http://localhost:8080`

## Routed Prefixes

- `/users/*` -> `user-service`
- `/zones/*` and `/spots/*` -> `parking-service`
- `/subscriptions/*` -> `subscription-service`
- `/reservations/*` -> `reservation-service`
- `/history/*` and legacy `/events/*` -> `history-service`
- `/notifications/*` -> `notification-service`

## Rate-Limited Endpoints

- `POST /users/register` -> `5 requests / minute`
- `POST /users/login` -> `10 requests / minute`
- `POST /subscriptions` -> `20 requests / minute`
- `POST /reservations` and reservation state transitions -> `15 requests / minute`
- `POST /history/archive` -> `2 requests / minute`

## Balancing Model

Each service is placed behind an Nginx `upstream` block. At the moment every upstream contains a single service instance, but the gateway is already prepared for horizontal scaling by adding more backend servers to the same upstream group.
