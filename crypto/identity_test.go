package crypto

import (
	"bytes"
	"crypto/ed25519"
	"testing"
)

func TestIdentitySignVerifyRoundTrip(t *testing.T) {
	id, err := GenerateIdentityKey()
	if err != nil {
		t.Fatalf("GenerateIdentityKey: %v", err)
	}
	pub := id.PublicKey()
	if len(pub) != IdentityPublicKeySize {
		t.Fatalf("public key size = %d, want %d", len(pub), IdentityPublicKeySize)
	}

	msg := []byte("handshake transcript digest")
	sig, err := id.Sign(msg)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if len(sig) != IdentitySignatureSize {
		t.Fatalf("signature size = %d, want %d", len(sig), IdentitySignatureSize)
	}
	if !IdentityVerify(pub, msg, sig) {
		t.Fatal("IdentityVerify rejected a valid hybrid signature")
	}
}

func TestIdentityFromSeedDeterministic(t *testing.T) {
	id, err := GenerateIdentityKey()
	if err != nil {
		t.Fatalf("GenerateIdentityKey: %v", err)
	}
	seed := id.Seed()
	clone, err := IdentityKeyFromSeed(seed)
	if err != nil {
		t.Fatalf("IdentityKeyFromSeed: %v", err)
	}
	if !bytes.Equal(id.PublicKey(), clone.PublicKey()) {
		t.Fatal("same seed produced different public keys")
	}
	// A signature from the clone must verify under the original's public key.
	msg := []byte("x")
	sig, err := clone.Sign(msg)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if !IdentityVerify(id.PublicKey(), msg, sig) {
		t.Fatal("seed-derived identities disagree")
	}
}

func TestIdentityVerifyRejectsTamper(t *testing.T) {
	id, err := GenerateIdentityKey()
	if err != nil {
		t.Fatalf("GenerateIdentityKey: %v", err)
	}
	pub := id.PublicKey()
	msg := []byte("authentic")
	sig, err := id.Sign(msg)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	if IdentityVerify(pub, []byte("tampered"), sig) {
		t.Error("verified over the wrong message")
	}
	if IdentityVerify(pub[:len(pub)-1], msg, sig) {
		t.Error("verified under a malformed public key length")
	}
	if IdentityVerify(pub, msg, sig[:len(sig)-1]) {
		t.Error("verified a malformed signature length")
	}

	other, err := GenerateIdentityKey()
	if err != nil {
		t.Fatalf("GenerateIdentityKey: %v", err)
	}
	if IdentityVerify(other.PublicKey(), msg, sig) {
		t.Error("verified under the wrong identity")
	}
}

// TestIdentityRequiresBothComponents proves the hybrid is an AND: corrupting
// only the Ed25519 half or only the ML-DSA half must fail verification, so a
// forgery requires breaking both primitives.
func TestIdentityRequiresBothComponents(t *testing.T) {
	id, err := GenerateIdentityKey()
	if err != nil {
		t.Fatalf("GenerateIdentityKey: %v", err)
	}
	pub := id.PublicKey()
	msg := []byte("authentic")
	sig, err := id.Sign(msg)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	// Corrupt only the Ed25519 signature component (first SignatureSize bytes).
	edBad := bytes.Clone(sig)
	edBad[0] ^= 0xFF
	if IdentityVerify(pub, msg, edBad) {
		t.Error("verified with a corrupt Ed25519 component")
	}

	// Corrupt only the ML-DSA signature component (after the Ed25519 part).
	pqBad := bytes.Clone(sig)
	pqBad[ed25519.SignatureSize] ^= 0xFF
	if IdentityVerify(pub, msg, pqBad) {
		t.Error("verified with a corrupt ML-DSA component")
	}
}

func TestIdentityKeyFromSeedRejectsAndZero(t *testing.T) {
	id, err := GenerateIdentityKey()
	if err != nil {
		t.Fatalf("GenerateIdentityKey: %v", err)
	}
	id.Zero()
	if seed := id.Seed(); seed != ([IdentitySeedSize]byte{}) {
		t.Error("Zero did not wipe the seed")
	}
}
