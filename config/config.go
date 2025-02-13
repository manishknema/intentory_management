package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var SecretKey string

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	SecretKey = os.Getenv("JWT_SECRET_KEY")
	if SecretKey == "" {
		log.Fatal("JWT_SECRET_KEY is missing in .env")
	}
}
