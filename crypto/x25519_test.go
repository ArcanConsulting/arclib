package crypto

import (
	"bytes"
	"testing"
)

func TestGenerateX25519KeyPair(t *testing.T) {
	kp, err := GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf("GenerateX25519KeyPair: %v", err)
	}

	var zero [X25519KeySize]byte
	if kp.Private == zero {
		t.Error("private key should not be zero")
	}
	if kp.Public == zero {
		t.Error("public key should not be zero")
	}
}

func TestGenerateX25519KeyPair_Clamping(t *testing.T) {
	kp, _ := GenerateX25519KeyPair()

	if kp.Private[0]&7 != 0 {
		t.Error("lower 3 bits of first byte should be cleared")
	}
	if kp.Private[31]&128 != 0 {
		t.Error("bit 255 should be cleared")
	}
	if kp.Private[31]&64 == 0 {
		t.Error("bit 254 should be set")
	}
}

func TestGenerateX25519KeyPair_Unique(t *testing.T) {
	kp1, _ := GenerateX25519KeyPair()
	kp2, _ := GenerateX25519KeyPair()

	if kp1.Private == kp2.Private {
		t.Error("private keys should be unique")
	}
	if kp1.Public == kp2.Public {
		t.Error("public keys should be unique")
	}
}

func TestX25519KeyPairFromPrivate(t *testing.T) {
	original, _ := GenerateX25519KeyPair()
	recreated := X25519KeyPairFromPrivate(original.Private)

	if original.Public != recreated.Public {
		t.Error("public key should be derived identically from same private key")
	}
}

func TestX25519KeyPairFromPrivate_Clamps(t *testing.T) {
	var unclamped [X25519KeySize]byte
	unclamped[0] = 0xFF
	unclamped[31] = 0xFF

	kp := X25519KeyPairFromPrivate(unclamped)

	if kp.Private[0]&7 != 0 {
		t.Error("should clamp lower 3 bits")
	}
	if kp.Private[31]&128 != 0 {
		t.Error("should clamp bit 255")
	}
	if kp.Private[31]&64 == 0 {
		t.Error("should set bit 254")
	}
}

func TestECDH_SharedSecretAgreement(t *testing.T) {
	alice, _ := GenerateX25519KeyPair()
	bob, _ := GenerateX25519KeyPair()

	aliceShared, err := alice.ECDH(bob.Public)
	if err != nil {
		t.Fatalf("Alice ECDH: %v", err)
	}

	bobShared, err := bob.ECDH(alice.Public)
	if err != nil {
		t.Fatalf("Bob ECDH: %v", err)
	}

	if aliceShared != bobShared {
		t.Error("shared secrets must match")
	}

	var zero [SharedSecretSize]byte
	if aliceShared == zero {
		t.Error("shared secret should not be all zeros")
	}
}

func TestECDH_DifferentPeers(t *testing.T) {
	alice, _ := GenerateX25519KeyPair()
	bob, _ := GenerateX25519KeyPair()
	carol, _ := GenerateX25519KeyPair()

	sharedAB, _ := alice.ECDH(bob.Public)
	sharedAC, _ := alice.ECDH(carol.Public)

	if sharedAB == sharedAC {
		t.Error("shared secrets with different peers should differ")
	}
}

func TestECDH_Deterministic(t *testing.T) {
	alice, _ := GenerateX25519KeyPair()
	bob, _ := GenerateX25519KeyPair()

	s1, _ := alice.ECDH(bob.Public)
	s2, _ := alice.ECDH(bob.Public)

	if s1 != s2 {
		t.Error("ECDH should be deterministic for same inputs")
	}
}

func TestECDH_InvalidPublicKey(t *testing.T) {
	alice, _ := GenerateX25519KeyPair()

	var zeroKey [X25519KeySize]byte
	_, err := alice.ECDH(zeroKey)
	if err != ErrInvalidPublicKey {
		t.Errorf("ECDH(zero key) error = %v, want ErrInvalidPublicKey", err)
	}
}

func TestECDH_UsableWithAEAD(t *testing.T) {
	alice, _ := GenerateX25519KeyPair()
	bob, _ := GenerateX25519KeyPair()

	shared, _ := alice.ECDH(bob.Public)

	key, err := HKDFDeriveKey(shared[:], nil, []byte("test"), KeySize)
	if err != nil {
		t.Fatalf("HKDFDeriveKey: %v", err)
	}

	aead, err := NewAEAD(key)
	if err != nil {
		t.Fatalf("NewAEAD: %v", err)
	}

	nonce, _ := GenerateNonce()
	plaintext := []byte("encrypted with X25519-derived key")

	ciphertext, err := aead.Encrypt(nonce[:], plaintext, nil)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	decrypted, err := aead.Decrypt(nonce[:], ciphertext, nil)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("full pipeline X25519 → HKDF → AEAD failed")
	}
}

func BenchmarkGenerateX25519KeyPair(b *testing.B) {
	for b.Loop() {
		GenerateX25519KeyPair() //nolint:errcheck // benchmark: error path is crypto/rand failure
	}
}

func BenchmarkECDH(b *testing.B) {
	alice, _ := GenerateX25519KeyPair()
	bob, _ := GenerateX25519KeyPair()

	b.ResetTimer()
	for b.Loop() {
		alice.ECDH(bob.Public) //nolint:errcheck // benchmark
	}
}
