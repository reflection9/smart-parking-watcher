package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reservation-service/internal/dto"
	"reservation-service/internal/messaging"
	"reservation-service/internal/model"
	"reservation-service/internal/repository"
	"time"

	"gorm.io/gorm"
)

type reservationService struct {
	reservationRepo         repository.ReservationRepository
	userLookupClient        UserLookupClient
	parkingSpotClient       ParkingSpotClient
	reservationTTL          time.Duration
	eventPublisher          messaging.ReservationEventPublisher
	parkingCommandPublisher messaging.ParkingCommandPublisher
}

func NewReservationService(
	reservationRepo repository.ReservationRepository,
	userLookupClient UserLookupClient,
	parkingSpotClient ParkingSpotClient,
	reservationTTL time.Duration,
	eventPublisher messaging.ReservationEventPublisher,
	parkingCommandPublisher messaging.ParkingCommandPublisher,
) ReservationService {
	if reservationTTL <= 0 {
		reservationTTL = 5 * time.Minute
	}
	if eventPublisher == nil {
		eventPublisher = messaging.NewNoopReservationEventPublisher()
	}
	if parkingCommandPublisher == nil {
		parkingCommandPublisher = messaging.NewNoopParkingCommandPublisher()
	}

	return &reservationService{
		reservationRepo:         reservationRepo,
		userLookupClient:        userLookupClient,
		parkingSpotClient:       parkingSpotClient,
		reservationTTL:          reservationTTL,
		eventPublisher:          eventPublisher,
		parkingCommandPublisher: parkingCommandPublisher,
	}
}

func (s *reservationService) Create(
	ctx context.Context,
	req dto.CreateReservationRequest,
) (*dto.ReservationResponse, error) {
	userExists, err := s.userLookupClient.Exists(ctx, req.UserID)
	if err != nil {
		return nil, wrapDependencyError(err)
	}
	if !userExists {
		return nil, ErrUserNotFound
	}

	spot, err := s.parkingSpotClient.GetByID(ctx, req.ZoneID, req.SpotID)
	if err != nil {
		if errors.Is(err, ErrSpotNotFound) {
			return nil, err
		}

		return nil, wrapDependencyError(err)
	}

	currentReservation, err := s.reservationRepo.GetOpenBySpot(ctx, req.ZoneID, req.SpotID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if currentReservation != nil {
		if s.isExpired(currentReservation) {
			if err := s.expireExistingReservation(ctx, currentReservation, spot.Status); err != nil {
				return nil, err
			}
			spot.Status = "FREE"
		} else {
			return nil, ErrActiveReservationExists
		}
	}

	if spot.Status != "FREE" {
		return nil, ErrSpotUnavailable
	}

	now := time.Now()
	reservation := &model.Reservation{
		UserID:    req.UserID,
		ZoneID:    req.ZoneID,
		SpotID:    req.SpotID,
		Status:    model.ReservationStatusPending,
		ExpiresAt: now.Add(s.reservationTTL),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.reservationRepo.Create(ctx, reservation); err != nil {
		return nil, err
	}

	if err := s.publishSpotCommand(ctx, reservation, messaging.SpotReserveRequestedCommand); err != nil {
		s.markReservationFailed(ctx, reservation)
		return nil, wrapDependencyError(err)
	}

	response := toReservationResponse(*reservation)
	return &response, nil
}

func (s *reservationService) GetByID(ctx context.Context, id uint) (*dto.ReservationResponse, error) {
	reservation, err := s.reservationRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReservationNotFound
		}

		return nil, err
	}

	response := toReservationResponse(*reservation)
	return &response, nil
}

func (s *reservationService) ListByUserID(
	ctx context.Context,
	userID int64,
) ([]dto.ReservationResponse, error) {
	reservations, err := s.reservationRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	response := make([]dto.ReservationResponse, 0, len(reservations))
	for _, reservation := range reservations {
		response = append(response, toReservationResponse(reservation))
	}

	return response, nil
}

