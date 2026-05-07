package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// Protocol version.
const (
	VersionMajor = 1
	VersionMinor = 0
)

// Maximum sizes.
const (
	MaxPayloadSize  = 1 << 20 // 1 MB
	MaxMessageSize  = MaxPayloadSize + 256
	MaxFragmentSize = 1 << 16 // 64 KB per fragment
)

// Flag bits in the flags byte.
const (
	FlagVersionMask   = 0x03 // Bits 0-1: Version (0-3)
	FlagTierMask      = 0x1C // Bits 2-4: Tier (0-7, only 0-5 valid)
	FlagTierShift     = 2
	FlagCompressed    = 0x20 // Bit 5: Payload is compressed
	FlagFragmented    = 0x40 // Bit 6: Message is fragmented
	FlagHasExtensions = 0x80 // Bit 7: Extension headers follow
)

// Errors.
var (
	ErrInvalidVersion   = errors.New("invalid protocol version")
	ErrInvalidTier      = errors.New("invalid security tier")
	ErrPayloadTooLarge  = errors.New("payload exceeds maximum size")
	ErrMessageTooShort  = errors.New("message too short")
	ErrInvalidChecksum  = errors.New("checksum verification failed")
	ErrInvalidHMAC      = errors.New("HMAC verification failed")
	ErrDecryptionFailed = errors.New("decryption failed")
	ErrInvalidFragment  = errors.New("invalid fragment")
)

// Header represents the protocol message header.
// The header structure varies based on the security tier.
type Header struct {
	// Version is the protocol version (0-3).
	Version uint8

	// Tier is the security tier (0-5).
	Tier Tier

	// Compressed indicates payload compression.
	Compressed bool

	// Fragmented indicates this is part of a larger message.
	Fragmented bool

	// HasExtensions indicates extension headers follow.
	HasExtensions bool

	// OpCode is the operation code.
	OpCode OpCode

	// Sequence is the message sequence number (Tier 2+).
	Sequence uint32

	// SessionID identifies the session (Tier 3+).
	SessionID uint64

	// Timestamp is the message timestamp in milliseconds (Tier 5).
	Timestamp uint64

	// Nonce is the encryption nonce (Tier 4+).
	Nonce [12]byte

	// KeyID identifies the encryption key (Tier 4+).
	KeyID uint32

	// ECDHPublic is the ephemeral ECDH public key (Tier 4+).
	// 32 bytes for X25519.
	ECDHPublic [32]byte

	// FragmentInfo contains fragmentation metadata (if Fragmented).
	FragmentInfo *FragmentInfo
}

// FragmentInfo contains fragmentation metadata.
type FragmentInfo struct {
	// MessageID identifies the original message.
	MessageID uint32

	// FragmentIndex is the fragment number (0-based).
	FragmentIndex uint16

	// TotalFragments is the total number of fragments.
	TotalFragments uint16
}

// Message represents a complete protocol message.
type Message struct {
	// Header contains the message header.
	Header Header

	// Payload is the message payload (may be encrypted/compressed).
	Payload []byte

	// Trailer contains integrity data (CRC, HMAC, or auth tag).
	Trailer []byte
}

// NewMessage creates a new message with the given parameters.
func NewMessage(tier Tier, op OpCode, payload []byte) *Message {
	return &Message{
		Header: Header{
			Version: VersionMajor - 1, // Version field is 0-indexed
			Tier:    tier,
			OpCode:  op,
		},
		Payload: payload,
	}
}

// Flags returns the packed flags byte.
func (h *Header) Flags() byte {
	flags := h.Version & FlagVersionMask
	flags |= (byte(h.Tier) << FlagTierShift) & FlagTierMask
	if h.Compressed {
		flags |= FlagCompressed
	}
	if h.Fragmented {
		flags |= FlagFragmented
	}
	if h.HasExtensions {
		flags |= FlagHasExtensions
	}
	return flags
}

// ParseFlags unpacks the flags byte into header fields.
func (h *Header) ParseFlags(flags byte) error {
	h.Version = flags & FlagVersionMask
	h.Tier = Tier((flags & FlagTierMask) >> FlagTierShift)
	h.Compressed = flags&FlagCompressed != 0
	h.Fragmented = flags&FlagFragmented != 0
	h.HasExtensions = flags&FlagHasExtensions != 0

	if !h.Tier.IsValid() {
		return ErrInvalidTier
	}
	return nil
}

