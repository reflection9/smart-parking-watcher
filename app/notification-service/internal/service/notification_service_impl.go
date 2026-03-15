package service

import (
	"context"
	"errors"
	"fmt"
	"notification-service/internal/dto"
	"notification-service/internal/model"
	"notification-service/internal/repository"
	"strings"
	"time"

	"gorm.io/gorm"
)

type notificationService struct {
	notificationRepo         repository.NotificationRepository
	subscriptionLookupClient SubscriptionLookupClient
	userLookupClient         UserLookupClient
	emailSender              EmailSender
}

func NewNotificationService(
	notificationRepo repository.NotificationRepository,
	subscriptionLookupClient SubscriptionLookupClient,
	userLookupClient UserLookupClient,
	emailSender EmailSender,
) NotificationService {
	return &notificationService{
		notificationRepo:         notificationRepo,
		subscriptionLookupClient: subscriptionLookupClient,
		userLookupClient:         userLookupClient,
		emailSender:              emailSender,
	}
}

func (s *notificationService) DispatchSpotFreed(
	ctx context.Context,
	req dto.DispatchNotificationRequest,
) (*dto.DispatchNotificationResponse, error) {
	if !strings.EqualFold(req.EventType, "spot_freed") {
		return nil, ErrInvalidEventType
	}

	eventType := strings.ToLower(req.EventType)
	userIDs, err := s.subscriptionLookupClient.ListUserIDsByZoneID(ctx, req.ZoneID)
	if err != nil {
		return nil, err
	}

	response := &dto.DispatchNotificationResponse{
		EventID:          req.EventID,
		EventType:        eventType,
		ZoneID:           req.ZoneID,
		SpotID:           req.SpotID,
		TotalSubscribers: len(userIDs),
		Notifications:    make([]dto.NotificationResponse, 0, len(userIDs)),
	}

	for _, userID := range userIDs {
		notification := newPendingNotification(req, eventType, userID)

		created, err := s.notificationRepo.CreatePending(ctx, notification)
		if err != nil {
			return nil, err
		}
		if !created {
			response.DuplicatesSkipped++
			continue
		}

		response.Processed++

		user, err := s.userLookupClient.GetByID(ctx, userID)
		if err != nil {
			if err := s.markFailed(ctx, notification, fmt.Sprintf("failed to fetch user email: %v", err)); err != nil {
				return nil, err
			}

			response.Failed++
			response.Notifications = append(response.Notifications, toNotificationResponse(*notification))
			continue
		}

		if user == nil || strings.TrimSpace(user.Email) == "" {
			if err := s.markFailed(ctx, notification, "user email not found"); err != nil {
				return nil, err
			}

			response.Failed++
			response.Notifications = append(response.Notifications, toNotificationResponse(*notification))
			continue
		}

		notification.RecipientEmail = user.Email

		sendErr := s.emailSender.Send(ctx, EmailMessage{
			To:      user.Email,
			Subject: notification.Subject,
			Body:    notification.Body,
		})
		if sendErr != nil {
			if err := s.markFailed(ctx, notification, fmt.Sprintf("failed to send email: %v", sendErr)); err != nil {
				return nil, err
			}

			response.Failed++
			response.Notifications = append(response.Notifications, toNotificationResponse(*notification))
			continue
		}

		now := time.Now()
		notification.Status = model.NotificationStatusSent
		notification.ErrorMessage = nil
		notification.SentAt = &now
		notification.UpdatedAt = now

		if err := s.notificationRepo.Update(ctx, notification); err != nil {
			return nil, err
		}

		response.Sent++
		response.Notifications = append(response.Notifications, toNotificationResponse(*notification))
	}

	return response, nil
}

func (s *notificationService) List(
	ctx context.Context,
	query dto.ListNotificationsQuery,
) ([]dto.NotificationResponse, error) {
	notifications, err := s.notificationRepo.List(ctx, repository.NotificationFilter{
		UserID:  query.UserID,
		ZoneID:  query.ZoneID,
		EventID: query.EventID,
		Status:  query.Status,
		Limit:   query.Limit,
	})
	if err != nil {
		return nil, err
	}

	response := make([]dto.NotificationResponse, 0, len(notifications))
	for _, notification := range notifications {
		response = append(response, toNotificationResponse(notification))
	}

	return response, nil
}

func (s *notificationService) GetByID(ctx context.Context, id uint) (*dto.NotificationResponse, error) {
	notification, err := s.notificationRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotificationNotFound
		}

		return nil, err
	}

	response := toNotificationResponse(*notification)
	return &response, nil
}

func newPendingNotification(
	req dto.DispatchNotificationRequest,
	eventType string,
	userID int64,
) *model.Notification {
	now := time.Now()
	subject := fmt.Sprintf("Parking spot available in zone #%d", req.ZoneID)
	body := fmt.Sprintf("A parking spot has been freed in zone #%d.", req.ZoneID)
	if req.SpotID != nil {
		body = fmt.Sprintf("%s Spot ID: %d.", body, *req.SpotID)
	}
	body = fmt.Sprintf("%s Event ID: %s.", body, req.EventID)

	return &model.Notification{
		EventID:        req.EventID,
		EventType:      eventType,
		UserID:         userID,
		ZoneID:         req.ZoneID,
		SpotID:         req.SpotID,
		RecipientEmail: "",
		Subject:        subject,
		Body:           body,
		Status:         model.NotificationStatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (s *notificationService) markFailed(
	ctx context.Context,
	notification *model.Notification,
	message string,
) error {
	now := time.Now()
	notification.Status = model.NotificationStatusFailed
	notification.ErrorMessage = &message
	notification.SentAt = nil
	notification.UpdatedAt = now

	return s.notificationRepo.Update(ctx, notification)
}

func toNotificationResponse(notification model.Notification) dto.NotificationResponse {
	return dto.NotificationResponse{
		ID:             notification.ID,
		EventID:        notification.EventID,
		EventType:      notification.EventType,
		UserID:         notification.UserID,
		ZoneID:         notification.ZoneID,
		SpotID:         notification.SpotID,
		RecipientEmail: notification.RecipientEmail,
		Subject:        notification.Subject,
		Body:           notification.Body,
		Status:         notification.Status,
		ErrorMessage:   notification.ErrorMessage,
		SentAt:         notification.SentAt,
		CreatedAt:      notification.CreatedAt,
		UpdatedAt:      notification.UpdatedAt,
	}
}
