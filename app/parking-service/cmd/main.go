package main

import (
	"context"
	"log"
	"parking-service/internal/config"
	"parking-service/internal/handler"
	"parking-service/internal/messaging"
	"parking-service/internal/repository"
	"parking-service/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()
	db := config.NewDatabase(cfg)

	parkingRepo := repository.NewGormParkingRepository(db)
	eventPublisher := messaging.NewNoopSpotEventPublisher()
	commandConsumer := messaging.NewNoopParkingCommandConsumer()
	if len(cfg.KafkaBrokers) > 0 {
		eventPublisher = messaging.NewKafkaSpotEventPublisher(cfg.KafkaBrokers, cfg.KafkaSpotTopic)
		commandConsumer = messaging.NewKafkaParkingCommandConsumer(
			cfg.KafkaBrokers,
			cfg.KafkaParkingCommandTopic,
			cfg.KafkaParkingCommandGroupID,
		)
		log.Println("parking-service Kafka orchestration enabled")
	} else {
		log.Println("parking-service Kafka orchestration disabled")
	}
	defer func() {
		if err := eventPublisher.Close(); err != nil {
			log.Println("failed to close spot event publisher:", err)
		}
	}()
	defer func() {
		if err := commandConsumer.Close(); err != nil {
			log.Println("failed to close parking command consumer:", err)
		}
	}()

	parkingService := service.NewParkingService(parkingRepo, eventPublisher)
	parkingHandler := handler.NewParkingHandler(parkingService)

	go func() {
		if err := commandConsumer.Start(context.Background(), parkingService); err != nil {
			log.Println("parking-service command consumer stopped:", err)
		}
	}()

	router := gin.Default()

	router.POST("/zones", parkingHandler.CreateZone)
	router.GET("/zones", parkingHandler.ListZones)
	router.GET("/zones/:id", parkingHandler.GetZoneByID)
	router.POST("/zones/:zoneId/spots", parkingHandler.AddSpot)
	router.GET("/spots/:spotId/zones/:zoneId", parkingHandler.GetSpotByID)
	router.PATCH("/zones/:zoneId/spots/:spotId/status", parkingHandler.UpdateSpotStatus)
	router.POST("/zones/:zoneId/spots/:spotId/reserve", parkingHandler.ReserveSpot)
	router.POST("/zones/:zoneId/spots/:spotId/release", parkingHandler.ReleaseSpot)
	router.POST("/zones/:zoneId/spots/:spotId/occupy", parkingHandler.OccupySpot)

	log.Println("parking-service started on port", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
