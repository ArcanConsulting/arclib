package crypto

import (
	"crypto/rand"
	"errors"
	"fmt"

	"golang.org/x/crypto/curve25519"
)

const (
	// X25519KeySize is the size of an X25519 public or private key in bytes.
	X25519KeySize = 32

	// SharedSecretSize is the size of the X25519 shared secret in bytes.
	SharedSecretSize = 32
)

var (
	// ErrInvalidPublicKey is returned when a public key is invalid (e.g., low-order point).
	ErrInvalidPublicKey = errors.New("crypto: invalid public key")
)

// X25519KeyPair holds an X25519 private/public key pair for Elliptic Curve
// Diffie-Hellman key exchange per RFC 7748.
type X25519KeyPair struct {
	Private [X25519KeySize]byte
	Public  [X25519KeySize]byte
}

// GenerateX25519KeyPair generates a random X25519 key pair.
//
// The private key is clamped per RFC 7748 Section 5:
//   - Clear the three least significant bits of the first byte
//   - Clear the most significant bit of the last byte
//   - Set the second most significant bit of the last byte
func GenerateX25519KeyPair() (*X25519KeyPair, error) {
	kp := &X25519KeyPair{}

	if _, err := rand.Read(kp.Private[:]); err != nil {
		return nil, fmt.Errorf("crypto: generating X25519 private key: %w", err)
	}

	clampX25519Private(&kp.Private)
	deriveX25519Public(kp)

	return kp, nil
}

// X25519KeyPairFromPrivate reconstructs a key pair from an existing private key.
// The private key is clamped and the public key is derived.
func X25519KeyPairFromPrivate(private [X25519KeySize]byte) *X25519KeyPair {
	kp := &X25519KeyPair{Private: private}
	clampX25519Private(&kp.Private)
	deriveX25519Public(kp)
	return kp
}

// ECDH computes the X25519 shared secret between this key pair's private key
// and the peer's public key.
//
// Returns ErrInvalidPublicKey if the peer's public key is a low-order point
// (e.g., all zeros), which would produce an all-zero shared secret.
func (kp *X25519KeyPair) ECDH(peerPublic [X25519KeySize]byte) ([SharedSecretSize]byte, error) {
	var shared [SharedSecretSize]byte

	result, err := curve25519.X25519(kp.Private[:], peerPublic[:])
	if err != nil {
		return shared, ErrInvalidPublicKey
	}

	copy(shared[:], result)
	return shared, nil
}

func clampX25519Private(key *[X25519KeySize]byte) {
	key[0] &= 248
	key[31] &= 127
	key[31] |= 64
}

func deriveX25519Public(kp *X25519KeyPair) {
	pub, _ := curve25519.X25519(kp.Private[:], curve25519.Basepoint)
	copy(kp.Public[:], pub)
}
