package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/classius/server/internal/db"
	"github.com/classius/server/internal/handlers"
	"github.com/classius/server/internal/middleware"
	"github.com/classius/server/internal/services"
	"github.com/spf13/viper"
)

func main() {
	// Load configuration
	loadConfig()

	// Initialize database
	database := db.InitDB()
	defer db.CloseDB(database)

	// Run database migrations
	// Temporarily disabled as tables already exist
	// if err := db.RunMigrations(); err != nil {
	// 	log.Fatalf("Failed to run migrations: %v", err)
	// }

	// Initialize AI Sage service
	aiProvider := services.AIProvider(viper.GetString("ai.provider"))
	if string(aiProvider) == "" {
		aiProvider = services.ProviderOpenAI // Default to OpenAI
	}

	aiConfig := map[string]interface{}{
		"api_key":     viper.GetString("ai.openai.api_key"),
		"model":       viper.GetString("ai.openai.model"),
		"base_url":    viper.GetString("ai.local.base_url"),
		"max_tokens":  viper.GetInt("ai.local.max_tokens"),
		"temperature": 0.7,
	}

	sageService, err := services.NewSageService(aiProvider, aiConfig)
	if err != nil {
		log.Printf("Warning: Failed to initialize Sage service: %v", err)
		log.Println("Sage endpoints will not be available")
	}

	// Initialize Book service
	uploadPath := viper.GetString("storage.upload_path")
	if uploadPath == "" {
		uploadPath = "./uploads"
	}
	maxFileSize := viper.GetInt64("storage.max_file_size")
	if maxFileSize == 0 {
		maxFileSize = 100 * 1024 * 1024 // 100MB default
	}
	
	bookService := services.NewBookService(database, uploadPath, maxFileSize)

	// Initialize router
	router := setupRouter(sageService, bookService)

	// Server configuration
	port := viper.GetString("server.port")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Classius server starting on port %s", port)
		log.Printf("📖 API documentation available at http://localhost:%s/docs", port)
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down server...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server stopped")
}

func loadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Default values
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.name", "classius_dev")
	viper.SetDefault("database.user", "classius")
	viper.SetDefault("database.password", "classius_dev_password")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	
	// AI service defaults
	viper.SetDefault("ai.provider", "openai")
	viper.SetDefault("ai.openai.model", "gpt-4")
	viper.SetDefault("ai.local.base_url", "http://localhost:8000")
	viper.SetDefault("ai.local.model", "classius-sage-7b")
	viper.SetDefault("ai.local.max_tokens", 2048)
	
	// Storage defaults
	viper.SetDefault("storage.upload_path", "./uploads")
	viper.SetDefault("storage.max_file_size", 104857600) // 100MB

	// Read environment variables
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not read config file: %v", err)
		log.Println("Using default configuration and environment variables")
	}
}

