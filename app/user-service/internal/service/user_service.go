package service

import (
	"context"
	"user-service/internal/dto"
)

type UserService interface {
	Register(ctx context.Context, req dto.RegisterUserRequest) (*dto.UserResponse, error)
	Login(ctx context.Context, req dto.LoginUserRequest) (*dto.LoginResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.UserResponse, error)
}
