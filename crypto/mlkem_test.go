package crypto

import (
	"bytes"
	"testing"
)

func TestGenerateMLKEMKeyPair(t *testing.T) {
	kp, err := GenerateMLKEMKeyPair()
	if err != nil {
		t.Fatalf("GenerateMLKEMKeyPair: %v", err)
	}
	if kp.DecapKey == nil {
		t.Fatal("DecapKey should not be nil")
	}
}

func TestMLKEMKeyPair_EncapsulationKeyBytes(t *testing.T) {
	kp, _ := GenerateMLKEMKeyPair()
	ekBytes := kp.EncapsulationKeyBytes()

	if len(ekBytes) != MLKEMEncapKeySize {
		t.Errorf("encap key size = %d, want %d", len(ekBytes), MLKEMEncapKeySize)
	}
}

func TestMLKEMKeyPair_EncapsulationKeyDeterministic(t *testing.T) {
	kp, _ := GenerateMLKEMKeyPair()
	a := kp.EncapsulationKeyBytes()
	b := kp.EncapsulationKeyBytes()

	if !bytes.Equal(a, b) {
		t.Error("encapsulation key should be deterministic for same key pair")
	}
}

func TestMLKEM_EncapsulateDecapsulate(t *testing.T) {
	kp, _ := GenerateMLKEMKeyPair()
	ekBytes := kp.EncapsulationKeyBytes()

	sharedEncap, ciphertext, err := MLKEMEncapsulate(ekBytes)
	if err != nil {
		t.Fatalf("MLKEMEncapsulate: %v", err)
	}

	if len(ciphertext) != MLKEMCiphertextSize {
		t.Errorf("ciphertext size = %d, want %d", len(ciphertext), MLKEMCiphertextSize)
	}

	sharedDecap, err := kp.Decapsulate(ciphertext)
	if err != nil {
		t.Fatalf("Decapsulate: %v", err)
	}

	if sharedEncap != sharedDecap {
		t.Error("encapsulated and decapsulated shared secrets must match")
	}

	var zero [MLKEMSharedKeySize]byte
	if sharedEncap == zero {
		t.Error("shared secret should not be all zeros")
	}
}

func TestMLKEM_DifferentKeyPairsDifferentSecrets(t *testing.T) {
	kp1, _ := GenerateMLKEMKeyPair()
	kp2, _ := GenerateMLKEMKeyPair()

	ss1, _, _ := MLKEMEncapsulate(kp1.EncapsulationKeyBytes())
	ss2, _, _ := MLKEMEncapsulate(kp2.EncapsulationKeyBytes())

	if ss1 == ss2 {
		t.Error("different key pairs should produce different shared secrets")
	}
}

func TestMLKEM_EachEncapsulationUnique(t *testing.T) {
	kp, _ := GenerateMLKEMKeyPair()
	ekBytes := kp.EncapsulationKeyBytes()

	ss1, ct1, _ := MLKEMEncapsulate(ekBytes)
	ss2, ct2, _ := MLKEMEncapsulate(ekBytes)

	if bytes.Equal(ct1, ct2) {
		t.Error("each encapsulation should produce unique ciphertext")
	}
	if ss1 == ss2 {
		t.Error("each encapsulation should produce unique shared secret")
	}
}

func TestMLKEMEncapsulate_InvalidKey(t *testing.T) {
	_, _, err := MLKEMEncapsulate([]byte("too short"))
	if err != ErrMLKEMEncapsulationKey {
		t.Errorf("MLKEMEncapsulate(bad key) error = %v, want ErrMLKEMEncapsulationKey", err)
	}

	_, _, err = MLKEMEncapsulate(nil)
	if err != ErrMLKEMEncapsulationKey {
		t.Errorf("MLKEMEncapsulate(nil) error = %v, want ErrMLKEMEncapsulationKey", err)
	}
}

func TestMLKEM_DecapsulateTamperedCiphertext(t *testing.T) {
	kp, _ := GenerateMLKEMKeyPair()
	ekBytes := kp.EncapsulationKeyBytes()

	sharedEncap, ciphertext, _ := MLKEMEncapsulate(ekBytes)

	tampered := make([]byte, len(ciphertext))
	copy(tampered, ciphertext)
	tampered[0] ^= 0xFF

	sharedTampered, err := kp.Decapsulate(tampered)
	// ML-KEM uses implicit rejection: tampered ciphertext returns a
	// deterministic but incorrect shared secret rather than an error.
	// Per FIPS 203, this prevents chosen-ciphertext attacks.
	if err != nil {
		t.Fatalf("Decapsulate should succeed with implicit rejection, got: %v", err)
	}
	if sharedTampered == sharedEncap {
		t.Error("tampered ciphertext should produce different shared secret (implicit rejection)")
	}
}

