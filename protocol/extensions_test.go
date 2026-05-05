package protocol

import (
	"encoding/binary"
	"testing"
)

func TestExtensions_ReplyTo_Roundtrip(t *testing.T) {
	var ext Extensions
	ext.SetReplyTo(42)

	data := ext.Marshal()
	if data == nil {
		t.Fatal("Marshal returned nil for non-empty extensions")
	}

	parsed, consumed, err := ParseExtensions(data)
	if err != nil {
		t.Fatalf("ParseExtensions failed: %v", err)
	}
	if consumed != len(data) {
		t.Fatalf("consumed %d bytes, want %d", consumed, len(data))
	}

	seq, ok := parsed.ReplyTo()
	if !ok {
		t.Fatal("ReplyTo not present after roundtrip")
	}
	if seq != 42 {
		t.Fatalf("ReplyTo = %d, want 42", seq)
	}
}

func TestExtensions_TargetService_Roundtrip(t *testing.T) {
	var ext Extensions
	ext.SetTargetService("chat.v1")

	data := ext.Marshal()
	if data == nil {
		t.Fatal("Marshal returned nil for non-empty extensions")
	}

	parsed, consumed, err := ParseExtensions(data)
	if err != nil {
		t.Fatalf("ParseExtensions failed: %v", err)
	}
	if consumed != len(data) {
		t.Fatalf("consumed %d bytes, want %d", consumed, len(data))
	}

	svc, ok := parsed.TargetService()
	if !ok {
		t.Fatal("TargetService not present after roundtrip")
	}
	if svc != "chat.v1" {
		t.Fatalf("TargetService = %q, want %q", svc, "chat.v1")
	}
}

func TestExtensions_Multiple_Combined(t *testing.T) {
	var ext Extensions
	ext.SetReplyTo(100)
	ext.SetTargetService("auth.v2")

	data := ext.Marshal()
	if data == nil {
		t.Fatal("Marshal returned nil")
	}

	// Size should be: ReplyTo(2+2+4) + TargetService(2+2+7) = 17
	expectedSize := 4 + 4 + 4 + len("auth.v2")
	if ext.Size() != expectedSize {
		t.Fatalf("Size() = %d, want %d", ext.Size(), expectedSize)
	}

	parsed, consumed, err := ParseExtensions(data)
	if err != nil {
		t.Fatalf("ParseExtensions failed: %v", err)
	}
	if consumed != len(data) {
		t.Fatalf("consumed %d, want %d", consumed, len(data))
	}

	seq, ok := parsed.ReplyTo()
	if !ok || seq != 100 {
		t.Fatalf("ReplyTo = (%d, %v), want (100, true)", seq, ok)
	}

	svc, ok := parsed.TargetService()
	if !ok || svc != "auth.v2" {
		t.Fatalf("TargetService = (%q, %v), want (\"auth.v2\", true)", svc, ok)
	}
}

func TestExtensions_Empty(t *testing.T) {
	var ext Extensions

	data := ext.Marshal()
	if data != nil {
		t.Fatalf("Marshal of empty extensions should return nil, got %v", data)
	}

	if ext.Size() != 0 {
		t.Fatalf("Size of empty extensions should be 0, got %d", ext.Size())
	}

	// Parsing empty data should succeed with zero consumed.
	parsed, consumed, err := ParseExtensions(nil)
	if err != nil {
		t.Fatalf("ParseExtensions(nil) failed: %v", err)
	}
	if consumed != 0 {
		t.Fatalf("consumed = %d, want 0", consumed)
	}
	_, ok := parsed.ReplyTo()
	if ok {
		t.Fatal("empty extensions should not have ReplyTo")
	}
}

func TestExtensions_UnknownType_Preserved(t *testing.T) {
	// Build wire data with an unknown extension type 0x00FF.
	unknownType := ExtensionType(0x00FF)
	unknownVal := []byte{0xDE, 0xAD, 0xBE, 0xEF}

	// Manually construct TLV: type(2) + length(2) + value(4) = 8 bytes
	data := make([]byte, 8)
	binary.BigEndian.PutUint16(data[0:], uint16(unknownType))
	binary.BigEndian.PutUint16(data[2:], uint16(len(unknownVal))) //nolint:gosec // test value is 4 bytes
	copy(data[4:], unknownVal)

	parsed, consumed, err := ParseExtensions(data)
	if err != nil {
		t.Fatalf("ParseExtensions failed: %v", err)
	}
	if consumed != 8 {
		t.Fatalf("consumed = %d, want 8", consumed)
	}

	// The unknown type should be preserved in the map.
	val, ok := parsed.entries[unknownType]
	if !ok {
		t.Fatal("unknown extension type not preserved")
	}
	if len(val) != 4 || val[0] != 0xDE || val[1] != 0xAD || val[2] != 0xBE || val[3] != 0xEF {
		t.Fatalf("unknown extension value = %x, want DEADBEEF", val)
	}

	// Re-marshal and verify roundtrip.
	remarshaled := parsed.Marshal()
	reparsed, _, err := ParseExtensions(remarshaled)
	if err != nil {
		t.Fatalf("re-parse failed: %v", err)
	}
	val2, ok := reparsed.entries[unknownType]
	if !ok {
		t.Fatal("unknown extension lost after re-marshal")
	}
	if len(val2) != 4 || val2[0] != 0xDE {
		t.Fatalf("value corrupted after roundtrip: %x", val2)
	}
}

func TestExtensions_TerminatorStopsParsing(t *testing.T) {
	// Build: ReplyTo extension + terminator + trailing garbage
	var ext Extensions
	ext.SetReplyTo(7)
	tlv := ext.Marshal()

	// Append terminator: type=0x0000, length=0x0000
	terminator := make([]byte, 4)
	// trailing garbage after terminator
	garbage := []byte{0xFF, 0xFF, 0xFF, 0xFF}

	data := make([]byte, 0, len(tlv)+len(terminator)+len(garbage))
	data = append(data, tlv...)
	data = append(data, terminator...)
	data = append(data, garbage...)

	parsed, consumed, err := ParseExtensions(data)
	if err != nil {
		t.Fatalf("ParseExtensions failed: %v", err)
	}

	// Should consume TLV (8 bytes) + terminator (4 bytes) = 12, NOT the garbage.
	expectedConsumed := len(tlv) + 4
	if consumed != expectedConsumed {
		t.Fatalf("consumed = %d, want %d (should stop at terminator)", consumed, expectedConsumed)
	}

	seq, ok := parsed.ReplyTo()
	if !ok || seq != 7 {
		t.Fatalf("ReplyTo = (%d, %v), want (7, true)", seq, ok)
	}
}

func TestExtensions_TruncatedData(t *testing.T) {
	// Build valid extension then truncate it.
	data := make([]byte, 6) // type(2) + length(2) says 10 bytes, but only 2 available
	binary.BigEndian.PutUint16(data[0:], uint16(ExtReplyTo))
	binary.BigEndian.PutUint16(data[2:], 10) // claims 10 bytes of value

	_, _, err := ParseExtensions(data)
	if err != ErrMessageTooShort {
		t.Fatalf("expected ErrMessageTooShort, got %v", err)
	}
}
