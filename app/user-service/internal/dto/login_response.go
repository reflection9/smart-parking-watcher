package dto

import "user-service/internal/model"

type LoginResponse struct {
	ID    int64      `json:"id"`
	Name  string     `json:"name"`
	Email string     `json:"email"`
	Role  model.Role `json:"role"`
}
