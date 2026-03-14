package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort         string
	MongoURI        string
	MongoDBName     string
	MongoCollection string
}

func LoadConfig() *Config {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println(".env file not found, using system env")
	}

	return &Config{
		AppPort:         os.Getenv("APP_PORT"),
		MongoURI:        os.Getenv("MONGO_URI"),
		MongoDBName:     os.Getenv("MONGO_DB_NAME"),
		MongoCollection: os.Getenv("MONGO_COLLECTION"),
	}
}
