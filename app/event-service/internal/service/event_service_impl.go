package service

import (
	"context"
	"event-service/internal/dto"
	"event-service/internal/messaging"
	"event-service/internal/model"
	"event-service/internal/repository"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
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
	now := time.Now()
	event := &model.Event{
		EventID:    "manual-" + bson.NewObjectID().Hex(),
		Source:     normalizeValue(req.Source, "manual"),
		ZoneID:     req.ZoneID,
		SpotID:     req.SpotID,
		EventType:  strings.ToUpper(req.EventType),
		OldStatus:  strings.ToUpper(req.OldStatus),
		NewStatus:  strings.ToUpper(req.NewStatus),
		OccurredAt: now,
		CreatedAt:  now,
	}

	_, err := s.eventRepo.Create(ctx, event)
	if err != nil {
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

func (s *eventService) ListByReservationID(
	ctx context.Context,
	reservationID int64,
) ([]dto.EventResponse, error) {
	events, err := s.eventRepo.ListByReservationID(ctx, reservationID)
	if err != nil {
		return nil, err
	}

	return toEventResponses(events), nil
}

func (s *eventService) HandleReservationEvent(
	ctx context.Context,
	event messaging.ReservationLifecycleEvent,
) error {
	historyEvent := &model.Event{
		EventID:       event.EventID,
		Source:        normalizeValue(event.Source, "reservation-service"),
		ReservationID: &event.ReservationID,
		UserID:        &event.UserID,
		ZoneID:        event.ZoneID,
		SpotID:        event.SpotID,
		EventType:     event.EventType,
		Status:        event.Status,
		ExpiresAt:     event.ExpiresAt,
		ConfirmedAt:   event.ConfirmedAt,
		OccurredAt:    event.OccurredAt,
		CreatedAt:     time.Now(),
	}

	_, err := s.eventRepo.Create(ctx, historyEvent)
	return err
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
		ID:            event.ID.Hex(),
		EventID:       event.EventID,
		Source:        event.Source,
		ReservationID: event.ReservationID,
		UserID:        event.UserID,
		ZoneID:        event.ZoneID,
		SpotID:        event.SpotID,
		EventType:     event.EventType,
		Status:        event.Status,
		OldStatus:     event.OldStatus,
		NewStatus:     event.NewStatus,
		ExpiresAt:     event.ExpiresAt,
		ConfirmedAt:   event.ConfirmedAt,
		OccurredAt:    event.OccurredAt,
		CreatedAt:     event.CreatedAt,
	}
}

func normalizeValue(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}

	return value
}
