package dto

type DispatchNotificationResponse struct {
	EventID           string                 `json:"event_id"`
	EventType         string                 `json:"event_type"`
	ZoneID            int64                  `json:"zone_id"`
	SpotID            *int64                 `json:"spot_id,omitempty"`
	TotalSubscribers  int                    `json:"total_subscribers"`
	Processed         int                    `json:"processed"`
	Sent              int                    `json:"sent"`
	Failed            int                    `json:"failed"`
	DuplicatesSkipped int                    `json:"duplicates_skipped"`
	Notifications     []NotificationResponse `json:"notifications"`
}
