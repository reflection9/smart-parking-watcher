package service

import (
	"context"
	"errors"
	"parking-service/internal/dto"
	"parking-service/internal/model"
	"parking-service/internal/repository"
	"strings"
	"time"

	"gorm.io/gorm"
)

type parkingService struct {
	parkingRepo repository.ParkingRepository
}

func NewParkingService(parkingRepo repository.ParkingRepository) ParkingService {
	return &parkingService{
		parkingRepo: parkingRepo,
	}
}

func (s *parkingService) CreateZone(ctx context.Context, req dto.CreateZoneRequest) (*dto.ZoneResponse, error) {
	existingZone, err := s.parkingRepo.GetZoneByName(ctx, req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existingZone != nil {
		return nil, ErrZoneAlreadyExists
	}

	now := time.Now()
	zone := &model.Zone{
		Name:      req.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.parkingRepo.CreateZone(ctx, zone); err != nil {
		return nil, err
	}

	response := toZoneResponse(*zone)
	return &response, nil
}

func (s *parkingService) ListZones(ctx context.Context) ([]dto.ZoneResponse, error) {
	zones, err := s.parkingRepo.ListZones(ctx)
	if err != nil {
		return nil, err
	}

	response := make([]dto.ZoneResponse, 0, len(zones))
	for _, zone := range zones {
		response = append(response, toZoneResponse(zone))
	}

	return response, nil
}

func (s *parkingService) GetZoneByID(ctx context.Context, id int64) (*dto.ZoneResponse, error) {
	zone, err := s.parkingRepo.GetZoneByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrZoneNotFound
		}
		return nil, err
	}

	response := toZoneResponse(*zone)
	return &response, nil
}

func (s *parkingService) AddSpot(ctx context.Context, zoneID int64, req dto.AddSpotRequest) (*dto.ParkingSpotResponse, error) {
	_, err := s.parkingRepo.GetZoneByID(ctx, zoneID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrZoneNotFound
		}
		return nil, err
	}

	existingSpot, err := s.parkingRepo.GetSpotByZoneAndNumber(ctx, zoneID, req.Number)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existingSpot != nil {
		return nil, ErrSpotAlreadyExists
	}

	now := time.Now()
	spot := &model.ParkingSpot{
		ZoneID:    zoneID,
		Number:    req.Number,
		Status:    model.SpotStatusFree,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.parkingRepo.CreateSpot(ctx, spot); err != nil {
		return nil, err
	}

	response := toSpotResponse(*spot)
	return &response, nil
}

func (s *parkingService) UpdateSpotStatus(ctx context.Context, zoneID, spotID int64, req dto.UpdateSpotStatusRequest) (*dto.ParkingSpotResponse, error) {
	if !model.IsValidSpotStatus(req.Status) {
		return nil, ErrInvalidSpotStatus
	}

	_, err := s.parkingRepo.GetZoneByID(ctx, zoneID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrZoneNotFound
		}
		return nil, err
	}

	spot, err := s.parkingRepo.GetSpotByIDAndZoneID(ctx, spotID, zoneID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSpotNotFound
		}
		return nil, err
	}

	spot.Status = model.SpotStatus(strings.ToUpper(req.Status))
	spot.UpdatedAt = time.Now()

	if err := s.parkingRepo.UpdateSpot(ctx, spot); err != nil {
		return nil, err
	}

	response := toSpotResponse(*spot)
	return &response, nil
}

func toZoneResponse(zone model.Zone) dto.ZoneResponse {
	spots := make([]dto.ParkingSpotResponse, 0, len(zone.Spots))
	availableSpots := 0

	for _, spot := range zone.Spots {
		spots = append(spots, toSpotResponse(spot))
		if spot.Status == model.SpotStatusFree {
			availableSpots++
		}
	}

	return dto.ZoneResponse{
		ID:             zone.ID,
		Name:           zone.Name,
		TotalSpots:     len(zone.Spots),
		AvailableSpots: availableSpots,
		CreatedAt:      zone.CreatedAt,
		UpdatedAt:      zone.UpdatedAt,
		Spots:          spots,
	}
}

func toSpotResponse(spot model.ParkingSpot) dto.ParkingSpotResponse {
	return dto.ParkingSpotResponse{
		ID:        spot.ID,
		ZoneID:    spot.ZoneID,
		Number:    spot.Number,
		Status:    string(spot.Status),
		CreatedAt: spot.CreatedAt,
		UpdatedAt: spot.UpdatedAt,
	}
}
