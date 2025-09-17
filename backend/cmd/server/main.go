package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/Morningstarl2504/BalkanID_repo/backend/internal/config"
	"github.com/Morningstarl2504/BalkanID_repo/backend/internal/database"
	"github.com/Morningstarl2504/BalkanID_repo/backend/internal/handlers"
	"github.com/Morningstarl2504/BalkanID_repo/backend/internal/middleware"
	"github.com/Morningstarl2504/BalkanID_repo/backend/internal/services"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	if err := database.Connect(cfg.DatabaseURL); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := database.Migrate(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Create upload directory
	if err := os.MkdirAll(cfg.UploadPath, 0755); err != nil {
		log.Fatal("Failed to create upload directory:", err)
	}

	// Initialize services
	authService := services.NewAuthService(cfg.JWTSecret)
	storageService := services.NewStorageService(cfg.UploadPath)
	fileService := services.NewFileService(storageService, cfg.MaxFileSize)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	fileHandler := handlers.NewFileHandler(fileService, storageService)
	adminHandler := handlers.NewAdminHandler(fileService, storageService)

	// Initialize middleware
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit, 10) // 10 burst

	// Setup router
	router := gin.Default()

	// Global middleware
	router.Use(middleware.CORS())
	router.Use(rateLimiter.Middleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	api := router.Group("/api/v1")

	// Auth routes (no auth required)
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Public routes (no auth required)
	public := api.Group("/public")
	{
		public.GET("/files", fileHandler.GetPublicFiles)
		public.GET("/files/:id/download", fileHandler.DownloadFile) // Public files
	}

	// Protected routes (auth required)
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware(authService))
	{
		// User profile
		protected.GET("/profile", authHandler.GetProfile)

		// File operations
		files := protected.Group("/files")
		{
			files.POST("/upload", fileHandler.UploadFile)
			files.GET("/", fileHandler.GetUserFiles)
			files.GET("/:id", fileHandler.GetFile)
			files.GET("/:id/download", fileHandler.DownloadFile)
			files.DELETE("/:id", fileHandler.DeleteFile)
			files.PUT("/:id/share", fileHandler.ShareFile)
		}

		// Folder operations
		folders := protected.Group("/folders")
		{
			folders.POST("/", fileHandler.CreateFolder)
			folders.GET("/", fileHandler.GetUserFolders)
		}

		// Storage stats
		protected.GET("/storage/stats", fileHandler.GetStorageStats)
	}

	// Admin routes (admin auth required)
	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware(authService))
	admin.Use(middleware.AdminMiddleware())
	{
		admin.GET("/files", adminHandler.GetAllFiles)
		admin.GET("/stats", adminHandler.GetSystemStats)
		admin.GET("/users", adminHandler.GetUsers)
		admin.GET("/audit-logs", adminHandler.GetAuditLogs)
	}

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}