// HeaderSize returns the header size in bytes for this tier.
func (h *Header) HeaderSize() int {
	size := 3 // flags (1) + opcode (2)

	if h.Fragmented {
		size += 8 // MessageID (4) + FragmentIndex (2) + TotalFragments (2)
	}

	switch h.Tier {
	case TierPlaintext, TierChecksum:
		// No additional header fields
	case TierAuthenticated:
		size += 4 // Sequence
	case TierEncrypted:
		size += 12 // Sequence (4) + SessionID (8)
	case TierPFS:
		size += 60 // Sequence (4) + SessionID (8) + Nonce (12) + KeyID (4) + ECDH (32)
	case TierMaxSecurity:
		size += 68 // Above + Timestamp (8)
	}

	return size
}

// TrailerSize returns the trailer size in bytes for this tier.
func (h *Header) TrailerSize() int {
	switch h.Tier {
	case TierPlaintext:
		return 0
	case TierChecksum:
		return 2 // CRC-16
	case TierAuthenticated:
		return 8 // Truncated HMAC-SHA256
	case TierEncrypted, TierPFS:
		return 16 // Poly1305 auth tag
	case TierMaxSecurity:
		return 32 // Full auth tag + additional MAC
	default:
		return 0
	}
}

// MarshalHeader writes the header to a byte slice.
func (h *Header) MarshalHeader() []byte {
	buf := make([]byte, h.HeaderSize())
	h.marshalHeaderTo(buf)
	return buf
}

// marshalHeaderTo writes the header to an existing buffer.
func (h *Header) marshalHeaderTo(buf []byte) {
	offset := 0

	// Flags byte
	buf[offset] = h.Flags()
	offset++

	// OpCode (big-endian)
	binary.BigEndian.PutUint16(buf[offset:], uint16(h.OpCode))
	offset += 2

	// Fragment info (if fragmented)
	if h.Fragmented && h.FragmentInfo != nil {
		binary.BigEndian.PutUint32(buf[offset:], h.FragmentInfo.MessageID)
		offset += 4
		binary.BigEndian.PutUint16(buf[offset:], h.FragmentInfo.FragmentIndex)
		offset += 2
		binary.BigEndian.PutUint16(buf[offset:], h.FragmentInfo.TotalFragments)
		offset += 2
	}

	// Tier-specific fields
	switch h.Tier {
	case TierAuthenticated:
		binary.BigEndian.PutUint32(buf[offset:], h.Sequence)

	case TierEncrypted:
		binary.BigEndian.PutUint32(buf[offset:], h.Sequence)
		offset += 4
		binary.BigEndian.PutUint64(buf[offset:], h.SessionID)

	case TierPFS:
		binary.BigEndian.PutUint32(buf[offset:], h.Sequence)
		offset += 4
		binary.BigEndian.PutUint64(buf[offset:], h.SessionID)
		offset += 8
		copy(buf[offset:], h.Nonce[:])
		offset += 12
		binary.BigEndian.PutUint32(buf[offset:], h.KeyID)
		offset += 4
		copy(buf[offset:], h.ECDHPublic[:])

	case TierMaxSecurity:
		binary.BigEndian.PutUint32(buf[offset:], h.Sequence)
		offset += 4
		binary.BigEndian.PutUint64(buf[offset:], h.SessionID)
		offset += 8
		binary.BigEndian.PutUint64(buf[offset:], h.Timestamp)
		offset += 8
		copy(buf[offset:], h.Nonce[:])
		offset += 12
		binary.BigEndian.PutUint32(buf[offset:], h.KeyID)
		offset += 4
		copy(buf[offset:], h.ECDHPublic[:])
	}
}

