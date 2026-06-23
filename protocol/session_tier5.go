package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	"arcan-it.de/arclib/crypto"
)

// Tier-5 (MaxSecurity) session codec: Tier-3 AEAD plus automatic key rotation
// (draft-myclerk-protocol-03 §3.7, §5.3). Each direction advances through key
// epochs; the epoch is carried in the frame as a KeyID so the receiver can
// follow. Rotation gives forward secrecy WITHIN a long-lived session: because
// each epoch key is a one-way HKDF of the previous (crypto.RotateKey), capturing
// the current key reveals neither past nor (without the next derivation) future
// epochs.
//
// A direction rotates when either the per-epoch message budget or the time
// budget is reached (§5.3 lists 2^32 messages / 24h / explicit). The nonce is
// the directional seed XOR the per-epoch sequence, which is safe across epochs
// because the key differs every epoch.
//
//	frame = [flags:1][opcode:2][seq:4][session id:8][key epoch:4] || AEAD(ct||tag)
//	        \__________________ 19-byte header, also the AEAD AAD _____________/

const (
	tier5HeaderSize = 19
	// tier5Flags marks a Tier-5 frame: version 0, tier 5 (5<<FlagTierShift).
	tier5Flags = byte(TierMaxSecurity) << FlagTierShift // 0x14

	defaultRekeyMsgs     = 1 << 20 // rotate after ~1M messages in an epoch
	defaultRekeyInterval = 24 * time.Hour
	maxEpochJump         = 16 // bound receiver fast-forward work (anti-DoS)
)

// Tier-5 specific errors.
var (
	ErrFrameNotTier5 = errors.New("protocol: not a tier-5 frame")
	ErrReplayedEpoch = errors.New("protocol: replayed or out-of-order key epoch")
	ErrEpochJump     = errors.New("protocol: key epoch advanced too far")
)

// Tier5Session is a Tier-5 frame codec with directional key rotation. Encode is
// safe for concurrent use; Decode is intended for a single read loop but guards
// its own state. Build one from a completed handshake via DerivedSession.
type Tier5Session struct {
	sessionID  uint64
	rekeyMsgs  uint32
	rekeyEvery time.Duration

	sendMu     sync.Mutex
	sendSeed   [crypto.NonceSize]byte
	sendKey    [crypto.KeySize]byte
	sendAEAD   *crypto.AEAD
	sendEpoch  uint32
	sendSeq    uint32
	sendEpochT time.Time

	recvMu      sync.Mutex
	recvSeed    [crypto.NonceSize]byte
	recvKey     [crypto.KeySize]byte
	recvAEAD    *crypto.AEAD
	recvEpoch   uint32
	lastRecvSeq int64
}

// Tier5Session builds a Tier-5 session from the derived handshake keys using the
// default rotation policy.
func (d *DerivedSession) Tier5Session() (*Tier5Session, error) {
	return NewTier5Session(d, defaultRekeyMsgs, defaultRekeyInterval)
}

// NewTier5Session builds a Tier-5 session with an explicit rotation policy.
// rekeyMsgs must be in (0, math.MaxUint32) so the per-epoch sequence never wraps.
func NewTier5Session(d *DerivedSession, rekeyMsgs uint32, rekeyEvery time.Duration) (*Tier5Session, error) {
	if rekeyMsgs == 0 {
		rekeyMsgs = defaultRekeyMsgs
	}
	if rekeyEvery <= 0 {
		rekeyEvery = defaultRekeyInterval
	}
	s := &Tier5Session{
		sessionID:   d.SessionID,
		rekeyMsgs:   rekeyMsgs,
		rekeyEvery:  rekeyEvery,
		sendSeed:    d.SendSeed,
		sendKey:     d.SendKey,
		recvSeed:    d.RecvSeed,
		recvKey:     d.RecvKey,
		lastRecvSeq: -1,
		sendEpochT:  time.Now(),
	}
	var err error
	if s.sendAEAD, err = crypto.NewAEAD(s.sendKey[:]); err != nil {
		return nil, fmt.Errorf("protocol: tier5 send aead: %w", err)
	}
	if s.recvAEAD, err = crypto.NewAEAD(s.recvKey[:]); err != nil {
		return nil, fmt.Errorf("protocol: tier5 recv aead: %w", err)
	}
	return s, nil
}

// SessionID returns the session identifier both endpoints derived.
func (s *Tier5Session) SessionID() uint64 { return s.sessionID }

