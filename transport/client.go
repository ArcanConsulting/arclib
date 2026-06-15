package transport

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"arcan-it.de/arclib/msgpack"
	"arcan-it.de/arclib/protocol"
)

// Client errors.
var (
	// ErrClientClosed is returned once the client connection has been closed.
	ErrClientClosed = errors.New("transport: client closed")
	// ErrRequestTimeout is returned when a request exceeds RequestTimeout.
	ErrRequestTimeout = errors.New("transport: request timed out")
	// ErrNotConnected is returned when an operation is attempted before Connect.
	ErrNotConnected = errors.New("transport: not connected")
)

// ClientConfig configures a transport Client.
type ClientConfig struct {
	URL              string            // ws:// or wss:// endpoint
	HandshakeTimeout time.Duration     // dial + SESSION_INIT/ACK deadline
	WriteTimeout     time.Duration     // per-frame write deadline (0 = none)
	RequestTimeout   time.Duration     // per-request response deadline (0 = none)
	MaxMessageSize   int64             // max inbound frame size (0 = library default)
	Dialer           *websocket.Dialer // optional custom dialer
}

func (c *ClientConfig) applyDefaults() {
	if c.HandshakeTimeout == 0 {
		c.HandshakeTimeout = 10 * time.Second
	}
	if c.RequestTimeout == 0 {
		c.RequestTimeout = 120 * time.Second
	}
	if c.MaxMessageSize == 0 {
		c.MaxMessageSize = 8 << 20
	}
}

// Client is a MyClerk-Protocol client: it dials a Server, completes the Tier-3
// handshake, and issues request/response and streaming calls. Requests are
// serialized; the single read loop correlates each reply with the in-flight call.
type Client struct {
	cfg  ClientConfig
	ws   *websocket.Conn
	aead *protocol.AeadSession

	reqMu   sync.Mutex // serializes Request/Stream
	writeMu sync.Mutex // serializes frame writes
	inbox   chan inboxItem
	closed  chan struct{}
	once    sync.Once
}

type inboxItem struct {
	op      protocol.OpCode
	payload []byte
	err     error
}

// NewClient creates a Client for the given config. Call Connect before use.
func NewClient(cfg ClientConfig) *Client {
	cfg.applyDefaults()
	return &Client{
		cfg:    cfg,
		inbox:  make(chan inboxItem, 16),
		closed: make(chan struct{}),
	}
}

// Connect dials the server and performs the Tier-3 handshake.
func (c *Client) Connect(ctx context.Context) error {
	dialer := c.cfg.Dialer
	if dialer == nil {
		dialer = &websocket.Dialer{HandshakeTimeout: c.cfg.HandshakeTimeout}
	}
	ws, resp, err := dialer.DialContext(ctx, c.cfg.URL, nil)
	if resp != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		return err
	}
	if c.cfg.MaxMessageSize > 0 {
		ws.SetReadLimit(c.cfg.MaxMessageSize)
	}
	c.ws = ws

	write := func(op protocol.OpCode, payload []byte) error {
		c.setWriteDeadline()
		return writeTier0(ws, op, payload)
	}
	read := func() (protocol.OpCode, []byte, error) {
		_ = ws.SetReadDeadline(time.Now().Add(c.cfg.HandshakeTimeout))
		return readTier0(ws)
	}
	aead, err := clientHandshake(write, read)
	if err != nil {
		_ = ws.Close()
		return err
	}
	_ = ws.SetReadDeadline(time.Time{}) // clear handshake deadline
	c.aead = aead

	go c.readLoop()
	return nil
}

func (c *Client) readLoop() {
	for {
		_, data, err := c.ws.ReadMessage()
		if err != nil {
			c.push(inboxItem{err: err})
			return
		}
		op, plaintext, derr := c.aead.Decode(data)
		c.push(inboxItem{op: op, payload: plaintext, err: derr})
		if derr != nil {
			return
		}
	}
}

func (c *Client) push(item inboxItem) {
	select {
	case c.inbox <- item:
	case <-c.closed:
	}
}

// Request sends op with payload and returns the single response message. payload
// may be a []byte (sent verbatim) or any msgpack-encodable value.
func (c *Client) Request(ctx context.Context, op protocol.OpCode, payload any) (*protocol.Message, error) {
	c.reqMu.Lock()
	defer c.reqMu.Unlock()
	if err := c.send(op, payload); err != nil {
		return nil, err
	}
	return c.recv(ctx)
}

// Stream sends op with payload and invokes fn for each response frame until fn
// returns stop == true or a non-nil error. It is used for MAYA_STREAM where the
// caller stops on the chunk whose payload reports completion.
func (c *Client) Stream(ctx context.Context, op protocol.OpCode, payload any, fn func(*protocol.Message) (stop bool, err error)) error {
	c.reqMu.Lock()
	defer c.reqMu.Unlock()
	if err := c.send(op, payload); err != nil {
		return err
	}
	for {
		msg, err := c.recv(ctx)
		if err != nil {
			return err
		}
		stop, ferr := fn(msg)
		if ferr != nil {
			return ferr
		}
		if stop {
			return nil
		}
	}
}

// Close shuts the connection down and unblocks any pending call.
func (c *Client) Close() error {
	c.once.Do(func() { close(c.closed) })
	if c.ws != nil {
		return c.ws.Close()
	}
	return nil
}

func (c *Client) send(op protocol.OpCode, payload any) error {
	if c.aead == nil {
		return ErrNotConnected
	}
	data, err := toBytes(payload)
	if err != nil {
		return err
	}
	frame, err := c.aead.Encode(op, data)
	if err != nil {
		return err
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	c.setWriteDeadline()
	return c.ws.WriteMessage(websocket.BinaryMessage, frame)
}

func (c *Client) recv(ctx context.Context) (*protocol.Message, error) {
	var timeout <-chan time.Time
	if c.cfg.RequestTimeout > 0 {
		t := time.NewTimer(c.cfg.RequestTimeout)
		defer t.Stop()
		timeout = t.C
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.closed:
		return nil, ErrClientClosed
	case <-timeout:
		return nil, ErrRequestTimeout
	case item := <-c.inbox:
		if item.err != nil {
			return nil, item.err
		}
		if item.op == protocol.OpError {
			return nil, decodeError(item.payload)
		}
		return &protocol.Message{
			Header:  protocol.Header{Tier: protocol.TierEncrypted, OpCode: item.op},
			Payload: item.payload,
		}, nil
	}
}

func (c *Client) setWriteDeadline() {
	if c.cfg.WriteTimeout > 0 {
		_ = c.ws.SetWriteDeadline(time.Now().Add(c.cfg.WriteTimeout))
	}
}

func toBytes(payload any) ([]byte, error) {
	switch v := payload.(type) {
	case nil:
		return nil, nil
	case []byte:
		return v, nil
	default:
		return msgpack.Marshal(v)
	}
}