func (s *reservationService) Confirm(ctx context.Context, id uint) (*dto.ReservationResponse, error) {
	reservation, err := s.getReservationByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if reservation.Status != model.ReservationStatusActive {
		return nil, ErrReservationActionNotAllowed
	}
	if time.Now().After(reservation.ExpiresAt) || time.Now().Equal(reservation.ExpiresAt) {
		return nil, ErrReservationExpired
	}

	now := time.Now()
	reservation.Status = model.ReservationStatusConfirming
	reservation.ConfirmedAt = nil
	reservation.UpdatedAt = now
	if err := s.reservationRepo.Update(ctx, reservation); err != nil {
		return nil, err
	}

	if err := s.publishSpotCommand(ctx, reservation, messaging.SpotOccupyRequestedCommand); err != nil {
		reservation.Status = model.ReservationStatusActive
		reservation.UpdatedAt = time.Now()
		_ = s.reservationRepo.Update(ctx, reservation)
		return nil, wrapDependencyError(err)
	}

	response := toReservationResponse(*reservation)
	return &response, nil
}

func (s *reservationService) Cancel(ctx context.Context, id uint) (*dto.ReservationResponse, error) {
	reservation, err := s.getReservationByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	switch reservation.Status {
	case model.ReservationStatusPending:
		reservation.Status = model.ReservationStatusCancelled
		reservation.ConfirmedAt = nil
		reservation.UpdatedAt = now

		if err := s.reservationRepo.Update(ctx, reservation); err != nil {
			return nil, err
		}

		s.publishReservationEvent(ctx, reservation, messaging.ReservationCancelledEvent)
	case model.ReservationStatusActive:
		reservation.Status = model.ReservationStatusCancelling
		reservation.ConfirmedAt = nil
		reservation.UpdatedAt = now
		if err := s.reservationRepo.Update(ctx, reservation); err != nil {
			return nil, err
		}
		if err := s.publishSpotCommand(ctx, reservation, messaging.SpotReleaseRequestedCommand); err != nil {
			reservation.Status = model.ReservationStatusActive
			reservation.UpdatedAt = time.Now()
			_ = s.reservationRepo.Update(ctx, reservation)
			return nil, wrapDependencyError(err)
		}
	default:
		return nil, ErrReservationActionNotAllowed
	}

	response := toReservationResponse(*reservation)
	return &response, nil
}

func (s *reservationService) Expire(ctx context.Context, id uint) (*dto.ReservationResponse, error) {
	reservation, err := s.getReservationByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	switch reservation.Status {
	case model.ReservationStatusPending:
		reservation.Status = model.ReservationStatusExpired
		reservation.ConfirmedAt = nil
		reservation.UpdatedAt = now

		if err := s.reservationRepo.Update(ctx, reservation); err != nil {
			return nil, err
		}

		s.publishReservationEvent(ctx, reservation, messaging.ReservationExpiredEvent)
	case model.ReservationStatusActive:
		reservation.Status = model.ReservationStatusExpiring
		reservation.ConfirmedAt = nil
		reservation.UpdatedAt = now
		if err := s.reservationRepo.Update(ctx, reservation); err != nil {
			return nil, err
		}
		if err := s.publishSpotCommand(ctx, reservation, messaging.SpotReleaseRequestedCommand); err != nil {
			reservation.Status = model.ReservationStatusActive
			reservation.UpdatedAt = time.Now()
			_ = s.reservationRepo.Update(ctx, reservation)
			return nil, wrapDependencyError(err)
		}
	default:
		return nil, ErrReservationActionNotAllowed
	}

	response := toReservationResponse(*reservation)
	return &response, nil
}

func (s *reservationService) HandleSpotEvent(ctx context.Context, event messaging.SpotStatusEvent) error {
	if event.ReservationID == nil {
		return nil
	}

	reservation, err := s.reservationRepo.GetByID(ctx, uint(*event.ReservationID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}

		return err
	}

	switch event.EventType {
	case messaging.SpotReservedEvent:
		return s.handleSpotReserved(ctx, reservation)
	case messaging.SpotReservationRejectedEvent:
		return s.handleSpotReservationRejected(ctx, reservation)
	case messaging.SpotFreedEvent:
		return s.handleSpotFreed(ctx, reservation)
	case messaging.SpotReleaseRejectedEvent:
		return s.handleSpotReleaseRejected(ctx, reservation)
	case messaging.SpotOccupiedEvent:
		return s.handleSpotOccupied(ctx, reservation)
	case messaging.SpotOccupationRejectedEvent:
		return s.handleSpotOccupationRejected(ctx, reservation)
	default:
		return nil
	}
}

