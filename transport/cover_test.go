package transport

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"arcan-it.de/arclib/crypto"
	"arcan-it.de/arclib/msgpack"
	"arcan-it.de/arclib/protocol"
)

func noopWrite(protocol.OpCode, []byte) error { return nil }

func readReturning(op protocol.OpCode, p []byte) readFrameFunc {
	return func() (protocol.OpCode, []byte, error) { return op, p, nil }
}

func TestHandlerRegistryAPI(t *testing.T) {
	r := NewHandlerRegistry()
	if r.Count() != 0 || r.Has(opEcho) {
		t.Fatal("empty registry expected")
	}
	if err := r.Register(opEcho, nil); err == nil {
		t.Fatal("nil handler should error")
	}
	h := HandlerFunc(func(context.Context, *Session, *protocol.Message) (*protocol.Message, error) { return nil, nil }) //nolint:nilnil // no-op handler for registry tests
	if err := r.RegisterFunc(opEcho, h); err != nil {
		t.Fatalf("register: %v", err)
	}
	if !r.Has(opEcho) || r.Count() != 1 {
		t.Fatal("expected registered handler")
	}
	if err := r.Register(opEcho, h); err == nil {
		t.Fatal("duplicate registration should error")
	}
}

func TestSessionUnit(t *testing.T) {
	s := newSession(42, nil, func([]byte) error { return nil })
	if s.ID() != 42 {
		t.Fatalf("ID = %d", s.ID())
	}
	if s.UserID() != "" || s.FamilyID() != "" || s.DeviceID() != "" {
		t.Fatal("identity should start empty")
	}
	s.SetIdentity("u", "f", "d")
	if s.UserID() != "u" || s.FamilyID() != "f" || s.DeviceID() != "d" {
		t.Fatalf("identity = %s/%s/%s", s.UserID(), s.FamilyID(), s.DeviceID())
	}
}

func TestWireErrorRoundTrip(t *testing.T) {
	if got := decodeError(encodeError(errors.New("oops"))); got.Error() != "oops" {
		t.Fatalf("got %v", got)
	}
	if decodeError([]byte{0xff, 0xff}) == nil {
		t.Fatal("malformed error payload should still yield an error")
	}
	if decodeError(nil) == nil {
		t.Fatal("nil error payload should still yield an error")
	}
}

