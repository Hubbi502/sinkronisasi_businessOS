package worker

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

// Task type constants for the message queue.
const (
	TaskSyncEntityToSheet = "sync:entity_to_sheet"  // Sync single entity DB → Sheet
	TaskProcessSheetBatch = "sync:process_batch"     // Process batch from GAS → DB
	TaskRevertSheetRow    = "sync:revert_row"        // Revert a conflicting sheet row
	TaskFullSync          = "sync:full_sync"         // Full DB → Sheet sync for one entity type
	TaskPullFromSheets    = "sync:pull_from_sheets"  // Pull from Sheet → DB (periodic sync)
)

// SyncEntityPayload is the payload for TaskSyncEntityToSheet.
type SyncEntityPayload struct {
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
}

// ProcessSheetBatchPayload wraps a batch of items from GAS for a specific entity type.
type ProcessSheetBatchPayload struct {
	EntityType string              `json:"entity_type"`
	Items      []map[string]string `json:"items"` // Column name → value
}

// RevertSheetRowPayload is the payload for TaskRevertSheetRow.
type RevertSheetRowPayload struct {
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
}

// FullSyncPayload triggers a full DB → Sheet sync for one entity type.
type FullSyncPayload struct {
	EntityType string `json:"entity_type"`
}

// PullFromSheetsPayload triggers a Sheet → DB pull for one entity type or "all".
type PullFromSheetsPayload struct {
	EntityType string `json:"entity_type"` // entity type or "all"
}

// ─── Task Constructors ──────────────────────────────────────────

func NewSyncEntityTask(entityType, entityID string) (*asynq.Task, error) {
	payload := SyncEntityPayload{EntityType: entityType, EntityID: entityID}
	return asynq.NewTask(TaskSyncEntityToSheet, mustMarshal(payload)), nil
}

func NewProcessSheetBatchTask(entityType string, items []map[string]string) (*asynq.Task, error) {
	payload := ProcessSheetBatchPayload{EntityType: entityType, Items: items}
	return asynq.NewTask(TaskProcessSheetBatch, mustMarshal(payload)), nil
}

func NewRevertSheetRowTask(entityType, entityID string) (*asynq.Task, error) {
	payload := RevertSheetRowPayload{EntityType: entityType, EntityID: entityID}
	return asynq.NewTask(TaskRevertSheetRow, mustMarshal(payload)), nil
}

func NewFullSyncTask(entityType string) (*asynq.Task, error) {
	payload := FullSyncPayload{EntityType: entityType}
	return asynq.NewTask(TaskFullSync, mustMarshal(payload)), nil
}

func NewPullFromSheetsTask(entityType string) (*asynq.Task, error) {
	payload := PullFromSheetsPayload{EntityType: entityType}
	return asynq.NewTask(TaskPullFromSheets, mustMarshal(payload)), nil
}

func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
