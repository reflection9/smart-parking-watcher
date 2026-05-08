package service

import (
	"context"
	"event-service/internal/dto"
	"event-service/internal/messaging"
)

type EventService interface {
	Create(ctx context.Context, req dto.CreateEventRequest) (*dto.EventResponse, error)
	ListByZoneID(ctx context.Context, zoneID int64) ([]dto.EventResponse, error)
	ListBySpotID(ctx context.Context, spotID int64) ([]dto.EventResponse, error)
	ListByReservationID(ctx context.Context, reservationID int64) ([]dto.EventResponse, error)
	HandleReservationEvent(ctx context.Context, event messaging.ReservationLifecycleEvent) error
	HandleSpotEvent(ctx context.Context, event messaging.SpotStatusEvent) error
}
