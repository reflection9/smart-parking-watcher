package main

import (
	"context"
	"log"
	observability "smart-parking-observability"
	"subscription-service/internal/cache"
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
	notificationDispatchClient := service.NewHTTPNotificationDispatchClient(cfg.NotificationServiceURL)
	subscriptionCache := cache.NewNoopSubscriptionCache()
	if cfg.RedisAddr != "" {
		subscriptionCache = cache.NewRedisSubscriptionCache(
			cfg.RedisAddr,
			cfg.RedisPassword,
			cfg.RedisDB,
			cfg.RedisKeyPrefix,
			cfg.RedisCacheTTL,
		)
		log.Println("subscription-service Redis cache enabled")
	} else {
		log.Println("subscription-service Redis cache disabled")
	}
	defer func() {
		if err := subscriptionCache.Close(); err != nil {
			log.Println("failed to close subscriptions cache:", err)
		}
	}()

	subscriptionService := service.NewSubscriptionService(
		subscriptionRepo,
		userLookupClient,
		zoneLookupClient,
		notificationDispatchClient,
		subscriptionCache,
	)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService)
	observer, err := observability.NewHTTPObserver(context.Background(), "subscription-service", cfg.OTLPEndpoint)
	if err != nil {
		log.Fatal("failed to initialize observability:", err)
	}
	defer func() {
		if shutdownErr := observer.Shutdown(context.Background()); shutdownErr != nil {
			log.Println("failed to shut down observability:", shutdownErr)
		}
	}()

	router := gin.Default()
	observer.Attach(router)

	router.POST("/subscriptions", subscriptionHandler.Create)
	router.GET("/subscriptions/users/:userId", subscriptionHandler.ListByUserID)
	router.GET("/subscriptions/zones/:zoneId", subscriptionHandler.ListByZoneID)
	router.DELETE("/subscriptions/users/:userId/zones/:zoneId", subscriptionHandler.Delete)

	log.Println("subscription-service started on port", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
