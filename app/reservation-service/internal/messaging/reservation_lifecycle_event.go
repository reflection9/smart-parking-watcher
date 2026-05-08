package messaging

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	ReservationCreatedEvent   = "reservation.created"
	ReservationConfirmedEvent = "reservation.confirmed"
	ReservationCancelledEvent = "reservation.cancelled"
	ReservationExpiredEvent   = "reservation.expired"
)

type ReservationLifecycleEvent struct {
	EventID       string     `json:"event_id"`
	EventType     string     `json:"event_type"`
	Source        string     `json:"source"`
	OccurredAt    time.Time  `json:"occurred_at"`
	ReservationID int64      `json:"reservation_id"`
	UserID        int64      `json:"user_id"`
	ZoneID        int64      `json:"zone_id"`
	SpotID        int64      `json:"spot_id"`
	Status        string     `json:"status"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	ConfirmedAt   *time.Time `json:"confirmed_at,omitempty"`
}

func NewEventID() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return fmt.Sprintf("evt-%d", time.Now().UnixNano())
	}

	return "evt-" + hex.EncodeToString(buffer)
}
