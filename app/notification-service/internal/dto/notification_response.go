package dto

import (
	"notification-service/internal/model"
	"time"
)

type NotificationResponse struct {
	ID             uint                     `json:"id"`
	EventID        string                   `json:"event_id"`
	EventType      string                   `json:"event_type"`
	UserID         int64                    `json:"user_id"`
	ZoneID         int64                    `json:"zone_id"`
	SpotID         *int64                   `json:"spot_id,omitempty"`
	RecipientEmail string                   `json:"recipient_email"`
	Subject        string                   `json:"subject"`
	Body           string                   `json:"body"`
	Status         model.NotificationStatus `json:"status"`
	ErrorMessage   *string                  `json:"error_message,omitempty"`
	SentAt         *time.Time               `json:"sent_at,omitempty"`
	CreatedAt      time.Time                `json:"created_at"`
	UpdatedAt      time.Time                `json:"updated_at"`
}
