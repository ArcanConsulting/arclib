package protocol

// Category represents an operation category (the high byte of an OpCode).
//
// The category vocabulary is the authoritative MyClerk catalog: every consuming
// application (MyClerk core, mayaservices, ArcShell, ArcMail, ArcHub, the
// SilverWidget terminal) classifies opcodes by these categories, so they live in
// arclib — the single source of truth — rather than being mirrored per project.
type Category uint8

// Protocol Infrastructure Categories (0x00-0x07).
const (
	CategorySystem          Category = 0x00 // Core: System, Discovery, Key Management
	CategoryAuth            Category = 0x01 // Standard: Authentication, Identity, Messaging
	CategoryResourceSharing Category = 0x02 // Resource Sharing (Streams, GPU, Routing)
	CategoryFamily          Category = 0x03 // Federation: Family, Roles, Permissions
	CategoryDevice          Category = 0x04 // Billing: Device Management, DevLink
	CategoryVFS             Category = 0x05 // VFS: Virtual File System (draft-myclerk-vfs)
	CategoryHardware        Category = 0x06 // Hardware Passthrough: USB, Serial, GPIO, I2C, SPI, CAN, 1-Wire
	CategoryKnowledge       Category = 0x07 // Knowledge & Memory Operations
)

// MyClerk Application Categories (0x08-0x31) - Per draft-myclerk-protocol Section 6, Table 2.
const (
	CategorySettings     Category = 0x08 // Settings & Configuration
	CategoryRoom         Category = 0x09 // Room/Channel Management
	CategoryMedia        Category = 0x0A // Media Handling
	CategoryCall         Category = 0x0B // Voice/Video Calls
	CategoryCalendar     Category = 0x0C // Calendar & Events
	CategoryTask         Category = 0x0D // Tasks, Lists, Shopping
	CategorySmartHome    Category = 0x0E // Smart Home Devices
	CategoryAutomation   Category = 0x0F // Automation Rules
	CategoryMaya         Category = 0x10 // Maya AI Assistant
	CategorySync         Category = 0x11 // Synchronization, VFS Federation
	CategoryOrch         Category = 0x12 // Orchestration, Notifications, Backup
	CategoryPresence     Category = 0x13 // Presence, Status, Gamification
	CategoryHealth       Category = 0x14 // Health, Social, Community
	CategoryLocation     Category = 0x15 // Location Services, Mobility
	CategoryDebug        Category = 0x16 // Debug/Testing
	CategoryMDM          Category = 0x17 // Mobile Device Management
	CategoryLegal        Category = 0x18 // Legal Documents & Care Directives
	CategoryMobility     Category = 0x19 // Mobility & Vehicle Management
	CategoryTravel       Category = 0x1A // Travel & Vacation
	CategoryAudit        Category = 0x1B // Audit & Compliance
	CategoryAudio        Category = 0x1C // Audio Pipeline
	CategoryDesktop      Category = 0x1D // Desktop Client
	CategoryTenant       Category = 0x1E // Tenant Management
	CategoryAnalytics    Category = 0x1F // Business Analytics
	CategoryCommunityFed Category = 0x20 // Community Federation (Emergency & Registry)
	CategoryMarketData   Category = 0x30 // Market Data (SilverWidget Financial Services)
	CategoryArcMail      Category = 0x31 // ArcMail Mail Service (draft-myclerk-arcmail)
	CategoryReserved     Category = 0xFF // Reserved (IETF)
)

// categoryNames maps categories to human-readable names.
var categoryNames = map[Category]string{
	CategorySystem:          "system",
	CategoryAuth:            "auth",
	CategoryResourceSharing: "resource_sharing",
	CategoryFamily:          "family",
	CategoryDevice:          "device",
	CategoryVFS:             "vfs",
	CategoryHardware:        "hardware",
	CategoryKnowledge:       "knowledge",
	CategorySettings:        "settings",
	CategoryRoom:            "room",
	CategoryMedia:           "media",
	CategoryCall:            "call",
	CategoryCalendar:        "calendar",
	CategoryTask:            "task",
	CategorySmartHome:       "smarthome",
	CategoryAutomation:      "automation",
	CategoryMaya:            "maya",
	CategorySync:            "sync",
	CategoryOrch:            "orchestration",
	CategoryPresence:        "presence",
	CategoryHealth:          "health",
	CategoryLocation:        "location",
	CategoryDebug:           "debug",
	CategoryMDM:             "mdm",
	CategoryLegal:           "legal",
	CategoryMobility:        "mobility",
	CategoryTravel:          "travel",
	CategoryAudit:           "audit",
	CategoryAudio:           "audio",
	CategoryDesktop:         "desktop",
	CategoryTenant:          "tenant",
	CategoryAnalytics:       "analytics",
	CategoryCommunityFed:    "community_fed",
	CategoryMarketData:      "market_data",
	CategoryArcMail:         "arcmail",
	CategoryReserved:        "reserved",
}

// String returns the category name, or "unknown" for an unregistered category.
func (c Category) String() string {
	if name, ok := categoryNames[c]; ok {
		return name
	}
	return "unknown"
}
