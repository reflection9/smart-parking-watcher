package service

import (
	"context"
	"errors"
	"subscription-service/internal/dto"
	"subscription-service/internal/model"
	"subscription-service/internal/repository"
	"time"

	"gorm.io/gorm"
)

type subscriptionService struct {
	subscriptionRepo repository.SubscriptionRepository
	userLookupClient UserLookupClient
	zoneLookupClient ZoneLookupClient
}

func NewSubscriptionService(
	subscriptionRepo repository.SubscriptionRepository,
	userLookupClient UserLookupClient,
	zoneLookupClient ZoneLookupClient,
) SubscriptionService {
	return &subscriptionService{
		subscriptionRepo: subscriptionRepo,
		userLookupClient: userLookupClient,
		zoneLookupClient: zoneLookupClient,
	}
}

func (s *subscriptionService) Create(ctx context.Context, req dto.CreateSubscriptionRequest) (*dto.SubscriptionResponse, error) {
	userExists, err := s.userLookupClient.Exists(ctx, req.UserID)
	if err != nil {
		return nil, ErrDependencyUnavailable
	}
	if !userExists {
		return nil, ErrUserNotFound
	}

	zoneExists, err := s.zoneLookupClient.Exists(ctx, req.ZoneID)
	if err != nil {
		return nil, ErrDependencyUnavailable
	}
	if !zoneExists {
		return nil, ErrZoneNotFound
	}

	existingSubscription, err := s.subscriptionRepo.GetByUserAndZone(ctx, req.UserID, req.ZoneID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existingSubscription != nil {
		return nil, ErrSubscriptionAlreadyExists
	}

	now := time.Now()

	subscription := &model.Subscription{
		UserID:    req.UserID,
		ZoneID:    req.ZoneID,
		CreatedAt: now,
	}

	if err := s.subscriptionRepo.Create(ctx, subscription); err != nil {
		return nil, err
	}

	return &dto.SubscriptionResponse{
		ID:        subscription.ID,
		UserID:    subscription.UserID,
		ZoneID:    subscription.ZoneID,
		CreatedAt: subscription.CreatedAt,
	}, nil
}

func (s *subscriptionService) ListByUserID(ctx context.Context, userID int64) ([]dto.SubscriptionResponse, error) {
	subscriptions, err := s.subscriptionRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	response := make([]dto.SubscriptionResponse, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		response = append(response, dto.SubscriptionResponse{
			ID:        subscription.ID,
			UserID:    subscription.UserID,
			ZoneID:    subscription.ZoneID,
			CreatedAt: subscription.CreatedAt,
		})
	}

	return response, nil
}

func (s *subscriptionService) ListByZoneID(ctx context.Context, zoneID int64) ([]dto.SubscriptionResponse, error) {
	subscriptions, err := s.subscriptionRepo.ListByZoneID(ctx, zoneID)
	if err != nil {
		return nil, err
	}

	response := make([]dto.SubscriptionResponse, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		response = append(response, dto.SubscriptionResponse{
			ID:        subscription.ID,
			UserID:    subscription.UserID,
			ZoneID:    subscription.ZoneID,
			CreatedAt: subscription.CreatedAt,
		})
	}

	return response, nil
}

func (s *subscriptionService) Delete(ctx context.Context, userID, zoneID int64) error {
	err := s.subscriptionRepo.DeleteByUserAndZone(ctx, userID, zoneID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSubscriptionNotFound
		}

		return err
	}

	return nil
}
