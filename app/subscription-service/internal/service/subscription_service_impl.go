package service

import (
	"context"
	"errors"
	"log"
	"subscription-service/internal/cache"
	"subscription-service/internal/dto"
	"subscription-service/internal/model"
	"subscription-service/internal/repository"
	"time"

	"gorm.io/gorm"
)

type subscriptionService struct {
	subscriptionRepo           repository.SubscriptionRepository
	userLookupClient           UserLookupClient
	zoneLookupClient           ZoneLookupClient
	notificationDispatchClient NotificationDispatchClient
	subscriptionCache          cache.SubscriptionCache
}

func NewSubscriptionService(
	subscriptionRepo repository.SubscriptionRepository,
	userLookupClient UserLookupClient,
	zoneLookupClient ZoneLookupClient,
	notificationDispatchClient NotificationDispatchClient,
	subscriptionCache cache.SubscriptionCache,
) SubscriptionService {
	if subscriptionCache == nil {
		subscriptionCache = cache.NewNoopSubscriptionCache()
	}

	if notificationDispatchClient == nil {
		notificationDispatchClient = noopNotificationDispatchClient{}
	}

	return &subscriptionService{
		subscriptionRepo:           subscriptionRepo,
		userLookupClient:           userLookupClient,
		zoneLookupClient:           zoneLookupClient,
		notificationDispatchClient: notificationDispatchClient,
		subscriptionCache:          subscriptionCache,
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

	zone, err := s.zoneLookupClient.GetByID(ctx, req.ZoneID)
	if err != nil {
		return nil, ErrDependencyUnavailable
	}
	if zone == nil {
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

	s.invalidateZoneCache(ctx, subscription.ZoneID)
	s.invalidateUserCache(ctx, subscription.UserID)
	s.dispatchCurrentAvailabilityNotification(ctx, subscription.UserID, zone)

	return &dto.SubscriptionResponse{
		ID:        subscription.ID,
		UserID:    subscription.UserID,
		ZoneID:    subscription.ZoneID,
		CreatedAt: subscription.CreatedAt,
	}, nil
}

func (s *subscriptionService) dispatchCurrentAvailabilityNotification(
	ctx context.Context,
	userID int64,
	zone *ZoneDetails,
) {
	if zone == nil || zone.AvailableSpots <= 0 {
		return
	}

	if err := s.notificationDispatchClient.NotifyCurrentAvailability(ctx, userID, zone); err != nil {
		log.Printf(
			"failed to send immediate availability notification for user %d and zone %d: %v",
			userID,
			zone.ID,
			err,
		)
	}
}

func (s *subscriptionService) ListByUserID(ctx context.Context, userID int64) ([]dto.SubscriptionResponse, error) {
	cachedSubscriptions, found, err := s.subscriptionCache.GetByUserID(ctx, userID)
	if err != nil {
		log.Printf("failed to read user subscriptions cache for user %d: %v", userID, err)
	} else if found {
		return cachedSubscriptions, nil
	}

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

	if err := s.subscriptionCache.SetByUserID(ctx, userID, response); err != nil {
		log.Printf("failed to write user subscriptions cache for user %d: %v", userID, err)
	}

	return response, nil
}

func (s *subscriptionService) ListByZoneID(ctx context.Context, zoneID int64) ([]dto.SubscriptionResponse, error) {
	cachedSubscriptions, found, err := s.subscriptionCache.GetByZoneID(ctx, zoneID)
	if err != nil {
		log.Printf("failed to read zone subscribers cache for zone %d: %v", zoneID, err)
	} else if found {
		return cachedSubscriptions, nil
	}

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

	if err := s.subscriptionCache.SetByZoneID(ctx, zoneID, response); err != nil {
		log.Printf("failed to write zone subscribers cache for zone %d: %v", zoneID, err)
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

	s.invalidateZoneCache(ctx, zoneID)
	s.invalidateUserCache(ctx, userID)

	return nil
}

func (s *subscriptionService) invalidateZoneCache(ctx context.Context, zoneID int64) {
	if err := s.subscriptionCache.InvalidateZoneID(ctx, zoneID); err != nil {
		log.Printf("failed to invalidate zone subscribers cache for zone %d: %v", zoneID, err)
	}
}

func (s *subscriptionService) invalidateUserCache(ctx context.Context, userID int64) {
	if err := s.subscriptionCache.InvalidateUserID(ctx, userID); err != nil {
		log.Printf("failed to invalidate user subscriptions cache for user %d: %v", userID, err)
	}
}
