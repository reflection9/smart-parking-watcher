package handler

import (
	"errors"
	"net/http"
	"strconv"
	"subscription-service/internal/dto"
	"subscription-service/internal/service"

	"github.com/gin-gonic/gin"
)

type SubscriptionHandler struct {
	subscriptionService service.SubscriptionService
}

func NewSubscriptionHandler(subscriptionService service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
	}
}

func (h *SubscriptionHandler) Create(c *gin.Context) {
	var req dto.CreateSubscriptionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.subscriptionService.Create(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) || errors.Is(err, service.ErrZoneNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if errors.Is(err, service.ErrSubscriptionAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		if errors.Is(err, service.ErrDependencyUnavailable) {
			c.JSON(http.StatusBadGateway, gin.H{"error": "dependency service unavailable"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create subscription"})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *SubscriptionHandler) ListByUserID(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	response, err := h.subscriptionService.ListByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get subscriptions"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *SubscriptionHandler) Delete(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	zoneID, err := strconv.ParseInt(c.Param("zoneId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone id"})
		return
	}

	err = h.subscriptionService.Delete(c.Request.Context(), userID, zoneID)
	if err != nil {
		if errors.Is(err, service.ErrSubscriptionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete subscription"})
		return
	}

	c.Status(http.StatusNoContent)
}
