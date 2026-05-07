package crypto

import (
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

// HKDFDeriveKey derives a single key of the given length using HKDF-SHA256.
//
// Per RFC 5869, HKDF consists of two stages:
//   - Extract: PRK = HMAC-SHA256(salt, ikm)
//   - Expand: OKM = HKDF-Expand(PRK, info, length)
//
// Parameters:
//   - ikm: Input keying material (shared secret, concatenated secrets, etc.)
//   - salt: Optional salt (can be nil; HKDF uses a zero-filled salt internally)
//   - info: Context and application-specific info (domain separator)
//   - length: Desired output length in bytes (max 255 * 32 = 8160)
func HKDFDeriveKey(ikm, salt, info []byte, length int) ([]byte, error) {
	reader := hkdf.New(sha256.New, ikm, salt, info)
	out := make([]byte, length)
	if _, err := io.ReadFull(reader, out); err != nil {
		return nil, fmt.Errorf("crypto: HKDF expand: %w", err)
	}
	return out, nil
}

// HKDFDeriveKeys derives multiple keys of specified lengths in a single pass.
//
// Each key is derived sequentially from the same HKDF stream, which is
// equivalent to deriving one large key and splitting it. This is the pattern
// used by the MyClerk Protocol to derive send_key, receive_key,
// send_nonce_seed, and receive_nonce_seed from a single shared secret.
//
// Example:
//
//	keys, _ := HKDFDeriveKeys(secret, nil, info, 32, 32, 12, 12)
//	sendKey := keys[0]       // 32 bytes
//	recvKey := keys[1]       // 32 bytes
//	sendNonce := keys[2]     // 12 bytes
//	recvNonce := keys[3]     // 12 bytes
func HKDFDeriveKeys(ikm, salt, info []byte, lengths ...int) ([][]byte, error) {
	total := 0
	for _, l := range lengths {
		total += l
	}

	reader := hkdf.New(sha256.New, ikm, salt, info)
	buf := make([]byte, total)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return nil, fmt.Errorf("crypto: HKDF expand: %w", err)
	}

	keys := make([][]byte, len(lengths))
	offset := 0
	for i, l := range lengths {
		keys[i] = make([]byte, l)
		copy(keys[i], buf[offset:offset+l])
		offset += l
	}

	ZeroBytes(buf)
	return keys, nil
}
