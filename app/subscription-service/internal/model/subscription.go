package model

import "time"

type Subscription struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    int64     `json:"user_id" gorm:"not null;index:idx_user_zone,unique"`
	ZoneID    int64     `json:"zone_id" gorm:"not null;index:idx_user_zone,unique"`
	CreatedAt time.Time `json:"created_at"`
}
