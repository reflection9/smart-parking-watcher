package model

import "time"

type Zone struct {
	ID        int64         `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string        `json:"name" gorm:"type:varchar(255);not null;uniqueIndex"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Spots     []ParkingSpot `json:"spots" gorm:"foreignKey:ZoneID"`
}
