package protocol

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"
)

// matchedPair returns two DerivedSessions whose directions are crossed (client
// send == server recv) by running a real hybrid handshake over a pipe.
func matchedPair(t *testing.T) (client, server *DerivedSession) {
	t.Helper()
	c, s := runHandshake(t, KexModeHybridMLKEM768, KexModeHybridMLKEM768)
	if c.err != nil || s.err != nil {
		t.Fatalf("handshake: client=%v server=%v", c.err, s.err)
	}
	return c.d, s.d
}

func TestTier5RoundTripBothDirections(t *testing.T) {
	cd, sd := matchedPair(t)
	cs, err := cd.Tier5Session()
	if err != nil {
		t.Fatalf("client Tier5Session: %v", err)
	}
	ss, err := sd.Tier5Session()
	if err != nil {
		t.Fatalf("server Tier5Session: %v", err)
	}

	// client -> server
	frame, err := cs.Encode(OpKeepalive, []byte("c2s"))
	if err != nil {
		t.Fatalf("client Encode: %v", err)
	}
	op, pt, err := ss.Decode(frame)
	if err != nil || op != OpKeepalive || string(pt) != "c2s" {
		t.Fatalf("server Decode: op=%v pt=%q err=%v", op, pt, err)
	}
	// server -> client
	frame, err = ss.Encode(OpKeepaliveAck, []byte("s2c"))
	if err != nil {
		t.Fatalf("server Encode: %v", err)
	}
	op, pt, err = cs.Decode(frame)
	if err != nil || op != OpKeepaliveAck || string(pt) != "s2c" {
		t.Fatalf("client Decode: op=%v pt=%q err=%v", op, pt, err)
	}
}

func TestTier5RotationInterop(t *testing.T) {
	cd, sd := matchedPair(t)
	cs, err := NewTier5Session(cd, 4, time.Hour) // rotate every 4 messages
	if err != nil {
		t.Fatalf("client session: %v", err)
	}
	ss, err := NewTier5Session(sd, 4, time.Hour)
	if err != nil {
		t.Fatalf("server session: %v", err)
	}

	const n = 20
	for i := 0; i < n; i++ {
		msg := []byte{byte(i)}
		frame, eerr := cs.Encode(OpNop, msg)
		if eerr != nil {
			t.Fatalf("Encode %d: %v", i, eerr)
		}
		_, pt, derr := ss.Decode(frame)
		if derr != nil {
			t.Fatalf("Decode %d: %v", i, derr)
		}
		if !bytes.Equal(pt, msg) {
			t.Fatalf("message %d mangled: got %v", i, pt)
		}
	}
	if cs.sendEpoch == 0 {
		t.Fatal("send key never rotated despite exceeding the message budget")
	}
	if ss.recvEpoch != cs.sendEpoch {
		t.Fatalf("receiver epoch %d did not follow sender epoch %d", ss.recvEpoch, cs.sendEpoch)
	}
}

// TestTier5ForgedFrameDoesNotPoison verifies a forged frame at a future epoch
// fails to decrypt AND leaves the receive state intact, so a legit frame at the
// current epoch still decodes afterward.
func TestTier5ForgedFrameDoesNotPoison(t *testing.T) {
	cd, sd := matchedPair(t)
	cs, _ := cd.Tier5Session()
	ss, _ := sd.Tier5Session()

	// Forged frame claiming epoch 1 with garbage ciphertext.
	forged := make([]byte, tier5HeaderSize+40)
	forged[0] = tier5Flags
	binary.BigEndian.PutUint64(forged[7:], ss.sessionID)
	binary.BigEndian.PutUint32(forged[15:], 1) // epoch 1
	for i := tier5HeaderSize; i < len(forged); i++ {
		forged[i] = 0xAB
	}
	if _, _, err := ss.Decode(forged); err == nil {
		t.Fatal("forged future-epoch frame was accepted")
	}
	if ss.recvEpoch != 0 {
		t.Fatalf("forged frame poisoned receive epoch: now %d", ss.recvEpoch)
	}

	// A legitimate current-epoch frame must still decode.
	frame, _ := cs.Encode(OpNop, []byte("ok"))
	if _, pt, err := ss.Decode(frame); err != nil || string(pt) != "ok" {
		t.Fatalf("legit frame after forgery failed: pt=%q err=%v", pt, err)
	}
}

func TestTier5RejectsReplayAndTamper(t *testing.T) {
	cd, sd := matchedPair(t)
	cs, _ := cd.Tier5Session()
	ss, _ := sd.Tier5Session()

	frame, _ := cs.Encode(OpNop, []byte("once"))
	if _, _, err := ss.Decode(frame); err != nil {
		t.Fatalf("first decode: %v", err)
	}
	if _, _, err := ss.Decode(frame); err == nil {
		t.Fatal("replayed frame was accepted")
	}

	frame2, _ := cs.Encode(OpNop, []byte("twice"))
	bad := bytes.Clone(frame2)
	bad[len(bad)-1] ^= 0xFF // corrupt the auth tag
	if _, _, err := ss.Decode(bad); err == nil {
		t.Fatal("tampered frame was accepted")
	}
}
