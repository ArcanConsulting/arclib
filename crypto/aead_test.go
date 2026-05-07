package crypto

import (
	"bytes"
	"testing"
)

func TestNewAEAD(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	aead, err := NewAEAD(key[:])
	if err != nil {
		t.Fatalf("NewAEAD: %v", err)
	}
	if aead == nil {
		t.Fatal("NewAEAD returned nil")
	}
}

func TestNewAEAD_InvalidKeySize(t *testing.T) {
	sizes := []int{0, 1, 16, 31, 33, 64}
	for _, size := range sizes {
		key := make([]byte, size)
		_, err := NewAEAD(key)
		if err != ErrInvalidKeySize {
			t.Errorf("NewAEAD(key[%d]) error = %v, want ErrInvalidKeySize", size, err)
		}
	}
}

func TestAEAD_EncryptDecrypt(t *testing.T) {
	key, _ := GenerateKey()
	aead, _ := NewAEAD(key[:])

	tests := []struct {
		name      string
		plaintext []byte
		aad       []byte
	}{
		{"normal message", []byte("Hello, MyClerk Protocol!"), []byte("additional data")},
		{"empty plaintext", []byte{}, nil},
		{"empty AAD", []byte("plaintext"), nil},
		{"both empty", []byte{}, []byte{}},
		{"large plaintext", bytes.Repeat([]byte("A"), 65536), []byte("bulk")},
		{"binary data", []byte{0x00, 0xFF, 0x80, 0x7F, 0x01}, []byte{0xDE, 0xAD}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nonce, _ := GenerateNonce()

			ciphertext, err := aead.Encrypt(nonce[:], tt.plaintext, tt.aad)
			if err != nil {
				t.Fatalf("Encrypt: %v", err)
			}

			expectedLen := len(tt.plaintext) + TagSize
			if len(ciphertext) != expectedLen {
				t.Errorf("ciphertext length = %d, want %d", len(ciphertext), expectedLen)
			}

			decrypted, err := aead.Decrypt(nonce[:], ciphertext, tt.aad)
			if err != nil {
				t.Fatalf("Decrypt: %v", err)
			}

			if !bytes.Equal(decrypted, tt.plaintext) {
				t.Errorf("decrypted = %q, want %q", decrypted, tt.plaintext)
			}
		})
	}
}

func TestAEAD_CiphertextDiffersFromPlaintext(t *testing.T) {
	key, _ := GenerateKey()
	aead, _ := NewAEAD(key[:])
	nonce, _ := GenerateNonce()

	plaintext := []byte("secret message that must not appear in ciphertext")
	ciphertext, _ := aead.Encrypt(nonce[:], plaintext, nil)

	if bytes.Equal(ciphertext[:len(plaintext)], plaintext) {
		t.Error("ciphertext should not equal plaintext")
	}
}

func TestAEAD_DecryptWrongKey(t *testing.T) {
	key1, _ := GenerateKey()
	key2, _ := GenerateKey()
	aead1, _ := NewAEAD(key1[:])
	aead2, _ := NewAEAD(key2[:])

	nonce, _ := GenerateNonce()
	ciphertext, _ := aead1.Encrypt(nonce[:], []byte("secret"), nil)

	_, err := aead2.Decrypt(nonce[:], ciphertext, nil)
	if err != ErrAuthFailed {
		t.Errorf("Decrypt(wrong key) error = %v, want ErrAuthFailed", err)
	}
}

func TestAEAD_DecryptTamperedCiphertext(t *testing.T) {
	key, _ := GenerateKey()
	aead, _ := NewAEAD(key[:])
	nonce, _ := GenerateNonce()

	ciphertext, _ := aead.Encrypt(nonce[:], []byte("secret"), nil)

	tampered := make([]byte, len(ciphertext))
	copy(tampered, ciphertext)
	tampered[0] ^= 0xFF

	_, err := aead.Decrypt(nonce[:], tampered, nil)
	if err != ErrAuthFailed {
		t.Errorf("Decrypt(tampered) error = %v, want ErrAuthFailed", err)
	}
}

