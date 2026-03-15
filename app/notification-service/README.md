# notification-service

Email notification service for users subscribed to parking zones.

## Endpoints

- `POST /notifications/spot-freed`
- `GET /notifications`
- `GET /notifications/:id`

## Notes

- The service stores notification history in PostgreSQL.
- Duplicate notifications are blocked by the unique pair `event_id + user_id`.
- Subscriber IDs are requested from `subscription-service` over HTTP.
- Recipient emails are fetched from `user-service`.
- If `EMAIL_TRANSPORT=log`, email delivery is logged instead of sent through SMTP.
