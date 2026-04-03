package handler

import (
	"log"
	"net/http"

	"sinkronisasi_db/internal/service"
	"sinkronisasi_db/internal/worker"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

// WebhookHandler handles incoming webhook requests from Google Apps Script.
type WebhookHandler struct {
	client   *asynq.Client
	registry map[string]*service.EntityConfig
}

// NewWebhookHandler creates a new WebhookHandler.
func NewWebhookHandler(client *asynq.Client, registry map[string]*service.EntityConfig) *WebhookHandler {
	return &WebhookHandler{
		client:   client,
		registry: registry,
	}
}

// SheetSyncRequest is the expected JSON body from GAS.
type SheetSyncRequest struct {
	EntityType string              `json:"entity_type" binding:"required"`
	Items      []map[string]string `json:"items" binding:"required"`
}

// HandleSheetSync handles POST /api/webhook/sheet-sync
func (h *WebhookHandler) HandleSheetSync(c *gin.Context) {
	var req SheetSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, ok := h.registry[req.EntityType]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "unknown entity_type",
			"entity_type": req.EntityType,
			"valid_types": getEntityTypeKeys(h.registry),
		})
		return
	}

	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no items in batch"})
		return
	}

	task, err := worker.NewProcessSheetBatchTask(req.EntityType, req.Items)
	if err != nil {
		log.Printf("[WebhookHandler] Failed to create batch task: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create processing task"})
		return
	}

	if _, err := h.client.Enqueue(task); err != nil {
		log.Printf("[WebhookHandler] Failed to enqueue batch task: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue processing task"})
		return
	}

	log.Printf("[WebhookHandler] Enqueued batch of %d items for entity type: %s", len(req.Items), req.EntityType)

	c.JSON(http.StatusAccepted, gin.H{
		"message":     "batch accepted for processing",
		"entity_type": req.EntityType,
		"count":       len(req.Items),
	})
}

// HandleFullSync handles POST /api/webhook/full-sync (DB → Sheet)
func (h *WebhookHandler) HandleFullSync(c *gin.Context) {
	entityType := c.Query("entity_type")

	if entityType != "" {
		if _, ok := h.registry[entityType]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":       "unknown entity_type",
				"valid_types": getEntityTypeKeys(h.registry),
			})
			return
		}

		task, _ := worker.NewFullSyncTask(entityType)
		if _, err := h.client.Enqueue(task); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue full sync task"})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{"message": "full sync enqueued", "entity_type": entityType})
		return
	}

	// Sync all
	var enqueued []string
	for et := range h.registry {
		task, _ := worker.NewFullSyncTask(et)
		if _, err := h.client.Enqueue(task); err != nil {
			log.Printf("[WebhookHandler] Failed to enqueue full sync for %s: %v", et, err)
			continue
		}
		enqueued = append(enqueued, et)
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "full sync enqueued for all entity types", "enqueued": enqueued})
}

// HandlePullSync handles POST /api/webhook/pull-sync (Sheet → DB)
// Replicates 11gawe's sync-from-sheets.cjs behavior.
func (h *WebhookHandler) HandlePullSync(c *gin.Context) {
	entityType := c.DefaultQuery("entity_type", "all")

	if entityType != "all" {
		if _, ok := h.registry[entityType]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":       "unknown entity_type",
				"valid_types": getEntityTypeKeys(h.registry),
			})
			return
		}
	}

	task, _ := worker.NewPullFromSheetsTask(entityType)
	if _, err := h.client.Enqueue(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue pull sync task"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":     "pull sync enqueued",
		"entity_type": entityType,
	})
}

func getEntityTypeKeys(registry map[string]*service.EntityConfig) []string {
	keys := make([]string, 0, len(registry))
	for k := range registry {
		keys = append(keys, k)
	}
	return keys
}
