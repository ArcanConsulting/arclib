package crypto

import (
	"bytes"
	"testing"
)

// TestAESKeyWrapRFC3394 checks AESKeyWrap/AESKeyUnwrap against the official
// RFC 3394 §4 test vectors (the canonical AES Key Wrap known-answer set).
func TestAESKeyWrapRFC3394(t *testing.T) {
	cases := []struct {
		name           string
		kek, key, want string
	}{
		{ // RFC 3394 §4.1: 128-bit KEK, 128-bit key.
			name: "128/128",
			kek:  "000102030405060708090A0B0C0D0E0F",
			key:  "00112233445566778899AABBCCDDEEFF",
			want: "1FA68B0A8112B447AEF34BD8FB5A7B829D3E862371D2CFE5",
		},
		{ // RFC 3394 §4.6: 256-bit KEK, 256-bit key.
			name: "256/256",
			kek:  "000102030405060708090A0B0C0D0E0F101112131415161718191A1B1C1D1E1F",
			key:  "00112233445566778899AABBCCDDEEFF000102030405060708090A0B0C0D0E0F",
			want: "28C9F404C4B810F4CBCCB35CFB87F8263F5786E2D80ED326CBC7F0E71A99F43BFB988B9B7A02DD21",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kek, key, want := mustHex(t, tc.kek), mustHex(t, tc.key), mustHex(t, tc.want)
			got, err := AESKeyWrap(kek, key)
			if err != nil {
				t.Fatalf("AESKeyWrap: %v", err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("AESKeyWrap = %X, want %X", got, want)
			}
			back, err := AESKeyUnwrap(kek, got)
			if err != nil {
				t.Fatalf("AESKeyUnwrap: %v", err)
			}
			if !bytes.Equal(back, key) {
				t.Fatalf("AESKeyUnwrap = %X, want %X", back, key)
			}
		})
	}
}

// TestAESKeyUnwrapIntegrity verifies a single bit flip is rejected (the RFC 3394
// integrity check), and that malformed lengths error rather than panic.
func TestAESKeyUnwrapIntegrity(t *testing.T) {
	kek := mustHex(t, "000102030405060708090A0B0C0D0E0F101112131415161718191A1B1C1D1E1F")
	wrapped, err := AESKeyWrap(kek, mustHex(t, "00112233445566778899AABBCCDDEEFF"))
	if err != nil {
		t.Fatalf("AESKeyWrap: %v", err)
	}
	tampered := append([]byte(nil), wrapped...)
	tampered[0] ^= 0x01
	if _, err := AESKeyUnwrap(kek, tampered); err == nil {
		t.Fatal("AESKeyUnwrap accepted a tampered wrapping")
	}
	if _, err := AESKeyUnwrap(kek, wrapped[:16]); err == nil {
		t.Fatal("AESKeyUnwrap accepted an under-length input")
	}
	if _, err := AESKeyWrap(kek, mustHex(t, "0011223344556677")); err == nil {
		t.Fatal("AESKeyWrap accepted single-block (under-16) key data")
	}
}
