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
	reservationRepo   repository.ReservationRepository
	userLookupClient  UserLookupClient
	parkingSpotClient ParkingSpotClient
	reservationTTL    time.Duration
	eventPublisher    messaging.ReservationEventPublisher
}

func NewReservationService(
	reservationRepo repository.ReservationRepository,
	userLookupClient UserLookupClient,
	parkingSpotClient ParkingSpotClient,
	reservationTTL time.Duration,
	eventPublisher messaging.ReservationEventPublisher,
) ReservationService {
	if reservationTTL <= 0 {
		reservationTTL = 5 * time.Minute
	}
	if eventPublisher == nil {
		eventPublisher = messaging.NewNoopReservationEventPublisher()
	}

	return &reservationService{
		reservationRepo:   reservationRepo,
		userLookupClient:  userLookupClient,
		parkingSpotClient: parkingSpotClient,
		reservationTTL:    reservationTTL,
		eventPublisher:    eventPublisher,
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

	activeReservation, err := s.reservationRepo.GetActiveBySpot(ctx, req.ZoneID, req.SpotID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if activeReservation != nil {
		if time.Now().After(activeReservation.ExpiresAt) || time.Now().Equal(activeReservation.ExpiresAt) {
			if err := s.expireExistingReservation(ctx, activeReservation, spot.Status); err != nil {
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

	if err := s.parkingSpotClient.ReserveSpot(ctx, req.ZoneID, req.SpotID); err != nil {
		if errors.Is(err, ErrSpotNotFound) || errors.Is(err, ErrSpotUnavailable) {
			return nil, err
		}

		return nil, wrapDependencyError(err)
	}

	now := time.Now()
	reservation := &model.Reservation{
		UserID:    req.UserID,
		ZoneID:    req.ZoneID,
		SpotID:    req.SpotID,
		Status:    model.ReservationStatusActive,
		ExpiresAt: now.Add(s.reservationTTL),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.reservationRepo.Create(ctx, reservation); err != nil {
		_ = s.parkingSpotClient.ReleaseSpot(ctx, req.ZoneID, req.SpotID)
		return nil, err
	}

	s.publishReservationEvent(ctx, reservation, messaging.ReservationCreatedEvent)

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
	reservation, err := s.getActiveReservation(ctx, id)
	if err != nil {
		return nil, err
	}
	if time.Now().After(reservation.ExpiresAt) || time.Now().Equal(reservation.ExpiresAt) {
		return nil, ErrReservationExpired
	}

	if err := s.parkingSpotClient.OccupySpot(ctx, reservation.ZoneID, reservation.SpotID); err != nil {
		if errors.Is(err, ErrSpotNotFound) || errors.Is(err, ErrSpotUnavailable) {
			return nil, err
		}

		return nil, wrapDependencyError(err)
	}

	now := time.Now()
	reservation.Status = model.ReservationStatusConfirmed
	reservation.ConfirmedAt = &now
	reservation.UpdatedAt = now

	if err := s.reservationRepo.Update(ctx, reservation); err != nil {
		return nil, err
	}

	s.publishReservationEvent(ctx, reservation, messaging.ReservationConfirmedEvent)

	response := toReservationResponse(*reservation)
	return &response, nil
}

func (s *reservationService) Cancel(ctx context.Context, id uint) (*dto.ReservationResponse, error) {
	reservation, err := s.getActiveReservation(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.parkingSpotClient.ReleaseSpot(ctx, reservation.ZoneID, reservation.SpotID); err != nil {
		if errors.Is(err, ErrSpotNotFound) || errors.Is(err, ErrSpotUnavailable) {
			return nil, err
		}

		return nil, wrapDependencyError(err)
	}

	now := time.Now()
	reservation.Status = model.ReservationStatusCancelled
	reservation.ConfirmedAt = nil
	reservation.UpdatedAt = now

	if err := s.reservationRepo.Update(ctx, reservation); err != nil {
		return nil, err
	}

	s.publishReservationEvent(ctx, reservation, messaging.ReservationCancelledEvent)

	response := toReservationResponse(*reservation)
	return &response, nil
}

func (s *reservationService) Expire(ctx context.Context, id uint) (*dto.ReservationResponse, error) {
	reservation, err := s.getActiveReservation(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.parkingSpotClient.ReleaseSpot(ctx, reservation.ZoneID, reservation.SpotID); err != nil {
		if errors.Is(err, ErrSpotNotFound) || errors.Is(err, ErrSpotUnavailable) {
			return nil, err
		}

		return nil, wrapDependencyError(err)
	}

	now := time.Now()
	reservation.Status = model.ReservationStatusExpired
	reservation.ConfirmedAt = nil
	reservation.UpdatedAt = now

	if err := s.reservationRepo.Update(ctx, reservation); err != nil {
		return nil, err
	}

	s.publishReservationEvent(ctx, reservation, messaging.ReservationExpiredEvent)

	response := toReservationResponse(*reservation)
	return &response, nil
}

func (s *reservationService) getActiveReservation(
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

	if reservation.Status != model.ReservationStatusActive {
		return nil, ErrReservationNotActive
	}

	return reservation, nil
}

func (s *reservationService) expireExistingReservation(
	ctx context.Context,
	reservation *model.Reservation,
	spotStatus string,
) error {
	if reservation.Status != model.ReservationStatusActive {
		return nil
	}

	if spotStatus == "RESERVED" {
		if err := s.parkingSpotClient.ReleaseSpot(ctx, reservation.ZoneID, reservation.SpotID); err != nil {
			if errors.Is(err, ErrSpotNotFound) {
				return err
			}
			if !errors.Is(err, ErrSpotUnavailable) {
				return wrapDependencyError(err)
			}
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
