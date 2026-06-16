// Code generated from internal/protocol/opcode.go; DO NOT EDIT.
// Migrated MyClerk audit operation codes (category 0x1B).

package protocol

// Audit operations (category 0x1B).
const (
	OpAuditLogQuery            OpCode = 0x1B00 // Query audit log entries
	OpAuditLogExport           OpCode = 0x1B01 // Export audit log (CSV/JSON)
	OpAuditLogStats            OpCode = 0x1B02 // Get audit statistics
	OpAuditLogPurge            OpCode = 0x1B03 // Purge old audit entries
	OpAuditLogDetail           OpCode = 0x1B04 // Get single audit entry detail
	OpRetentionPolicyCreate    OpCode = 0x1B08 // Create retention policy
	OpRetentionPolicyGet       OpCode = 0x1B09 // Get retention policy
	OpRetentionPolicyUpdate    OpCode = 0x1B0A // Update retention policy
	OpRetentionPolicyDelete    OpCode = 0x1B0B // Delete retention policy
	OpRetentionPolicyList      OpCode = 0x1B0C // List retention policies
	OpRetentionPolicyRun       OpCode = 0x1B0D // Run retention policy now
	OpGDPRExportRequest        OpCode = 0x1B10 // Request data export (Art. 15)
	OpGDPRExportStatus         OpCode = 0x1B11 // Check export status
	OpGDPRExportDownload       OpCode = 0x1B12 // Download export
	OpGDPRDeleteRequest        OpCode = 0x1B13 // Request data deletion (Art. 17)
	OpGDPRDeleteStatus         OpCode = 0x1B14 // Check deletion status
	OpGDPRConsentList          OpCode = 0x1B15 // List user consents
	OpComplianceReportGenerate OpCode = 0x1B18 // Generate compliance report
	OpComplianceReportGet      OpCode = 0x1B19 // Get report
	OpComplianceReportList     OpCode = 0x1B1A // List reports
	OpComplianceReportDelete   OpCode = 0x1B1B // Delete report
	OpAccessReviewCreate       OpCode = 0x1B20 // Create access review
	OpAccessReviewGet          OpCode = 0x1B21 // Get access review
	OpAccessReviewUpdate       OpCode = 0x1B22 // Update review (approve/deny)
	OpAccessReviewList         OpCode = 0x1B23 // List reviews
	OpAccessReviewComplete     OpCode = 0x1B24 // Complete review cycle
	OpAccessReviewSchedule     OpCode = 0x1B25 // Schedule periodic review
)
