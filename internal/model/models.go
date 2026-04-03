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
	ID            string    `gorm:"column:id;primaryKey;type:varchar(60)" json:"id"`
	OrderNumber   string    `gorm:"column:orderNumber;type:varchar(60);uniqueIndex" json:"order_number"`
	Client        string    `gorm:"column:client;type:varchar(255);not null" json:"client"`
	Description   *string   `gorm:"column:description;type:text" json:"description"`
	Amount        float64   `gorm:"column:amount;type:decimal(15,2);not null;default:0" json:"amount"`
	Priority      string    `gorm:"column:priority;type:varchar(20);not null;default:'medium'" json:"priority"`
	Status        string    `gorm:"column:status;type:varchar(20);not null;default:'pending'" json:"status"`
	ImageUrl      *string   `gorm:"column:imageUrl;type:text" json:"image_url"`
	CreatedByName *string   `gorm:"column:createdByName;type:varchar(255)" json:"created_by_name"`
	CreatedBy     string    `gorm:"column:createdBy;type:varchar(60);not null" json:"created_by"`
	ApprovedBy    *string   `gorm:"column:approvedBy;type:varchar(60)" json:"approved_by"`
	Notes         *string   `gorm:"column:notes;type:text" json:"notes"`
	Version       int       `gorm:"column:version;type:int;not null;default:1" json:"version"`
	IsDeleted     bool      `gorm:"column:isDeleted;type:bool;not null;default:false" json:"is_deleted"`
	CreatedAt     time.Time `gorm:"column:createdAt;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updated_at"`
}

func (e *SalesOrder) BeforeCreate(tx *gorm.DB) error {
	generateUUID(&e.ID)
	return nil
}

func (SalesOrder) TableName() string {
	return "SalesOrder"
}

// ─── PurchaseRequest ────────────────────────────────────────────

type PurchaseRequest struct {
	ID            string    `gorm:"column:id;primaryKey;type:varchar(60)" json:"id"`
	RequestNumber string    `gorm:"column:requestNumber;type:varchar(60);uniqueIndex" json:"request_number"`
	Category      string    `gorm:"column:category;type:varchar(100);not null" json:"category"`
	Supplier      *string   `gorm:"column:supplier;type:varchar(255)" json:"supplier"`
	Description   *string   `gorm:"column:description;type:text" json:"description"`
	Amount        float64   `gorm:"column:amount;type:decimal(15,2);not null;default:0" json:"amount"`
	Priority      string    `gorm:"column:priority;type:varchar(20);not null;default:'medium'" json:"priority"`
	Status        string    `gorm:"column:status;type:varchar(20);not null;default:'pending'" json:"status"`
	ImageUrl      *string   `gorm:"column:imageUrl;type:text" json:"image_url"`
	CreatedByName *string   `gorm:"column:createdByName;type:varchar(255)" json:"created_by_name"`
	CreatedBy     string    `gorm:"column:createdBy;type:varchar(60);not null" json:"created_by"`
	ApprovedBy    *string   `gorm:"column:approvedBy;type:varchar(60)" json:"approved_by"`
	Notes         *string   `gorm:"column:notes;type:text" json:"notes"`
	IsAnomaly     bool      `gorm:"column:isAnomaly;type:bool;not null;default:false" json:"is_anomaly"`
	AnomalyReason *string   `gorm:"column:anomalyReason;type:text" json:"anomaly_reason"`
	OcrData       *string   `gorm:"column:ocrData;type:text" json:"ocr_data"`
	Version       int       `gorm:"column:version;type:int;not null;default:1" json:"version"`
	IsDeleted     bool      `gorm:"column:isDeleted;type:bool;not null;default:false" json:"is_deleted"`
	CreatedAt     time.Time `gorm:"column:createdAt;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updated_at"`
}

func (e *PurchaseRequest) BeforeCreate(tx *gorm.DB) error {
	generateUUID(&e.ID)
	return nil
}

func (PurchaseRequest) TableName() string {
	return "PurchaseRequest"
}

// ─── TaskItem ───────────────────────────────────────────────────

