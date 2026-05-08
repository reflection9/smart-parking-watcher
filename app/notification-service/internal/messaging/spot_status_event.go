package messaging

import "time"

const (
	SpotReservedEvent = "spot_reserved"
	SpotFreedEvent    = "spot_freed"
	SpotOccupiedEvent = "spot_occupied"
)

type SpotStatusEvent struct {
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
	Source     string    `json:"source"`
	OccurredAt time.Time `json:"occurred_at"`
	ZoneID     int64     `json:"zone_id"`
	SpotID     int64     `json:"spot_id"`
	Status     string    `json:"status"`
	OldStatus  string    `json:"old_status,omitempty"`
	NewStatus  string    `json:"new_status,omitempty"`
}
