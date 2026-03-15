package repository

import (
	"context"
	"errors"
	"subscription-service/internal/model"

	"gorm.io/gorm"
)

type GormSubscriptionRepository struct {
	db *gorm.DB
}

func NewGormSubscriptionRepository(db *gorm.DB) *GormSubscriptionRepository {
	return &GormSubscriptionRepository{db: db}
}

func (r *GormSubscriptionRepository) Create(ctx context.Context, subscription *model.Subscription) error {
	return r.db.WithContext(ctx).Create(subscription).Error
}

func (r *GormSubscriptionRepository) GetByUserAndZone(ctx context.Context, userID, zoneID int64) (*model.Subscription, error) {
	var subscription model.Subscription

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND zone_id = ?", userID, zoneID).
		First(&subscription).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}

	return &subscription, nil
}

func (r *GormSubscriptionRepository) ListByUserID(ctx context.Context, userID int64) ([]model.Subscription, error) {
	var subscriptions []model.Subscription

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&subscriptions).Error

	if err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func (r *GormSubscriptionRepository) ListByZoneID(ctx context.Context, zoneID int64) ([]model.Subscription, error) {
	var subscriptions []model.Subscription

	err := r.db.WithContext(ctx).
		Where("zone_id = ?", zoneID).
		Order("created_at DESC").
		Find(&subscriptions).Error

	if err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func (r *GormSubscriptionRepository) DeleteByUserAndZone(ctx context.Context, userID, zoneID int64) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND zone_id = ?", userID, zoneID).
		Delete(&model.Subscription{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
