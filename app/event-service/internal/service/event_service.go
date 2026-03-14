package service

import (
	"context"
	"event-service/internal/dto"
)

type EventService interface {
	Create(ctx context.Context, req dto.CreateEventRequest) (*dto.EventResponse, error)
	ListByZoneID(ctx context.Context, zoneID int64) ([]dto.EventResponse, error)
	ListBySpotID(ctx context.Context, spotID int64) ([]dto.EventResponse, error)
}
