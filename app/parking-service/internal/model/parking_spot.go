package model

import "time"

type ParkingSpot struct {
	ID        int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	ZoneID    int64      `json:"zone_id" gorm:"not null;index:idx_zone_number,unique"`
	Number    string     `json:"number" gorm:"type:varchar(50);not null;index:idx_zone_number,unique"`
	Status    SpotStatus `json:"status" gorm:"type:varchar(50);not null;default:FREE"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
