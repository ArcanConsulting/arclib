// MyClerk opcode catalog (canonical reference in arclib; apps copy locally, see cmd/opcode-verify).
// Migrated MyClerk settings operation codes (category 0x08).

package protocol

// Settings operations (category 0x08).
const (
	OpSettingsGet      OpCode = 0x0800 // Get single setting
	OpSettingsSet      OpCode = 0x0801 // Set single setting
	OpSettingsList     OpCode = 0x0810 // List settings by namespace
	OpSettingsListAll  OpCode = 0x0811 // List all settings
	OpSettingsBatchSet OpCode = 0x0820 // Set multiple settings atomically
	OpSettingsDelete   OpCode = 0x0821 // Delete setting
	OpSettingsReset    OpCode = 0x0830 // Reset namespace to defaults
	OpSettingsResetAll OpCode = 0x0831 // Reset all to defaults
	OpSettingsExport   OpCode = 0x0840 // Export config as YAML
	OpSettingsImport   OpCode = 0x0841 // Import config from YAML
)
