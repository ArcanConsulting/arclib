package transport

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"arcan-it.de/arclib/protocol"
)

const (
	opEcho   = protocol.OpCode(0x1000)
	opResp   = protocol.OpCode(0x1001)
	opStream = protocol.OpCode(0x1002)
	opFail   = protocol.OpCode(0x10F0)
	opSetID  = protocol.OpCode(0x10F1)
	opGetID  = protocol.OpCode(0x10F2)
	opSlow   = protocol.OpCode(0x10F3)
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	s, err := NewServer(ServerConfig{Address: "127.0.0.1:0"})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	reg := s.Handlers()
	mustRegister(t, reg, opEcho, func(_ context.Context, _ *Session, msg *protocol.Message) (*protocol.Message, error) {
		return &protocol.Message{Header: protocol.Header{OpCode: opResp}, Payload: msg.Payload}, nil
	})
	mustRegister(t, reg, opStream, func(_ context.Context, s *Session, _ *protocol.Message) (*protocol.Message, error) {
		for i := 0; i < 3; i++ {
			done := byte(0)
			if i == 2 {
				done = 1
			}
			if err := s.Send(opStream, []byte{byte(i), done}); err != nil {
				return nil, err
			}
		}
		return nil, nil //nolint:nilnil // streaming handler already replied via Session.Send
	})
	mustRegister(t, reg, opFail, func(_ context.Context, _ *Session, _ *protocol.Message) (*protocol.Message, error) {
		return nil, errors.New("boom")
	})
	mustRegister(t, reg, opSetID, func(_ context.Context, s *Session, _ *protocol.Message) (*protocol.Message, error) {
		s.SetIdentity("user-1", "fam-1", "dev-1")
		return &protocol.Message{Header: protocol.Header{OpCode: opResp}}, nil
	})
	mustRegister(t, reg, opGetID, func(_ context.Context, s *Session, _ *protocol.Message) (*protocol.Message, error) {
		return &protocol.Message{Header: protocol.Header{OpCode: opResp},
			Payload: []byte(s.UserID() + "/" + s.FamilyID() + "/" + s.DeviceID())}, nil
	})
	mustRegister(t, reg, opSlow, func(_ context.Context, _ *Session, _ *protocol.Message) (*protocol.Message, error) {
		return nil, nil //nolint:nilnil // deliberately never replies (timeout test)
	})

	if err := s.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = s.Stop(context.Background()) })
	return s
}

func mustRegister(t *testing.T, r *HandlerRegistry, op protocol.OpCode, f HandlerFunc) {
	t.Helper()
	if err := r.RegisterFunc(op, f); err != nil {
		t.Fatalf("register %s: %v", op, err)
	}
}

func dialClient(t *testing.T, s *Server) *Client {
	t.Helper()
	c := NewClient(ClientConfig{URL: "ws://" + s.Addr() + "/ws", RequestTimeout: 2 * time.Second})
	if err := c.Connect(context.Background()); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func TestRequestResponse(t *testing.T) {
	s := newTestServer(t)
	c := dialClient(t, s)

	resp, err := c.Request(context.Background(), opEcho, []byte("hello"))
	if err != nil {
		t.Fatalf("Request: %v", err)
	}
	if resp.Header.OpCode != opResp || string(resp.Payload) != "hello" {
		t.Fatalf("unexpected response: op=%s payload=%q", resp.Header.OpCode, resp.Payload)
	}
}

func TestRequestEncodesMsgpack(t *testing.T) {
	s := newTestServer(t)
	c := dialClient(t, s)
	// non-[]byte payload is msgpack-encoded; echo returns the raw bytes.
	resp, err := c.Request(context.Background(), opEcho, map[string]int{"n": 7})
	if err != nil {
		t.Fatalf("Request: %v", err)
	}
	if len(resp.Payload) == 0 {
		t.Fatal("expected msgpack payload echoed back")
	}
}

func TestStreaming(t *testing.T) {
	s := newTestServer(t)
	c := dialClient(t, s)

	var chunks [][]byte
	err := c.Stream(context.Background(), opStream, nil, func(msg *protocol.Message) (bool, error) {
		chunks = append(chunks, msg.Payload)
		return msg.Payload[1] == 1, nil
	})
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	if len(chunks) != 3 || chunks[0][0] != 0 || chunks[2][1] != 1 {
		t.Fatalf("unexpected chunks: %v", chunks)
	}
}

func TestHandlerError(t *testing.T) {
	s := newTestServer(t)
	c := dialClient(t, s)
	_, err := c.Request(context.Background(), opFail, nil)
	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected boom error, got %v", err)
	}
}

func TestUnknownOpcode(t *testing.T) {
	s := newTestServer(t)
	c := dialClient(t, s)
	_, err := c.Request(context.Background(), protocol.OpCode(0x10FF), nil)
	if err == nil {
		t.Fatal("expected error for unknown opcode")
	}
}

