package crypto

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("invalid hex: %v", err)
	}
	return b
}

// RFC 5869 test vectors for HKDF-SHA256.
func TestHKDFDeriveKey_RFC5869(t *testing.T) {
	tests := []struct {
		name   string
		ikm    string
		salt   string
		info   string
		length int
		okm    string
	}{
		{
			// RFC 5869 Test Case 1
			name:   "TC1 basic",
			ikm:    "0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b",
			salt:   "000102030405060708090a0b0c",
			info:   "f0f1f2f3f4f5f6f7f8f9",
			length: 42,
			okm:    "3cb25f25faacd57a90434f64d0362f2a2d2d0a90cf1a5a4c5db02d56ecc4c5bf34007208d5b887185865",
		},
		{
			// RFC 5869 Test Case 2
			name:   "TC2 longer inputs",
			ikm:    "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f404142434445464748494a4b4c4d4e4f",
			salt:   "606162636465666768696a6b6c6d6e6f707172737475767778797a7b7c7d7e7f808182838485868788898a8b8c8d8e8f909192939495969798999a9b9c9d9e9fa0a1a2a3a4a5a6a7a8a9aaabacadaeaf",
			info:   "b0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c8c9cacbcccdcecfd0d1d2d3d4d5d6d7d8d9dadbdcdddedfe0e1e2e3e4e5e6e7e8e9eaebecedeeeff0f1f2f3f4f5f6f7f8f9fafbfcfdfeff",
			length: 82,
			okm:    "b11e398dc80327a1c8e7f78c596a49344f012eda2d4efad8a050cc4c19afa97c59045a99cac7827271cb41c65e590e09da3275600c2f09b8367793a9aca3db71cc30c58179ec3e87c14c01d5c1f3434f1d87",
		},
		{
			// RFC 5869 Test Case 3 — zero-length salt and info
			name:   "TC3 empty salt and info",
			ikm:    "0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b",
			salt:   "",
			info:   "",
			length: 42,
			okm:    "8da4e775a563c18f715f802a063c5a31b8a11f5c5ee1879ec3454e5f3c738d2d9d201395faa4b61a96c8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ikm := mustHex(t, tt.ikm)
			var salt []byte
			if tt.salt != "" {
				salt = mustHex(t, tt.salt)
			}
			var info []byte
			if tt.info != "" {
				info = mustHex(t, tt.info)
			}
			expected := mustHex(t, tt.okm)

			got, err := HKDFDeriveKey(ikm, salt, info, tt.length)
			if err != nil {
				t.Fatalf("HKDFDeriveKey: %v", err)
			}

			if !bytes.Equal(got, expected) {
				t.Errorf("HKDFDeriveKey =\n  %s\nwant\n  %s",
					hex.EncodeToString(got), hex.EncodeToString(expected))
			}
		})
	}
}

func TestHKDFDeriveKey_Deterministic(t *testing.T) {
	ikm := []byte("input keying material")
	salt := []byte("salt")
	info := []byte("info")

	a, _ := HKDFDeriveKey(ikm, salt, info, 32)
	b, _ := HKDFDeriveKey(ikm, salt, info, 32)
	if !bytes.Equal(a, b) {
		t.Error("HKDF should be deterministic")
	}
}

func TestHKDFDeriveKey_DifferentInputsDifferentOutput(t *testing.T) {
	ikm := []byte("shared secret")
	a, _ := HKDFDeriveKey(ikm, nil, []byte("info-a"), 32)
	b, _ := HKDFDeriveKey(ikm, nil, []byte("info-b"), 32)
	if bytes.Equal(a, b) {
		t.Error("different info should produce different keys")
	}
}

func TestHKDFDeriveKey_NilSalt(t *testing.T) {
	ikm := []byte("shared secret")
	key, err := HKDFDeriveKey(ikm, nil, []byte("info"), 32)
	if err != nil {
		t.Fatalf("HKDFDeriveKey with nil salt: %v", err)
	}
	if len(key) != 32 {
		t.Errorf("key length = %d, want 32", len(key))
	}
}

