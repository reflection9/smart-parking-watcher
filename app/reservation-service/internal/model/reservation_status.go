package model

type ReservationStatus string

const (
	ReservationStatusActive    ReservationStatus = "ACTIVE"
	ReservationStatusConfirmed ReservationStatus = "CONFIRMED"
	ReservationStatusCancelled ReservationStatus = "CANCELLED"
	ReservationStatusExpired   ReservationStatus = "EXPIRED"
)
