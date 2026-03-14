package dto

import "time"

type EventResponse struct {
	ID        string    `json:"id"`
	ZoneID    int64     `json:"zone_id"`
	SpotID    int64     `json:"spot_id"`
	EventType string    `json:"event_type"`
	OldStatus string    `json:"old_status,omitempty"`
	NewStatus string    `json:"new_status"`
	CreatedAt time.Time `json:"created_at"`
}