func (s *reservationService) getReservationByID(
	ctx context.Context,
	id uint,
) (*model.Reservation, error) {
	reservation, err := s.reservationRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReservationNotFound
		}

		return nil, err
	}

	return reservation, nil
}

func (s *reservationService) expireExistingReservation(
	ctx context.Context,
	reservation *model.Reservation,
	spotStatus string,
) error {
	if !isOpenReservationStatus(reservation.Status) {
		return nil
	}

	if spotStatus == "RESERVED" || spotStatus == "OCCUPIED" {
		if err := s.publishSpotCommand(ctx, reservation, messaging.SpotReleaseRequestedCommand); err != nil {
			return wrapDependencyError(err)
		}
	}

	now := time.Now()
	reservation.Status = model.ReservationStatusExpired
	reservation.ConfirmedAt = nil
	reservation.UpdatedAt = now

	if err := s.reservationRepo.Update(ctx, reservation); err != nil {
		return err
	}

	s.publishReservationEvent(ctx, reservation, messaging.ReservationExpiredEvent)

	return nil
}

func (s *reservationService) publishSpotCommand(
	ctx context.Context,
	reservation *model.Reservation,
	commandType string,
) error {
	command := messaging.SpotCommand{
		EventID:       messaging.NewEventID(),
		EventType:     commandType,
		Source:        "reservation-service",
		OccurredAt:    time.Now(),
		ReservationID: int64(reservation.ID),
		UserID:        reservation.UserID,
		ZoneID:        reservation.ZoneID,
		SpotID:        reservation.SpotID,
	}

	return s.parkingCommandPublisher.Publish(ctx, command)
}

func (s *reservationService) handleSpotReserved(
	ctx context.Context,
	reservation *model.Reservation,
) error {
	switch reservation.Status {
	case model.ReservationStatusPending:
	case model.ReservationStatusCancelling, model.ReservationStatusExpiring,
		model.ReservationStatusExpired, model.ReservationStatusCancelled, model.ReservationStatusFailed:
		if err := s.publishSpotCommand(ctx, reservation, messaging.SpotReleaseRequestedCommand); err != nil {
			return wrapDependencyError(err)
		}
		return nil
	default:
		return nil
	}

	if s.isExpired(reservation) {
		if err := s.expireExistingReservation(ctx, reservation, "RESERVED"); err != nil {
			return err
		}

		return nil
	}

	now := time.Now()
	reservation.Status = model.ReservationStatusActive
	reservation.UpdatedAt = now

	if err := s.reservationRepo.Update(ctx, reservation); err != nil {
		return err
	}

	s.publishReservationEvent(ctx, reservation, messaging.ReservationCreatedEvent)
	return nil
}

func (s *reservationService) handleSpotReservationRejected(
	ctx context.Context,
	reservation *model.Reservation,
) error {
	if reservation.Status != model.ReservationStatusPending {
		return nil
	}

	now := time.Now()
	reservation.Status = model.ReservationStatusFailed
	reservation.ConfirmedAt = nil
	reservation.UpdatedAt = now

	if err := s.reservationRepo.Update(ctx, reservation); err != nil {
		return err
	}

	s.publishReservationEvent(ctx, reservation, messaging.ReservationFailedEvent)
	return nil
}

func (s *reservationService) handleSpotFreed(
	ctx context.Context,
	reservation *model.Reservation,
) error {
	now := time.Now()
	switch reservation.Status {
	case model.ReservationStatusCancelling:
		reservation.Status = model.ReservationStatusCancelled
	case model.ReservationStatusExpiring:
		reservation.Status = model.ReservationStatusExpired
	case model.ReservationStatusCancelled, model.ReservationStatusExpired, model.ReservationStatusFailed:
		return nil
	default:
		return nil
	}

	reservation.ConfirmedAt = nil
	reservation.UpdatedAt = now
	if err := s.reservationRepo.Update(ctx, reservation); err != nil {
		return err
	}

	if reservation.Status == model.ReservationStatusCancelled {
		s.publishReservationEvent(ctx, reservation, messaging.ReservationCancelledEvent)
	} else {
		s.publishReservationEvent(ctx, reservation, messaging.ReservationExpiredEvent)
	}

	return nil
}

