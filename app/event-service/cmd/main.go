package main

import (
	"context"
	"event-service/internal/config"
	"event-service/internal/handler"
	"event-service/internal/messaging"
	"event-service/internal/repository"
	"event-service/internal/service"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()
	collection := config.NewMongoCollection(cfg)

	eventRepo := repository.NewMongoEventRepository(collection)
	eventService := service.NewEventService(eventRepo)
	eventHandler := handler.NewEventHandler(eventService)

	reservationConsumer := messaging.NewNoopReservationEventConsumer()
	spotConsumer := messaging.NewNoopSpotEventConsumer()
	if len(cfg.KafkaBrokers) > 0 {
		reservationConsumer = messaging.NewKafkaReservationEventConsumer(
			cfg.KafkaBrokers,
			cfg.ReservationKafkaTopic,
			cfg.ReservationGroupID,
		)
		spotConsumer = messaging.NewKafkaSpotEventConsumer(
			cfg.KafkaBrokers,
			cfg.SpotKafkaTopic,
			cfg.SpotGroupID,
		)
		log.Println("event-service reservation Kafka consumer enabled for topic", cfg.ReservationKafkaTopic)
		log.Println("event-service spot Kafka consumer enabled for topic", cfg.SpotKafkaTopic)

		go func() {
			if err := reservationConsumer.Start(context.Background(), eventService); err != nil {
				log.Println("reservation event consumer stopped with error:", err)
			}
		}()

		go func() {
			if err := spotConsumer.Start(context.Background(), eventService); err != nil {
				log.Println("spot event consumer stopped with error:", err)
			}
		}()
	} else {
		log.Println("event-service Kafka consumers disabled")
	}
	defer func() {
		if err := reservationConsumer.Close(); err != nil {
			log.Println("failed to close reservation event consumer:", err)
		}
		if err := spotConsumer.Close(); err != nil {
			log.Println("failed to close spot event consumer:", err)
		}
	}()

	router := gin.Default()

	router.POST("/events", eventHandler.Create)
	router.GET("/events/zones/:zoneId", eventHandler.ListByZoneID)
	router.GET("/events/spots/:spotId", eventHandler.ListBySpotID)
	router.GET("/events/reservations/:reservationId", eventHandler.ListByReservationID)

	log.Println("event-service started on port", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
