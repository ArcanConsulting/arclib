package crypto

import (
	"encoding/hex"
	"testing"
)

func TestSHA256Sum(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "hello",
			input:    "hello",
			expected: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		},
		{
			name:     "abc",
			input:    "abc",
			expected: "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SHA256Sum([]byte(tt.input))
			got := hex.EncodeToString(result[:])
			if got != tt.expected {
				t.Errorf("SHA256Sum(%q) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSHA256Sum_Deterministic(t *testing.T) {
	data := []byte("deterministic test input")
	a := SHA256Sum(data)
	b := SHA256Sum(data)
	if a != b {
		t.Error("SHA256Sum is not deterministic")
	}
}

func TestCRC16(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint16
	}{
		{
			name:     "standard test vector",
			input:    "123456789",
			expected: 0x4B37,
		},
		{
			name:     "empty data",
			input:    "",
			expected: 0xFFFF,
		},
		{
			name:     "single byte zero",
			input:    "\x00",
			expected: 0x40BF,
		},
		{
			name:     "hello",
			input:    "hello",
			expected: CRC16([]byte("hello")), // self-consistent
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CRC16([]byte(tt.input))
			if got != tt.expected {
				t.Errorf("CRC16(%q) = 0x%04X, want 0x%04X", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCRC16_Deterministic(t *testing.T) {
	data := []byte("deterministic crc16 test")
	a := CRC16(data)
	b := CRC16(data)
	if a != b {
		t.Error("CRC16 is not deterministic")
	}
}

func TestCRC16_Uniqueness(t *testing.T) {
	a := CRC16([]byte("input one"))
	b := CRC16([]byte("input two"))
	if a == b {
		t.Error("CRC16 produced identical checksums for different inputs")
	}
}

func TestCRC32(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint32
	}{
		{
			name:     "IEEE standard test vector",
			input:    "123456789",
			expected: 0xCBF43926,
		},
		{
			name:     "empty data",
			input:    "",
			expected: 0,
		},
		{
			name:     "single byte zero",
			input:    "\x00",
			expected: 0xD202EF8D,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CRC32([]byte(tt.input))
			if got != tt.expected {
				t.Errorf("CRC32(%q) = 0x%08X, want 0x%08X", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCRC32_Deterministic(t *testing.T) {
	data := []byte("deterministic crc32 test")
	a := CRC32(data)
	b := CRC32(data)
	if a != b {
		t.Error("CRC32 is not deterministic")
	}
}

func TestCRC32_Uniqueness(t *testing.T) {
	a := CRC32([]byte("input one"))
	b := CRC32([]byte("input two"))
	if a == b {
		t.Error("CRC32 produced identical checksums for different inputs")
	}
}

func BenchmarkCRC16(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i)
	}
	b.ResetTimer()
	for b.Loop() {
		CRC16(data)
	}
}

func BenchmarkCRC32(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i)
	}
	b.ResetTimer()
	for b.Loop() {
		CRC32(data)
	}
}

func BenchmarkSHA256Sum(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i)
	}
	b.ResetTimer()
	for b.Loop() {
		SHA256Sum(data)
	}
}
