package dto

type CreateReservationRequest struct {
	UserID int64 `json:"user_id" binding:"required,gt=0"`
	ZoneID int64 `json:"zone_id" binding:"required,gt=0"`
	SpotID int64 `json:"spot_id" binding:"required,gt=0"`
}
