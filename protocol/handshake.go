package protocol

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"

	"arcan-it.de/arclib/crypto"
	"arcan-it.de/arclib/msgpack"
)

// Hybrid key-exchange handshake (MyClerk draft-myclerk-protocol-03 §5.1, §7).
//
// Two endpoints establish a Tier-3+ session by exchanging SESSION_INIT /
// SESSION_ACK in the clear: each contributes an ephemeral X25519 public key and
// (in hybrid mode) the initiator an ephemeral ML-KEM-768 encapsulation key, the
// responder the matching ciphertext. The shared secret is
// combined = ss_x25519 ‖ ss_mlkem, secure as long as EITHER primitive holds.
// Session keys are HKDF-derived with the handshake transcript folded into the
// info parameter, so any on-path modification makes the two ends derive
// different keys and the first encrypted frame fails AEAD verification (§10.6).
//
// This handshake is anonymous: it provides confidentiality, forward secrecy and
// integrity, but NOT peer authentication. Callers that need to authenticate the
// remote endpoint layer an identity exchange (see auth.go) over the channel.

// KexMode selects the key-establishment algorithm. The values are carried on
// the wire and MUST match the specification.
type KexMode uint8

const (
	// KexModeUnset is the invalid zero value.
	KexModeUnset KexMode = 0
	// KexModeClassicalOnly derives the session key from X25519 alone.
	KexModeClassicalOnly KexMode = 1
	// KexModeHybridMLKEM768 combines X25519 with ML-KEM-768 (the default).
	KexModeHybridMLKEM768 KexMode = 2
)

// String returns the mode's label, also used as the HKDF info domain separator.
func (m KexMode) String() string {
	switch m {
	case KexModeClassicalOnly:
		return "myclerk-session-v1-classical"
	case KexModeHybridMLKEM768:
		return "myclerk-session-v1-hybrid"
	default:
		return "myclerk-session-v1-unset"
	}
}

// Handshake errors.
var (
	ErrKexMode          = errors.New("protocol: unacceptable key-exchange mode")
	ErrHandshakeFraming = errors.New("protocol: malformed handshake frame")
	ErrHandshakeKey     = errors.New("protocol: invalid handshake public key material")
)

const (
	handshakeNonceSize = 8
	maxHandshakeFrame  = 1 << 16 // generous cap; hybrid INIT is ~1.2 KB
)

// sessionInit is the initiator's first handshake message (SESSION_INIT).
type sessionInit struct {
	Nonce        []byte  `msgpack:"nonce"`
	Timestamp    uint64  `msgpack:"timestamp"`
	KexMode      KexMode `msgpack:"kex_mode"`
	X25519Public []byte  `msgpack:"x25519_public"`
	MLKEMPublic  []byte  `msgpack:"mlkem_public,omitempty"`
}

// sessionAck is the responder's reply (SESSION_ACK).
type sessionAck struct {
	Nonce           []byte  `msgpack:"nonce"`
	KexMode         KexMode `msgpack:"kex_mode"`
	X25519Public    []byte  `msgpack:"x25519_public"`
	MLKEMCiphertext []byte  `msgpack:"mlkem_ciphertext,omitempty"`
}

// DerivedSession is the output of a completed handshake: the directional Tier-3+
// key material plus the transcript hash (for a subsequent identity exchange).
type DerivedSession struct {
	SessionID  uint64
	SendKey    [crypto.KeySize]byte
	RecvKey    [crypto.KeySize]byte
	SendSeed   [crypto.NonceSize]byte
	RecvSeed   [crypto.NonceSize]byte
	Transcript [32]byte
}

// AeadSession builds a Tier-3 session from the derived key material.
func (d *DerivedSession) AeadSession() (*AeadSession, error) {
	return NewAeadSessionFromKeys(d.SessionID, d.SendKey[:], d.RecvKey[:], d.SendSeed, d.RecvSeed)
}

