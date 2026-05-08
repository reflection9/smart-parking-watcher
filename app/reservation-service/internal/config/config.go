package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	UserServiceURL           string
	ParkingServiceURL        string
	KafkaBrokers             []string
	KafkaReservationTopic    string
	KafkaParkingCommandTopic string
	KafkaSpotTopic           string
	KafkaSpotGroupID         string

	ReservationTTL time.Duration
}

func LoadConfig() *Config {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println(".env file not found, using system env")
	}

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "reservation_service_db")

	ttlMinutes := getEnvAsInt("RESERVATION_TTL_MINUTES", 5)
	if ttlMinutes <= 0 {
		ttlMinutes = 5
	}

	return &Config{
		AppPort: getEnv("APP_PORT", "8086"),

		DBHost:     dbHost,
		DBPort:     dbPort,
		DBUser:     dbUser,
		DBPassword: dbPassword,
		DBName:     dbName,

		UserServiceURL:           getEnv("USER_SERVICE_URL", "http://localhost:8081"),
		ParkingServiceURL:        getEnv("PARKING_SERVICE_URL", "http://localhost:8083"),
		KafkaBrokers:             getEnvAsList("KAFKA_BROKERS"),
		KafkaReservationTopic:    getEnv("KAFKA_RESERVATION_TOPIC", "reservation-lifecycle-events"),
		KafkaParkingCommandTopic: getEnv("KAFKA_PARKING_COMMAND_TOPIC", "parking-spot-commands"),
		KafkaSpotTopic:           getEnv("KAFKA_SPOT_TOPIC", "spot-status-events"),
		KafkaSpotGroupID:         getEnv("KAFKA_SPOT_GROUP_ID", "reservation-service-spots"),

		ReservationTTL: time.Duration(ttlMinutes) * time.Minute,
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

func getEnvAsInt(key string, fallback int) int {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
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
