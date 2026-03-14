package dto

type UpdateSpotStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
