// Code generated from internal/protocol/opcode.go; DO NOT EDIT.
// Migrated MyClerk analytics operation codes (category 0x1F).

package protocol

// Analytics operations (category 0x1F).
const (
	OpAnalyticsDashboardCreate  OpCode = 0x1F00 // Create dashboard
	OpAnalyticsDashboardGet     OpCode = 0x1F01 // Get dashboard
	OpAnalyticsDashboardUpdate  OpCode = 0x1F02 // Update dashboard
	OpAnalyticsDashboardDelete  OpCode = 0x1F03 // Delete dashboard
	OpAnalyticsDashboardList    OpCode = 0x1F04 // List dashboards
	OpAnalyticsDashboardDefault OpCode = 0x1F05 // Get or create default dashboard
	OpAnalyticsWidgetCreate     OpCode = 0x1F10 // Create widget
	OpAnalyticsWidgetGet        OpCode = 0x1F11 // Get widget
	OpAnalyticsWidgetUpdate     OpCode = 0x1F12 // Update widget
	OpAnalyticsWidgetDelete     OpCode = 0x1F13 // Delete widget
	OpAnalyticsWidgetList       OpCode = 0x1F14 // List widgets for dashboard
	OpAnalyticsUsageQuery       OpCode = 0x1F20 // Usage time-series and summary
	OpAnalyticsCostQuery        OpCode = 0x1F21 // Cost time-series and breakdown
	OpAnalyticsAPIQuery         OpCode = 0x1F22 // API analytics
	OpAnalyticsTrendQuery       OpCode = 0x1F23 // Trend analysis
	OpAnalyticsSnapshotQuery    OpCode = 0x1F24 // Latest usage snapshot
	OpAnalyticsReportCreate     OpCode = 0x1F30 // Generate report
	OpAnalyticsReportStatus     OpCode = 0x1F31 // Check report status
	OpAnalyticsReportDownload   OpCode = 0x1F32 // Download report data
	OpAnalyticsReportList       OpCode = 0x1F33 // List reports
)
