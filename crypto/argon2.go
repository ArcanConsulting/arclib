package crypto

import "golang.org/x/crypto/argon2"

// Argon2idParams configures the Argon2id password-based KDF (RFC 9106).
//
// Use Argon2id to derive a key from a LOW-entropy secret — a user password or
// passphrase. For HIGH-entropy key material (shared secrets, existing keys) use
// HKDFDeriveKey instead: HKDF is fast and has no work factor, so it must never
// be used directly on a password.
type Argon2idParams struct {
	// Time is the number of passes over memory (>= 1).
	Time uint32
	// Memory is the memory cost in KiB.
	Memory uint32
	// Threads is the degree of parallelism (lanes, >= 1).
	Threads uint8
	// KeyLen is the derived key length in bytes.
	KeyLen uint32
}

// DefaultArgon2idParams returns interactive-grade parameters: 64 MiB memory,
// 3 passes, 4 lanes, and a 32-byte key (a KeySize key for the AEAD).
//
// Store the chosen parameters alongside the salt so the same key can be
// re-derived later, and so the cost can be raised over time without breaking
// existing material.
func DefaultArgon2idParams() Argon2idParams {
	return Argon2idParams{
		Time:    3,
		Memory:  64 * 1024,
		Threads: 4,
		KeyLen:  KeySize,
	}
}

// DeriveKeyArgon2id derives a key from a password and salt using Argon2id.
//
// Argon2id (RFC 9106) is memory-hard, resisting GPU/ASIC brute force — the
// property a low-entropy password needs. The salt must be unique per secret
// (>= 16 bytes recommended), generated with a CSPRNG, and stored with the
// parameters so the key can be re-derived.
func DeriveKeyArgon2id(password, salt []byte, p Argon2idParams) []byte {
	return argon2.IDKey(password, salt, p.Time, p.Memory, p.Threads, p.KeyLen)
}
