package main

import (
	"log"

	"sinkronisasi_db/config"
	"sinkronisasi_db/internal/model"
	"sinkronisasi_db/internal/repository"
	"sinkronisasi_db/internal/service"
	"sinkronisasi_db/internal/worker"

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

	// Initialize Google Sheets service
	sheetsService, err := service.NewSheetsService(
		cfg.GoogleSheetsCredentials,
		cfg.GoogleSheetsID,
	)
	if err != nil {
		log.Fatalf("Failed to initialize Google Sheets service: %v", err)
	}
	log.Println("Google Sheets service initialized")

	// Build entity registry
	registry := service.BuildEntityRegistry()

	// Initialize generic repository
	genericRepo := repository.NewGenericRepository(db)

	// Create Asynq client (for publishing revert jobs from SyncToDB handler)
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.RedisAddr})
	defer asynqClient.Close()

	// Initialize task handlers
	syncToSheetHandler := worker.NewSyncToSheetHandler(genericRepo, sheetsService, registry)
	syncToDBHandler := worker.NewSyncToDBHandler(genericRepo, asynqClient, sheetsService, registry)

	// Create Asynq server (consumer)
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisAddr},
		asynq.Config{
			Concurrency: 5,
			Queues: map[string]int{
				"default": 1,
			},
			RetryDelayFunc: asynq.DefaultRetryDelayFunc,
		},
	)

	// Register task handlers using ServeMux
	mux := asynq.NewServeMux()
	mux.HandleFunc(worker.TaskSyncEntityToSheet, syncToSheetHandler.HandleSyncEntityToSheet)
	mux.HandleFunc(worker.TaskRevertSheetRow, syncToSheetHandler.HandleRevertSheetRow)
	mux.HandleFunc(worker.TaskFullSync, syncToSheetHandler.HandleFullSync)
	mux.HandleFunc(worker.TaskProcessSheetBatch, syncToDBHandler.HandleProcessSheetBatch)
	mux.HandleFunc(worker.TaskPullFromSheets, syncToDBHandler.HandlePullFromSheets)

	log.Printf("Worker started with %d entity types, waiting for tasks...", len(registry))

	if err := srv.Run(mux); err != nil {
		log.Fatalf("Failed to start worker: %v", err)
	}
}
