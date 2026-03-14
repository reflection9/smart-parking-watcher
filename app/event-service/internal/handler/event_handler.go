package handler

import (
	"net/http"
	"strconv"

	"event-service/internal/dto"
	"event-service/internal/service"

	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	eventService service.EventService
}

func NewEventHandler(eventService service.EventService) *EventHandler {
	return &EventHandler{
		eventService: eventService,
	}
}

func (h *EventHandler) Create(c *gin.Context) {
	var req dto.CreateEventRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.eventService.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create event"})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *EventHandler) ListByZoneID(c *gin.Context) {
	zoneID, err := strconv.ParseInt(c.Param("zoneId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone id"})
		return
	}

	response, err := h.eventService.ListByZoneID(c.Request.Context(), zoneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get events"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *EventHandler) ListBySpotID(c *gin.Context) {
	spotID, err := strconv.ParseInt(c.Param("spotId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid spot id"})
		return
	}

	response, err := h.eventService.ListBySpotID(c.Request.Context(), spotID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get events"})
		return
	}

	c.JSON(http.StatusOK, response)
}