func validInit(t *testing.T) []byte {
	t.Helper()
	cx, err := crypto.GenerateX25519KeyPair()
	if err != nil {
		t.Fatal(err)
	}
	cm, err := crypto.GenerateMLKEMKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	salt, err := protocol.GenerateSessionSalt()
	if err != nil {
		t.Fatal(err)
	}
	b, err := msgpack.Marshal(handshakeInit{X25519Pub: cx.Public[:], MLKEMEncap: cm.EncapsulationKeyBytes(), ClientSalt: salt})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestServerHandshakeWriteAckError(t *testing.T) {
	sentinel := errors.New("write fail")
	_, err := serverHandshake(func(protocol.OpCode, []byte) error { return sentinel }, readReturning(protocol.OpSessionInit, validInit(t)))
	if !errors.Is(err, sentinel) {
		t.Fatalf("got %v", err)
	}
}

func TestServerHandshakeMLKEMFail(t *testing.T) {
	init, _ := msgpack.Marshal(handshakeInit{X25519Pub: make([]byte, 32), MLKEMEncap: []byte{1, 2, 3}, ClientSalt: make([]byte, protocol.SessionSaltSize)})
	if _, err := serverHandshake(noopWrite, readReturning(protocol.OpSessionInit, init)); err == nil {
		t.Fatal("expected ml-kem encapsulate error")
	}
}

func TestClientHandshakeBadAckKey(t *testing.T) {
	ack, _ := msgpack.Marshal(handshakeAck{X25519Pub: []byte{1}, MLKEMCiphertext: make([]byte, 1088), ServerSalt: make([]byte, 16)})
	if _, err := clientHandshake(noopWrite, readReturning(protocol.OpSessionAck, ack)); !errors.Is(err, ErrHandshakeBadKey) {
		t.Fatalf("got %v", err)
	}
}

func TestClientHandshakeBadAckSalt(t *testing.T) {
	ack, _ := msgpack.Marshal(handshakeAck{X25519Pub: make([]byte, 32), MLKEMCiphertext: make([]byte, 1088), ServerSalt: []byte{1}})
	if _, err := clientHandshake(noopWrite, readReturning(protocol.OpSessionAck, ack)); !errors.Is(err, ErrHandshakeBadKey) {
		t.Fatalf("got %v", err)
	}
}

func TestClientHandshakeBadAckDecode(t *testing.T) {
	if _, err := clientHandshake(noopWrite, readReturning(protocol.OpSessionAck, []byte{0xff})); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestClientHandshakeECDHFail(t *testing.T) {
	// All-zero peer public is a low-order point -> ECDH error.
	ack, _ := msgpack.Marshal(handshakeAck{X25519Pub: make([]byte, 32), MLKEMCiphertext: make([]byte, 1088), ServerSalt: make([]byte, 16)})
	if _, err := clientHandshake(noopWrite, readReturning(protocol.OpSessionAck, ack)); err == nil {
		t.Fatal("expected x25519 ecdh error")
	}
}

func TestClientHandshakeMLKEMFail(t *testing.T) {
	kp, _ := crypto.GenerateX25519KeyPair()
	ack, _ := msgpack.Marshal(handshakeAck{X25519Pub: kp.Public[:], MLKEMCiphertext: []byte{1, 2, 3}, ServerSalt: make([]byte, 16)})
	if _, err := clientHandshake(noopWrite, readReturning(protocol.OpSessionAck, ack)); err == nil {
		t.Fatal("expected ml-kem decapsulate error")
	}
}

func TestServerRejectsBadTier3Frame(t *testing.T) {
	s := newTestServer(t)
	ws, resp, err := websocket.DefaultDialer.Dial("ws://"+s.Addr()+"/ws", nil)
	if resp != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = ws.Close() }()

	aead, err := clientHandshake(
		func(op protocol.OpCode, p []byte) error { return writeTier0(ws, op, p) },
		func() (protocol.OpCode, []byte, error) { return readTier0(ws) },
	)
	if err != nil {
		t.Fatalf("handshake: %v", err)
	}
	_ = aead
	// A frame with valid Tier-3 flags but undecryptable body must drop the conn.
	bad := make([]byte, 40)
	bad[0] = 0x8C
	if err := ws.WriteMessage(websocket.BinaryMessage, bad); err != nil {
		t.Fatalf("write: %v", err)
	}
	_ = ws.SetReadDeadline(time.Now().Add(time.Second))
	if _, _, err := ws.ReadMessage(); err == nil {
		t.Fatal("expected server to close on bad frame")
	}
}

func TestServerHandshakeShortFrame(t *testing.T) {
	s := newTestServer(t)
	ws, resp, err := websocket.DefaultDialer.Dial("ws://"+s.Addr()+"/ws", nil)
	if resp != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = ws.Close() }()
	// 1-byte frame cannot be parsed as a message -> handshake read error.
	if err := ws.WriteMessage(websocket.BinaryMessage, []byte{0x00}); err != nil {
		t.Fatalf("write: %v", err)
	}
	_ = ws.SetReadDeadline(time.Now().Add(time.Second))
	if _, _, err := ws.ReadMessage(); err == nil {
		t.Fatal("expected server to close")
	}
}

func TestStreamCallbackError(t *testing.T) {
	s := newTestServer(t)
	c := dialClient(t, s)
	sentinel := errors.New("stop now")
	err := c.Stream(context.Background(), opStream, nil, func(*protocol.Message) (bool, error) {
		return false, sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("got %v", err)
	}
}

func TestServerClosePropagatesToClient(t *testing.T) {
	s := newTestServer(t)
	c := dialClient(t, s)
	// Sanity round-trip first.
	if _, err := c.Request(context.Background(), opEcho, []byte("x")); err != nil {
		t.Fatalf("warmup: %v", err)
	}
	_ = s.Stop(context.Background())
	// Subsequent request must fail once the read loop observes the closed conn.
	if _, err := c.Request(context.Background(), opSlow, nil); err == nil {
		t.Fatal("expected error after server stop")
	}
}

func TestTimeoutsConfigured(t *testing.T) {
	s, err := NewServer(ServerConfig{Address: "127.0.0.1:0", ReadTimeout: 2 * time.Second, WriteTimeout: 2 * time.Second})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	mustRegister(t, s.Handlers(), opEcho, func(_ context.Context, _ *Session, msg *protocol.Message) (*protocol.Message, error) {
		return &protocol.Message{Header: protocol.Header{OpCode: opResp}, Payload: msg.Payload}, nil
	})
	if err := s.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = s.Stop(context.Background()) })

	c := NewClient(ClientConfig{URL: "ws://" + s.Addr() + "/ws", WriteTimeout: 2 * time.Second, RequestTimeout: 2 * time.Second})
	if err := c.Connect(context.Background()); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer func() { _ = c.Close() }()
	resp, err := c.Request(context.Background(), opEcho, []byte("y"))
	if err != nil || string(resp.Payload) != "y" {
		t.Fatalf("request with timeouts: resp=%v err=%v", resp, err)
	}
}
