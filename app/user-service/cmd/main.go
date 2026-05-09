package main

import (
	"context"
	"log"
	"user-service/internal/config"
	"user-service/internal/handler"
	"user-service/internal/repository"
	"user-service/internal/service"
	observability "smart-parking-observability"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()
	db := config.NewDatabase(cfg)

	userRepo := repository.NewGormUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)
	observer, err := observability.NewHTTPObserver(context.Background(), "user-service", cfg.OTLPEndpoint)
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

	router.POST("/users/register", userHandler.Register)
	router.POST("/users/login", userHandler.Login)
	router.GET("/users/:id", userHandler.GetByID)

	log.Println("user-service started on port", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
