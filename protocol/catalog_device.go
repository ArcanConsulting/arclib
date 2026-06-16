// Code generated from internal/protocol/opcode.go; DO NOT EDIT.
// Migrated MyClerk device operation codes (category 0x04).

package protocol

// Device operations (category 0x04).
const (
	OpDeviceCreate     OpCode = 0x0400 // Register device
	OpDeviceGet        OpCode = 0x0401 // Get device info
	OpDeviceUpdate     OpCode = 0x0402 // Update device
	OpDeviceRemove     OpCode = 0x0403 // Remove device
	OpDeviceListFamily OpCode = 0x0404 // List family devices
	OpDeviceListUser   OpCode = 0x0405 // List user devices
	OpDevicePair       OpCode = 0x0410 // Pair device
	OpDeviceUnpair     OpCode = 0x0411 // Unpair device
	OpDeviceClaim      OpCode = 0x0412 // Claim device
	OpDeviceRelease    OpCode = 0x0413 // Release device
	OpDeviceCommand    OpCode = 0x0420 // Send command
	OpDeviceStatus     OpCode = 0x0421 // Get status
	OpDeviceConfig     OpCode = 0x0422 // Configure device
	OpDeviceOTA        OpCode = 0x0430 // OTA update
	OpDeviceOTAStatus  OpCode = 0x0431 // OTA status
	OpDeviceReboot     OpCode = 0x0432 // Reboot device
	OpDeviceReset      OpCode = 0x0433 // Factory reset
)