// Zero wipes the derived key material.
func (d *DerivedSession) Zero() {
	crypto.ZeroBytes(d.SendKey[:])
	crypto.ZeroBytes(d.RecvKey[:])
	crypto.ZeroBytes(d.SendSeed[:])
	crypto.ZeroBytes(d.RecvSeed[:])
}

// ClientHandshake runs the initiator side of the hybrid KEX over conn using the
// given mode, returning the derived session keys and transcript. The handshake
// is aborted (conn closed) if ctx is canceled.
func ClientHandshake(ctx context.Context, conn io.ReadWriteCloser, mode KexMode) (*DerivedSession, error) {
	if err := validKexMode(mode); err != nil {
		return nil, err
	}
	defer guardHandshakeConn(ctx, conn)()

	eph, err := crypto.GenerateX25519KeyPair()
	if err != nil {
		return nil, fmt.Errorf("protocol: client ephemeral x25519: %w", err)
	}
	nonceC, err := randomBytes(handshakeNonceSize)
	if err != nil {
		return nil, err
	}
	init, kem, err := buildClientInit(eph, nonceC, mode)
	if err != nil {
		return nil, err
	}

	initBytes, err := msgpack.Marshal(init)
	if err != nil {
		return nil, fmt.Errorf("protocol: encode session-init: %w", err)
	}
	if err := writeHandshakeFrame(conn, initBytes); err != nil {
		return nil, err
	}

	ackBytes, err := readHandshakeFrame(conn)
	if err != nil {
		return nil, err
	}
	var ack sessionAck
	if err := msgpack.Unmarshal(ackBytes, &ack); err != nil {
		return nil, fmt.Errorf("%w: session-ack: %v", ErrHandshakeFraming, err)
	}
	if ack.KexMode != mode {
		return nil, ErrKexMode
	}

	combined, err := clientCombined(eph, kem, mode, ack)
	if err != nil {
		return nil, err
	}
	defer crypto.ZeroBytes(combined) // wipe the raw KEX master secret after derivation
	transcript := TranscriptHash(initBytes, ackBytes)
	return deriveSession(combined, mode, transcript, nonceC, ack.Nonce, true)
}

// buildClientInit assembles the SESSION_INIT message and, in hybrid mode, the
// initiator's ephemeral ML-KEM key pair (returned so the caller can decapsulate).
func buildClientInit(eph *crypto.X25519KeyPair, nonceC []byte, mode KexMode) (sessionInit, *crypto.MLKEMKeyPair, error) {
	init := sessionInit{
		Nonce:        nonceC,
		Timestamp:    nowUnix(),
		KexMode:      mode,
		X25519Public: eph.Public[:],
	}
	if mode != KexModeHybridMLKEM768 {
		return init, nil, nil
	}
	kem, err := crypto.GenerateMLKEMKeyPair()
	if err != nil {
		return sessionInit{}, nil, fmt.Errorf("protocol: client ml-kem keygen: %w", err)
	}
	init.MLKEMPublic = kem.EncapsulationKeyBytes()
	return init, kem, nil
}

// clientCombined computes the initiator's combined shared secret from the ACK.
func clientCombined(eph *crypto.X25519KeyPair, kem *crypto.MLKEMKeyPair, mode KexMode, ack sessionAck) ([]byte, error) {
	peerPub, err := publicKey(ack.X25519Public)
	if err != nil {
		return nil, err
	}
	ssClassical, err := eph.ECDH(peerPub)
	if err != nil {
		return nil, fmt.Errorf("protocol: client ecdh: %w", err)
	}
	combined := append([]byte(nil), ssClassical[:]...)
	if mode != KexModeHybridMLKEM768 {
		return combined, nil
	}
	if len(ack.MLKEMCiphertext) != crypto.MLKEMCiphertextSize {
		return nil, ErrHandshakeKey
	}
	ssPQ, err := kem.Decapsulate(ack.MLKEMCiphertext)
	if err != nil {
		return nil, fmt.Errorf("protocol: client ml-kem decapsulate: %w", err)
	}
	return append(combined, ssPQ[:]...), nil
}

