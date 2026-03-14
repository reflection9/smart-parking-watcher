package repository

import (
	"context"
	"parking-service/internal/model"
)

type ParkingRepository interface {
	CreateZone(ctx context.Context, zone *model.Zone) error
	GetZoneByID(ctx context.Context, id int64) (*model.Zone, error)
	GetZoneByName(ctx context.Context, name string) (*model.Zone, error)
	ListZones(ctx context.Context) ([]model.Zone, error)
	CreateSpot(ctx context.Context, spot *model.ParkingSpot) error
	GetSpotByZoneAndNumber(ctx context.Context, zoneID int64, number string) (*model.ParkingSpot, error)
	GetSpotByIDAndZoneID(ctx context.Context, spotID, zoneID int64) (*model.ParkingSpot, error)
	UpdateSpot(ctx context.Context, spot *model.ParkingSpot) error
}
