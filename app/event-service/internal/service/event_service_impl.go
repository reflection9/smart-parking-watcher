package service

import (
	"context"
	"event-service/internal/dto"
	"event-service/internal/model"
	"event-service/internal/repository"
	"strings"
	"time"
)

type eventService struct {
	eventRepo repository.EventRepository
}

func NewEventService(eventRepo repository.EventRepository) EventService {
	return &eventService{
		eventRepo: eventRepo,
	}
}

func (s *eventService) Create(ctx context.Context, req dto.CreateEventRequest) (*dto.EventResponse, error) {
	event := &model.Event{
		ZoneID:    req.ZoneID,
		SpotID:    req.SpotID,
		EventType: strings.ToUpper(req.EventType),
		OldStatus: strings.ToUpper(req.OldStatus),
		NewStatus: strings.ToUpper(req.NewStatus),
		CreatedAt: time.Now(),
	}

	if err := s.eventRepo.Create(ctx, event); err != nil {
		return nil, err
	}

	response := toEventResponse(*event)
	return &response, nil
}

func (s *eventService) ListByZoneID(ctx context.Context, zoneID int64) ([]dto.EventResponse, error) {
	events, err := s.eventRepo.ListByZoneID(ctx, zoneID)
	if err != nil {
		return nil, err
	}

	return toEventResponses(events), nil
}

func (s *eventService) ListBySpotID(ctx context.Context, spotID int64) ([]dto.EventResponse, error) {
	events, err := s.eventRepo.ListBySpotID(ctx, spotID)
	if err != nil {
		return nil, err
	}

	return toEventResponses(events), nil
}

func toEventResponses(events []model.Event) []dto.EventResponse {
	response := make([]dto.EventResponse, 0, len(events))
	for _, event := range events {
		response = append(response, toEventResponse(event))
	}

	return response
}

func toEventResponse(event model.Event) dto.EventResponse {
	return dto.EventResponse{
		ID:        event.ID.Hex(),
		ZoneID:    event.ZoneID,
		SpotID:    event.SpotID,
		EventType: event.EventType,
		OldStatus: event.OldStatus,
		NewStatus: event.NewStatus,
		CreatedAt: event.CreatedAt,
	}
}
