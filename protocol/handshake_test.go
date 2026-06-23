package protocol

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"
)

type handshakeResult struct {
	d   *DerivedSession
	err error
}

func runHandshake(t *testing.T, clientMode, serverMin KexMode) (clientRes, serverRes handshakeResult) {
	t.Helper()
	cConn, sConn := net.Pipe()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cc := make(chan handshakeResult, 1)
	sc := make(chan handshakeResult, 1)
	go func() {
		d, err := ClientHandshake(ctx, cConn, clientMode)
		cc <- handshakeResult{d, err}
	}()
	go func() {
		d, err := ServerHandshake(ctx, sConn, serverMin)
		sc <- handshakeResult{d, err}
	}()
	return <-cc, <-sc
}

func TestHandshakeHybridRoundTrip(t *testing.T) {
	c, s := runHandshake(t, KexModeHybridMLKEM768, KexModeHybridMLKEM768)
	if c.err != nil || s.err != nil {
		t.Fatalf("handshake errors: client=%v server=%v", c.err, s.err)
	}
	assertMatchingSessions(t, c.d, s.d)
	assertAeadInterop(t, c.d, s.d)
}

func TestHandshakeClassicalRoundTrip(t *testing.T) {
	c, s := runHandshake(t, KexModeClassicalOnly, KexModeClassicalOnly)
	if c.err != nil || s.err != nil {
		t.Fatalf("handshake errors: client=%v server=%v", c.err, s.err)
	}
	assertMatchingSessions(t, c.d, s.d)
	assertAeadInterop(t, c.d, s.d)
}

func TestHandshakeRejectsDowngrade(t *testing.T) {
	c, s := runHandshake(t, KexModeClassicalOnly, KexModeHybridMLKEM768)
	if s.err == nil {
		t.Fatal("server accepted a classical handshake despite hybrid minimum")
	}
	// The client may or may not observe the error depending on timing; the
	// security property is that the server refuses.
	_ = c
}

func assertMatchingSessions(t *testing.T, c, s *DerivedSession) {
	t.Helper()
	if c.SessionID != s.SessionID {
		t.Errorf("session id mismatch: client=%d server=%d", c.SessionID, s.SessionID)
	}
	if c.Transcript != s.Transcript {
		t.Error("transcript mismatch")
	}
	if !bytes.Equal(c.SendKey[:], s.RecvKey[:]) {
		t.Error("client send key != server recv key")
	}
	if !bytes.Equal(c.RecvKey[:], s.SendKey[:]) {
		t.Error("client recv key != server send key")
	}
	if bytes.Equal(c.SendKey[:], c.RecvKey[:]) {
		t.Error("directional keys are identical (no direction separation)")
	}
}

func assertAeadInterop(t *testing.T, c, s *DerivedSession) {
	t.Helper()
	cs, err := c.AeadSession()
	if err != nil {
		t.Fatalf("client AeadSession: %v", err)
	}
	ss, err := s.AeadSession()
	if err != nil {
		t.Fatalf("server AeadSession: %v", err)
	}
	frame, err := cs.Encode(OpKeepalive, []byte("ping"))
	if err != nil {
		t.Fatalf("client Encode: %v", err)
	}
	op, pt, err := ss.Decode(frame)
	if err != nil {
		t.Fatalf("server Decode: %v", err)
	}
	if op != OpKeepalive || string(pt) != "ping" {
		t.Fatalf("decoded op=%v payload=%q", op, pt)
	}
}

func TestHandshakeContextCancel(t *testing.T) {
	cConn, _ := net.Pipe() // no server reading; client should block then abort
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := ClientHandshake(ctx, cConn, KexModeHybridMLKEM768)
		done <- err
	}()
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected error after context cancel")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("handshake did not abort on context cancel")
	}
}
