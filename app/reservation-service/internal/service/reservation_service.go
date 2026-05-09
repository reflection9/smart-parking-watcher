package service

import (
	"context"
	"reservation-service/internal/dto"
	"reservation-service/internal/messaging"
)

type ReservationService interface {
	Create(ctx context.Context, req dto.CreateReservationRequest) (*dto.ReservationResponse, error)
	GetByID(ctx context.Context, id uint) (*dto.ReservationResponse, error)
	ListByUserID(ctx context.Context, userID int64) ([]dto.ReservationResponse, error)
	Confirm(ctx context.Context, id uint) (*dto.ReservationResponse, error)
	Cancel(ctx context.Context, id uint) (*dto.ReservationResponse, error)
	Expire(ctx context.Context, id uint) (*dto.ReservationResponse, error)
	HandleTTLExpiration(ctx context.Context, id uint) error
	HandleSpotEvent(ctx context.Context, event messaging.SpotStatusEvent) error
}
