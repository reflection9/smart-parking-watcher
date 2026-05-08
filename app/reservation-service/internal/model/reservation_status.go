package model

type ReservationStatus string

const (
	ReservationStatusPending    ReservationStatus = "PENDING"
	ReservationStatusActive     ReservationStatus = "ACTIVE"
	ReservationStatusConfirming ReservationStatus = "CONFIRMING"
	ReservationStatusCancelling ReservationStatus = "CANCELLING"
	ReservationStatusExpiring   ReservationStatus = "EXPIRING"
	ReservationStatusConfirmed  ReservationStatus = "CONFIRMED"
	ReservationStatusCancelled  ReservationStatus = "CANCELLED"
	ReservationStatusExpired    ReservationStatus = "EXPIRED"
	ReservationStatusFailed     ReservationStatus = "FAILED"
)
