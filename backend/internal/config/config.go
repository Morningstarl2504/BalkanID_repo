// morningstarl2504/balkanid_repo/BalkanID_repo-f1fc3ed153144eb6d79e3c90f73a0f3d312b9c79/backend/internal/config/config.go
package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL  string
	JWTSecret    string
	Port         string
	UploadPath   string
	MaxFileSize  int64
	RateLimit    float64
	StorageQuota int64
}

func Load() *Config {
	// Explicitly find and load the .env file from the backend directory
	cwd, err := os.Getwd()
	if err != nil {
		log.Println("Warning: could not get current working directory. Assuming .env is in the same folder.")
	}
	
	// Check if running from project root or from within /backend
	if filepath.Base(cwd) != "backend" {
		envPath := filepath.Join(cwd, "backend", ".env")
		if _, err := os.Stat(envPath); err == nil {
			err = godotenv.Load(envPath)
			if err != nil {
				log.Fatalf("Error loading .env file from path: %s", envPath)
			}
		}
	} else {
		// Already in /backend, load .env directly
		err = godotenv.Load()
		if err != nil {
			log.Println("Warning: .env file not found in current directory.")
		}
	}


	maxFileSize, _ := strconv.ParseInt(getEnv("MAX_FILE_SIZE", "52428800"), 10, 64) // 50MB default
	rateLimit, _ := strconv.ParseFloat(getEnv("RATE_LIMIT", "2"), 64)
	storageQuota, _ := strconv.ParseInt(getEnv("STORAGE_QUOTA", "10485760"), 10, 64) // 10MB default

	return &Config{
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/filevault?sslmode=disable"),
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