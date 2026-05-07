package crypto

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	if len(key) != KeySize {
		t.Errorf("key length = %d, want %d", len(key), KeySize)
	}

	var zero [KeySize]byte
	if key == zero {
		t.Error("key should not be all zeros")
	}
}

func TestGenerateKey_Unique(t *testing.T) {
	key1, _ := GenerateKey()
	key2, _ := GenerateKey()
	if key1 == key2 {
		t.Error("two generated keys should not be equal")
	}
}

func TestGenerateNonce(t *testing.T) {
	nonce, err := GenerateNonce()
	if err != nil {
		t.Fatalf("GenerateNonce: %v", err)
	}

	var zero [NonceSize]byte
	if nonce == zero {
		t.Error("nonce should not be all zeros")
	}
}

func TestGenerateNonce_Unique(t *testing.T) {
	nonce1, _ := GenerateNonce()
	nonce2, _ := GenerateNonce()
	if nonce1 == nonce2 {
		t.Error("two generated nonces should not be equal")
	}
}

func TestNonceFromSequence(t *testing.T) {
	tests := []struct {
		name      string
		sequence  uint32
		sessionID uint64
	}{
		{"zeros", 0, 0},
		{"ones", 1, 1},
		{"max values", 0xFFFFFFFF, 0xFFFFFFFFFFFFFFFF},
		{"typical", 12345, 9876543210},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nonce := NonceFromSequence(tt.sequence, tt.sessionID)

			if len(nonce) != NonceSize {
				t.Errorf("nonce length = %d, want %d", len(nonce), NonceSize)
			}

			gotSeq := binary.BigEndian.Uint32(nonce[0:4])
			if gotSeq != tt.sequence {
				t.Errorf("sequence in nonce = %d, want %d", gotSeq, tt.sequence)
			}

			gotSID := binary.BigEndian.Uint64(nonce[4:12])
			if gotSID != tt.sessionID {
				t.Errorf("sessionID in nonce = %d, want %d", gotSID, tt.sessionID)
			}
		})
	}
}

func TestNonceFromSequence_Deterministic(t *testing.T) {
	a := NonceFromSequence(42, 12345)
	b := NonceFromSequence(42, 12345)
	if a != b {
		t.Error("same inputs should produce same nonce")
	}
}

func TestNonceFromSequence_Unique(t *testing.T) {
	n1 := NonceFromSequence(1, 1)
	n2 := NonceFromSequence(2, 1)
	n3 := NonceFromSequence(1, 2)

	if n1 == n2 {
		t.Error("different sequence should produce different nonce")
	}
	if n1 == n3 {
		t.Error("different sessionID should produce different nonce")
	}
}

func TestNonceFromSequence_UsableForEncryption(t *testing.T) {
	key, _ := GenerateKey()
	aead, _ := NewAEAD(key[:])

	nonce := NonceFromSequence(1, 1000)
	plaintext := []byte("encrypted with derived nonce")

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

func TestZeroBytes(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0xFF, 0xAB}
	ZeroBytes(data)

	for i, b := range data {
		if b != 0 {
			t.Errorf("byte %d = 0x%02X, want 0x00", i, b)
		}
	}
}

func TestZeroBytes_Empty(t *testing.T) {
	ZeroBytes(nil)
	ZeroBytes([]byte{})
}
