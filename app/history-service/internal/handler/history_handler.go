package handler

import (
	"net/http"
	"strconv"
	"time"

	"history-service/internal/dto"
	"history-service/internal/service"

	"github.com/gin-gonic/gin"
)

type HistoryHandler struct {
	historyService service.HistoryService
}

func NewHistoryHandler(historyService service.HistoryService) *HistoryHandler {
	return &HistoryHandler{
		historyService: historyService,
	}
}

func (h *HistoryHandler) Create(c *gin.Context) {
	var req dto.CreateEventRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.historyService.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create event"})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *HistoryHandler) ListByZoneID(c *gin.Context) {
	zoneID, err := strconv.ParseInt(c.Param("zoneId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone id"})
		return
	}

	response, err := h.historyService.ListByZoneID(c.Request.Context(), zoneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get events"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *HistoryHandler) ListBySpotID(c *gin.Context) {
	spotID, err := strconv.ParseInt(c.Param("spotId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid spot id"})
		return
	}

	response, err := h.historyService.ListBySpotID(c.Request.Context(), spotID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get events"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *HistoryHandler) ListByReservationID(c *gin.Context) {
	reservationID, err := strconv.ParseInt(c.Param("reservationId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reservation id"})
		return
	}

	response, err := h.historyService.ListByReservationID(c.Request.Context(), reservationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get events"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *HistoryHandler) Archive(c *gin.Context) {
	olderThan := time.Duration(0)
	if raw := c.Query("older_than_hours"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid older_than_hours"})
			return
		}
		if parsed == 0 {
			olderThan = time.Nanosecond
		} else {
			olderThan = time.Duration(parsed) * time.Hour
		}
	}

	response, err := h.historyService.ArchiveOldEvents(c.Request.Context(), olderThan)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to archive history"})
		return
	}

	c.JSON(http.StatusOK, response)
}
