package dto

type CreateEventRequest struct {
	ZoneID    int64  `json:"zone_id" binding:"required"`
	SpotID    int64  `json:"spot_id" binding:"required"`
	EventType string `json:"event_type" binding:"required"`
	OldStatus string `json:"old_status"`
	NewStatus string `json:"new_status" binding:"required"`
}
