package expiration

import (
	"context"
	"time"
)

type ReservationTTLTracker interface {
	TrackReservation(ctx context.Context, reservationID uint, expiresAt time.Time) error
	RemoveReservation(ctx context.Context, reservationID uint) error
	Start(ctx context.Context, onExpired func(context.Context, uint) error) error
	Close() error
}

type noopReservationTTLTracker struct{}

func NewNoopReservationTTLTracker() ReservationTTLTracker {
	return &noopReservationTTLTracker{}
}

func (t *noopReservationTTLTracker) TrackReservation(context.Context, uint, time.Time) error {
	return nil
}

func (t *noopReservationTTLTracker) RemoveReservation(context.Context, uint) error {
	return nil
}

func (t *noopReservationTTLTracker) Start(context.Context, func(context.Context, uint) error) error {
	return nil
}

func (t *noopReservationTTLTracker) Close() error {
	return nil
}
