package transport

import (
	"context"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"arcan-it.de/arclib/protocol"
)

// ServerConfig configures a transport Server.
type ServerConfig struct {
	Address          string        // TCP listen address, e.g. ":8094"
	Path             string        // WebSocket path, default "/ws"
	HandshakeTimeout time.Duration // deadline for the SESSION_INIT/ACK exchange
	ReadTimeout      time.Duration // per-frame read deadline (0 = none)
	WriteTimeout     time.Duration // per-frame write deadline (0 = none)
	MaxMessageSize   int64         // max inbound frame size (0 = library default)
	AllowedOrigins   []string      // permitted WS Origins (empty = allow all)
}

func (c *ServerConfig) applyDefaults() {
	if c.Path == "" {
		c.Path = "/ws"
	}
	if c.HandshakeTimeout == 0 {
		c.HandshakeTimeout = 10 * time.Second
	}
	if c.MaxMessageSize == 0 {
		c.MaxMessageSize = 8 << 20 // 8 MiB
	}
}

// Server accepts MyClerk-Protocol WebSocket connections, performs the Tier-3
// handshake for each, and dispatches decrypted messages to registered handlers.
type Server struct {
	cfg      ServerConfig
	handlers *HandlerRegistry
	upgrader websocket.Upgrader
	httpSrv  *http.Server
	nextID   atomic.Uint64

	mu     sync.Mutex
	ln     net.Listener
	conns  map[*serverConn]struct{}
	closed bool
}

// NewServer creates a Server. Register handlers via Handlers() before Start.
func NewServer(cfg ServerConfig) (*Server, error) {
	cfg.applyDefaults()
	s := &Server{
		cfg:      cfg,
		handlers: NewHandlerRegistry(),
		conns:    make(map[*serverConn]struct{}),
	}
	s.upgrader = websocket.Upgrader{
		HandshakeTimeout: cfg.HandshakeTimeout,
		CheckOrigin:      s.checkOrigin,
	}
	mux := http.NewServeMux()
	mux.HandleFunc(cfg.Path, s.handleWS)
	s.httpSrv = &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	return s, nil
}

// Handlers returns the server's handler registry.
func (s *Server) Handlers() *HandlerRegistry { return s.handlers }

// Start binds the listener and serves in the background.
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.cfg.Address)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.ln = ln
	s.mu.Unlock()
	go func() { _ = s.httpSrv.Serve(ln) }()
	return nil
}

// Addr returns the actual listen address (useful when Address used port 0).
func (s *Server) Addr() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ln != nil {
		return s.ln.Addr().String()
	}
	return s.cfg.Address
}

// Stop closes all live connections and shuts the HTTP server down.
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	s.closed = true
	conns := make([]*serverConn, 0, len(s.conns))
	for c := range s.conns {
		conns = append(conns, c)
	}
	s.mu.Unlock()
	for _, c := range conns {
		c.close()
	}
	return s.httpSrv.Shutdown(ctx)
}

func (s *Server) checkOrigin(r *http.Request) bool {
	if len(s.cfg.AllowedOrigins) == 0 {
		return true
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	for _, allowed := range s.cfg.AllowedOrigins {
		if allowed == origin {
			return true
		}
	}
	return false
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	if s.cfg.MaxMessageSize > 0 {
		ws.SetReadLimit(s.cfg.MaxMessageSize)
	}
	c := &serverConn{ws: ws, server: s}

	aead, err := serverHandshake(c.handshakeWrite, c.handshakeRead)
	if err != nil {
		_ = ws.Close()
		return
	}
	c.session = newSession(s.nextID.Add(1), aead, c.writeFrame)

	s.register(c)
	defer s.unregister(c)
	c.readLoop()
}

func (s *Server) register(c *serverConn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		c.close()
		return
	}
	s.conns[c] = struct{}{}
}

func (s *Server) unregister(c *serverConn) {
	s.mu.Lock()
	delete(s.conns, c)
	s.mu.Unlock()
	c.close()
}

// serverConn is one accepted connection.
type serverConn struct {
	ws        *websocket.Conn
	server    *Server
	session   *Session
	writeMu   sync.Mutex
	closeOnce sync.Once
}

func (c *serverConn) handshakeWrite(op protocol.OpCode, payload []byte) error {
	c.setWriteDeadline()
	return writeTier0(c.ws, op, payload)
}

func (c *serverConn) handshakeRead() (protocol.OpCode, []byte, error) {
	if c.server.cfg.HandshakeTimeout > 0 {
		_ = c.ws.SetReadDeadline(time.Now().Add(c.server.cfg.HandshakeTimeout))
	}
	return readTier0(c.ws)
}

func (c *serverConn) writeFrame(frame []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	c.setWriteDeadline()
	return c.ws.WriteMessage(websocket.BinaryMessage, frame)
}

func (c *serverConn) setWriteDeadline() {
	if c.server.cfg.WriteTimeout > 0 {
		_ = c.ws.SetWriteDeadline(time.Now().Add(c.server.cfg.WriteTimeout))
	}
}

func (c *serverConn) readLoop() {
	for {
		if c.server.cfg.ReadTimeout > 0 {
			_ = c.ws.SetReadDeadline(time.Now().Add(c.server.cfg.ReadTimeout))
		}
		_, data, err := c.ws.ReadMessage()
		if err != nil {
			return
		}
		op, plaintext, err := c.session.aead.Decode(data)
		if err != nil {
			return // a frame that fails authentication ends the connection
		}
		msg := &protocol.Message{
			Header:  protocol.Header{Tier: protocol.TierEncrypted, OpCode: op},
			Payload: plaintext,
		}
		resp, herr := c.server.handlers.dispatch(context.Background(), c.session, msg)
		if herr != nil {
			if sendErr := c.session.Send(protocol.OpError, encodeError(herr)); sendErr != nil {
				return
			}
			continue
		}
		if resp != nil {
			if err := c.session.Send(resp.Header.OpCode, resp.Payload); err != nil {
				return
			}
		}
	}
}

func (c *serverConn) close() {
	c.closeOnce.Do(func() { _ = c.ws.Close() })
}
