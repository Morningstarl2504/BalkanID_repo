package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL    string
	JWTSecret     string
	Port          string
	UploadPath    string
	MaxFileSize   int64
	RateLimit     float64
	StorageQuota  int64
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	maxFileSize, _ := strconv.ParseInt(getEnv("MAX_FILE_SIZE", "52428800"), 10, 64) // 50MB default
	rateLimit, _ := strconv.ParseFloat(getEnv("RATE_LIMIT", "2"), 64)
	storageQuota, _ := strconv.ParseInt(getEnv("STORAGE_QUOTA", "10485760"), 10, 64) // 10MB default

	return &Config{
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/filevault?sslmode=disable"),
		JWTSecret:    getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		Port:         getEnv("PORT", "8080"),
		UploadPath:   getEnv("UPLOAD_PATH", "./uploads"),
		MaxFileSize:  maxFileSize,
		RateLimit:    rateLimit,
		StorageQuota: storageQuota,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
