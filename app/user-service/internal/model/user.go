package model

import "time"

type User struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name         string    `json:"name" gorm:"type:varchar(255);not null"`
	Email        string    `json:"email" gorm:"type:varchar(255);not null;uniqueIndex"`
	PasswordHash string    `json:"-" gorm:"type:text;not null"`
	Role         Role      `json:"role" gorm:"type:varchar(50);not null;default:USER"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
