package model

import "time"

type Reservation struct {
	ID          uint              `gorm:"primaryKey" json:"id"`
	UserID      int64             `gorm:"not null;index" json:"user_id"`
	ZoneID      int64             `gorm:"not null;index" json:"zone_id"`
	SpotID      int64             `gorm:"not null;index" json:"spot_id"`
	Status      ReservationStatus `gorm:"type:varchar(20);not null;default:'ACTIVE';index" json:"status"`
	ExpiresAt   time.Time         `gorm:"not null;index" json:"expires_at"`
	ConfirmedAt *time.Time        `json:"confirmed_at,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}
