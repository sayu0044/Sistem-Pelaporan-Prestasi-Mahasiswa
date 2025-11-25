package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBDriver    string
	Port        string
	JWTSecret   string
	JWTExpiry   string
	Environment string
)

// LoadEnv memuat environment variables dari .env file
func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("File .env tidak ditemukan, menggunakan environment variables sistem")
	}

	// Database configuration
	DBHost = getEnv("DB_HOST", "localhost")
	DBPort = getEnv("DB_PORT", "5432")
	DBUser = getEnv("DB_USER", "postgres")
	DBPassword = getEnv("DB_PASSWORD", "")
	DBName = getEnv("DB_NAME", "backend")
	DBDriver = getEnv("DB_DRIVER", "postgres")

	// Server configuration
	Port = getEnv("PORT", "3000")
	Environment = getEnv("ENVIRONMENT", "development")

	// JWT configuration
	JWTSecret = getEnv("JWT_SECRET", "your-secret-key-change-in-production")
	JWTExpiry = getEnv("JWT_EXPIRY", "24h")
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