type TaskItem struct {
	ID           string     `gorm:"column:id;primaryKey;type:varchar(60)" json:"id"`
	Title        string     `gorm:"column:title;type:varchar(500);not null" json:"title"`
	Description  *string    `gorm:"column:description;type:text" json:"description"`
	Status       string     `gorm:"column:status;type:varchar(20);not null;default:'todo'" json:"status"`
	Priority     string     `gorm:"column:priority;type:varchar(20);not null;default:'medium'" json:"priority"`
	AssigneeId   *string    `gorm:"column:assigneeId;type:varchar(60)" json:"assignee_id"`
	AssigneeName *string    `gorm:"column:assigneeName;type:varchar(255)" json:"assignee_name"`
	DueDate      *time.Time `gorm:"column:dueDate;type:timestamptz" json:"due_date"`
	CreatedBy    string     `gorm:"column:createdBy;type:varchar(60);not null" json:"created_by"`
	Version      int        `gorm:"column:version;type:int;not null;default:1" json:"version"`
	IsDeleted    bool       `gorm:"column:isDeleted;type:bool;not null;default:false" json:"is_deleted"`
	CreatedAt    time.Time  `gorm:"column:createdAt;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updatedAt;autoUpdateTime" json:"updated_at"`
}

func (e *TaskItem) BeforeCreate(tx *gorm.DB) error {
	generateUUID(&e.ID)
	return nil
}

func (TaskItem) TableName() string {
	return "TaskItem"
}

// ─── ShipmentDetail ─────────────────────────────────────────────

type ShipmentDetail struct {
	ID             string     `gorm:"column:id;primaryKey;type:varchar(60)" json:"id"`
	ShipmentNumber string     `gorm:"column:shipmentNumber;type:varchar(60);uniqueIndex" json:"shipment_number"`
	DealId         *string    `gorm:"column:dealId;type:varchar(60)" json:"deal_id"`
	Status         string     `gorm:"column:status;type:varchar(20);not null;default:'draft'" json:"status"`
	Buyer          string     `gorm:"column:buyer;type:varchar(255);not null" json:"buyer"`
	Supplier       *string    `gorm:"column:supplier;type:varchar(255)" json:"supplier"`
	IsBlending     bool       `gorm:"column:isBlending;type:bool;not null;default:false" json:"is_blending"`
	IupOp          *string    `gorm:"column:iupOp;type:varchar(255)" json:"iup_op"`
	VesselName     *string    `gorm:"column:vesselName;type:varchar(255)" json:"vessel_name"`
	BargeName      *string    `gorm:"column:bargeName;type:varchar(255)" json:"barge_name"`
	LoadingPort    *string    `gorm:"column:loadingPort;type:varchar(255)" json:"loading_port"`
	DischargePort  *string    `gorm:"column:dischargePort;type:varchar(255)" json:"discharge_port"`
	QuantityLoaded *float64   `gorm:"column:quantityLoaded;type:decimal(15,2)" json:"quantity_loaded"`
	BlDate         *time.Time `gorm:"column:blDate;type:timestamptz" json:"bl_date"`
	Eta            *time.Time `gorm:"column:eta;type:timestamptz" json:"eta"`
	SalesPrice     *float64   `gorm:"column:salesPrice;type:decimal(15,2)" json:"sales_price"`
	MarginMt       *float64   `gorm:"column:marginMt;type:decimal(15,2)" json:"margin_mt"`
	PicName        *string    `gorm:"column:picName;type:varchar(255)" json:"pic_name"`
	Type           string     `gorm:"column:type;type:varchar(20);not null;default:'export'" json:"type"`
	Milestones     *string    `gorm:"column:milestones;type:text" json:"milestones"`
	Version        int        `gorm:"column:version;type:int;not null;default:1" json:"version"`
	IsDeleted      bool       `gorm:"column:isDeleted;type:bool;not null;default:false" json:"is_deleted"`
	CreatedAt      time.Time  `gorm:"column:createdAt;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"column:updatedAt;autoUpdateTime" json:"updated_at"`
}

func (e *ShipmentDetail) BeforeCreate(tx *gorm.DB) error {
	generateUUID(&e.ID)
	return nil
}

func (ShipmentDetail) TableName() string {
	return "ShipmentDetail"
}

// ─── SourceSupplier ─────────────────────────────────────────────

