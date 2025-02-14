package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

var SecretKey string

func LoadConfig() {

	wd, _ := os.Getwd()
	fmt.Println("📂 Current Working Directory:", wd)

	// Ensure we are loading .env from the correct location
	envPath := filepath.Join(wd, ".env")
	fmt.Println("📂 .env Path:", envPath)
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		log.Fatalf("❌ ERROR: .env file not found at %s", envPath)
	}

	// Load .env file
	err := godotenv.Load(envPath)
	if err != nil {
		log.Fatalf("❌ ERROR: Could not load .env file: %v", err)
	}

	SecretKey = os.Getenv("JWT_SECRET_KEY")
	if SecretKey == "" {
		log.Fatal("JWT_SECRET_KEY is missing in .env")
	}
}
