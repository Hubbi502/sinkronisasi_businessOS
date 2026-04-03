package main

import (
	"fmt"
	"log"

	"sinkronisasi_db/config"
	"sinkronisasi_db/internal/handler"
	"sinkronisasi_db/internal/middleware"
	"sinkronisasi_db/internal/model"
	"sinkronisasi_db/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to PostgreSQL
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	// Auto-migrate all models
	if err := db.AutoMigrate(model.AllModels()...); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}
	log.Println("Database migration complete for all models")

	// Create Asynq client (for publishing jobs)
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.RedisAddr})
	defer asynqClient.Close()

	// Build entity registry
	registry := service.BuildEntityRegistry()

	// Initialize webhook handler
	webhookHandler := handler.NewWebhookHandler(asynqClient, registry)

	// Setup Gin router
	router := gin.Default()

	// API routes
	api := router.Group("/api")
	{
		// Webhook (with HMAC signature verification)
		webhook := api.Group("/webhook")
		webhook.Use(middleware.VerifySignature(cfg.MaintenanceSyncSecret))
		{
			webhook.POST("/sheet-sync", webhookHandler.HandleSheetSync)
			webhook.POST("/full-sync", webhookHandler.HandleFullSync)
			webhook.POST("/pull-sync", webhookHandler.HandlePullSync)
		}

		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":   "ok",
				"entities": getEntityTypes(registry),
			})
		})
	}

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("API server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEntityTypes(registry map[string]*service.EntityConfig) []string {
	types := make([]string, 0, len(registry))
	for k := range registry {
		types = append(types, k)
	}
	return types
}
