package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"sinkronisasi_db/internal/model"
)

// SheetConfig defines the column mapping for a single sheet tab.
type SheetConfig struct {
	SheetName string   // Name of the tab in Google Sheets (must match exactly)
	Columns   []string // Ordered column headers
	ColRange  string   // e.g. "A:K" — range for read/write
}

// EntityToRow converts a DB entity to a sheet row ([]interface{}).
type EntityToRow func(entity interface{}) []interface{}

// RowToUpdates converts a sheet row (map[string]string) to DB updates.
// Returns: (id, updates map, create entity for upsert)
type RowToUpdates func(row map[string]string) (id string, updates map[string]interface{}, createEntity interface{})

// EntityConfig holds all sync configuration for an entity type.
type EntityConfig struct {
	Sheet     SheetConfig
	ToRow     EntityToRow
	ToUpdates RowToUpdates
	NewModel  func() interface{} // Factory for empty model pointer
	NewSlice  func() interface{} // Factory for empty model slice
	// ModelName is the Prisma/GORM model identifier used for deletion reconciliation
	ModelName string
	// SkipDeletionReconciliation: if true, don't mark missing records as deleted
	SkipDeletionReconciliation bool
}

// ─── Helpers ────────────────────────────────────────────────────

func strPtr(s string) *string {
	if s == "" || s == "-" {
		return nil
	}
	return &s
}

func floatPtr(s string) *float64 {
	if s == "" || s == "-" {
		return nil
	}
	s = strings.ReplaceAll(s, ",", ".")
	// Handle thousand separators: "1.000.000" → "1000000"
	// If it has dots followed by 3 digits and ends with a dot-2digit pattern, it's European
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &v
}

func parseFloat(s string) float64 {
	if s == "" || s == "-" {
		return 0
	}
	s = strings.ReplaceAll(s, ",", ".")
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func parseBool(s string) bool {
	s = strings.ToUpper(strings.TrimSpace(s))
	return s == "TRUE" || s == "YES" || s == "1"
}

func parseTime(s string) *time.Time {
	if s == "" || s == "-" {
		return nil
	}
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, strings.TrimSpace(s)); err == nil {
			return &t
		}
	}
	return nil
}

func parseTimeVal(s string) time.Time {
	t := parseTime(s)
	if t == nil {
		return time.Time{}
	}
	return *t
}

func fmtTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func fmtTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func fmtFloat(f *float64) interface{} {
	if f == nil {
		return ""
	}
	return *f
}

func fmtStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func fmtBool(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func getCol(row map[string]string, key string) string {
	return strings.TrimSpace(row[key])
}

// NormalizeHeader normalizes a header for case-insensitive matching (matches 11gawe pattern)
func NormalizeHeader(h string) string {
	return strings.ToUpper(strings.TrimSpace(strings.Join(strings.Fields(h), " ")))
}

// ─── Entity Registry ───────────────────────────────────────────

// BuildEntityRegistry returns the full registry of entity sync configurations.
// Tab names match EXACTLY what 11gawe uses.
func BuildEntityRegistry() map[string]*EntityConfig {
	return map[string]*EntityConfig{
		"tasks":          buildTaskItemConfig(),
		"sales":          buildSalesOrderConfig(),
		"expenses":       buildPurchaseRequestConfig(),
		"shipments":      buildShipmentDetailConfig(),
		"sources":        buildSourceSupplierConfig(),
		"quality":        buildQualityResultConfig(),
		"market_price":   buildMarketPriceConfig(),
		"meetings":       buildMeetingItemConfig(),
		"pl_forecast":    buildPLForecastConfig(),
		"projects":       buildSalesDealConfig(),
		"partners":       buildPartnerConfig(),
		"blending":       buildBlendingConfig(),
	}
}

// ─── Tasks (tab: "Tasks") ─────────────────────────────────────

func buildTaskItemConfig() *EntityConfig {
	cols := []string{"ID", "Title", "Description", "Status", "Priority", "Assignee", "Due Date", "Image Preview", "Updated At"}
	return &EntityConfig{
		Sheet: SheetConfig{SheetName: "Tasks", Columns: cols, ColRange: "A:I"},
		ModelName: "TaskItem",
		NewModel:  func() interface{} { return &model.TaskItem{} },
		NewSlice:  func() interface{} { return &[]model.TaskItem{} },
		ToRow: func(entity interface{}) []interface{} {
			e := entity.(*model.TaskItem)
			return []interface{}{
				e.ID, e.Title, fmtStr(e.Description), e.Status, e.Priority,
				fmtStr(e.AssigneeName), fmtTimePtr(e.DueDate), "", fmtTime(e.UpdatedAt),
			}
		},
		ToUpdates: func(row map[string]string) (string, map[string]interface{}, interface{}) {
			id := getCol(row, "ID")
			updates := map[string]interface{}{
				"title":         getCol(row, "TITLE"),
				"description":   strPtr(getCol(row, "DESCRIPTION")),
				"status":        orDefault(getCol(row, "STATUS"), "todo"),
				"priority":      orDefault(getCol(row, "PRIORITY"), "medium"),
				"assigneeName": strPtr(getCol(row, "ASSIGNEE")),
				"dueDate":      parseTime(getCol(row, "DUE DATE")),
			}
			entity := &model.TaskItem{
				ID: id, Title: orDefault(getCol(row, "TITLE"), "Untitled"),
				Description: strPtr(getCol(row, "DESCRIPTION")),
				Status: orDefault(getCol(row, "STATUS"), "todo"), Priority: orDefault(getCol(row, "PRIORITY"), "medium"),
				AssigneeName: strPtr(getCol(row, "ASSIGNEE")), DueDate: parseTime(getCol(row, "DUE DATE")),
				CreatedBy: "system", Version: 1,
			}
			return id, updates, entity
		},
	}
}

// ─── Sales Orders (tab: "Sales") ──────────────────────────────

func buildSalesOrderConfig() *EntityConfig {
	cols := []string{"ID", "Order #", "Date", "Client", "Description", "Amount", "Priority", "Status", "Created By", "Image Preview", "Updated At"}
	return &EntityConfig{
		Sheet: SheetConfig{SheetName: "Sales", Columns: cols, ColRange: "A:K"},
		ModelName: "SalesOrder",
		NewModel:  func() interface{} { return &model.SalesOrder{} },
		NewSlice:  func() interface{} { return &[]model.SalesOrder{} },
		ToRow: func(entity interface{}) []interface{} {
			e := entity.(*model.SalesOrder)
			return []interface{}{
				e.ID, e.OrderNumber, fmtTime(e.CreatedAt), e.Client,
				fmtStr(e.Description), e.Amount, e.Priority, e.Status,
				fmtStr(e.CreatedByName), fmtStr(e.ImageUrl), fmtTime(e.UpdatedAt),
			}
		},
		ToUpdates: func(row map[string]string) (string, map[string]interface{}, interface{}) {
			id := getCol(row, "ID")
			updates := map[string]interface{}{
				"orderNumber":    getCol(row, "ORDER #"),
				"client":          orDefault(getCol(row, "CLIENT"), "Unknown"),
				"description":     strPtr(getCol(row, "DESCRIPTION")),
				"amount":          parseFloat(getCol(row, "AMOUNT")),
				"priority":        orDefault(getCol(row, "PRIORITY"), "medium"),
				"status":          orDefault(getCol(row, "STATUS"), "pending"),
				"createdByName": strPtr(getCol(row, "CREATED BY")),
				"imageUrl":       strPtr(getCol(row, "IMAGE PREVIEW")),
			}
			entity := &model.SalesOrder{
				ID: id, OrderNumber: orDefault(getCol(row, "ORDER #"), id),
				Client: orDefault(getCol(row, "CLIENT"), "Unknown"),
				Description: strPtr(getCol(row, "DESCRIPTION")), Amount: parseFloat(getCol(row, "AMOUNT")),
				Priority: orDefault(getCol(row, "PRIORITY"), "medium"), Status: orDefault(getCol(row, "STATUS"), "pending"),
				CreatedByName: strPtr(getCol(row, "CREATED BY")), ImageUrl: strPtr(getCol(row, "IMAGE PREVIEW")),
				CreatedBy: "system", Version: 1,
			}
			return id, updates, entity
		},
	}
}

// ─── Purchase Requests (tab: "Expenses") ──────────────────────

func buildPurchaseRequestConfig() *EntityConfig {
	cols := []string{"ID", "Request #", "Date", "Category", "Supplier", "Description", "Amount", "Priority", "Status", "Created By", "Image Preview", "Updated At"}
	return &EntityConfig{
		Sheet: SheetConfig{SheetName: "Expenses", Columns: cols, ColRange: "A:L"},
		ModelName: "PurchaseRequest",
		NewModel:  func() interface{} { return &model.PurchaseRequest{} },
		NewSlice:  func() interface{} { return &[]model.PurchaseRequest{} },
		ToRow: func(entity interface{}) []interface{} {
			e := entity.(*model.PurchaseRequest)
			return []interface{}{
				e.ID, e.RequestNumber, fmtTime(e.CreatedAt), e.Category,
				fmtStr(e.Supplier), fmtStr(e.Description), e.Amount, e.Priority,
				e.Status, fmtStr(e.CreatedByName), fmtStr(e.ImageUrl), fmtTime(e.UpdatedAt),
			}
		},
		ToUpdates: func(row map[string]string) (string, map[string]interface{}, interface{}) {
			id := getCol(row, "ID")
			updates := map[string]interface{}{
				"requestNumber":  getCol(row, "REQUEST #"),
				"category":        orDefault(getCol(row, "CATEGORY"), "Other"),
				"supplier":        strPtr(getCol(row, "SUPPLIER")),
				"description":     strPtr(getCol(row, "DESCRIPTION")),
				"amount":          parseFloat(getCol(row, "AMOUNT")),
				"priority":        orDefault(getCol(row, "PRIORITY"), "medium"),
				"status":          orDefault(getCol(row, "STATUS"), "pending"),
				"createdByName": strPtr(getCol(row, "CREATED BY")),
				"imageUrl":       strPtr(getCol(row, "IMAGE PREVIEW")),
			}
			entity := &model.PurchaseRequest{
				ID: id, RequestNumber: orDefault(getCol(row, "REQUEST #"), id),
				Category: orDefault(getCol(row, "CATEGORY"), "Other"),
				Supplier: strPtr(getCol(row, "SUPPLIER")), Description: strPtr(getCol(row, "DESCRIPTION")),
				Amount: parseFloat(getCol(row, "AMOUNT")), Priority: orDefault(getCol(row, "PRIORITY"), "medium"),
				Status: orDefault(getCol(row, "STATUS"), "pending"),
				CreatedByName: strPtr(getCol(row, "CREATED BY")), ImageUrl: strPtr(getCol(row, "IMAGE PREVIEW")),
				CreatedBy: "system", Version: 1,
			}
			return id, updates, entity
		},
	}
}

// ─── Shipment Details (tab: "Shipments") ──────────────────────

func buildShipmentDetailConfig() *EntityConfig {
	cols := []string{"ID", "Shipment #", "Deal ID", "Status", "Buyer", "Supplier", "Is Blending", "IUP OP", "Vessel Name", "Barge Name", "Loading Port", "Discharge Port", "Qty Loaded (MT)", "BL Date", "ETA", "Sales Price", "Margin/MT", "PIC", "Type", "Milestones", "Created At", "Updated At"}
	return &EntityConfig{
		Sheet: SheetConfig{SheetName: "Shipments", Columns: cols, ColRange: "A:V"},
		ModelName: "ShipmentDetail",
		NewModel:  func() interface{} { return &model.ShipmentDetail{} },
		NewSlice:  func() interface{} { return &[]model.ShipmentDetail{} },
		ToRow: func(entity interface{}) []interface{} {
			e := entity.(*model.ShipmentDetail)
			return []interface{}{
				e.ID, e.ShipmentNumber, fmtStr(e.DealId), e.Status, e.Buyer, fmtStr(e.Supplier),
				fmtBool(e.IsBlending), fmtStr(e.IupOp), fmtStr(e.VesselName), fmtStr(e.BargeName),
				fmtStr(e.LoadingPort), fmtStr(e.DischargePort), fmtFloat(e.QuantityLoaded),
				fmtTimePtr(e.BlDate), fmtTimePtr(e.Eta), fmtFloat(e.SalesPrice),
				fmtFloat(e.MarginMt), fmtStr(e.PicName), e.Type,
				fmtStr(e.Milestones), fmtTime(e.CreatedAt), fmtTime(e.UpdatedAt),
			}
		},
		ToUpdates: func(row map[string]string) (string, map[string]interface{}, interface{}) {
			id := getCol(row, "ID")
			updates := map[string]interface{}{
				"shipmentNumber": getCol(row, "SHIPMENT #"),
				"dealId":         strPtr(getCol(row, "DEAL ID")),
				"status":          normalizeShipmentStatus(getCol(row, "STATUS")),
				"buyer":           orDefault(getCol(row, "BUYER"), "Unknown"),
				"supplier":        strPtr(getCol(row, "SUPPLIER")),
				"isBlending":     parseBool(getCol(row, "IS BLENDING")),
				"iupOp":          strPtr(getCol(row, "IUP OP")),
				"vesselName":     strPtr(getCol(row, "VESSEL NAME")),
				"bargeName":      strPtr(getCol(row, "BARGE NAME")),
				"loadingPort":    strPtr(getCol(row, "LOADING PORT")),
				"dischargePort":  strPtr(getCol(row, "DISCHARGE PORT")),
				"quantityLoaded": floatPtr(getCol(row, "QTY LOADED (MT)")),
				"blDate":         parseTime(getCol(row, "BL DATE")),
				"eta":             parseTime(getCol(row, "ETA")),
				"salesPrice":     floatPtr(getCol(row, "SALES PRICE")),
				"marginMt":       floatPtr(getCol(row, "MARGIN/MT")),
				"picName":        strPtr(getCol(row, "PIC")),
				"type":            orDefault(getCol(row, "TYPE"), "export"),
			}
			entity := &model.ShipmentDetail{
				ID: id, ShipmentNumber: orDefault(getCol(row, "SHIPMENT #"), id),
				DealId: strPtr(getCol(row, "DEAL ID")),
				Status: normalizeShipmentStatus(getCol(row, "STATUS")), Buyer: orDefault(getCol(row, "BUYER"), "Unknown"),
				Supplier: strPtr(getCol(row, "SUPPLIER")), IsBlending: parseBool(getCol(row, "IS BLENDING")),
				IupOp: strPtr(getCol(row, "IUP OP")), VesselName: strPtr(getCol(row, "VESSEL NAME")),
				BargeName: strPtr(getCol(row, "BARGE NAME")), LoadingPort: strPtr(getCol(row, "LOADING PORT")),
				DischargePort: strPtr(getCol(row, "DISCHARGE PORT")),
				QuantityLoaded: floatPtr(getCol(row, "QTY LOADED (MT)")),
				BlDate: parseTime(getCol(row, "BL DATE")), Eta: parseTime(getCol(row, "ETA")),
				SalesPrice: floatPtr(getCol(row, "SALES PRICE")), MarginMt: floatPtr(getCol(row, "MARGIN/MT")),
				PicName: strPtr(getCol(row, "PIC")), Type: orDefault(getCol(row, "TYPE"), "export"),
				Version: 1,
			}
			return id, updates, entity
		},
	}
}

// ─── Source Suppliers (tab: "Sources") ─────────────────────────

func buildSourceSupplierConfig() *EntityConfig {
	cols := []string{"ID", "Name", "Region", "Calorie Range", "GAR", "TS", "Ash", "TM", "Jetty Port", "Anchorage", "Stock Available", "Min Stock Alert", "KYC Status", "PSI Status", "FOB Barge Only", "Price Linked Index", "FOB Barge Price (USD)", "Contract Type", "PIC", "IUP Number", "Updated At"}
	return &EntityConfig{
		Sheet: SheetConfig{SheetName: "Sources", Columns: cols, ColRange: "A:U"},
		ModelName: "SourceSupplier",
		NewModel:  func() interface{} { return &model.SourceSupplier{} },
		NewSlice:  func() interface{} { return &[]model.SourceSupplier{} },
		ToRow: func(entity interface{}) []interface{} {
			e := entity.(*model.SourceSupplier)
			return []interface{}{
				e.ID, e.Name, e.Region, fmtStr(e.CalorieRange),
				fmtFloat(e.Gar), fmtFloat(e.Ts), fmtFloat(e.Ash), fmtFloat(e.Tm),
				fmtStr(e.JettyPort), fmtStr(e.Anchorage), e.StockAvailable,
				fmtFloat(e.MinStockAlert), e.KycStatus, e.PsiStatus,
				fmtBool(e.FobBargeOnly), fmtStr(e.PriceLinkedIndex),
				fmtFloat(e.FobBargePriceUsd), fmtStr(e.ContractType),
				fmtStr(e.PicName), fmtStr(e.IupNumber), fmtTime(e.UpdatedAt),
			}
		},
		ToUpdates: func(row map[string]string) (string, map[string]interface{}, interface{}) {
			id := getCol(row, "ID")
			updates := map[string]interface{}{
				"name": orDefault(getCol(row, "NAME"), "Unknown"), "region": orDefault(getCol(row, "REGION"), "Unknown"),
				"calorieRange": strPtr(getCol(row, "CALORIE RANGE")),
				"gar": floatPtr(getCol(row, "GAR")), "ts": floatPtr(getCol(row, "TS")),
				"ash": floatPtr(getCol(row, "ASH")), "tm": floatPtr(getCol(row, "TM")),
				"jettyPort": strPtr(getCol(row, "JETTY PORT")), "anchorage": strPtr(getCol(row, "ANCHORAGE")),
				"stockAvailable": parseFloat(getCol(row, "STOCK AVAILABLE")),
				"minStockAlert": floatPtr(getCol(row, "MIN STOCK ALERT")),
				"kycStatus": orDefault(getCol(row, "KYC STATUS"), "not_started"),
				"psiStatus": orDefault(getCol(row, "PSI STATUS"), "not_started"),
				"fobBargeOnly": parseBool(getCol(row, "FOB BARGE ONLY")),
				"priceLinkedIndex": strPtr(getCol(row, "PRICE LINKED INDEX")),
				"fobBargePriceUsd": floatPtr(getCol(row, "FOB BARGE PRICE (USD)")),
				"contractType": strPtr(getCol(row, "CONTRACT TYPE")),
				"picName": strPtr(getCol(row, "PIC")), "iupNumber": strPtr(getCol(row, "IUP NUMBER")),
			}
			entity := &model.SourceSupplier{
				ID: id, Name: orDefault(getCol(row, "NAME"), "Unknown"), Region: orDefault(getCol(row, "REGION"), "Unknown"),
				CalorieRange: strPtr(getCol(row, "CALORIE RANGE")),
				Gar: floatPtr(getCol(row, "GAR")), Ts: floatPtr(getCol(row, "TS")),
				Ash: floatPtr(getCol(row, "ASH")), Tm: floatPtr(getCol(row, "TM")),
				JettyPort: strPtr(getCol(row, "JETTY PORT")), Anchorage: strPtr(getCol(row, "ANCHORAGE")),
				StockAvailable: parseFloat(getCol(row, "STOCK AVAILABLE")),
				MinStockAlert: floatPtr(getCol(row, "MIN STOCK ALERT")),
				KycStatus: orDefault(getCol(row, "KYC STATUS"), "not_started"),
				PsiStatus: orDefault(getCol(row, "PSI STATUS"), "not_started"),
				FobBargeOnly: parseBool(getCol(row, "FOB BARGE ONLY")),
				PriceLinkedIndex: strPtr(getCol(row, "PRICE LINKED INDEX")),
				FobBargePriceUsd: floatPtr(getCol(row, "FOB BARGE PRICE (USD)")),
				ContractType: strPtr(getCol(row, "CONTRACT TYPE")),
				PicName: strPtr(getCol(row, "PIC")), IupNumber: strPtr(getCol(row, "IUP NUMBER")),
				Version: 1,
			}
			return id, updates, entity
		},
	}
}

// ─── Quality Results (tab: "Quality") ─────────────────────────

func buildQualityResultConfig() *EntityConfig {
	cols := []string{"ID", "Cargo ID", "Cargo Name", "Surveyor", "Sampling Date", "GAR", "TS", "Ash", "TM", "Status", "Updated At"}
	return &EntityConfig{
		Sheet: SheetConfig{SheetName: "Quality", Columns: cols, ColRange: "A:K"},
		ModelName: "QualityResult",
		NewModel:  func() interface{} { return &model.QualityResult{} },
		NewSlice:  func() interface{} { return &[]model.QualityResult{} },
		ToRow: func(entity interface{}) []interface{} {
			e := entity.(*model.QualityResult)
			return []interface{}{
				e.ID, e.CargoId, e.CargoName, fmtStr(e.Surveyor),
				fmtTimePtr(e.SamplingDate), fmtFloat(e.Gar), fmtFloat(e.Ts),
				fmtFloat(e.Ash), fmtFloat(e.Tm), e.Status, fmtTime(e.UpdatedAt),
			}
		},
		ToUpdates: func(row map[string]string) (string, map[string]interface{}, interface{}) {
			id := getCol(row, "ID")
			updates := map[string]interface{}{
				"cargoId": orDefault(getCol(row, "CARGO ID"), id), "cargoName": orDefault(getCol(row, "CARGO NAME"), "Unknown"),
				"surveyor": strPtr(getCol(row, "SURVEYOR")), "samplingDate": parseTime(getCol(row, "SAMPLING DATE")),
				"gar": floatPtr(getCol(row, "GAR")), "ts": floatPtr(getCol(row, "TS")),
				"ash": floatPtr(getCol(row, "ASH")), "tm": floatPtr(getCol(row, "TM")),
				"status": orDefault(getCol(row, "STATUS"), "pending"),
			}
			entity := &model.QualityResult{
				ID: id, CargoId: orDefault(getCol(row, "CARGO ID"), id),
				CargoName: orDefault(getCol(row, "CARGO NAME"), "Unknown"),
				Surveyor: strPtr(getCol(row, "SURVEYOR")), SamplingDate: parseTime(getCol(row, "SAMPLING DATE")),
				Gar: floatPtr(getCol(row, "GAR")), Ts: floatPtr(getCol(row, "TS")),
				Ash: floatPtr(getCol(row, "ASH")), Tm: floatPtr(getCol(row, "TM")),
				Status: orDefault(getCol(row, "STATUS"), "pending"), Version: 1,
			}
			return id, updates, entity
		},
	}
}

// ─── Market Prices (tab: "Market Price") ──────────────────────

func buildMarketPriceConfig() *EntityConfig {
	cols := []string{"ID", "Date", "ICI 1", "ICI 2", "ICI 3", "ICI 4", "ICI 5", "Newcastle", "HBA", "Source", "Updated At"}
	return &EntityConfig{
		Sheet: SheetConfig{SheetName: "Market Price", Columns: cols, ColRange: "A:K"},
		ModelName:                  "MarketPrice",
		SkipDeletionReconciliation: true, // Match 11gawe: no deletion for market prices
		NewModel:                   func() interface{} { return &model.MarketPrice{} },
		NewSlice:                   func() interface{} { return &[]model.MarketPrice{} },
		ToRow: func(entity interface{}) []interface{} {
			e := entity.(*model.MarketPrice)
			return []interface{}{
				e.ID, fmtTime(e.Date), fmtFloat(e.Ici1), fmtFloat(e.Ici2),
				fmtFloat(e.Ici3), fmtFloat(e.Ici4), fmtFloat(e.Ici5),
				fmtFloat(e.Newcastle), fmtFloat(e.Hba), fmtStr(e.Source), fmtTime(e.UpdatedAt),
			}
		},
		ToUpdates: func(row map[string]string) (string, map[string]interface{}, interface{}) {
			id := getCol(row, "ID")
			updates := map[string]interface{}{
				"date": parseTimeVal(getCol(row, "DATE")),
				"ici1": floatPtr(getCol(row, "ICI 1")), "ici2": floatPtr(getCol(row, "ICI 2")),
				"ici3": floatPtr(getCol(row, "ICI 3")), "ici4": floatPtr(getCol(row, "ICI 4")),
				"ici5": floatPtr(getCol(row, "ICI 5")), "newcastle": floatPtr(getCol(row, "NEWCASTLE")),
				"hba": floatPtr(getCol(row, "HBA")), "source": strPtr(getCol(row, "SOURCE")),
			}
			entity := &model.MarketPrice{
				ID: id, Date: parseTimeVal(getCol(row, "DATE")),
				Ici1: floatPtr(getCol(row, "ICI 1")), Ici2: floatPtr(getCol(row, "ICI 2")),
				Ici3: floatPtr(getCol(row, "ICI 3")), Ici4: floatPtr(getCol(row, "ICI 4")),
				Ici5: floatPtr(getCol(row, "ICI 5")), Newcastle: floatPtr(getCol(row, "NEWCASTLE")),
				Hba: floatPtr(getCol(row, "HBA")), Source: strPtr(getCol(row, "SOURCE")),
				Version: 1,
			}
			return id, updates, entity
		},
	}
}

// ─── Meetings (tab: "Meetings") ───────────────────────────────

func buildMeetingItemConfig() *EntityConfig {
	cols := []string{"ID", "Title", "Date", "Time", "Location", "Status", "Attendees", "Voice Note URL", "MoM Content", "AI Summary", "Created By", "Updated At"}
	return &EntityConfig{
		Sheet: SheetConfig{SheetName: "Meetings", Columns: cols, ColRange: "A:L"},
		ModelName: "MeetingItem",
		NewModel:  func() interface{} { return &model.MeetingItem{} },
		NewSlice:  func() interface{} { return &[]model.MeetingItem{} },
		ToRow: func(entity interface{}) []interface{} {
			e := entity.(*model.MeetingItem)
			return []interface{}{
				e.ID, e.Title, fmtTimePtr(e.Date), fmtStr(e.Time),
				fmtStr(e.Location), e.Status, fmtStr(e.Attendees),
				fmtStr(e.VoiceNoteUrl), fmtStr(e.MomContent), fmtStr(e.AiSummary),
				fmtStr(e.CreatedByName), fmtTime(e.UpdatedAt),
			}
		},
		ToUpdates: func(row map[string]string) (string, map[string]interface{}, interface{}) {
			id := getCol(row, "ID")
			updates := map[string]interface{}{
				"title": orDefault(getCol(row, "TITLE"), "Untitled"),
				"date": parseTime(getCol(row, "DATE")), "time": strPtr(getCol(row, "TIME")),
				"location": strPtr(getCol(row, "LOCATION")), "status": orDefault(getCol(row, "STATUS"), "scheduled"),
				"attendees": strPtr(getCol(row, "ATTENDEES")),
				"voiceNoteUrl": strPtr(getCol(row, "VOICE NOTE URL")),
				"momContent": strPtr(getCol(row, "MOM CONTENT")),
				"aiSummary": strPtr(getCol(row, "AI SUMMARY")),
				"createdByName": strPtr(getCol(row, "CREATED BY")),
			}
			entity := &model.MeetingItem{
				ID: id, Title: orDefault(getCol(row, "TITLE"), "Untitled"),
				Date: parseTime(getCol(row, "DATE")), Time: strPtr(getCol(row, "TIME")),
				Location: strPtr(getCol(row, "LOCATION")), Status: orDefault(getCol(row, "STATUS"), "scheduled"),
				Attendees: strPtr(getCol(row, "ATTENDEES")),
				VoiceNoteUrl: strPtr(getCol(row, "VOICE NOTE URL")),
				MomContent: strPtr(getCol(row, "MOM CONTENT")),
				AiSummary: strPtr(getCol(row, "AI SUMMARY")),
				CreatedByName: strPtr(getCol(row, "CREATED BY")), CreatedBy: "system", Version: 1,
			}
			return id, updates, entity
		},
	}
}

// ─── P&L Forecast (tab: "P&L Forecast") ──────────────────────

func buildPLForecastConfig() *EntityConfig {
	cols := []string{"ID", "Project / Buyer", "Quantity", "Selling Price", "Buying Price", "Freight Cost", "Other Cost", "Gross Profit / MT", "Total Gross Profit", "Updated At"}
	return &EntityConfig{
		Sheet: SheetConfig{SheetName: "P&L Forecast", Columns: cols, ColRange: "A:J"},
		ModelName: "PLForecast",
		NewModel:  func() interface{} { return &model.PLForecast{} },
		NewSlice:  func() interface{} { return &[]model.PLForecast{} },
		ToRow: func(entity interface{}) []interface{} {
			e := entity.(*model.PLForecast)
			return []interface{}{
				e.ID, fmtStr(e.Buyer), e.Quantity, e.SellingPrice,
				e.BuyingPrice, e.FreightCost, e.OtherCost,
				e.GrossProfitMt, e.TotalGrossProfit, fmtTime(e.UpdatedAt),
			}
		},
		ToUpdates: func(row map[string]string) (string, map[string]interface{}, interface{}) {
			id := getCol(row, "ID")
			buyer := getCol(row, "PROJECT / BUYER")
			if buyer == "" {
				buyer = getCol(row, "BUYER")
			}
			updates := map[string]interface{}{
				"buyer": strPtr(orDefault(buyer, "Unknown")),
				"quantity": parseFloat(getCol(row, "QUANTITY")),
				"sellingPrice": parseFloat(getCol(row, "SELLING PRICE")),
				"buyingPrice": parseFloat(getCol(row, "BUYING PRICE")),
				"freightCost": parseFloat(getCol(row, "FREIGHT COST")),
				"otherCost": parseFloat(getCol(row, "OTHER COST")),
				"grossProfitMt": parseFloat(getCol(row, "GROSS PROFIT / MT")),
				"totalGrossProfit": parseFloat(getCol(row, "TOTAL GROSS PROFIT")),
			}
			entity := &model.PLForecast{
				ID: id, Buyer: strPtr(orDefault(buyer, "Unknown")),
				Quantity: parseFloat(getCol(row, "QUANTITY")), SellingPrice: parseFloat(getCol(row, "SELLING PRICE")),
				BuyingPrice: parseFloat(getCol(row, "BUYING PRICE")), FreightCost: parseFloat(getCol(row, "FREIGHT COST")),
				OtherCost: parseFloat(getCol(row, "OTHER COST")), GrossProfitMt: parseFloat(getCol(row, "GROSS PROFIT / MT")),
				TotalGrossProfit: parseFloat(getCol(row, "TOTAL GROSS PROFIT")), Version: 1,
			}
			return id, updates, entity
		},
	}
}

// ─── Sales Deals (tab: "Projects") ────────────────────────────

func buildSalesDealConfig() *EntityConfig {
	cols := []string{"ID", "Buyer", "Country", "Type", "Quantity (MT)", "Price/MT", "Total Value", "Status", "Vessel", "Laycan Start", "Laycan End", "PIC", "Updated At"}
	return &EntityConfig{
		Sheet: SheetConfig{SheetName: "Projects", Columns: cols, ColRange: "A:M"},
		ModelName: "SalesDeal",
		NewModel:  func() interface{} { return &model.SalesDeal{} },
		NewSlice:  func() interface{} { return &[]model.SalesDeal{} },
		ToRow: func(entity interface{}) []interface{} {
			e := entity.(*model.SalesDeal)
			return []interface{}{
				e.ID, e.Buyer, fmtStr(e.BuyerCountry), e.Type,
				e.Quantity, fmtFloat(e.PricePerMt), fmtFloat(e.TotalValue),
				e.Status, fmtStr(e.VesselName), fmtTimePtr(e.LaycanStart),
				fmtTimePtr(e.LaycanEnd), fmtStr(e.PicName), fmtTime(e.UpdatedAt),
			}
		},
		ToUpdates: func(row map[string]string) (string, map[string]interface{}, interface{}) {
			id := getCol(row, "ID")
			dealNum := id
			if !strings.HasPrefix(id, "DEAL-") {
				dealNum = fmt.Sprintf("DEAL-%s", id)
			}
			updates := map[string]interface{}{
				"dealNumber":   dealNum,
				"status":        orDefault(getCol(row, "STATUS"), "confirmed"),
				"buyer":         orDefault(getCol(row, "BUYER"), "Unknown"),
				"buyerCountry": strPtr(getCol(row, "COUNTRY")),
				"type":          orDefault(getCol(row, "TYPE"), "export"),
				"quantity":      parseFloat(getCol(row, "QUANTITY (MT)")),
				"pricePerMt":  floatPtr(getCol(row, "PRICE/MT")),
				"totalValue":   floatPtr(getCol(row, "TOTAL VALUE")),
				"laycanStart":  parseTime(getCol(row, "LAYCAN START")),
				"laycanEnd":    parseTime(getCol(row, "LAYCAN END")),
				"picName":      strPtr(getCol(row, "PIC")),
				"vesselName":   strPtr(getCol(row, "VESSEL")),
			}
			entity := &model.SalesDeal{
				ID: id, DealNumber: dealNum, Buyer: orDefault(getCol(row, "BUYER"), "Unknown"),
				BuyerCountry: strPtr(getCol(row, "COUNTRY")), Type: orDefault(getCol(row, "TYPE"), "export"),
				Quantity: parseFloat(getCol(row, "QUANTITY (MT)")), PricePerMt: floatPtr(getCol(row, "PRICE/MT")),
				TotalValue: floatPtr(getCol(row, "TOTAL VALUE")), Status: orDefault(getCol(row, "STATUS"), "confirmed"),
				VesselName: strPtr(getCol(row, "VESSEL")), LaycanStart: parseTime(getCol(row, "LAYCAN START")),
				LaycanEnd: parseTime(getCol(row, "LAYCAN END")), PicName: strPtr(getCol(row, "PIC")),
				CreatedBy: "system", Version: 1,
			}
			return id, updates, entity
		},
	}
}

// ─── Partners (tab: "Partners") ───────────────────────────────

func buildPartnerConfig() *EntityConfig {
	cols := []string{"ID", "Name", "Type", "Category", "Contact Person", "Phone", "Email", "Address", "City", "Country", "Tax ID", "Status", "Notes", "Updated At"}
	return &EntityConfig{
		Sheet: SheetConfig{SheetName: "Partners", Columns: cols, ColRange: "A:N"},
		ModelName: "Partner",
		NewModel:  func() interface{} { return &model.Partner{} },
		NewSlice:  func() interface{} { return &[]model.Partner{} },
		ToRow: func(entity interface{}) []interface{} {
			e := entity.(*model.Partner)
			return []interface{}{
				e.ID, e.Name, e.Type, fmtStr(e.Category),
				fmtStr(e.ContactPerson), fmtStr(e.Phone), fmtStr(e.Email),
				fmtStr(e.Address), fmtStr(e.City), fmtStr(e.Country),
				fmtStr(e.TaxId), e.Status, fmtStr(e.Notes), fmtTime(e.UpdatedAt),
			}
		},
		ToUpdates: func(row map[string]string) (string, map[string]interface{}, interface{}) {
			id := getCol(row, "ID")
			updates := map[string]interface{}{
				"name": orDefault(getCol(row, "NAME"), "Unknown"), "type": orDefault(getCol(row, "TYPE"), "buyer"),
				"category": strPtr(getCol(row, "CATEGORY")), "contactPerson": strPtr(getCol(row, "CONTACT PERSON")),
				"phone": strPtr(getCol(row, "PHONE")), "email": strPtr(getCol(row, "EMAIL")),
				"address": strPtr(getCol(row, "ADDRESS")), "city": strPtr(getCol(row, "CITY")),
				"country": strPtr(getCol(row, "COUNTRY")), "taxId": strPtr(getCol(row, "TAX ID")),
				"status": orDefault(getCol(row, "STATUS"), "active"), "notes": strPtr(getCol(row, "NOTES")),
			}
			entity := &model.Partner{
				ID: id, Name: orDefault(getCol(row, "NAME"), "Unknown"), Type: orDefault(getCol(row, "TYPE"), "buyer"),
				Category: strPtr(getCol(row, "CATEGORY")), ContactPerson: strPtr(getCol(row, "CONTACT PERSON")),
				Phone: strPtr(getCol(row, "PHONE")), Email: strPtr(getCol(row, "EMAIL")),
				Address: strPtr(getCol(row, "ADDRESS")), City: strPtr(getCol(row, "CITY")),
				Country: strPtr(getCol(row, "COUNTRY")), TaxId: strPtr(getCol(row, "TAX ID")),
				Status: orDefault(getCol(row, "STATUS"), "active"), Notes: strPtr(getCol(row, "NOTES")),
				Version: 1,
			}
			return id, updates, entity
		},
	}
}

// ─── Blending (tab: "Blending") ───────────────────────────────

func buildBlendingConfig() *EntityConfig {
	cols := []string{"ID", "Inputs", "Total Qty", "Result GAR", "Result TS", "Result Ash", "Result TM", "Created By", "Created At"}
	return &EntityConfig{
		Sheet: SheetConfig{SheetName: "Blending", Columns: cols, ColRange: "A:I"},
		ModelName: "BlendingSimulation",
		NewModel:  func() interface{} { return &model.BlendingSimulation{} },
		NewSlice:  func() interface{} { return &[]model.BlendingSimulation{} },
		ToRow: func(entity interface{}) []interface{} {
			e := entity.(*model.BlendingSimulation)
			return []interface{}{
				e.ID, e.Inputs, e.TotalQuantity, e.ResultGar,
				e.ResultTs, e.ResultAsh, e.ResultTm,
				e.CreatedBy, fmtTime(e.CreatedAt),
			}
		},
		ToUpdates: func(row map[string]string) (string, map[string]interface{}, interface{}) {
			id := getCol(row, "ID")
			updates := map[string]interface{}{
				"inputs":         orDefault(getCol(row, "INPUTS"), "[]"),
				"totalQuantity": parseFloat(getCol(row, "TOTAL QTY")),
				"resultGar":     parseFloat(getCol(row, "RESULT GAR")),
				"resultTs":      parseFloat(getCol(row, "RESULT TS")),
				"resultAsh":     parseFloat(getCol(row, "RESULT ASH")),
				"resultTm":      parseFloat(getCol(row, "RESULT TM")),
				"createdBy":     orDefault(getCol(row, "CREATED BY"), "system"),
			}
			entity := &model.BlendingSimulation{
				ID: id, Inputs: orDefault(getCol(row, "INPUTS"), "[]"),
				TotalQuantity: parseFloat(getCol(row, "TOTAL QTY")),
				ResultGar: parseFloat(getCol(row, "RESULT GAR")), ResultTs: parseFloat(getCol(row, "RESULT TS")),
				ResultAsh: parseFloat(getCol(row, "RESULT ASH")), ResultTm: parseFloat(getCol(row, "RESULT TM")),
				CreatedBy: orDefault(getCol(row, "CREATED BY"), "system"), Version: 1,
			}
			return id, updates, entity
		},
	}
}

// ─── Utilities ──────────────────────────────────────────────────

func orDefault(val, def string) string {
	if val == "" {
		return def
	}
	return val
}

func normalizeShipmentStatus(val string) string {
	s := strings.ToLower(strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(val, " ", "_"), "-", "_")))
	switch s {
	case "waiting_for_loading", "waiting":
		return "waiting_loading"
	case "intransit", "transit":
		return "in_transit"
	case "discharged", "discharge":
		return "discharging"
	case "complete", "done":
		return "completed"
	case "cancel", "canceled":
		return "cancelled"
	}
	validStatuses := []string{"draft", "confirmed", "waiting_loading", "loading", "in_transit", "anchorage", "discharging", "completed", "cancelled"}
	for _, vs := range validStatuses {
		if s == vs {
			return s
		}
	}
	return "draft"
}