func TestMLKEM_HybridWithX25519(t *testing.T) {
	// Simulate hybrid key exchange: X25519 + ML-KEM-768
	alice25519, _ := GenerateX25519KeyPair()
	bob25519, _ := GenerateX25519KeyPair()
	aliceMLKEM, _ := GenerateMLKEMKeyPair()

	// Bob encapsulates against Alice's ML-KEM key
	mlkemSharedBob, ciphertext, err := MLKEMEncapsulate(aliceMLKEM.EncapsulationKeyBytes())
	if err != nil {
		t.Fatalf("MLKEMEncapsulate: %v", err)
	}

	// Alice decapsulates
	mlkemSharedAlice, err := aliceMLKEM.Decapsulate(ciphertext)
	if err != nil {
		t.Fatalf("Decapsulate: %v", err)
	}

	// X25519 shared secrets
	x25519SharedAlice, _ := alice25519.ECDH(bob25519.Public)
	x25519SharedBob, _ := bob25519.ECDH(alice25519.Public)

	// Combine: X25519 || ML-KEM
	combinedAlice := make([]byte, 0, 64)
	combinedAlice = append(combinedAlice, x25519SharedAlice[:]...)
	combinedAlice = append(combinedAlice, mlkemSharedAlice[:]...)

	combinedBob := make([]byte, 0, 64)
	combinedBob = append(combinedBob, x25519SharedBob[:]...)
	combinedBob = append(combinedBob, mlkemSharedBob[:]...)

	if !bytes.Equal(combinedAlice, combinedBob) {
		t.Error("hybrid combined secrets must match")
	}

	// Derive session keys via HKDF
	info := []byte("myclerk-session-v1-hybrid")
	keysAlice, err := HKDFDeriveKeys(combinedAlice, nil, info, KeySize, KeySize, NonceSize, NonceSize)
	if err != nil {
		t.Fatalf("HKDFDeriveKeys: %v", err)
	}

	keysBob, err := HKDFDeriveKeys(combinedBob, nil, info, KeySize, KeySize, NonceSize, NonceSize)
	if err != nil {
		t.Fatalf("HKDFDeriveKeys: %v", err)
	}

	if !bytes.Equal(keysAlice[0], keysBob[0]) {
		t.Error("derived send keys must match")
	}

	// Verify encryption works end-to-end
	aead, _ := NewAEAD(keysAlice[0])
	var nonce [NonceSize]byte
	copy(nonce[:], keysAlice[2])

	ct, _ := aead.Encrypt(nonce[:], []byte("post-quantum secure"), nil)

	aeadBob, _ := NewAEAD(keysBob[0])
	pt, err := aeadBob.Decrypt(nonce[:], ct, nil)
	if err != nil {
		t.Fatalf("hybrid pipeline decrypt failed: %v", err)
	}
	if !bytes.Equal(pt, []byte("post-quantum secure")) {
		t.Error("hybrid pipeline: decrypted text mismatch")
	}
}

func BenchmarkMLKEMGenerateKey(b *testing.B) {
	for b.Loop() {
		GenerateMLKEMKeyPair() //nolint:errcheck // benchmark: error is crypto/rand failure
	}
}

func BenchmarkMLKEMEncapsulate(b *testing.B) {
	kp, _ := GenerateMLKEMKeyPair()
	ekBytes := kp.EncapsulationKeyBytes()

	b.ResetTimer()
	for b.Loop() {
		MLKEMEncapsulate(ekBytes) //nolint:errcheck // benchmark
	}
}

func BenchmarkMLKEMDecapsulate(b *testing.B) {
	kp, _ := GenerateMLKEMKeyPair()
	_, ct, _ := MLKEMEncapsulate(kp.EncapsulationKeyBytes())

	b.ResetTimer()
	for b.Loop() {
		kp.Decapsulate(ct) //nolint:errcheck // benchmark
	}
}
