package service

import (
	"context"
	"parking-service/internal/dto"
	"parking-service/internal/messaging"
)

type ParkingService interface {
	CreateZone(ctx context.Context, req dto.CreateZoneRequest) (*dto.ZoneResponse, error)
	ListZones(ctx context.Context) ([]dto.ZoneResponse, error)
	GetZoneByID(ctx context.Context, id int64) (*dto.ZoneResponse, error)
	AddSpot(ctx context.Context, zoneID int64, req dto.AddSpotRequest) (*dto.ParkingSpotResponse, error)
	GetSpotByID(ctx context.Context, zoneID, spotID int64) (*dto.ParkingSpotResponse, error)
	UpdateSpotStatus(ctx context.Context, zoneID, spotID int64, req dto.UpdateSpotStatusRequest) (*dto.ParkingSpotResponse, error)
	ReserveSpot(ctx context.Context, zoneID, spotID int64) (*dto.ParkingSpotResponse, error)
	ReleaseSpot(ctx context.Context, zoneID, spotID int64) (*dto.ParkingSpotResponse, error)
	OccupySpot(ctx context.Context, zoneID, spotID int64) (*dto.ParkingSpotResponse, error)
	HandleSpotCommand(ctx context.Context, command messaging.SpotCommand) error
}
