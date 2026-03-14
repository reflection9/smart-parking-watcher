package dto

type CreateSubscriptionRequest struct {
	UserID int64 `json:"user_id" binding:"required"`
	ZoneID int64 `json:"zone_id" binding:"required"`
}