type SourceSupplier struct {
	ID               string    `gorm:"column:id;primaryKey;type:varchar(60)" json:"id"`
	Name             string    `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Region           string    `gorm:"column:region;type:varchar(255);not null" json:"region"`
	CalorieRange     *string   `gorm:"column:calorieRange;type:varchar(100)" json:"calorie_range"`
	Gar              *float64  `gorm:"column:gar;type:decimal(10,2)" json:"gar"`
	Ts               *float64  `gorm:"column:ts;type:decimal(10,2)" json:"ts"`
	Ash              *float64  `gorm:"column:ash;type:decimal(10,2)" json:"ash"`
	Tm               *float64  `gorm:"column:tm;type:decimal(10,2)" json:"tm"`
	Im               *float64  `gorm:"column:im;type:decimal(10,2)" json:"im"`
	Fc               *float64  `gorm:"column:fc;type:decimal(10,2)" json:"fc"`
	Nar              *float64  `gorm:"column:nar;type:decimal(10,2)" json:"nar"`
	Adb              *float64  `gorm:"column:adb;type:decimal(10,2)" json:"adb"`
	JettyPort        *string   `gorm:"column:jettyPort;type:varchar(255)" json:"jetty_port"`
	Anchorage        *string   `gorm:"column:anchorage;type:varchar(255)" json:"anchorage"`
	StockAvailable   float64   `gorm:"column:stockAvailable;type:decimal(15,2);not null;default:0" json:"stock_available"`
	MinStockAlert    *float64  `gorm:"column:minStockAlert;type:decimal(15,2)" json:"min_stock_alert"`
	KycStatus        string    `gorm:"column:kycStatus;type:varchar(20);not null;default:'not_started'" json:"kyc_status"`
	PsiStatus        string    `gorm:"column:psiStatus;type:varchar(20);not null;default:'not_started'" json:"psi_status"`
	FobBargeOnly     bool      `gorm:"column:fobBargeOnly;type:bool;not null;default:false" json:"fob_barge_only"`
	PriceLinkedIndex *string   `gorm:"column:priceLinkedIndex;type:varchar(255)" json:"price_linked_index"`
	FobBargePriceUsd *float64  `gorm:"column:fobBargePriceUsd;type:decimal(15,2)" json:"fob_barge_price_usd"`
	ContractType     *string   `gorm:"column:contractType;type:varchar(100)" json:"contract_type"`
	PicName          *string   `gorm:"column:picName;type:varchar(255)" json:"pic_name"`
	IupNumber        *string   `gorm:"column:iupNumber;type:varchar(255)" json:"iup_number"`
	Version          int       `gorm:"column:version;type:int;not null;default:1" json:"version"`
	IsDeleted        bool      `gorm:"column:isDeleted;type:bool;not null;default:false" json:"is_deleted"`
	CreatedAt        time.Time `gorm:"column:createdAt;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updated_at"`
}

func (e *SourceSupplier) BeforeCreate(tx *gorm.DB) error {
	generateUUID(&e.ID)
	return nil
}

func (SourceSupplier) TableName() string {
	return "SourceSupplier"
}

// ─── QualityResult ──────────────────────────────────────────────

type QualityResult struct {
	ID           string     `gorm:"column:id;primaryKey;type:varchar(60)" json:"id"`
	CargoId      string     `gorm:"column:cargoId;type:varchar(60);not null" json:"cargo_id"`
	CargoName    string     `gorm:"column:cargoName;type:varchar(255);not null" json:"cargo_name"`
	Surveyor     *string    `gorm:"column:surveyor;type:varchar(255)" json:"surveyor"`
	SamplingDate *time.Time `gorm:"column:samplingDate;type:timestamptz" json:"sampling_date"`
	Gar          *float64   `gorm:"column:gar;type:decimal(10,2)" json:"gar"`
	Ts           *float64   `gorm:"column:ts;type:decimal(10,2)" json:"ts"`
	Ash          *float64   `gorm:"column:ash;type:decimal(10,2)" json:"ash"`
	Tm           *float64   `gorm:"column:tm;type:decimal(10,2)" json:"tm"`
	Status       string     `gorm:"column:status;type:varchar(20);not null;default:'pending'" json:"status"`
	Version      int        `gorm:"column:version;type:int;not null;default:1" json:"version"`
	IsDeleted    bool       `gorm:"column:isDeleted;type:bool;not null;default:false" json:"is_deleted"`
	CreatedAt    time.Time  `gorm:"column:createdAt;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updatedAt;autoUpdateTime" json:"updated_at"`
}

func (e *QualityResult) BeforeCreate(tx *gorm.DB) error {
	generateUUID(&e.ID)
	return nil
}

func (QualityResult) TableName() string {
	return "QualityResult"
}

// ─── MarketPrice ────────────────────────────────────────────────

type MarketPrice struct {
	ID        string    `gorm:"column:id;primaryKey;type:varchar(60)" json:"id"`
	Date      time.Time `gorm:"column:date;type:timestamptz;not null" json:"date"`
	Ici1      *float64  `gorm:"column:ici1;type:decimal(10,2)" json:"ici1"`
	Ici2      *float64  `gorm:"column:ici2;type:decimal(10,2)" json:"ici2"`
	Ici3      *float64  `gorm:"column:ici3;type:decimal(10,2)" json:"ici3"`
	Ici4      *float64  `gorm:"column:ici4;type:decimal(10,2)" json:"ici4"`
	Ici5      *float64  `gorm:"column:ici5;type:decimal(10,2)" json:"ici5"`
	Newcastle *float64  `gorm:"column:newcastle;type:decimal(10,2)" json:"newcastle"`
	Hba       *float64  `gorm:"column:hba;type:decimal(10,2)" json:"hba"`
	Source    *string   `gorm:"column:source;type:text" json:"source"`
	Version   int       `gorm:"column:version;type:int;not null;default:1" json:"version"`
	IsDeleted bool      `gorm:"column:isDeleted;type:bool;not null;default:false" json:"is_deleted"`
	CreatedAt time.Time `gorm:"column:createdAt;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updated_at"`
}

func (e *MarketPrice) BeforeCreate(tx *gorm.DB) error {
	generateUUID(&e.ID)
	return nil
}

func (MarketPrice) TableName() string {
	return "MarketPrice"
}

// ─── MeetingItem ────────────────────────────────────────────────

type MeetingItem struct {
	ID            string     `gorm:"column:id;primaryKey;type:varchar(60)" json:"id"`
	Title         string     `gorm:"column:title;type:varchar(500);not null" json:"title"`
	Date          *time.Time `gorm:"column:date;type:timestamptz" json:"date"`
	Time          *string    `gorm:"column:time;type:varchar(20)" json:"time"`
	Location      *string    `gorm:"column:location;type:varchar(255)" json:"location"`
	Status        string     `gorm:"column:status;type:varchar(20);not null;default:'scheduled'" json:"status"`
	Attendees     *string    `gorm:"column:attendees;type:text" json:"attendees"`
	MomContent    *string    `gorm:"column:momContent;type:text" json:"mom_content"`
	VoiceNoteUrl  *string    `gorm:"column:voiceNoteUrl;type:text" json:"voice_note_url"`
	AiSummary     *string    `gorm:"column:aiSummary;type:text" json:"ai_summary"`
	CreatedByName *string    `gorm:"column:createdByName;type:varchar(255)" json:"created_by_name"`
	CreatedBy     string     `gorm:"column:createdBy;type:varchar(60);not null" json:"created_by"`
	Version       int        `gorm:"column:version;type:int;not null;default:1" json:"version"`
	IsDeleted     bool       `gorm:"column:isDeleted;type:bool;not null;default:false" json:"is_deleted"`
	CreatedAt     time.Time  `gorm:"column:createdAt;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updatedAt;autoUpdateTime" json:"updated_at"`
}

func (e *MeetingItem) BeforeCreate(tx *gorm.DB) error {
	generateUUID(&e.ID)
	return nil
}

func (MeetingItem) TableName() string {
	return "MeetingItem"
}

// ─── PLForecast ─────────────────────────────────────────────────

type PLForecast struct {
	ID               string    `gorm:"column:id;primaryKey;type:varchar(60)" json:"id"`
	DealId           *string   `gorm:"column:dealId;type:varchar(60)" json:"deal_id"`
	DealNumber       *string   `gorm:"column:dealNumber;type:varchar(60)" json:"deal_number"`
	ProjectName      *string   `gorm:"column:projectName;type:varchar(255)" json:"project_name"`
	Buyer            *string   `gorm:"column:buyer;type:varchar(255)" json:"buyer"`
	Type             string    `gorm:"column:type;type:varchar(20);not null;default:'export'" json:"type"`
	Status           string    `gorm:"column:status;type:varchar(20);not null;default:'forecast'" json:"status"`
	Quantity         float64   `gorm:"column:quantity;type:decimal(15,2);not null;default:0" json:"quantity"`
	SellingPrice     float64   `gorm:"column:sellingPrice;type:decimal(15,2);not null;default:0" json:"selling_price"`
	BuyingPrice      float64   `gorm:"column:buyingPrice;type:decimal(15,2);not null;default:0" json:"buying_price"`
	FreightCost      float64   `gorm:"column:freightCost;type:decimal(15,2);not null;default:0" json:"freight_cost"`
	OtherCost        float64   `gorm:"column:otherCost;type:decimal(15,2);not null;default:0" json:"other_cost"`
	GrossProfitMt    float64   `gorm:"column:grossProfitMt;type:decimal(15,2);not null;default:0" json:"gross_profit_mt"`
	TotalGrossProfit float64   `gorm:"column:totalGrossProfit;type:decimal(15,2);not null;default:0" json:"total_gross_profit"`
	CreatedBy        *string   `gorm:"column:createdBy;type:varchar(60)" json:"created_by"`
	Version          int       `gorm:"column:version;type:int;not null;default:1" json:"version"`
	IsDeleted        bool      `gorm:"column:isDeleted;type:bool;not null;default:false" json:"is_deleted"`
	CreatedAt        time.Time `gorm:"column:createdAt;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updated_at"`
}

func (e *PLForecast) BeforeCreate(tx *gorm.DB) error {
	generateUUID(&e.ID)
	return nil
}

func (PLForecast) TableName() string {
	return "PLForecast"
}

// ─── SalesDeal ──────────────────────────────────────────────────

type SalesDeal struct {
	ID            string     `gorm:"column:id;primaryKey;type:varchar(60)" json:"id"`
	DealNumber    string     `gorm:"column:dealNumber;type:varchar(60);uniqueIndex" json:"deal_number"`
	Status        string     `gorm:"column:status;type:varchar(20);not null;default:'pre_sale'" json:"status"`
	Buyer         string     `gorm:"column:buyer;type:varchar(255);not null" json:"buyer"`
	BuyerCountry  *string    `gorm:"column:buyerCountry;type:varchar(100)" json:"buyer_country"`
	Type          string     `gorm:"column:type;type:varchar(20);not null;default:'export'" json:"type"`
	ShippingTerms string     `gorm:"column:shippingTerms;type:varchar(20);not null;default:'FOB'" json:"shipping_terms"`
	Quantity      float64    `gorm:"column:quantity;type:decimal(15,2);not null;default:0" json:"quantity"`
	PricePerMt    *float64   `gorm:"column:pricePerMt;type:decimal(15,2)" json:"price_per_mt"`
	TotalValue    *float64   `gorm:"column:totalValue;type:decimal(15,2)" json:"total_value"`
	LaycanStart   *time.Time `gorm:"column:laycanStart;type:timestamptz" json:"laycan_start"`
	LaycanEnd     *time.Time `gorm:"column:laycanEnd;type:timestamptz" json:"laycan_end"`
	VesselName    *string    `gorm:"column:vesselName;type:varchar(255)" json:"vessel_name"`
	Gar           *float64   `gorm:"column:gar;type:decimal(10,2)" json:"gar"`
	Ts            *float64   `gorm:"column:ts;type:decimal(10,2)" json:"ts"`
	Ash           *float64   `gorm:"column:ash;type:decimal(10,2)" json:"ash"`
	Tm            *float64   `gorm:"column:tm;type:decimal(10,2)" json:"tm"`
	ProjectId     *string    `gorm:"column:projectId;type:varchar(60)" json:"project_id"`
	PicId         *string    `gorm:"column:picId;type:varchar(60)" json:"pic_id"`
	PicName       *string    `gorm:"column:picName;type:varchar(255)" json:"pic_name"`
	CreatedByName *string    `gorm:"column:createdByName;type:varchar(255)" json:"created_by_name"`
	CreatedBy     string     `gorm:"column:createdBy;type:varchar(60);not null" json:"created_by"`
	Version       int        `gorm:"column:version;type:int;not null;default:1" json:"version"`
	IsDeleted     bool       `gorm:"column:isDeleted;type:bool;not null;default:false" json:"is_deleted"`
	CreatedAt     time.Time  `gorm:"column:createdAt;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updatedAt;autoUpdateTime" json:"updated_at"`
}

func (e *SalesDeal) BeforeCreate(tx *gorm.DB) error {
	generateUUID(&e.ID)
	return nil
}

func (SalesDeal) TableName() string {
	return "SalesDeal"
}

// ─── Partner ────────────────────────────────────────────────────

type Partner struct {
	ID            string    `gorm:"column:id;primaryKey;type:varchar(60)" json:"id"`
	Name          string    `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Type          string    `gorm:"column:type;type:varchar(20);not null;default:'buyer'" json:"type"`
	Category      *string   `gorm:"column:category;type:varchar(100)" json:"category"`
	ContactPerson *string   `gorm:"column:contactPerson;type:varchar(255)" json:"contact_person"`
	Phone         *string   `gorm:"column:phone;type:varchar(50)" json:"phone"`
	Email         *string   `gorm:"column:email;type:varchar(255)" json:"email"`
	Address       *string   `gorm:"column:address;type:text" json:"address"`
	City          *string   `gorm:"column:city;type:varchar(100)" json:"city"`
	Country       *string   `gorm:"column:country;type:varchar(100)" json:"country"`
	TaxId         *string   `gorm:"column:taxId;type:varchar(100)" json:"tax_id"`
	Status        string    `gorm:"column:status;type:varchar(20);not null;default:'active'" json:"status"`
	Notes         *string   `gorm:"column:notes;type:text" json:"notes"`
	Version       int       `gorm:"column:version;type:int;not null;default:1" json:"version"`
	IsDeleted     bool      `gorm:"column:isDeleted;type:bool;not null;default:false" json:"is_deleted"`
	CreatedAt     time.Time `gorm:"column:createdAt;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updated_at"`
}

