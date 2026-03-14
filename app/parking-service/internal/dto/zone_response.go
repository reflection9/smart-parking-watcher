package dto

import "time"

type ZoneResponse struct {
	ID             int64                 `json:"id"`
	Name           string                `json:"name"`
	TotalSpots     int                   `json:"total_spots"`
	AvailableSpots int                   `json:"available_spots"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
	Spots          []ParkingSpotResponse `json:"spots"`
}
