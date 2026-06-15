package transport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"arcan-it.de/arclib/protocol"
)

func TestRequestMarshalError(t *testing.T) {
	s := newTestServer(t)
	c := dialClient(t, s)
	// A channel cannot be msgpack-encoded -> send fails before any I/O.
	if _, err := c.Request(context.Background(), opEcho, make(chan int)); err == nil {
		t.Fatal("expected marshal error")
	}
}

func TestServerStartError(t *testing.T) {
	s, err := NewServer(ServerConfig{Address: "127.0.0.1:99999"}) // port out of range
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	if err := s.Start(); err == nil {
		_ = s.Stop(context.Background())
		t.Fatal("expected listen error")
	}
}

func TestStreamSendError(t *testing.T) {
	s := newTestServer(t)
	c := dialClient(t, s)
	err := c.Stream(context.Background(), opStream, make(chan int), func(*protocol.Message) (bool, error) {
		return true, nil
	})
	if err == nil {
		t.Fatal("expected send marshal error")
	}
}

func TestConnectHandshakeFailure(t *testing.T) {
	up := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := up.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = ws.Close() }()
		// Consume the client's SESSION_INIT, then reply with the wrong opcode so
		// the client's handshake rejects it.
		_, _, _ = readTier0(ws)
		_ = writeTier0(ws, protocol.OpNop, nil)
	}))
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/"
	c := NewClient(ClientConfig{URL: url, HandshakeTimeout: time.Second})
	if err := c.Connect(context.Background()); err == nil {
		t.Fatal("expected handshake failure")
		_ = c.Close()
	}
}
