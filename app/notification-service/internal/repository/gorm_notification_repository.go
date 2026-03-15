package repository

import (
	"context"
	"notification-service/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormNotificationRepository struct {
	db *gorm.DB
}

func NewGormNotificationRepository(db *gorm.DB) *GormNotificationRepository {
	return &GormNotificationRepository{db: db}
}

func (r *GormNotificationRepository) CreatePending(ctx context.Context, notification *model.Notification) (bool, error) {
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "event_id"},
				{Name: "user_id"},
			},
			DoNothing: true,
		}).
		Create(notification)

	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected == 1, nil
}

func (r *GormNotificationRepository) Update(ctx context.Context, notification *model.Notification) error {
	return r.db.WithContext(ctx).Save(notification).Error
}

func (r *GormNotificationRepository) List(ctx context.Context, filter NotificationFilter) ([]model.Notification, error) {
	query := r.db.WithContext(ctx).Model(&model.Notification{})

	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.ZoneID != nil {
		query = query.Where("zone_id = ?", *filter.ZoneID)
	}
	if filter.EventID != "" {
		query = query.Where("event_id = ?", filter.EventID)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var notifications []model.Notification
	err := query.
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications).Error
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *GormNotificationRepository) GetByID(ctx context.Context, id uint) (*model.Notification, error) {
	var notification model.Notification

	err := r.db.WithContext(ctx).First(&notification, id).Error
	if err != nil {
		return nil, err
	}

	return &notification, nil
}