func setupRouter(sageService *services.SageService, bookService *services.BookService) *gin.Engine {
	// Set Gin mode
	if viper.GetString("environment") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "classius-server",
			"version": "0.1.0",
			"time":    time.Now().UTC(),
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
			auth.POST("/logout", handlers.Logout)
			auth.POST("/refresh", handlers.RefreshToken)
		}

		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.AuthRequired())
		{
			// User routes
			user := protected.Group("/user")
			{
				// Profile management
				user.GET("/profile", handlers.GetUserProfile)
				user.PUT("/profile", handlers.UpdateUserProfile)
				
				// Preferences and settings
				user.GET("/preferences", handlers.GetUserPreferences)
				user.PUT("/preferences", handlers.UpdateUserPreferences)
				
				// Reading goals
				user.GET("/goals", handlers.GetReadingGoals)
				user.PUT("/goals", handlers.UpdateReadingGoals)
				
				// Account management
				user.POST("/change-password", handlers.ChangePassword)
				user.GET("/stats", handlers.GetAccountStats)
				user.DELETE("/account", handlers.DeleteAccount)
				
				// Legacy endpoints
				user.GET("/progress", handlers.GetUserProgress)
				user.POST("/progress", handlers.SaveUserProgress)
			}

			// Book routes
			bookHandlers := handlers.NewBookHandlers(bookService)
			books := protected.Group("/books")
			{
				books.GET("/", bookHandlers.GetBooks)
				books.POST("/upload", bookHandlers.UploadBook)
				books.GET("/stats", bookHandlers.GetBookStats)
				books.GET("/tags", bookHandlers.GetTags)
				books.POST("/tags", bookHandlers.CreateTag)
				books.DELETE("/tags/:id", bookHandlers.DeleteTag)
				books.GET("/:id", bookHandlers.GetBook)
				books.PUT("/:id", bookHandlers.UpdateBook)
				books.DELETE("/:id", bookHandlers.DeleteBook)
				books.GET("/:id/download", bookHandlers.DownloadBook)
				books.GET("/:id/content", bookHandlers.GetBookContent)
				books.GET("/:id/text", bookHandlers.GetBookText)
			}

			// Annotation routes
			annotations := protected.Group("/annotations")
			{
				// Enhanced annotation management
				annotations.GET("/", handlers.GetAnnotationsAdvanced)
				annotations.POST("/", handlers.CreateAnnotationEnhanced)
				annotations.PUT("/:id", handlers.UpdateAnnotationEnhanced)
				annotations.DELETE("/:id", handlers.DeleteAnnotationEnhanced)
				
				// Bulk operations
				annotations.POST("/bulk", handlers.BulkAnnotationActions)
				
				// Export functionality
				annotations.GET("/export", handlers.ExportAnnotations)
			}

			// Reading Progress routes
			progress := protected.Group("/progress")
			{
				progress.GET("/", handlers.GetReadingProgress)
				progress.POST("/", handlers.UpdateReadingProgress)
				progress.GET("/stats", handlers.GetReadingStats)
			}

			// Reading Sessions routes
			sessions := protected.Group("/sessions")
			{
				sessions.GET("/", handlers.GetReadingSessions)
				sessions.POST("/start/:book_id", handlers.StartReadingSession)
				sessions.PUT("/end/:session_id", handlers.EndReadingSession)
			}

			// Bookmarks routes
			bookmarks := protected.Group("/bookmarks")
			{
				bookmarks.GET("/", handlers.GetBookmarks)
				bookmarks.POST("/", handlers.CreateBookmark)
				bookmarks.PUT("/:id", handlers.UpdateBookmark)
				bookmarks.DELETE("/:id", handlers.DeleteBookmark)
			}

			// Statistics routes
			stats := protected.Group("/stats")
			{
				stats.GET("/books", handlers.GetBookStats)
			}

			// Search routes
			search := protected.Group("/search")
			{
				search.GET("/", handlers.GlobalSearch)
			}

			// AI Sage routes (only if service is available)
			if sageService != nil {
				sageHandlers := handlers.NewSageHandlers(sageService)
				sage := protected.Group("/sage")
				{
					sage.POST("/ask", sageHandlers.AskSage)
					sage.GET("/capabilities", sageHandlers.GetSageCapabilities)
					sage.GET("/health", sageHandlers.CheckSageHealth)
					sage.GET("/conversations", sageHandlers.GetSageConversations)
					sage.GET("/conversations/:id", sageHandlers.GetSageConversation)
					sage.DELETE("/conversations/:id", sageHandlers.DeleteSageConversation)
					sage.GET("/stats", sageHandlers.GetSageStats)
					sage.GET("/export", sageHandlers.ExportSageData)
				}
			}

			// Community routes (placeholder - to be implemented later)
			// community := protected.Group("/community")
			// {
			//	community.GET("/discussions", handlers.GetDiscussions)
			//	community.POST("/discussions", handlers.CreateDiscussion)
			//	community.GET("/discussions/:id", handlers.GetDiscussion)
			//	community.POST("/discussions/:id/comments", handlers.AddComment)
			// }
		}
	}

	// Serve static files (if needed)
	router.Static("/static", "./static")

	return router
}