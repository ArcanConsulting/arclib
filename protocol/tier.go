package protocol

// Tier represents a security tier level (0-5).
// Higher tiers provide more security at the cost of overhead.
type Tier uint8

// Security tiers as defined in the protocol specification.
const (
	// TierPlaintext provides no security, minimal overhead (1 byte).
	// Use only for local/trusted connections.
	TierPlaintext Tier = 0

	// TierChecksum adds CRC-16 for integrity (3 bytes overhead).
	// Detects transmission errors but no security.
	TierChecksum Tier = 1

	// TierAuthenticated adds HMAC-SHA256 (truncated) for authentication (15 bytes).
	// Prevents tampering but no encryption.
	TierAuthenticated Tier = 2

	// TierEncrypted adds ChaCha20-Poly1305 session encryption (31 bytes).
	// Full encryption with session keys.
	TierEncrypted Tier = 3

	// TierPFS adds Perfect Forward Secrecy with ephemeral keys (67 bytes).
	// Compromise of long-term keys doesn't reveal past sessions.
	TierPFS Tier = 4

	// TierMaxSecurity provides maximum security with key rotation (147 bytes).
	// Full E2E encryption with frequent key rotation.
	TierMaxSecurity Tier = 5
)

// TierInfo contains metadata about a security tier.
type TierInfo struct {
	// Name is the human-readable tier name.
	Name string

	// HeaderOverhead is the additional header bytes for this tier.
	HeaderOverhead int

	// TrailerOverhead is the trailer bytes (CRC, HMAC, or auth tag).
	TrailerOverhead int

	// TotalOverhead is HeaderOverhead + TrailerOverhead.
	TotalOverhead int

	// Encrypted indicates if payload is encrypted.
	Encrypted bool

	// Authenticated indicates if messages are authenticated.
	Authenticated bool

	// HasSequence indicates if sequence numbers are used.
	HasSequence bool

	// HasSessionID indicates if session ID is included.
	HasSessionID bool

	// HasTimestamp indicates if timestamp is included.
	HasTimestamp bool

	// HasNonce indicates if explicit nonce is included.
	HasNonce bool

	// HasKeyID indicates if key identifier is included.
	HasKeyID bool

	// HasECDH indicates if ECDH public key is included (for PFS).
	HasECDH bool
}

// tierInfoTable contains metadata for all tiers.
var tierInfoTable = [6]TierInfo{
	TierPlaintext: {
		Name:            "plaintext",
		HeaderOverhead:  0,
		TrailerOverhead: 0,
		TotalOverhead:   1, // Just flags byte
	},
	TierChecksum: {
		Name:            "checksum",
		HeaderOverhead:  0,
		TrailerOverhead: 2, // CRC-16
		TotalOverhead:   3,
	},
	TierAuthenticated: {
		Name:            "authenticated",
		HeaderOverhead:  4, // Sequence number
		TrailerOverhead: 8, // HMAC-SHA256 truncated to 64 bits
		TotalOverhead:   15,
		Authenticated:   true,
		HasSequence:     true,
	},
	TierEncrypted: {
		Name:            "encrypted",
		HeaderOverhead:  12, // Sequence (4) + Session ID (8)
		TrailerOverhead: 16, // Poly1305 auth tag
		TotalOverhead:   31,
		Encrypted:       true,
		Authenticated:   true,
		HasSequence:     true,
		HasSessionID:    true,
	},
	TierPFS: {
		Name:            "pfs",
		HeaderOverhead:  48, // Sequence (4) + Session ID (8) + Nonce (12) + Key ID (4) + ECDH (32 compressed)
		TrailerOverhead: 16, // Poly1305 auth tag
		TotalOverhead:   67,
		Encrypted:       true,
		Authenticated:   true,
		HasSequence:     true,
		HasSessionID:    true,
		HasNonce:        true,
		HasKeyID:        true,
		HasECDH:         true,
	},
	TierMaxSecurity: {
		Name:            "max_security",
		HeaderOverhead:  112, // Full header with key rotation data
		TrailerOverhead: 32,  // Full HMAC + auth tag
		TotalOverhead:   147,
		Encrypted:       true,
		Authenticated:   true,
		HasSequence:     true,
		HasSessionID:    true,
		HasTimestamp:    true,
		HasNonce:        true,
		HasKeyID:        true,
		HasECDH:         true,
	},
}

// Info returns the TierInfo for this tier.
func (t Tier) Info() TierInfo {
	if t > TierMaxSecurity {
		return TierInfo{Name: "unknown"}
	}
	return tierInfoTable[t]
}

// String returns the tier name.
func (t Tier) String() string {
	return t.Info().Name
}

// IsValid checks if the tier is a valid value.
func (t Tier) IsValid() bool {
	return t <= TierMaxSecurity
}

// RequiresEncryption returns true if this tier encrypts payloads.
func (t Tier) RequiresEncryption() bool {
	return t >= TierEncrypted
}

// RequiresAuthentication returns true if this tier authenticates messages.
func (t Tier) RequiresAuthentication() bool {
	return t >= TierAuthenticated
}
