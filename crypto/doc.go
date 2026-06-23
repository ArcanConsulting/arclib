// Package crypto provides the cryptographic primitives for the Arc ecosystem.
//
// All Arc projects use this package for encryption, key derivation, and key exchange.
// The algorithms are chosen to match the MyClerk Protocol specification
// (IETF Internet-Draft draft-myclerk-protocol).
//
// Supported algorithms:
//   - SHA-256 (hashing)
//   - CRC-16-IBM, CRC-32-IEEE (checksums)
//   - HMAC-SHA256 (message authentication)
//   - ChaCha20-Poly1305 (AEAD encryption)
//   - HKDF-SHA256 (key derivation)
//   - X25519 (classical key exchange)
//   - ML-KEM-768 (post-quantum key exchange)
//   - Ed25519 (classical signatures)
//   - ML-DSA-65 (post-quantum signatures)
//   - Hybrid Ed25519+ML-DSA-65 identity keys (handshake authentication)
package crypto
