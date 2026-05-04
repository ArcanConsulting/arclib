package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
)

const (
	// HMACFullSize is the size of a full HMAC-SHA256 output in bytes.
	HMACFullSize = sha256.Size // 32 bytes
)

var (
	// ErrInvalidKeySize is returned when a key has an incorrect length.
	ErrInvalidKeySize = errors.New("crypto: invalid key size")

	// ErrInvalidTruncation is returned when a truncation length exceeds the full HMAC size.
	ErrInvalidTruncation = errors.New("crypto: truncation exceeds HMAC size")
)

// HMACSHA256 computes the full HMAC-SHA256 of data using the given key.
// Returns a 32-byte MAC.
func HMACSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

// HMACSHA256Truncated computes HMAC-SHA256 and truncates the result to n bytes.
//
// The MyClerk Protocol uses n=8 for Tier 2 authentication. Truncation to fewer
// bytes reduces security but saves bandwidth — 8 bytes (64 bits) provides
// adequate integrity for the protocol's use case.
//
// Returns an error if n exceeds HMACFullSize (32 bytes) or is zero.
func HMACSHA256Truncated(key, data []byte, n int) ([]byte, error) {
	if n <= 0 || n > HMACFullSize {
		return nil, ErrInvalidTruncation
	}
	full := HMACSHA256(key, data)
	result := make([]byte, n)
	copy(result, full[:n])
	return result, nil
}

// VerifyHMACSHA256 verifies a full HMAC-SHA256 tag.
// Uses constant-time comparison to prevent timing attacks.
func VerifyHMACSHA256(key, data, expectedMAC []byte) bool {
	computed := HMACSHA256(key, data)
	return hmac.Equal(computed, expectedMAC)
}

// VerifyHMACSHA256Truncated verifies a truncated HMAC-SHA256 tag.
// The expectedMAC length determines the truncation size.
// Uses constant-time comparison to prevent timing attacks.
func VerifyHMACSHA256Truncated(key, data, expectedMAC []byte) bool {
	n := len(expectedMAC)
	if n <= 0 || n > HMACFullSize {
		return false
	}
	full := HMACSHA256(key, data)
	return hmac.Equal(full[:n], expectedMAC)
}
