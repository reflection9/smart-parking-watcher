package repository

import (
	"context"
	"subscription-service/internal/model"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, subscription *model.Subscription) error
	GetByUserAndZone(ctx context.Context, userID, zoneID int64) (*model.Subscription, error)
	ListByUserID(ctx context.Context, userID int64) ([]model.Subscription, error)
	ListByZoneID(ctx context.Context, zoneID int64) ([]model.Subscription, error)
	DeleteByUserAndZone(ctx context.Context, userID, zoneID int64) error
}
