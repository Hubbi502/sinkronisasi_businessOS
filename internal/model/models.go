package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─── Helpers ────────────────────────────────────────────────────

func generateUUID(id *string) {
	if *id == "" {
		*id = uuid.New().String()
	}
}

// ─── SalesOrder ─────────────────────────────────────────────────

type SalesOrder struct {
	ID            string    `gorm:"primaryKey;type:varchar(60)" json:"id"`
	OrderNumber   string    `gorm:"type:varchar(60);uniqueIndex" json:"order_number"`
	Client        string    `gorm:"type:varchar(255);not null" json:"client"`
	Description   *string   `gorm:"type:text" json:"description"`
	Amount        float64   `gorm:"type:decimal(15,2);not null;default:0" json:"amount"`
	Priority      string    `gorm:"type:varchar(20);not null;default:'medium'" json:"priority"`
	Status        string    `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	ImageUrl      *string   `gorm:"type:text" json:"image_url"`
	CreatedByName *string   `gorm:"type:varchar(255)" json:"created_by_name"`
	CreatedBy     string    `gorm:"type:varchar(60);not null" json:"created_by"`
	ApprovedBy    *string   `gorm:"type:varchar(60)" json:"approved_by"`
	Notes         *string   `gorm:"type:text" json:"notes"`
	Version       int       `gorm:"type:int;not null;default:1" json:"version"`
	IsDeleted     bool      `gorm:"type:bool;not null;default:false" json:"is_deleted"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (e *SalesOrder) BeforeCreate(tx *gorm.DB) error {
	generateUUID(&e.ID)
	return nil
}

// ─── PurchaseRequest ────────────────────────────────────────────

type PurchaseRequest struct {
	ID            string    `gorm:"primaryKey;type:varchar(60)" json:"id"`
	RequestNumber string    `gorm:"type:varchar(60);uniqueIndex" json:"request_number"`
	Category      string    `gorm:"type:varchar(100);not null" json:"category"`
	Supplier      *string   `gorm:"type:varchar(255)" json:"supplier"`
	Description   *string   `gorm:"type:text" json:"description"`
	Amount        float64   `gorm:"type:decimal(15,2);not null;default:0" json:"amount"`
	Priority      string    `gorm:"type:varchar(20);not null;default:'medium'" json:"priority"`
	Status        string    `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	ImageUrl      *string   `gorm:"type:text" json:"image_url"`
	CreatedByName *string   `gorm:"type:varchar(255)" json:"created_by_name"`
	CreatedBy     string    `gorm:"type:varchar(60);not null" json:"created_by"`
	ApprovedBy    *string   `gorm:"type:varchar(60)" json:"approved_by"`
	Notes         *string   `gorm:"type:text" json:"notes"`
	IsAnomaly     bool      `gorm:"type:bool;not null;default:false" json:"is_anomaly"`
	AnomalyReason *string   `gorm:"type:text" json:"anomaly_reason"`
	OcrData       *string   `gorm:"type:text" json:"ocr_data"`
	Version       int       `gorm:"type:int;not null;default:1" json:"version"`
	IsDeleted     bool      `gorm:"type:bool;not null;default:false" json:"is_deleted"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (e *PurchaseRequest) BeforeCreate(tx *gorm.DB) error {
	generateUUID(&e.ID)
	return nil
}

// ─── TaskItem ───────────────────────────────────────────────────

type TaskItem struct {
	ID           string     `gorm:"primaryKey;type:varchar(60)" json:"id"`
	Title        string     `gorm:"type:varchar(500);not null" json:"title"`
	Description  *string    `gorm:"type:text" json:"description"`
	Status       string     `gorm:"type:varchar(20);not null;default:'todo'" json:"status"`
	Priority     string     `gorm:"type:varchar(20);not null;default:'medium'" json:"priority"`
	AssigneeId   *string    `gorm:"type:varchar(60)" json:"assignee_id"`
	AssigneeName *string    `gorm:"type:varchar(255)" json:"assignee_name"`
	DueDate      *time.Time `gorm:"type:timestamptz" json:"due_date"`
	CreatedBy    string     `gorm:"type:varchar(60);not null" json:"created_by"`
	Version      int        `gorm:"type:int;not null;default:1" json:"version"`
	IsDeleted    bool       `gorm:"type:bool;not null;default:false" json:"is_deleted"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (e *TaskItem) BeforeCreate(tx *gorm.DB) error {
	generateUUID(&e.ID)
	return nil
}

// ─── ShipmentDetail ─────────────────────────────────────────────

type ShipmentDetail struct {
	ID             string     `gorm:"primaryKey;type:varchar(60)" json:"id"`
	ShipmentNumber string     `gorm:"type:varchar(60);uniqueIndex" json:"shipment_number"`
	DealId         *string    `gorm:"type:varchar(60)" json:"deal_id"`
	Status         string     `gorm:"type:varchar(20);not null;default:'draft'" json:"status"`
	Buyer          string     `gorm:"type:varchar(255);not null" json:"buyer"`
	Supplier       *string    `gorm:"type:varchar(255)" json:"supplier"`
	IsBlending     bool       `gorm:"type:bool;not null;default:false" json:"is_blending"`
	IupOp          *string    `gorm:"type:varchar(255)" json:"iup_op"`
	VesselName     *string    `gorm:"type:varchar(255)" json:"vessel_name"`
	BargeName      *string    `gorm:"type:varchar(255)" json:"barge_name"`
	LoadingPort    *string    `gorm:"type:varchar(255)" json:"loading_port"`
	DischargePort  *string    `gorm:"type:varchar(255)" json:"discharge_port"`
	QuantityLoaded *float64   `gorm:"type:decimal(15,2)" json:"quantity_loaded"`
	BlDate         *time.Time `gorm:"type:timestamptz" json:"bl_date"`
	Eta            *time.Time `gorm:"type:timestamptz" json:"eta"`
	SalesPrice     *float64   `gorm:"type:decimal(15,2)" json:"sales_price"`
	MarginMt       *float64   `gorm:"type:decimal(15,2)" json:"margin_mt"`
	PicName        *string    `gorm:"type:varchar(255)" json:"pic_name"`
	Type           string     `gorm:"type:varchar(20);not null;default:'export'" json:"type"`
	Milestones     *string    `gorm:"type:text" json:"milestones"`
	Version        int        `gorm:"type:int;not null;default:1" json:"version"`
	IsDeleted      bool       `gorm:"type:bool;not null;default:false" json:"is_deleted"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (e *ShipmentDetail) BeforeCreate(tx *gorm.DB) error {
	generateUUID(&e.ID)
	return nil
}

// ─── SourceSupplier ─────────────────────────────────────────────

type SourceSupplier struct {
	ID               string    `gorm:"primaryKey;type:varchar(60)" json:"id"`
	Name             string    `gorm:"type:varchar(255);not null" json:"name"`
	Region           string    `gorm:"type:varchar(255);not null" json:"region"`
	CalorieRange     *string   `gorm:"type:varchar(100)" json:"calorie_range"`
	Gar              *float64  `gorm:"type:decimal(10,2)" json:"gar"`
	Ts               *float64  `gorm:"type:decimal(10,2)" json:"ts"`
	Ash              *float64  `gorm:"type:decimal(10,2)" json:"ash"`
	Tm               *float64  `gorm:"type:decimal(10,2)" json:"tm"`
	Im               *float64  `gorm:"type:decimal(10,2)" json:"im"`
	Fc               *float64  `gorm:"type:decimal(10,2)" json:"fc"`
	Nar              *float64  `gorm:"type:decimal(10,2)" json:"nar"`
	Adb              *float64  `gorm:"type:decimal(10,2)" json:"adb"`
	JettyPort        *string   `gorm:"type:varchar(255)" json:"jetty_port"`
	Anchorage        *string   `gorm:"type:varchar(255)" json:"anchorage"`
	StockAvailable   float64   `gorm:"type:decimal(15,2);not null;default:0" json:"stock_available"`
	MinStockAlert    *float64  `gorm:"type:decimal(15,2)" json:"min_stock_alert"`
	KycStatus        string    `gorm:"type:varchar(20);not null;default:'not_started'" json:"kyc_status"`
	PsiStatus        string    `gorm:"type:varchar(20);not null;default:'not_started'" json:"psi_status"`
	FobBargeOnly     bool      `gorm:"type:bool;not null;default:false" json:"fob_barge_only"`
	PriceLinkedIndex *string   `gorm:"type:varchar(255)" json:"price_linked_index"`
	FobBargePriceUsd *float64  `gorm:"type:decimal(15,2)" json:"fob_barge_price_usd"`
	ContractType     *string   `gorm:"type:varchar(100)" json:"contract_type"`
	PicName          *string   `gorm:"type:varchar(255)" json:"pic_name"`
	IupNumber        *string   `gorm:"type:varchar(255)" json:"iup_number"`
	Version          int       `gorm:"type:int;not null;default:1" json:"version"`
	IsDeleted        bool      `gorm:"type:bool;not null;default:false" json:"is_deleted"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (e *SourceSupplier) BeforeCreate(tx *gorm.DB) error {
	generateUUID(&e.ID)
	return nil
}

// ─── QualityResult ──────────────────────────────────────────────

type QualityResult struct {
	ID           string     `gorm:"primaryKey;type:varchar(60)" json:"id"`
	CargoId      string     `gorm:"type:varchar(60);not null" json:"cargo_id"`
	CargoName    string     `gorm:"type:varchar(255);not null" json:"cargo_name"`
	Surveyor     *string    `gorm:"type:varchar(255)" json:"surveyor"`
	SamplingDate *time.Time `gorm:"type:timestamptz" json:"sampling_date"`
	Gar          *float64   `gorm:"type:decimal(10,2)" json:"gar"`
	Ts           *float64   `gorm:"type:decimal(10,2)" json:"ts"`
	Ash          *float64   `gorm:"type:decimal(10,2)" json:"ash"`
	Tm           *float64   `gorm:"type:decimal(10,2)" json:"tm"`
	Status       string     `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	Version      int        `gorm:"type:int;not null;default:1" json:"version"`
	IsDeleted    bool       `gorm:"type:bool;not null;default:false" json:"is_deleted"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (e *QualityResult) BeforeCreate(tx *gorm.DB) error {
	generateUUID(&e.ID)
	return nil
}

// ─── MarketPrice ────────────────────────────────────────────────

type MarketPrice struct {
	ID        string    `gorm:"primaryKey;type:varchar(60)" json:"id"`
	Date      time.Time `gorm:"type:timestamptz;not null" json:"date"`
	Ici1      *float64  `gorm:"type:decimal(10,2)" json:"ici1"`
	Ici2      *float64  `gorm:"type:decimal(10,2)" json:"ici2"`
	Ici3      *float64  `gorm:"type:decimal(10,2)" json:"ici3"`
	Ici4      *float64  `gorm:"type:decimal(10,2)" json:"ici4"`
	Ici5      *float64  `gorm:"type:decimal(10,2)" json:"ici5"`
	Newcastle *float64  `gorm:"type:decimal(10,2)" json:"newcastle"`
	Hba       *float64  `gorm:"type:decimal(10,2)" json:"hba"`
	Source    *string   `gorm:"type:text" json:"source"`
	Version   int       `gorm:"type:int;not null;default:1" json:"version"`
	IsDeleted bool      `gorm:"type:bool;not null;default:false" json:"is_deleted"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (e *MarketPrice) BeforeCreate(tx *gorm.DB) error {
	generateUUID(&e.ID)
	return nil
}

// ─── MeetingItem ────────────────────────────────────────────────

type MeetingItem struct {
	ID            string     `gorm:"primaryKey;type:varchar(60)" json:"id"`
	Title         string     `gorm:"type:varchar(500);not null" json:"title"`
	Date          *time.Time `gorm:"type:timestamptz" json:"date"`
	Time          *string    `gorm:"type:varchar(20)" json:"time"`
	Location      *string    `gorm:"type:varchar(255)" json:"location"`
	Status        string     `gorm:"type:varchar(20);not null;default:'scheduled'" json:"status"`
	Attendees     *string    `gorm:"type:text" json:"attendees"`
	MomContent    *string    `gorm:"type:text" json:"mom_content"`
	VoiceNoteUrl  *string    `gorm:"type:text" json:"voice_note_url"`
	AiSummary     *string    `gorm:"type:text" json:"ai_summary"`
	CreatedByName *string    `gorm:"type:varchar(255)" json:"created_by_name"`
	CreatedBy     string     `gorm:"type:varchar(60);not null" json:"created_by"`
	Version       int        `gorm:"type:int;not null;default:1" json:"version"`
	IsDeleted     bool       `gorm:"type:bool;not null;default:false" json:"is_deleted"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (e *MeetingItem) BeforeCreate(tx *gorm.DB) error {
	generateUUID(&e.ID)
	return nil
}

// ─── PLForecast ─────────────────────────────────────────────────

type PLForecast struct {
	ID               string    `gorm:"primaryKey;type:varchar(60)" json:"id"`
	DealId           *string   `gorm:"type:varchar(60)" json:"deal_id"`
	DealNumber       *string   `gorm:"type:varchar(60)" json:"deal_number"`
	ProjectName      *string   `gorm:"type:varchar(255)" json:"project_name"`
	Buyer            *string   `gorm:"type:varchar(255)" json:"buyer"`
	Type             string    `gorm:"type:varchar(20);not null;default:'export'" json:"type"`
	Status           string    `gorm:"type:varchar(20);not null;default:'forecast'" json:"status"`
	Quantity         float64   `gorm:"type:decimal(15,2);not null;default:0" json:"quantity"`
	SellingPrice     float64   `gorm:"type:decimal(15,2);not null;default:0" json:"selling_price"`
	BuyingPrice      float64   `gorm:"type:decimal(15,2);not null;default:0" json:"buying_price"`
	FreightCost      float64   `gorm:"type:decimal(15,2);not null;default:0" json:"freight_cost"`
	OtherCost        float64   `gorm:"type:decimal(15,2);not null;default:0" json:"other_cost"`
	GrossProfitMt    float64   `gorm:"type:decimal(15,2);not null;default:0" json:"gross_profit_mt"`
	TotalGrossProfit float64   `gorm:"type:decimal(15,2);not null;default:0" json:"total_gross_profit"`
	CreatedBy        *string   `gorm:"type:varchar(60)" json:"created_by"`
	Version          int       `gorm:"type:int;not null;default:1" json:"version"`
	IsDeleted        bool      `gorm:"type:bool;not null;default:false" json:"is_deleted"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (e *PLForecast) BeforeCreate(tx *gorm.DB) error {
	generateUUID(&e.ID)
	return nil
}

// ─── SalesDeal ──────────────────────────────────────────────────

type SalesDeal struct {
	ID            string     `gorm:"primaryKey;type:varchar(60)" json:"id"`
	DealNumber    string     `gorm:"type:varchar(60);uniqueIndex" json:"deal_number"`
	Status        string     `gorm:"type:varchar(20);not null;default:'pre_sale'" json:"status"`
	Buyer         string     `gorm:"type:varchar(255);not null" json:"buyer"`
	BuyerCountry  *string    `gorm:"type:varchar(100)" json:"buyer_country"`
	Type          string     `gorm:"type:varchar(20);not null;default:'export'" json:"type"`
	ShippingTerms string     `gorm:"type:varchar(20);not null;default:'FOB'" json:"shipping_terms"`
	Quantity      float64    `gorm:"type:decimal(15,2);not null;default:0" json:"quantity"`
	PricePerMt    *float64   `gorm:"type:decimal(15,2)" json:"price_per_mt"`
	TotalValue    *float64   `gorm:"type:decimal(15,2)" json:"total_value"`
	LaycanStart   *time.Time `gorm:"type:timestamptz" json:"laycan_start"`
	LaycanEnd     *time.Time `gorm:"type:timestamptz" json:"laycan_end"`
	VesselName    *string    `gorm:"type:varchar(255)" json:"vessel_name"`
	Gar           *float64   `gorm:"type:decimal(10,2)" json:"gar"`
	Ts            *float64   `gorm:"type:decimal(10,2)" json:"ts"`
	Ash           *float64   `gorm:"type:decimal(10,2)" json:"ash"`
	Tm            *float64   `gorm:"type:decimal(10,2)" json:"tm"`
	ProjectId     *string    `gorm:"type:varchar(60)" json:"project_id"`
	PicId         *string    `gorm:"type:varchar(60)" json:"pic_id"`
	PicName       *string    `gorm:"type:varchar(255)" json:"pic_name"`
	CreatedByName *string    `gorm:"type:varchar(255)" json:"created_by_name"`
	CreatedBy     string     `gorm:"type:varchar(60);not null" json:"created_by"`
	Version       int        `gorm:"type:int;not null;default:1" json:"version"`
	IsDeleted     bool       `gorm:"type:bool;not null;default:false" json:"is_deleted"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (e *SalesDeal) BeforeCreate(tx *gorm.DB) error {
	generateUUID(&e.ID)
	return nil
}

// ─── Partner ────────────────────────────────────────────────────

type Partner struct {
	ID            string    `gorm:"primaryKey;type:varchar(60)" json:"id"`
	Name          string    `gorm:"type:varchar(255);not null" json:"name"`
	Type          string    `gorm:"type:varchar(20);not null;default:'buyer'" json:"type"`
	Category      *string   `gorm:"type:varchar(100)" json:"category"`
	ContactPerson *string   `gorm:"type:varchar(255)" json:"contact_person"`
	Phone         *string   `gorm:"type:varchar(50)" json:"phone"`
	Email         *string   `gorm:"type:varchar(255)" json:"email"`
	Address       *string   `gorm:"type:text" json:"address"`
	City          *string   `gorm:"type:varchar(100)" json:"city"`
	Country       *string   `gorm:"type:varchar(100)" json:"country"`
	TaxId         *string   `gorm:"type:varchar(100)" json:"tax_id"`
	Status        string    `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	Notes         *string   `gorm:"type:text" json:"notes"`
	Version       int       `gorm:"type:int;not null;default:1" json:"version"`
	IsDeleted     bool      `gorm:"type:bool;not null;default:false" json:"is_deleted"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (e *Partner) BeforeCreate(tx *gorm.DB) error {
	generateUUID(&e.ID)
	return nil
}

// ─── BlendingSimulation ─────────────────────────────────────────

type BlendingSimulation struct {
	ID            string    `gorm:"primaryKey;type:varchar(60)" json:"id"`
	Inputs        string    `gorm:"type:text;not null;default:'[]'" json:"inputs"`
	TotalQuantity float64   `gorm:"type:decimal(15,2);not null;default:0" json:"total_quantity"`
	ResultGar     float64   `gorm:"type:decimal(10,2);not null;default:0" json:"result_gar"`
	ResultTs      float64   `gorm:"type:decimal(10,2);not null;default:0" json:"result_ts"`
	ResultAsh     float64   `gorm:"type:decimal(10,2);not null;default:0" json:"result_ash"`
	ResultTm      float64   `gorm:"type:decimal(10,2);not null;default:0" json:"result_tm"`
	CreatedBy     string    `gorm:"type:varchar(60);not null" json:"created_by"`
	Version       int       `gorm:"type:int;not null;default:1" json:"version"`
	IsDeleted     bool      `gorm:"type:bool;not null;default:false" json:"is_deleted"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (e *BlendingSimulation) BeforeCreate(tx *gorm.DB) error {
	generateUUID(&e.ID)
	return nil
}

// ─── AllModels returns all GORM models for auto-migration ──────

func AllModels() []interface{} {
	return []interface{}{
		&SalesOrder{},
		&PurchaseRequest{},
		&TaskItem{},
		&ShipmentDetail{},
		&SourceSupplier{},
		&QualityResult{},
		&MarketPrice{},
		&MeetingItem{},
		&PLForecast{},
		&SalesDeal{},
		&Partner{},
		&BlendingSimulation{},
	}
}
