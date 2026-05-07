// Package protocol implements the core wire protocol for arclib.
//
// It provides message encoding/decoding, security tier definitions,
// operation codes, and error codes. The wire format is compatible with
// the MyClerk Protocol (draft-myclerk-protocol-00).
//
// # Wire Format
//
// Messages are encoded with a compact binary header:
//
//	[flags:1][opcode:2][tier-specific fields...][payload...][trailer...]
//
// The flags byte encodes:
//   - Bits 0-1: Protocol version (0-3)
//   - Bits 2-4: Security tier (0-5)
//   - Bit 5: Payload compressed
//   - Bit 6: Message fragmented
//   - Bit 7: Extension headers follow
//
// # Security Tiers
//
// The protocol supports six security tiers (0-5), each providing
// increasing security at the cost of header/trailer overhead:
//
//   - Tier 0 (Plaintext): No security, minimal overhead
//   - Tier 1 (Checksum): CRC-16 integrity check
//   - Tier 2 (Authenticated): HMAC-SHA256 authentication
//   - Tier 3 (Encrypted): ChaCha20-Poly1305 encryption
//   - Tier 4 (PFS): Perfect Forward Secrecy with ephemeral keys
//   - Tier 5 (MaxSecurity): Full rotation with maximum protection
//
// # OpCode Registry
//
// Operation codes follow an 8-bit category + 8-bit operation structure.
// Core opcodes are defined in this package. External packages may register
// additional opcodes using RegisterOpCodes.
package protocol
