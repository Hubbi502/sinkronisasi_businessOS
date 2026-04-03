package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"sinkronisasi_db/internal/repository"
	"sinkronisasi_db/internal/service"

	"github.com/hibiken/asynq"
)

// SyncToDBHandler processes TaskProcessSheetBatch and TaskPullFromSheets jobs.
// It validates incoming sheet data, upserts into DB, and handles deletion reconciliation.
type SyncToDBHandler struct {
	repo          *repository.GenericRepository
	client        *asynq.Client
	sheetsService *service.SheetsService
	registry      map[string]*service.EntityConfig
}

// NewSyncToDBHandler creates a new handler for DB sync tasks.
func NewSyncToDBHandler(
	repo *repository.GenericRepository,
	client *asynq.Client,
	sheetsService *service.SheetsService,
	registry map[string]*service.EntityConfig,
) *SyncToDBHandler {
	return &SyncToDBHandler{
		repo:          repo,
		client:        client,
		sheetsService: sheetsService,
		registry:      registry,
	}
}

// HandleProcessSheetBatch processes a batch of items from Google Apps Script (push).
func (h *SyncToDBHandler) HandleProcessSheetBatch(ctx context.Context, t *asynq.Task) error {
	var payload ProcessSheetBatchPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal ProcessSheetBatchPayload: %w", err)
	}

	cfg, ok := h.registry[payload.EntityType]
	if !ok {
		return fmt.Errorf("unknown entity type: %s", payload.EntityType)
	}

	log.Printf("[SyncToDB] Processing batch of %d items for entity type: %s", len(payload.Items), payload.EntityType)

	var processedCount, skippedCount int

	for i, item := range payload.Items {
		id, ok := item["ID"]
		if !ok || id == "" {
			log.Printf("[SyncToDB] Item %d has no ID — SKIPPING", i)
			skippedCount++
			continue
		}

		// Normalize headers to UPPERCASE (matching 11gawe pattern)
		normalizedItem := make(map[string]string)
		for k, v := range item {
			normalizedItem[service.NormalizeHeader(k)] = v
		}

		entityID, updates, createEntity := cfg.ToUpdates(normalizedItem)
		if entityID == "" {
			log.Printf("[SyncToDB] Item %d has empty entity ID — SKIPPING", i)
			skippedCount++
			continue
		}

		// Simple upsert without OCC — Sheets always wins (matching 11gawe behavior)
		err := h.repo.UpsertSimple(cfg.NewModel(), entityID, updates, createEntity)
		if err != nil {
			log.Printf("[SyncToDB] Item %d (ID: %s) DB operation failed: %v — SKIPPING", i, entityID, err)
			skippedCount++
			continue
		}

		processedCount++
	}

	log.Printf("[SyncToDB] Batch complete for %s: %d processed, %d skipped",
		payload.EntityType, processedCount, skippedCount)

	return nil
}

// HandlePullFromSheets processes a TaskPullFromSheets job.
// It reads data from a Google Sheet tab and syncs it to the DB.
// This replicates the 11gawe `sync-from-sheets.cjs` PullService logic including
// header normalization, ID-based upsert, and deletion reconciliation.
func (h *SyncToDBHandler) HandlePullFromSheets(ctx context.Context, t *asynq.Task) error {
	var payload PullFromSheetsPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal PullFromSheetsPayload: %w", err)
	}

	log.Printf("[PullSync] Starting pull sync for: %s", payload.EntityType)

	if payload.EntityType == "all" {
		return h.pullAllEntities(ctx)
	}

	return h.pullSingleEntity(ctx, payload.EntityType)
}

func (h *SyncToDBHandler) pullAllEntities(ctx context.Context) error {
	var errs []string
	for entityType := range h.registry {
		if err := h.pullSingleEntity(ctx, entityType); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", entityType, err))
		}
		// Sleep 1 second between tabs to avoid Google Sheets quota hit
		time.Sleep(1 * time.Second)
	}
	if len(errs) > 0 {
		return fmt.Errorf("pull errors: %s", strings.Join(errs, "; "))
	}
	log.Println("[PullSync] ✅ Full pull sync complete for all entities")
	return nil
}

func (h *SyncToDBHandler) pullSingleEntity(ctx context.Context, entityType string) error {
	cfg, ok := h.registry[entityType]
	if !ok {
		return fmt.Errorf("unknown entity type: %s", entityType)
	}

	log.Printf("[PullSync] Pulling %s from sheet '%s'...", entityType, cfg.Sheet.SheetName)

	// Read headers + data from sheet
	allRows, err := h.sheetsService.BatchReadWithHeaders(cfg.Sheet.SheetName)
	if err != nil {
		return fmt.Errorf("failed to read sheet '%s': %w", cfg.Sheet.SheetName, err)
	}

	if len(allRows) == 0 {
		log.Printf("[PullSync] %s sheet has no rows — skipping", entityType)
		return nil
	}

	headers := allRows[0]
	dataRows := allRows[1:]

	if len(dataRows) == 0 {
		log.Printf("[PullSync] %s sheet has no data rows — skipping deletion reconciliation to protect DB",
			entityType)
		return nil
	}

	// Normalize headers to UPPERCASE
	normalizedHeaders := make([]string, len(headers))
	for i, h := range headers {
		normalizedHeaders[i] = service.NormalizeHeader(h)
	}

	// Process rows
	remoteIDs := make(map[string]bool)
	var processedCount, skippedCount int

	for _, row := range dataRows {
		if len(row) == 0 || row[0] == "" {
			continue
		}

		// Map row to header → value
		item := make(map[string]string)
		for j, header := range normalizedHeaders {
			if j < len(row) {
				item[header] = row[j]
			}
		}

		rowID, ok := item["ID"]
		if !ok || rowID == "" {
			skippedCount++
			continue
		}

		entityID, updates, createEntity := cfg.ToUpdates(item)
		if entityID == "" {
			skippedCount++
			continue
		}

		err := h.repo.UpsertSimple(cfg.NewModel(), entityID, updates, createEntity)
		if err != nil {
			log.Printf("[PullSync] Failed upsert for %s ID %s: %v", entityType, entityID, err)
			skippedCount++
			continue
		}

		remoteIDs[entityID] = true
		processedCount++
	}

	// Deletion reconciliation (matching 11gawe logic)
	if !cfg.SkipDeletionReconciliation && len(remoteIDs) > 0 {
		deletedCount := h.repo.ReconcileDeletions(cfg.NewSlice(), remoteIDs)
		if deletedCount > 0 {
			log.Printf("[PullSync] Soft-deleted %d %s records not found in sheet", deletedCount, entityType)
		}
	}

	log.Printf("[PullSync] %s complete: %d processed, %d skipped", entityType, processedCount, skippedCount)
	return nil
}
