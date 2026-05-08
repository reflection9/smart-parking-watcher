package messaging

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	SpotReservedEvent            = "spot_reserved"
	SpotFreedEvent               = "spot_freed"
	SpotOccupiedEvent            = "spot_occupied"
	SpotReservationRejectedEvent = "spot_reservation_rejected"
	SpotReleaseRejectedEvent     = "spot_release_rejected"
	SpotOccupationRejectedEvent  = "spot_occupation_rejected"
)

type SpotStatusEvent struct {
	EventID       string    `json:"event_id"`
	EventType     string    `json:"event_type"`
	Source        string    `json:"source"`
	OccurredAt    time.Time `json:"occurred_at"`
	ZoneID        int64     `json:"zone_id"`
	SpotID        int64     `json:"spot_id"`
	Status        string    `json:"status,omitempty"`
	OldStatus     string    `json:"old_status,omitempty"`
	NewStatus     string    `json:"new_status,omitempty"`
	ReservationID *int64    `json:"reservation_id,omitempty"`
	UserID        *int64    `json:"user_id,omitempty"`
	FailureReason string    `json:"failure_reason,omitempty"`
}

func NewEventID() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return fmt.Sprintf("evt-%d", time.Now().UnixNano())
	}

	return "evt-" + hex.EncodeToString(buffer)
}
