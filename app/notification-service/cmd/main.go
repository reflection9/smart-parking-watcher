package main

import (
	"context"
	"log"
	"notification-service/internal/config"
	"notification-service/internal/handler"
	"notification-service/internal/messaging"
	"notification-service/internal/repository"
	"notification-service/internal/service"
	observability "smart-parking-observability"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	notificationDB := config.NewNotificationDatabase(cfg)

	notificationRepo := repository.NewGormNotificationRepository(notificationDB)
	subscriptionLookupClient := service.NewHTTPSubscriptionLookupClient(cfg.SubscriptionServiceURL)
	userLookupClient := service.NewHTTPUserLookupClient(cfg.UserServiceURL)

	var emailSender service.EmailSender
	if cfg.EmailTransport == "smtp" {
		emailSender = service.NewSMTPEmailSender(
			cfg.SMTPHost,
			cfg.SMTPPort,
			cfg.SMTPUsername,
			cfg.SMTPPassword,
			cfg.EmailFrom,
		)
	} else {
		emailSender = service.NewLogEmailSender(cfg.EmailFrom)
	}

	notificationService := service.NewNotificationService(
		notificationRepo,
		subscriptionLookupClient,
		userLookupClient,
		emailSender,
	)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	observer, err := observability.NewHTTPObserver(context.Background(), "notification-service", cfg.OTLPEndpoint)
	if err != nil {
		log.Fatal("failed to initialize observability:", err)
	}
	defer func() {
		if shutdownErr := observer.Shutdown(context.Background()); shutdownErr != nil {
			log.Println("failed to shut down observability:", shutdownErr)
		}
	}()

	spotConsumer := messaging.NewNoopSpotEventConsumer()
	if len(cfg.KafkaBrokers) > 0 {
		spotConsumer = messaging.NewKafkaSpotEventConsumer(cfg.KafkaBrokers, cfg.KafkaSpotTopic, cfg.KafkaGroupID)
		log.Println("notification-service spot Kafka consumer enabled for topic", cfg.KafkaSpotTopic)

		go func() {
			if err := spotConsumer.Start(context.Background(), notificationService); err != nil {
				log.Println("spot event consumer stopped with error:", err)
			}
		}()
	} else {
		log.Println("notification-service spot Kafka consumer disabled")
	}
	defer func() {
		if err := spotConsumer.Close(); err != nil {
			log.Println("failed to close spot event consumer:", err)
		}
	}()

	router := gin.Default()
	observer.Attach(router)

	router.POST("/notifications/spot-freed", notificationHandler.DispatchSpotFreed)
	router.GET("/notifications", notificationHandler.List)
	router.GET("/notifications/:id", notificationHandler.GetByID)

	log.Println("notification-service started on port", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
