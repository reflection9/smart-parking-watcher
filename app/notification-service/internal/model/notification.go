package model

import "time"

type Notification struct {
	ID             uint               `gorm:"primaryKey" json:"id"`
	EventID        string             `gorm:"not null;index;uniqueIndex:idx_notifications_event_user,priority:1" json:"event_id"`
	EventType      string             `gorm:"not null" json:"event_type"`
	UserID         int64              `gorm:"not null;index;uniqueIndex:idx_notifications_event_user,priority:2" json:"user_id"`
	ZoneID         int64              `gorm:"not null;index" json:"zone_id"`
	SpotID         *int64             `gorm:"index" json:"spot_id,omitempty"`
	RecipientEmail string             `gorm:"not null;default:''" json:"recipient_email"`
	Subject        string             `gorm:"not null" json:"subject"`
	Body           string             `gorm:"type:text;not null" json:"body"`
	Status         NotificationStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	ErrorMessage   *string            `gorm:"type:text" json:"error_message,omitempty"`
	SentAt         *time.Time         `json:"sent_at,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}