func (e *Partner) BeforeCreate(tx *gorm.DB) error {
	generateUUID(&e.ID)
	return nil
}

func (Partner) TableName() string {
	return "Partner"
}

// ─── BlendingSimulation ─────────────────────────────────────────

type BlendingSimulation struct {
	ID            string    `gorm:"column:id;primaryKey;type:varchar(60)" json:"id"`
	Inputs        string    `gorm:"column:inputs;type:text;not null;default:'[]'" json:"inputs"`
	TotalQuantity float64   `gorm:"column:totalQuantity;type:decimal(15,2);not null;default:0" json:"total_quantity"`
	ResultGar     float64   `gorm:"column:resultGar;type:decimal(10,2);not null;default:0" json:"result_gar"`
	ResultTs      float64   `gorm:"column:resultTs;type:decimal(10,2);not null;default:0" json:"result_ts"`
	ResultAsh     float64   `gorm:"column:resultAsh;type:decimal(10,2);not null;default:0" json:"result_ash"`
	ResultTm      float64   `gorm:"column:resultTm;type:decimal(10,2);not null;default:0" json:"result_tm"`
	CreatedBy     string    `gorm:"column:createdBy;type:varchar(60);not null" json:"created_by"`
	Version       int       `gorm:"column:version;type:int;not null;default:1" json:"version"`
	IsDeleted     bool      `gorm:"column:isDeleted;type:bool;not null;default:false" json:"is_deleted"`
	CreatedAt     time.Time `gorm:"column:createdAt;autoCreateTime" json:"created_at"`
}

func (e *BlendingSimulation) BeforeCreate(tx *gorm.DB) error {
	generateUUID(&e.ID)
	return nil
}

func (BlendingSimulation) TableName() string {
	return "BlendingSimulation"
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
