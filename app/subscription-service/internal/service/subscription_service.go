package service

import (
	"context"
	"subscription-service/internal/dto"
)

type SubscriptionService interface {
	Create(ctx context.Context, req dto.CreateSubscriptionRequest) (*dto.SubscriptionResponse, error)
	ListByUserID(ctx context.Context, userID int64) ([]dto.SubscriptionResponse, error)
	ListByZoneID(ctx context.Context, zoneID int64) ([]dto.SubscriptionResponse, error)
	Delete(ctx context.Context, userID, zoneID int64) error
}
