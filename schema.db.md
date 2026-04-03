
model Account {
  id                String  @id @default(cuid())
  userId            String
  type              String
  provider          String
  providerAccountId String
  refresh_token     String?
  access_token      String?
  expires_at        Int?
  token_type        String?
  scope             String?
  id_token          String?
  session_state     String?

  user User @relation(fields: [userId], references: [id], onDelete: Cascade)

  @@unique([provider, providerAccountId])
}

model Session {
  id           String   @id @default(cuid())
  sessionToken String   @unique
  userId       String
  expires      DateTime
  user         User     @relation(fields: [userId], references: [id], onDelete: Cascade)
}

model User {
  id            String    @id @default(cuid())
  name          String?
  email         String?   @unique
  emailVerified DateTime?
  image         String?
  password      String?
  role          String    @default("staff") // Roles: ceo, director, manager, staff

  accounts      Account[]
  sessions      Session[]
  chatHistories ChatHistory[]
  meetingMedia  MeetingMedia[]
  auditLogs     AuditLog[]
}

model VerificationToken {
  identifier String
  token      String   @unique
  expires    DateTime

  @@unique([identifier, token])
}

model ChatHistory {
  id        String   @id @default(cuid())
  userId    String
  role      String // "user" or "assistant"
  content   String
  createdAt DateTime @default(now())

  user User @relation(fields: [userId], references: [id], onDelete: Cascade)
}

model MeetingMedia {
  id              String   @id @default(cuid())
  meetingId       String // ID from Google Sheets
  uploaderId      String
  fileName        String
  fileUrl         String
  transcribedText String?
  generatedMom    String?
  extractedTasks  String?
  createdAt       DateTime @default(now())

  uploader User @relation(fields: [uploaderId], references: [id], onDelete: Cascade)
}

model TimelineMilestone {
  id          String   @id @default(cuid())
  shipmentId  String
  title       String
  date        DateTime
  description String?
  createdAt   DateTime @default(now())
  updatedAt   DateTime @updatedAt
}

// --- Memory B Schema Additions ---

model SyncState {
  id           String   @id @default("singleton")
  lastSyncTime DateTime @default(now())
  updatedAt    DateTime @updatedAt
}

model AuditLog {
  id        String   @id @default(cuid())
  userId    String
  userName  String
  action    String // CREATE, UPDATE, DELETE
  entity    String // Task, Sales, Shipment, etc
  entityId  String
  details   String? // JSON stringified changes
  createdAt DateTime @default(now())

  user User @relation(fields: [userId], references: [id], onDelete: Cascade)
}

model TaskItem {
  id           String    @id @default(cuid())
  title        String
  description  String?
  status       String    @default("todo") // todo, in_progress, review, done
  priority     String    @default("medium") // low, medium, high, urgent
  assigneeId   String?
  assigneeName String?
  dueDate      DateTime?
  createdBy    String
  createdAt    DateTime  @default(now())
  updatedAt    DateTime  @updatedAt
  isDeleted    Boolean   @default(false)
}

model SalesOrder {
  id            String   @id @default(cuid())
  orderNumber   String   @unique
  client        String
  description   String?
  amount        Float    @default(0)
  priority      String   @default("medium")
  status        String   @default("pending") // draft, pending, approved, rejected
  imageUrl      String?
  createdByName String?
  createdBy     String
  approvedBy    String?
  notes         String?
  createdAt     DateTime @default(now())
  updatedAt     DateTime @updatedAt
  isDeleted     Boolean  @default(false)
}

model PurchaseRequest {
  id            String   @id @default(cuid())
  requestNumber String   @unique
  category      String
  supplier      String?
  description   String?
  amount        Float    @default(0)
  priority      String   @default("medium")
  status        String   @default("pending")
  imageUrl      String?
  createdByName String?
  createdBy     String
  approvedBy    String?
  notes         String?
  isAnomaly     Boolean  @default(false)
  anomalyReason String?
  ocrData       String? // JSON
  createdAt     DateTime @default(now())
  updatedAt     DateTime @updatedAt
  isDeleted     Boolean  @default(false)
}

model ShipmentDetail {
  id             String    @id @default(cuid())
  shipmentNumber String    @unique
  dealId         String?
  status         String    @default("draft")
  buyer          String
  supplier       String?
  isBlending     Boolean   @default(false)
  iupOp          String?
  vesselName     String?
  bargeName      String?
  loadingPort    String?
  dischargePort  String?
  quantityLoaded Float?
  blDate         DateTime?
  eta            DateTime?
  salesPrice     Float?
  marginMt       Float?
  picName        String?
  type           String    @default("export")
  milestones     String? // JSON
  createdAt      DateTime  @default(now())
  updatedAt      DateTime  @updatedAt
  isDeleted      Boolean   @default(false)
}

