package dto

import (
	"time"
	"user-service/internal/model"
)

type UserResponse struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Role      model.Role `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
