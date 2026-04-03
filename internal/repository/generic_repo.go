package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// ErrVersionConflict is returned when an OCC check fails.
var ErrVersionConflict = errors.New("version conflict: data has been modified by another process")

// GenericRepository handles all database operations for any entity.
type GenericRepository struct {
	db *gorm.DB
}

// NewGenericRepository creates a new GenericRepository.
func NewGenericRepository(db *gorm.DB) *GenericRepository {
	return &GenericRepository{db: db}
}

// Create inserts a new entity into the database.
func (r *GenericRepository) Create(entity interface{}) error {
	return r.db.Create(entity).Error
}

// GetByID fetches a single entity by its ID. `dest` must be a pointer to the model struct.
func (r *GenericRepository) GetByID(dest interface{}, id string) error {
	return r.db.Where("id = ?", id).First(dest).Error
}

// GetAll returns all non-deleted entities. `dest` must be a pointer to a slice.
func (r *GenericRepository) GetAll(dest interface{}) error {
	return r.db.Where("is_deleted = ?", false).Order("created_at DESC").Find(dest).Error
}

// Update performs a standard update on the entity.
func (r *GenericRepository) Update(entity interface{}) error {
	return r.db.Save(entity).Error
}

// SoftDelete marks an entity as deleted by setting is_deleted = true.
func (r *GenericRepository) SoftDelete(model interface{}, id string) error {
	return r.db.Model(model).Where("id = ?", id).Updates(map[string]interface{}{
		"is_deleted": true,
		"version":    gorm.Expr("version + 1"),
	}).Error
}

// UpdateWithOCC performs an Optimistic Concurrency Control update.
// `model` is a pointer to an empty struct of the entity type (e.g., &model.SalesOrder{}).
// It only updates if the current DB version matches expectedVersion.
func (r *GenericRepository) UpdateWithOCC(model interface{}, id string, updates map[string]interface{}, expectedVersion int) error {
	// Always bump version
	updates["version"] = gorm.Expr("version + 1")

	result := r.db.Model(model).
		Where("id = ? AND version = ?", id, expectedVersion).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: expected version %d for entity %s", ErrVersionConflict, expectedVersion, id)
	}

	return nil
}

// Upsert creates the entity if it doesn't exist, or updates it with OCC if it does.
// `model` is a pointer to an empty struct of the entity type.
// `entity` is the fully populated entity to upsert.
func (r *GenericRepository) Upsert(model interface{}, id string, updates map[string]interface{}, expectedVersion int, createEntity interface{}) error {
	// Try to find existing
	result := r.db.Where("id = ?", id).First(model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Create new
			return r.db.Create(createEntity).Error
		}
		return result.Error
	}

	// Exists — do OCC update
	return r.UpdateWithOCC(model, id, updates, expectedVersion)
}

// UpsertSimple creates the entity if it doesn't exist, or updates it without OCC.
// This matches the 11gawe behavior where Sheets always wins.
func (r *GenericRepository) UpsertSimple(model interface{}, id string, updates map[string]interface{}, createEntity interface{}) error {
	result := r.db.Where("id = ?", id).First(model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return r.db.Create(createEntity).Error
		}
		return result.Error
	}

	// Exists — update directly (Sheets wins, no version check)
	return r.db.Model(model).Where("id = ?", id).Updates(updates).Error
}

// ReconcileDeletions soft-deletes local records that are not present in the remote set.
// This matches the 11gawe deletion reconciliation logic from sync-from-sheets.cjs.
// Returns the number of records deleted.
func (r *GenericRepository) ReconcileDeletions(slicePtr interface{}, remoteIDs map[string]bool) int {
	// Find all non-deleted records
	r.db.Where("is_deleted = ?", false).Find(slicePtr)

	// We need to iterate over the slice; use GORM's approach
	type IDHolder struct {
		ID string
	}

	var localRecords []IDHolder
	r.db.Model(slicePtr).
		Where("is_deleted = ?", false).
		Select("id").
		Find(&localRecords)

	var deletedCount int
	for _, local := range localRecords {
		if !remoteIDs[local.ID] {
			r.db.Model(slicePtr).
				Where("id = ?", local.ID).
				Updates(map[string]interface{}{"is_deleted": true})
			deletedCount++
		}
	}
	return deletedCount
}

