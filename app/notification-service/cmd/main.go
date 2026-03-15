package main

import (
	"log"
	"notification-service/internal/config"
	"notification-service/internal/handler"
	"notification-service/internal/repository"
	"notification-service/internal/service"

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

	router := gin.Default()

	router.POST("/notifications/spot-freed", notificationHandler.DispatchSpotFreed)
	router.GET("/notifications", notificationHandler.List)
	router.GET("/notifications/:id", notificationHandler.GetByID)

	log.Println("notification-service started on port", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
