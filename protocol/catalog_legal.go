// MyClerk opcode catalog (canonical reference in arclib; apps copy locally, see cmd/opcode-verify).
// Migrated MyClerk legal operation codes (category 0x18).

package protocol

// Legal operations (category 0x18).
const (
	OpLegalDocCreate        OpCode = 0x1800 // Create legal document
	OpLegalDocGet           OpCode = 0x1801 // Get legal document
	OpLegalDocUpdate        OpCode = 0x1802 // Update legal document
	OpLegalDocDelete        OpCode = 0x1803 // Delete legal document
	OpLegalDocList          OpCode = 0x1804 // List legal documents
	OpLegalVersionList      OpCode = 0x1808 // List document versions
	OpLegalVersionGet       OpCode = 0x1809 // Get specific version
	OpLegalVersionRestore   OpCode = 0x180A // Restore document to version
	OpLegalAccessGrant      OpCode = 0x1810 // Grant document access
	OpLegalAccessRevoke     OpCode = 0x1811 // Revoke document access
	OpLegalAccessList       OpCode = 0x1812 // List document access grants
	OpLegalAccessCheck      OpCode = 0x1813 // Check access for user
	OpLegalShareCreate      OpCode = 0x1818 // Create Shamir shares
	OpLegalShareList        OpCode = 0x1819 // List shares for document
	OpLegalShareVerify      OpCode = 0x181A // Verify share holder access
	OpLegalShareRecover     OpCode = 0x181B // Recover document from shares
	OpLegalExtShareCreate   OpCode = 0x1820 // Create external share link
	OpLegalExtShareRevoke   OpCode = 0x1821 // Revoke external share
	OpLegalExtShareList     OpCode = 0x1822 // List external shares
	OpLegalExtShareAccess   OpCode = 0x1823 // Access via external share
	OpLegalReminderCreate   OpCode = 0x1828 // Create reminder
	OpLegalReminderUpdate   OpCode = 0x1829 // Update reminder
	OpLegalReminderDelete   OpCode = 0x182A // Delete reminder
	OpLegalReminderList     OpCode = 0x182B // List reminders
	OpLegalReminderSnooze   OpCode = 0x182C // Snooze reminder
	OpLegalReminderComplete OpCode = 0x182D // Mark reminder complete
	OpLegalAuditList        OpCode = 0x1830 // List audit entries
	OpLegalAccountCreate    OpCode = 0x1838 // Add digital account
	OpLegalAccountUpdate    OpCode = 0x1839 // Update digital account
	OpLegalAccountDelete    OpCode = 0x183A // Delete digital account
	OpLegalAccountList      OpCode = 0x183B // List digital accounts
	OpLegalTemplateList     OpCode = 0x1840 // List document templates
	OpLegalTemplateGet      OpCode = 0x1841 // Get document template
	OpLegalEmergencyRequest OpCode = 0x1848 // Request emergency access
	OpLegalEmergencyApprove OpCode = 0x1849 // Approve emergency access
	OpLegalEmergencyDeny    OpCode = 0x184A // Deny emergency access
	OpLegalEmergencyCheck   OpCode = 0x184B // Check emergency access status
)
