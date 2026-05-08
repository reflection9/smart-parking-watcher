package messaging

import "time"

const (
	SpotReserveRequestedCommand = "spot_reserve_requested"
	SpotReleaseRequestedCommand = "spot_release_requested"
	SpotOccupyRequestedCommand  = "spot_occupy_requested"
)

type SpotCommand struct {
	EventID       string    `json:"event_id"`
	EventType     string    `json:"event_type"`
	Source        string    `json:"source"`
	OccurredAt    time.Time `json:"occurred_at"`
	ReservationID int64     `json:"reservation_id"`
	UserID        int64     `json:"user_id"`
	ZoneID        int64     `json:"zone_id"`
	SpotID        int64     `json:"spot_id"`
}
