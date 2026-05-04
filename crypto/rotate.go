package crypto

import (
	"encoding/binary"
)

const (
	// rotationSalt is the fixed salt for key rotation HKDF derivation.
	// Per MyClerk Protocol draft-myclerk-protocol-03 Section 5.3.
	rotationSalt = "rotate"
)

// RotateKey derives a new key from the current key and a rotation counter.
//
// Per the MyClerk Protocol specification (Section 5.3), key rotation uses:
//
//	new_key = HKDF-SHA256(
//	    IKM  = current_key,
//	    salt = "rotate",
//	    info = rotation_counter (4 bytes, big-endian),
//	    L    = 32
//	)
//
// Rotation is triggered by:
//   - 2^32 messages sent (nonce space exhaustion)
//   - 24 hours elapsed (time-based)
//   - Explicit KEY_ROTATE operation
//
// The rotation counter must be monotonically incremented. Using the same
// counter twice with the same key produces the same derived key, which
// would constitute nonce reuse if both are used for encryption.
func RotateKey(currentKey [KeySize]byte, counter uint32) ([KeySize]byte, error) {
	var info [4]byte
	binary.BigEndian.PutUint32(info[:], counter)

	derived, err := HKDFDeriveKey(
		currentKey[:],
		[]byte(rotationSalt),
		info[:],
		KeySize,
	)
	if err != nil {
		var zero [KeySize]byte
		return zero, err
	}

	var newKey [KeySize]byte
	copy(newKey[:], derived)
	ZeroBytes(derived)
	return newKey, nil
}

// RotateKeyN performs N sequential key rotations starting from counter.
//
// This is equivalent to calling RotateKey N times in sequence,
// each time using the previous output as the new current key.
// Useful for fast-forwarding key state after reconnection.
func RotateKeyN(currentKey [KeySize]byte, startCounter, n uint32) ([KeySize]byte, error) {
	key := currentKey
	for i := uint32(0); i < n; i++ {
		var err error
		key, err = RotateKey(key, startCounter+i)
		if err != nil {
			var zero [KeySize]byte
			return zero, err
		}
	}
	return key, nil
}
