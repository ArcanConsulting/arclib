package crypto

import (
	"bytes"
	"testing"
)

func TestRotateKey(t *testing.T) {
	key, _ := GenerateKey()

	rotated, err := RotateKey(key, 0)
	if err != nil {
		t.Fatalf("RotateKey: %v", err)
	}

	if rotated == key {
		t.Error("rotated key should differ from original")
	}

	var zero [KeySize]byte
	if rotated == zero {
		t.Error("rotated key should not be zero")
	}
}

func TestRotateKey_Deterministic(t *testing.T) {
	key, _ := GenerateKey()

	a, _ := RotateKey(key, 1)
	b, _ := RotateKey(key, 1)

	if a != b {
		t.Error("same key + same counter must produce same result")
	}
}

func TestRotateKey_DifferentCounters(t *testing.T) {
	key, _ := GenerateKey()

	r0, _ := RotateKey(key, 0)
	r1, _ := RotateKey(key, 1)
	r2, _ := RotateKey(key, 2)

	if r0 == r1 {
		t.Error("different counters should produce different keys")
	}
	if r1 == r2 {
		t.Error("different counters should produce different keys")
	}
	if r0 == r2 {
		t.Error("different counters should produce different keys")
	}
}

func TestRotateKey_DifferentBaseKeys(t *testing.T) {
	key1, _ := GenerateKey()
	key2, _ := GenerateKey()

	r1, _ := RotateKey(key1, 0)
	r2, _ := RotateKey(key2, 0)

	if r1 == r2 {
		t.Error("different base keys should produce different rotated keys")
	}
}

func TestRotateKey_ChainedRotation(t *testing.T) {
	key, _ := GenerateKey()

	k1, _ := RotateKey(key, 0)
	k2, _ := RotateKey(k1, 1)
	k3, _ := RotateKey(k2, 2)

	if k1 == k2 || k2 == k3 || k1 == k3 {
		t.Error("chained rotations should produce distinct keys")
	}

	if k1 == key || k2 == key || k3 == key {
		t.Error("no rotated key should equal the original")
	}
}

func TestRotateKey_UsableForEncryption(t *testing.T) {
	key, _ := GenerateKey()
	rotated, _ := RotateKey(key, 42)

	aead, err := NewAEAD(rotated[:])
	if err != nil {
		t.Fatalf("NewAEAD with rotated key: %v", err)
	}

	nonce, _ := GenerateNonce()
	plaintext := []byte("encrypted with rotated key")

	ct, err := aead.Encrypt(nonce[:], plaintext, nil)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	pt, err := aead.Decrypt(nonce[:], ct, nil)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if !bytes.Equal(pt, plaintext) {
		t.Error("decrypt after rotation failed")
	}
}

func TestRotateKeyN(t *testing.T) {
	key, _ := GenerateKey()

	// RotateKeyN(key, 0, 3) should equal three sequential RotateKey calls
	k1, _ := RotateKey(key, 0)
	k2, _ := RotateKey(k1, 1)
	k3, _ := RotateKey(k2, 2)

	result, err := RotateKeyN(key, 0, 3)
	if err != nil {
		t.Fatalf("RotateKeyN: %v", err)
	}

	if result != k3 {
		t.Error("RotateKeyN(3) should equal three sequential RotateKey calls")
	}
}

func TestRotateKeyN_Zero(t *testing.T) {
	key, _ := GenerateKey()

	result, err := RotateKeyN(key, 0, 0)
	if err != nil {
		t.Fatalf("RotateKeyN: %v", err)
	}

	if result != key {
		t.Error("RotateKeyN(0) should return the original key")
	}
}

func TestRotateKeyN_One(t *testing.T) {
	key, _ := GenerateKey()

	single, _ := RotateKey(key, 5)
	multi, _ := RotateKeyN(key, 5, 1)

	if single != multi {
		t.Error("RotateKeyN(1) should equal single RotateKey")
	}
}

func BenchmarkRotateKey(b *testing.B) {
	key, _ := GenerateKey()
	for b.Loop() {
		RotateKey(key, 0) //nolint:errcheck // benchmark
	}
}

func BenchmarkRotateKeyN_10(b *testing.B) {
	key, _ := GenerateKey()
	for b.Loop() {
		RotateKeyN(key, 0, 10) //nolint:errcheck // benchmark
	}
}
