// Package crypto provides the cryptographic primitives for the Arc ecosystem.
//
// All Arc projects use this package for encryption, key derivation, and key exchange.
// The algorithms are chosen to match the MyClerk Protocol specification (IETF Internet-Draft).
//
// Supported algorithms:
//   - ChaCha20-Poly1305 (AEAD encryption)
//   - HKDF-SHA256 (key derivation)
//   - X25519 (classical key exchange)
//   - ML-KEM-768 (post-quantum key exchange)
package crypto
