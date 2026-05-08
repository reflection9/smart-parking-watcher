package repository

import (
	"context"
	"reservation-service/internal/model"
)

type ReservationRepository interface {
	Create(ctx context.Context, reservation *model.Reservation) error
	Update(ctx context.Context, reservation *model.Reservation) error
	GetByID(ctx context.Context, id uint) (*model.Reservation, error)
	ListByUserID(ctx context.Context, userID int64) ([]model.Reservation, error)
	GetOpenBySpot(ctx context.Context, zoneID, spotID int64) (*model.Reservation, error)
}
