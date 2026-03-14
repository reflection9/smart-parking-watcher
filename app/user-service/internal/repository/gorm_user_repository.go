package repository

import (
	"context"
	"errors"
	"user-service/internal/model"

	"gorm.io/gorm"
)

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *GormUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User

	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *GormUserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User

	err := r.db.WithContext(ctx).
		First(&user, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}

	return &user, nil
}
