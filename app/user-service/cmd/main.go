package main

import (
	"log"
	"user-service/internal/config"
	"user-service/internal/handler"
	"user-service/internal/repository"
	"user-service/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()
	db := config.NewDatabase(cfg)

	userRepo := repository.NewGormUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	router := gin.Default()

	router.POST("/users/register", userHandler.Register)
	router.POST("/users/login", userHandler.Login)
	router.GET("/users/:id", userHandler.GetByID)

	log.Println("user-service started on port", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
