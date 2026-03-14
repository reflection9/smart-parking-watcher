package dto

import "time"

type SubscriptionResponse struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	ZoneID    int64     `json:"zone_id"`
	CreatedAt time.Time `json:"created_at"`
}
