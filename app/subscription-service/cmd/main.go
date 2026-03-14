package main

import (
	"log"
	"subscription-service/internal/config"
	"subscription-service/internal/handler"
	"subscription-service/internal/repository"
	"subscription-service/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()
	db := config.NewDatabase(cfg)

	subscriptionRepo := repository.NewGormSubscriptionRepository(db)
	userLookupClient := service.NewHTTPUserLookupClient(cfg.UserServiceURL)
	zoneLookupClient := service.NewHTTPZoneLookupClient(cfg.ParkingServiceURL)
	subscriptionService := service.NewSubscriptionService(subscriptionRepo, userLookupClient, zoneLookupClient)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService)

	router := gin.Default()

	router.POST("/subscriptions", subscriptionHandler.Create)
	router.GET("/subscriptions/users/:userId", subscriptionHandler.ListByUserID)
	router.DELETE("/subscriptions/users/:userId/zones/:zoneId", subscriptionHandler.Delete)

	log.Println("subscription-service started on port", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
