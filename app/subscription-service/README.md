# subscription-service

Minimal service for user subscriptions to parking zones.

## Endpoints

- `POST /subscriptions`
- `GET /subscriptions/users/:userId`
- `DELETE /subscriptions/users/:userId/zones/:zoneId`

## Notes

- Before creating a subscription, the service checks that the user exists in `user-service`.
- It also checks that the parking zone exists in `parking-service`.
