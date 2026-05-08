package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort               string
	MongoURI              string
	MongoDBName           string
	MongoCollection       string
	KafkaBrokers          []string
	ReservationKafkaTopic string
	ReservationGroupID    string
	SpotKafkaTopic        string
	SpotGroupID           string
}

func LoadConfig() *Config {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println(".env file not found, using system env")
	}

	return &Config{
		AppPort:               getEnv("APP_PORT", "8084"),
		MongoURI:              getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName:           getEnv("MONGO_DB_NAME", "event_service_db"),
		MongoCollection:       getEnv("MONGO_COLLECTION", "parking_events"),
		KafkaBrokers:          getEnvAsList("KAFKA_BROKERS"),
		ReservationKafkaTopic: getEnv("KAFKA_RESERVATION_TOPIC", "reservation-lifecycle-events"),
		ReservationGroupID:    getEnv("KAFKA_RESERVATION_GROUP_ID", "event-service-reservations"),
		SpotKafkaTopic:        getEnv("KAFKA_SPOT_TOPIC", "spot-status-events"),
		SpotGroupID:           getEnv("KAFKA_SPOT_GROUP_ID", "event-service-spots"),
	}
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}

	return value
}

func getEnvAsList(key string) []string {
	value := getEnv(key, "")
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}
