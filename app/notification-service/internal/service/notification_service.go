package service

import (
	"context"
	"notification-service/internal/dto"
)

type NotificationService interface {
	DispatchSpotFreed(ctx context.Context, req dto.DispatchNotificationRequest) (*dto.DispatchNotificationResponse, error)
	List(ctx context.Context, query dto.ListNotificationsQuery) ([]dto.NotificationResponse, error)
	GetByID(ctx context.Context, id uint) (*dto.NotificationResponse, error)
}