func TestAEAD_DecryptWrongAAD(t *testing.T) {
	key, _ := GenerateKey()
	aead, _ := NewAEAD(key[:])
	nonce, _ := GenerateNonce()

	ciphertext, _ := aead.Encrypt(nonce[:], []byte("secret"), []byte("correct"))

	_, err := aead.Decrypt(nonce[:], ciphertext, []byte("wrong"))
	if err != ErrAuthFailed {
		t.Errorf("Decrypt(wrong AAD) error = %v, want ErrAuthFailed", err)
	}
}

func TestAEAD_DecryptWrongNonce(t *testing.T) {
	key, _ := GenerateKey()
	aead, _ := NewAEAD(key[:])
	nonce1, _ := GenerateNonce()
	nonce2, _ := GenerateNonce()

	ciphertext, _ := aead.Encrypt(nonce1[:], []byte("secret"), nil)

	_, err := aead.Decrypt(nonce2[:], ciphertext, nil)
	if err != ErrAuthFailed {
		t.Errorf("Decrypt(wrong nonce) error = %v, want ErrAuthFailed", err)
	}
}

func TestAEAD_InvalidNonce(t *testing.T) {
	key, _ := GenerateKey()
	aead, _ := NewAEAD(key[:])

	badNonces := [][]byte{nil, {}, make([]byte, 11), make([]byte, 13)}
	for _, nonce := range badNonces {
		_, err := aead.Encrypt(nonce, []byte("test"), nil)
		if err != ErrInvalidNonceSize {
			t.Errorf("Encrypt(nonce[%d]) error = %v, want ErrInvalidNonceSize", len(nonce), err)
		}
		_, err = aead.Decrypt(nonce, []byte("test"), nil)
		if err != ErrInvalidNonceSize {
			t.Errorf("Decrypt(nonce[%d]) error = %v, want ErrInvalidNonceSize", len(nonce), err)
		}
	}
}

func TestAEAD_EncryptTo(t *testing.T) {
	key, _ := GenerateKey()
	aead, _ := NewAEAD(key[:])
	nonce, _ := GenerateNonce()

	plaintext := []byte("in-place encryption test")
	dst := make([]byte, 0, len(plaintext)+TagSize)

	result, err := aead.EncryptTo(dst, nonce[:], plaintext, nil)
	if err != nil {
		t.Fatalf("EncryptTo: %v", err)
	}

	if len(result) != len(plaintext)+TagSize {
		t.Errorf("result length = %d, want %d", len(result), len(plaintext)+TagSize)
	}

	decrypted, err := aead.Decrypt(nonce[:], result, nil)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
	}
}

func TestAEAD_EncryptTo_InvalidNonce(t *testing.T) {
	key, _ := GenerateKey()
	aead, _ := NewAEAD(key[:])

	_, err := aead.EncryptTo(nil, []byte("short"), []byte("test"), nil)
	if err != ErrInvalidNonceSize {
		t.Errorf("EncryptTo(short nonce) error = %v, want ErrInvalidNonceSize", err)
	}
}

func TestAEAD_Overhead(t *testing.T) {
	key, _ := GenerateKey()
	aead, _ := NewAEAD(key[:])

	if aead.Overhead() != TagSize {
		t.Errorf("Overhead = %d, want %d", aead.Overhead(), TagSize)
	}
}

func BenchmarkAEAD_Encrypt(b *testing.B) {
	key, _ := GenerateKey()
	aead, _ := NewAEAD(key[:])
	nonce, _ := GenerateNonce()
	plaintext := make([]byte, 1024)

	b.ResetTimer()
	for b.Loop() {
		_, _ = aead.Encrypt(nonce[:], plaintext, nil)
	}
}

func BenchmarkAEAD_Decrypt(b *testing.B) {
	key, _ := GenerateKey()
	aead, _ := NewAEAD(key[:])
	nonce, _ := GenerateNonce()
	ciphertext, _ := aead.Encrypt(nonce[:], make([]byte, 1024), nil)

	b.ResetTimer()
	for b.Loop() {
		_, _ = aead.Decrypt(nonce[:], ciphertext, nil)
	}
}
