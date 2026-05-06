package repository

import (
	"context"
	"errors"
	"parking-service/internal/model"
	"time"

	"gorm.io/gorm"
)

type GormParkingRepository struct {
	db *gorm.DB
}

func NewGormParkingRepository(db *gorm.DB) *GormParkingRepository {
	return &GormParkingRepository{db: db}
}

func (r *GormParkingRepository) CreateZone(ctx context.Context, zone *model.Zone) error {
	return r.db.WithContext(ctx).Create(zone).Error
}

func (r *GormParkingRepository) GetZoneByID(ctx context.Context, id int64) (*model.Zone, error) {
	var zone model.Zone

	err := r.db.WithContext(ctx).
		Preload("Spots", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		First(&zone, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}

	return &zone, nil
}

func (r *GormParkingRepository) GetZoneByName(ctx context.Context, name string) (*model.Zone, error) {
	var zone model.Zone

	err := r.db.WithContext(ctx).
		Where("name = ?", name).
		First(&zone).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}

	return &zone, nil
}

func (r *GormParkingRepository) ListZones(ctx context.Context) ([]model.Zone, error) {
	var zones []model.Zone

	err := r.db.WithContext(ctx).
		Preload("Spots", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Order("id ASC").
		Find(&zones).Error

	if err != nil {
		return nil, err
	}

	return zones, nil
}

func (r *GormParkingRepository) CreateSpot(ctx context.Context, spot *model.ParkingSpot) error {
	return r.db.WithContext(ctx).Create(spot).Error
}

func (r *GormParkingRepository) GetSpotByZoneAndNumber(ctx context.Context, zoneID int64, number string) (*model.ParkingSpot, error) {
	var spot model.ParkingSpot

	err := r.db.WithContext(ctx).
		Where("zone_id = ? AND number = ?", zoneID, number).
		First(&spot).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}

	return &spot, nil
}

func (r *GormParkingRepository) GetSpotByIDAndZoneID(ctx context.Context, spotID, zoneID int64) (*model.ParkingSpot, error) {
	var spot model.ParkingSpot

	err := r.db.WithContext(ctx).
		Where("id = ? AND zone_id = ?", spotID, zoneID).
		First(&spot).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}

	return &spot, nil
}

func (r *GormParkingRepository) UpdateSpot(ctx context.Context, spot *model.ParkingSpot) error {
	return r.db.WithContext(ctx).Save(spot).Error
}

func (r *GormParkingRepository) UpdateSpotStatusIfCurrent(
	ctx context.Context,
	spotID, zoneID int64,
	current []model.SpotStatus,
	next model.SpotStatus,
	updatedAt time.Time,
) (bool, error) {
	result := r.db.WithContext(ctx).
		Model(&model.ParkingSpot{}).
		Where("id = ? AND zone_id = ? AND status IN ?", spotID, zoneID, current).
		Updates(map[string]any{
			"status":     next,
			"updated_at": updatedAt,
		})

	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected == 1, nil
}
