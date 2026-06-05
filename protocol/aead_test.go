package protocol

import (
	"bytes"
	"errors"
	"math"
	"testing"

	"arcan-it.de/arclib/crypto"
)

func testSecret() []byte {
	s := make([]byte, crypto.KeySize)
	for i := range s {
		s[i] = byte(i + 1)
	}
	return s
}

// sessionPair derives a fresh client/server session from the same secret+salts.
func sessionPair(t *testing.T) (client, server *AeadSession) {
	t.Helper()
	cs := bytes.Repeat([]byte{0xA1}, SessionSaltSize)
	ss := bytes.Repeat([]byte{0xB2}, SessionSaltSize)
	c, err := NewAeadSession(testSecret(), cs, ss, true)
	if err != nil {
		t.Fatalf("client session: %v", err)
	}
	s, err := NewAeadSession(testSecret(), cs, ss, false)
	if err != nil {
		t.Fatalf("server session: %v", err)
	}
	return c, s
}

func TestAeadSessionRoundTrip(t *testing.T) {
	client, server := sessionPair(t)
	if client.SessionID() != server.SessionID() {
		t.Fatalf("session ids differ: %x vs %x", client.SessionID(), server.SessionID())
	}

	// Client -> server.
	frame, err := client.Encode(OpArcMailAccountList, []byte("ping"))
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	op, pt, err := server.Decode(frame)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if op != OpArcMailAccountList || string(pt) != "ping" {
		t.Errorf("decoded (0x%04X, %q), want (ACCOUNT_LIST, ping)", uint16(op), pt)
	}

	// Server -> client uses the opposite directional key/seed.
	rframe, err := server.Encode(OpArcMailEvent, []byte("event"))
	if err != nil {
		t.Fatalf("server encode: %v", err)
	}
	rop, rpt, err := client.Decode(rframe)
	if err != nil {
		t.Fatalf("client decode: %v", err)
	}
	if rop != OpArcMailEvent || string(rpt) != "event" {
		t.Errorf("decoded (0x%04X, %q), want (EVENT, event)", uint16(rop), rpt)
	}
}

func TestAeadSessionTamperRejected(t *testing.T) {
	for _, idx := range []int{2 /* opcode (AAD) */, aeadHeaderSize + 1 /* ciphertext */} {
		client, server := sessionPair(t)
		frame, err := client.Encode(OpArcMailMessageGet, []byte("payload"))
		if err != nil {
			t.Fatalf("encode: %v", err)
		}
		frame[idx] ^= 0xFF
		if _, _, err := server.Decode(frame); err == nil {
			t.Errorf("tampering byte %d was not detected", idx)
		}
	}
}

func TestAeadSessionReplayRejected(t *testing.T) {
	client, server := sessionPair(t)
	frame, err := client.Encode(OpArcMailMessageList, []byte("x"))
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if _, _, err := server.Decode(frame); err != nil {
		t.Fatalf("first decode: %v", err)
	}
	if _, _, err := server.Decode(frame); !errors.Is(err, ErrReplayedSequence) {
		t.Errorf("replay err = %v, want ErrReplayedSequence", err)
	}
}

func TestAeadSessionOutOfOrderRejected(t *testing.T) {
	client, server := sessionPair(t)
	f0, _ := client.Encode(OpArcMailMessagePage, []byte("0"))
	f1, _ := client.Encode(OpArcMailMessagePage, []byte("1"))
	if _, _, err := server.Decode(f1); err != nil {
		t.Fatalf("decode f1: %v", err)
	}
	// f0 carries an earlier sequence than the watermark -> rejected.
	if _, _, err := server.Decode(f0); !errors.Is(err, ErrReplayedSequence) {
		t.Errorf("out-of-order err = %v, want ErrReplayedSequence", err)
	}
}

func TestAeadSessionWrongSessionRejected(t *testing.T) {
	_, server := sessionPair(t)
	// A different connection (different salts) yields a different session id.
	other, err := NewAeadSession(testSecret(),
		bytes.Repeat([]byte{0x11}, SessionSaltSize), bytes.Repeat([]byte{0x22}, SessionSaltSize), true)
	if err != nil {
		t.Fatalf("other session: %v", err)
	}
	frame, _ := other.Encode(OpArcMailAccountList, []byte("x"))
	if _, _, err := server.Decode(frame); !errors.Is(err, ErrSessionMismatch) {
		t.Errorf("wrong-session err = %v, want ErrSessionMismatch", err)
	}
}

func TestAeadSessionBadInputs(t *testing.T) {
	salt := bytes.Repeat([]byte{0x01}, SessionSaltSize)
	if _, err := NewAeadSession(make([]byte, 16), salt, salt, true); !errors.Is(err, ErrSessionSecretSize) {
		t.Errorf("short secret err = %v, want ErrSessionSecretSize", err)
	}
	if _, err := NewAeadSession(testSecret(), salt[:4], salt, true); !errors.Is(err, ErrSessionSaltSize) {
		t.Errorf("short salt err = %v, want ErrSessionSaltSize", err)
	}
	client, server := sessionPair(t)
	frame, _ := client.Encode(OpArcMailAccountList, nil)
	if _, _, err := server.Decode(frame[:aeadHeaderSize+aeadTagSize-1]); !errors.Is(err, ErrFrameTooShort) {
		t.Errorf("short frame err = %v, want ErrFrameTooShort", err)
	}
}

func TestAeadSessionSequenceExhausted(t *testing.T) {
	client, _ := sessionPair(t)
	// Drive the counter so the next Encode consumes the last valid wire sequence
	// (MaxUint32); the one after that would wrap to 0 and reuse a nonce, so it must
	// error instead of sealing.
	client.sendSeq = math.MaxUint32
	if _, err := client.Encode(OpArcMailAccountList, []byte("last")); err != nil {
		t.Fatalf("final valid frame errored: %v", err)
	}
	if _, err := client.Encode(OpArcMailAccountList, []byte("over")); !errors.Is(err, ErrSequenceExhausted) {
		t.Errorf("exhausted err = %v, want ErrSequenceExhausted", err)
	}
}

func TestGenerateSessionSalt(t *testing.T) {
	a, err := GenerateSessionSalt()
	if err != nil || len(a) != SessionSaltSize {
		t.Fatalf("salt a = %d bytes, %v", len(a), err)
	}
	b, _ := GenerateSessionSalt()
	if bytes.Equal(a, b) {
		t.Error("two salts are identical (RNG broken?)")
	}
}
