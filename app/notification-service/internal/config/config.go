package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	SubscriptionServiceURL string
	UserServiceURL         string

	EmailTransport string
	EmailFrom      string
	SMTPHost       string
	SMTPPort       string
	SMTPUsername   string
	SMTPPassword   string
	KafkaBrokers   []string
	KafkaSpotTopic string
	KafkaGroupID   string
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
	dbName := getEnv("DB_NAME", "notification_service_db")

	return &Config{
		AppPort: getEnv("APP_PORT", "8085"),

		DBHost:     dbHost,
		DBPort:     dbPort,
		DBUser:     dbUser,
		DBPassword: dbPassword,
		DBName:     dbName,

		SubscriptionServiceURL: getEnv("SUBSCRIPTION_SERVICE_URL", "http://localhost:8082"),
		UserServiceURL:         getEnv("USER_SERVICE_URL", "http://localhost:8081"),

		EmailTransport: strings.ToLower(getEnv("EMAIL_TRANSPORT", "log")),
		EmailFrom:      getEnv("EMAIL_FROM", "noreply@smartparking.local"),
		SMTPHost:       getEnv("SMTP_HOST", "localhost"),
		SMTPPort:       getEnv("SMTP_PORT", "587"),
		SMTPUsername:   getEnv("SMTP_USERNAME", ""),
		SMTPPassword:   getEnv("SMTP_PASSWORD", ""),
		KafkaBrokers:   getEnvAsList("KAFKA_BROKERS"),
		KafkaSpotTopic: getEnv("KAFKA_SPOT_TOPIC", "spot-status-events"),
		KafkaGroupID:   getEnv("KAFKA_GROUP_ID", "notification-service-spots"),
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