// ServerHandshake runs the responder side of the hybrid KEX over conn. It
// rejects any offered mode weaker than minMode. The handshake is aborted (conn
// closed) if ctx is canceled.
func ServerHandshake(ctx context.Context, conn io.ReadWriteCloser, minMode KexMode) (*DerivedSession, error) {
	defer guardHandshakeConn(ctx, conn)()

	initBytes, err := readHandshakeFrame(conn)
	if err != nil {
		return nil, err
	}
	var init sessionInit
	if err := msgpack.Unmarshal(initBytes, &init); err != nil {
		return nil, fmt.Errorf("%w: session-init: %v", ErrHandshakeFraming, err)
	}
	mode := init.KexMode
	if err := acceptableMode(mode, minMode); err != nil {
		return nil, err
	}
	peerPub, err := publicKey(init.X25519Public)
	if err != nil {
		return nil, err
	}

	eph, err := crypto.GenerateX25519KeyPair()
	if err != nil {
		return nil, fmt.Errorf("protocol: server ephemeral x25519: %w", err)
	}
	nonceS, err := randomBytes(handshakeNonceSize)
	if err != nil {
		return nil, err
	}
	ack := sessionAck{Nonce: nonceS, KexMode: mode, X25519Public: eph.Public[:]}

	var ssPQ [crypto.MLKEMSharedKeySize]byte
	if mode == KexModeHybridMLKEM768 {
		shared, ct, eerr := serverEncapsulate(init.MLKEMPublic)
		if eerr != nil {
			return nil, eerr
		}
		ssPQ = shared
		ack.MLKEMCiphertext = ct
	}

	ackBytes, err := msgpack.Marshal(ack)
	if err != nil {
		return nil, fmt.Errorf("protocol: encode session-ack: %w", err)
	}
	if err := writeHandshakeFrame(conn, ackBytes); err != nil {
		return nil, err
	}

	combined, err := serverCombined(eph, peerPub, mode, ssPQ)
	if err != nil {
		return nil, err
	}
	defer crypto.ZeroBytes(combined) // wipe the raw KEX master secret after derivation
	transcript := TranscriptHash(initBytes, ackBytes)
	return deriveSession(combined, mode, transcript, init.Nonce, nonceS, false)
}

// serverEncapsulate validates the initiator's ML-KEM key and encapsulates to it.
func serverEncapsulate(mlkemPub []byte) (shared [crypto.MLKEMSharedKeySize]byte, ciphertext []byte, err error) {
	if len(mlkemPub) != crypto.MLKEMEncapKeySize {
		return shared, nil, ErrHandshakeKey
	}
	shared, ciphertext, err = crypto.MLKEMEncapsulate(mlkemPub)
	if err != nil {
		return shared, nil, fmt.Errorf("protocol: server ml-kem encapsulate: %w", err)
	}
	return shared, ciphertext, nil
}

// serverCombined computes the responder's combined shared secret.
func serverCombined(eph *crypto.X25519KeyPair, peerPub [crypto.X25519KeySize]byte, mode KexMode, ssPQ [crypto.MLKEMSharedKeySize]byte) ([]byte, error) {
	ssClassical, err := eph.ECDH(peerPub)
	if err != nil {
		return nil, fmt.Errorf("protocol: server ecdh: %w", err)
	}
	combined := append([]byte(nil), ssClassical[:]...)
	if mode == KexModeHybridMLKEM768 {
		combined = append(combined, ssPQ[:]...)
	}
	return combined, nil
}

func validKexMode(mode KexMode) error {
	if mode != KexModeClassicalOnly && mode != KexModeHybridMLKEM768 {
		return ErrKexMode
	}
	return nil
}

// acceptableMode checks an offered mode is valid and at least as strong as the
// minimum (the constant values increase with strength, so a smaller value is a
// downgrade and is rejected).
func acceptableMode(mode, minMode KexMode) error {
	if err := validKexMode(mode); err != nil {
		return err
	}
	if mode < minMode {
		return ErrKexMode
	}
	return nil
}

