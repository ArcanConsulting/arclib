package protocol

import "testing"

func TestOpCodeString(t *testing.T) {
	tests := []struct {
		op       OpCode
		expected string
	}{
		{OpNop, "NOP"},
		{OpKeepalive, "KEEPALIVE"},
		{OpKeepaliveAck, "KEEPALIVE_ACK"},
		{OpSessionInit, "SESSION_INIT"},
		{OpSessionAck, "SESSION_ACK"},
		{OpSessionClose, "SESSION_CLOSE"},
		{OpSessionCloseAck, "SESSION_CLOSE_ACK"},
		{OpSessionResume, "SESSION_RESUME"},
		{OpSessionResumed, "SESSION_RESUMED"},
		{OpKeyExchangeInit, "KEY_EXCHANGE_INIT"},
		{OpKeyExchangeResponse, "KEY_EXCHANGE_RESPONSE"},
		{OpKeyExchangeComplete, "KEY_EXCHANGE_COMPLETE"},
		{OpSessionRotate, "SESSION_ROTATE"},
		{OpSessionRevoke, "SESSION_REVOKE"},
		{OpError, "ERROR"},
	}

	for _, tt := range tests {
		if got := tt.op.String(); got != tt.expected {
			t.Errorf("OpCode(0x%04X).String() = %q, want %q", uint16(tt.op), got, tt.expected)
		}
	}
}

func TestOpCodeStringUnknown(t *testing.T) {
	op := OpCode(0xBEEF)
	got := op.String()
	expected := "0xBEEF"
	if got != expected {
		t.Errorf("OpCode(0xBEEF).String() = %q, want %q", got, expected)
	}
}

func TestOpCodeCategory(t *testing.T) {
	tests := []struct {
		op       OpCode
		category uint8
	}{
		{OpNop, 0x00},
		{OpError, 0x00},
		{OpCode(0x0100), 0x01},
		{OpCode(0xFF00), 0xFF},
	}

	for _, tt := range tests {
		if got := tt.op.Category(); got != tt.category {
			t.Errorf("OpCode(0x%04X).Category() = 0x%02X, want 0x%02X",
				uint16(tt.op), got, tt.category)
		}
	}
}

func TestOpCodeOperation(t *testing.T) {
	tests := []struct {
		op        OpCode
		operation uint8
	}{
		{OpNop, 0x00},
		{OpKeepalive, 0x01},
		{OpError, 0xFF},
		{OpKeyExchangeInit, 0x10},
	}

	for _, tt := range tests {
		if got := tt.op.Operation(); got != tt.operation {
			t.Errorf("OpCode(0x%04X).Operation() = 0x%02X, want 0x%02X",
				uint16(tt.op), got, tt.operation)
		}
	}
}

func TestNewOpCode(t *testing.T) {
	op := NewOpCode(0x01, 0x50)
	if op != OpCode(0x0150) {
		t.Errorf("NewOpCode(0x01, 0x50) = 0x%04X, want 0x0150", uint16(op))
	}
	if op.Category() != 0x01 {
		t.Errorf("Category = 0x%02X, want 0x01", op.Category())
	}
	if op.Operation() != 0x50 {
		t.Errorf("Operation = 0x%02X, want 0x50", op.Operation())
	}
}

func TestRegisterOpCodes(t *testing.T) {
	custom := map[OpCode]string{
		OpCode(0xF001): "CUSTOM_OP_1",
		OpCode(0xF002): "CUSTOM_OP_2",
	}
	RegisterOpCodes("test", custom)

	if got := OpCode(0xF001).String(); got != "CUSTOM_OP_1" {
		t.Errorf("registered OpCode(0xF001).String() = %q, want %q", got, "CUSTOM_OP_1")
	}
	if got := OpCode(0xF002).String(); got != "CUSTOM_OP_2" {
		t.Errorf("registered OpCode(0xF002).String() = %q, want %q", got, "CUSTOM_OP_2")
	}
}

func TestLookupOpCode(t *testing.T) {
	// Core opcode
	if got := LookupOpCode(OpNop); got != "NOP" {
		t.Errorf("LookupOpCode(OpNop) = %q, want %q", got, "NOP")
	}

	// Unknown opcode
	if got := LookupOpCode(OpCode(0x9999)); got != "" {
		t.Errorf("LookupOpCode(0x9999) = %q, want empty", got)
	}

	// Register and lookup
	RegisterOpCodes("lookup_test", map[OpCode]string{
		OpCode(0xF010): "LOOKUP_TEST",
	})
	if got := LookupOpCode(OpCode(0xF010)); got != "LOOKUP_TEST" {
		t.Errorf("LookupOpCode(0xF010) = %q, want %q", got, "LOOKUP_TEST")
	}
}

func TestOpCodeValues(t *testing.T) {
	// Verify exact opcode values match the wire protocol
	if OpNop != 0x0000 {
		t.Errorf("OpNop = 0x%04X, want 0x0000", uint16(OpNop))
	}
	if OpKeepalive != 0x0001 {
		t.Errorf("OpKeepalive = 0x%04X, want 0x0001", uint16(OpKeepalive))
	}
	if OpSessionInit != 0x0003 {
		t.Errorf("OpSessionInit = 0x%04X, want 0x0003", uint16(OpSessionInit))
	}
	if OpKeyExchangeInit != 0x0010 {
		t.Errorf("OpKeyExchangeInit = 0x%04X, want 0x0010", uint16(OpKeyExchangeInit))
	}
	if OpError != 0x00FF {
		t.Errorf("OpError = 0x%04X, want 0x00FF", uint16(OpError))
	}
}
