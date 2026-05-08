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

	eventConsumer := messaging.NewNoopReservationEventConsumer()
	if len(cfg.KafkaBrokers) > 0 {
		eventConsumer = messaging.NewKafkaReservationEventConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID)
		log.Println("event-service Kafka consumer enabled for topic", cfg.KafkaTopic)

		go func() {
			if err := eventConsumer.Start(context.Background(), eventService); err != nil {
				log.Println("reservation event consumer stopped with error:", err)
			}
		}()
	} else {
		log.Println("event-service Kafka consumer disabled")
	}
	defer func() {
		if err := eventConsumer.Close(); err != nil {
			log.Println("failed to close reservation event consumer:", err)
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
