package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"sinkronisasi_db/internal/repository"
	"sinkronisasi_db/internal/service"

	"github.com/hibiken/asynq"
)

// SyncToSheetHandler processes TaskSyncEntityToSheet, TaskRevertSheetRow, and TaskFullSync jobs.
type SyncToSheetHandler struct {
	repo          *repository.GenericRepository
	sheetsService *service.SheetsService
	registry      map[string]*service.EntityConfig
}

// NewSyncToSheetHandler creates a new handler for sheet sync tasks.
func NewSyncToSheetHandler(
	repo *repository.GenericRepository,
	sheetsService *service.SheetsService,
	registry map[string]*service.EntityConfig,
) *SyncToSheetHandler {
	return &SyncToSheetHandler{
		repo:          repo,
		sheetsService: sheetsService,
		registry:      registry,
	}
}

// HandleSyncEntityToSheet processes a TaskSyncEntityToSheet job.
// It reads the entity from DB and pushes it to the correct Google Sheet tab.
func (h *SyncToSheetHandler) HandleSyncEntityToSheet(ctx context.Context, t *asynq.Task) error {
	var payload SyncEntityPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal SyncEntityPayload: %w", err)
	}

	cfg, ok := h.registry[payload.EntityType]
	if !ok {
		return fmt.Errorf("unknown entity type: %s", payload.EntityType)
	}

	log.Printf("[SyncToSheet] Syncing %s/%s to sheet", payload.EntityType, payload.EntityID)

	// Fetch entity from DB
	entity := cfg.NewModel()
	if err := h.repo.GetByID(entity, payload.EntityID); err != nil {
		return fmt.Errorf("failed to fetch %s/%s from DB: %w", payload.EntityType, payload.EntityID, err)
	}

	// Convert to sheet row
	values := cfg.ToRow(entity)

	// Upsert to Google Sheets
	if err := h.sheetsService.UpsertRow(
		cfg.Sheet.SheetName,
		cfg.Sheet.ColRange,
		payload.EntityID,
		values,
	); err != nil {
		return fmt.Errorf("failed to upsert %s/%s to sheet: %w", payload.EntityType, payload.EntityID, err)
	}

	log.Printf("[SyncToSheet] Successfully synced %s/%s", payload.EntityType, payload.EntityID)
	return nil
}

// HandleRevertSheetRow processes a TaskRevertSheetRow job.
func (h *SyncToSheetHandler) HandleRevertSheetRow(ctx context.Context, t *asynq.Task) error {
	var payload RevertSheetRowPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal RevertSheetRowPayload: %w", err)
	}

	cfg, ok := h.registry[payload.EntityType]
	if !ok {
		return fmt.Errorf("unknown entity type: %s", payload.EntityType)
	}

	log.Printf("[SyncToSheet] Reverting sheet row for %s/%s", payload.EntityType, payload.EntityID)

	entity := cfg.NewModel()
	if err := h.repo.GetByID(entity, payload.EntityID); err != nil {
		return fmt.Errorf("failed to fetch %s/%s for revert: %w", payload.EntityType, payload.EntityID, err)
	}

	values := cfg.ToRow(entity)
	if err := h.sheetsService.UpsertRow(
		cfg.Sheet.SheetName,
		cfg.Sheet.ColRange,
		payload.EntityID,
		values,
	); err != nil {
		return fmt.Errorf("failed to revert sheet row for %s/%s: %w", payload.EntityType, payload.EntityID, err)
	}

	log.Printf("[SyncToSheet] Successfully reverted %s/%s", payload.EntityType, payload.EntityID)
	return nil
}

// HandleFullSync processes a TaskFullSync job.
// It reads all entities from DB and writes them to the corresponding sheet tab.
func (h *SyncToSheetHandler) HandleFullSync(ctx context.Context, t *asynq.Task) error {
	var payload FullSyncPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal FullSyncPayload: %w", err)
	}

	cfg, ok := h.registry[payload.EntityType]
	if !ok {
		return fmt.Errorf("unknown entity type: %s", payload.EntityType)
	}

	log.Printf("[SyncToSheet] Full sync for entity type: %s", payload.EntityType)

	// Fetch all entities from DB
	entities := cfg.NewSlice()
	if err := h.repo.GetAll(entities); err != nil {
		return fmt.Errorf("failed to fetch all %s from DB: %w", payload.EntityType, err)
	}

	// Convert slice to rows — we need to use reflection-free approach
	// The slice is already populated; convert each to a row
	rows := h.convertSliceToRows(cfg, entities)

	// Write all to sheet
	if err := h.sheetsService.BatchWrite(
		cfg.Sheet.SheetName,
		cfg.Sheet.ColRange,
		cfg.Sheet.Columns,
		rows,
	); err != nil {
		return fmt.Errorf("failed to batch write %s to sheet: %w", payload.EntityType, err)
	}

	log.Printf("[SyncToSheet] Full sync complete for %s: %d rows", payload.EntityType, len(rows))
	return nil
}

// convertSliceToRows converts a slice of entities to sheet rows using JSON marshal/unmarshal.
func (h *SyncToSheetHandler) convertSliceToRows(cfg *service.EntityConfig, slice interface{}) [][]interface{} {
	// Marshal to JSON and back to get []interface{} slice
	data, err := json.Marshal(slice)
	if err != nil {
		log.Printf("[SyncToSheet] Failed to marshal slice: %v", err)
		return nil
	}

	var items []json.RawMessage
	if err := json.Unmarshal(data, &items); err != nil {
		log.Printf("[SyncToSheet] Failed to unmarshal to raw items: %v", err)
		return nil
	}

	var rows [][]interface{}
	for _, item := range items {
		entity := cfg.NewModel()
		if err := json.Unmarshal(item, entity); err != nil {
			log.Printf("[SyncToSheet] Failed to unmarshal item: %v", err)
			continue
		}
		rows = append(rows, cfg.ToRow(entity))
	}

	return rows
}
