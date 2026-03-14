package dto

import "time"

type ParkingSpotResponse struct {
	ID        int64     `json:"id"`
	ZoneID    int64     `json:"zone_id"`
	Number    string    `json:"number"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