model SourceSupplier {
  id               String   @id @default(cuid())
  name             String
  region           String
  calorieRange     String?
  gar              Float?
  ts               Float?
  ash              Float?
  tm               Float?
  im               Float?
  fc               Float?
  nar              Float?
  adb              Float?
  jettyPort        String?
  anchorage        String?
  stockAvailable   Float    @default(0)
  minStockAlert    Float?
  kycStatus        String   @default("not_started")
  psiStatus        String   @default("not_started")
  fobBargeOnly     Boolean  @default(false)
  priceLinkedIndex String?
  fobBargePriceUsd Float?
  contractType     String?
  picName          String?
  iupNumber        String?
  createdAt        DateTime @default(now())
  updatedAt        DateTime @updatedAt
  isDeleted        Boolean  @default(false)
}

model QualityResult {
  id           String    @id @default(cuid())
  cargoId      String
  cargoName    String
  surveyor     String?
  samplingDate DateTime?
  gar          Float?
  ts           Float?
  ash          Float?
  tm           Float?
  status       String    @default("pending")
  createdAt    DateTime  @default(now())
  updatedAt    DateTime  @updatedAt
  isDeleted    Boolean   @default(false)
}

model MarketPrice {
  id        String   @id @default(cuid())
  date      DateTime
  ici1      Float?
  ici2      Float?
  ici3      Float?
  ici4      Float?
  ici5      Float?
  newcastle Float?
  hba       Float?
  source    String?
  createdAt DateTime @default(now())
  updatedAt DateTime @updatedAt
  isDeleted Boolean  @default(false)
}

model MeetingItem {
  id            String    @id @default(cuid())
  title         String
  date          DateTime?
  time          String?
  location      String?
  status        String    @default("scheduled")
  attendees     String? // JSON array
  momContent    String? // Plain text MOM transcript
  voiceNoteUrl  String? // URL to uploaded audio/video
  aiSummary     String? // AI-generated summary (markdown)
  createdByName String?
  createdBy     String
  createdAt     DateTime  @default(now())
  updatedAt     DateTime  @updatedAt
  isDeleted     Boolean   @default(false)
}

model PLForecast {
  id               String   @id @default(cuid())
  dealId           String?
  dealNumber       String?
  projectName      String?
  buyer            String?
  type             String   @default("export")
  status           String   @default("forecast")
  quantity         Float    @default(0)
  sellingPrice     Float    @default(0)
  buyingPrice      Float    @default(0)
  freightCost      Float    @default(0)
  otherCost        Float    @default(0)
  grossProfitMt    Float    @default(0)
  totalGrossProfit Float    @default(0)
  createdBy        String?
  createdAt        DateTime @default(now())
  updatedAt        DateTime @updatedAt
  isDeleted        Boolean  @default(false)
}

model SalesDeal {
  id            String    @id @default(cuid())
  dealNumber    String    @unique
  status        String    @default("pre_sale")
  buyer         String
  buyerCountry  String?
  type          String    @default("export")
  shippingTerms String    @default("FOB")
  quantity      Float     @default(0)
  pricePerMt    Float?
  totalValue    Float?
  laycanStart   DateTime?
  laycanEnd     DateTime?
  vesselName    String?
  gar           Float?
  ts            Float?
  ash           Float?
  tm            Float?
  projectId     String?
  picId         String?
  picName       String?
  createdByName String?
  createdBy     String
  createdAt     DateTime  @default(now())
  updatedAt     DateTime  @updatedAt
  isDeleted     Boolean   @default(false)
}

model Partner {
  id            String   @id @default(cuid())
  name          String
  type          String   @default("buyer") // buyer, vendor, fleet
  category      String?
  contactPerson String?
  phone         String?
  email         String?
  address       String?
  city          String?
  country       String?
  taxId         String?
  status        String   @default("active") // active, under_review, inactive
  notes         String?
  createdAt     DateTime @default(now())
  updatedAt     DateTime @updatedAt
  isDeleted     Boolean  @default(false)
}

model BlendingSimulation {
  id            String   @id @default(cuid())
  inputs        String   @default("[]") // JSON array of { name, quantity, gar, ts, ash, tm }
  totalQuantity Float    @default(0)
  resultGar     Float    @default(0)
  resultTs      Float    @default(0)
  resultAsh     Float    @default(0)
  resultTm      Float    @default(0)
  createdBy     String
  createdAt     DateTime @default(now())
  isDeleted     Boolean  @default(false)
}
