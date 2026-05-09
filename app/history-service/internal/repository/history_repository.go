package repository

import (
	"context"
	"history-service/internal/model"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type EventRepository interface {
	Create(ctx context.Context, event *model.Event) (bool, error)
	ListByZoneID(ctx context.Context, zoneID int64) ([]model.Event, error)
	ListBySpotID(ctx context.Context, spotID int64) ([]model.Event, error)
	ListByReservationID(ctx context.Context, reservationID int64) ([]model.Event, error)
	ListOlderThan(ctx context.Context, cutoff time.Time, limit int64) ([]model.Event, error)
	DeleteByIDs(ctx context.Context, ids []bson.ObjectID) error
}
