package service

import "errors"

var (
	ErrUserNotFound            = errors.New("user not found")
	ErrSpotNotFound            = errors.New("parking spot not found")
	ErrSpotUnavailable         = errors.New("parking spot is not available")
	ErrReservationNotFound     = errors.New("reservation not found")
	ErrReservationNotActive    = errors.New("reservation is not active")
	ErrReservationExpired      = errors.New("reservation has already expired")
	ErrActiveReservationExists = errors.New("an active reservation already exists for this spot")
	ErrDependencyUnavailable   = errors.New("dependent service unavailable")
)