// Encode seals plaintext for op into a Tier-5 frame, rotating the send key first
// if the message or time budget for the current epoch is exhausted.
func (s *Tier5Session) Encode(op OpCode, plaintext []byte) ([]byte, error) {
	s.sendMu.Lock()
	defer s.sendMu.Unlock()

	if s.sendSeq >= s.rekeyMsgs || time.Since(s.sendEpochT) >= s.rekeyEvery {
		if err := s.rotateSend(); err != nil {
			return nil, err
		}
	}
	seq := s.sendSeq
	s.sendSeq++

	header := make([]byte, tier5HeaderSize)
	header[0] = tier5Flags
	binary.BigEndian.PutUint16(header[1:], uint16(op))
	binary.BigEndian.PutUint32(header[3:], seq)
	binary.BigEndian.PutUint64(header[7:], s.sessionID)
	binary.BigEndian.PutUint32(header[15:], s.sendEpoch)

	nonce := deriveNonce(s.sendSeed, seq)
	ct, err := s.sendAEAD.Encrypt(nonce, plaintext, header)
	if err != nil {
		return nil, fmt.Errorf("protocol: tier5 encrypt: %w", err)
	}
	frame := make([]byte, tier5HeaderSize+len(ct))
	copy(frame, header)
	copy(frame[tier5HeaderSize:], ct)
	return frame, nil
}

// rotateSend advances the send direction to the next key epoch.
func (s *Tier5Session) rotateSend() error {
	newKey, err := crypto.RotateKey(s.sendKey, s.sendEpoch)
	if err != nil {
		return fmt.Errorf("protocol: tier5 rotate send: %w", err)
	}
	aead, err := crypto.NewAEAD(newKey[:])
	if err != nil {
		return fmt.Errorf("protocol: tier5 send aead: %w", err)
	}
	crypto.ZeroBytes(s.sendKey[:])
	s.sendKey = newKey
	s.sendAEAD = aead
	s.sendEpoch++
	s.sendSeq = 0
	s.sendEpochT = time.Now()
	return nil
}

// Decode opens a Tier-5 frame. It follows the sender's key epoch (carried as the
// KeyID), but only COMMITS an epoch advance after the frame authenticates, so a
// forged frame can neither be accepted nor poison the receive key state. Replay
// is rejected per epoch by a monotonic sequence and across epochs by a monotonic
// epoch.
func (s *Tier5Session) Decode(frame []byte) (OpCode, []byte, error) {
	if len(frame) < tier5HeaderSize+aeadTagSize {
		return 0, nil, ErrFrameTooShort
	}
	if frame[0] != tier5Flags {
		return 0, nil, ErrFrameNotTier5
	}
	op := OpCode(binary.BigEndian.Uint16(frame[1:]))
	seq := binary.BigEndian.Uint32(frame[3:])
	sid := binary.BigEndian.Uint64(frame[7:])
	epoch := binary.BigEndian.Uint32(frame[15:])
	if sid != s.sessionID {
		return 0, nil, ErrSessionMismatch
	}

	s.recvMu.Lock()
	defer s.recvMu.Unlock()

	if epoch < s.recvEpoch {
		return 0, nil, ErrReplayedEpoch
	}

	// Derive the candidate key for this frame's epoch WITHOUT mutating state.
	candKey := s.recvKey
	candAEAD := s.recvAEAD
	if epoch > s.recvEpoch {
		jump := epoch - s.recvEpoch
		if jump > maxEpochJump {
			return 0, nil, ErrEpochJump
		}
		rotated, err := crypto.RotateKeyN(s.recvKey, s.recvEpoch, jump)
		if err != nil {
			return 0, nil, fmt.Errorf("protocol: tier5 rotate recv: %w", err)
		}
		candKey = rotated
		if candAEAD, err = crypto.NewAEAD(candKey[:]); err != nil {
			return 0, nil, fmt.Errorf("protocol: tier5 recv aead: %w", err)
		}
	}

	nonce := deriveNonce(s.recvSeed, seq)
	pt, err := candAEAD.Decrypt(nonce, frame[tier5HeaderSize:], frame[:tier5HeaderSize])
	if err != nil {
		return 0, nil, fmt.Errorf("protocol: tier5 decrypt: %w", err)
	}

	// Authenticated: now commit any epoch advance and enforce the replay window.
	if epoch > s.recvEpoch {
		crypto.ZeroBytes(s.recvKey[:])
		s.recvKey = candKey
		s.recvAEAD = candAEAD
		s.recvEpoch = epoch
		s.lastRecvSeq = -1
	}
	if s.lastRecvSeq >= 0 && int64(seq) <= s.lastRecvSeq {
		return 0, nil, ErrReplayedSequence
	}
	s.lastRecvSeq = int64(seq)
	return op, pt, nil
}

// Zero wipes the in-memory key material.
func (s *Tier5Session) Zero() {
	s.sendMu.Lock()
	crypto.ZeroBytes(s.sendKey[:])
	s.sendMu.Unlock()
	s.recvMu.Lock()
	crypto.ZeroBytes(s.recvKey[:])
	s.recvMu.Unlock()
}
