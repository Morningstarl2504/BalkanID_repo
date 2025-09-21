// morningstarl2504/balkanid_repo/BalkanID_repo-f1fc3ed153144eb6d79e3c90f73a0f3d312b9c79/backend/cmd/server/main.go
package main

import (
	"log"
	"os"

	"filevault-backend/internal/config"
	"filevault-backend/internal/database"
	"filevault-backend/internal/handlers"
	"filevault-backend/internal/middleware"
	"filevault-backend/internal/services"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	if err := database.Connect(cfg.DatabaseURL); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	if err := database.Migrate(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}
	if err := os.MkdirAll(cfg.UploadPath, 0755); err != nil {
		log.Fatal("Failed to create upload directory:", err)
	}

	authService := services.NewAuthService(cfg.JWTSecret)
	storageService := services.NewStorageService(cfg.UploadPath)
	fileService := services.NewFileService(storageService, cfg.MaxFileSize)
	auditService := services.NewAuditService()
	authHandler := handlers.NewAuthHandler(authService)
	fileHandler := handlers.NewFileHandler(fileService, storageService, auditService)
	adminHandler := handlers.NewAdminHandler(fileService, storageService, auditService)
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit, 10)

	router := gin.Default()

	// Apply CORS and Rate Limiter to all routes
	router.Use(middleware.CORS())
	router.Use(rateLimiter.Middleware())

	router.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	api := router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Group for all routes that require authentication
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			protected.GET("/profile", authHandler.GetProfile)
			protected.GET("/storage/stats", fileHandler.GetStorageStats)

			files := protected.Group("/files")
			{
				files.POST("/upload", fileHandler.UploadFile)
				files.GET("", fileHandler.GetUserFiles) // Use "" for the group's root
				files.GET("/:id/download", fileHandler.DownloadFile)
				files.DELETE("/:id", fileHandler.DeleteFile)
			}
		}

		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware(authService), middleware.AdminMiddleware())
		{
			admin.GET("/files", adminHandler.GetAllFiles)
			admin.GET("/stats", adminHandler.GetSystemStats)
			admin.GET("/users", adminHandler.GetUsers)
			admin.GET("/audit-logs", adminHandler.GetAuditLogs)
		}
	}

	log.Printf("Server starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}