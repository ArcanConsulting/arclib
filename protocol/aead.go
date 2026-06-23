package protocol

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"

	"arcan-it.de/arclib/crypto"
)

// Tier-3 AEAD frame parameters. The frame layout is wire-compatible with the
// reference MyClerk AeadFrameCodec: a fixed 15-byte cleartext header used as the
// AEAD associated data, followed by the ChaCha20-Poly1305 ciphertext and tag.
//
//	[0]      flags = 0x8C  (Tier 3 Encrypted, extensions bit set)
//	[1:3]    opcode        (big-endian uint16)
//	[3:7]    sequence      (big-endian uint32, per-direction, drives the nonce)
//	[7:15]   session id    (big-endian uint64)
//	[15:]    ChaCha20-Poly1305(key, nonce, plaintext, aad=header) => ciphertext||tag
//
// The nonce is derived as the directional seed XOR the big-endian sequence in its
// first four bytes, matching the reference codec. Keys, seeds and the session id
// are established by NewAeadSession from a shared secret and two exchanged salts.
const (
	// SessionSaltSize is the length of each endpoint's handshake salt.
	SessionSaltSize = 16

	aeadHeaderSize = 15
	aeadFlagsT3    = 0x8C
	aeadTagSize    = 16
)

// AEAD session/frame errors.
var (
	ErrSessionSecretSize = errors.New("protocol: shared secret must be 32 bytes")
	ErrSessionSaltSize   = errors.New("protocol: salt must be SessionSaltSize bytes")
	ErrFrameTooShort     = errors.New("protocol: aead frame too short")
	ErrFrameNotTier3     = errors.New("protocol: not a tier-3 aead frame")
	ErrSessionMismatch   = errors.New("protocol: session id mismatch")
	ErrReplayedSequence  = errors.New("protocol: replayed or out-of-order sequence")
	ErrSequenceExhausted = errors.New("protocol: send sequence exhausted, reconnect required")
)

// AeadSession is per-connection Tier-3 state established from a shared secret and
// two exchanged random salts. Keys and nonce seeds are directional, so the two
// endpoints never share a (key, nonce) pair; the send sequence is monotonic and
// the receive side rejects any non-increasing sequence (replay protection). A
// session is safe for concurrent Encode calls; Decode is meant to be driven by a
// single read loop but guards its replay watermark regardless.
type AeadSession struct {
	sessionID uint64
	sendAEAD  *crypto.AEAD
	recvAEAD  *crypto.AEAD
	sendSeed  [crypto.NonceSize]byte
	recvSeed  [crypto.NonceSize]byte

	sendSeq uint64 // atomic counter; the wire sequence is the low 32 bits, never wrapped

	recvMu      sync.Mutex
	lastRecvSeq int64 // -1 before the first received frame
}

// GenerateSessionSalt returns a fresh random salt for the connection handshake.
func GenerateSessionSalt() ([]byte, error) {
	salt := make([]byte, SessionSaltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("protocol: generate salt: %w", err)
	}
	return salt, nil
}

// NewAeadSession derives a Tier-3 session from a 32-byte shared secret and the
// two SessionSaltSize salts exchanged at connection start. Both endpoints pass
// the same (secret, clientSalt, serverSalt); isClient selects the direction so
// that one side's send key/seed is the other's receive key/seed.
func NewAeadSession(sharedSecret, clientSalt, serverSalt []byte, isClient bool) (*AeadSession, error) {
	if len(sharedSecret) != crypto.KeySize {
		return nil, ErrSessionSecretSize
	}
	if len(clientSalt) != SessionSaltSize || len(serverSalt) != SessionSaltSize {
		return nil, ErrSessionSaltSize
	}

	salt := make([]byte, 0, 2*SessionSaltSize)
	salt = append(salt, clientSalt...)
	salt = append(salt, serverSalt...)

	keys, err := crypto.HKDFDeriveKeys(sharedSecret, salt, []byte("arcmail/myclerk tier3 v1"),
		crypto.KeySize, crypto.KeySize, crypto.NonceSize, crypto.NonceSize)
	if err != nil {
		return nil, fmt.Errorf("protocol: derive session keys: %w", err)
	}
	c2sKey, s2cKey, c2sSeed, s2cSeed := keys[0], keys[1], keys[2], keys[3]

	idBytes, err := crypto.HKDFDeriveKeys(sharedSecret, salt, []byte("arcmail/myclerk tier3 session-id"), 8)
	if err != nil {
		return nil, fmt.Errorf("protocol: derive session id: %w", err)
	}

	sendKey, recvKey, sendSeed, recvSeed := c2sKey, s2cKey, c2sSeed, s2cSeed
	if !isClient {
		sendKey, recvKey, sendSeed, recvSeed = s2cKey, c2sKey, s2cSeed, c2sSeed
	}

	s := &AeadSession{sessionID: binary.BigEndian.Uint64(idBytes[0]), lastRecvSeq: -1}
	if s.sendAEAD, err = crypto.NewAEAD(sendKey); err != nil {
		return nil, fmt.Errorf("protocol: send aead: %w", err)
	}
	if s.recvAEAD, err = crypto.NewAEAD(recvKey); err != nil {
		return nil, fmt.Errorf("protocol: recv aead: %w", err)
	}
	copy(s.sendSeed[:], sendSeed)
	copy(s.recvSeed[:], recvSeed)
	return s, nil
}

