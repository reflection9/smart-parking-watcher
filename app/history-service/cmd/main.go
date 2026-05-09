package main

import (
	"context"
	"log"
	"time"

	"history-service/internal/archive"
	"history-service/internal/config"
	"history-service/internal/handler"
	"history-service/internal/messaging"
	"history-service/internal/repository"
	"history-service/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()
	collection := config.NewMongoCollection(cfg)

	eventRepo := repository.NewMongoEventRepository(collection)
	coldStorage := archive.NewNoopColdStorage()
	if cfg.MinIOEndpoint != "" {
		storage, err := archive.NewMinIOColdStorage(
			context.Background(),
			cfg.MinIOEndpoint,
			cfg.MinIOAccessKey,
			cfg.MinIOSecretKey,
			cfg.MinIOBucket,
			cfg.MinIOUseSSL,
		)
		if err != nil {
			log.Fatal("failed to initialize MinIO cold storage:", err)
		}
		coldStorage = storage
		log.Println("history-service MinIO cold storage enabled")
	} else {
		log.Println("history-service MinIO cold storage disabled")
	}
	defer func() {
		if err := coldStorage.Close(); err != nil {
			log.Println("failed to close MinIO cold storage:", err)
		}
	}()

	historyService := service.NewHistoryService(
		eventRepo,
		coldStorage,
		time.Duration(cfg.ArchiveAfterHours)*time.Hour,
		cfg.ArchiveBatchSize,
	)
	historyHandler := handler.NewHistoryHandler(historyService)

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
		log.Println("history-service reservation Kafka consumer enabled for topic", cfg.ReservationKafkaTopic)
		log.Println("history-service spot Kafka consumer enabled for topic", cfg.SpotKafkaTopic)

		go func() {
			if err := reservationConsumer.Start(context.Background(), historyService); err != nil {
				log.Println("reservation event consumer stopped with error:", err)
			}
		}()

		go func() {
			if err := spotConsumer.Start(context.Background(), historyService); err != nil {
				log.Println("spot event consumer stopped with error:", err)
			}
		}()
	} else {
		log.Println("history-service Kafka consumers disabled")
	}
	defer func() {
		if err := reservationConsumer.Close(); err != nil {
			log.Println("failed to close reservation event consumer:", err)
		}
		if err := spotConsumer.Close(); err != nil {
			log.Println("failed to close spot event consumer:", err)
		}
	}()

	if cfg.ArchiveInterval > 0 && cfg.MinIOEndpoint != "" {
		go func() {
			ticker := time.NewTicker(cfg.ArchiveInterval)
			defer ticker.Stop()

			for range ticker.C {
				if _, err := historyService.ArchiveOldEvents(context.Background(), 0); err != nil {
					log.Println("history-service background archive failed:", err)
				}
			}
		}()
	}

	router := gin.Default()

	router.POST("/events", historyHandler.Create)
	router.GET("/events/zones/:zoneId", historyHandler.ListByZoneID)
	router.GET("/events/spots/:spotId", historyHandler.ListBySpotID)
	router.GET("/events/reservations/:reservationId", historyHandler.ListByReservationID)

	router.POST("/history/events", historyHandler.Create)
	router.GET("/history/zones/:zoneId", historyHandler.ListByZoneID)
	router.GET("/history/spots/:spotId", historyHandler.ListBySpotID)
	router.GET("/history/reservations/:reservationId", historyHandler.ListByReservationID)
	router.POST("/history/archive", historyHandler.Archive)

	log.Println("history-service started on port", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
