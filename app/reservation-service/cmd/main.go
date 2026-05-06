package main

import (
	"log"
	"reservation-service/internal/config"
	"reservation-service/internal/handler"
	"reservation-service/internal/repository"
	"reservation-service/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	reservationDB := config.NewReservationDatabase(cfg)
	reservationRepo := repository.NewGormReservationRepository(reservationDB)
	userLookupClient := service.NewHTTPUserLookupClient(cfg.UserServiceURL)
	parkingSpotClient := service.NewHTTPParkingSpotClient(cfg.ParkingServiceURL)

	reservationService := service.NewReservationService(
		reservationRepo,
		userLookupClient,
		parkingSpotClient,
		cfg.ReservationTTL,
	)
	reservationHandler := handler.NewReservationHandler(reservationService)

	router := gin.Default()

	router.POST("/reservations", reservationHandler.Create)
	router.GET("/reservations/:id", reservationHandler.GetByID)
	router.GET("/reservations/users/:userId", reservationHandler.ListByUserID)
	router.POST("/reservations/:id/confirm", reservationHandler.Confirm)
	router.POST("/reservations/:id/cancel", reservationHandler.Cancel)
	router.POST("/reservations/:id/expire", reservationHandler.Expire)

	log.Println("reservation-service started on port", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
