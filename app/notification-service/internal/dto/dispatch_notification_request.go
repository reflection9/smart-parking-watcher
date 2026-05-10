package dto

type DispatchNotificationRequest struct {
	EventID   string  `json:"event_id" binding:"required"`
	EventType string  `json:"event_type" binding:"required"`
	ZoneID    int64   `json:"zone_id" binding:"required"`
	SpotID    *int64  `json:"spot_id,omitempty"`
	UserIDs   []int64 `json:"user_ids,omitempty"`
}
