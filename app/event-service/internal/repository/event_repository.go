package repository

import (
	"context"
	"event-service/internal/model"
)

type EventRepository interface {
	Create(ctx context.Context, event *model.Event) (bool, error)
	ListByZoneID(ctx context.Context, zoneID int64) ([]model.Event, error)
	ListBySpotID(ctx context.Context, spotID int64) ([]model.Event, error)
	ListByReservationID(ctx context.Context, reservationID int64) ([]model.Event, error)
}
