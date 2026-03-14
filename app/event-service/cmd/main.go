package main

import (
	"event-service/internal/config"
	"event-service/internal/handler"
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

	router := gin.Default()

	router.POST("/events", eventHandler.Create)
	router.GET("/events/zones/:zoneId", eventHandler.ListByZoneID)
	router.GET("/events/spots/:spotId", eventHandler.ListBySpotID)

	log.Println("event-service started on port", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
