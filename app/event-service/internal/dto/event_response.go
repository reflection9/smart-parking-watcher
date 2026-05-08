package dto

import "time"

type EventResponse struct {
	ID            string     `json:"id"`
	EventID       string     `json:"event_id"`
	Source        string     `json:"source"`
	ReservationID *int64     `json:"reservation_id,omitempty"`
	UserID        *int64     `json:"user_id,omitempty"`
	ZoneID        int64      `json:"zone_id"`
	SpotID        int64      `json:"spot_id"`
	EventType     string     `json:"event_type"`
	Status        string     `json:"status,omitempty"`
	OldStatus     string     `json:"old_status,omitempty"`
	NewStatus     string     `json:"new_status,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	ConfirmedAt   *time.Time `json:"confirmed_at,omitempty"`
	OccurredAt    time.Time  `json:"occurred_at"`
	CreatedAt     time.Time  `json:"created_at"`
}
