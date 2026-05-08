package main

import (
	"context"
	"log"
	"reservation-service/internal/config"
	"reservation-service/internal/handler"
	"reservation-service/internal/messaging"
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
	eventPublisher := messaging.NewNoopReservationEventPublisher()
	parkingCommandPublisher := messaging.NewNoopParkingCommandPublisher()
	spotEventConsumer := messaging.NewNoopSpotEventConsumer()
	if len(cfg.KafkaBrokers) > 0 {
		eventPublisher = messaging.NewKafkaReservationEventPublisher(cfg.KafkaBrokers, cfg.KafkaReservationTopic)
		parkingCommandPublisher = messaging.NewKafkaParkingCommandPublisher(cfg.KafkaBrokers, cfg.KafkaParkingCommandTopic)
		spotEventConsumer = messaging.NewKafkaSpotEventConsumer(cfg.KafkaBrokers, cfg.KafkaSpotTopic, cfg.KafkaSpotGroupID)
		log.Println("reservation-service Kafka orchestration enabled")
	} else {
		log.Println("reservation-service Kafka orchestration disabled")
	}
	defer func() {
		if err := eventPublisher.Close(); err != nil {
			log.Println("failed to close reservation event publisher:", err)
		}
	}()
	defer func() {
		if err := parkingCommandPublisher.Close(); err != nil {
			log.Println("failed to close parking command publisher:", err)
		}
	}()
	defer func() {
		if err := spotEventConsumer.Close(); err != nil {
			log.Println("failed to close reservation spot event consumer:", err)
		}
	}()

	reservationService := service.NewReservationService(
		reservationRepo,
		userLookupClient,
		parkingSpotClient,
		cfg.ReservationTTL,
		eventPublisher,
		parkingCommandPublisher,
	)
	reservationHandler := handler.NewReservationHandler(reservationService)

	go func() {
		if err := spotEventConsumer.Start(context.Background(), reservationService); err != nil {
			log.Println("reservation-service spot event consumer stopped:", err)
		}
	}()

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
