package crypto

import (
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/cloudflare/circl/sign/mldsa/mldsa65"
)

const (
	// MLDSAPublicKeySize is the size of an ML-DSA-65 public key in bytes.
	MLDSAPublicKeySize = mldsa65.PublicKeySize

	// MLDSAPrivateKeySize is the size of an ML-DSA-65 private key in bytes.
	MLDSAPrivateKeySize = mldsa65.PrivateKeySize

	// MLDSASignatureSize is the size of an ML-DSA-65 signature in bytes.
	MLDSASignatureSize = mldsa65.SignatureSize

	// MLDSASeedSize is the size of the seed an ML-DSA-65 key is derived from.
	MLDSASeedSize = mldsa65.SeedSize
)

var (
	// ErrMLDSAPublicKey is returned when an ML-DSA public key is malformed.
	ErrMLDSAPublicKey = errors.New("crypto: invalid ML-DSA public key")

	// ErrMLDSASign is returned when ML-DSA signing fails.
	ErrMLDSASign = errors.New("crypto: ML-DSA signing failed")
)

// MLDSAKeyPair holds an ML-DSA-65 (NIST FIPS 204) signature key pair.
//
// ML-DSA-65 targets NIST security category 3 (≈AES-192), the signature
// counterpart to the ML-KEM-768 key encapsulation used for key exchange. It is
// resistant to both classical and quantum forgery.
type MLDSAKeyPair struct {
	pub  *mldsa65.PublicKey
	priv *mldsa65.PrivateKey
}

// GenerateMLDSAKeyPair generates a random ML-DSA-65 key pair.
func GenerateMLDSAKeyPair() (*MLDSAKeyPair, error) {
	pub, priv, err := mldsa65.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("crypto: generating ML-DSA-65 key: %w", err)
	}
	return &MLDSAKeyPair{pub: pub, priv: priv}, nil
}

// MLDSAKeyPairFromSeed deterministically derives an ML-DSA-65 key pair from a
// MLDSASeedSize-byte seed. The same seed always yields the same key pair, so a
// long-term identity can be stored as just the seed.
func MLDSAKeyPairFromSeed(seed *[MLDSASeedSize]byte) *MLDSAKeyPair {
	pub, priv := mldsa65.NewKeyFromSeed(seed)
	return &MLDSAKeyPair{pub: pub, priv: priv}
}

// PublicKeyBytes returns the MLDSAPublicKeySize-byte packed public key.
func (kp *MLDSAKeyPair) PublicKeyBytes() []byte {
	return kp.pub.Bytes()
}

// Sign produces a hedged (randomized) ML-DSA-65 signature over msg.
func (kp *MLDSAKeyPair) Sign(msg []byte) ([]byte, error) {
	sig := make([]byte, MLDSASignatureSize)
	if err := mldsa65.SignTo(kp.priv, msg, nil, true, sig); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMLDSASign, err)
	}
	return sig, nil
}

// MLDSAVerify reports whether sig is a valid ML-DSA-65 signature over msg under
// the packed public key pubBytes. It returns false (never panics) for a
// malformed public key or signature length.
func MLDSAVerify(pubBytes, msg, sig []byte) bool {
	if len(pubBytes) != MLDSAPublicKeySize || len(sig) != MLDSASignatureSize {
		return false
	}
	var buf [MLDSAPublicKeySize]byte
	copy(buf[:], pubBytes)
	var pk mldsa65.PublicKey
	pk.Unpack(&buf)
	return mldsa65.Verify(&pk, msg, nil, sig)
}
