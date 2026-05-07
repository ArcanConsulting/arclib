package crypto

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

// GenerateKey generates a cryptographically random 32-byte key
// suitable for use with ChaCha20-Poly1305 or HMAC-SHA256.
func GenerateKey() ([KeySize]byte, error) {
	var key [KeySize]byte
	if _, err := rand.Read(key[:]); err != nil {
		return key, fmt.Errorf("crypto: generating key: %w", err)
	}
	return key, nil
}

// GenerateNonce generates a cryptographically random 12-byte nonce
// suitable for use with ChaCha20-Poly1305.
//
// For most use cases, prefer NonceFromSequence which derives deterministic
// nonces from a sequence number and session ID, avoiding the birthday-bound
// collision risk of random nonces.
func GenerateNonce() ([NonceSize]byte, error) {
	var nonce [NonceSize]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nonce, fmt.Errorf("crypto: generating nonce: %w", err)
	}
	return nonce, nil
}

// NonceFromSequence derives a deterministic 12-byte nonce from a sequence
// number and session ID.
//
// Layout (per MyClerk Protocol draft-myclerk-protocol-03, Section 4):
//
//	Bytes 0-3:  Sequence number (big-endian uint32)
//	Bytes 4-11: Session ID (big-endian uint64)
//
// This construction guarantees unique nonces within a session as long as
// the sequence number is monotonically incremented. Different sessions
// produce different nonces due to the session ID component.
func NonceFromSequence(sequence uint32, sessionID uint64) [NonceSize]byte {
	var nonce [NonceSize]byte
	binary.BigEndian.PutUint32(nonce[0:4], sequence)
	binary.BigEndian.PutUint64(nonce[4:12], sessionID)
	return nonce
}

// ZeroBytes overwrites a byte slice with zeros.
// Use this to clear key material from memory after use.
func ZeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
