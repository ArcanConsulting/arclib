package crypto

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"testing"
)

func testKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	return key
}

func TestHMACSHA256(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		data     string
		expected string
	}{
		{
			// RFC 4231 Test Case 2
			name:     "RFC 4231 TC2",
			key:      "4a656665",
			data:     "7768617420646f2079612077616e7420666f72206e6f7468696e673f",
			expected: "5bdcc146bf60754e6a042426089575c75a003f089d2739839dec58b964ec3843",
		},
		{
			// RFC 4231 Test Case 1 (key=0x0b * 20, data="Hi There")
			name:     "RFC 4231 TC1",
			key:      "0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b",
			data:     "4869205468657265",
			expected: "b0344c61d8db38535ca8afceaf0bf12b881dc200c9833da726e9376c2e32cff7",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := hex.DecodeString(tt.key)
			if err != nil {
				t.Fatalf("invalid key hex: %v", err)
			}
			data, err := hex.DecodeString(tt.data)
			if err != nil {
				t.Fatalf("invalid data hex: %v", err)
			}
			got := hex.EncodeToString(HMACSHA256(key, data))
			if got != tt.expected {
				t.Errorf("HMACSHA256() = %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestHMACSHA256_Size(t *testing.T) {
	key := testKey(t)
	mac := HMACSHA256(key, []byte("test data"))
	if len(mac) != HMACFullSize {
		t.Errorf("HMAC size = %d, want %d", len(mac), HMACFullSize)
	}
}

func TestHMACSHA256_Deterministic(t *testing.T) {
	key := testKey(t)
	data := []byte("deterministic hmac test")
	a := HMACSHA256(key, data)
	b := HMACSHA256(key, data)
	if !bytes.Equal(a, b) {
		t.Error("HMACSHA256 is not deterministic")
	}
}

func TestHMACSHA256_DifferentKeys(t *testing.T) {
	key1 := testKey(t)
	key2 := testKey(t)
	data := []byte("same data different keys")
	a := HMACSHA256(key1, data)
	b := HMACSHA256(key2, data)
	if bytes.Equal(a, b) {
		t.Error("different keys produced same HMAC")
	}
}

func TestHMACSHA256Truncated(t *testing.T) {
	key := testKey(t)
	data := []byte("truncation test")

	tests := []struct {
		name    string
		n       int
		wantErr bool
	}{
		{"8 bytes (MyClerk Tier 2)", 8, false},
		{"16 bytes", 16, false},
		{"32 bytes (full)", 32, false},
		{"1 byte minimum", 1, false},
		{"0 bytes invalid", 0, true},
		{"33 bytes exceeds HMAC size", 33, true},
		{"-1 invalid", -1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HMACSHA256Truncated(key, data, tt.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && len(got) != tt.n {
				t.Errorf("truncated size = %d, want %d", len(got), tt.n)
			}
		})
	}
}

func TestHMACSHA256Truncated_PrefixOfFull(t *testing.T) {
	key := testKey(t)
	data := []byte("prefix consistency check")

	full := HMACSHA256(key, data)
	truncated, err := HMACSHA256Truncated(key, data, 8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(truncated, full[:8]) {
		t.Error("truncated HMAC is not a prefix of the full HMAC")
	}
}

func TestVerifyHMACSHA256(t *testing.T) {
	key := testKey(t)
	data := []byte("verification test")
	mac := HMACSHA256(key, data)

	if !VerifyHMACSHA256(key, data, mac) {
		t.Error("valid HMAC was rejected")
	}

	tampered := make([]byte, len(mac))
	copy(tampered, mac)
	tampered[0] ^= 0xFF
	if VerifyHMACSHA256(key, data, tampered) {
		t.Error("tampered HMAC was accepted")
	}

	wrongKey := testKey(t)
	if VerifyHMACSHA256(wrongKey, data, mac) {
		t.Error("wrong key HMAC was accepted")
	}
}

func TestVerifyHMACSHA256Truncated(t *testing.T) {
	key := testKey(t)
	data := []byte("truncated verification test")
	mac, err := HMACSHA256Truncated(key, data, 8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !VerifyHMACSHA256Truncated(key, data, mac) {
		t.Error("valid truncated HMAC was rejected")
	}

	tampered := make([]byte, len(mac))
	copy(tampered, mac)
	tampered[0] ^= 0xFF
	if VerifyHMACSHA256Truncated(key, data, tampered) {
		t.Error("tampered truncated HMAC was accepted")
	}
}

func TestVerifyHMACSHA256Truncated_InvalidSize(t *testing.T) {
	key := testKey(t)
	data := []byte("test")

	if VerifyHMACSHA256Truncated(key, data, nil) {
		t.Error("nil MAC should not verify")
	}
	if VerifyHMACSHA256Truncated(key, data, []byte{}) {
		t.Error("empty MAC should not verify")
	}
	oversized := make([]byte, 33)
	if VerifyHMACSHA256Truncated(key, data, oversized) {
		t.Error("oversized MAC should not verify")
	}
}

func BenchmarkHMACSHA256(b *testing.B) {
	key := make([]byte, 32)
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i)
	}
	b.ResetTimer()
	for b.Loop() {
		HMACSHA256(key, data)
	}
}