// deriveSession expands the combined secret into directional Tier-3 key material,
// folding the transcript into the HKDF info for downgrade protection.
func deriveSession(combined []byte, mode KexMode, transcript [32]byte, nonceInit, nonceResp []byte, initiator bool) (*DerivedSession, error) {
	salt := make([]byte, 0, len(nonceInit)+len(nonceResp))
	salt = append(salt, nonceInit...)
	salt = append(salt, nonceResp...)

	info := make([]byte, 0, len(mode.String())+len(transcript))
	info = append(info, mode.String()...)
	info = append(info, transcript[:]...)

	keys, err := crypto.HKDFDeriveKeys(combined, salt, info,
		crypto.KeySize, crypto.KeySize, crypto.NonceSize, crypto.NonceSize, 8)
	if err != nil {
		return nil, fmt.Errorf("protocol: derive session keys: %w", err)
	}
	defer func() {
		for _, k := range keys {
			crypto.ZeroBytes(k)
		}
	}()

	// keys[0]=initiator-send (c2s), keys[1]=initiator-recv (s2c); responder swaps.
	sendKey, recvKey, sendSeed, recvSeed := keys[0], keys[1], keys[2], keys[3]
	if !initiator {
		sendKey, recvKey, sendSeed, recvSeed = keys[1], keys[0], keys[3], keys[2]
	}

	d := &DerivedSession{
		SessionID:  binary.BigEndian.Uint64(keys[4]),
		Transcript: transcript,
	}
	copy(d.SendKey[:], sendKey)
	copy(d.RecvKey[:], recvKey)
	copy(d.SendSeed[:], sendSeed)
	copy(d.RecvSeed[:], recvSeed)
	return d, nil
}

// TranscriptHash computes SHA-256(initPayload || ackPayload). Both endpoints
// must hash the identical wire bytes or they derive different keys.
func TranscriptHash(initPayload, ackPayload []byte) [32]byte {
	h := sha256.New()
	_, _ = h.Write(initPayload)
	_, _ = h.Write(ackPayload)
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out
}

func publicKey(b []byte) ([crypto.X25519KeySize]byte, error) {
	var pub [crypto.X25519KeySize]byte
	if len(b) != crypto.X25519KeySize {
		return pub, ErrHandshakeKey
	}
	copy(pub[:], b)
	return pub, nil
}

func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("protocol: random bytes: %w", err)
	}
	return b, nil
}

func nowUnix() uint64 {
	t := time.Now().Unix()
	if t < 0 {
		return 0
	}
	return uint64(t)
}

// writeHandshakeFrame writes a 4-byte big-endian length prefix followed by payload.
func writeHandshakeFrame(w io.Writer, payload []byte) error {
	if len(payload) > maxHandshakeFrame {
		return ErrHandshakeFraming
	}
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(len(payload))) //nolint:gosec // G115: len bounded by maxHandshakeFrame above
	if _, err := w.Write(hdr[:]); err != nil {
		return fmt.Errorf("protocol: write handshake frame: %w", err)
	}
	if _, err := w.Write(payload); err != nil {
		return fmt.Errorf("protocol: write handshake frame: %w", err)
	}
	return nil
}

// readHandshakeFrame reads a length-prefixed handshake payload, bounded by maxHandshakeFrame.
func readHandshakeFrame(r io.Reader) ([]byte, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return nil, fmt.Errorf("protocol: read handshake length: %w", err)
	}
	n := binary.BigEndian.Uint32(hdr[:])
	if n == 0 || n > maxHandshakeFrame {
		return nil, ErrHandshakeFraming
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, fmt.Errorf("protocol: read handshake frame: %w", err)
	}
	return buf, nil
}

// guardHandshakeConn closes conn if ctx is canceled before the returned stop
// func runs, so a stalled handshake cannot block forever. The caller defers the
// returned stop.
func guardHandshakeConn(ctx context.Context, conn io.Closer) func() {
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-done:
		}
	}()
	return func() { close(done) }
}
