package repository

import (
	"context"
	"reservation-service/internal/model"

	"gorm.io/gorm"
)

type GormReservationRepository struct {
	db *gorm.DB
}

func NewGormReservationRepository(db *gorm.DB) *GormReservationRepository {
	return &GormReservationRepository{db: db}
}

func (r *GormReservationRepository) Create(ctx context.Context, reservation *model.Reservation) error {
	return r.db.WithContext(ctx).Create(reservation).Error
}

func (r *GormReservationRepository) Update(ctx context.Context, reservation *model.Reservation) error {
	return r.db.WithContext(ctx).Save(reservation).Error
}

func (r *GormReservationRepository) GetByID(ctx context.Context, id uint) (*model.Reservation, error) {
	var reservation model.Reservation
	if err := r.db.WithContext(ctx).First(&reservation, id).Error; err != nil {
		return nil, err
	}

	return &reservation, nil
}

func (r *GormReservationRepository) ListByUserID(ctx context.Context, userID int64) ([]model.Reservation, error) {
	var reservations []model.Reservation
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&reservations).Error
	if err != nil {
		return nil, err
	}

	return reservations, nil
}

func (r *GormReservationRepository) GetActiveBySpot(ctx context.Context, zoneID, spotID int64) (*model.Reservation, error) {
	var reservation model.Reservation
	err := r.db.WithContext(ctx).
		Where("zone_id = ? AND spot_id = ? AND status = ?", zoneID, spotID, model.ReservationStatusActive).
		Order("created_at DESC").
		First(&reservation).Error
	if err != nil {
		return nil, err
	}

	return &reservation, nil
}
