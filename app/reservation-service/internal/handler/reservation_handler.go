package handler

import (
	"errors"
	"net/http"
	"reservation-service/internal/dto"
	"reservation-service/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ReservationHandler struct {
	reservationService service.ReservationService
}

func NewReservationHandler(reservationService service.ReservationService) *ReservationHandler {
	return &ReservationHandler{
		reservationService: reservationService,
	}
}

func (h *ReservationHandler) Create(c *gin.Context) {
	var req dto.CreateReservationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.reservationService.Create(c.Request.Context(), req)
	if err != nil {
		h.writeError(c, err, "failed to create reservation")
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *ReservationHandler) GetByID(c *gin.Context) {
	id, err := parseReservationID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reservation id"})
		return
	}

	response, err := h.reservationService.GetByID(c.Request.Context(), id)
	if err != nil {
		h.writeError(c, err, "failed to get reservation")
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ReservationHandler) ListByUserID(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	response, err := h.reservationService.ListByUserID(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err, "failed to list reservations")
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ReservationHandler) Confirm(c *gin.Context) {
	id, err := parseReservationID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reservation id"})
		return
	}

	response, err := h.reservationService.Confirm(c.Request.Context(), id)
	if err != nil {
		h.writeError(c, err, "failed to confirm reservation")
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ReservationHandler) Cancel(c *gin.Context) {
	id, err := parseReservationID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reservation id"})
		return
	}

	response, err := h.reservationService.Cancel(c.Request.Context(), id)
	if err != nil {
		h.writeError(c, err, "failed to cancel reservation")
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ReservationHandler) Expire(c *gin.Context) {
	id, err := parseReservationID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reservation id"})
		return
	}

	response, err := h.reservationService.Expire(c.Request.Context(), id)
	if err != nil {
		h.writeError(c, err, "failed to expire reservation")
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ReservationHandler) writeError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, service.ErrUserNotFound), errors.Is(err, service.ErrSpotNotFound), errors.Is(err, service.ErrReservationNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrSpotUnavailable), errors.Is(err, service.ErrReservationNotActive), errors.Is(err, service.ErrReservationExpired), errors.Is(err, service.ErrActiveReservationExists), errors.Is(err, service.ErrReservationActionNotAllowed):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrDependencyUnavailable):
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": fallback})
	}
}

func parseReservationID(raw string) (uint, error) {
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint(id), nil
}