func TestIdentityPersistsAcrossRequests(t *testing.T) {
	s := newTestServer(t)
	c := dialClient(t, s)
	if _, err := c.Request(context.Background(), opSetID, nil); err != nil {
		t.Fatalf("setID: %v", err)
	}
	resp, err := c.Request(context.Background(), opGetID, nil)
	if err != nil {
		t.Fatalf("getID: %v", err)
	}
	if string(resp.Payload) != "user-1/fam-1/dev-1" {
		t.Fatalf("identity not persisted: %q", resp.Payload)
	}
}

func TestRequestTimeout(t *testing.T) {
	s := newTestServer(t)
	c := NewClient(ClientConfig{URL: "ws://" + s.Addr() + "/ws", RequestTimeout: 150 * time.Millisecond})
	if err := c.Connect(context.Background()); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer func() { _ = c.Close() }()
	_, err := c.Request(context.Background(), opSlow, nil)
	if !errors.Is(err, ErrRequestTimeout) {
		t.Fatalf("expected ErrRequestTimeout, got %v", err)
	}
}

func TestContextCancel(t *testing.T) {
	s := newTestServer(t)
	c := dialClient(t, s)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := c.Request(ctx, opSlow, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestCloseUnblocksAndRejects(t *testing.T) {
	s := newTestServer(t)
	c := dialClient(t, s)
	if err := c.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	// recv after close returns ErrClientClosed (send may also fail first).
	_, err := c.Request(context.Background(), opSlow, nil)
	if err == nil {
		t.Fatal("expected error after close")
	}
}

func TestSendBeforeConnect(t *testing.T) {
	c := NewClient(ClientConfig{URL: "ws://127.0.0.1:0/ws"})
	_, err := c.Request(context.Background(), opEcho, nil)
	if !errors.Is(err, ErrNotConnected) {
		t.Fatalf("expected ErrNotConnected, got %v", err)
	}
	if err := c.Close(); err != nil {
		t.Fatalf("Close on unconnected: %v", err)
	}
}

func TestConnectBadURL(t *testing.T) {
	c := NewClient(ClientConfig{URL: "ws://127.0.0.1:1/ws", HandshakeTimeout: 200 * time.Millisecond})
	if err := c.Connect(context.Background()); err == nil {
		t.Fatal("expected dial error")
	}
}

func TestServerHandshakeRejectsGarbage(t *testing.T) {
	s := newTestServer(t)
	// Raw client that skips the protocol handshake and sends junk.
	ws, resp, err := websocket.DefaultDialer.Dial("ws://"+s.Addr()+"/ws", nil)
	if resp != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = ws.Close() }()
	if err := ws.WriteMessage(websocket.BinaryMessage, []byte{0x00, 0x99, 0x99}); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Server must close the connection (handshake parse / unexpected op).
	_ = ws.SetReadDeadline(time.Now().Add(time.Second))
	if _, _, err := ws.ReadMessage(); err == nil {
		t.Fatal("expected server to close connection")
	}
}

func TestConcurrentClients(t *testing.T) {
	s := newTestServer(t)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := dialClient(t, s)
			resp, err := c.Request(context.Background(), opEcho, []byte("x"))
			if err != nil || string(resp.Payload) != "x" {
				t.Errorf("client request failed: %v", err)
			}
		}()
	}
	wg.Wait()
}

func TestServerConfigDefaultsAndAddr(t *testing.T) {
	s, err := NewServer(ServerConfig{})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	if s.cfg.Path != "/ws" || s.cfg.HandshakeTimeout == 0 || s.cfg.MaxMessageSize == 0 {
		t.Fatalf("defaults not applied: %+v", s.cfg)
	}
	if s.Addr() != "" {
		t.Fatalf("Addr before Start should be empty, got %q", s.Addr())
	}
}

func TestCheckOrigin(t *testing.T) {
	s, _ := NewServer(ServerConfig{AllowedOrigins: []string{"https://ok.example"}})
	tests := []struct {
		origin string
		want   bool
	}{
		{"", true},
		{"https://ok.example", true},
		{"https://evil.example", false},
	}
	for _, tt := range tests {
		r, _ := http.NewRequest(http.MethodGet, "/ws", http.NoBody)
		if tt.origin != "" {
			r.Header.Set("Origin", tt.origin)
		}
		if got := s.checkOrigin(r); got != tt.want {
			t.Errorf("origin %q: got %v want %v", tt.origin, got, tt.want)
		}
	}
	// empty allow-list permits everything.
	open, _ := NewServer(ServerConfig{})
	r, _ := http.NewRequest(http.MethodGet, "/ws", http.NoBody)
	r.Header.Set("Origin", "https://anything")
	if !open.checkOrigin(r) {
		t.Error("empty allow-list should permit all origins")
	}
}
