package handler

import (
	"errors"
	"net/http"
	"parking-service/internal/dto"
	"parking-service/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ParkingHandler struct {
	parkingService service.ParkingService
}

func NewParkingHandler(parkingService service.ParkingService) *ParkingHandler {
	return &ParkingHandler{
		parkingService: parkingService,
	}
}

func (h *ParkingHandler) CreateZone(c *gin.Context) {
	var req dto.CreateZoneRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.parkingService.CreateZone(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrZoneAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create zone"})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *ParkingHandler) ListZones(c *gin.Context) {
	response, err := h.parkingService.ListZones(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list zones"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ParkingHandler) GetZoneByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone id"})
		return
	}

	response, err := h.parkingService.GetZoneByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrZoneNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get zone"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ParkingHandler) AddSpot(c *gin.Context) {
	zoneID, err := strconv.ParseInt(c.Param("zoneId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone id"})
		return
	}

	var req dto.AddSpotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.parkingService.AddSpot(c.Request.Context(), zoneID, req)
	if err != nil {
		if errors.Is(err, service.ErrZoneNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrSpotAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add spot"})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *ParkingHandler) UpdateSpotStatus(c *gin.Context) {
	zoneID, err := strconv.ParseInt(c.Param("zoneId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone id"})
		return
	}

	spotID, err := strconv.ParseInt(c.Param("spotId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid spot id"})
		return
	}

	var req dto.UpdateSpotStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.parkingService.UpdateSpotStatus(c.Request.Context(), zoneID, spotID, req)
	if err != nil {
		if errors.Is(err, service.ErrSpotNotFound) || errors.Is(err, service.ErrZoneNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrInvalidSpotStatus) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update spot status"})
		return
	}

	c.JSON(http.StatusOK, response)
}