// UnmarshalHeader reads the header from a byte slice.
// Returns the number of bytes consumed and any error.
func (h *Header) UnmarshalHeader(data []byte) (int, error) {
	if len(data) < 3 {
		return 0, ErrMessageTooShort
	}

	offset := 0

	// Parse flags
	if err := h.ParseFlags(data[offset]); err != nil {
		return 0, err
	}
	offset++

	// OpCode
	h.OpCode = OpCode(binary.BigEndian.Uint16(data[offset:]))
	offset += 2

	// Check we have enough data for the full header
	expectedSize := h.HeaderSize()
	if len(data) < expectedSize {
		return 0, ErrMessageTooShort
	}

	// Fragment info (if fragmented)
	if h.Fragmented {
		h.FragmentInfo = &FragmentInfo{
			MessageID:      binary.BigEndian.Uint32(data[offset:]),
			FragmentIndex:  binary.BigEndian.Uint16(data[offset+4:]),
			TotalFragments: binary.BigEndian.Uint16(data[offset+6:]),
		}
		offset += 8
	}

	// Tier-specific fields
	switch h.Tier {
	case TierAuthenticated:
		h.Sequence = binary.BigEndian.Uint32(data[offset:])
		offset += 4

	case TierEncrypted:
		h.Sequence = binary.BigEndian.Uint32(data[offset:])
		offset += 4
		h.SessionID = binary.BigEndian.Uint64(data[offset:])
		offset += 8

	case TierPFS:
		h.Sequence = binary.BigEndian.Uint32(data[offset:])
		offset += 4
		h.SessionID = binary.BigEndian.Uint64(data[offset:])
		offset += 8
		copy(h.Nonce[:], data[offset:offset+12])
		offset += 12
		h.KeyID = binary.BigEndian.Uint32(data[offset:])
		offset += 4
		copy(h.ECDHPublic[:], data[offset:offset+32])
		offset += 32

	case TierMaxSecurity:
		h.Sequence = binary.BigEndian.Uint32(data[offset:])
		offset += 4
		h.SessionID = binary.BigEndian.Uint64(data[offset:])
		offset += 8
		h.Timestamp = binary.BigEndian.Uint64(data[offset:])
		offset += 8
		copy(h.Nonce[:], data[offset:offset+12])
		offset += 12
		h.KeyID = binary.BigEndian.Uint32(data[offset:])
		offset += 4
		copy(h.ECDHPublic[:], data[offset:offset+32])
		offset += 32
	}

	return offset, nil
}

// Marshal serializes the message to bytes.
func (m *Message) Marshal() ([]byte, error) {
	if len(m.Payload) > MaxPayloadSize {
		return nil, ErrPayloadTooLarge
	}

	headerSize := m.Header.HeaderSize()
	trailerSize := m.Header.TrailerSize()
	totalSize := headerSize + len(m.Payload) + trailerSize

	buf := make([]byte, totalSize)

	// Write header
	m.Header.marshalHeaderTo(buf[:headerSize])

	// Write payload
	copy(buf[headerSize:], m.Payload)

	// Trailer is written separately (after encryption/checksumming)
	if len(m.Trailer) > 0 {
		copy(buf[headerSize+len(m.Payload):], m.Trailer)
	}

	return buf, nil
}

// Unmarshal deserializes a message from bytes.
func (m *Message) Unmarshal(data []byte) error {
	// Parse header
	headerLen, err := m.Header.UnmarshalHeader(data)
	if err != nil {
		return err
	}

	trailerSize := m.Header.TrailerSize()
	if len(data) < headerLen+trailerSize {
		return ErrMessageTooShort
	}

	// Extract payload (between header and trailer)
	payloadEnd := len(data) - trailerSize
	m.Payload = make([]byte, payloadEnd-headerLen)
	copy(m.Payload, data[headerLen:payloadEnd])

	// Extract trailer
	if trailerSize > 0 {
		m.Trailer = make([]byte, trailerSize)
		copy(m.Trailer, data[payloadEnd:])
	}

	return nil
}

// WriteTo writes the message to a writer.
func (m *Message) WriteTo(w io.Writer) (int64, error) {
	data, err := m.Marshal()
	if err != nil {
		return 0, err
	}

	n, err := w.Write(data)
	return int64(n), err
}

// ReadWithLength reads a message from a reader with known length.
func (m *Message) ReadWithLength(r io.Reader, length int) error {
	if length > MaxMessageSize {
		return ErrPayloadTooLarge
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return fmt.Errorf("reading message: %w", err)
	}

	return m.Unmarshal(data)
}

// String returns a human-readable representation of the message.
func (m *Message) String() string {
	return fmt.Sprintf("Message{Op: 0x%04X, Tier: %s, PayloadLen: %d}",
		uint16(m.Header.OpCode), m.Header.Tier, len(m.Payload))
}
