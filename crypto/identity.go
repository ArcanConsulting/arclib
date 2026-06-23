package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"
)

// Hybrid identity-key sizes. A hybrid identity is a classical Ed25519 key
// concatenated with a post-quantum ML-DSA-65 key, so a forgery requires
// breaking BOTH primitives. Public keys and signatures are the two components
// concatenated, Ed25519 first.
const (
	// IdentitySeedSize is the size of the master seed a hybrid identity is
	// derived from. The full long-term identity can be stored as just this seed.
	IdentitySeedSize = 32

	// IdentityPublicKeySize is the size of a packed hybrid public key
	// (Ed25519 || ML-DSA-65).
	IdentityPublicKeySize = ed25519.PublicKeySize + MLDSAPublicKeySize

	// IdentitySignatureSize is the size of a hybrid signature
	// (Ed25519 || ML-DSA-65).
	IdentitySignatureSize = ed25519.SignatureSize + MLDSASignatureSize
)

// identityKDFInfo domain-separates the two component seeds derived from the
// single master seed.
var identityKDFInfo = []byte("arcmail/identity-key v1")

// ErrIdentitySeedSize is returned when a seed of the wrong length is supplied.
var ErrIdentitySeedSize = errors.New("crypto: identity seed must be IdentitySeedSize bytes")

// IdentityKey is a long-term hybrid signing identity (Ed25519 + ML-DSA-65)
// derived deterministically from a single IdentitySeedSize-byte seed. It is used
// to authenticate the endpoints of a session by signing the handshake transcript.
type IdentityKey struct {
	seed   [IdentitySeedSize]byte
	edPriv ed25519.PrivateKey
	edPub  ed25519.PublicKey
	pq     *MLDSAKeyPair
}

// GenerateIdentityKey creates a fresh random hybrid identity.
func GenerateIdentityKey() (*IdentityKey, error) {
	var seed [IdentitySeedSize]byte
	if _, err := rand.Read(seed[:]); err != nil {
		return nil, fmt.Errorf("crypto: generating identity seed: %w", err)
	}
	return IdentityKeyFromSeed(seed)
}

// IdentityKeyFromSeed deterministically reconstructs a hybrid identity from a
// previously generated seed. The Ed25519 and ML-DSA component seeds are derived
// from the master seed via HKDF so the two are independent yet reproducible.
func IdentityKeyFromSeed(seed [IdentitySeedSize]byte) (*IdentityKey, error) {
	subs, err := HKDFDeriveKeys(seed[:], nil, identityKDFInfo, ed25519.SeedSize, MLDSASeedSize)
	if err != nil {
		return nil, fmt.Errorf("crypto: deriving identity component seeds: %w", err)
	}
	defer func() {
		ZeroBytes(subs[0])
		ZeroBytes(subs[1])
	}()

	edPriv := ed25519.NewKeyFromSeed(subs[0])
	var pqSeed [MLDSASeedSize]byte
	copy(pqSeed[:], subs[1])
	pq := MLDSAKeyPairFromSeed(&pqSeed)
	ZeroBytes(pqSeed[:])

	k := &IdentityKey{
		seed:   seed,
		edPriv: edPriv,
		edPub:  edPriv.Public().(ed25519.PublicKey),
		pq:     pq,
	}
	return k, nil
}

// Seed returns the master seed, suitable for persisting the identity. Treat it
// as secret key material.
func (k *IdentityKey) Seed() [IdentitySeedSize]byte { return k.seed }

// PublicKey returns the packed hybrid public key (Ed25519 || ML-DSA-65),
// IdentityPublicKeySize bytes. This is the value a peer pins / enrolls.
func (k *IdentityKey) PublicKey() []byte {
	out := make([]byte, 0, IdentityPublicKeySize)
	out = append(out, k.edPub...)
	out = append(out, k.pq.PublicKeyBytes()...)
	return out
}

// Sign produces a hybrid signature (Ed25519 || ML-DSA-65) over msg. Both
// component signatures must later verify for the hybrid signature to be valid.
func (k *IdentityKey) Sign(msg []byte) ([]byte, error) {
	pqSig, err := k.pq.Sign(msg)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0, IdentitySignatureSize)
	out = append(out, ed25519.Sign(k.edPriv, msg)...)
	out = append(out, pqSig...)
	return out, nil
}

// IdentityVerify reports whether sig is a valid hybrid signature over msg under
// the packed hybrid public key pub. BOTH the Ed25519 and the ML-DSA-65 component
// must verify. It returns false (never panics) for malformed inputs.
func IdentityVerify(pub, msg, sig []byte) bool {
	if len(pub) != IdentityPublicKeySize || len(sig) != IdentitySignatureSize {
		return false
	}
	edPub := pub[:ed25519.PublicKeySize]
	pqPub := pub[ed25519.PublicKeySize:]
	edSig := sig[:ed25519.SignatureSize]
	pqSig := sig[ed25519.SignatureSize:]
	return ed25519.Verify(edPub, msg, edSig) && MLDSAVerify(pqPub, msg, pqSig)
}

// Zero wipes the seed and the Ed25519 private key, and releases the ML-DSA key
// for garbage collection. The ML-DSA private key cannot be wiped in place (circl
// exposes no zeroizer and keeps it in expanded form), so dropping the reference
// is the best available. The IdentityKey must not be used after Zero.
func (k *IdentityKey) Zero() {
	ZeroBytes(k.seed[:])
	if k.edPriv != nil {
		ZeroBytes(k.edPriv)
	}
	k.pq = nil
}
