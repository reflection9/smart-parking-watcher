package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"history-service/internal/archive"
	"history-service/internal/dto"
	"history-service/internal/messaging"
	"history-service/internal/model"
	"history-service/internal/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type historyService struct {
	eventRepo         repository.EventRepository
	coldStorage       archive.ColdStorage
	archiveAfter      time.Duration
	archiveBatchLimit int64
}

func NewHistoryService(
	eventRepo repository.EventRepository,
	coldStorage archive.ColdStorage,
	archiveAfter time.Duration,
	archiveBatchLimit int64,
) HistoryService {
	if coldStorage == nil {
		coldStorage = archive.NewNoopColdStorage()
	}
	if archiveAfter <= 0 {
		archiveAfter = 30 * 24 * time.Hour
	}
	if archiveBatchLimit <= 0 {
		archiveBatchLimit = 500
	}

	return &historyService{
		eventRepo:         eventRepo,
		coldStorage:       coldStorage,
		archiveAfter:      archiveAfter,
		archiveBatchLimit: archiveBatchLimit,
	}
}

func (s *historyService) Create(ctx context.Context, req dto.CreateEventRequest) (*dto.EventResponse, error) {
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

func (s *historyService) ListByZoneID(ctx context.Context, zoneID int64) ([]dto.EventResponse, error) {
	events, err := s.eventRepo.ListByZoneID(ctx, zoneID)
	if err != nil {
		return nil, err
	}

	return toEventResponses(events), nil
}

func (s *historyService) ListBySpotID(ctx context.Context, spotID int64) ([]dto.EventResponse, error) {
	events, err := s.eventRepo.ListBySpotID(ctx, spotID)
	if err != nil {
		return nil, err
	}

	return toEventResponses(events), nil
}

func (s *historyService) ListByReservationID(
	ctx context.Context,
	reservationID int64,
) ([]dto.EventResponse, error) {
	events, err := s.eventRepo.ListByReservationID(ctx, reservationID)
	if err != nil {
		return nil, err
	}

	return toEventResponses(events), nil
}

func (s *historyService) ArchiveOldEvents(
	ctx context.Context,
	olderThan time.Duration,
) (*dto.ArchiveHistoryResponse, error) {
	if olderThan <= 0 {
		olderThan = s.archiveAfter
	}

	cutoff := time.Now().Add(-olderThan)
	events, err := s.eventRepo.ListOlderThan(ctx, cutoff, s.archiveBatchLimit)
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return &dto.ArchiveHistoryResponse{
			ArchivedCount: 0,
			Cutoff:        cutoff,
		}, nil
	}

	payload, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return nil, err
	}

	objectKey := buildArchiveObjectKey(events[0].OccurredAt)
	if err := s.coldStorage.UploadArchive(ctx, objectKey, payload, "application/json"); err != nil {
		return nil, err
	}

	ids := make([]bson.ObjectID, 0, len(events))
	for _, event := range events {
		ids = append(ids, event.ID)
	}

	if err := s.eventRepo.DeleteByIDs(ctx, ids); err != nil {
		return nil, err
	}

	return &dto.ArchiveHistoryResponse{
		ArchivedCount: len(events),
		ObjectKey:     objectKey,
		Cutoff:        cutoff,
	}, nil
}

func (s *historyService) HandleReservationEvent(
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

func (s *historyService) HandleSpotEvent(
	ctx context.Context,
	event messaging.SpotStatusEvent,
) error {
	historyEvent := &model.Event{
		EventID:    event.EventID,
		Source:     normalizeValue(event.Source, "parking-service"),
		ZoneID:     event.ZoneID,
		SpotID:     event.SpotID,
		EventType:  event.EventType,
		Status:     event.Status,
		OldStatus:  event.OldStatus,
		NewStatus:  event.NewStatus,
		OccurredAt: event.OccurredAt,
		CreatedAt:  time.Now(),
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

func buildArchiveObjectKey(occurredAt time.Time) string {
	now := time.Now().UTC()
	return fmt.Sprintf(
		"history/%04d/%02d/%02d/archive-%s.json",
		occurredAt.UTC().Year(),
		occurredAt.UTC().Month(),
		occurredAt.UTC().Day(),
		now.Format("20060102T150405Z"),
	)
}
