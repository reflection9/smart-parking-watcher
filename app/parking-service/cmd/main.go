package main

import (
	"log"
	"parking-service/internal/config"
	"parking-service/internal/handler"
	"parking-service/internal/repository"
	"parking-service/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()
	db := config.NewDatabase(cfg)

	parkingRepo := repository.NewGormParkingRepository(db)
	parkingService := service.NewParkingService(parkingRepo)
	parkingHandler := handler.NewParkingHandler(parkingService)

	router := gin.Default()

	router.POST("/zones", parkingHandler.CreateZone)
	router.GET("/zones", parkingHandler.ListZones)
	router.GET("/zones/:id", parkingHandler.GetZoneByID)
	router.POST("/zones/:zoneId/spots", parkingHandler.AddSpot)
	router.GET("/zones/:zoneId/spots/:spotId", parkingHandler.GetSpotByID)
	router.PATCH("/zones/:zoneId/spots/:spotId/status", parkingHandler.UpdateSpotStatus)
	router.POST("/zones/:zoneId/spots/:spotId/reserve", parkingHandler.ReserveSpot)
	router.POST("/zones/:zoneId/spots/:spotId/release", parkingHandler.ReleaseSpot)
	router.POST("/zones/:zoneId/spots/:spotId/occupy", parkingHandler.OccupySpot)

	log.Println("parking-service started on port", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
