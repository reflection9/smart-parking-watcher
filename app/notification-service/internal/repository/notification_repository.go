package repository

import (
	"context"
	"notification-service/internal/model"
)

type NotificationFilter struct {
	UserID  *int64
	ZoneID  *int64
	EventID string
	Status  *model.NotificationStatus
	Limit   int
}

type NotificationRepository interface {
	CreatePending(ctx context.Context, notification *model.Notification) (bool, error)
	Update(ctx context.Context, notification *model.Notification) error
	List(ctx context.Context, filter NotificationFilter) ([]model.Notification, error)
	GetByID(ctx context.Context, id uint) (*model.Notification, error)
}
