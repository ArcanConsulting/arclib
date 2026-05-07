package crypto

import (
	"crypto/mlkem"
	"errors"
	"fmt"
)

const (
	// MLKEMEncapKeySize is the size of an ML-KEM-768 encapsulation key in bytes.
	MLKEMEncapKeySize = 1184

	// MLKEMCiphertextSize is the size of an ML-KEM-768 ciphertext in bytes.
	MLKEMCiphertextSize = 1088

	// MLKEMSharedKeySize is the size of the ML-KEM-768 shared secret in bytes.
	MLKEMSharedKeySize = mlkem.SharedKeySize // 32
)

var (
	// ErrMLKEMEncapsulationKey is returned when an encapsulation key is invalid.
	ErrMLKEMEncapsulationKey = errors.New("crypto: invalid ML-KEM encapsulation key")

	// ErrMLKEMDecapsulation is returned when decapsulation fails.
	ErrMLKEMDecapsulation = errors.New("crypto: ML-KEM decapsulation failed")
)

// MLKEMKeyPair holds an ML-KEM-768 key pair for post-quantum key encapsulation.
//
// ML-KEM-768 (NIST FIPS 203) provides NIST Category 3 security
// (approximately AES-192 equivalent), resistant to both classical
// and quantum attacks.
type MLKEMKeyPair struct {
	DecapKey *mlkem.DecapsulationKey768
}

// GenerateMLKEMKeyPair generates a random ML-KEM-768 key pair.
func GenerateMLKEMKeyPair() (*MLKEMKeyPair, error) {
	dk, err := mlkem.GenerateKey768()
	if err != nil {
		return nil, fmt.Errorf("crypto: generating ML-KEM-768 key: %w", err)
	}
	return &MLKEMKeyPair{DecapKey: dk}, nil
}

// EncapsulationKeyBytes returns the 1184-byte encapsulation (public) key
// suitable for transmission to a peer.
func (kp *MLKEMKeyPair) EncapsulationKeyBytes() []byte {
	return kp.DecapKey.EncapsulationKey().Bytes()
}

// Decapsulate recovers the 32-byte shared secret from an encapsulation
// ciphertext produced by the peer using our encapsulation key.
func (kp *MLKEMKeyPair) Decapsulate(ciphertext []byte) ([MLKEMSharedKeySize]byte, error) {
	var shared [MLKEMSharedKeySize]byte
	ss, err := kp.DecapKey.Decapsulate(ciphertext)
	if err != nil {
		return shared, ErrMLKEMDecapsulation
	}
	copy(shared[:], ss)
	return shared, nil
}

// MLKEMEncapsulate performs encapsulation against a peer's encapsulation key.
//
// Returns the 32-byte shared secret and the 1088-byte ciphertext.
// The ciphertext must be sent to the peer who holds the corresponding
// decapsulation key.
func MLKEMEncapsulate(encapKeyBytes []byte) (sharedSecret [MLKEMSharedKeySize]byte, ciphertext []byte, err error) {
	ek, parseErr := mlkem.NewEncapsulationKey768(encapKeyBytes)
	if parseErr != nil {
		return sharedSecret, nil, ErrMLKEMEncapsulationKey
	}
	ss, ct := ek.Encapsulate()
	copy(sharedSecret[:], ss)
	return sharedSecret, ct, nil
}
