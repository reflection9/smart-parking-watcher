package dto

import (
	"reservation-service/internal/model"
	"time"
)

type ReservationResponse struct {
	ID          uint                    `json:"id"`
	UserID      int64                   `json:"user_id"`
	ZoneID      int64                   `json:"zone_id"`
	SpotID      int64                   `json:"spot_id"`
	Status      model.ReservationStatus `json:"status"`
	ExpiresAt   time.Time               `json:"expires_at"`
	ConfirmedAt *time.Time              `json:"confirmed_at,omitempty"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
}