// NewAeadSessionFromKeys builds a Tier-3 session directly from pre-derived
// directional keys, nonce seeds and a session id — for example the output of an
// authenticated key exchange (see handshake.go) rather than the static-key salt
// exchange. sendKey/recvKey must be crypto.KeySize bytes. Callers are
// responsible for deriving the material so that one endpoint's send key/seed is
// the other's receive key/seed.
func NewAeadSessionFromKeys(sessionID uint64, sendKey, recvKey []byte, sendSeed, recvSeed [crypto.NonceSize]byte) (*AeadSession, error) {
	if len(sendKey) != crypto.KeySize || len(recvKey) != crypto.KeySize {
		return nil, ErrSessionSecretSize
	}
	s := &AeadSession{sessionID: sessionID, lastRecvSeq: -1}
	var err error
	if s.sendAEAD, err = crypto.NewAEAD(sendKey); err != nil {
		return nil, fmt.Errorf("protocol: send aead: %w", err)
	}
	if s.recvAEAD, err = crypto.NewAEAD(recvKey); err != nil {
		return nil, fmt.Errorf("protocol: recv aead: %w", err)
	}
	s.sendSeed = sendSeed
	s.recvSeed = recvSeed
	return s, nil
}

// SessionID returns the session identifier both endpoints derived.
func (s *AeadSession) SessionID() uint64 { return s.sessionID }

// Encode seals plaintext for op into a Tier-3 frame. Each call consumes the next
// send sequence, so callers must serialize Encode with the actual write to the
// connection to keep wire order equal to sequence order (the peer's replay check
// requires strictly increasing sequences).
//
// The wire sequence is a uint32, but the counter is a uint64 that is never
// allowed to wrap: once 2^32 frames have been sent on a direction, Encode returns
// ErrSequenceExhausted instead of reusing sequence 0. This is a security
// requirement, not merely a counter limit — the nonce is seed XOR sequence, so a
// wrapped sequence would reuse a (key, nonce) pair and break ChaCha20-Poly1305.
// The peer must reconnect (a fresh handshake derives fresh keys and seeds).
func (s *AeadSession) Encode(op OpCode, plaintext []byte) ([]byte, error) {
	n := atomic.AddUint64(&s.sendSeq, 1) - 1
	if n > math.MaxUint32 {
		return nil, ErrSequenceExhausted
	}
	seq := uint32(n)

	header := make([]byte, aeadHeaderSize)
	header[0] = aeadFlagsT3
	binary.BigEndian.PutUint16(header[1:], uint16(op))
	binary.BigEndian.PutUint32(header[3:], seq)
	binary.BigEndian.PutUint64(header[7:], s.sessionID)

	nonce := deriveNonce(s.sendSeed, seq)
	ct, err := s.sendAEAD.Encrypt(nonce, plaintext, header)
	if err != nil {
		return nil, fmt.Errorf("protocol: aead encrypt: %w", err)
	}

	frame := make([]byte, aeadHeaderSize+len(ct))
	copy(frame, header)
	copy(frame[aeadHeaderSize:], ct)
	return frame, nil
}

// Decode opens a Tier-3 frame: it authenticates the header (as AAD) and payload,
// then — only once authenticity is proven — enforces the monotonic replay check,
// so a forged or tampered frame can neither be accepted nor poison the watermark.
func (s *AeadSession) Decode(frame []byte) (OpCode, []byte, error) {
	if len(frame) < aeadHeaderSize+aeadTagSize {
		return 0, nil, ErrFrameTooShort
	}
	if frame[0] != aeadFlagsT3 {
		return 0, nil, ErrFrameNotTier3
	}
	op := OpCode(binary.BigEndian.Uint16(frame[1:]))
	seq := binary.BigEndian.Uint32(frame[3:])
	sid := binary.BigEndian.Uint64(frame[7:])
	if sid != s.sessionID {
		return 0, nil, ErrSessionMismatch
	}

	nonce := deriveNonce(s.recvSeed, seq)
	pt, err := s.recvAEAD.Decrypt(nonce, frame[aeadHeaderSize:], frame[:aeadHeaderSize])
	if err != nil {
		return 0, nil, fmt.Errorf("protocol: aead decrypt: %w", err)
	}
	if err := s.checkRecvSeq(seq); err != nil {
		return 0, nil, err
	}
	return op, pt, nil
}

// checkRecvSeq enforces strictly increasing receive sequences. No uint32
// wraparound logic is needed because the peer's Encode refuses to wrap (it errors
// at ErrSequenceExhausted), so a received sequence is always in [0, 2^32-1] and
// int64 comparison is exact and monotonic.
func (s *AeadSession) checkRecvSeq(seq uint32) error {
	s.recvMu.Lock()
	defer s.recvMu.Unlock()
	if s.lastRecvSeq >= 0 && int64(seq) <= s.lastRecvSeq {
		return ErrReplayedSequence
	}
	s.lastRecvSeq = int64(seq)
	return nil
}

// deriveNonce XORs the big-endian sequence into the first four bytes of the
// directional seed (reference AeadFrameCodec scheme). Distinct sequences yield
// distinct nonces within a direction; distinct directional seeds and keys keep
// the two directions disjoint.
func deriveNonce(seed [crypto.NonceSize]byte, seq uint32) []byte {
	n := make([]byte, crypto.NonceSize)
	copy(n, seed[:])
	n[0] ^= byte((seq >> 24) & 0xFF)
	n[1] ^= byte((seq >> 16) & 0xFF)
	n[2] ^= byte((seq >> 8) & 0xFF)
	n[3] ^= byte(seq & 0xFF)
	return n
}
