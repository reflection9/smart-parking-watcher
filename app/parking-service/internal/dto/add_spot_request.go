package dto

type AddSpotRequest struct {
	Number string `json:"number" binding:"required"`
}
