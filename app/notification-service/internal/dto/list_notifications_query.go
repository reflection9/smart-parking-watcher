package dto

import "notification-service/internal/model"

type ListNotificationsQuery struct {
	UserID  *int64
	ZoneID  *int64
	EventID string
	Status  *model.NotificationStatus
	Limit   int
}
