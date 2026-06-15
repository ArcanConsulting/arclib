package crypto

import (
	"encoding/binary"
	"fmt"
	"hash"
)

// ANSIX963KDF derives keyLen bytes from the shared secret z and the optional
// sharedInfo using the ANSI-X9.63 key derivation function (SEC1 v2 §3.6.1) — the
// KDF RFC 5753 §7.2 mandates for ECDH key agreement in CMS. It concatenates
// Hash(z ‖ counter ‖ sharedInfo) for an incrementing 32-bit big-endian counter
// starting at 1, truncating the concatenation to keyLen bytes. h selects the
// hash: SHA-256 for dhSinglePass-stdDH-sha256kdf-scheme, SHA-1 for the legacy
// sha1kdf variant OpenSSL emits by default. keyLen must be small enough that the
// counter does not overflow (2^32-1 blocks), which any real key-wrap size is.
func ANSIX963KDF(h func() hash.Hash, z, sharedInfo []byte, keyLen int) ([]byte, error) {
	if keyLen <= 0 {
		return nil, fmt.Errorf("crypto: ANSI-X9.63 KDF: key length must be positive, got %d", keyLen)
	}
	hashSize := h().Size()
	// The X9.63 counter is a 32-bit value; cap the output well below the point
	// where the block count could overflow it (any real KEK is a few blocks).
	if keyLen > hashSize<<20 {
		return nil, fmt.Errorf("crypto: ANSI-X9.63 KDF: requested key length too large")
	}
	reps := (keyLen + hashSize - 1) / hashSize
	hasher := h()
	out := make([]byte, 0, reps*hashSize)
	var counter [4]byte
	for i := 1; i <= reps; i++ {
		binary.BigEndian.PutUint32(counter[:], uint32(i)) //nolint:gosec // i <= reps <= 2^32-1, checked above
		hasher.Reset()
		hasher.Write(z)
		hasher.Write(counter[:])
		hasher.Write(sharedInfo)
		out = hasher.Sum(out)
	}
	return out[:keyLen], nil
}
