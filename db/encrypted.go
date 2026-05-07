package db

import (
	"database/sql/driver"
	"encoding/base64"
	"fmt"

	"arcan-it.de/arclib/crypto"
)

// Encrypted wraps a string value that is transparently encrypted when stored
// and decrypted when loaded. The encryption uses ChaCha20-Poly1305 via arclib/crypto.
//
// Usage:
//
//	type User struct {
//	    Name  string
//	    Email Encrypted
//	}
//
// The encryption key and nonce must be provided via SetEncryptionKey before
// any database operations. In production, the key is derived per-service
// via HKDF from the service's keyfile.
type Encrypted struct {
	Plaintext string
}

// encryptionKey is the package-level encryption key for Encrypted fields.
// Set via SetEncryptionKey. In production this is derived per-service.
var encryptionKey []byte

// SetEncryptionKey sets the key used for Encrypted field operations.
// Must be exactly 32 bytes (crypto.KeySize).
func SetEncryptionKey(key []byte) error {
	if len(key) != crypto.KeySize {
		return crypto.ErrInvalidKeySize
	}
	encryptionKey = make([]byte, crypto.KeySize)
	copy(encryptionKey, key)
	return nil
}

// NewEncrypted creates an Encrypted value from plaintext.
func NewEncrypted(plaintext string) Encrypted {
	return Encrypted{Plaintext: plaintext}
}

// Value implements driver.Valuer for database writes.
// Encrypts the plaintext and returns base64-encoded ciphertext.
func (e Encrypted) Value() (driver.Value, error) {
	if e.Plaintext == "" {
		return "", nil
	}
	if len(encryptionKey) != crypto.KeySize {
		return nil, fmt.Errorf("db: encryption key not set (call SetEncryptionKey)")
	}

	aead, err := crypto.NewAEAD(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("db: creating AEAD: %w", err)
	}

	nonce, err := crypto.GenerateNonce()
	if err != nil {
		return nil, fmt.Errorf("db: generating nonce: %w", err)
	}

	ciphertext, err := aead.Encrypt(nonce[:], []byte(e.Plaintext), nil)
	if err != nil {
		return nil, fmt.Errorf("db: encrypting: %w", err)
	}

	// Store as: base64(nonce || ciphertext)
	combined := make([]byte, crypto.NonceSize+len(ciphertext))
	copy(combined[:crypto.NonceSize], nonce[:])
	copy(combined[crypto.NonceSize:], ciphertext)

	return base64.StdEncoding.EncodeToString(combined), nil
}

// Scan implements sql.Scanner for database reads.
// Decrypts base64-encoded ciphertext back to plaintext.
func (e *Encrypted) Scan(src interface{}) error {
	if src == nil {
		e.Plaintext = ""
		return nil
	}

	var encoded string
	switch v := src.(type) {
	case string:
		encoded = v
	case []byte:
		encoded = string(v)
	default:
		return fmt.Errorf("db: Encrypted.Scan: unsupported type %T", src)
	}

	if encoded == "" {
		e.Plaintext = ""
		return nil
	}

	if len(encryptionKey) != crypto.KeySize {
		return fmt.Errorf("db: encryption key not set (call SetEncryptionKey)")
	}

	combined, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return fmt.Errorf("db: decoding base64: %w", err)
	}

	if len(combined) < crypto.NonceSize {
		return fmt.Errorf("db: ciphertext too short")
	}

	nonce := combined[:crypto.NonceSize]
	ciphertext := combined[crypto.NonceSize:]

	aead, err := crypto.NewAEAD(encryptionKey)
	if err != nil {
		return fmt.Errorf("db: creating AEAD: %w", err)
	}

	plaintext, err := aead.Decrypt(nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("db: decrypting: %w", err)
	}

	e.Plaintext = string(plaintext)
	return nil
}

// String returns the plaintext value.
func (e Encrypted) String() string {
	return e.Plaintext
}
