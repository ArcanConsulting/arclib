package crypto

import (
	"crypto/cipher"
	"errors"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
)

const (
	// KeySize is the size of a ChaCha20-Poly1305 key in bytes (256 bits).
	KeySize = chacha20poly1305.KeySize // 32

	// NonceSize is the size of a ChaCha20-Poly1305 nonce in bytes (96 bits).
	NonceSize = chacha20poly1305.NonceSize // 12

	// TagSize is the size of the Poly1305 authentication tag in bytes.
	TagSize = 16
)

var (
	// ErrInvalidNonceSize is returned when a nonce has an incorrect length.
	ErrInvalidNonceSize = errors.New("crypto: invalid nonce size")

	// ErrAuthFailed is returned when ciphertext authentication fails.
	// This indicates either the wrong key, tampered ciphertext, or incorrect AAD.
	ErrAuthFailed = errors.New("crypto: authentication failed")
)

// AEAD provides ChaCha20-Poly1305 authenticated encryption with associated data.
//
// This is the primary encryption primitive for the MyClerk Protocol (Tier 3+).
// Per RFC 8439, ChaCha20-Poly1305 provides both confidentiality and integrity
// in a single operation, using a 256-bit key and 96-bit nonce.
type AEAD struct {
	aead cipher.AEAD
}

// NewAEAD creates a new ChaCha20-Poly1305 AEAD cipher from a 32-byte key.
func NewAEAD(key []byte) (*AEAD, error) {
	if len(key) != KeySize {
		return nil, ErrInvalidKeySize
	}

	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: creating AEAD: %w", err)
	}

	return &AEAD{aead: aead}, nil
}

// Encrypt encrypts plaintext with the given nonce and optional additional data.
//
// Returns ciphertext (len(plaintext) + TagSize bytes) containing the encrypted
// data followed by the Poly1305 authentication tag.
//
// The nonce must be exactly NonceSize (12) bytes. Each nonce must only be used
// once with the same key — reuse compromises both confidentiality and integrity.
func (a *AEAD) Encrypt(nonce, plaintext, additionalData []byte) ([]byte, error) {
	if len(nonce) != NonceSize {
		return nil, ErrInvalidNonceSize
	}
	return a.aead.Seal(nil, nonce, plaintext, additionalData), nil
}

// Decrypt decrypts and authenticates ciphertext with the given nonce and AAD.
//
// Returns the plaintext if authentication succeeds, or ErrAuthFailed if the
// ciphertext was tampered with, the key is wrong, or the AAD doesn't match.
func (a *AEAD) Decrypt(nonce, ciphertext, additionalData []byte) ([]byte, error) {
	if len(nonce) != NonceSize {
		return nil, ErrInvalidNonceSize
	}

	plaintext, err := a.aead.Open(nil, nonce, ciphertext, additionalData)
	if err != nil {
		return nil, ErrAuthFailed
	}
	return plaintext, nil
}

// EncryptTo encrypts plaintext and appends the result to dst.
//
// This avoids an allocation when the caller already has a buffer.
// The dst slice is used as the starting point — the ciphertext is appended
// beginning at dst[:0], so any existing content in dst is overwritten.
func (a *AEAD) EncryptTo(dst, nonce, plaintext, additionalData []byte) ([]byte, error) {
	if len(nonce) != NonceSize {
		return nil, ErrInvalidNonceSize
	}
	return a.aead.Seal(dst[:0], nonce, plaintext, additionalData), nil
}

// Overhead returns the number of bytes added by encryption (the tag size).
func (a *AEAD) Overhead() int {
	return a.aead.Overhead()
}
