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
	authHandler := handlers.NewAuthHandler(authService)
	fileHandler := handlers.NewFileHandler(fileService, storageService)
	adminHandler := handlers.NewAdminHandler(fileService, storageService)
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit, 10)

	router := gin.Default()

	// Use only ONE CORS middleware
	router.Use(middleware.CORS())
	router.Use(rateLimiter.Middleware())

	router.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	api := router.Group("/api/v1")

	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	public := api.Group("/public")
	{
		public.GET("/files", fileHandler.GetPublicFiles)
		public.GET("/files/:id/download", fileHandler.DownloadFile)
		public.POST("/files/upload", fileHandler.UploadFile)
	}

	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware(authService))
	{
		protected.GET("/profile", authHandler.GetProfile)
		files := protected.Group("/files")
		{
			files.POST("/upload", fileHandler.UploadFile)
			files.GET("/", fileHandler.GetUserFiles)
			files.GET("/:id", fileHandler.GetFile)
			files.GET("/:id/download", fileHandler.DownloadFile)
			files.DELETE("/:id", fileHandler.DeleteFile)
			files.PUT("/:id/share", fileHandler.ShareFile)
		}
		folders := protected.Group("/folders")
		{
			folders.POST("/", fileHandler.CreateFolder)
			folders.GET("/", fileHandler.GetUserFolders)
		}
		protected.GET("/storage/stats", fileHandler.GetStorageStats)
	}

	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware(authService), middleware.AdminMiddleware())
	{
		admin.GET("/files", adminHandler.GetAllFiles)
		admin.GET("/stats", adminHandler.GetSystemStats)
		admin.GET("/users", adminHandler.GetUsers)
		admin.GET("/audit-logs", adminHandler.GetAuditLogs)
	}

	log.Printf("Server starting on port %s", cfg.Port)
	log.Printf("CORS enabled for all origins")
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}