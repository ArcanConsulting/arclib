package transport

import (
	"errors"
	"testing"

	"arcan-it.de/arclib/msgpack"
	"arcan-it.de/arclib/protocol"
)

type frame struct {
	op      protocol.OpCode
	payload []byte
}

func TestHandshakeInMemoryInterop(t *testing.T) {
	toServer := make(chan frame, 2)
	toClient := make(chan frame, 2)

	clientWrite := func(op protocol.OpCode, p []byte) error { toServer <- frame{op, p}; return nil }
	clientRead := func() (protocol.OpCode, []byte, error) { f := <-toClient; return f.op, f.payload, nil }
	serverWrite := func(op protocol.OpCode, p []byte) error { toClient <- frame{op, p}; return nil }
	serverRead := func() (protocol.OpCode, []byte, error) { f := <-toServer; return f.op, f.payload, nil }

	var serverAead *protocol.AeadSession
	var serverErr error
	done := make(chan struct{})
	go func() {
		serverAead, serverErr = serverHandshake(serverWrite, serverRead)
		close(done)
	}()

	clientAead, err := clientHandshake(clientWrite, clientRead)
	if err != nil {
		t.Fatalf("clientHandshake: %v", err)
	}
	<-done
	if serverErr != nil {
		t.Fatalf("serverHandshake: %v", serverErr)
	}
	if clientAead.SessionID() != serverAead.SessionID() {
		t.Fatalf("session id mismatch: %d != %d", clientAead.SessionID(), serverAead.SessionID())
	}

	// Bidirectional Tier-3 interop proves both derived identical keys/seeds.
	cFrame, err := clientAead.Encode(opEcho, []byte("ping"))
	if err != nil {
		t.Fatalf("client encode: %v", err)
	}
	op, pt, err := serverAead.Decode(cFrame)
	if err != nil || op != opEcho || string(pt) != "ping" {
		t.Fatalf("server decode: op=%s pt=%q err=%v", op, pt, err)
	}
	sFrame, err := serverAead.Encode(opResp, []byte("pong"))
	if err != nil {
		t.Fatalf("server encode: %v", err)
	}
	op2, pt2, err := clientAead.Decode(sFrame)
	if err != nil || op2 != opResp || string(pt2) != "pong" {
		t.Fatalf("client decode: op=%s pt=%q err=%v", op2, pt2, err)
	}
}

func TestClientHandshakeUnexpectedOp(t *testing.T) {
	write := func(protocol.OpCode, []byte) error { return nil }
	read := func() (protocol.OpCode, []byte, error) { return protocol.OpNop, nil, nil }
	if _, err := clientHandshake(write, read); !errors.Is(err, ErrHandshakeUnexpectedOp) {
		t.Fatalf("got %v", err)
	}
}

func TestServerHandshakeUnexpectedOp(t *testing.T) {
	write := func(protocol.OpCode, []byte) error { return nil }
	read := func() (protocol.OpCode, []byte, error) { return protocol.OpNop, nil, nil }
	if _, err := serverHandshake(write, read); !errors.Is(err, ErrHandshakeUnexpectedOp) {
		t.Fatalf("got %v", err)
	}
}

func TestServerHandshakeBadInitKey(t *testing.T) {
	bad, _ := msgpack.Marshal(handshakeInit{
		X25519Pub:  []byte{1, 2, 3}, // too short
		MLKEMEncap: make([]byte, 1184),
		ClientSalt: make([]byte, protocol.SessionSaltSize),
	})
	read := func() (protocol.OpCode, []byte, error) { return protocol.OpSessionInit, bad, nil }
	write := func(protocol.OpCode, []byte) error { return nil }
	if _, err := serverHandshake(write, read); !errors.Is(err, ErrHandshakeBadKey) {
		t.Fatalf("got %v", err)
	}
}

func TestServerHandshakeBadSalt(t *testing.T) {
	bad, _ := msgpack.Marshal(handshakeInit{
		X25519Pub:  make([]byte, 32),
		MLKEMEncap: make([]byte, 1184),
		ClientSalt: []byte{1}, // wrong salt size
	})
	read := func() (protocol.OpCode, []byte, error) { return protocol.OpSessionInit, bad, nil }
	write := func(protocol.OpCode, []byte) error { return nil }
	if _, err := serverHandshake(write, read); !errors.Is(err, ErrHandshakeBadKey) {
		t.Fatalf("got %v", err)
	}
}

func TestServerHandshakeBadInitDecode(t *testing.T) {
	read := func() (protocol.OpCode, []byte, error) { return protocol.OpSessionInit, []byte{0xff, 0xff}, nil }
	write := func(protocol.OpCode, []byte) error { return nil }
	if _, err := serverHandshake(write, read); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestClientHandshakeReadError(t *testing.T) {
	sentinel := errors.New("read fail")
	write := func(protocol.OpCode, []byte) error { return nil }
	read := func() (protocol.OpCode, []byte, error) { return 0, nil, sentinel }
	if _, err := clientHandshake(write, read); !errors.Is(err, sentinel) {
		t.Fatalf("got %v", err)
	}
}

func TestClientHandshakeWriteError(t *testing.T) {
	sentinel := errors.New("write fail")
	write := func(protocol.OpCode, []byte) error { return sentinel }
	read := func() (protocol.OpCode, []byte, error) { return protocol.OpSessionAck, nil, nil }
	if _, err := clientHandshake(write, read); !errors.Is(err, sentinel) {
		t.Fatalf("got %v", err)
	}
}

func TestServerHandshakeReadError(t *testing.T) {
	sentinel := errors.New("read fail")
	write := func(protocol.OpCode, []byte) error { return nil }
	read := func() (protocol.OpCode, []byte, error) { return 0, nil, sentinel }
	if _, err := serverHandshake(write, read); !errors.Is(err, sentinel) {
		t.Fatalf("got %v", err)
	}
}
