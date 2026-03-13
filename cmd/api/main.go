package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ifinu/radio-api/internal/cache"
	"github.com/ifinu/radio-api/internal/config"
	"github.com/ifinu/radio-api/internal/handlers"
	"github.com/ifinu/radio-api/internal/models"
	"github.com/ifinu/radio-api/internal/repository"
	"github.com/ifinu/radio-api/internal/services"
	"github.com/ifinu/radio-api/internal/stream"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load .env file
	godotenv.Load()

	// Setup logger
	setupLogger()

	log.Info().Msg("Starting IFINU Radio API")

	// Load config
	cfg := config.Load()

	// Connect to database
	db, err := connectDB(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// Auto-migrate models
	if err := db.AutoMigrate(&models.Radio{}); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}

	log.Info().Msg("Database connected and migrated")

	// Initialize components
	radioCache := cache.NewRadioCache(5 * time.Minute)
	radioRepo := repository.NewRadioRepository(db)
	radioService := services.NewRadioService(radioRepo, radioCache)
	streamer := stream.NewStreamer(cfg)

	// Start periodic sync
	radioService.StartPeriodicSync(cfg.SyncInterval)

	// Initialize handlers
	radioHandler := handlers.NewRadioHandler(radioService, streamer)

	// Setup router
	router := setupRouter(radioHandler)

	// Start server
	addr := ":" + cfg.ServerPort
	log.Info().Str("addr", addr).Msg("Starting HTTP server")

	if err := router.Run(addr); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}

func setupLogger() {
	// Pretty console logging for development
	if os.Getenv("ENV") != "production" {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		})
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}

func connectDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func setupRouter(radioHandler *handlers.RadioHandler) *gin.Engine {
	// Set Gin to release mode in production
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// CORS middleware
	router.Use(corsMiddleware())

	// Health check
	router.GET("/health", radioHandler.Health)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		radios := v1.Group("/radios")
		{
			radios.GET("", radioHandler.ListRadios)
			radios.GET("/search", radioHandler.SearchRadios)
			radios.GET("/:id", radioHandler.GetRadio)
			radios.GET("/:id/stream", radioHandler.StreamRadio)
		}

		// Admin routes (protected)
		admin := v1.Group("/admin")
		admin.Use(adminAuthMiddleware())
		{
			admin.POST("/sync", radioHandler.SyncRadios)
		}
	}

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow specific origins in production
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
		if allowedOrigins == "" {
			allowedOrigins = "*" // Fallback for development
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigins)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Admin authentication middleware
func adminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := os.Getenv("ADMIN_API_KEY")

		// If no API key is set, deny all admin requests
		if apiKey == "" {
			log.Warn().Msg("Admin endpoint accessed but ADMIN_API_KEY not configured")
			c.JSON(401, gin.H{
				"sucesso": false,
				"mensagem": "Admin API key not configured",
			})
			c.Abort()
			return
		}

		// Check X-API-Key header
		providedKey := c.GetHeader("X-API-Key")
		if providedKey == "" {
			c.JSON(401, gin.H{
				"sucesso": false,
				"mensagem": "Missing API key",
			})
			c.Abort()
			return
		}

		// Constant-time comparison to prevent timing attacks
		if providedKey != apiKey {
			log.Warn().
				Str("ip", c.ClientIP()).
				Msg("Failed admin authentication attempt")
			c.JSON(403, gin.H{
				"sucesso": false,
				"mensagem": "Invalid API key",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