func TestHKDFDeriveKeys(t *testing.T) {
	ikm := []byte("shared secret for multiple keys")
	salt := []byte("optional salt")
	info := []byte("myclerk-session-v1-hybrid")

	keys, err := HKDFDeriveKeys(ikm, salt, info, 32, 32, 12, 12)
	if err != nil {
		t.Fatalf("HKDFDeriveKeys: %v", err)
	}

	if len(keys) != 4 {
		t.Fatalf("expected 4 keys, got %d", len(keys))
	}

	expectedLengths := []int{32, 32, 12, 12}
	for i, expected := range expectedLengths {
		if len(keys[i]) != expected {
			t.Errorf("keys[%d] length = %d, want %d", i, len(keys[i]), expected)
		}
	}
}

func TestHKDFDeriveKeys_ConsistentWithDeriveKey(t *testing.T) {
	ikm := []byte("consistency check")
	info := []byte("test-info")

	single, _ := HKDFDeriveKey(ikm, nil, info, 88)

	multi, _ := HKDFDeriveKeys(ikm, nil, info, 32, 32, 12, 12)

	combined := make([]byte, 0, 88)
	for _, k := range multi {
		combined = append(combined, k...)
	}

	if !bytes.Equal(single, combined) {
		t.Error("HKDFDeriveKeys should produce the same bytes as a single HKDFDeriveKey of the total length")
	}
}

func TestHKDFDeriveKeys_AllDifferent(t *testing.T) {
	ikm := []byte("different keys test")
	info := []byte("myclerk-session-v1-classical")

	keys, _ := HKDFDeriveKeys(ikm, nil, info, 32, 32, 12, 12)

	if bytes.Equal(keys[0], keys[1]) {
		t.Error("send key and receive key should differ")
	}
	if bytes.Equal(keys[2], keys[3]) {
		t.Error("send nonce seed and receive nonce seed should differ")
	}
}

func TestHKDFDeriveKeys_Deterministic(t *testing.T) {
	ikm := []byte("deterministic test")
	info := []byte("test")

	a, _ := HKDFDeriveKeys(ikm, nil, info, 32, 32)
	b, _ := HKDFDeriveKeys(ikm, nil, info, 32, 32)

	if !bytes.Equal(a[0], b[0]) || !bytes.Equal(a[1], b[1]) {
		t.Error("HKDFDeriveKeys should be deterministic")
	}
}

func TestHKDFDeriveKeys_UsableForEncryption(t *testing.T) {
	ikm := []byte("secret shared between two parties")
	info := []byte("myclerk-session-v1-hybrid")

	keys, err := HKDFDeriveKeys(ikm, nil, info, KeySize, KeySize, NonceSize, NonceSize)
	if err != nil {
		t.Fatalf("HKDFDeriveKeys: %v", err)
	}

	aead, err := NewAEAD(keys[0])
	if err != nil {
		t.Fatalf("NewAEAD with derived key: %v", err)
	}

	plaintext := []byte("encrypted with HKDF-derived key")
	var nonce [NonceSize]byte
	copy(nonce[:], keys[2])

	ciphertext, err := aead.Encrypt(nonce[:], plaintext, nil)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	decrypted, err := aead.Decrypt(nonce[:], ciphertext, nil)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
	}
}

func BenchmarkHKDFDeriveKey_32(b *testing.B) {
	ikm := make([]byte, 64)
	info := []byte("myclerk-session-v1-hybrid")
	for b.Loop() {
		HKDFDeriveKey(ikm, nil, info, 32) //nolint:errcheck // benchmark
	}
}

func BenchmarkHKDFDeriveKeys_88(b *testing.B) {
	ikm := make([]byte, 64)
	info := []byte("myclerk-session-v1-hybrid")
	for b.Loop() {
		HKDFDeriveKeys(ikm, nil, info, 32, 32, 12, 12) //nolint:errcheck // benchmark
	}
}