func (s *reservationService) handleSpotReleaseRejected(
	ctx context.Context,
	reservation *model.Reservation,
) error {
	switch reservation.Status {
	case model.ReservationStatusCancelling, model.ReservationStatusExpiring:
	default:
		return nil
	}

	reservation.Status = model.ReservationStatusActive
	reservation.UpdatedAt = time.Now()
	return s.reservationRepo.Update(ctx, reservation)
}

func (s *reservationService) handleSpotOccupied(
	ctx context.Context,
	reservation *model.Reservation,
) error {
	if reservation.Status != model.ReservationStatusConfirming {
		return nil
	}

	now := time.Now()
	reservation.Status = model.ReservationStatusConfirmed
	reservation.ConfirmedAt = &now
	reservation.UpdatedAt = now

	if err := s.reservationRepo.Update(ctx, reservation); err != nil {
		return err
	}

	s.publishReservationEvent(ctx, reservation, messaging.ReservationConfirmedEvent)
	return nil
}

func (s *reservationService) handleSpotOccupationRejected(
	ctx context.Context,
	reservation *model.Reservation,
) error {
	if reservation.Status != model.ReservationStatusConfirming {
		return nil
	}

	reservation.Status = model.ReservationStatusActive
	reservation.ConfirmedAt = nil
	reservation.UpdatedAt = time.Now()
	return s.reservationRepo.Update(ctx, reservation)
}

func (s *reservationService) markReservationFailed(
	ctx context.Context,
	reservation *model.Reservation,
) {
	if reservation == nil || reservation.Status != model.ReservationStatusPending {
		return
	}

	reservation.Status = model.ReservationStatusFailed
	reservation.UpdatedAt = time.Now()
	if err := s.reservationRepo.Update(ctx, reservation); err != nil {
		log.Printf("failed to mark reservation %d as failed: %v", reservation.ID, err)
		return
	}

	s.publishReservationEvent(ctx, reservation, messaging.ReservationFailedEvent)
}

func (s *reservationService) isExpired(reservation *model.Reservation) bool {
	if reservation == nil {
		return false
	}

	now := time.Now()
	return now.After(reservation.ExpiresAt) || now.Equal(reservation.ExpiresAt)
}

func isOpenReservationStatus(status model.ReservationStatus) bool {
	switch status {
	case model.ReservationStatusPending,
		model.ReservationStatusActive,
		model.ReservationStatusConfirming,
		model.ReservationStatusCancelling,
		model.ReservationStatusExpiring:
		return true
	default:
		return false
	}
}

func wrapDependencyError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrDependencyUnavailable) {
		return err
	}

	return fmt.Errorf("%w: %v", ErrDependencyUnavailable, err)
}

func toReservationResponse(reservation model.Reservation) dto.ReservationResponse {
	return dto.ReservationResponse{
		ID:          reservation.ID,
		UserID:      reservation.UserID,
		ZoneID:      reservation.ZoneID,
		SpotID:      reservation.SpotID,
		Status:      reservation.Status,
		ExpiresAt:   reservation.ExpiresAt,
		ConfirmedAt: reservation.ConfirmedAt,
		CreatedAt:   reservation.CreatedAt,
		UpdatedAt:   reservation.UpdatedAt,
	}
}

func (s *reservationService) publishReservationEvent(
	ctx context.Context,
	reservation *model.Reservation,
	eventType string,
) {
	if reservation == nil {
		return
	}

	occurredAt := reservation.UpdatedAt
	if occurredAt.IsZero() {
		occurredAt = time.Now()
	}

	event := messaging.ReservationLifecycleEvent{
		EventID:       messaging.NewEventID(),
		EventType:     eventType,
		Source:        "reservation-service",
		OccurredAt:    occurredAt,
		ReservationID: int64(reservation.ID),
		UserID:        reservation.UserID,
		ZoneID:        reservation.ZoneID,
		SpotID:        reservation.SpotID,
		Status:        string(reservation.Status),
		ExpiresAt:     &reservation.ExpiresAt,
		ConfirmedAt:   reservation.ConfirmedAt,
	}

	if err := s.eventPublisher.Publish(ctx, event); err != nil {
		log.Printf("failed to publish reservation event %s for reservation %d: %v", eventType, reservation.ID, err)
	}
}
