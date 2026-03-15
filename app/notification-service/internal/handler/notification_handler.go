package handler

import (
	"errors"
	"net/http"
	"notification-service/internal/dto"
	"notification-service/internal/model"
	"notification-service/internal/service"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notificationService service.NotificationService
}

func NewNotificationHandler(notificationService service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

func (h *NotificationHandler) DispatchSpotFreed(c *gin.Context) {
	var req dto.DispatchNotificationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.notificationService.DispatchSpotFreed(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidEventType) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to dispatch notifications"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *NotificationHandler) List(c *gin.Context) {
	query, err := parseListNotificationsQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.notificationService.List(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list notifications"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *NotificationHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification id"})
		return
	}

	response, err := h.notificationService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrNotificationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get notification"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func parseListNotificationsQuery(c *gin.Context) (dto.ListNotificationsQuery, error) {
	var query dto.ListNotificationsQuery

	if userIDParam := strings.TrimSpace(c.Query("user_id")); userIDParam != "" {
		userID, err := strconv.ParseInt(userIDParam, 10, 64)
		if err != nil {
			return dto.ListNotificationsQuery{}, errors.New("invalid user_id")
		}
		query.UserID = &userID
	}

	if zoneIDParam := strings.TrimSpace(c.Query("zone_id")); zoneIDParam != "" {
		zoneID, err := strconv.ParseInt(zoneIDParam, 10, 64)
		if err != nil {
			return dto.ListNotificationsQuery{}, errors.New("invalid zone_id")
		}
		query.ZoneID = &zoneID
	}

	query.EventID = strings.TrimSpace(c.Query("event_id"))

	if statusParam := strings.TrimSpace(c.Query("status")); statusParam != "" {
		status := model.NotificationStatus(statusParam)
		switch status {
		case model.NotificationStatusPending, model.NotificationStatusSent, model.NotificationStatusFailed:
			query.Status = &status
		default:
			return dto.ListNotificationsQuery{}, errors.New("invalid status")
		}
	}

	if limitParam := strings.TrimSpace(c.Query("limit")); limitParam != "" {
		limit, err := strconv.Atoi(limitParam)
		if err != nil {
			return dto.ListNotificationsQuery{}, errors.New("invalid limit")
		}
		query.Limit = limit
	}

	return query, nil
}
