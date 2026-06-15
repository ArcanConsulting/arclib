package crypto

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// TestANSIX963KDFKnownAnswer checks ANSIX963KDF against an independently derived
// known answer. The X9.63 KDF with an empty SharedInfo is, by definition,
// SHA256(Z‖00000001) ‖ SHA256(Z‖00000002) ‖ …; the expected values below were
// computed out-of-band with sha256sum over Z=0x00..0x1F, so this pins both the
// single-block and the multi-block (counter increment) paths.
func TestANSIX963KDFKnownAnswer(t *testing.T) {
	z := make([]byte, 32)
	for i := range z {
		z[i] = byte(i)
	}
	cases := []struct {
		keyLen int
		want   string
	}{
		{16, "04a6950a06d3e3308ad7d3606ef810eb"},
		{48, "04a6950a06d3e3308ad7d3606ef810eb124e3943404ca746a12c51c7bf7768390f8d842ac9cb62349779a7537a78327d"},
	}
	for _, tc := range cases {
		got, err := ANSIX963KDF(sha256.New, z, nil, tc.keyLen)
		if err != nil {
			t.Fatalf("ANSIX963KDF(%d): %v", tc.keyLen, err)
		}
		want, _ := hex.DecodeString(tc.want)
		if !bytes.Equal(got, want) {
			t.Errorf("ANSIX963KDF(%d) = %x, want %s", tc.keyLen, got, tc.want)
		}
	}
}

// TestANSIX963KDFSharedInfo checks that SharedInfo is mixed in (a different
// SharedInfo yields a different key) and that an invalid length errors.
func TestANSIX963KDFSharedInfo(t *testing.T) {
	z := bytes.Repeat([]byte{0x11}, 32)
	a, err := ANSIX963KDF(sha256.New, z, []byte("info-A"), 32)
	if err != nil {
		t.Fatalf("kdf A: %v", err)
	}
	b, err := ANSIX963KDF(sha256.New, z, []byte("info-B"), 32)
	if err != nil {
		t.Fatalf("kdf B: %v", err)
	}
	if bytes.Equal(a, b) {
		t.Error("ANSIX963KDF ignored SharedInfo (A and B match)")
	}
	if _, err := ANSIX963KDF(sha256.New, z, nil, 0); err == nil {
		t.Error("ANSIX963KDF accepted a non-positive key length")
	}
}
