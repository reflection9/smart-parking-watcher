package dto

type CreateZoneRequest struct {
	Name string `json:"name" binding:"required"`
}
