package crypto

import (
	"bytes"
	"testing"
)

func TestMLDSASignVerifyRoundTrip(t *testing.T) {
	kp, err := GenerateMLDSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateMLDSAKeyPair: %v", err)
	}
	if got := len(kp.PublicKeyBytes()); got != MLDSAPublicKeySize {
		t.Fatalf("public key size = %d, want %d", got, MLDSAPublicKeySize)
	}

	msg := []byte("the quick brown fox")
	sig, err := kp.Sign(msg)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if len(sig) != MLDSASignatureSize {
		t.Fatalf("signature size = %d, want %d", len(sig), MLDSASignatureSize)
	}
	if !MLDSAVerify(kp.PublicKeyBytes(), msg, sig) {
		t.Fatal("MLDSAVerify rejected a valid signature")
	}
}

func TestMLDSAFromSeedDeterministic(t *testing.T) {
	var seed [MLDSASeedSize]byte
	for i := range seed {
		seed[i] = byte(i)
	}
	a := MLDSAKeyPairFromSeed(&seed)
	b := MLDSAKeyPairFromSeed(&seed)
	if !bytes.Equal(a.PublicKeyBytes(), b.PublicKeyBytes()) {
		t.Fatal("same seed produced different public keys")
	}
	// A signature from one key must verify under the other (same key pair).
	sig, err := a.Sign([]byte("x"))
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if !MLDSAVerify(b.PublicKeyBytes(), []byte("x"), sig) {
		t.Fatal("seed-derived key pairs disagree")
	}
}

func TestMLDSAVerifyRejects(t *testing.T) {
	kp, err := GenerateMLDSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateMLDSAKeyPair: %v", err)
	}
	msg := []byte("authentic")
	sig, err := kp.Sign(msg)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	pub := kp.PublicKeyBytes()

	if MLDSAVerify(pub, []byte("tampered"), sig) {
		t.Error("verified a signature over the wrong message")
	}
	bad := bytes.Clone(sig)
	bad[0] ^= 0xFF
	if MLDSAVerify(pub, msg, bad) {
		t.Error("verified a tampered signature")
	}
	if MLDSAVerify(pub[:len(pub)-1], msg, sig) {
		t.Error("verified under a truncated public key")
	}
	if MLDSAVerify(pub, msg, sig[:len(sig)-1]) {
		t.Error("verified a truncated signature")
	}

	other, err := GenerateMLDSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateMLDSAKeyPair: %v", err)
	}
	if MLDSAVerify(other.PublicKeyBytes(), msg, sig) {
		t.Error("verified under the wrong public key")
	}
}
