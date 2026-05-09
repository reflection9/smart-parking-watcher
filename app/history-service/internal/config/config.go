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
	AppPort               string
	MongoURI              string
	MongoDBName           string
	MongoCollection       string
	KafkaBrokers          []string
	ReservationKafkaTopic string
	ReservationGroupID    string
	SpotKafkaTopic        string
	SpotGroupID           string
	MinIOEndpoint         string
	MinIOAccessKey        string
	MinIOSecretKey        string
	MinIOBucket           string
	MinIOUseSSL           bool
	ArchiveAfterHours     int
	ArchiveInterval       time.Duration
	ArchiveBatchSize      int64
}

func LoadConfig() *Config {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println(".env file not found, using system env")
	}

	return &Config{
		AppPort:               getEnv("APP_PORT", "8084"),
		MongoURI:              getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName:           getEnv("MONGO_DB_NAME", "history_service_db"),
		MongoCollection:       getEnv("MONGO_COLLECTION", "parking_events"),
		KafkaBrokers:          getEnvAsList("KAFKA_BROKERS"),
		ReservationKafkaTopic: getEnv("KAFKA_RESERVATION_TOPIC", "reservation-lifecycle-events"),
		ReservationGroupID:    getEnv("KAFKA_RESERVATION_GROUP_ID", "history-service-reservations"),
		SpotKafkaTopic:        getEnv("KAFKA_SPOT_TOPIC", "spot-status-events"),
		SpotGroupID:           getEnv("KAFKA_SPOT_GROUP_ID", "history-service-spots"),
		MinIOEndpoint:         getEnv("MINIO_ENDPOINT", ""),
		MinIOAccessKey:        getEnv("MINIO_ACCESS_KEY", ""),
		MinIOSecretKey:        getEnv("MINIO_SECRET_KEY", ""),
		MinIOBucket:           getEnv("MINIO_BUCKET", "history-archives"),
		MinIOUseSSL:           getEnvAsBool("MINIO_USE_SSL", false),
		ArchiveAfterHours:     getEnvAsInt("ARCHIVE_AFTER_HOURS", 720),
		ArchiveInterval:       time.Duration(getEnvAsInt("ARCHIVE_INTERVAL_SECONDS", 300)) * time.Second,
		ArchiveBatchSize:      int64(getEnvAsInt("ARCHIVE_BATCH_SIZE", 500)),
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

func getEnvAsBool(key string, fallback bool) bool {
	value := strings.ToLower(getEnv(key, ""))
	switch value {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return fallback
	}
}
