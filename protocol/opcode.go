package protocol

import (
	"fmt"
	"sync"
)

// OpCode represents a 16-bit operation code.
// Format: CCCC_CCCC_OOOO_OOOO (8-bit category + 8-bit operation).
type OpCode uint16

// Core Session Operations (0x0000-0x0008).
const (
	OpNop             OpCode = 0x0000 // No operation (no response)
	OpKeepalive       OpCode = 0x0001 // Keepalive request
	OpKeepaliveAck    OpCode = 0x0002 // Keepalive acknowledgement
	OpSessionInit     OpCode = 0x0003 // Session initialization
	OpSessionAck      OpCode = 0x0004 // Session initialization acknowledgement
	OpSessionClose    OpCode = 0x0005 // Session close request
	OpSessionCloseAck OpCode = 0x0006 // Session close acknowledgement
	OpSessionResume   OpCode = 0x0007 // Session resume request
	OpSessionResumed  OpCode = 0x0008 // Session resumed acknowledgement
)

// Key Management Operations (0x0010-0x0017).
const (
	OpKeyExchangeInit     OpCode = 0x0010 // X25519 key exchange initiation
	OpKeyExchangeResponse OpCode = 0x0011 // Key exchange response
	OpKeyExchangeComplete OpCode = 0x0012 // Key exchange completion
	OpSessionRotate       OpCode = 0x0016 // Session key rotation
	OpSessionRevoke       OpCode = 0x0017 // Session key revocation
)

// Core error response.
const (
	OpError OpCode = 0x00FF // Error response
)

// coreOpCodeNames maps core opcodes to their string representation.
var coreOpCodeNames = map[OpCode]string{
	OpNop:                 "NOP",
	OpKeepalive:           "KEEPALIVE",
	OpKeepaliveAck:        "KEEPALIVE_ACK",
	OpSessionInit:         "SESSION_INIT",
	OpSessionAck:          "SESSION_ACK",
	OpSessionClose:        "SESSION_CLOSE",
	OpSessionCloseAck:     "SESSION_CLOSE_ACK",
	OpSessionResume:       "SESSION_RESUME",
	OpSessionResumed:      "SESSION_RESUMED",
	OpKeyExchangeInit:     "KEY_EXCHANGE_INIT",
	OpKeyExchangeResponse: "KEY_EXCHANGE_RESPONSE",
	OpKeyExchangeComplete: "KEY_EXCHANGE_COMPLETE",
	OpSessionRotate:       "SESSION_ROTATE",
	OpSessionRevoke:       "SESSION_REVOKE",
	OpError:               "ERROR",
}

// opCodeRegistry holds externally registered opcode names.
var (
	opCodeRegistryMu sync.RWMutex
	opCodeRegistry   = make(map[OpCode]string)
)

// String returns the opcode name or a hex representation.
func (op OpCode) String() string {
	if name, ok := coreOpCodeNames[op]; ok {
		return name
	}

	opCodeRegistryMu.RLock()
	name, ok := opCodeRegistry[op]
	opCodeRegistryMu.RUnlock()

	if ok {
		return name
	}
	return fmt.Sprintf("0x%04X", uint16(op))
}

// Category extracts the category (high byte) from the opcode.
func (op OpCode) Category() Category {
	return Category((op >> 8) & 0xFF)
}

// Operation extracts the operation byte (low byte) from the opcode.
func (op OpCode) Operation() uint8 {
	return uint8(op & 0xFF)
}

// NewOpCode creates an opcode from a category and operation number.
func NewOpCode(category Category, operation uint8) OpCode {
	return OpCode(uint16(category)<<8 | uint16(operation))
}

// RegisterOpCodes registers a set of opcodes with their names.
// This allows external packages to extend the protocol with their own operations.
// The name parameter identifies the registering module (for debugging).
func RegisterOpCodes(name string, codes map[OpCode]string) {
	_ = name // Reserved for future diagnostics
	opCodeRegistryMu.Lock()
	defer opCodeRegistryMu.Unlock()
	for op, opName := range codes {
		opCodeRegistry[op] = opName
	}
}

// LookupOpCode returns the registered name for an opcode, or empty string if unknown.
func LookupOpCode(op OpCode) string {
	if name, ok := coreOpCodeNames[op]; ok {
		return name
	}

	opCodeRegistryMu.RLock()
	name := opCodeRegistry[op]
	opCodeRegistryMu.RUnlock()

	return name
}
