// Code generated from internal/protocol/opcode.go; DO NOT EDIT.
// Migrated MyClerk desktop operation codes (category 0x1D).

package protocol

// Desktop operations (category 0x1D).
const (
	OpDesktopStatus       OpCode = 0x1D00 // Get desktop client status
	OpDesktopConnect      OpCode = 0x1D01 // Connect to server
	OpDesktopDisconnect   OpCode = 0x1D02 // Disconnect from server
	OpDesktopTrayStatus   OpCode = 0x1D03 // Get/set tray status
	OpDesktopNotify       OpCode = 0x1D04 // Send desktop notification
	OpDesktopUpdateCheck  OpCode = 0x1D10 // Check for updates
	OpDesktopUpdateStatus OpCode = 0x1D11 // Get update status
	OpDesktopConfigGet    OpCode = 0x1D20 // Get desktop config
	OpDesktopConfigSet    OpCode = 0x1D21 // Set desktop config
	OpDesktopHotkeyList   OpCode = 0x1D22 // List registered hotkeys
)
